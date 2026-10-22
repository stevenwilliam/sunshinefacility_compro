package filetype

import (
	"archive/zip"
	"bytes"
	"testing"
)

func zipWith(names ...string) []byte {
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for _, n := range names {
		f, _ := w.Create(n)
		_, _ = f.Write([]byte("x"))
	}
	_ = w.Close()
	return b.Bytes()
}

func TestBR6_1_Documents(t *testing.T) {
	cases := []struct {
		name string
		b    []byte
		want string
		ok   bool
	}{
		{"pdf", []byte("%PDF-1.7\n..."), PDF, true},
		{"doc", []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1, 0, 0}, DOC, true},
		{"docx", zipWith("[Content_Types].xml", "word/document.xml"), DOCX, true},
		{"xlsx is not a CV", zipWith("[Content_Types].xml", "xl/workbook.xml"), "", false},
		{"html renamed .pdf", []byte("<html><script>alert(1)</script>"), "", false},
		{"empty", nil, "", false},
		{"png is not a CV", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}, "", false},
	}
	for _, c := range cases {
		got, ok := Document(c.b)
		if got != c.want || ok != c.ok {
			t.Errorf("%s: Document = (%q, %v), want (%q, %v)", c.name, got, ok, c.want, c.ok)
		}
	}
}

func TestBR4_1_Images(t *testing.T) {
	cases := []struct {
		name string
		b    []byte
		want string
		ok   bool
	}{
		{"png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n', 0}, PNG, true},
		{"jpeg", []byte{0xFF, 0xD8, 0xFF, 0xE0}, JPEG, true},
		{"webp", append([]byte("RIFF\x00\x00\x00\x00WEBP"), 0), WebP, true},
		{"svg is refused (BR-4.4)", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script/></svg>`), "", false},
		{"gif is refused", []byte("GIF89a"), "", false},
	}
	for _, c := range cases {
		got, ok := Image(c.b)
		if got != c.want || ok != c.ok {
			t.Errorf("%s: Image = (%q, %v), want (%q, %v)", c.name, got, ok, c.want, c.ok)
		}
	}
}
