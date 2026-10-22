// Command sfg is the whole service: HTTP server, migrations, seeding and the
// scheduled jobs. Wiring only — no business logic lives here.
//
//	sfg serve
//	sfg migrate up|down|status
//	sfg seed-assets <design/assets dir>
//	sfg seed-dev                      development admin (refused in production)
//	sfg reset-password <email>        break-glass; new password on stdin
//	sfg job deliver-mail|certificate-reminders|purge-applications
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/stevenwilliam/sunshinefacility_compro/db"
	httpapi "github.com/stevenwilliam/sunshinefacility_compro/internal/adapter/http"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/adapter/media"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/adapter/notify"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/adapter/postgres"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/adapter/storage"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/config"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/database"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/logging"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/security"
	"github.com/stevenwilliam/sunshinefacility_compro/web"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "sfg:", err)
		os.Exit(1)
	}
}

type services struct {
	cfg         config.Config
	log         *slog.Logger
	store       *postgres.Store
	storage     *storage.MinIO
	settings    *app.Settings
	public      *app.Public
	auth        *app.Auth
	content     *app.Content
	media       *app.Media
	submissions *app.Submissions
	jobs        *app.Jobs
	audit       *app.AuditLog
	ping        func(context.Context) error
}

func wire(cfg config.Config, log *slog.Logger) (*services, error) {
	gdb, err := database.Open(cfg.DatabaseURL, false)
	if err != nil {
		return nil, err
	}
	sqlDB, _ := gdb.DB()
	st, err := storage.NewMinIO(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3UseSSL)
	if err != nil {
		return nil, err
	}
	store := postgres.New(gdb)
	now := time.Now
	adminURL := cfg.PublicBaseURL + "/admin"
	settings := app.NewSettings(store)
	public := app.NewPublic(store, cfg.PublicBaseURL, now)
	var mailer app.Mailer = notify.SMTP{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Password: cfg.SMTPPassword}
	if !cfg.NotifyEnabled {
		mailer = notify.Log{L: log}
	}
	s := &services{
		cfg: cfg, log: log, store: store, storage: st, settings: settings, public: public,
		auth:        app.NewAuth(store, settings, security.NewTokenIssuer(cfg.JWTSecret, cfg.AccessTokenTTL), cfg.RefreshTokenTTL, cfg.SetupToken, now),
		content:     app.NewContent(store, public, settings, now),
		media:       app.NewMedia(store, st, media.WebP{Quality: 82}, settings, public),
		submissions: app.NewSubmissions(store, st, settings, public, adminURL, now),
		jobs:        app.NewJobs(store, st, mailer, settings, adminURL, log, now),
		audit:       app.NewAuditLog(store),
	}
	s.ping = func(ctx context.Context) error {
		if err := sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("database: %w", err)
		}
		return st.Ping(ctx)
	}
	return s, nil
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: sfg serve | migrate up|down|status | seed-assets <dir> | seed-dev | job <name>")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logging.New(cfg.LogLevel)

	switch args[0] {
	case "migrate":
		return migrate(cfg, args[1:])
	case "serve":
		return serve(cfg, log)
	case "seed-assets":
		if len(args) < 2 {
			return errors.New("usage: sfg seed-assets <path to design/assets>")
		}
		s, err := wire(cfg, log)
		if err != nil {
			return err
		}
		ctx := context.Background()
		if err := s.storage.EnsureBucket(ctx); err != nil {
			return err
		}
		return s.media.SeedAssets(ctx, args[1], func(m string) { fmt.Println(m) })
	case "seed-dev":
		s, err := wire(cfg, log)
		if err != nil {
			return err
		}
		email := envOr("SFG_DEV_ADMIN_EMAIL", "admin@sfg.local")
		pw := os.Getenv("SFG_DEV_ADMIN_PASSWORD")
		if pw == "" {
			return errors.New("SFG_DEV_ADMIN_PASSWORD is not set")
		}
		created, err := s.auth.SeedDevAdmin(context.Background(), cfg.Env, email, pw)
		if err != nil {
			return err
		}
		if created {
			fmt.Println("seed-dev: created", email)
		} else {
			fmt.Println("seed-dev:", email, "already exists; unchanged")
		}
		return nil
	case "reset-password":
		// Break-glass: sfg reset-password <email>, new password on stdin.
		if len(args) < 2 {
			return errors.New("usage: echo 'new password' | sfg reset-password <email>")
		}
		s, err := wire(cfg, log)
		if err != nil {
			return err
		}
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return errors.New("read the new password from stdin")
		}
		if err := s.auth.BreakGlassReset(context.Background(), args[1], strings.TrimRight(line, "\r\n")); err != nil {
			return err
		}
		fmt.Println("reset-password: done; the account is unlocked and active, its sessions ended")
		return nil
	case "job":
		if len(args) < 2 {
			return errors.New("usage: sfg job deliver-mail|certificate-reminders|purge-applications")
		}
		s, err := wire(cfg, log)
		if err != nil {
			return err
		}
		return job(s, args[1])
	}
	return fmt.Errorf("unknown command %q", args[0])
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func migrate(cfg config.Config, args []string) error {
	gdb, err := database.Open(cfg.DatabaseURL, false)
	if err != nil {
		return err
	}
	ms, err := database.Load(db.Migrations, "migrations")
	if err != nil {
		return err
	}
	sub := "up"
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "up":
		done, err := database.Up(gdb, ms)
		for _, m := range done {
			fmt.Printf("applied %04d_%s\n", m.Version, m.Name)
		}
		if err == nil && len(done) == 0 {
			fmt.Println("up to date")
		}
		return err
	case "down":
		if cfg.IsProduction() {
			return errors.New("migrate down is refused in production: migrations are forward-only")
		}
		m, err := database.Down(gdb, ms)
		if m != nil {
			fmt.Printf("rolled back %04d_%s\n", m.Version, m.Name)
		}
		return err
	case "status":
		applied, pending, err := database.Status(gdb, ms)
		if err != nil {
			return err
		}
		for _, a := range applied {
			fmt.Printf("applied  %04d_%s  %s\n", a.Version, a.Name, a.AppliedAt.Format(time.RFC3339))
		}
		for _, p := range pending {
			fmt.Printf("pending  %04d_%s\n", p.Version, p.Name)
		}
		return nil
	}
	return fmt.Errorf("unknown migrate command %q", sub)
}

