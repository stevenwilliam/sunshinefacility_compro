// Package certificate holds the certificate rules (02-business-rules.md BR-3).
//
// A regulated claim the company cannot back is worse than no claim. The
// database trigger refuses to publish an expired certificate; this package is
// the same rule for the API's answer and the CMS's disabled toggle, so the
// three agree.
package certificate

import "github.com/stevenwilliam/sunshinefacility_compro/internal/domain/calendar"

type Status string

const (
	StatusValid    Status = "valid"
	StatusNoExpiry Status = "no_expiry"
	StatusExpired  Status = "expired"
)

// StatusOn — BR-3.1: a certificate whose valid-until date is before today
// (Asia/Jakarta) is Expired. Valid ON its last day: "valid until 30 Sep"
// includes 30 Sep.
func StatusOn(validUntil *calendar.Date, today calendar.Date) Status {
	if validUntil == nil {
		return StatusNoExpiry
	}
	if validUntil.Before(today) {
		return StatusExpired
	}
	return StatusValid
}

// CanPublish — BR-3.2.
func CanPublish(validUntil *calendar.Date, today calendar.Date) bool {
	return StatusOn(validUntil, today) != StatusExpired
}

// PubliclyVisible — BR-3.3: a published certificate that has since expired
// disappears from the site on its own; nobody has to remember to unpublish it.
func PubliclyVisible(published bool, validUntil *calendar.Date, today calendar.Date) bool {
	return published && CanPublish(validUntil, today)
}

// ReminderDue — BR-3.4: the reminder goes out once the expiry is within
// `days` days, and not after expiry (an expired certificate is already
// hidden; the CMS shows it as Expired). Whether it was already sent for this
// date is the caller's ledger, not this function's.
func ReminderDue(validUntil *calendar.Date, today calendar.Date, days int) bool {
	if validUntil == nil || days < 0 {
		return false
	}
	left := today.DaysUntil(*validUntil)
	return left >= 0 && left <= days
}
