package test

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"strings"
	"testing"
)

// BR-8.1 — negative authorisation, per role, per route. Every admin route is
// called by every role; the expected status is spelled out, so a route added
// without a permission fails here.
func TestBR8_1_AuthorisationMatrix(t *testing.T) {
	e := fresh(t)
	roles := map[string]*client{"superadmin": e.user("superadmin"), "editor": e.user("editor"), "sales": e.user("sales"), "hr": e.user("hr")}
	anon := e.anon()
	cert := e.id(`SELECT id::text FROM certificates WHERE name = 'SLHS'`)
	type route struct {
		method, path string
		allowed      []string
	}
	routes := []route{
		{"GET", "/api/v1/admin/r/clients", []string{"superadmin", "editor"}},
		{"GET", "/api/v1/admin/r/certificates/" + cert, []string{"superadmin", "editor"}},
		{"GET", "/api/v1/admin/r/parameters", []string{"superadmin", "editor"}},
		{"GET", "/api/v1/admin/media", []string{"superadmin", "editor"}},
		{"GET", "/api/v1/admin/schema", []string{"superadmin", "editor"}},
		{"GET", "/api/v1/admin/quotations", []string{"superadmin", "sales"}},
		{"GET", "/api/v1/admin/applications", []string{"superadmin", "hr"}},
		{"GET", "/api/v1/admin/users", []string{"superadmin"}},
		{"GET", "/api/v1/admin/audit", []string{"superadmin"}},
		{"GET", "/api/v1/admin/outbox", []string{"superadmin"}},
	}
	for _, r := range routes {
		allowed := map[string]bool{}
		for _, a := range r.allowed {
			allowed[a] = true
		}
		for name, c := range roles {
			want := 403
			if allowed[name] {
				want = 200
			}
			c.do(r.method, r.path, nil, nil, want)
		}
		anon.do(r.method, r.path, nil, nil, 401)
	}
	// Writes: sales and hr cannot change content.
	for _, name := range []string{"sales", "hr"} {
		roles[name].do("PUT", "/api/v1/admin/r/certificates/"+cert, map[string]any{"name": "x"}, nil, 403)
	}
}

// IDOR: an id from another inbox is not readable by the wrong role, and a
// made-up id is a 404, not a 500.
func TestIDORInbox(t *testing.T) {
	e := fresh(t)
	e.anon().do("POST", "/api/v1/public/quotation", quote("idor@contoh.co.id"), nil, 200)
	qid := e.id(`SELECT id::text FROM quotation_requests LIMIT 1`)
	e.user("hr").do("GET", "/api/v1/admin/quotations/"+qid, nil, nil, 403)
	sales := e.user("sales")
	sales.do("GET", "/api/v1/admin/quotations/"+qid, nil, nil, 200)
	sales.do("GET", "/api/v1/admin/quotations/00000000-0000-7000-8000-000000000000", nil, nil, 404)
	sales.do("GET", "/api/v1/admin/quotations/not-a-uuid", nil, nil, 404)
	sales.do("POST", "/api/v1/admin/applications/"+qid+"/cv", nil, nil, 403)
}

// BR-8.3 — identical answers for unknown email and wrong password; lockout.
func TestBR8_3_LoginGenericAndLockout(t *testing.T) {
	e := fresh(t)
	e.user("editor")
	var a, b struct {
		Error struct{ Code, Message string }
	}
	e.anon().do("POST", "/api/v1/auth/login", map[string]string{"email": "nobody@sfg.test", "password": "whatever-whatever"}, &a, 401)
	e.anon().do("POST", "/api/v1/auth/login", map[string]string{"email": "editor@sfg.test", "password": "wrong-password!"}, &b, 401)
	if a.Error != b.Error {
		t.Fatalf("unknown email and wrong password differ: %+v vs %+v", a, b)
	}
	for i := 0; i < 4; i++ {
		e.anon().do("POST", "/api/v1/auth/login", map[string]string{"email": "editor@sfg.test", "password": "wrong-password!"}, nil, 0)
	}
	// Locked now: even the right password is refused, with the same message.
	var c struct {
		Error struct{ Code, Message string }
	}
	e.anon().do("POST", "/api/v1/auth/login", map[string]string{"email": "editor@sfg.test", "password": "correct horse battery"}, &c, 0)
	if c.Error.Message != a.Error.Message {
		t.Fatalf("locked account answered differently: %+v", c)
	}
	if e.count(`SELECT count(*) FROM users WHERE email = 'editor@sfg.test' AND locked_until > now()`) != 1 {
		t.Fatal("account not locked after repeated failures")
	}
}

