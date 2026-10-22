package forms

import (
	"strings"
	"testing"
)

var bt = []string{"office", "hospital"}
var lines = []string{"food", "digital", "workforce", "remote", "facility"}

func good() QuotationInput {
	return QuotationInput{Name: "Budi Santoso", Company: "PT Contoh", Email: "budi@contoh.co.id", Phone: "+62 812 3456 7890",
		BuildingType: "office", Area: "5000", Location: "Tangerang", Services: []string{"facility"}, Lang: "id"}
}

func TestBR5_1_ValidQuotation(t *testing.T) {
	q, errs := ValidateQuotation(good(), bt, lines)
	if errs != nil {
		t.Fatalf("unexpected errors %v", errs)
	}
	if *q.AreaM2 != 5000 || q.Lang != "id" || q.Services[0] != "facility" {
		t.Fatalf("got %+v", q)
	}
}

func TestBR5_1_RequiredFields(t *testing.T) {
	_, errs := ValidateQuotation(QuotationInput{}, bt, lines)
	for _, f := range []string{"name", "company", "email", "phone", "buildingType", "location", "services"} {
		if errs[f] != ErrRequired {
			t.Errorf("%s: got %q, want required", f, errs[f])
		}
	}
	if _, ok := errs["area"]; ok {
		t.Error("area is optional but was reported")
	}
	if _, ok := errs["message"]; ok {
		t.Error("message is optional but was reported")
	}
}

// Reject, do not repair (CLAUDE.md §7).
func TestBR5_1_AreaDigitsOnly(t *testing.T) {
	for _, a := range []string{"5,000", "5000 m2", "-5", "5.5"} {
		in := good()
		in.Area = a
		_, errs := ValidateQuotation(in, bt, lines)
		if errs["area"] != ErrDigits {
			t.Errorf("area %q: got %q, want digits", a, errs["area"])
		}
	}
	in := good()
	in.Area = "0"
	if _, errs := ValidateQuotation(in, bt, lines); errs["area"] != ErrRange {
		t.Errorf("area 0: got %q, want out_of_range", errs["area"])
	}
}

func TestBR5_1_ForgedKeysRefused(t *testing.T) {
	in := good()
	in.BuildingType = "spaceship"
	in.Services = []string{"facility", "<script>"}
	_, errs := ValidateQuotation(in, bt, lines)
	if errs["buildingType"] != ErrNotAllowed || errs["services"] != ErrNotAllowed {
		t.Fatalf("forged keys accepted: %v", errs)
	}
}

func TestBR5_1_LengthAndControl(t *testing.T) {
	in := good()
	in.Name = strings.Repeat("a", Rules.Name+1)
	in.Company = "PT\x00Evil"
	in.Email = "not-an-email"
	in.Phone = "call me"
	_, errs := ValidateQuotation(in, bt, lines)
	want := map[string]string{"name": ErrTooLong, "company": ErrControl, "email": ErrEmail, "phone": ErrPhone}
	for f, code := range want {
		if errs[f] != code {
			t.Errorf("%s: got %q, want %q", f, errs[f], code)
		}
	}
}

func TestBR6_1_Application(t *testing.T) {
	in := ApplicationInput{Name: "Sari", Phone: "0812-3456-7890", Email: "sari@example.com", Position: OpenApplication}
	a, errs := ValidateApplication(in, []string{"abc"})
	if errs != nil || a.OpeningID != "" {
		t.Fatalf("open application refused: %v", errs)
	}
	in.Position = "abc"
	if a, errs = ValidateApplication(in, []string{"abc"}); errs != nil || a.OpeningID != "abc" {
		t.Fatalf("published opening refused: %v", errs)
	}
	in.Position = "closed-or-forged"
	if _, errs = ValidateApplication(in, []string{"abc"}); errs["position"] != ErrNotAllowed {
		t.Fatalf("unknown opening accepted")
	}
}

func TestFirstName(t *testing.T) {
	if FirstName("  Budi Santoso ") != "Budi" || FirstName("") != "" {
		t.Fatal("FirstName")
	}
}
