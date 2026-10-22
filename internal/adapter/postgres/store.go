// Package postgres implements app.Store with explicit SQL through gorm's
// Raw/Exec. Every value is a bound parameter; the only identifiers ever
// spliced into SQL come from domain/resource definitions, which a test pins
// to plain identifiers.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
)

type Store struct{ db *gorm.DB }

func New(db *gorm.DB) *Store { return &Store{db: db} }

var _ app.Store = (*Store)(nil)

func (s *Store) InTx(ctx context.Context, fn func(app.Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Store{db: tx})
	})
}

func (s *Store) q(ctx context.Context) *gorm.DB { return s.db.WithContext(ctx) }

// translate maps a driver error to the API's error model. The driver's text
// is kept as the cause for the log and never shown.
func translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		return app.ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23001": // restrict_violation (ON DELETE RESTRICT)
			return apierror.Wrap(apierror.CodeInUse, "Can’t delete: other items still use this. Change or remove them first.", err)
		case "23503": // foreign_key_violation
			if strings.Contains(strings.ToLower(pg.Message), "update or delete") {
				return apierror.Wrap(apierror.CodeInUse, "Can’t delete: other items still use this. Change or remove them first.", err)
			}
			return apierror.Wrap(apierror.CodeValidation, "One of the chosen items no longer exists. Reload and choose again.", err)
		case "23505": // unique_violation
			return apierror.Wrap(apierror.CodeConflict, "Another item already uses this value.", err)
		case "23514": // check_violation
			if pg.ConstraintName == "certificates_expired_unpublished" {
				return apierror.Wrap(apierror.CodeExpired, "This certificate has expired and cannot be published.", err)
			}
			return apierror.Wrap(apierror.CodeValidation, "A value is outside what the system allows.", err)
		case "23502": // not_null_violation
			return apierror.Wrap(apierror.CodeValidation, "A required value is missing.", err)
		case "42501": // insufficient_privilege (append-only audit trigger)
			return apierror.Wrap(apierror.CodeForbidden, "This history cannot be changed.", err)
		}
	}
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func jsonOf(v any) any {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// Users and sessions
// ---------------------------------------------------------------------------

const userCols = `id::text, email, name, password_hash, role, active, failed_logins, locked_until, last_login_at, password_changed_at, created_at`

func scanUser(sc interface{ Scan(...any) error }) (*app.User, error) {
	var u app.User
	var locked, last sql.NullTime
	if err := sc.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role, &u.Active, &u.FailedLogins, &locked, &last, &u.PasswordChangedAt, &u.CreatedAt); err != nil {
		return nil, err
	}
	if locked.Valid {
		u.LockedUntil = &locked.Time
	}
	if last.Valid {
		u.LastLoginAt = &last.Time
	}
	return &u, nil
}

func (s *Store) row(ctx context.Context, query string, args ...any) *sql.Row {
	return s.q(ctx).Raw(query, args...).Row()
}

func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var n int
	err := s.row(ctx, `SELECT count(*) FROM users`).Scan(&n)
	return n, translate(err)
}

func (s *Store) CountActiveSuperadmins(ctx context.Context) (int, error) {
	var n int
	err := s.row(ctx, `SELECT count(*) FROM users WHERE role = 'superadmin' AND active`).Scan(&n)
	return n, translate(err)
}

func (s *Store) UserByEmail(ctx context.Context, email string) (*app.User, error) {
	u, err := scanUser(s.row(ctx, `SELECT `+userCols+` FROM users WHERE email = ?`, strings.ToLower(email)))
	return u, translate(err)
}

func (s *Store) UserByID(ctx context.Context, id string) (*app.User, error) {
	u, err := scanUser(s.row(ctx, `SELECT `+userCols+` FROM users WHERE id = ?::uuid`, id))
	return u, translate(err)
}

