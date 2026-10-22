package test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/resource"
)

// CLAUDE.md §13 — what a link-preview bot sees, per route, without JS.
func TestSEOHeadPerRoute(t *testing.T) {
	e := fresh(t)
	a := e.anon()
	cases := map[string]string{
		"/":              "<title>Sunshine Facilities Group · Integrated facility services in Indonesia</title>",
		"/services":      "<title>Services · Sunshine Facilities Group</title>",
		"/services/food": "<title>Food Services · Sunshine Facilities Group</title>",
		"/about?lang=id": "<title>Tentang Kami · Sunshine Facilities Group</title>",
		"/contact":       `<link rel="canonical" href="https://www.sunshinefacility.com/contact">`,
	}
	for path, want := range cases {
		st, _, body := a.raw("GET", path)
		if st != 200 || !strings.Contains(body, want) {
			t.Errorf("%s: status %d, missing %s", path, st, want)
		}
		if !strings.Contains(body, `property="og:title"`) || !strings.Contains(body, `application/ld+json`) || !strings.Contains(body, `id="sfg-bundle"`) {
			t.Errorf("%s: missing OG, JSON-LD or bundle", path)
		}
	}
	if _, _, body := a.raw("GET", "/about?lang=id"); !strings.Contains(body, `<html lang="id">`) {
		t.Error("?lang=id did not set <html lang>")
	}
	// Placeholder contact details stay out of structured data.
	if _, _, body := a.raw("GET", "/"); strings.Contains(body, `"telephone"`) {
		t.Error("a placeholder phone number is in the JSON-LD")
	}
}

func TestNotFoundAndRedirects(t *testing.T) {
	e := fresh(t)
	a := e.anon()
	if st, _, body := a.raw("GET", "/no-such-page"); st != 404 || !strings.Contains(body, "noindex") {
		t.Errorf("unknown route: %d", st)
	}
	if st, _, _ := a.raw("GET", "/services/spaceships"); st != 404 {
		t.Errorf("unknown line: %d", st)
	}
	if st, h, _ := a.raw("GET", "/services?line=remote"); st != 301 || h.Get("Location") != "/services/remote" {
		t.Errorf("?line redirect: %d %s", st, h.Get("Location"))
	}
	if st, _, body := a.raw("GET", "/tytb/r/clients"); st != 200 || !strings.Contains(body, "noindex") {
		t.Errorf("admin: %d", st)
	}
	if st, _, body := a.raw("GET", "/admin"); st != 404 || strings.Contains(body, "Admin ·") {
		t.Errorf("/admin must no longer serve the CMS: %d", st)
	}
	_, _, robots := a.raw("GET", "/robots.txt")
	if strings.Contains(robots, "tytb") || strings.Contains(robots, "/admin") || !strings.Contains(robots, "Sitemap: https://www.sunshinefacility.com/sitemap.xml") {
		t.Errorf("robots: %s", robots)
	}
	_, _, sm := a.raw("GET", "/sitemap.xml")
	if strings.Count(sm, "<loc>") != 11 {
		t.Errorf("sitemap has %d urls, want 11", strings.Count(sm, "<loc>"))
	}
}

// impeccable §2 — a missing catalogue key renders as the key. Every t('…')
// in the web source must exist in site_text. The oracle is the database
// after the migrations ran (the source of truth), not a list in a comment.
func TestCatalogueCoversEveryKey(t *testing.T) {
	e := fresh(t)
	have := map[string]bool{}
	rows, err := e.gdb.Raw(`SELECT key FROM site_text`).Rows()
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var k string
		_ = rows.Scan(&k)
		have[k] = true
	}
	rows.Close()
	// A literal key only: t('err.' + code) is built at run time and is
	// checked against the server's codes below.
	re := regexp.MustCompile(`\bt\(\s*'([a-z][a-zA-Z0-9_.]*)'\s*[,)]`)
	var missing []string
	_ = filepath.Walk("../web/src", func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !(strings.HasSuffix(p, ".tsx") || strings.HasSuffix(p, ".ts")) || strings.Contains(p, "/admin/") {
			return nil
		}
		b, _ := os.ReadFile(p)
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			if !have[m[1]] {
				missing = append(missing, p+": "+m[1])
			}
		}
		return nil
	})
	// err.<code> keys are built from server codes; check each code the
	// server can send has a catalogue entry.
	for _, code := range []string{"required", "too_long", "email", "not_allowed", "digits", "out_of_range", "control", "phone", "file_type", "file_size"} {
		if !have["err."+code] {
			missing = append(missing, "err."+code)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("site text keys used but not defined:\n%s", strings.Join(missing, "\n"))
	}
}

// The icons an editor can pick must all render: the Go allow-list and the
// web icon map hold the same names.
func TestIconListsAgree(t *testing.T) {
	b, err := os.ReadFile("../web/src/lib/icons.tsx")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`'([a-z0-9-]+)':\s*[A-Z]`)
	web := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(string(b), -1) {
		web[m[1]] = true
	}
	for _, i := range resource.Icons {
		if !web[i] {
			t.Errorf("icon %q is pickable in the CMS but not in web/src/lib/icons.tsx", i)
		}
		delete(web, i)
	}
	for i := range web {
		t.Errorf("icon %q is in the web map but not in the Go allow-list", i)
	}
}
