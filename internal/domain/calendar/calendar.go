// Package calendar is the business-day logic. Every "today" in the product
// is the date in Asia/Jakarta, stated explicitly — never the server's local
// zone (CLAUDE.md §6). A server in UTC would otherwise flip a certificate to
// Expired seven hours late.
package calendar

import (
	"fmt"
	"time"
	// Embedded zone data: the business date must not depend on whether the
	// host has /usr/share/zoneinfo.
	_ "time/tzdata"
)

// Zone is the operating timezone.
var Zone = mustLoad("Asia/Jakarta")

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// The binary imports time/tzdata, so this only fails if that import
		// is removed. Failing loudly beats silently computing in UTC.
		panic(fmt.Sprintf("calendar: cannot load %s: %v", name, err))
	}
	return loc
}

// Date is a civil date with no time and no zone.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// Today is the civil date in Asia/Jakarta at the instant now.
func Today(now time.Time) Date {
	y, m, d := now.In(Zone).Date()
	return Date{y, m, d}
}

// ParseDate accepts exactly YYYY-MM-DD and rejects impossible dates
// (2027-02-30) rather than normalising them.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return Date{}, fmt.Errorf("not a date in YYYY-MM-DD form: %q", s)
	}
	return Date{t.Year(), t.Month(), t.Day()}, nil
}

func FromTime(t time.Time) Date {
	y, m, d := t.Date()
	return Date{y, m, d}
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, int(d.Month), d.Day)
}

func (d Date) Time() time.Time { return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC) }

func (d Date) Before(o Date) bool { return d.Time().Before(o.Time()) }

func (d Date) AddDays(n int) Date { return FromTime(d.Time().AddDate(0, 0, n)) }

// DaysUntil is o − d in whole days.
func (d Date) DaysUntil(o Date) int { return int(o.Time().Sub(d.Time()).Hours() / 24) }

func (d Date) IsZero() bool { return d.Year == 0 }