func (s *Store) ListUsers(ctx context.Context, q string) ([]app.User, error) {
	where, args := searchClause([]string{"email", "name", "role"}, q)
	rows, err := s.q(ctx).Raw(`SELECT `+userCols+` FROM users WHERE `+where+` ORDER BY name, email`, args...).Rows()
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	out := []app.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (s *Store) CreateUser(ctx context.Context, u app.User) error {
	return translate(s.q(ctx).Exec(`INSERT INTO users (id, email, name, password_hash, role, active, password_changed_at, created_at)
		VALUES (?::uuid, ?, ?, ?, ?, ?, ?, ?)`, u.ID, u.Email, u.Name, u.PasswordHash, u.Role, u.Active, u.PasswordChangedAt, u.CreatedAt).Error)
}

func (s *Store) UpdateUser(ctx context.Context, u app.User) error {
	return translate(s.q(ctx).Exec(`UPDATE users SET name = ?, role = ?, active = ?, updated_at = now() WHERE id = ?::uuid`,
		u.Name, u.Role, u.Active, u.ID).Error)
}

func (s *Store) SetPassword(ctx context.Context, userID, hash string) error {
	return translate(s.q(ctx).Exec(`UPDATE users SET password_hash = ?, password_changed_at = now(), failed_logins = 0,
		locked_until = NULL, updated_at = now() WHERE id = ?::uuid`, hash, userID).Error)
}

func (s *Store) RecordLoginFailure(ctx context.Context, userID string, lockUntil *time.Time) error {
	if lockUntil != nil {
		return translate(s.q(ctx).Exec(`UPDATE users SET failed_logins = 0, locked_until = ? WHERE id = ?::uuid`, *lockUntil, userID).Error)
	}
	return translate(s.q(ctx).Exec(`UPDATE users SET failed_logins = failed_logins + 1 WHERE id = ?::uuid`, userID).Error)
}

func (s *Store) RecordLoginSuccess(ctx context.Context, userID string) error {
	return translate(s.q(ctx).Exec(`UPDATE users SET failed_logins = 0, locked_until = NULL, last_login_at = now() WHERE id = ?::uuid`, userID).Error)
}

func (s *Store) CreateRefreshToken(ctx context.Context, t app.RefreshToken) error {
	return translate(s.q(ctx).Exec(`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, ip, user_agent)
		VALUES (?::uuid, ?::uuid, ?, ?, ?, ?)`, t.ID, t.UserID, t.TokenHash, t.ExpiresAt, t.IP, t.UserAgent).Error)
}

func (s *Store) RefreshTokenByHash(ctx context.Context, hash string) (*app.RefreshToken, error) {
	var t app.RefreshToken
	var revoked sql.NullTime
	var replaced sql.NullString
	err := s.row(ctx, `SELECT id::text, user_id::text, token_hash, expires_at, revoked_at, replaced_by::text FROM refresh_tokens WHERE token_hash = ?`, hash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &revoked, &replaced)
	if err != nil {
		return nil, translate(err)
	}
	if revoked.Valid {
		t.RevokedAt = &revoked.Time
	}
	t.ReplacedBy = replaced.String
	return &t, nil
}

func (s *Store) RevokeRefreshToken(ctx context.Context, id string, replacedBy string) error {
	return translate(s.q(ctx).Exec(`UPDATE refresh_tokens SET revoked_at = COALESCE(revoked_at, now()), replaced_by = ?::uuid WHERE id = ?::uuid`,
		nullStr(replacedBy), id).Error)
}

func (s *Store) RevokeUserRefreshTokens(ctx context.Context, userID string) error {
	return translate(s.q(ctx).Exec(`UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = ?::uuid AND revoked_at IS NULL`, userID).Error)
}

func (s *Store) RevokeJTI(ctx context.Context, jti string, until time.Time) error {
	if err := s.q(ctx).Exec(`DELETE FROM revoked_jti WHERE expires_at < now()`).Error; err != nil {
		return translate(err)
	}
	return translate(s.q(ctx).Exec(`INSERT INTO revoked_jti (jti, expires_at) VALUES (?, ?) ON CONFLICT DO NOTHING`, jti, until).Error)
}

func (s *Store) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	var n int
	err := s.row(ctx, `SELECT count(*) FROM revoked_jti WHERE jti = ?`, jti).Scan(&n)
	return n > 0, translate(err)
}

// ---------------------------------------------------------------------------
// Audit
// ---------------------------------------------------------------------------

func (s *Store) AppendAudit(ctx context.Context, e app.AuditEntry) error {
	return translate(s.q(ctx).Exec(`INSERT INTO audit_log (id, at, actor_id, actor_email, action, entity, entity_id, summary, before, after, ip, user_agent, request_id)
		VALUES (?::uuid, ?, ?::uuid, ?, ?, ?, ?, ?, ?::jsonb, ?::jsonb, ?, ?, ?)`,
		e.ID, e.At, nullStr(e.ActorID), e.ActorEmail, e.Action, e.Entity, e.EntityID, e.Summary,
		jsonOf(e.Before), jsonOf(e.After), e.IP, e.UserAgent, e.RequestID).Error)
}

func (s *Store) ListAudit(ctx context.Context, q app.ListQuery) ([]app.AuditEntry, int, error) {
	where, args := searchClause([]string{"actor_email", "action", "entity", "entity_id", "summary"}, q.Q)
	if v := q.Filters["entity"]; v != "" {
		where += " AND entity = ?"
		args = append(args, v)
	}
	if v := q.Filters["action"]; v != "" {
		where += " AND action = ?"
		args = append(args, v)
	}
	limit, offset := page(q)
	args = append(args, limit, offset)
	rows, err := s.q(ctx).Raw(`SELECT id::text, at, COALESCE(actor_id::text, ''), actor_email, action, entity, entity_id, summary,
		COALESCE(before::text, ''), COALESCE(after::text, ''), ip, user_agent, request_id, count(*) OVER ()
		FROM audit_log WHERE `+where+` ORDER BY at DESC LIMIT ? OFFSET ?`, args...).Rows()
	if err != nil {
		return nil, 0, translate(err)
	}
	defer rows.Close()
	out := []app.AuditEntry{}
	total := 0
	for rows.Next() {
		var e app.AuditEntry
		var before, after string
		if err := rows.Scan(&e.ID, &e.At, &e.ActorID, &e.ActorEmail, &e.Action, &e.Entity, &e.EntityID, &e.Summary,
			&before, &after, &e.IP, &e.UserAgent, &e.RequestID, &total); err != nil {
			return nil, 0, err
		}
		e.Before, e.After = rawJSON(before), rawJSON(after)
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func rawJSON(s string) any {
	if s == "" {
		return nil
	}
	return json.RawMessage(s)
}

func page(q app.ListQuery) (int, int) {
	limit := q.Limit
	if limit <= 0 || limit > 5000 {
		limit = 5000
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// searchClause builds an AND of ORs: every word must appear in at least one
// of the columns. Wildcards in the search text are escaped, so a search for
// "50%" means the text "50%", not "50 followed by anything".
func searchClause(cols []string, q string) (string, []any) {
	words := strings.Fields(q)
	if len(words) == 0 || len(cols) == 0 {
		return "true", nil
	}
	if len(words) > 8 {
		words = words[:8]
	}
	var parts []string
	var args []any
	esc := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	for _, w := range words {
		var ors []string
		for _, c := range cols {
			ors = append(ors, c+"::text ILIKE ? ESCAPE '\\'")
			args = append(args, "%"+esc.Replace(w)+"%")
		}
		parts = append(parts, "("+strings.Join(ors, " OR ")+")")
	}
	return strings.Join(parts, " AND "), args
}
