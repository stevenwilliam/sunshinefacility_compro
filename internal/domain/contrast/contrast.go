// Package contrast measures WCAG 2.x contrast. A colour pairing that carries
// text is a number, not an opinion (impeccable §3).
//
// Floating point is fine here: this is colour arithmetic, nowhere near money.
package contrast

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// AA is the minimum ratio for body text (WCAG 1.4.3).
const AA = 4.5

func channel(c uint8) float64 {
	v := float64(c) / 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// Luminance of #RRGGBB.
func Luminance(hex string) (float64, error) {
	h := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(h) != 6 {
		return 0, fmt.Errorf("not a #RRGGBB colour: %q", hex)
	}
	n, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("not a #RRGGBB colour: %q", hex)
	}
	r, g, b := uint8(n>>16), uint8(n>>8), uint8(n)
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b), nil
}

// Ratio is (L1 + .05) / (L2 + .05) with L1 the lighter.
func Ratio(a, b string) (float64, error) {
	la, err := Luminance(a)
	if err != nil {
		return 0, err
	}
	lb, err := Luminance(b)
	if err != nil {
		return 0, err
	}
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05), nil
}

// MeetsAA — BR-11.1: the text colour on a line's ink field must reach 4.5:1.
// An editor choosing a pretty pair that nobody can read is refused, with the
// number, rather than shipped.
func MeetsAA(fg, bg string) (bool, float64, error) {
	r, err := Ratio(fg, bg)
	if err != nil {
		return false, 0, err
	}
	return r >= AA, r, nil
}
