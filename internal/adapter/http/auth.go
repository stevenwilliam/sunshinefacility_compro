package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
)

// The refresh token lives in an httpOnly cookie scoped to the auth routes,
// SameSite=Strict. The access token is held in memory by the CMS and sent as
// a Bearer header, which no other site can make a browser send — so neither
// half is usable for CSRF.
const refreshCookie = "sfg_rt"

func (s *Server) setRefresh(c *gin.Context, sess *app.Session) {
	secure := s.Production || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	http.SetCookie(c.Writer, &http.Cookie{Name: refreshCookie, Value: sess.RefreshToken, Path: "/api/v1/auth",
		Expires: sess.RefreshExpiresAt, HttpOnly: true, Secure: secure, SameSite: http.SameSiteStrictMode})
}

func (s *Server) clearRefresh(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{Name: refreshCookie, Value: "", Path: "/api/v1/auth", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteStrictMode})
}

func (s *Server) getSetup(c *gin.Context) {
	need, err := s.auth.SetupNeeded(c.Request.Context())
	if err != nil {
		s.fail(c, err)
		return
	}
	ok(c, gin.H{"setupNeeded": need})
}

func (s *Server) postSetup(c *gin.Context) {
	var b struct{ Token, Email, Name, Password string }
	if err := c.ShouldBindJSON(&b); err != nil {
		s.fail(c, apierror.Validation("Invalid request.", nil))
		return
	}
	sess, err := s.auth.Setup(c.Request.Context(), b.Token, b.Email, b.Name, b.Password, actorOf(c))
	if err != nil {
		s.fail(c, err)
		return
	}
	s.setRefresh(c, sess)
	ok(c, sess)
}

func (s *Server) postLogin(c *gin.Context) {
	var b struct{ Email, Password string }
	if err := c.ShouldBindJSON(&b); err != nil {
		s.fail(c, apierror.Validation("Enter your email and password.", nil))
		return
	}
	// Per-email limit on top of the per-IP one, so a password spray from
	// many addresses at one account still runs into a wall (BR-8.3).
	key := strings.ToLower(strings.TrimSpace(b.Email))
	if allowed, _ := s.limiter("login-email", 10, 5).Allow(key); !allowed {
		s.fail(c, apierror.New(apierror.CodeRateLimited, "Too many sign-in attempts. Please wait a few minutes."))
		return
	}
	sess, err := s.auth.Login(c.Request.Context(), b.Email, b.Password, clientIP(c), c.Request.UserAgent())
	if err != nil {
		s.fail(c, err)
		return
	}
	s.setRefresh(c, sess)
	ok(c, sess)
}

func (s *Server) postRefresh(c *gin.Context) {
	ck, _ := c.Cookie(refreshCookie)
	sess, err := s.auth.Refresh(c.Request.Context(), ck, clientIP(c), c.Request.UserAgent())
	if err != nil {
		s.clearRefresh(c)
		s.fail(c, err)
		return
	}
	s.setRefresh(c, sess)
	ok(c, sess)
}

func (s *Server) postLogout(c *gin.Context) {
	ck, _ := c.Cookie(refreshCookie)
	var jti string
	exp := time.Now().Add(time.Hour)
	if tok := bearer(c); tok != "" {
		if _, claims, err := s.auth.Authenticate(c.Request.Context(), tok); err == nil {
			jti = claims.JTI
			if claims.ExpiresAt != nil {
				exp = claims.ExpiresAt.Time
			}
		}
	}
	if err := s.auth.Logout(c.Request.Context(), jti, exp, ck); err != nil {
		s.fail(c, err)
		return
	}
	s.clearRefresh(c)
	ok(c, gin.H{"ok": true})
}

func (s *Server) getMe(c *gin.Context) {
	a := actorOf(c)
	u, _, err := s.auth.Authenticate(c.Request.Context(), bearer(c))
	if err != nil || u.ID != a.UserID {
		s.fail(c, apierror.New(apierror.CodeUnauthenticated, "Please sign in."))
		return
	}
	ok(c, s.auth.Me(u))
}

func (s *Server) postPassword(c *gin.Context) {
	var b struct{ Current, Password string }
	if err := c.ShouldBindJSON(&b); err != nil {
		s.fail(c, apierror.Validation("Invalid request.", nil))
		return
	}
	if err := s.auth.ChangePassword(c.Request.Context(), actorOf(c), b.Current, b.Password); err != nil {
		s.fail(c, err)
		return
	}
	s.clearRefresh(c)
	ok(c, gin.H{"ok": true})
}
