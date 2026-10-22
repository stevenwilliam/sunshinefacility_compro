package certificate

import (
	"testing"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/calendar"
)

func d(s string) *calendar.Date {
	x, err := calendar.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return &x
}

func TestBR3_1_StatusBoundaries(t *testing.T) {
	today := *d("2026-10-22")
	cases := []struct {
		until *calendar.Date
		want  Status
	}{
		{nil, StatusNoExpiry},
		{d("2026-10-21"), StatusExpired},
		{d("2026-10-22"), StatusValid}, // valid on its last day
		{d("2027-09-30"), StatusValid},
		{d("2025-08-25"), StatusExpired}, // the halal certificate
	}
	for _, c := range cases {
		if got := StatusOn(c.until, today); got != c.want {
			t.Errorf("StatusOn(%v) = %s, want %s", c.until, got, c.want)
		}
	}
}

func TestBR3_2_ExpiredCannotPublish(t *testing.T) {
	today := *d("2026-10-22")
	if CanPublish(d("2025-08-25"), today) {
		t.Fatal("an expired certificate was allowed to publish")
	}
	if !CanPublish(nil, today) {
		t.Fatal("a certificate with no expiry was refused")
	}
}

// BR-3.1 uses the date in Jakarta. At 18:00 UTC on 22 Oct it is already
// 23 Oct in Jakarta, so a certificate valid until 22 Oct has expired there.
func TestBR3_1_TodayIsJakarta(t *testing.T) {
	now := time.Date(2026, 10, 22, 18, 0, 0, 0, time.UTC)
	today := calendar.Today(now)
	if today.String() != "2026-10-23" {
		t.Fatalf("today in Jakarta = %s, want 2026-10-23", today)
	}
	if StatusOn(d("2026-10-22"), today) != StatusExpired {
		t.Fatal("expiry was computed in UTC instead of Asia/Jakarta")
	}
}

func TestBR3_3_PublishedButExpiredIsHidden(t *testing.T) {
	today := *d("2026-10-22")
	if PubliclyVisible(true, d("2026-10-01"), today) {
		t.Fatal("an expired certificate stayed on the site")
	}
	if !PubliclyVisible(true, d("2027-01-01"), today) {
		t.Fatal("a valid published certificate was hidden")
	}
	if PubliclyVisible(false, d("2027-01-01"), today) {
		t.Fatal("an unpublished certificate was shown")
	}
}

func TestBR3_4_ReminderWindow(t *testing.T) {
	today := *d("2026-10-22")
	cases := []struct {
		until *calendar.Date
		days  int
		want  bool
	}{
		{d("2026-12-21"), 60, true},  // exactly 60 days out
		{d("2026-12-22"), 60, false}, // 61 days out
		{d("2026-10-22"), 60, true},  // last valid day
		{d("2026-10-21"), 60, false}, // already expired
		{nil, 60, false},
	}
	for _, c := range cases {
		if got := ReminderDue(c.until, today, c.days); got != c.want {
			t.Errorf("ReminderDue(%v, %d) = %v, want %v", c.until, c.days, got, c.want)
		}
	}
}
