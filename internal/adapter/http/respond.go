// Package http is the HTTP adapter: routing, middleware, request/response
// mapping. It holds no business rules — handlers parse, call one use-case,
// and render.
package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/calendar"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/csvexport"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/logging"
)

// fail renders the one error shape (04-api-specification.md §2). An
// internal error is logged with its cause and shown as a generic message.
func (s *Server) fail(c *gin.Context, err error) {
	if errors.Is(err, app.ErrNotFound) {
		err = apierror.NotFound("Item")
	}
	e := apierror.As(err)
	status := e.HTTPStatus()
	if status >= 500 {
		logging.FromContext(c.Request.Context(), s.log).Error("request failed",
			slog.String("path", c.FullPath()), slog.String("error", err.Error()))
		e = &apierror.Error{Code: apierror.CodeInternal, Message: "Something went wrong. Please try again."}
	} else if e.Cause != nil {
		logging.FromContext(c.Request.Context(), s.log).Info("request refused",
			slog.String("path", c.FullPath()), slog.String("code", string(e.Code)), slog.String("cause", e.Cause.Error()))
	}
	if e.Code == apierror.CodeRateLimited {
		c.Header("Retry-After", "600")
	}
	c.AbortWithStatusJSON(status, gin.H{"error": e})
}

func ok(c *gin.Context, v any) { c.JSON(http.StatusOK, v) }

// csv writes a pipe-delimited export (BR-9.2). Every cell goes through the
// formula guard in csvexport.
func writeCSV(c *gin.Context, name string, header []string, rows [][]string) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, csvexport.Filename(name, time.Now(), calendar.Zone)))
	c.Header("Cache-Control", "no-store")
	c.Status(http.StatusOK)
	w := csvexport.New(c.Writer)
	w.Write(header)
	for _, r := range rows {
		w.Write(r)
	}
	_ = w.Flush()
}

func wantsCSV(c *gin.Context) bool { return c.Query("format") == "csv" }

func clientIP(c *gin.Context) string { return c.ClientIP() }

func bearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

func ts(t time.Time) string { return csvexport.Timestamp(t, calendar.Zone) }