func job(s *services, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	switch name {
	case "deliver-mail":
		sent, failed, err := s.jobs.DeliverMail(ctx)
		fmt.Printf("deliver-mail: %d sent, %d failed\n", sent, failed)
		return err
	case "certificate-reminders":
		n, err := s.jobs.CertificateReminders(ctx)
		fmt.Printf("certificate-reminders: %d queued\n", n)
		if err != nil {
			return err
		}
		_, _, err = s.jobs.DeliverMail(ctx)
		return err
	case "purge-applications":
		n, err := s.jobs.PurgeApplications(ctx)
		fmt.Printf("purge-applications: %d deleted\n", n)
		return err
	}
	return fmt.Errorf("unknown job %q", name)
}

func serve(cfg config.Config, log *slog.Logger) error {
	s, err := wire(cfg, log)
	if err != nil {
		return err
	}
	// Refuse to serve on a schema that is behind or edited (platform/database).
	gdb, err := database.Open(cfg.DatabaseURL, false)
	if err != nil {
		return err
	}
	ms, err := database.Load(db.Migrations, "migrations")
	if err != nil {
		return err
	}
	_, pending, err := database.Status(gdb, ms)
	if err != nil {
		return err
	}
	if len(pending) > 0 {
		return fmt.Errorf("%d migration(s) pending — run `sfg migrate up` before serving", len(pending))
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := s.storage.EnsureBucket(ctx); err != nil {
		return err
	}
	dist, err := web.Dist()
	if err != nil {
		return err
	}
	_, handler, err := httpapi.New(httpapi.Deps{
		Log: log, Production: cfg.IsProduction(), BaseURL: cfg.PublicBaseURL,
		Auth: s.auth, Content: s.content, Media: s.media, Public: s.public, Submissions: s.submissions, Audit: s.audit,
		Ready: s.ping, Web: dist, Proxies: cfg.TrustedProxies,
	})
	if err != nil {
		return err
	}
	go s.jobs.RunWorker(ctx, 30*time.Second)

	srv := &nethttp.Server{Addr: cfg.HTTPAddr, Handler: handler, ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 120 * time.Second}
	log.Info("serving", "config", cfg.Redacted())
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	shut, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shut)
}
