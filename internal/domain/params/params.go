// Package params is the registry of parameters the code depends on, with the
// type each value must have (CLAUDE.md §8).
//
// A parameter is a row the business edits without a deploy — and therefore a
// value that can be wrong. Every known key declares its type here, so a
// retention of "twelve" or a recipient of "itdept" is refused at save time
// with a reason, instead of breaking a job at 02:00.
package params

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Kind string

const (
	KindText      Kind = "text"
	KindInt       Kind = "int"
	KindEmail     Kind = "email"
	KindEmailList Kind = "email_list"
	KindURL       Kind = "url"
	KindDecimal   Kind = "decimal"
	KindColour    Kind = "colour"
	KindGA4       Kind = "ga4"
	KindSlug      Kind = "media_slug"
)

type Spec struct {
	Kind     Kind
	Min, Max int64 // for KindInt and KindDecimal (decimal bounds are whole degrees)
	// Bilingual: the Indonesian value is shown to Indonesian visitors and
	// must be supplied. Non-bilingual values copy EN into ID.
	Bilingual bool
}

// Group describes a parameter group.
type Group struct {
	Key   string
	Label string
	// System groups change how the system behaves and need params.system.
	System bool
	// Open groups are option lists: rows may be added and removed. In a
	// closed group the keys are fixed by the code; only values change.
	Open bool
	// Spec for every row of an open group.
	RowSpec Spec
}

var Groups = []Group{
	{Key: "building_types", Label: "Building types", Open: true, RowSpec: Spec{Kind: KindText, Bilingual: true}},
	{Key: "locations", Label: "Locations", Open: true, RowSpec: Spec{Kind: KindText, Bilingual: true}},
	{Key: "job_types", Label: "Job types", Open: true, RowSpec: Spec{Kind: KindText, Bilingual: true}},
	{Key: "brand_colours", Label: "Brand colours", Open: true, RowSpec: Spec{Kind: KindColour}},
	{Key: "contact", Label: "Contact details"},
	{Key: "social", Label: "Social links"},
	{Key: "notify", Label: "Notifications", System: true},
	{Key: "certificates", Label: "Certificates", System: true},
	{Key: "careers", Label: "Careers", System: true},
	{Key: "forms", Label: "Forms and spam protection", System: true},
	{Key: "media", Label: "Media uploads", System: true},
	{Key: "auth", Label: "Sign-in security", System: true},
	{Key: "analytics", Label: "Analytics", System: true},
	{Key: "seo", Label: "Search and link previews", System: true},
}

// Keys of the closed groups and the type of each.
var Keys = map[string]Spec{
	"contact.phone":    {Kind: KindText},
	"contact.whatsapp": {Kind: KindText},
	"contact.email":    {Kind: KindEmail},
	"contact.address":  {Kind: KindText, Bilingual: true},
	"contact.hours":    {Kind: KindText, Bilingual: true},
	"contact.map_lat":  {Kind: KindDecimal, Min: -90, Max: 90},
	"contact.map_lng":  {Kind: KindDecimal, Min: -180, Max: 180},

	"social.instagram": {Kind: KindURL},
	"social.linkedin":  {Kind: KindURL},
	"social.facebook":  {Kind: KindURL},
	"social.youtube":   {Kind: KindURL},

	"notify.quotation_to":   {Kind: KindEmailList},
	"notify.career_to":      {Kind: KindEmailList},
	"notify.certificate_to": {Kind: KindEmailList},
	"notify.from":           {Kind: KindEmail},

	"certificates.reminder_days":   {Kind: KindInt, Min: 1, Max: 365},
	"careers.retention_months":     {Kind: KindInt, Min: 1, Max: 120},
	"careers.cv_max_mb":            {Kind: KindInt, Min: 1, Max: 20},
	"forms.min_fill_seconds":       {Kind: KindInt, Min: 0, Max: 120},
	"forms.per_ip_per_hour":        {Kind: KindInt, Min: 1, Max: 1000},
	"forms.per_email_per_hour":     {Kind: KindInt, Min: 1, Max: 1000},
	"media.max_upload_mb":          {Kind: KindInt, Min: 1, Max: 50},
	"media.max_dimension_px":       {Kind: KindInt, Min: 500, Max: 20000},
	"auth.password_min_length":     {Kind: KindInt, Min: 8, Max: 128},
	"auth.max_failed_logins":       {Kind: KindInt, Min: 3, Max: 50},
	"auth.lockout_minutes":         {Kind: KindInt, Min: 1, Max: 1440},
	"analytics.ga4_measurement_id": {Kind: KindGA4},
	"seo.og_image":                 {Kind: KindSlug},
}

