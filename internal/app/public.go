package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/calendar"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/certificate"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/forms"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/params"
)

// Snapshot is the published content as the repository reads it. Only rows
// marked published are in it (BR-1.3); certificates are filtered again here
// by date (BR-3.3), because "published" alone does not make one valid.
type Snapshot struct {
	Text         []TextRow
	Sections     []SectionRow
	Params       []Parameter
	Media        map[string]MediaRecord
	Lines        []LineRow
	Items        []ItemRow
	Facts        []FactRow
	Groups       []GroupRow
	Industries   []IndustryRow
	Clients      []ClientRow
	Certificates []CertRow
	Lists        []ListRow
	Stats        []StatRow
	Openings     []OpeningRow
	Testimonials []TestimonialRow
	Management   []PersonRow
}

type TextRow struct {
	Key, EN, ID string
	Placeholder bool
}

type SectionRow struct {
	Page, Section, TitleEN, TitleID, BodyEN, BodyID, ImageID string
	Published, Placeholder                                   bool
}

type LineRow struct {
	ID, Key                                                     string
	Sort                                                        int
	TitleEN, TitleID, ShortEN, ShortID, LabelEN, LabelID        string
	SummaryEN, SummaryID, DescriptionEN, DescriptionID          string
	LeadEN, LeadID                                              string
	Base, Ink, Shade, OnInk, Label, MarkID, PhotoID, MapImageID string
	ShowMap, ShowCerts                                          bool
}

type ItemRow struct{ LineID, Kind, EN, ID string }
type FactRow struct{ LineID, ValueEN, ValueID, LabelEN, LabelID string }
type GroupRow struct {
	LineID, TitleEN, TitleID, BodyEN, BodyID, ItemsEN, ItemsID string
}
type IndustryRow struct {
	ID, Key, Icon, TitleEN, TitleID, BodyEN, BodyID string
	LineIDs                                         []string
}
type ClientRow struct{ Name, LogoID, IndustryID string }
type CertRow struct {
	ID, Name, Issuer, ValidUntil, ImageID, DescriptionEN, DescriptionID string
	Published                                                           bool
}
type ListRow struct {
	ListKey, Icon, TitleEN, TitleID, BodyEN, BodyID, MetaEN, MetaID, Color string
	Placeholder                                                            bool
}
type StatRow struct{ Scope, ValueEN, ValueID, LabelEN, LabelID string }
type OpeningRow struct {
	ID, TitleEN, TitleID, LineID, LocationKey, TypeKey, ClosingDate string
	SummaryEN, SummaryID, RequirementsEN, RequirementsID            string
	Placeholder                                                     bool
}
type TestimonialRow struct {
	QuoteEN, QuoteID, Name, Role, Company string
	Placeholder                           bool
}
type PersonRow struct {
	Name, RoleEN, RoleID, PhotoID string
	Placeholder                   bool
}

// ---------------------------------------------------------------------------
// Bundle: what the public site receives. Every bilingual value is [en, id].
// ---------------------------------------------------------------------------

type L2 [2]string

type Img struct {
	Src    string `json:"src"`
	SrcSet string `json:"srcset,omitempty"`
	W      int    `json:"w,omitempty"`
	H      int    `json:"h,omitempty"`
	Alt    L2     `json:"alt"`
	SVG    bool   `json:"svg,omitempty"`
}

type Section struct {
	Title       L2   `json:"title"`
	Body        L2   `json:"body"`
	Image       *Img `json:"image,omitempty"`
	Published   bool `json:"published"`
	Placeholder bool `json:"placeholder"`
}

type ParamValue struct {
	V           L2   `json:"v"`
	Placeholder bool `json:"placeholder,omitempty"`
}

type OptionV struct {
	Key   string `json:"key"`
	Label L2     `json:"label"`
}

type LineV struct {
	ID          string   `json:"id"`
	Key         string   `json:"key"`
	N           string   `json:"n"`
	Title       L2       `json:"title"`
	Short       L2       `json:"short"`
	Label       L2       `json:"label"`
	Summary     L2       `json:"summary"`
	Description L2       `json:"description"`
	Lead        L2       `json:"lead"`
	Base        string   `json:"base"`
	Ink         string   `json:"ink"`
	Shade       string   `json:"shade"`
	OnInk       string   `json:"onInk"`
	LabelColor  string   `json:"labelColor"`
	Mark        *Img     `json:"mark"`
	Photo       *Img     `json:"photo"`
	Map         *Img     `json:"map,omitempty"`
	ShowMap     bool     `json:"showMap"`
	ShowCerts   bool     `json:"showCerts"`
	CardTags    []L2     `json:"cardTags"`
	ServiceTags []L2     `json:"serviceTags"`
	HeroPoints  []L2     `json:"heroPoints"`
	Facts       []FactV  `json:"facts"`
	Groups      []GroupV `json:"groups"`
	Industries  []string `json:"industries"`
}

