package http

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/ratelimit"
)

type Deps struct {
	Log         *slog.Logger
	Production  bool
	BaseURL     string // public origin, for canonical URLs and link previews
	Auth        *app.Auth
	Content     *app.Content
	Media       *app.Media
	Public      *app.Public
	Submissions *app.Submissions
	Audit       *app.AuditLog
	Ready       func(ctx context.Context) error
	Web         fs.FS // the built SPA (web/dist)
	Proxies     []string
}

type Server struct {
	Deps
	auth *app.Auth
	log  *slog.Logger

	mu       sync.Mutex
	limiters map[string]*ratelimit.Limiter
	index    []byte
}

func (s *Server) limiter(name string, burst int, perMinute float64) *ratelimit.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, ok := s.limiters[name]; ok {
		return l
	}
	l := ratelimit.New(burst, perMinute)
	s.limiters[name] = l
	return l
}

func New(d Deps) (*Server, http.Handler, error) {
	s := &Server{Deps: d, auth: d.Auth, log: d.Log, limiters: map[string]*ratelimit.Limiter{}}
	if d.Web != nil {
		b, err := fs.ReadFile(d.Web, "index.html")
		if err != nil {
			return nil, nil, err
		}
		s.index = b
	}
	if d.Production {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	if err := r.SetTrustedProxies(d.Proxies); err != nil {
		return nil, nil, err
	}
	r.Use(gin.Recovery(), requestID(), securityHeaders(d.Production), accessLog(d.Log))
	r.MaxMultipartMemory = 8 << 20
	r.RedirectTrailingSlash = true
	r.HandleMethodNotAllowed = false

	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := d.Ready(ctx); err != nil {
			s.log.Error("not ready", "error", err.Error())
			c.String(http.StatusServiceUnavailable, "not ready")
			return
		}
		c.String(http.StatusOK, "ready")
	})

	api := r.Group("/api/v1", noStore())
	{
		pub := api.Group("/public")
		pub.GET("/bundle", s.getBundle)
		pub.POST("/quotation", s.limit("forms", 20, 20), bodyLimit(64<<10), s.postQuotation)
		pub.POST("/application", s.limit("forms", 20, 20), bodyLimit(21<<20), s.postApplication)

		au := api.Group("/auth", s.limit("auth", 10, 10))
		au.GET("/setup", s.getSetup)
		au.POST("/setup", bodyLimit(16<<10), s.postSetup)
		au.POST("/login", bodyLimit(16<<10), s.postLogin)
		au.POST("/refresh", s.postRefresh)
		au.POST("/logout", s.postLogout)

		me := api.Group("/me", s.requireAuth())
		me.GET("", s.getMe)
		me.POST("/password", bodyLimit(16<<10), s.postPassword)

		// Admin routes: a separate group, authenticated, every route
		// declaring its permission (CLAUDE.md §7).
		ad := api.Group("/admin", s.requireAuth(), bodyLimit(1<<20))
		ad.GET("/schema", s.require(access.ContentRead, access.ParamsContent), s.getSchema)
		ad.GET("/options", s.require(access.ContentRead, access.ParamsContent), s.getOptions)
		ad.GET("/placeholders", s.require(access.ContentRead), s.getPlaceholders)
		ad.GET("/r/:name", s.require(access.ContentRead, access.ParamsContent), s.listResource)
		ad.GET("/r/:name/:id", s.require(access.ContentRead, access.ParamsContent), s.getResource)
		ad.POST("/r/:name", s.require(access.ContentWrite, access.ParamsContent), s.createResource)
		ad.PUT("/r/:name/:id", s.require(access.ContentWrite, access.ParamsContent), s.updateResource)
		ad.DELETE("/r/:name/:id", s.require(access.ContentWrite, access.ParamsContent), s.deleteResource)

		ad.GET("/media", s.require(access.ContentRead), s.listMedia)
		ad.PUT("/media/:id", s.require(access.MediaWrite), s.updateMedia)
		ad.DELETE("/media/:id", s.require(access.MediaWrite), s.deleteMedia)

		ad.GET("/quotations", s.require(access.QuotationsRead), s.listQuotations)
		ad.GET("/quotations/:id", s.require(access.QuotationsRead), s.getQuotation)
		ad.GET("/applications", s.require(access.ApplicationsRead), s.listApplications)
		ad.GET("/applications/:id", s.require(access.ApplicationsRead), s.getApplication)
		ad.POST("/applications/:id/cv", s.require(access.ApplicationsRead), s.cvLink)

		ad.GET("/users", s.require(access.UsersManage), s.listUsers)
		ad.POST("/users", s.require(access.UsersManage), s.createUser)
		ad.PUT("/users/:id", s.require(access.UsersManage), s.updateUser)
		ad.POST("/users/:id/password", s.require(access.UsersManage), s.resetPassword)
		ad.GET("/audit", s.require(access.AuditRead), s.listAudit)
		ad.GET("/outbox", s.require(access.AuditRead), s.listOutbox)
	}
	// Uploads have their own, larger body limit.
	up := r.Group("/api/v1/admin", noStore(), s.requireAuth(), bodyLimit(52<<20))
	up.POST("/media", s.require(access.MediaWrite), s.uploadMedia)

	r.GET("/media/:id/:file", s.getMediaFile)
	r.GET("/robots.txt", s.robots)
	r.GET("/sitemap.xml", s.sitemap)
	if d.Web != nil {
		r.NoRoute(s.spa)
	}
	return s, r, nil
}

func noStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
