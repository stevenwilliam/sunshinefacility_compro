package params

import "testing"

func TestValidate(t *testing.T) {
	cases := []struct {
		key, v string
		ok     bool
		want   string
	}{
		{"careers.retention_months", "12", true, "12"},
		{"careers.retention_months", "twelve", false, ""},
		{"careers.retention_months", "0", false, ""},
		{"notify.quotation_to", "a@x.com, b@y.co.id", true, "a@x.com, b@y.co.id"},
		{"notify.quotation_to", "itdept", false, ""},
		{"notify.quotation_to", "", false, ""},
		{"notify.from", "noreply@sunshinefacility.com", true, "noreply@sunshinefacility.com"},
		{"social.instagram", "javascript:alert(1)", false, ""},
		{"social.instagram", "http://insecure.example", false, ""},
		{"social.instagram", "https://www.instagram.com/sunshine", true, "https://www.instagram.com/sunshine"},
		{"analytics.ga4_measurement_id", "", true, ""},
		{"analytics.ga4_measurement_id", "G-ABC123XYZ", true, "G-ABC123XYZ"},
		{"analytics.ga4_measurement_id", "UA-1234-1", false, ""},
		{"contact.map_lat", "-6.2265", true, "-6.2265"},
		{"contact.map_lat", "95", false, ""},
	}
	for _, c := range cases {
		grp, key := split(c.key)
		s, ok := SpecFor(grp, key)
		if !ok {
			t.Fatalf("no spec for %s", c.key)
		}
		got, err := Validate(s, c.v)
		if (err == nil) != c.ok || (c.ok && got != c.want) {
			t.Errorf("Validate(%s, %q) = (%q, %v), want ok=%v %q", c.key, c.v, got, err, c.ok, c.want)
		}
	}
}

func TestColourIsUppercased(t *testing.T) {
	s, _ := SpecFor("brand_colours", "anything")
	got, err := Validate(s, "#c6241c")
	if err != nil || got != "#C6241C" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestWhatsAppDigits(t *testing.T) {
	if got := WhatsAppDigits("+62 1234 5678"); got != "6212345678" {
		t.Fatalf("got %q", got)
	}
}

func split(k string) (string, string) {
	for i := range k {
		if k[i] == '.' {
			return k[:i], k[i+1:]
		}
	}
	return k, ""
}
