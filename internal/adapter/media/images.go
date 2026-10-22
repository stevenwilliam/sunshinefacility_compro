// Package media implements app.ImageProcessor: decode, bound, resize and
// re-encode to WebP (BR-4.2). Re-encoding is the point — whatever else was in
// the uploaded file (metadata, a polyglot payload) does not survive it.
package media

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // decoders
	_ "image/png"

	"github.com/gen2brain/webp"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
)

type WebP struct{ Quality int }

var _ app.ImageProcessor = WebP{}

func (p WebP) Process(b []byte, widths []int, maxDim int) (*app.ProcessedImage, error) {
	// Read the dimensions first: a 20 000 × 20 000 PNG is a few KB on disk
	// and 1.6 GB in memory. Refuse it before decoding.
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("not a readable image: %w", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxDim || cfg.Height > maxDim {
		return nil, fmt.Errorf("image is %d×%d; the limit is %d px", cfg.Width, cfg.Height, maxDim)
	}
	src, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("not a readable image: %w", err)
	}
	q := p.Quality
	if q == 0 {
		q = 82
	}
	out := &app.ProcessedImage{Width: cfg.Width, Height: cfg.Height}
	seen := map[int]bool{}
	for _, w := range widths {
		if w > cfg.Width {
			w = cfg.Width // never upscale; the original width becomes the top rendition
		}
		if seen[w] {
			continue
		}
		seen[w] = true
		h := int(float64(cfg.Height) * float64(w) / float64(cfg.Width))
		if h < 1 {
			h = 1
		}
		dst := image.NewNRGBA(image.Rect(0, 0, w, h))
		if w == cfg.Width {
			draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
		} else {
			xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Src, nil)
		}
		var buf bytes.Buffer
		if err := webp.Encode(&buf, dst, webp.Options{Quality: q, Method: 4}); err != nil {
			return nil, fmt.Errorf("webp encode: %w", err)
		}
		out.Variants = append(out.Variants, app.EncodedVariant{Width: w, Bytes: buf.Bytes()})
	}
	return out, nil
}
