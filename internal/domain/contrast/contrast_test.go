package contrast

import (
	"math"
	"testing"
)

// The README states these ratios. If the arithmetic disagrees with the
// design sheet, one of them is wrong — this test says which.
func TestBR11_DesignRatios(t *testing.T) {
	cases := []struct {
		fg, bg string
		want   float64
		passAA bool
	}{
		{"#FFFFFF", "#C6241C", 5.7, true},  // Strong Red button
		{"#1F211C", "#FD572A", 5.0, true},  // Remote: charcoal on tangerine
		{"#FFFFFF", "#EA2915", 4.4, false}, // white on Signal Red fails — why text sits on ink
		{"#FFFFFF", "#28937A", 3.8, false}, // white on Mint fails
		{"#FFFFFF", "#1F7360", 0, true},    // Digital ink
		{"#FFFFFF", "#4A7023", 0, true},    // Workforce ink
		{"#FFFFFF", "#A3181A", 0, true},    // Facility ink
	}
	for _, c := range cases {
		ok, r, err := MeetsAA(c.fg, c.bg)
		if err != nil {
			t.Fatal(err)
		}
		if c.want > 0 && math.Abs(r-c.want) > 0.1 {
			t.Errorf("%s on %s = %.2f, README says %.1f", c.fg, c.bg, r, c.want)
		}
		if ok != c.passAA {
			t.Errorf("%s on %s = %.2f, AA pass = %v, want %v", c.fg, c.bg, r, ok, c.passAA)
		}
	}
}

func TestRejectsMalformed(t *testing.T) {
	for _, bad := range []string{"", "#FFF", "red", "#GGGGGG", "#1234567"} {
		if _, err := Luminance(bad); err == nil {
			t.Errorf("Luminance(%q) accepted a malformed colour", bad)
		}
	}
}
