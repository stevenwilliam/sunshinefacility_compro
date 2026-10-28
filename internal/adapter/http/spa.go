package http

import (
	"bytes"
	"encoding/json"
	"html"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
)

// Public routes, for the sitemap and for telling a real page from a 404.
// AdminPath is where the CMS is served (web/src/admin/paths.ts must match).
// Not /admin, which every scanner probes; robots.txt does not name it either,
// the pages carry noindex instead.
const AdminPath = "/tytb"

var staticRoutes = []string{"/", "/services", "/industries", "/about", "/careers", "/contact"}

type pageMeta struct {
	title, description, canonical string
	status                        int
	noindex                       bool
	admin                         bool
}

func pick(v app.L2, lang string) string {
	if lang == "id" && v[1] != "" {
		return v[1]
	}
	return v[0]
}

func (s *Server) meta(b *app.Bundle, p, lang string) pageMeta {
	t := func(k string) string { return pick(b.Text[k], lang) }
	suffix := t("seo.suffix")
	m := pageMeta{status: http.StatusOK, canonical: s.BaseURL + p}
	switch {
	case p == "/":
		m.title, m.description = t("seo.home.title"), t("seo.home.description")
	case p == "/services" || p == "/industries" || p == "/about" || p == "/careers" || p == "/contact":
		key := strings.TrimPrefix(p, "/")
		m.title, m.description = t("seo."+key+".title")+suffix, t("seo."+key+".description")
	case strings.HasPrefix(p, "/services/"):
		key := strings.TrimPrefix(p, "/services/")
		for _, l := range b.Lines {
			if l.Key == key {
				m.title, m.description = pick(l.Title, lang)+suffix, pick(l.Lead, lang)
				return m
			}
		}
		fallthrough
	default:
		m.title, m.description = t("notfound.title")+suffix, t("notfound.body")
		m.status, m.noindex, m.canonical = http.StatusNotFound, true, ""
	}
	return m
}

func (s *Server) head(b *app.Bundle, m pageMeta, lang string) string {
	e := html.EscapeString
	var h strings.Builder
	h.WriteString("<title>" + e(m.title) + "</title>\n")
	h.WriteString(`<meta name="description" content="` + e(m.description) + "\">\n")
	if m.noindex {
		h.WriteString(`<meta name="robots" content="noindex, nofollow">` + "\n")
	}
	if m.admin {
		return h.String()
	}
	if m.canonical != "" {
		h.WriteString(`<link rel="canonical" href="` + e(m.canonical) + "\">\n")
	}
	// Link-preview bots do not run JavaScript; these tags are the whole of
	// what WhatsApp, LinkedIn and Slack will ever show (CLAUDE.md §13).
	locale := "en_US"
	if lang == "id" {
		locale = "id_ID"
	}
	og := [][2]string{
		{"og:type", "website"}, {"og:site_name", "Sunshine Facilities Group"}, {"og:title", m.title},
		{"og:description", m.description}, {"og:url", m.canonical}, {"og:locale", locale},
	}
	if b.OGImage != "" {
		og = append(og, [2]string{"og:image", s.BaseURL + b.OGImage})
	}
	for _, kv := range og {
		h.WriteString(`<meta property="` + kv[0] + `" content="` + e(kv[1]) + "\">\n")
	}
	h.WriteString(`<meta name="twitter:card" content="summary_large_image">` + "\n")
	h.WriteString(`<meta name="twitter:title" content="` + e(m.title) + "\">\n")
	h.WriteString(`<meta name="twitter:description" content="` + e(m.description) + "\">\n")
	if b.OGImage != "" {
		h.WriteString(`<meta name="twitter:image" content="` + e(s.BaseURL+b.OGImage) + "\">\n")
	}
	// Organization JSON-LD. Contact details still marked placeholder are
	// left out: structured data is read by machines that will not see the
	// on-page "Placeholder" tag, and a fake phone number would spread.
	// An empty phone is hidden on the site, so it is left out here too.
	org := map[string]any{
		"@context": "https://schema.org", "@type": "Organization",
		"name": "Sunshine Facilities Group", "url": s.BaseURL + "/",
		"parentOrganization": map[string]any{"@type": "Organization", "name": "Lippo Indonesia"},
	}
	if logo, ok := b.Logos["sfg-logo-primary"]; ok && logo != nil {
		org["logo"] = s.BaseURL + logo.Src
	}
	if v, ok := b.Params["contact.phone"]; ok && !v.Placeholder && v.V[0] != "" {
		org["telephone"] = v.V[0]
	}
	if v, ok := b.Params["contact.email"]; ok && !v.Placeholder {
		org["email"] = v.V[0]
	}
	if v, ok := b.Params["contact.address"]; ok && !v.Placeholder {
		org["address"] = v.V[0]
	}
	ld, _ := json.Marshal(org) // json.Marshal escapes <, > and & — safe inside <script>
	h.WriteString(`<script type="application/ld+json">` + string(ld) + "</script>\n")
	return h.String()
}

