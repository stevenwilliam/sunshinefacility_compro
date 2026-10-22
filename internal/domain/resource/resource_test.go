package resource

import (
	"regexp"
	"strings"
	"testing"
)

// Identifiers in the SQL come from these definitions, so they must be plain
// identifiers. This is what makes building SQL from them safe.
func TestIdentifiersArePlain(t *testing.T) {
	ident := regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
	for _, d := range Defs {
		if !ident.MatchString(d.Table) || !ident.MatchString(d.IDColumn) {
			t.Errorf("%s: table/id column not a plain identifier", d.Name)
		}
		for _, f := range d.Fields {
			if !ident.MatchString(f.Key) {
				t.Errorf("%s.%s: not a plain identifier", d.Name, f.Key)
			}
		}
		for _, fl := range d.Filters {
			if !d.Has(fl.Key) {
				t.Errorf("%s: filter %s is not a field", d.Name, fl.Key)
			}
		}
		if d.NameField != "" && !d.Has(d.NameField) {
			t.Errorf("%s: name field %s is not a field", d.Name, d.NameField)
		}
		for _, f := range d.Fields {
			if f.Type == Refs {
				if _, ok := d.RefJoin[f.Key]; !ok {
					t.Errorf("%s.%s: refs field without a join table", d.Name, f.Key)
				}
			}
		}
	}
}

// BR-1.1: every bilingual pair is required together or optional together.
func TestBR1_1_PairsAgree(t *testing.T) {
	for _, d := range Defs {
		for _, f := range d.Fields {
			if !strings.HasSuffix(f.Key, "_en") {
				continue
			}
			twin, ok := d.Field(strings.TrimSuffix(f.Key, "_en") + "_id")
			if !ok {
				continue
			}
			if twin.Required != f.Required || twin.Type != f.Type || twin.Max != f.Max {
				t.Errorf("%s: %s and %s disagree", d.Name, f.Key, twin.Key)
			}
		}
	}
}

func TestValidate(t *testing.T) {
	d, _ := Lookup("testimonials")
	_, errs := Validate(d, map[string]any{}, true)
	for _, k := range []string{"name", "company", "quote_en", "quote_id"} {
		if errs[k] == "" {
			t.Errorf("missing required %s not reported", k)
		}
	}
	vals, errs := Validate(d, map[string]any{"name": "  Budi  ", "company": "PT X", "quote_en": "Good", "quote_id": "Bagus",
		"published": true, "sort": "3"}, true)
	if errs != nil {
		t.Fatalf("unexpected %v", errs)
	}
	if vals["name"] != "Budi" || vals["sort"] != int64(3) || vals["published"] != true {
		t.Fatalf("normalisation: %v", vals)
	}
}

func TestImmutableRefusedOnUpdate(t *testing.T) {
	d, _ := Lookup("services")
	_, errs := Validate(d, map[string]any{"key": "new-key"}, false)
	if errs["key"] == "" {
		t.Fatal("changing a service key was allowed")
	}
}

func TestLinesNormalised(t *testing.T) {
	d, _ := Lookup("service-groups")
	vals, errs := Validate(d, map[string]any{"items_en": " one \n\n  two\r\n"}, false)
	if errs != nil || vals["items_en"] != "one\ntwo" {
		t.Fatalf("got %q %v", vals["items_en"], errs)
	}
}

func TestSingleLineText(t *testing.T) {
	d, _ := Lookup("clients")
	if _, errs := Validate(d, map[string]any{"name": "a\nb"}, false); errs["name"] == "" {
		t.Fatal("a newline in a single-line field was accepted")
	}
}