type FactV struct {
	Value L2 `json:"value"`
	Label L2 `json:"label"`
}

type GroupV struct {
	Title L2          `json:"title"`
	Body  L2          `json:"body"`
	Items [2][]string `json:"items"`
}

type IndustryV struct {
	Key     string   `json:"key"`
	Icon    string   `json:"icon"`
	Title   L2       `json:"title"`
	Body    L2       `json:"body"`
	Lines   []string `json:"lines"`
	Clients []int    `json:"clients"`
}

type ClientV struct {
	Name     string `json:"name"`
	Logo     *Img   `json:"logo"`
	Industry string `json:"industry"`
}

type CertV struct {
	Name        string `json:"name"`
	Issuer      string `json:"issuer"`
	ValidUntil  string `json:"validUntil"`
	Description L2     `json:"description"`
	Image       *Img   `json:"image"`
	Full        string `json:"full"` // largest rendition, for the lightbox
}

type ListItemV struct {
	Icon        string `json:"icon,omitempty"`
	Title       L2     `json:"title"`
	Body        L2     `json:"body"`
	Meta        L2     `json:"meta"`
	Color       string `json:"color,omitempty"`
	Placeholder bool   `json:"placeholder,omitempty"`
}

type StatV struct {
	Value L2 `json:"value"`
	Label L2 `json:"label"`
}

type OpeningV struct {
	ID           string      `json:"id"`
	Title        L2          `json:"title"`
	Line         string      `json:"line"`
	Location     string      `json:"location"`
	Type         string      `json:"type"`
	ClosingDate  string      `json:"closingDate,omitempty"`
	Summary      L2          `json:"summary"`
	Requirements [2][]string `json:"requirements"`
	Placeholder  bool        `json:"placeholder,omitempty"`
}

