package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/logging"
)

const ctxActor = "sfg.actor"
const ctxClaimsJTI = "sfg.jti"
const ctxClaimsExp = "sfg.exp"

// requestID propagates (or creates) a request id through the context and
// back to the client, so a reported problem can be found in the log.
func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if len(rid) > 64 || rid == "" || strings.ContainsAny(rid, "\r\n\t ") {
			rid = id.NewString()
		}
		c.Header("X-Request-Id", rid)
		c.Request = c.Request.WithContext(logging.WithTrace(c.Request.Context(), rid))
		c.Next()
	}
}

// accessLog records method, route, status and duration. No query strings,
// no bodies, no IPs: there is personal data in all three (CLAUDE.md §7).
func accessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "(unmatched)"
		}
		logging.FromContext(c.Request.Context(), log).Info("http",
			slog.String("method", c.Request.Method), slog.String("route", route),
			slog.Int("status", c.Writer.Status()), slog.Int64("ms", time.Since(start).Milliseconds()))
	}
}

// Content-Security-Policy. No 'unsafe-inline' anywhere: the page's data is a
// JSON <script type="application/json"> (never executed), React sets styles
// through the CSSOM (not governed by style-src), and the GA4 bootstrap lives
// in the bundle rather than inline. Google is allowed only for analytics
// (when switched on) and the map frame (decision Q20).
const csp = "default-src 'self'; " +
	"script-src 'self' https://www.googletagmanager.com; " +
	"style-src 'self'; " +
	"img-src 'self' data: blob: https://www.google-analytics.com https://www.googletagmanager.com; " +
	"font-src 'self'; " +
	"connect-src 'self' https://*.google-analytics.com https://*.analytics.google.com https://www.googletagmanager.com; " +
	"frame-src https://www.google.com https://maps.google.com; " +
	"frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'"

func securityHeaders(production bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		if production {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// requireAuth resolves the access token to an actor. Deny by default: a
// handler under it either gets an actor or never runs.
func (s *Server) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, claims, err := s.auth.Authenticate(c.Request.Context(), bearer(c))
		if err != nil {
			s.fail(c, err)
			return
		}
		c.Set(ctxActor, app.Actor{UserID: u.ID, Email: u.Email, Role: u.Role, IP: clientIP(c),
			UserAgent: c.Request.UserAgent(), RequestID: c.Writer.Header().Get("X-Request-Id")})
		c.Set(ctxClaimsJTI, claims.JTI)
		if claims.ExpiresAt != nil {
			c.Set(ctxClaimsExp, claims.ExpiresAt.Time)
		}
		c.Next()
	}
}

// require declares the permission a route needs (BR-8.1). The use-cases
// check again; this makes the route table itself the readable statement of
// who may call what.
func (s *Server) require(perms ...access.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		a := actorOf(c)
		for _, p := range perms {
			if access.Can(access.Role(a.Role), p) {
				c.Next()
				return
			}
		}
		s.fail(c, apierror.New(apierror.CodeForbidden, "You do not have access to this."))
	}
}

func actorOf(c *gin.Context) app.Actor {
	if v, ok := c.Get(ctxActor); ok {
		return v.(app.Actor)
	}
	return app.Actor{IP: clientIP(c), UserAgent: c.Request.UserAgent()}
}

// limit is a per-IP token bucket for an endpoint group.
func (s *Server) limit(name string, burst int, perMinute float64) gin.HandlerFunc {
	l := s.limiter(name, burst, perMinute)
	return func(c *gin.Context) {
		if allowed, wait := l.Allow(clientIP(c)); !allowed {
			c.Header("Retry-After", itoa(int(wait.Seconds())+1))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": apierror.New(apierror.CodeRateLimited, "Too many requests. Please wait a moment and try again.")})
			return
		}
		c.Next()
	}
}

// bodyLimit caps a request body before anything reads it.
func bodyLimit(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}

func itoa(n int) string {
	if n <= 0 {
		return "1"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
