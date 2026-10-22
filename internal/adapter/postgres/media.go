package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
)

// usageExpr counts every reference to media m — the same columns whose
// foreign keys refuse a delete (BR-4.3), plus the link-preview parameter,
// which points at a slug rather than an id.
const usageExpr = `(
	(SELECT count(*) FROM page_sections   x WHERE x.image_id = m.id) +
	(SELECT count(*) FROM service_lines   x WHERE m.id IN (x.mark_id, x.photo_id, x.map_image_id)) +
	(SELECT count(*) FROM clients         x WHERE x.logo_id = m.id) +
	(SELECT count(*) FROM certificates    x WHERE x.image_id = m.id) +
	(SELECT count(*) FROM management      x WHERE x.photo_id = m.id) +
	(SELECT count(*) FROM parameters      x WHERE x.grp = 'seo' AND x.key = 'og_image' AND x.value_en = m.slug)
)`

const mediaCols = `m.id::text, COALESCE(m.slug, ''), m.kind, m.name, m.alt_en, m.alt_id, COALESCE(m.hash, ''), m.mime,
	m.width, m.height, m.bytes, m.object_key, m.variants::text, m.created_at`

func scanMedia(sc interface{ Scan(...any) error }, withUsage bool) (*app.MediaRecord, error) {
	var r app.MediaRecord
	var variants string
	dest := []any{&r.ID, &r.Slug, &r.Kind, &r.Name, &r.AltEN, &r.AltID, &r.Hash, &r.Mime, &r.Width, &r.Height, &r.Bytes, &r.ObjectKey, &variants, &r.CreatedAt}
	if withUsage {
		dest = append(dest, &r.Usage)
	}
	if err := sc.Scan(dest...); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(variants), &r.Variants); err != nil {
		return nil, err
	}
	if r.Variants == nil {
		r.Variants = []app.MediaVariant{}
	}
	return &r, nil
}

func (s *Store) InsertMedia(ctx context.Context, m app.MediaRecord) error {
	v, _ := json.Marshal(m.Variants)
	return translate(s.q(ctx).Exec(`INSERT INTO media (id, slug, kind, name, alt_en, alt_id, hash, mime, width, height, bytes, object_key, variants)
		VALUES (?::uuid, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb)`,
		m.ID, nullStr(m.Slug), m.Kind, m.Name, m.AltEN, m.AltID, nullStr(m.Hash), m.Mime, m.Width, m.Height, m.Bytes, m.ObjectKey, string(v)).Error)
}

func (s *Store) UpdateMediaFile(ctx context.Context, m app.MediaRecord) error {
	v, _ := json.Marshal(m.Variants)
	res := s.q(ctx).Exec(`UPDATE media SET hash = ?, mime = ?, width = ?, height = ?, bytes = ?, object_key = ?, variants = ?::jsonb, updated_at = now()
		WHERE id = ?::uuid`, nullStr(m.Hash), m.Mime, m.Width, m.Height, m.Bytes, m.ObjectKey, string(v), m.ID)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected != 1 {
		return app.ErrNotFound
	}
	return nil
}

func (s *Store) UpdateMediaMeta(ctx context.Context, mid, name, altEN, altID string) error {
	return translate(s.q(ctx).Exec(`UPDATE media SET name = ?, alt_en = ?, alt_id = ?, updated_at = now() WHERE id = ?::uuid`, name, altEN, altID, mid).Error)
}

func (s *Store) GetMedia(ctx context.Context, mid string) (*app.MediaRecord, error) {
	if _, err := id.Parse(mid); err != nil {
		return nil, app.ErrNotFound
	}
	r, err := scanMedia(s.row(ctx, `SELECT `+mediaCols+`, `+usageExpr+` FROM media m WHERE m.id = ?::uuid`, mid), true)
	return r, translate(err)
}

func (s *Store) MediaBySlug(ctx context.Context, slug string) (*app.MediaRecord, error) {
	r, err := scanMedia(s.row(ctx, `SELECT `+mediaCols+`, `+usageExpr+` FROM media m WHERE m.slug = ?`, slug), true)
	return r, translate(err)
}

func (s *Store) ListMedia(ctx context.Context, q, kind string) ([]app.MediaRecord, error) {
	where, args := searchClause([]string{"m.name", "m.kind", "m.slug", "m.alt_en", "m.alt_id"}, q)
	if kind != "" {
		where += " AND m.kind = ?"
		args = append(args, kind)
	}
	rows, err := s.q(ctx).Raw(`SELECT `+mediaCols+`, `+usageExpr+` FROM media m WHERE `+where+` ORDER BY m.kind, m.name`, args...).Rows()
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	out := []app.MediaRecord{}
	for rows.Next() {
		r, err := scanMedia(rows, true)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, rows.Err()
}

func (s *Store) MediaUsage(ctx context.Context, mid string) (int, error) {
	var n int
	err := s.row(ctx, `SELECT `+usageExpr+` FROM media m WHERE m.id = ?::uuid`, mid).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, app.ErrNotFound
	}
	return n, translate(err)
}

func (s *Store) DeleteMedia(ctx context.Context, mid string) error {
	res := s.q(ctx).Exec(`DELETE FROM media WHERE id = ?::uuid`, mid)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected != 1 {
		return app.ErrNotFound
	}
	return nil
}
