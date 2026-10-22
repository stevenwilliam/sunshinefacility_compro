package resource

import (
	"slices"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
)

// Icons the site can render (Phosphor names). The frontend's icon map in
// web/src/lib/icons.tsx must hold exactly these — a guard test compares the
// two, so an icon an editor can pick always renders.
var Icons = []string{
	"arrows-in-simple", "briefcase", "broom", "buildings", "bug", "calendar-check", "camera",
	"chart-line-up", "check-circle", "chef-hat", "clock", "cooking-pot", "device-mobile",
	"device-mobile-camera", "drop", "factory", "file-text", "first-aid-kit", "fork-knife",
	"graduation-cap", "handshake", "hard-hat", "heart", "house-line", "identification-badge",
	"leaf", "lightning", "map-pin", "medal", "path", "phone-call", "plant", "recycle",
	"shield-check", "signature", "sparkle", "storefront", "truck", "user-circle-check",
	"users-three", "wrench",
}

func IconAllowed(s string) bool { return slices.Contains(Icons, s) }

// Field shorthands.
func txt(key, label string, max int) Field {
	return Field{Key: key, Label: label, Type: Text, Max: max}
}

func (f Field) req() Field             { f.Required = true; return f }
func (f Field) content() Field         { f.Content = true; return f }
func (f Field) list() Field            { f.List = true; return f }
func (f Field) search() Field          { f.Search = true; return f }
func (f Field) fixed() Field           { f.Immutable = true; return f }
func (f Field) help(h string) Field    { f.Help = h; return f }
func (f Field) pattern(p string) Field { f.Pattern = p; return f }

// pair is the EN + ID twin of a text field (BR-1.1: both languages).
func pair(key, label string, t FieldType, max int, required bool) []Field {
	en := Field{Key: key + "_en", Label: label + " (EN)", Type: t, Max: max, Required: required, Content: true, Search: true}
	id := Field{Key: key + "_id", Label: label + " (ID)", Type: t, Max: max, Required: required, Content: true, Search: true}
	return []Field{en, id}
}

func sortF() Field {
	return Field{Key: "sort", Label: "Order", Type: Int, MinInt: 0, MaxInt: 9999, List: true}
}
func publishedF() Field {
	return Field{Key: "published", Label: "Published", Type: Bool, List: true}
}
func placeholderF() Field {
	return Field{Key: "is_placeholder", Label: "Placeholder", Type: Bool, List: true,
		Help: "Shown on the site with a visible Placeholder tag. Clears automatically when you change the content."}
}
func lineRef() Field {
	return Field{Key: "line_id", Label: "Service line", Type: Ref, Ref: "services", Required: true, List: true}
}

func fields(groups ...any) []Field {
	var out []Field
	for _, g := range groups {
		switch x := g.(type) {
		case Field:
			out = append(out, x)
		case []Field:
			out = append(out, x...)
		}
	}
	return out
}

var keyPattern = `^[a-z][a-z0-9-]{1,30}$`

const fixedSections = "Sections are fixed by the page layout. Edit the existing ones."

var pageOptions = []Option{
	{"home", "Home"}, {"services", "Services"}, {"service_detail", "Service detail"},
	{"industries", "Industries"}, {"about", "About Us"}, {"careers", "Careers"}, {"contact", "Contact"},
}

var listOptions = []Option{
	{"home_why", "Home · Why Sunshine points"},
	{"services_benefits", "Services · One contract, one lead, one report"},
	{"service_proof", "Service detail · Every job, on record"},
	{"careers_reasons", "Careers · Why people stay"},
	{"about_journey", "About · Journey timeline"},
	{"about_mission", "About · Mission"},
	{"about_values", "About · Values"},
	{"about_ecosystem", "About · Lippo ecosystem"},
}

