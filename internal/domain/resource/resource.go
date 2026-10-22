// Package resource declares every CMS-editable entity in one place: its
// table, its fields, their types and limits, which columns are content
// (BR-2.2), who may write it, and whether rows can be added or removed.
//
// One declaration drives four things that would otherwise drift apart: the
// server-side validation, the SQL the repository builds (identifiers come
// from here, never from a request), the admin forms the CMS renders, and the
// columns of the CSV export. The CMS fetches these definitions from the API
// instead of keeping its own copy (CLAUDE.md §7, "same rules, one source").
package resource

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/calendar"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/sanitize"
)

type FieldType string

const (
	Text     FieldType = "text"
	Textarea FieldType = "textarea"
	Lines    FieldType = "lines" // textarea, one item per line
	Int      FieldType = "int"
	Bool     FieldType = "bool"
	Date     FieldType = "date"
	Media    FieldType = "media"
	Colour   FieldType = "colour"
	Select   FieldType = "select" // static options
	Param    FieldType = "param"  // options from a parameter group
	Ref      FieldType = "ref"    // a row of another resource
	Refs     FieldType = "refs"   // several rows of another resource (join table)
	Icon     FieldType = "icon"
)

type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type Field struct {
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Type      FieldType `json:"type"`
	Required  bool      `json:"required,omitempty"`
	Max       int       `json:"max,omitempty"`    // characters, for text types
	MinInt    int64     `json:"minInt,omitempty"` // for Int
	MaxInt    int64     `json:"maxInt,omitempty"`
	Pattern   string    `json:"pattern,omitempty"`
	Options   []Option  `json:"options,omitempty"`
	Group     string    `json:"group,omitempty"`     // Param
	Ref       string    `json:"ref,omitempty"`       // Ref / Refs: resource name
	MediaKind string    `json:"mediaKind,omitempty"` // Media: picker filter
	// Content fields: a change clears the placeholder flag (BR-2.2).
	Content bool `json:"content,omitempty"`
	// List: shown as a grid column. Search: matched by the search box.
	List   bool `json:"list,omitempty"`
	Search bool `json:"search,omitempty"`
	// Fixed once the row exists (a key used in URLs, a section name the
	// page layout depends on). Settable on create.
	Immutable bool   `json:"immutable,omitempty"`
	Help      string `json:"help,omitempty"`
}

type Filter struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	Options []Option `json:"options,omitempty"`
	Ref     string   `json:"ref,omitempty"`
	Group   string   `json:"group,omitempty"`
}

type Def struct {
	Name        string            `json:"name"`
	Table       string            `json:"-"`
	IDColumn    string            `json:"idColumn"`
	Label       string            `json:"label"`
	One         string            `json:"one"`
	Description string            `json:"description"`
	Icon        string            `json:"icon"`
	Nav         string            `json:"nav"` // sidebar group
	WritePerm   access.Permission `json:"writePerm"`
	Fields      []Field           `json:"fields"`
	Filters     []Filter          `json:"filters,omitempty"`
	OrderBy     string            `json:"-"`
	NameField   string            `json:"nameField"`
	CanCreate   bool              `json:"canCreate"`
	CreateWhy   string            `json:"createWhy,omitempty"`
	CanDelete   bool              `json:"canDelete"`
	DeleteWhy   string            `json:"deleteWhy,omitempty"`
	// RefJoin describes a Refs field's join table.
	RefJoin map[string]Join `json:"-"`
}

type Join struct {
	Table, OwnerCol, RefCol string
}

func (d Def) Field(key string) (Field, bool) {
	for _, f := range d.Fields {
		if f.Key == key {
			return f, true
		}
	}
	return Field{}, false
}

func (d Def) Has(key string) bool { _, ok := d.Field(key); return ok }

