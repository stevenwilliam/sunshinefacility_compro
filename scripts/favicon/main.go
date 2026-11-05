// Command favicon cuts the "G" of the SFG wordmark (red, with the green
// notch) out of the brand logo and writes the site's icons into web/public:
//
//	favicon.ico           16, 32, 48 (PNG entries), transparent
//	icon-192.png          192, transparent (Android, rel=icon sizes=192x192)
//	apple-touch-icon.png  180, white background (iOS masks transparency black)
//
// Run with `make favicon`. Deterministic: same logo in, same bytes out.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/image/draw"
)

const (
	logo = "docs/design/handoff/design/assets/brand/sfg-logo-primary.png"
	out  = "web/public"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "favicon:", err)
		os.Exit(1)
	}
}

func run() error {
	f, err := os.Open(logo)
	if err != nil {
		return err
	}
	defer f.Close()
	logoImg, err := png.Decode(f)
	if err != nil {
		return err
	}
	src, g, err := letterG(logoImg)
	if err != nil {
		return err
	}

	var entries [][]byte
	for _, n := range []int{16, 32, 48} {
		b, err := encode(square(src, g, n, 0.04, nil))
		if err != nil {
			return err
		}
		entries = append(entries, b)
	}
	if err := os.WriteFile(filepath.Join(out, "favicon.ico"), ico(entries, []int{16, 32, 48}), 0o644); err != nil {
		return err
	}
	if err := save("icon-192.png", square(src, g, 192, 0.08, nil)); err != nil {
		return err
	}
	return save("apple-touch-icon.png", square(src, g, 180, 0.16, color.White))
}

// letterG returns the G of the SFG mark (its red body and the green notch)
// on a transparent canvas. The letters overlap column-wise (slanted cuts),
// so it labels connected shapes in the mark (the first run of inked
// columns; the wordmark follows a gap) and keeps the rightmost sizeable one:
// in this file S and F are joined, and the notch touches the G.
func letterG(img image.Image) (*image.RGBA, image.Rectangle, error) {
	b := img.Bounds()
	inked := func(x, y int) bool { _, _, _, a := img.At(x, y).RGBA(); return a > 0x8000 }
	markEnd := b.Min.X
	for x := b.Min.X; x < b.Max.X; x++ {
		any := false
		for y := b.Min.Y; y < b.Max.Y && !any; y++ {
			any = inked(x, y)
		}
		if !any && x > b.Min.X+10 {
			break
		}
		markEnd = x + 1
	}
	mark := image.Rect(b.Min.X, b.Min.Y, markEnd, b.Max.Y)

	label := make(map[image.Point]int)
	type shape struct {
		px  []image.Point
		box image.Rectangle
	}
	var shapes []shape
	for y := mark.Min.Y; y < mark.Max.Y; y++ {
		for x := mark.Min.X; x < mark.Max.X; x++ {
			p := image.Pt(x, y)
			if label[p] != 0 || !inked(x, y) {
				continue
			}
			id := len(shapes) + 1
			sh := shape{box: image.Rectangle{Min: p, Max: p.Add(image.Pt(1, 1))}}
			stack := []image.Point{p}
			label[p] = id
			for len(stack) > 0 {
				q := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				sh.px = append(sh.px, q)
				sh.box = sh.box.Union(image.Rectangle{Min: q, Max: q.Add(image.Pt(1, 1))})
				for _, d := range []image.Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					n := q.Add(d)
					if n.In(mark) && label[n] == 0 && inked(n.X, n.Y) {
						label[n] = id
						stack = append(stack, n)
					}
				}
			}
			shapes = append(shapes, sh)
		}
	}
	var big []shape
	for _, sh := range shapes {
		if len(sh.px) > 500 { // anti-aliasing specks are not letters
			big = append(big, sh)
		}
	}
	sort.Slice(big, func(i, j int) bool { return big[i].box.Min.X < big[j].box.Min.X })
	if os.Getenv("FAVICON_DEBUG") != "" {
		for _, sh := range big {
			fmt.Println(sh.box, len(sh.px))
		}
	}
	if len(big) < 2 {
		return nil, image.Rectangle{}, fmt.Errorf("expected S/F and G in the mark; found %d shapes", len(big))
	}
	g := big[len(big)-1]
	gid := label[g.px[0]]
	out := image.NewRGBA(b)
	// Copy the G's box except pixels of another letter, so the G's faint
	// anti-aliased edge (below the ink threshold, unlabelled) comes along.
	r := g.box.Inset(-2).Intersect(mark)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if id := label[image.Pt(x, y)]; id != 0 && id != gid {
				continue
			}
			if x < g.box.Min.X+(g.box.Dx()/2) && label[image.Pt(x, y)] == 0 && nearOther(label, gid, x, y) {
				continue
			}
			out.Set(x, y, img.At(x, y))
		}
	}
	return out, g.box, nil
}

// nearOther reports whether an unlabelled (faint) pixel borders a pixel of
// another letter, i.e. it is that letter's anti-aliased edge, not the G's.
func nearOther(label map[image.Point]int, gid, x, y int) bool {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if id := label[image.Pt(x+dx, y+dy)]; id != 0 && id != gid {
				return true
			}
		}
	}
	return false
}

// square scales the crop into an n×n canvas with pad (fraction of n) on
// every side, centred; bg nil means transparent.
func square(src image.Image, crop image.Rectangle, n int, pad float64, bg color.Color) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, n, n))
	if bg != nil {
		draw.Draw(dst, dst.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	}
	box := float64(n) * (1 - 2*pad)
	w, h := float64(crop.Dx()), float64(crop.Dy())
	s := box / max(w, h)
	dw, dh := int(w*s+0.5), int(h*s+0.5)
	x0, y0 := (n-dw)/2, (n-dh)/2
	draw.CatmullRom.Scale(dst, image.Rect(x0, y0, x0+dw, y0+dh), src, crop, draw.Over, nil)
	return dst
}

func encode(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&buf, img)
	return buf.Bytes(), err
}

func save(name string, img image.Image) error {
	b, err := encode(img)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, name), b, 0o644)
}

// ico wraps PNG images in an ICO container (PNG entries: Vista and later,
// every current browser).
func ico(pngs [][]byte, sizes []int) []byte {
	var buf bytes.Buffer
	le := binary.LittleEndian
	_ = binary.Write(&buf, le, [3]uint16{0, 1, uint16(len(pngs))})
	offset := 6 + 16*len(pngs)
	for i, p := range pngs {
		s := byte(sizes[i] % 256) // 256 is written as 0
		buf.Write([]byte{s, s, 0, 0})
		_ = binary.Write(&buf, le, [2]uint16{1, 32})
		_ = binary.Write(&buf, le, [2]uint32{uint32(len(p)), uint32(offset)})
		offset += len(p)
	}
	for _, p := range pngs {
		buf.Write(p)
	}
	return buf.Bytes()
}
