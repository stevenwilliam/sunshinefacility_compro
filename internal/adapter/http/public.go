package http

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/forms"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
)

func (s *Server) getBundle(c *gin.Context) {
	b, raw, err := s.Public.Get(c.Request.Context())
	if err != nil {
		s.fail(c, err)
		return
	}
	etag := `"` + b.Version + `"`
	c.Header("Cache-Control", "no-cache")
	c.Header("ETag", etag)
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", raw)
}

type quotationBody struct {
	forms.QuotationInput
	Website   string `json:"website"` // the honeypot: hidden from people
	ElapsedMS int64  `json:"elapsedMs"`
}

func (s *Server) postQuotation(c *gin.Context) {
	var body quotationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		s.fail(c, apierror.Validation("check", map[string]string{"_form": "invalid"}))
		return
	}
	first, err := s.Submissions.SubmitQuotation(c.Request.Context(), body.QuotationInput, app.Envelope{
		IP: clientIP(c), UserAgent: c.Request.UserAgent(), Honeypot: body.Website, ElapsedMS: body.ElapsedMS})
	if errors.Is(err, app.ErrDiscarded) {
		s.log.Info("quotation discarded as automated")
		ok(c, gin.H{"ok": true, "firstName": forms.FirstName(body.Name)})
		return
	}
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, gin.H{"ok": true, "firstName": first})
}

func (s *Server) postApplication(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(8 << 20); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			s.fail(c, apierror.Validation("check", map[string]string{"cv": "file_size"}))
			return
		}
		s.fail(c, apierror.Validation("check", map[string]string{"_form": "invalid"}))
		return
	}
	elapsed, _ := strconv.ParseInt(c.PostForm("elapsedMs"), 10, 64)
	in := forms.ApplicationInput{Name: c.PostForm("name"), Phone: c.PostForm("phone"), Email: c.PostForm("email"),
		Position: c.PostForm("position"), Message: c.PostForm("message"), Lang: c.PostForm("lang")}
	var cv *app.CV
	if fh, err := c.FormFile("cv"); err == nil {
		f, err := fh.Open()
		if err != nil {
			s.fail(c, err)
			return
		}
		// Read one byte past the largest allowed CV so an oversized file is
		// recognised without buffering all of it.
		data, err := io.ReadAll(io.LimitReader(f, 21<<20))
		_ = f.Close()
		if err != nil {
			s.fail(c, err)
			return
		}
		cv = &app.CV{Filename: fh.Filename, Data: data}
	}
	err := s.Submissions.SubmitApplication(c.Request.Context(), in, cv, app.Envelope{
		IP: clientIP(c), UserAgent: c.Request.UserAgent(), Honeypot: c.PostForm("website"), ElapsedMS: elapsed})
	if errors.Is(err, app.ErrDiscarded) {
		s.log.Info("application discarded as automated")
		ok(c, gin.H{"ok": true})
		return
	}
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, gin.H{"ok": true})
}

// getMediaFile serves /media/<id>/<hash12>-<w>.webp and /media/<id>/<hash12>.svg.
// The hash is in the URL, so the response is cached for a year and marked
// immutable — the filename tells the truth (impeccable §5).
func (s *Server) getMediaFile(c *gin.Context) {
	file := c.Param("file")
	var hash string
	var width int
	svg := false
	switch {
	case strings.HasSuffix(file, ".svg"):
		hash, svg = strings.TrimSuffix(file, ".svg"), true
	case strings.HasSuffix(file, ".webp"):
		base := strings.TrimSuffix(file, ".webp")
		i := strings.LastIndex(base, "-")
		if i < 0 {
			c.Status(http.StatusNotFound)
			return
		}
		hash = base[:i]
		w, err := strconv.Atoi(base[i+1:])
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		width = w
	default:
		c.Status(http.StatusNotFound)
		return
	}
	data, ct, err := s.Media.Object(c.Request.Context(), c.Param("id"), hash, width, svg)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	if svg {
		// An SVG opened directly is a document; this keeps it inert.
		c.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	}
	c.Data(http.StatusOK, ct, data)
}
