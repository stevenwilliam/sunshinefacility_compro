package media

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func pngOf(w, h int) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{uint8(x), uint8(y), 120, 255})
		}
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}

// BR-4.2: renditions at the requested widths, never upscaled, and WebP.
func TestBR4_2_Renditions(t *testing.T) {
	out, err := WebP{}.Process(pngOf(1200, 800), []int{480, 960, 1600}, 8000)
	if err != nil {
		t.Fatal(err)
	}
	var widths []int
	for _, v := range out.Variants {
		widths = append(widths, v.Width)
		if !bytes.HasPrefix(v.Bytes, []byte("RIFF")) || !bytes.Equal(v.Bytes[8:12], []byte("WEBP")) {
			t.Fatalf("rendition %d is not WebP", v.Width)
		}
	}
	if len(widths) != 3 || widths[0] != 480 || widths[1] != 960 || widths[2] != 1200 {
		t.Fatalf("widths = %v, want [480 960 1200] (1600 capped at the original)", widths)
	}
}

// BR-4.1: the dimension limit is checked before decoding.
func TestBR4_1_DimensionLimit(t *testing.T) {
	if _, err := (WebP{}).Process(pngOf(600, 100), []int{480}, 500); err == nil {
		t.Fatal("an image wider than the limit was accepted")
	}
}

func TestRejectsGarbage(t *testing.T) {
	if _, err := (WebP{}).Process([]byte("\x89PNG\r\n\x1a\nnot really"), []int{480}, 8000); err == nil {
		t.Fatal("a corrupt PNG was accepted")
	}
}