func GroupOf(key string) (Group, bool) {
	for _, g := range Groups {
		if g.Key == key {
			return g, true
		}
	}
	return Group{}, false
}

// SpecFor returns the spec for grp.key.
func SpecFor(grp, key string) (Spec, bool) {
	g, ok := GroupOf(grp)
	if !ok {
		return Spec{}, false
	}
	if g.Open {
		return g.RowSpec, true
	}
	s, ok := Keys[grp+"."+key]
	return s, ok
}

var (
	ga4Re    = regexp.MustCompile(`^G-[A-Z0-9]{4,16}$`)
	colourRe = regexp.MustCompile(`^#[0-9A-F]{6}$`)
	slugRe   = regexp.MustCompile(`^[a-z]+/[a-z0-9-]+$`)
)

// Validate checks one value against its spec and returns the value to store
// (normalised: trimmed, colour upper-cased) or a reason.
func Validate(s Spec, v string) (string, error) {
	v = strings.TrimSpace(v)
	switch s.Kind {
	case KindText:
		if len([]rune(v)) > 500 {
			return "", fmt.Errorf("must be 500 characters or fewer")
		}
		return v, nil
	case KindInt:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return "", fmt.Errorf("must be a whole number")
		}
		if n < s.Min || n > s.Max {
			return "", fmt.Errorf("must be between %d and %d", s.Min, s.Max)
		}
		return strconv.FormatInt(n, 10), nil
	case KindDecimal:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", fmt.Errorf("must be a number such as -6.2265")
		}
		if f < float64(s.Min) || f > float64(s.Max) {
			return "", fmt.Errorf("must be between %d and %d", s.Min, s.Max)
		}
		return v, nil
	case KindEmail:
		a, err := mail.ParseAddress(v)
		if err != nil || a.Address != v {
			return "", fmt.Errorf("must be one email address, such as team@sunshinefacility.com")
		}
		return v, nil
	case KindEmailList:
		var out []string
		for part := range strings.SplitSeq(v, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			a, err := mail.ParseAddress(part)
			if err != nil || a.Address != part {
				return "", fmt.Errorf("%q is not an email address", part)
			}
			out = append(out, part)
		}
		if len(out) == 0 {
			return "", fmt.Errorf("needs at least one email address")
		}
		return strings.Join(out, ", "), nil
	case KindURL:
		if v == "" {
			return "", nil
		}
		u, err := url.Parse(v)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return "", fmt.Errorf("must be a full https:// address")
		}
		return v, nil
	case KindColour:
		v = strings.ToUpper(v)
		if !colourRe.MatchString(v) {
			return "", fmt.Errorf("must be a colour such as #C6241C")
		}
		return v, nil
	case KindGA4:
		if v == "" {
			return "", nil
		}
		if !ga4Re.MatchString(v) {
			return "", fmt.Errorf("must look like G-XXXXXXXXXX, or be empty for no analytics")
		}
		return v, nil
	case KindSlug:
		if !slugRe.MatchString(v) {
			return "", fmt.Errorf("must be a media slug such as photo/city-aerial")
		}
		return v, nil
	}
	return "", fmt.Errorf("unknown parameter type %q", s.Kind)
}

// RecipientList splits a stored email list.
func RecipientList(v string) []string {
	var out []string
	for p := range strings.SplitSeq(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// WhatsAppDigits turns "+62 1234 5678" into "6212345678" for a wa.me link.
func WhatsAppDigits(v string) string {
	var b strings.Builder
	for _, r := range v {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
