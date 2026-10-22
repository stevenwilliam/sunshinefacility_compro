// Package config loads configuration from the environment only. Nothing
// secret is in git; the documented surface is .env.example and the real file
// lives at /etc/sfg/sfg.env.
//
// Only what is needed to BOOT lives here. Everything the business may want to
// change without a deploy — recipients, retention, thresholds, contact
// details — is a row in `parameters` (CLAUDE.md §7).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env             string
	HTTPAddr        string
	PublicBaseURL   string
	DatabaseURL     string
	TestDatabaseURL string

	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	// SetupToken gates the first-run setup screen. With no users in the
	// database and no token configured, setup is refused: a fresh install
	// must never be claimable by whoever reaches it first.
	SetupToken string

	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPassword  string
	NotifyEnabled bool

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool

	LogLevel       string
	MetricsEnabled bool
	TrustedProxies []string
}

func (c Config) IsProduction() bool { return c.Env == "production" }

// Redacted is what may be printed at boot. A startup line that prints the
// database URL leaks the password into the journal, where it stays.
func (c Config) Redacted() map[string]string {
	return map[string]string{
		"env":         c.Env,
		"http_addr":   c.HTTPAddr,
		"public_base": c.PublicBaseURL,
		"database":    redactURL(c.DatabaseURL),
		"smtp":        fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort),
		"s3":          c.S3Endpoint + "/" + c.S3Bucket,
		"jwt_secret":  mask(c.JWTSecret),
		"setup_token": mask(c.SetupToken),
		"log_level":   c.LogLevel,
	}
}

func mask(s string) string {
	if s == "" {
		return "(unset)"
	}
	return "••••••"
}

func redactURL(u string) string {
	at := strings.LastIndex(u, "@")
	if at < 0 {
		return u
	}
	scheme := strings.Index(u, "://")
	if scheme < 0 {
		return u
	}
	creds := u[scheme+3 : at]
	if colon := strings.Index(creds, ":"); colon >= 0 {
		creds = creds[:colon] + ":••••••"
	}
	return u[:scheme+3] + creds + u[at:]
}

// Load reads the environment and refuses to start on a missing required
// value. A service that boots with an empty JWT secret is worse than one that
// does not boot.
func Load() (Config, error) {
	c := Config{
		Env:             env("SFG_ENV", "development"),
		HTTPAddr:        env("SFG_HTTP_ADDR", "127.0.0.1:8098"),
		PublicBaseURL:   strings.TrimRight(env("SFG_PUBLIC_BASE_URL", "https://sunshinefacility.sunshinefood.co.id"), "/"),
		DatabaseURL:     os.Getenv("SFG_DATABASE_URL"),
		TestDatabaseURL: os.Getenv("SFG_TEST_DATABASE_URL"),
		JWTSecret:       os.Getenv("SFG_JWT_SECRET"),
		SetupToken:      os.Getenv("SFG_SETUP_TOKEN"),
		SMTPHost:        env("SFG_SMTP_HOST", "127.0.0.1"),
		SMTPUser:        os.Getenv("SFG_SMTP_USER"),
		SMTPPassword:    os.Getenv("SFG_SMTP_PASSWORD"),
		S3Endpoint:      os.Getenv("SFG_S3_ENDPOINT"),
		S3AccessKey:     os.Getenv("SFG_S3_ACCESS_KEY"),
		S3SecretKey:     os.Getenv("SFG_S3_SECRET_KEY"),
		S3Bucket:        env("SFG_S3_BUCKET", "sfg"),
		LogLevel:        env("SFG_LOG_LEVEL", "info"),
	}
	c.SMTPPort = envInt("SFG_SMTP_PORT", 1025)
	c.AccessTokenTTL = envDuration("SFG_ACCESS_TOKEN_TTL", 15*time.Minute)
	c.RefreshTokenTTL = envDuration("SFG_REFRESH_TOKEN_TTL", 168*time.Hour)
	c.NotifyEnabled = envBool("SFG_NOTIFY_ENABLED", true)
	c.S3UseSSL = envBool("SFG_S3_USE_SSL", false)
	c.MetricsEnabled = envBool("SFG_METRICS_ENABLED", true)
	if p := os.Getenv("SFG_TRUSTED_PROXIES"); p != "" {
		c.TrustedProxies = strings.Split(p, ",")
	}

	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "SFG_DATABASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "SFG_JWT_SECRET")
	}
	if c.S3Endpoint == "" {
		missing = append(missing, "SFG_S3_ENDPOINT")
	}
	if len(missing) > 0 {
		return c, fmt.Errorf("required configuration missing: %s", strings.Join(missing, ", "))
	}
	if len(c.JWTSecret) < 32 {
		return c, fmt.Errorf("SFG_JWT_SECRET must be at least 32 characters, it is %d", len(c.JWTSecret))
	}
	return c, nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
