// Package filetype identifies uploads by their MAGIC BYTES, never by the
// extension or the client's Content-Type (CLAUDE.md §7). A file named
// cv.pdf that is really an HTML page is refused, not stored.
package filetype

import (
	"archive/zip"
	"bytes"
)

const (
	PNG  = "image/png"
	JPEG = "image/jpeg"
	WebP = "image/webp"
	PDF  = "application/pdf"
	DOC  = "application/msword"
	DOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

// Image — BR-4.1: PNG, JPEG or WebP. SVG is deliberately absent: an SVG is
// a document that can carry script, and the only SVGs on the site are the
// brand marks, which arrive through the seed (BR-4.4).
func Image(b []byte) (string, bool) {
	switch {
	case len(b) >= 8 && bytes.Equal(b[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}):
		return PNG, true
	case len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return JPEG, true
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")):
		return WebP, true
	}
	return "", false
}

// Document — BR-6.1: a CV is PDF, legacy Word (OLE compound file) or DOCX.
// DOCX is a zip, and so is every other Office file and every .jar; it only
// counts if the archive actually contains a Word document part.
func Document(b []byte) (string, bool) {
	switch {
	case len(b) >= 5 && bytes.Equal(b[:5], []byte("%PDF-")):
		return PDF, true
	case len(b) >= 8 && bytes.Equal(b[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}):
		return DOC, true
	case len(b) >= 4 && bytes.Equal(b[:4], []byte{'P', 'K', 0x03, 0x04}):
		zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
		if err != nil {
			return "", false
		}
		for _, f := range zr.File {
			if f.Name == "word/document.xml" {
				return DOCX, true
			}
		}
	}
	return "", false
}

// Extension for a stored document, chosen from the sniffed type — never
// echoed from the client's filename.
func Extension(mime string) string {
	switch mime {
	case PDF:
		return ".pdf"
	case DOC:
		return ".doc"
	case DOCX:
		return ".docx"
	case PNG:
		return ".png"
	case JPEG:
		return ".jpg"
	case WebP:
		return ".webp"
	}
	return ".bin"
}
