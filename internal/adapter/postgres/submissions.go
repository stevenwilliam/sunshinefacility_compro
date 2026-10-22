package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
)

func (s *Store) InsertQuotation(ctx context.Context, q app.QuotationRecord) error {
	return translate(s.q(ctx).Exec(`INSERT INTO quotation_requests (id, created_at, name, company, email, phone, building_type_key, building_type,
		area_m2, location, services, message, lang, ip, user_agent) VALUES (?::uuid, ?, ?, ?, ?, ?, ?, ?, ?, ?, ARRAY(SELECT json_array_elements_text(?::json)), ?, ?, ?, ?)`,
		q.ID, q.CreatedAt, q.Name, q.Company, q.Email, q.Phone, q.BuildingTypeKey, q.BuildingType, q.AreaM2, q.Location,
		jsonArray(q.Services), q.Message, q.Lang, q.IP, q.UserAgent).Error)
}

func (s *Store) InsertApplication(ctx context.Context, a app.ApplicationRecord) error {
	return translate(s.q(ctx).Exec(`INSERT INTO career_applications (id, created_at, name, phone, email, position, opening_id, message,
		cv_key, cv_filename, cv_mime, cv_bytes, lang, ip, user_agent) VALUES (?::uuid, ?, ?, ?, ?, ?, ?::uuid, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.CreatedAt, a.Name, a.Phone, a.Email, a.Position, nullStr(a.OpeningID), a.Message,
		a.CVKey, a.CVFilename, a.CVMime, a.CVBytes, a.Lang, a.IP, a.UserAgent).Error)
}

// CountRecentSubmissions backs the per-IP and per-email limits (BR-5.4).
// Counting stored rows rather than an in-memory bucket means a restart does
// not reset anybody's allowance. Table and column are fixed by the caller.
func (s *Store) CountRecentSubmissions(ctx context.Context, table, column, value string, since time.Time) (int, error) {
	if (table != "quotation_requests" && table != "career_applications") || (column != "ip" && column != "email") {
		return 0, fmt.Errorf("CountRecentSubmissions: %s.%s is not allowed", table, column) // a programming error, never request input
	}
	var n int
	err := s.row(ctx, `SELECT count(*) FROM `+table+` WHERE `+column+` = ? AND created_at >= ?`, value, since).Scan(&n)
	return n, translate(err)
}

const quoteCols = `id::text, created_at, name, company, email, phone, building_type_key, building_type, area_m2, location,
	array_to_json(services)::text, message, lang`

func scanQuote(sc interface{ Scan(...any) error }, total *int) (*app.QuotationRecord, error) {
	var q app.QuotationRecord
	var area sql.NullInt64
	var services string
	dest := []any{&q.ID, &q.CreatedAt, &q.Name, &q.Company, &q.Email, &q.Phone, &q.BuildingTypeKey, &q.BuildingType, &area, &q.Location, &services, &q.Message, &q.Lang}
	if total != nil {
		dest = append(dest, total)
	}
	if err := sc.Scan(dest...); err != nil {
		return nil, err
	}
	if area.Valid {
		n := int(area.Int64)
		q.AreaM2 = &n
	}
	q.Services = jsonStrings(services)
	return &q, nil
}

func (s *Store) ListQuotations(ctx context.Context, lq app.ListQuery) ([]app.QuotationRecord, int, error) {
	where, args := searchClause([]string{"name", "company", "email", "phone", "building_type", "location", "message", "array_to_string(services, ' ')"}, lq.Q)
	if v := lq.Filters["building_type_key"]; v != "" {
		where += " AND building_type_key = ?"
		args = append(args, v)
	}
	if v := lq.Filters["service"]; v != "" {
		where += " AND ? = ANY(services)"
		args = append(args, v)
	}
	limit, offset := page(lq)
	args = append(args, limit, offset)
	rows, err := s.q(ctx).Raw(`SELECT `+quoteCols+`, count(*) OVER () FROM quotation_requests WHERE `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, args...).Rows()
	if err != nil {
		return nil, 0, translate(err)
	}
	defer rows.Close()
	out := []app.QuotationRecord{}
	total := 0
	for rows.Next() {
		q, err := scanQuote(rows, &total)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *q)
	}
	return out, total, rows.Err()
}

func (s *Store) GetQuotation(ctx context.Context, qid string) (*app.QuotationRecord, error) {
	if _, err := id.Parse(qid); err != nil {
		return nil, app.ErrNotFound
	}
	q, err := scanQuote(s.row(ctx, `SELECT `+quoteCols+` FROM quotation_requests WHERE id = ?::uuid`, qid), nil)
	return q, translate(err)
}

const appCols = `id::text, created_at, name, phone, email, position, COALESCE(opening_id::text, ''), message, cv_key, cv_filename, cv_mime, cv_bytes, lang`

func scanApp(sc interface{ Scan(...any) error }, total *int) (*app.ApplicationRecord, error) {
	var a app.ApplicationRecord
	dest := []any{&a.ID, &a.CreatedAt, &a.Name, &a.Phone, &a.Email, &a.Position, &a.OpeningID, &a.Message, &a.CVKey, &a.CVFilename, &a.CVMime, &a.CVBytes, &a.Lang}
	if total != nil {
		dest = append(dest, total)
	}
	if err := sc.Scan(dest...); err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) ListApplications(ctx context.Context, lq app.ListQuery) ([]app.ApplicationRecord, int, error) {
	where, args := searchClause([]string{"name", "phone", "email", "position", "message"}, lq.Q)
	limit, offset := page(lq)
	args = append(args, limit, offset)
	rows, err := s.q(ctx).Raw(`SELECT `+appCols+`, count(*) OVER () FROM career_applications WHERE `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, args...).Rows()
	if err != nil {
		return nil, 0, translate(err)
	}
	defer rows.Close()
	out := []app.ApplicationRecord{}
	total := 0
	for rows.Next() {
		a, err := scanApp(rows, &total)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *a)
	}
	return out, total, rows.Err()
}

func (s *Store) GetApplication(ctx context.Context, aid string) (*app.ApplicationRecord, error) {
	if _, err := id.Parse(aid); err != nil {
		return nil, app.ErrNotFound
	}
	a, err := scanApp(s.row(ctx, `SELECT `+appCols+` FROM career_applications WHERE id = ?::uuid`, aid), nil)
	return a, translate(err)
}

func (s *Store) ApplicationsOlderThan(ctx context.Context, cutoff time.Time) ([]app.ApplicationRecord, error) {
	rows, err := s.q(ctx).Raw(`SELECT `+appCols+` FROM career_applications WHERE created_at < ? ORDER BY created_at`, cutoff).Rows()
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	var out []app.ApplicationRecord
	for rows.Next() {
		a, err := scanApp(rows, nil)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (s *Store) DeleteApplication(ctx context.Context, aid string) error {
	res := s.q(ctx).Exec(`DELETE FROM career_applications WHERE id = ?::uuid`, aid)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected != 1 {
		return app.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Outbox
// ---------------------------------------------------------------------------

func (s *Store) EnqueueMail(ctx context.Context, m app.OutboxMail) error {
	return translate(s.q(ctx).Exec(`INSERT INTO notification_outbox (id, kind, ref_id, recipients, subject, body, next_attempt_at)
		VALUES (?::uuid, ?, ?, ARRAY(SELECT json_array_elements_text(?::json)), ?, ?, ?)`, m.ID, m.Kind, m.RefID, jsonArray(m.Recipients), m.Subject, m.Body, m.NextAt).Error)
}

const outboxCols = `id::text, created_at, kind, ref_id, array_to_json(recipients)::text, subject, body, attempts, next_attempt_at, sent_at, last_error`

func scanOutbox(sc interface{ Scan(...any) error }) (*app.OutboxMail, error) {
	var m app.OutboxMail
	var rcpt string
	var sent sql.NullTime
	if err := sc.Scan(&m.ID, &m.CreatedAt, &m.Kind, &m.RefID, &rcpt, &m.Subject, &m.Body, &m.Attempts, &m.NextAt, &sent, &m.LastError); err != nil {
		return nil, err
	}
	m.Recipients = jsonStrings(rcpt)
	if sent.Valid {
		m.SentAt = &sent.Time
	}
	return &m, nil
}

// DueMail claims mail whose time has come. "Now" is the application's
// clock, the same one the back-off was computed from.
func (s *Store) DueMail(ctx context.Context, limit int, now time.Time) ([]app.OutboxMail, error) {
	rows, err := s.q(ctx).Raw(`SELECT `+outboxCols+` FROM notification_outbox WHERE sent_at IS NULL AND next_attempt_at <= ?
		ORDER BY next_attempt_at LIMIT ? FOR UPDATE SKIP LOCKED`, now, limit).Rows()
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	var out []app.OutboxMail
	for rows.Next() {
		m, err := scanOutbox(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (s *Store) MarkMailSent(ctx context.Context, mid string) error {
	return translate(s.q(ctx).Exec(`UPDATE notification_outbox SET sent_at = now(), attempts = attempts + 1, last_error = '' WHERE id = ?::uuid`, mid).Error)
}

func (s *Store) MarkMailFailed(ctx context.Context, mid string, errText string, next time.Time) error {
	return translate(s.q(ctx).Exec(`UPDATE notification_outbox SET attempts = attempts + 1, last_error = ?, next_attempt_at = ? WHERE id = ?::uuid`, errText, next, mid).Error)
}

func (s *Store) ListOutbox(ctx context.Context, limit int) ([]app.OutboxMail, error) {
	rows, err := s.q(ctx).Raw(`SELECT `+outboxCols+` FROM notification_outbox ORDER BY created_at DESC LIMIT ?`, limit).Rows()
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	out := []app.OutboxMail{}
	for rows.Next() {
		m, err := scanOutbox(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Certificates
// ---------------------------------------------------------------------------

func (s *Store) CertificatesForReminder(ctx context.Context) ([]app.CertificateRecord, error) {
	rows, err := s.q(ctx).Raw(`SELECT id::text, name, COALESCE(to_char(valid_until, 'YYYY-MM-DD'), ''), published
		FROM certificates WHERE published AND valid_until IS NOT NULL ORDER BY valid_until`).Rows()
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	var out []app.CertificateRecord
	for rows.Next() {
		var c app.CertificateRecord
		if err := rows.Scan(&c.ID, &c.Name, &c.ValidUntil, &c.Published); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// RecordCertificateReminder claims the (certificate, date) slot. It returns
// false when the reminder for that date was already sent (BR-3.4).
func (s *Store) RecordCertificateReminder(ctx context.Context, certID, validUntil string) (bool, error) {
	res := s.q(ctx).Exec(`INSERT INTO certificate_reminders (certificate_id, valid_until) VALUES (?::uuid, ?::date) ON CONFLICT DO NOTHING`, certID, validUntil)
	if res.Error != nil {
		return false, translate(res.Error)
	}
	return res.RowsAffected == 1, nil
}

func jsonStrings(s string) []string {
	var out []string
	_ = json.Unmarshal([]byte(s), &out)
	if out == nil {
		out = []string{}
	}
	return out
}

// jsonArray passes a []string as JSON; the SQL turns it into text[]. (gorm
// would otherwise expand a slice into a row constructor.)
func jsonArray(v []string) string {
	if v == nil {
		v = []string{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}