// Defs is every CMS resource, in sidebar order.
var Defs = []Def{
	{
		Name: "pages", Table: "page_sections", IDColumn: "id", Label: "Pages", One: "section", Icon: "file-text", Nav: "Content",
		Description: "Titles, text and images for each page section, in English and Bahasa Indonesia.",
		WritePerm:   access.ContentWrite, NameField: "title_en", OrderBy: "page, sort",
		Filters:   []Filter{{Key: "page", Label: "Page", Options: pageOptions}},
		CreateWhy: fixedSections, DeleteWhy: fixedSections,
		Fields: fields(
			Field{Key: "page", Label: "Page", Type: Select, Options: pageOptions, Required: true, List: true, Immutable: true},
			txt("section", "Section", 40).req().list().search().fixed(),
			pair("title", "Title", Text, 300, false),
			pair("body", "Body", Textarea, 4000, false),
			Field{Key: "image_id", Label: "Image", Type: Media, MediaKind: "photo", Content: true},
			publishedF(), placeholderF(),
		),
	},
	{
		Name: "text", Table: "site_text", IDColumn: "key", Label: "Site text", One: "text", Icon: "text-aa", Nav: "Content",
		Description: "Every label, button and message on the public site, in both languages.",
		WritePerm:   access.ContentWrite, NameField: "key", OrderBy: "key",
		CreateWhy: "Interface text keys are fixed by the site's code. Edit the existing ones.",
		DeleteWhy: "Interface text keys are fixed by the site's code; removing one would leave a blank on the site.",
		Fields: fields(
			txt("key", "Key", 80).req().list().search().fixed(),
			Field{Key: "value_en", Label: "English", Type: Textarea, Max: 2000, Required: true, Content: true, List: true, Search: true},
			Field{Key: "value_id", Label: "Bahasa Indonesia", Type: Textarea, Max: 2000, Required: true, Content: true, List: true, Search: true},
			txt("note", "Note", 300).search(),
			placeholderF(),
		),
	},
	{
		Name: "services", Table: "service_lines", IDColumn: "id", Label: "Services", One: "service line", Icon: "squares-four", Nav: "Services",
		Description: "The business lines shown on Home, Services, Service detail and the hero ball.",
		WritePerm:   access.ContentWrite, NameField: "title_en", OrderBy: "sort, title_en",
		CanCreate: true, CanDelete: true,
		Fields: fields(
			Field{Key: "photo_id", Label: "Photo", Type: Media, MediaKind: "photo", Required: true, List: true, Content: true},
			sortF(),
			Field{Key: "title_en", Label: "Title (EN)", Type: Text, Max: 120, Required: true, Content: true, List: true, Search: true},
			Field{Key: "title_id", Label: "Title (ID)", Type: Text, Max: 120, Required: true, Content: true, Search: true},
			txt("key", "Key", 31).req().search().fixed().pattern(keyPattern).help("must be lowercase letters, digits or dashes, e.g. food. It is used in page links (/services/food) and cannot change later."),
			pair("short", "Short label", Text, 40, true),
			pair("label", "Filter label", Text, 60, true),
			pair("summary", "Summary (Home card, hero panel)", Textarea, 400, true),
			pair("description", "Description (Services page)", Textarea, 600, true),
			pair("lead", "Lead (Service detail hero)", Textarea, 400, true),
			Field{Key: "color_base", Label: "Base colour", Type: Colour, Required: true, List: true, Help: "Tiles, chips and markers."},
			Field{Key: "color_ink", Label: "Ink colour", Type: Colour, Required: true, Help: "Background behind text on the poster cards."},
			Field{Key: "color_shade", Label: "Shade colour", Type: Colour, Required: true, Help: "The mark inside the colour divider strip."},
			Field{Key: "color_on_ink", Label: "Text on ink", Type: Colour, Required: true, Help: "Must reach 4.5:1 against the ink colour."},
			Field{Key: "color_label", Label: "Label on white", Type: Colour, Required: true, Help: "Must reach 4.5:1 against white."},
			Field{Key: "mark_id", Label: "Line mark (SVG)", Type: Media, MediaKind: "brand", Required: true},
			Field{Key: "map_image_id", Label: "Coverage map", Type: Media, MediaKind: "photo"},
			Field{Key: "show_map", Label: "Show coverage map", Type: Bool},
			Field{Key: "show_certs", Label: "Show certification badges", Type: Bool},
			publishedF(),
		),
	},
	{
		Name: "service-items", Table: "service_items", IDColumn: "id", Label: "Service tags & points", One: "item", Icon: "list-bullets", Nav: "Services",
		Description: "The short lists on each line: the Home card tags, the Services page tags and the hero panel points.",
		WritePerm:   access.ContentWrite, NameField: "text_en", OrderBy: "line_id, kind, sort",
		CanCreate: true, CanDelete: true,
		Filters: []Filter{{Key: "line_id", Label: "Service line", Ref: "services"},
			{Key: "kind", Label: "Kind", Options: itemKinds}},
		Fields: fields(
			lineRef(),
			Field{Key: "kind", Label: "Kind", Type: Select, Options: itemKinds, Required: true, List: true},
			Field{Key: "text_en", Label: "Text (EN)", Type: Text, Max: 200, Required: true, Content: true, List: true, Search: true},
			Field{Key: "text_id", Label: "Text (ID)", Type: Text, Max: 200, Required: true, Content: true, Search: true},
			sortF(), publishedF(),
		),
	},
	{
		Name: "service-facts", Table: "service_facts", IDColumn: "id", Label: "Key figures", One: "figure", Icon: "chart-bar", Nav: "Services",
		Description: "The figures under each Service detail hero, such as 5M+ meals a year.",
		WritePerm:   access.ContentWrite, NameField: "value_en", OrderBy: "line_id, sort",
		CanCreate: true, CanDelete: true,
		Filters: []Filter{{Key: "line_id", Label: "Service line", Ref: "services"}},
		Fields: fields(
			lineRef(),
			Field{Key: "value_en", Label: "Value (EN)", Type: Text, Max: 20, Required: true, Content: true, List: true, Search: true},
			Field{Key: "value_id", Label: "Value (ID)", Type: Text, Max: 20, Required: true, Content: true, Search: true},
			Field{Key: "label_en", Label: "Label (EN)", Type: Text, Max: 80, Required: true, Content: true, List: true, Search: true},
			Field{Key: "label_id", Label: "Label (ID)", Type: Text, Max: 80, Required: true, Content: true, Search: true},
			sortF(), publishedF(),
		),
	},
	{
		Name: "service-groups", Table: "service_groups", IDColumn: "id", Label: "What's included", One: "group", Icon: "stack", Nav: "Services",
		Description: "The cards under \"What's included\" on each Service detail page: a list of sub-items, or a paragraph.",
		WritePerm:   access.ContentWrite, NameField: "title_en", OrderBy: "line_id, sort",
		CanCreate: true, CanDelete: true,
		Filters: []Filter{{Key: "line_id", Label: "Service line", Ref: "services"}},
		Fields: fields(
			lineRef(),
			Field{Key: "title_en", Label: "Title (EN)", Type: Text, Max: 120, Required: true, Content: true, List: true, Search: true},
			Field{Key: "title_id", Label: "Title (ID)", Type: Text, Max: 120, Required: true, Content: true, Search: true},
			pair("body", "Paragraph", Textarea, 800, false),
			Field{Key: "items_en", Label: "Sub-items (EN), one per line", Type: Lines, Max: 2000, Content: true, Search: true},
			Field{Key: "items_id", Label: "Sub-items (ID), one per line", Type: Lines, Max: 2000, Content: true, Search: true},
			sortF(), publishedF(),
		),
	},
	{
		Name: "industries", Table: "industries", IDColumn: "id", Label: "Industries", One: "industry", Icon: "buildings", Nav: "Content",
		Description: "The sectors on the Industries page, and which lines serve each (also shown on Service detail).",
		WritePerm:   access.ContentWrite, NameField: "title_en", OrderBy: "sort, title_en",
		CanCreate: true, CanDelete: true,
		RefJoin: map[string]Join{"lines": {Table: "industry_lines", OwnerCol: "industry_id", RefCol: "line_id"}},
		Fields: fields(
			Field{Key: "icon", Label: "Icon", Type: Icon, List: true},
			sortF(),
			Field{Key: "title_en", Label: "Title (EN)", Type: Text, Max: 120, Required: true, Content: true, List: true, Search: true},
			Field{Key: "title_id", Label: "Title (ID)", Type: Text, Max: 120, Required: true, Content: true, Search: true},
			txt("key", "Key", 31).req().search().fixed().pattern(keyPattern).help("must be lowercase letters, digits or dashes, e.g. healthcare."),
			pair("body", "Description", Textarea, 600, true),
			Field{Key: "lines", Label: "Services we usually run (in order)", Type: Refs, Ref: "services", Content: true},
			publishedF(),
		),
	},
	{
		Name: "clients", Table: "clients", IDColumn: "id", Label: "Clients", One: "client", Icon: "handshake", Nav: "Content",
		Description: "Logos in the Home trust bar and on the Industries page.",
		WritePerm:   access.ContentWrite, NameField: "name", OrderBy: "sort, name",
		CanCreate: true, CanDelete: true,
		Filters: []Filter{{Key: "industry_id", Label: "Industry", Ref: "industries"}},
		Fields: fields(
			Field{Key: "logo_id", Label: "Logo", Type: Media, MediaKind: "logo", Required: true, List: true, Help: "PNG with a transparent background works best."},
			txt("name", "Name", 120).req().list().search(),
			Field{Key: "industry_id", Label: "Industry", Type: Ref, Ref: "industries", List: true},
			sortF(), publishedF(),
		),
	},
	{
		Name: "certificates", Table: "certificates", IDColumn: "id", Label: "Certificates", One: "certificate", Icon: "seal-check", Nav: "Content",
		Description: "Shown on Home, About and Food Services. Expired certificates cannot be published.",
		WritePerm:   access.ContentWrite, NameField: "name", OrderBy: "sort, name",
		CanCreate: true, CanDelete: true,
		Fields: fields(
			Field{Key: "image_id", Label: "Certificate image (first page)", Type: Media, MediaKind: "certificate", Required: true, List: true},
			txt("name", "Certificate name", 120).req().list().search(),
			txt("issuer", "Issuer", 160).list().search(),
			Field{Key: "valid_until", Label: "Valid until", Type: Date, List: true, Help: "Leave empty if the certificate has no expiry date."},
			pair("description", "Description", Textarea, 400, false),
			sortF(), publishedF(),
		),
	},
	{
		Name: "careers", Table: "career_openings", IDColumn: "id", Label: "Careers", One: "opening", Icon: "briefcase", Nav: "Content",
		Description: "Openings on the Careers page.",
		WritePerm:   access.ContentWrite, NameField: "title_en", OrderBy: "sort, title_en",
		CanCreate: true, CanDelete: true,
		Filters: []Filter{{Key: "line_id", Label: "Service line", Ref: "services"},
			{Key: "location_key", Label: "Location", Group: "locations"}},
		Fields: fields(
			Field{Key: "title_en", Label: "Position (EN)", Type: Text, Max: 120, Required: true, Content: true, List: true, Search: true},
			Field{Key: "title_id", Label: "Position (ID)", Type: Text, Max: 120, Required: true, Content: true, Search: true},
			lineRef(),
			Field{Key: "location_key", Label: "Location", Type: Param, Group: "locations", Required: true, List: true, Content: true},
			Field{Key: "type_key", Label: "Type", Type: Param, Group: "job_types", Required: true, List: true, Content: true},
			Field{Key: "closing_date", Label: "Closing date", Type: Date, List: true, Content: true},
			pair("summary", "Summary", Textarea, 800, false),
			Field{Key: "requirements_en", Label: "Requirements (EN), one per line", Type: Lines, Max: 2000, Content: true, Search: true},
			Field{Key: "requirements_id", Label: "Requirements (ID), one per line", Type: Lines, Max: 2000, Content: true, Search: true},
			sortF(), publishedF(), placeholderF(),
		),
	},
	{
		Name: "testimonials", Table: "testimonials", IDColumn: "id", Label: "Testimonials", One: "testimonial", Icon: "quotes", Nav: "Content",
		Description: "Client quotes on Home. The three seeded rows are placeholders and stay unpublished.",
		WritePerm:   access.ContentWrite, NameField: "name", OrderBy: "sort, name",
		CanCreate: true, CanDelete: true,
		Fields: fields(
			txt("name", "Name", 120).req().list().search().content(),
			txt("company", "Company", 160).req().list().search().content(),
			txt("role", "Role", 120).search().content(),
			pair("quote", "Quote", Textarea, 600, true),
			sortF(), publishedF(), placeholderF(),
		),
	},
	{
		Name: "management", Table: "management", IDColumn: "id", Label: "Management", One: "person", Icon: "users-three", Nav: "Content",
		Description: "The management team on About Us.",
		WritePerm:   access.ContentWrite, NameField: "name", OrderBy: "sort, name",
		CanCreate: true, CanDelete: true,
		Fields: fields(
			Field{Key: "photo_id", Label: "Photo", Type: Media, MediaKind: "photo", List: true, Content: true},
			txt("name", "Name", 120).req().list().search().content(),
			pair("role", "Role", Text, 120, true),
			sortF(), publishedF(), placeholderF(),
		),
	},
	{
		Name: "lists", Table: "content_items", IDColumn: "id", Label: "Lists", One: "list item", Icon: "list-numbers", Nav: "Content",
		Description: "The short repeated lists: Why Sunshine, the journey timeline, mission, values, the Lippo ecosystem and more.",
		WritePerm:   access.ContentWrite, NameField: "title_en", OrderBy: "list_key, sort",
		CanCreate: true, CanDelete: true,
		Filters: []Filter{{Key: "list_key", Label: "List", Options: listOptions}},
		Fields: fields(
			Field{Key: "list_key", Label: "List", Type: Select, Options: listOptions, Required: true, List: true},
			Field{Key: "icon", Label: "Icon", Type: Icon},
			Field{Key: "title_en", Label: "Title (EN)", Type: Text, Max: 200, Content: true, List: true, Search: true},
			Field{Key: "title_id", Label: "Title (ID)", Type: Text, Max: 200, Content: true, Search: true},
			pair("body", "Text", Textarea, 800, false),
			pair("meta", "Year or label", Text, 40, false),
			Field{Key: "color", Label: "Accent colour", Type: Colour},
			sortF(), publishedF(), placeholderF(),
		),
	},
	{
		Name: "stats", Table: "stats", IDColumn: "id", Label: "Stats", One: "stat", Icon: "chart-line-up", Nav: "Content",
		Description: "The figures strip on Home and About Us.",
		WritePerm:   access.ContentWrite, NameField: "label_en", OrderBy: "scope, sort",
		CanCreate: true, CanDelete: true,
		Filters: []Filter{{Key: "scope", Label: "Page", Options: statScopes}},
		Fields: fields(
			Field{Key: "scope", Label: "Page", Type: Select, Options: statScopes, Required: true, List: true},
			Field{Key: "value_en", Label: "Value (EN)", Type: Text, Max: 20, Required: true, Content: true, List: true, Search: true},
			Field{Key: "value_id", Label: "Value (ID)", Type: Text, Max: 20, Required: true, Content: true, Search: true},
			Field{Key: "label_en", Label: "Label (EN)", Type: Text, Max: 80, Required: true, Content: true, List: true, Search: true},
			Field{Key: "label_id", Label: "Label (ID)", Type: Text, Max: 80, Required: true, Content: true, Search: true},
			sortF(), publishedF(),
		),
	},
	{
		Name: "parameters", Table: "parameters", IDColumn: "id", Label: "Parameters", One: "parameter", Icon: "sliders-horizontal", Nav: "Settings",
		Description: "Values used in forms and filters, contact details, notification recipients and thresholds.",
		WritePerm:   access.ParamsContent, NameField: "key", OrderBy: "grp, sort, key",
		CanCreate: true, CanDelete: true,
		Filters: []Filter{{Key: "grp", Label: "Group"}},
		Fields: fields(
			Field{Key: "grp", Label: "Group", Type: Text, Max: 40, Required: true, List: true, Search: true, Immutable: true},
			Field{Key: "key", Label: "Key", Type: Text, Max: 60, Required: true, List: true, Search: true, Immutable: true,
				Pattern: `^[a-z0-9][a-z0-9_.-]*$`, Help: "must be lowercase letters, digits, dots, dashes or underscores."},
			Field{Key: "value_en", Label: "Value (EN)", Type: Textarea, Max: 1000, List: true, Search: true, Content: true},
			Field{Key: "value_id", Label: "Value (ID)", Type: Textarea, Max: 1000, List: true, Search: true, Content: true},
			txt("description", "Description", 300).search(),
			sortF(), placeholderF(),
		),
	},
}

var itemKinds = []Option{{"card_tag", "Home card tag"}, {"service_tag", "Services page tag"}, {"hero_point", "Hero panel point"}}
var statScopes = []Option{{"home", "Home"}, {"about", "About Us"}}

// Lookup by URL name.
func Lookup(name string) (Def, bool) {
	for _, d := range Defs {
		if d.Name == name {
			return d, true
		}
	}
	return Def{}, false
}
