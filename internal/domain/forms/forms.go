// Package forms validates the two public forms (BR-5, BR-6).
//
// The browser validates for feedback; this package validates because the
// browser can be bypassed (CLAUDE.md §7). The limits below are ALSO sent to
// the browser in the public bundle (Rules), so the two sides read one source
// instead of two hand-written copies.
//
// Errors are machine codes per field. The site renders them in the visitor's
// language from its catalogue (err.required, err.email, …).
package forms

import (
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/sanitize"
)

// Field error codes.
const (
	ErrRequired   = "required"
	ErrTooLong    = "too_long"
	ErrEmail      = "email"
	ErrNotAllowed = "not_allowed"
	ErrDigits     = "digits"
	ErrRange      = "out_of_range"
	ErrControl    = "control"
	ErrPhone      = "phone"
)

// Limits, in characters.
type Limits struct {
	Name     int `json:"name"`
	Company  int `json:"company"`
	Email    int `json:"email"`
	Phone    int `json:"phone"`
	Location int `json:"location"`
	Message  int `json:"message"`
	AreaMax  int `json:"areaMax"`
}

var Rules = Limits{Name: 120, Company: 160, Email: 254, Phone: 30, Location: 200, Message: 2000, AreaMax: 100_000_000}

// A phone number as people type it: digits, spaces, + - ( ) and dots, with
// at least six digits. Not a full E.164 parse — Indonesian numbers arrive as
// 0812…, +62 812…, (021) 1234 5678, and all of those are fine.
var phoneRe = regexp.MustCompile(`^[0-9+()\-. ]+$`)

type Errors map[string]string

func (e Errors) any() bool { return len(e) > 0 }

func text(errs Errors, field, v string, max int, required bool) string {
	var out string
	var err error
	if required {
		out, err = sanitize.Text(v, max)
	} else {
		out, err = sanitize.Optional(v, max)
	}
	switch {
	case err == nil:
		return out
	case errors.Is(err, sanitize.ErrEmpty):
		errs[field] = ErrRequired
	case errors.Is(err, sanitize.ErrTooLong):
		errs[field] = ErrTooLong
	default:
		errs[field] = ErrControl
	}
	return ""
}

func email(errs Errors, v string) string {
	out, err := sanitize.Email(v)
	switch {
	case err == nil:
		return out
	case errors.Is(err, sanitize.ErrEmpty):
		errs["email"] = ErrRequired
	case errors.Is(err, sanitize.ErrTooLong):
		errs["email"] = ErrTooLong
	default:
		errs["email"] = ErrEmail
	}
	return ""
}

func phone(errs Errors, v string) string {
	p := text(errs, "phone", v, Rules.Phone, true)
	if p == "" {
		return ""
	}
	digits := 0
	for _, r := range p {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	if !phoneRe.MatchString(p) || digits < 6 {
		errs["phone"] = ErrPhone
		return ""
	}
	return p
}

// QuotationInput is what the visitor sent.
type QuotationInput struct {
	Name         string   `json:"name"`
	Company      string   `json:"company"`
	Email        string   `json:"email"`
	Phone        string   `json:"phone"`
	BuildingType string   `json:"buildingType"`
	Area         string   `json:"area"`
	Location     string   `json:"location"`
	Services     []string `json:"services"`
	Message      string   `json:"message"`
	Lang         string   `json:"lang"`
}

type Quotation struct {
	Name, Company, Email, Phone string
	BuildingType                string
	AreaM2                      *int
	Location                    string
	Services                    []string
	Message                     string
	Lang                        string
}

// ValidateQuotation — BR-5.1. buildingTypes and lines are the allowed keys
// (the published building types and service lines); anything else is
// refused, so a forged key never reaches the inbox.
func ValidateQuotation(in QuotationInput, buildingTypes, lines []string) (Quotation, Errors) {
	errs := Errors{}
	q := Quotation{}
	q.Name = text(errs, "name", in.Name, Rules.Name, true)
	q.Company = text(errs, "company", in.Company, Rules.Company, true)
	q.Email = email(errs, in.Email)
	q.Phone = phone(errs, in.Phone)
	q.Location = text(errs, "location", in.Location, Rules.Location, true)
	q.Message = text(errs, "message", in.Message, Rules.Message, false)

	bt := strings.TrimSpace(in.BuildingType)
	switch {
	case bt == "":
		errs["buildingType"] = ErrRequired
	case !contains(buildingTypes, bt):
		errs["buildingType"] = ErrNotAllowed
	default:
		q.BuildingType = bt
	}

	// Area: digits only, optional (BR-5.1). "5,000" is refused, not
	// repaired — the visitor is told digits only.
	if a := strings.TrimSpace(in.Area); a != "" {
		if !allDigits(a) {
			errs["area"] = ErrDigits
		} else if n, err := strconv.Atoi(a); err != nil || n < 1 || n > Rules.AreaMax {
			errs["area"] = ErrRange
		} else {
			q.AreaM2 = &n
		}
	}

	seen := map[string]bool{}
	for _, s := range in.Services {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		if !contains(lines, s) {
			errs["services"] = ErrNotAllowed
			break
		}
		seen[s] = true
		q.Services = append(q.Services, s)
	}
	if len(q.Services) == 0 && errs["services"] == "" {
		errs["services"] = ErrRequired
	}

	q.Lang = lang(in.Lang)
	if errs.any() {
		return Quotation{}, errs
	}
	return q, nil
}

// ApplicationInput — the text half of a job application. The CV is checked
// separately (type by magic bytes, size against careers.cv_max_mb).
type ApplicationInput struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Position string `json:"position"` // an opening id, or "open"
	Message  string `json:"message"`
	Lang     string `json:"lang"`
}

type Application struct {
	Name, Phone, Email string
	OpeningID          string // empty for an open application
	Message            string
	Lang               string
}

// OpenApplication is the position value for "Open application".
const OpenApplication = "open"

// ValidateApplication — BR-6.1. openings are the ids of published openings.
func ValidateApplication(in ApplicationInput, openings []string) (Application, Errors) {
	errs := Errors{}
	a := Application{}
	a.Name = text(errs, "name", in.Name, Rules.Name, true)
	a.Phone = phone(errs, in.Phone)
	a.Email = email(errs, in.Email)
	a.Message = text(errs, "message", in.Message, Rules.Message, false)
	p := strings.TrimSpace(in.Position)
	switch {
	case p == "":
		errs["position"] = ErrRequired
	case p == OpenApplication:
	case contains(openings, p):
		a.OpeningID = p
	default:
		errs["position"] = ErrNotAllowed
	}
	a.Lang = lang(in.Lang)
	if errs.any() {
		return Application{}, errs
	}
	return a, nil
}

func lang(s string) string {
	if s == "id" {
		return "id"
	}
	return "en"
}

func contains(list []string, v string) bool { return slices.Contains(list, v) }

func allDigits(s string) bool {
	if s == "" || !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// FirstName for the thank-you line ("Thank you, Budi."). The design takes
// the first word of the full name.
func FirstName(full string) string {
	f := strings.Fields(full)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}