func (s *Server) spa(c *gin.Context) {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		c.Status(http.StatusNotFound)
		return
	}
	p := path.Clean("/" + c.Request.URL.Path)
	if strings.HasPrefix(p, "/api/") {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Not found"}})
		return
	}
	// A real file from the build.
	if p != "/" && p != "/index.html" {
		if f, err := fs.Stat(s.Web, strings.TrimPrefix(p, "/")); err == nil && !f.IsDir() {
			s.static(c, strings.TrimPrefix(p, "/"))
			return
		}
	}
	// The design's links used ?line=; real URLs carry the line in the path.
	// Old links are redirected permanently rather than broken.
	if line := c.Query("line"); line != "" && (p == "/services" || p == "/service" || p == "/service-detail") {
		c.Redirect(http.StatusMovedPermanently, "/services/"+url(line))
		return
	}
	lang := "en"
	if c.Query("lang") == "id" {
		lang = "id"
	}
	var m pageMeta
	var data []byte
	if p == AdminPath || strings.HasPrefix(p, AdminPath+"/") {
		m = pageMeta{title: "Admin · Sunshine Facilities Group", status: http.StatusOK, noindex: true, admin: true}
		c.Header("X-Robots-Tag", "noindex, nofollow")
	} else {
		b, raw, err := s.Public.Get(c.Request.Context())
		if err != nil {
			s.fail(c, err)
			return
		}
		m = s.meta(b, p, lang)
		data = raw
		c.Header("Content-Language", lang)
		head := s.head(b, m, lang)
		s.writeIndex(c, m.status, lang, head, data)
		return
	}
	s.writeIndex(c, m.status, "en", s.head(nil, m, "en"), nil)
}

func (s *Server) writeIndex(c *gin.Context, status int, lang, head string, data []byte) {
	out := bytes.Replace(s.index, []byte(`<html lang="en">`), []byte(`<html lang="`+lang+`">`), 1)
	out = bytes.Replace(out, []byte("<!--SFG_HEAD-->"), []byte(head), 1)
	if data != nil {
		out = bytes.Replace(out, []byte("<!--SFG_DATA-->"),
			append(append([]byte(`<script id="sfg-bundle" type="application/json">`), data...), []byte("</script>")...), 1)
	}
	c.Header("Cache-Control", "no-cache")
	c.Data(status, "text/html; charset=utf-8", out)
}

// static serves a built file. Vite puts content-hashed files under assets/,
// and only those are immutable; anything else revalidates (impeccable §5).
func (s *Server) static(c *gin.Context, name string) {
	b, err := fs.ReadFile(s.Web, name)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if strings.HasPrefix(name, "assets/") {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "no-cache")
	}
	ct := mime.TypeByExtension(path.Ext(name))
	if ct == "" {
		ct = http.DetectContentType(b)
	}
	c.Data(http.StatusOK, ct, b)
}

func url(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (s *Server) robots(c *gin.Context) {
	c.Header("Cache-Control", "no-cache")
	c.String(http.StatusOK, "User-agent: *\nAllow: /\nDisallow: /api/\n\nSitemap: "+s.BaseURL+"/sitemap.xml\n")
}

func (s *Server) sitemap(c *gin.Context) {
	b, _, err := s.Public.Get(c.Request.Context())
	if err != nil {
		s.fail(c, err)
		return
	}
	var x strings.Builder
	x.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	routes := append([]string{}, staticRoutes...)
	for _, l := range b.Lines {
		routes = append(routes, "/services/"+l.Key)
	}
	for _, r := range routes {
		x.WriteString("  <url><loc>" + html.EscapeString(s.BaseURL+r) + "</loc></url>\n")
	}
	x.WriteString("</urlset>\n")
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(x.String()))
}
