// Package test holds the integration, authorisation and security suites.
// They run against a real PostgreSQL (SFG_TEST_DATABASE_URL) with the real
// migrations and seed, the real HTTP router, and in-memory fakes only for
// the object store and the mail relay.
//
//	make test    (sources /etc/sfg/sfg.env)
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/db"
	httpapi "github.com/stevenwilliam/sunshinefacility_compro/internal/adapter/http"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/adapter/media"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/adapter/postgres"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/database"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/security"
	"github.com/stevenwilliam/sunshinefacility_compro/web"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type memStorage struct {
	mu   sync.Mutex
	objs map[string][]byte
}

func (m *memStorage) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.objs[key] = b
	m.mu.Unlock()
	return nil
}
func (m *memStorage) Get(_ context.Context, key string) (io.ReadCloser, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.objs[key]
	if !ok {
		return nil, 0, app.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(b)), int64(len(b)), nil
}
func (m *memStorage) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.objs, key)
	m.mu.Unlock()
	return nil
}
func (m *memStorage) PresignGet(_ context.Context, key string, _ time.Duration, _ string) (string, error) {
	return "https://storage.invalid/" + key + "?sig=test", nil
}
func (m *memStorage) has(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.objs[key]
	return ok
}

type memMailer struct {
	mu   sync.Mutex
	sent []app.Mail
	fail bool
}

func (m *memMailer) Send(_ context.Context, mail app.Mail) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail {
		return io.ErrUnexpectedEOF
	}
	m.sent = append(m.sent, mail)
	return nil
}

// ---------------------------------------------------------------------------
// Environment
// ---------------------------------------------------------------------------

type env struct {
	t        *testing.T
	gdb      *gorm.DB
	store    *postgres.Store
	storage  *memStorage
	mailer   *memMailer
	srv      *httptest.Server
	auth     *app.Auth
	jobs     *app.Jobs
	public   *app.Public
	settings *app.Settings
	now      time.Time
}

const testSecret = "test-secret-test-secret-test-secret-1234"
const setupToken = "setup-token-for-tests"

// fresh drops the schema, migrates and seeds, and starts a server. Each test
// gets its own clean database: order dependence is a defect.
func fresh(t *testing.T) *env {
	t.Helper()
	dsn := os.Getenv("SFG_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("SFG_TEST_DATABASE_URL not set")
	}
	gdb, err := database.Open(dsn, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`).Error; err != nil {
		t.Fatal(err)
	}
	ms, err := database.Load(db.Migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Up(gdb, ms); err != nil {
		t.Fatal(err)
	}
	e := &env{t: t, gdb: gdb, store: postgres.New(gdb), storage: &memStorage{objs: map[string][]byte{}}, mailer: &memMailer{},
		// The real clock: tokens are verified against it. Tests that need a
		// date set it relative to now.
		now: time.Now().UTC().Truncate(time.Second)}
	clock := func() time.Time { return e.now }
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	e.settings = app.NewSettings(e.store)
	e.public = app.NewPublic(e.store, "https://www.sunshinefacility.com", clock)
	e.auth = app.NewAuth(e.store, e.settings, security.NewTokenIssuer(testSecret, 15*time.Minute), 7*24*time.Hour, setupToken, clock)
	content := app.NewContent(e.store, e.public, e.settings, clock)
	med := app.NewMedia(e.store, e.storage, media.WebP{}, e.settings, e.public)
	subs := app.NewSubmissions(e.store, e.storage, e.settings, e.public, "https://www.sunshinefacility.com/admin", clock)
	e.jobs = app.NewJobs(e.store, e.storage, e.mailer, e.settings, "https://www.sunshinefacility.com/admin", log, clock)
	dist, err := web.Dist()
	if err != nil {
		t.Fatal(err)
	}
	_, h, err := httpapi.New(httpapi.Deps{Log: log, BaseURL: "https://www.sunshinefacility.com", Auth: e.auth, Content: content,
		Media: med, Public: e.public, Submissions: subs, Audit: app.NewAuditLog(e.store),
		Ready: func(context.Context) error { return nil }, Web: dist})
	if err != nil {
		t.Fatal(err)
	}
	e.srv = httptest.NewServer(h)
	t.Cleanup(func() {
		e.srv.Close()
		if s, err := gdb.DB(); err == nil {
			_ = s.Close()
		}
	})
	return e
}

// user creates an account with a role and returns a signed-in client.
func (e *env) user(role string) *client {
	e.t.Helper()
	email := role + "@sfg.test"
	ctx := context.Background()
	if _, err := e.auth.CreateUser(ctx, app.Actor{Email: "test"}, email, strings.ToUpper(role[:1])+role[1:], role, "correct horse battery"); err != nil {
		e.t.Fatal(err)
	}
	c := &client{e: e}
	var r struct{ AccessToken string }
	c.do("POST", "/api/v1/auth/login", map[string]string{"email": email, "password": "correct horse battery"}, &r, 200)
	c.token = r.AccessToken
	return c
}

func (e *env) anon() *client { return &client{e: e} }

type client struct {
	e     *env
	token string
}

// do sends JSON and decodes JSON. wantStatus 0 means "don't check".
func (c *client) do(method, path string, body any, out any, wantStatus int) *http.Response {
	c.e.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.e.srv.URL+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := c.e.srv.Client().Do(req)
	if err != nil {
		c.e.t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if wantStatus != 0 && res.StatusCode != wantStatus {
		c.e.t.Fatalf("%s %s: status %d, want %d; body %s", method, path, res.StatusCode, wantStatus, truncate(string(raw)))
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			if s, ok := out.(*string); ok {
				*s = string(raw)
			} else {
				c.e.t.Fatalf("%s %s: decode: %v; body %s", method, path, err, truncate(string(raw)))
			}
		}
	}
	res.Body = io.NopCloser(bytes.NewReader(raw))
	return res
}

func (c *client) raw(method, path string) (int, http.Header, string) {
	c.e.t.Helper()
	req, _ := http.NewRequest(method, c.e.srv.URL+path, nil)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	cl := *c.e.srv.Client()
	cl.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	res, err := cl.Do(req)
	if err != nil {
		c.e.t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return res.StatusCode, res.Header, string(b)
}

func (c *client) multipart(path string, fields map[string]string, fileField, fileName string, file []byte) (int, string) {
	c.e.t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	if fileField != "" {
		fw, _ := w.CreateFormFile(fileField, fileName)
		_, _ = fw.Write(file)
	}
	_ = w.Close()
	req, _ := http.NewRequest("POST", c.e.srv.URL+path, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := c.e.srv.Client().Do(req)
	if err != nil {
		c.e.t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	return res.StatusCode, string(raw)
}

func truncate(s string) string {
	if len(s) > 600 {
		return s[:600] + "…"
	}
	return s
}

func (e *env) count(query string, args ...any) int {
	e.t.Helper()
	var n int
	if err := e.gdb.Raw(query, args...).Row().Scan(&n); err != nil {
		e.t.Fatal(err)
	}
	return n
}

func (e *env) id(query string, args ...any) string {
	e.t.Helper()
	var s string
	if err := e.gdb.Raw(query, args...).Row().Scan(&s); err != nil {
		e.t.Fatal(err)
	}
	return s
}

// A valid quotation body.
func quote(email string) map[string]any {
	return map[string]any{"name": "Budi Santoso", "company": "PT Contoh", "email": email, "phone": "+62 812 3456 7890",
		"buildingType": "office", "area": "5000", "location": "Tangerang", "services": []string{"facility"},
		"message": "Test", "lang": "en", "elapsedMs": 9000}
}