type TestimonialV struct {
	Quote       L2     `json:"quote"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Placeholder bool   `json:"placeholder,omitempty"`
}

type PersonV struct {
	Name        string `json:"name"`
	Role        L2     `json:"role"`
	Photo       *Img   `json:"photo,omitempty"`
	Placeholder bool   `json:"placeholder,omitempty"`
}

type Rules struct {
	Limits         forms.Limits `json:"limits"`
	CVMaxMB        int          `json:"cvMaxMb"`
	MinFillSeconds int          `json:"minFillSeconds"`
}

type Bundle struct {
	Version       string                 `json:"version"`
	Today         string                 `json:"today"`
	BaseURL       string                 `json:"baseUrl"`
	Text          map[string]L2          `json:"text"`
	TextPH        []string               `json:"textPlaceholders"`
	Sections      map[string]Section     `json:"sections"`
	Params        map[string]ParamValue  `json:"params"`
	BuildingTypes []OptionV              `json:"buildingTypes"`
	Locations     []OptionV              `json:"locations"`
	JobTypes      []OptionV              `json:"jobTypes"`
	Logos         map[string]*Img        `json:"logos"`
	Lines         []LineV                `json:"lines"`
	Industries    []IndustryV            `json:"industries"`
	Clients       []ClientV              `json:"clients"`
	Certificates  []CertV                `json:"certificates"`
	Lists         map[string][]ListItemV `json:"lists"`
	Stats         map[string][]StatV     `json:"stats"`
	Openings      []OpeningV             `json:"openings"`
	Testimonials  []TestimonialV         `json:"testimonials"`
	Management    []PersonV              `json:"management"`
	Rules         Rules                  `json:"rules"`
	GA4           string                 `json:"ga4"`
	OGImage       string                 `json:"ogImage"`
}

// publicParamGroups are the parameter groups the site may read. Recipients,
// thresholds and anything secret never leave the server.
var publicParamGroups = map[string]bool{"contact": true, "social": true}

// ImageURL is the public URL of one rendition. The hash is in the path, so
// the bytes at a URL never change and the response can be cached forever
// (impeccable §5).
func ImageURL(m MediaRecord, w int) string {
	h := m.Hash
	if len(h) > 12 {
		h = h[:12]
	}
	if m.Mime == "image/svg+xml" {
		return fmt.Sprintf("/media/%s/%s.svg", m.ID, h)
	}
	return fmt.Sprintf("/media/%s/%s-%d.webp", m.ID, h, w)
}

func toImg(m MediaRecord, ok bool) *Img {
	if !ok || m.Hash == "" {
		return nil // no file yet: render nothing rather than a broken image
	}
	img := &Img{Alt: L2{m.AltEN, m.AltID}, W: m.Width, H: m.Height}
	if m.Mime == "image/svg+xml" {
		img.Src = ImageURL(m, 0)
		img.SVG = true
		return img
	}
	if len(m.Variants) == 0 {
		return nil
	}
	vs := append([]MediaVariant(nil), m.Variants...)
	sort.Slice(vs, func(i, j int) bool { return vs[i].W < vs[j].W })
	var parts []string
	for _, v := range vs {
		parts = append(parts, fmt.Sprintf("%s %dw", ImageURL(m, v.W), v.W))
	}
	// Default src: the 960 rendition if there is one, else the largest.
	src := vs[len(vs)-1].W
	for _, v := range vs {
		if v.W == 960 {
			src = 960
		}
	}
	img.Src = ImageURL(m, src)
	img.SrcSet = strings.Join(parts, ", ")
	return img
}

func largest(m MediaRecord) string {
	best := 0
	for _, v := range m.Variants {
		if v.W > best {
			best = v.W
		}
	}
	if best == 0 {
		return ""
	}
	return ImageURL(m, best)
}

func splitLines(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

// BuildBundle assembles the public bundle. It is pure given its inputs.
func BuildBundle(s *Snapshot, today calendar.Date, baseURL string) *Bundle {
	img := func(id string) *Img {
		if id == "" {
			return nil
		}
		m, ok := s.Media[id]
		return toImg(m, ok)
	}
	b := &Bundle{
		Today: today.String(), BaseURL: baseURL,
		Text: map[string]L2{}, TextPH: []string{},
		Sections: map[string]Section{}, Params: map[string]ParamValue{},
		Logos: map[string]*Img{}, Lists: map[string][]ListItemV{}, Stats: map[string][]StatV{},
		Lines: []LineV{}, Industries: []IndustryV{}, Clients: []ClientV{}, Certificates: []CertV{},
		Openings: []OpeningV{}, Testimonials: []TestimonialV{}, Management: []PersonV{},
		BuildingTypes: []OptionV{}, Locations: []OptionV{}, JobTypes: []OptionV{},
	}
	for _, t := range s.Text {
		b.Text[t.Key] = L2{t.EN, t.ID}
		if t.Placeholder {
			b.TextPH = append(b.TextPH, t.Key)
		}
	}
	for _, sec := range s.Sections {
		b.Sections[sec.Page+"."+sec.Section] = Section{
			Title: L2{sec.TitleEN, sec.TitleID}, Body: L2{sec.BodyEN, sec.BodyID},
			Image: img(sec.ImageID), Published: sec.Published, Placeholder: sec.Placeholder,
		}
	}
	rules := Rules{Limits: forms.Rules, CVMaxMB: 5, MinFillSeconds: 3}
	for _, p := range s.Params {
		switch p.Group {
		case "building_types":
			b.BuildingTypes = append(b.BuildingTypes, OptionV{p.Key, L2{p.ValueEN, p.ValueID}})
		case "locations":
			b.Locations = append(b.Locations, OptionV{p.Key, L2{p.ValueEN, p.ValueID}})
		case "job_types":
			b.JobTypes = append(b.JobTypes, OptionV{p.Key, L2{p.ValueEN, p.ValueID}})
		}
		if publicParamGroups[p.Group] && !p.IsSecret {
			b.Params[p.Group+"."+p.Key] = ParamValue{V: L2{p.ValueEN, p.ValueID}, Placeholder: p.IsPlaceholder}
		}
		switch p.Group + "." + p.Key {
		case "careers.cv_max_mb":
			rules.CVMaxMB = atoiOr(p.ValueEN, 5)
		case "forms.min_fill_seconds":
			rules.MinFillSeconds = atoiOr(p.ValueEN, 3)
		case "analytics.ga4_measurement_id":
			b.GA4 = p.ValueEN
		case "seo.og_image":
			for _, m := range s.Media {
				if m.Slug == p.ValueEN {
					if i := toImg(m, true); i != nil {
						b.OGImage = largest(m)
					}
				}
			}
		}
	}
	// WhatsApp digits for wa.me links, derived once here.
	if wa, ok := b.Params["contact.whatsapp"]; ok {
		d := params.WhatsAppDigits(wa.V[0])
		b.Params["contact.whatsapp_digits"] = ParamValue{V: L2{d, d}, Placeholder: wa.Placeholder}
	}
	b.Rules = rules

	for _, m := range s.Media {
		if m.Kind == "brand" && m.Slug != "" {
			if i := toImg(m, true); i != nil {
				b.Logos[strings.TrimPrefix(m.Slug, "brand/")] = i
			}
		}
	}

	lineKey := map[string]string{}
	for _, l := range s.Lines {
		lineKey[l.ID] = l.Key
	}
	industryKey := map[string]string{}
	for _, in := range s.Industries {
		industryKey[in.ID] = in.Key
	}
	// D12: the line → industries relation is the inverse of industry_lines.
	lineIndustries := map[string][]string{}
	for _, in := range s.Industries {
		for _, lid := range in.LineIDs {
			lineIndustries[lid] = append(lineIndustries[lid], in.Key)
		}
	}

	for i, l := range s.Lines {
		lv := LineV{
			ID: l.ID, Key: l.Key, N: fmt.Sprintf("%02d", i+1),
			Title: L2{l.TitleEN, l.TitleID}, Short: L2{l.ShortEN, l.ShortID}, Label: L2{l.LabelEN, l.LabelID},
			Summary: L2{l.SummaryEN, l.SummaryID}, Description: L2{l.DescriptionEN, l.DescriptionID}, Lead: L2{l.LeadEN, l.LeadID},
			Base: l.Base, Ink: l.Ink, Shade: l.Shade, OnInk: l.OnInk, LabelColor: l.Label,
			Mark: img(l.MarkID), Photo: img(l.PhotoID), ShowMap: l.ShowMap, ShowCerts: l.ShowCerts,
			CardTags: []L2{}, ServiceTags: []L2{}, HeroPoints: []L2{}, Facts: []FactV{}, Groups: []GroupV{},
			Industries: lineIndustries[l.ID],
		}
		if lv.Industries == nil {
			lv.Industries = []string{}
		}
		if l.ShowMap {
			lv.Map = img(l.MapImageID)
		}
		for _, it := range s.Items {
			if it.LineID != l.ID {
				continue
			}
			switch it.Kind {
			case "card_tag":
				lv.CardTags = append(lv.CardTags, L2{it.EN, it.ID})
			case "service_tag":
				lv.ServiceTags = append(lv.ServiceTags, L2{it.EN, it.ID})
			case "hero_point":
				lv.HeroPoints = append(lv.HeroPoints, L2{it.EN, it.ID})
			}
		}
		for _, f := range s.Facts {
			if f.LineID == l.ID {
				lv.Facts = append(lv.Facts, FactV{L2{f.ValueEN, f.ValueID}, L2{f.LabelEN, f.LabelID}})
			}
		}
		for _, g := range s.Groups {
			if g.LineID == l.ID {
				lv.Groups = append(lv.Groups, GroupV{Title: L2{g.TitleEN, g.TitleID}, Body: L2{g.BodyEN, g.BodyID},
					Items: [2][]string{splitLines(g.ItemsEN), splitLines(g.ItemsID)}})
			}
		}
		b.Lines = append(b.Lines, lv)
	}

	clientIdx := map[string][]int{}
	for i, c := range s.Clients {
		b.Clients = append(b.Clients, ClientV{Name: c.Name, Logo: img(c.LogoID), Industry: industryKey[c.IndustryID]})
		clientIdx[c.IndustryID] = append(clientIdx[c.IndustryID], i)
	}
	for _, in := range s.Industries {
		iv := IndustryV{Key: in.Key, Icon: in.Icon, Title: L2{in.TitleEN, in.TitleID}, Body: L2{in.BodyEN, in.BodyID},
			Lines: []string{}, Clients: clientIdx[in.ID]}
		for _, lid := range in.LineIDs {
			if k, ok := lineKey[lid]; ok { // an unpublished line drops out of the chips
				iv.Lines = append(iv.Lines, k)
			}
		}
		if iv.Clients == nil {
			iv.Clients = []int{}
		}
		b.Industries = append(b.Industries, iv)
	}

	for _, c := range s.Certificates {
		var vu *calendar.Date
		if c.ValidUntil != "" {
			if d, err := calendar.ParseDate(c.ValidUntil); err == nil {
				vu = &d
			}
		}
		if !certificate.PubliclyVisible(c.Published, vu, today) { // BR-3.3
			continue
		}
		cv := CertV{Name: c.Name, Issuer: c.Issuer, ValidUntil: c.ValidUntil,
			Description: L2{c.DescriptionEN, c.DescriptionID}, Image: img(c.ImageID)}
		if m, ok := s.Media[c.ImageID]; ok {
			cv.Full = largest(m)
		}
		b.Certificates = append(b.Certificates, cv)
	}

	for _, li := range s.Lists {
		b.Lists[li.ListKey] = append(b.Lists[li.ListKey], ListItemV{
			Icon: li.Icon, Title: L2{li.TitleEN, li.TitleID}, Body: L2{li.BodyEN, li.BodyID},
			Meta: L2{li.MetaEN, li.MetaID}, Color: li.Color, Placeholder: li.Placeholder})
	}
	for _, st := range s.Stats {
		b.Stats[st.Scope] = append(b.Stats[st.Scope], StatV{L2{st.ValueEN, st.ValueID}, L2{st.LabelEN, st.LabelID}})
	}
	for _, o := range s.Openings {
		key, ok := lineKey[o.LineID]
		if !ok {
			continue // its line is unpublished
		}
		b.Openings = append(b.Openings, OpeningV{ID: o.ID, Title: L2{o.TitleEN, o.TitleID}, Line: key,
			Location: o.LocationKey, Type: o.TypeKey, ClosingDate: o.ClosingDate,
			Summary:      L2{o.SummaryEN, o.SummaryID},
			Requirements: [2][]string{splitLines(o.RequirementsEN), splitLines(o.RequirementsID)},
			Placeholder:  o.Placeholder})
	}
	for _, t := range s.Testimonials {
		b.Testimonials = append(b.Testimonials, TestimonialV{Quote: L2{t.QuoteEN, t.QuoteID}, Name: t.Name, Role: t.Role,
			Company: t.Company, Placeholder: t.Placeholder})
	}
	for _, p := range s.Management {
		b.Management = append(b.Management, PersonV{Name: p.Name, Role: L2{p.RoleEN, p.RoleID}, Photo: img(p.PhotoID), Placeholder: p.Placeholder})
	}

	raw, _ := json.Marshal(b)
	sum := sha256.Sum256(raw)
	b.Version = hex.EncodeToString(sum[:8])
	return b
}

func atoiOr(s string, def int) int {
	n := 0
	if s == "" {
		return def
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
		if n > 1_000_000 {
			return def
		}
	}
	return n
}

// Public serves the bundle from a cache. Any content write calls Invalidate;
// the cache key also carries today's date in Jakarta, so a certificate that
// expires overnight disappears at the next request after midnight without a
// write (BR-3.3).
type Public struct {
	store   Store
	baseURL string
	now     func() time.Time

	mu      sync.Mutex
	day     string
	gen     uint64
	builtAt uint64
	cached  *Bundle
	json    []byte
}

func NewPublic(store Store, baseURL string, now func() time.Time) *Public {
	return &Public{store: store, baseURL: baseURL, now: now, gen: 1}
}

func (p *Public) Invalidate() {
	p.mu.Lock()
	p.gen++
	p.mu.Unlock()
}

// Get returns the bundle and its JSON encoding.
func (p *Public) Get(ctx context.Context) (*Bundle, []byte, error) {
	today := calendar.Today(p.now())
	p.mu.Lock()
	if p.cached != nil && p.day == today.String() && p.builtAt == p.gen {
		b, j := p.cached, p.json
		p.mu.Unlock()
		return b, j, nil
	}
	gen := p.gen
	p.mu.Unlock()

	snap, err := p.store.PublicSnapshot(ctx)
	if err != nil {
		return nil, nil, err
	}
	b := BuildBundle(snap, today, p.baseURL)
	j, err := json.Marshal(b)
	if err != nil {
		return nil, nil, err
	}
	p.mu.Lock()
	p.cached, p.json, p.day, p.builtAt = b, j, today.String(), gen
	p.mu.Unlock()
	return b, j, nil
}