// JWT tampering and logout.
func TestJWTTamperingAndLogout(t *testing.T) {
	e := fresh(t)
	su := e.user("superadmin")
	parts := strings.Split(su.token, ".")
	// alg:none
	none := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`)) + "." + parts[1] + "."
	(&client{e: e, token: none}).do("GET", "/api/v1/me", nil, nil, 401)
	// flipped signature
	sig := []byte(parts[2])
	sig[0] ^= 1
	(&client{e: e, token: parts[0] + "." + parts[1] + "." + string(sig)}).do("GET", "/api/v1/me", nil, nil, 401)
	su.do("GET", "/api/v1/me", nil, nil, 200)
	su.do("POST", "/api/v1/auth/logout", nil, nil, 200)
	su.do("GET", "/api/v1/me", nil, nil, 401) // the jti is revoked
}

// BR-8.4 — the last superadmin cannot be demoted, nor demote themself.
func TestBR8_4_LastSuperadmin(t *testing.T) {
	e := fresh(t)
	su := e.user("superadmin")
	me := e.id(`SELECT id::text FROM users WHERE email = 'superadmin@sfg.test'`)
	su.do("PUT", "/api/v1/admin/users/"+me, map[string]any{"name": "S", "role": "editor", "active": true}, nil, 409)
	su.do("PUT", "/api/v1/admin/users/"+me, map[string]any{"name": "S", "role": "superadmin", "active": false}, nil, 409)
}

// BR-8.2 — setup works once, and only with the token.
func TestBR8_2_SetupOnce(t *testing.T) {
	e := fresh(t)
	body := map[string]string{"token": "wrong", "email": "first@sfg.test", "name": "First", "password": "a long enough password"}
	e.anon().do("POST", "/api/v1/auth/setup", body, nil, 422)
	body["token"] = setupToken
	e.anon().do("POST", "/api/v1/auth/setup", body, nil, 200)
	body["email"] = "second@sfg.test"
	e.anon().do("POST", "/api/v1/auth/setup", body, nil, 404)
}

// BR-5.1 / BR-5.4 — the server re-checks everything and keeps bots out.
func TestBR5_FormAbuse(t *testing.T) {
	e := fresh(t)
	a := e.anon()
	// Forged keys and injection-shaped values are refused field by field.
	bad := quote("x@y.co")
	bad["buildingType"] = "'; DROP TABLE users; --"
	bad["services"] = []string{"facility", "<script>"}
	bad["area"] = "5,000"
	var out struct {
		Error struct{ Fields map[string]string }
	}
	a.do("POST", "/api/v1/public/quotation", bad, &out, 422)
	if out.Error.Fields["buildingType"] != "not_allowed" || out.Error.Fields["services"] != "not_allowed" || out.Error.Fields["area"] != "digits" {
		t.Fatalf("fields: %+v", out.Error.Fields)
	}
	// Honeypot: accepted at the door, stored nowhere.
	hp := quote("bot@spam.example")
	hp["website"] = "http://spam.example"
	a.do("POST", "/api/v1/public/quotation", hp, nil, 200)
	// Too fast: refused visibly.
	fast := quote("fast@contoh.co.id")
	fast["elapsedMs"] = 200
	a.do("POST", "/api/v1/public/quotation", fast, nil, 422)
	if n := e.count(`SELECT count(*) FROM quotation_requests`); n != 0 {
		t.Fatalf("%d bot/invalid submissions stored", n)
	}
	// Per-email limit (parameter: 5 an hour).
	for i := 0; i < 5; i++ {
		a.do("POST", "/api/v1/public/quotation", quote("same@contoh.co.id"), nil, 200)
	}
	a.do("POST", "/api/v1/public/quotation", quote("same@contoh.co.id"), nil, 429)
	if n := e.count(`SELECT count(*) FROM notification_outbox WHERE kind = 'quotation'`); n != 5 {
		t.Fatalf("outbox has %d quotation mails, want 5", n)
	}
}

// BR-6.1 — a CV is identified by its bytes; size is capped.
func TestBR6_1_CVChecks(t *testing.T) {
	e := fresh(t)
	a := e.anon()
	fields := map[string]string{"name": "Sari", "phone": "0812-3456-7890", "email": "sari@example.com", "position": "open", "lang": "id", "elapsedMs": "8000"}
	st, body := a.multipart("/api/v1/public/application", fields, "cv", "cv.pdf", []byte("<html><script>alert(1)</script>"))
	if st != 422 || !strings.Contains(body, "file_type") {
		t.Fatalf("html-as-pdf: %d %s", st, body)
	}
	big := append([]byte("%PDF-1.7\n"), make([]byte, 6<<20)...)
	if st, body = a.multipart("/api/v1/public/application", fields, "cv", "big.pdf", big); st != 422 || !strings.Contains(body, "file_size") {
		t.Fatalf("6 MB CV: %d %s", st, body)
	}
	if st, body = a.multipart("/api/v1/public/application", fields, "cv", "../../etc/passwd.pdf", []byte("%PDF-1.4 ok")); st != 200 {
		t.Fatalf("good CV: %d %s", st, body)
	}
	key := e.id(`SELECT cv_key FROM career_applications LIMIT 1`)
	name := e.id(`SELECT cv_filename FROM career_applications LIMIT 1`)
	if !strings.HasPrefix(key, "cv/") || !strings.HasSuffix(key, ".pdf") || strings.Contains(name, "/") || strings.Contains(name, "..") {
		t.Fatalf("stored key %q name %q", key, name)
	}
	if !e.storage.has(key) {
		t.Fatal("the CV was not stored")
	}
}

// BR-9.2 — pipe-delimited, formula-guarded, honours the search on screen.
func TestBR9_2_CSVExport(t *testing.T) {
	e := fresh(t)
	ed := e.user("editor")
	id := e.id(`SELECT id::text FROM clients WHERE name = 'Cargill'`)
	ed.do("PUT", "/api/v1/admin/r/clients/"+id, map[string]any{"name": "=HYPERLINK(\"http://evil\")"}, nil, 200)
	st, h, body := ed.raw("GET", "/api/v1/admin/r/clients?q=HYPERLINK&format=csv")
	if st != 200 || !strings.HasPrefix(h.Get("Content-Type"), "text/csv") {
		t.Fatalf("status %d type %s", st, h.Get("Content-Type"))
	}
	r := csv.NewReader(strings.NewReader(strings.TrimPrefix(body, "\xef\xbb\xbf")))
	r.Comma = '|'
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("not a pipe CSV: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("export ignored the search: %d rows", len(rows))
	}
	found := false
	for _, c := range rows[1] {
		if strings.HasPrefix(c, "'=HYPERLINK") {
			found = true
		}
		if strings.HasPrefix(c, "=") {
			t.Fatalf("unguarded formula cell %q", c)
		}
	}
	if !found {
		t.Fatalf("guarded cell missing: %v", rows[1])
	}
}

// BR-4.1 / BR-4.4 — uploads by magic bytes; SVG refused; re-encoded to WebP.
func TestBR4_1_Uploads(t *testing.T) {
	e := fresh(t)
	ed := e.user("editor")
	st, body := ed.multipart("/api/v1/admin/media", map[string]string{"kind": "photo"}, "file", "x.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
	if st != 415 {
		t.Fatalf("svg upload: %d %s", st, body)
	}
	img := image.NewNRGBA(image.Rect(0, 0, 600, 400))
	for x := 0; x < 600; x++ {
		img.Set(x, 10, color.NRGBA{200, 30, 20, 255})
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	st, body = ed.multipart("/api/v1/admin/media", map[string]string{"kind": "photo"}, "file", "pic.jpg", b.Bytes())
	if st != 201 || !strings.Contains(body, `"mime":"image/png"`) {
		t.Fatalf("png named .jpg: %d %s", st, body)
	}
	if e.count(`SELECT count(*) FROM media WHERE name = 'pic.jpg' AND jsonb_array_length(variants) = 2`) != 1 {
		t.Fatal("renditions not recorded (want 480 and 600)")
	}
	e.user("sales").multipart("/api/v1/admin/media", map[string]string{"kind": "photo"}, "file", "p.png", b.Bytes())
}

// Security headers on every response (CLAUDE.md §7).
func TestSecurityHeaders(t *testing.T) {
	e := fresh(t)
	_, h, _ := e.anon().raw("GET", "/")
	csp := h.Get("Content-Security-Policy")
	if csp == "" || strings.Contains(csp, "unsafe-inline") || !strings.Contains(csp, "frame-ancestors 'none'") {
		t.Fatalf("CSP: %q", csp)
	}
	for _, k := range []string{"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Permissions-Policy"} {
		if h.Get(k) == "" {
			t.Errorf("missing %s", k)
		}
	}
	_, h, _ = e.anon().raw("GET", "/api/v1/public/bundle")
	if h.Get("ETag") == "" {
		t.Error("bundle has no ETag")
	}
	_ = http.StatusOK
}