// ContentFields lists the columns whose change clears the placeholder flag.
func (d Def) ContentFields() []string {
	var out []string
	for _, f := range d.Fields {
		if f.Content {
			out = append(out, f.Key)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// Value is a validated field value: string, int64, bool, nil (NULL) or
// []string (Refs).
type Value = any

// Validate checks a submitted row against the definition. creating says
// whether this is a new row (Immutable fields may be set only then). Fields
// absent from raw are left untouched on update; on create, Required fields
// must be present.
//
// Errors are English sentences: the CMS is English-only (decision Q16).
func Validate(d Def, raw map[string]any, creating bool) (map[string]Value, map[string]string) {
	out := map[string]Value{}
	errs := map[string]string{}
	for _, f := range d.Fields {
		v, present := raw[f.Key]
		if !present {
			if creating && f.Required {
				errs[f.Key] = f.Label + " is required."
			}
			continue
		}
		if f.Immutable && !creating {
			// Silently ignoring it would be a repair; saying so is not.
			errs[f.Key] = f.Label + " cannot be changed once created."
			continue
		}
		val, msg := validateField(f, v)
		if msg != "" {
			errs[f.Key] = msg
			continue
		}
		out[f.Key] = val
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return out, nil
}

func str(v any) (string, bool) {
	switch x := v.(type) {
	case nil:
		return "", true
	case string:
		return x, true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(x), true
	}
	return "", false
}

var colourRe = regexp.MustCompile(`^#[0-9A-F]{6}$`)
var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func validateField(f Field, v any) (Value, string) {
	switch f.Type {
	case Bool:
		b, ok := v.(bool)
		if !ok {
			return nil, f.Label + " must be on or off."
		}
		return b, ""
	case Refs:
		arr, ok := v.([]any)
		if !ok {
			return nil, f.Label + " must be a list."
		}
		var ids []string
		seen := map[string]bool{}
		for _, x := range arr {
			s, ok := x.(string)
			if !ok || !uuidRe.MatchString(s) {
				return nil, f.Label + " contains an invalid choice."
			}
			if !seen[s] {
				seen[s] = true
				ids = append(ids, s)
			}
		}
		if f.Required && len(ids) == 0 {
			return nil, f.Label + " needs at least one choice."
		}
		return ids, ""
	}

	s, ok := str(v)
	if !ok {
		return nil, f.Label + " has the wrong type."
	}
	if strings.TrimSpace(s) == "" {
		if f.Required {
			return nil, f.Label + " is required."
		}
		switch f.Type {
		case Media, Ref, Date:
			return nil, ""
		case Int:
			return int64(0), ""
		}
		return "", ""
	}

	switch f.Type {
	case Text, Textarea, Lines, Param, Icon:
		max := f.Max
		if max == 0 {
			max = 500
		}
		clean, err := sanitize.Text(s, max)
		if err != nil {
			switch err {
			case sanitize.ErrTooLong:
				return nil, fmt.Sprintf("%s must be %d characters or fewer (it is %d).", f.Label, max, utf8.RuneCountInString(strings.TrimSpace(s)))
			case sanitize.ErrControl:
				return nil, f.Label + " contains characters that are not allowed."
			default:
				return nil, f.Label + " is not valid text."
			}
		}
		if f.Type == Text && strings.ContainsAny(clean, "\n\r") {
			return nil, f.Label + " must be a single line."
		}
		if f.Type == Lines {
			clean = normaliseLines(clean)
		}
		if f.Pattern != "" && !regexp.MustCompile(f.Pattern).MatchString(clean) {
			return nil, f.Label + " " + f.Help
		}
		if f.Type == Icon && !IconAllowed(clean) {
			return nil, f.Label + " must be one of the listed icons."
		}
		return clean, ""
	case Select:
		for _, o := range f.Options {
			if o.Value == s {
				return s, ""
			}
		}
		return nil, f.Label + " must be one of the listed options."
	case Int:
		n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return nil, f.Label + " must be a whole number."
		}
		if (f.MinInt != 0 || f.MaxInt != 0) && (n < f.MinInt || n > f.MaxInt) {
			return nil, fmt.Sprintf("%s must be between %d and %d.", f.Label, f.MinInt, f.MaxInt)
		}
		return n, ""
	case Date:
		dt, err := calendar.ParseDate(strings.TrimSpace(s))
		if err != nil {
			return nil, f.Label + " must be a date (YYYY-MM-DD)."
		}
		return dt.String(), ""
	case Colour:
		c := strings.ToUpper(strings.TrimSpace(s))
		if !colourRe.MatchString(c) {
			return nil, f.Label + " must be a colour such as #C6241C."
		}
		return c, ""
	case Media, Ref:
		id := strings.TrimSpace(s)
		if !uuidRe.MatchString(id) {
			return nil, f.Label + " is not a valid choice."
		}
		return id, ""
	}
	return nil, f.Label + " has an unknown type."
}

// normaliseLines trims every line and drops blank ones: a list item that is
// only whitespace would render as an empty bullet.
func normaliseLines(s string) string {
	var out []string
	for line := range strings.SplitSeq(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, "\n")
}
