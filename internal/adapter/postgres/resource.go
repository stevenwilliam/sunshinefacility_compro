package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/resource"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
)

// selectExpr renders a field for reading. Ids and dates are cast to text so
// a row scans into plain Go values; a Refs field becomes a JSON array from
// its join table, in order.
func selectExpr(d resource.Def, f resource.Field) string {
	switch f.Type {
	case resource.Media, resource.Ref:
		return "t." + f.Key + "::text AS " + f.Key
	case resource.Date:
		return "to_char(t." + f.Key + ", 'YYYY-MM-DD') AS " + f.Key
	case resource.Refs:
		j := d.RefJoin[f.Key]
		return fmt.Sprintf("COALESCE((SELECT json_agg(j.%s::text ORDER BY j.sort) FROM %s j WHERE j.%s = t.%s)::text, '[]') AS %s",
			j.RefCol, j.Table, j.OwnerCol, d.IDColumn, f.Key)
	}
	return "t." + f.Key
}

func selectList(d resource.Def) string {
	cols := []string{"t." + d.IDColumn + "::text AS _id"}
	for _, f := range d.Fields {
		if f.Key == d.IDColumn {
			continue
		}
		cols = append(cols, selectExpr(d, f))
	}
	if d.IDColumn != "id" || d.Table == "site_text" {
		// site_text's key is both the id and a field.
		cols = append(cols, "t."+d.IDColumn)
	}
	cols = append(cols, "t.updated_at")
	if d.Name == "parameters" {
		cols = append(cols, "t.is_secret")
	}
	return strings.Join(cols, ", ")
}

func scanRows(d resource.Def, rows interface {
	Next() bool
	Scan(...any) error
	Columns() ([]string, error)
	Err() error
}, withTotal bool) ([]app.Row, int, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, 0, err
	}
	out := []app.Row{}
	total := 0
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, 0, err
		}
		r := app.Row{}
		for i, c := range cols {
			if c == "_total" {
				if n, ok := vals[i].(int64); ok {
					total = int(n)
				}
				continue
			}
			r[c] = normalise(vals[i])
		}
		for _, f := range d.Fields {
			if f.Type == resource.Refs {
				var ids []string
				_ = json.Unmarshal([]byte(fmt.Sprint(r[f.Key])), &ids)
				if ids == nil {
					ids = []string{}
				}
				r[f.Key] = ids
			}
		}
		r["id"] = r["_id"]
		delete(r, "_id")
		out = append(out, r)
	}
	if !withTotal {
		total = len(out)
	}
	return out, total, rows.Err()
}

func normalise(v any) any {
	switch x := v.(type) {
	case []byte:
		return string(x)
	case int32:
		return int64(x)
	}
	return v
}

func (s *Store) ListResource(ctx context.Context, d resource.Def, q app.ListQuery) ([]app.Row, int, error) {
	var search []string
	for _, f := range d.Fields {
		if f.Search {
			search = append(search, "t."+f.Key)
		}
	}
	where, args := searchClause(search, q.Q)
	for _, fl := range d.Filters {
		v, ok := q.Filters[fl.Key]
		if !ok || v == "" {
			continue
		}
		where += " AND t." + fl.Key + "::text = ?"
		args = append(args, v)
	}
	limit, offset := page(q)
	args = append(args, limit, offset)
	order := "t." + strings.ReplaceAll(d.OrderBy, ", ", ", t.")
	rows, err := s.q(ctx).Raw(`SELECT `+selectList(d)+`, count(*) OVER () AS _total FROM `+d.Table+` t WHERE `+where+
		` ORDER BY `+order+` LIMIT ? OFFSET ?`, args...).Rows()
	if err != nil {
		return nil, 0, translate(err)
	}
	defer rows.Close()
	return scanRows(d, rows, true)
}

func (s *Store) GetResource(ctx context.Context, d resource.Def, rowID string, forUpdate bool) (app.Row, error) {
	lock := ""
	if forUpdate {
		lock = " FOR UPDATE OF t"
	}
	cond := "t." + d.IDColumn + " = ?"
	if d.IDColumn == "id" {
		if _, err := id.Parse(rowID); err != nil {
			return nil, app.ErrNotFound
		}
		cond = "t.id = ?::uuid"
	}
	rows, err := s.q(ctx).Raw(`SELECT `+selectList(d)+` FROM `+d.Table+` t WHERE `+cond+lock, rowID).Rows()
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	out, _, err := scanRows(d, rows, false)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, app.ErrNotFound
	}
	return out[0], nil
}

// bind turns a validated value into a SQL placeholder expression and arg.
func bind(f resource.Field, v any) (string, any) {
	switch f.Type {
	case resource.Media, resource.Ref:
		if v == nil {
			return "NULL", nil
		}
		return "?::uuid", v
	case resource.Date:
		if v == nil {
			return "NULL", nil
		}
		return "?::date", v
	}
	return "?", v
}

func (s *Store) InsertResource(ctx context.Context, d resource.Def, vals map[string]any, actorID string) (string, error) {
	cols, exprs, args := []string{}, []string{}, []any{}
	newID := ""
	if d.IDColumn == "id" {
		newID = id.NewString()
		cols, exprs, args = append(cols, "id"), append(exprs, "?::uuid"), append(args, newID)
	}
	for _, f := range d.Fields {
		v, ok := vals[f.Key]
		if !ok || f.Type == resource.Refs {
			continue
		}
		e, a := bind(f, v)
		cols, exprs = append(cols, f.Key), append(exprs, e)
		if e != "NULL" {
			args = append(args, a)
		}
		if f.Key == d.IDColumn {
			newID = fmt.Sprint(v)
		}
	}
	cols, exprs = append(cols, "updated_by"), append(exprs, "?::uuid")
	args = append(args, nullStr(actorID))
	if err := s.q(ctx).Exec(`INSERT INTO `+d.Table+` (`+strings.Join(cols, ", ")+`) VALUES (`+strings.Join(exprs, ", ")+`)`, args...).Error; err != nil {
		return "", translate(err)
	}
	if err := s.setRefs(ctx, d, newID, vals); err != nil {
		return "", err
	}
	return newID, nil
}

func (s *Store) UpdateResource(ctx context.Context, d resource.Def, rowID string, vals map[string]any, actorID string) error {
	sets, args := []string{}, []any{}
	for _, f := range d.Fields {
		v, ok := vals[f.Key]
		if !ok || f.Type == resource.Refs {
			continue
		}
		e, a := bind(f, v)
		sets = append(sets, f.Key+" = "+e)
		if e != "NULL" {
			args = append(args, a)
		}
	}
	sets = append(sets, "updated_at = now()", "updated_by = ?::uuid")
	args = append(args, nullStr(actorID))
	cond := d.IDColumn + " = ?"
	if d.IDColumn == "id" {
		cond = "id = ?::uuid"
	}
	args = append(args, rowID)
	res := s.q(ctx).Exec(`UPDATE `+d.Table+` SET `+strings.Join(sets, ", ")+` WHERE `+cond, args...)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected != 1 {
		return app.ErrNotFound // an update that touched nothing is not a success
	}
	return s.setRefs(ctx, d, rowID, vals)
}

func (s *Store) setRefs(ctx context.Context, d resource.Def, ownerID string, vals map[string]any) error {
	for _, f := range d.Fields {
		if f.Type != resource.Refs {
			continue
		}
		ids, ok := vals[f.Key].([]string)
		if !ok {
			continue
		}
		j := d.RefJoin[f.Key]
		if err := s.q(ctx).Exec(`DELETE FROM `+j.Table+` WHERE `+j.OwnerCol+` = ?::uuid`, ownerID).Error; err != nil {
			return translate(err)
		}
		for i, ref := range ids {
			if err := s.q(ctx).Exec(`INSERT INTO `+j.Table+` (`+j.OwnerCol+`, `+j.RefCol+`, sort) VALUES (?::uuid, ?::uuid, ?)`, ownerID, ref, i).Error; err != nil {
				return translate(err)
			}
		}
	}
	return nil
}

func (s *Store) DeleteResource(ctx context.Context, d resource.Def, rowID string) error {
	cond := d.IDColumn + " = ?"
	if d.IDColumn == "id" {
		cond = "id = ?::uuid"
	}
	res := s.q(ctx).Exec(`DELETE FROM `+d.Table+` WHERE `+cond, rowID)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected != 1 {
		return app.ErrNotFound
	}
	return nil
}

func (s *Store) Options(ctx context.Context) (app.Options, error) {
	o := app.Options{Refs: map[string][]app.Option{}, Params: map[string][]app.Option{}}
	load := func(key, query string) error {
		rows, err := s.q(ctx).Raw(query).Rows()
		if err != nil {
			return translate(err)
		}
		defer rows.Close()
		list := []app.Option{}
		for rows.Next() {
			var v, l string
			if err := rows.Scan(&v, &l); err != nil {
				return err
			}
			list = append(list, app.Option{Value: v, Label: l})
		}
		o.Refs[key] = list
		return rows.Err()
	}
	if err := load("services", `SELECT id::text, title_en FROM service_lines ORDER BY sort, title_en`); err != nil {
		return o, err
	}
	if err := load("industries", `SELECT id::text, title_en FROM industries ORDER BY sort, title_en`); err != nil {
		return o, err
	}
	rows, err := s.q(ctx).Raw(`SELECT grp, key, value_en FROM parameters WHERE grp IN ('locations', 'job_types', 'building_types', 'brand_colours') ORDER BY grp, sort, key`).Rows()
	if err != nil {
		return o, translate(err)
	}
	defer rows.Close()
	for rows.Next() {
		var g, k, v string
		if err := rows.Scan(&g, &k, &v); err != nil {
			return o, err
		}
		o.Params[g] = append(o.Params[g], app.Option{Value: k, Label: v})
	}
	return o, rows.Err()
}

func (s *Store) AllParameters(ctx context.Context) ([]app.Parameter, error) {
	rows, err := s.q(ctx).Raw(`SELECT id::text, grp, key, value_en, value_id, description, is_secret, is_placeholder, sort, updated_at
		FROM parameters ORDER BY grp, sort, key`).Rows()
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	var out []app.Parameter
	for rows.Next() {
		var p app.Parameter
		if err := rows.Scan(&p.ID, &p.Group, &p.Key, &p.ValueEN, &p.ValueID, &p.Description, &p.IsSecret, &p.IsPlaceholder, &p.Sort, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// PlaceholderCounts — BR-2.1: the CMS shows how many placeholders remain.
func (s *Store) PlaceholderCounts(ctx context.Context) (map[string]int, error) {
	out := map[string]int{}
	for _, d := range resource.Defs {
		if !d.Has("is_placeholder") {
			continue
		}
		var n int
		if err := s.row(ctx, `SELECT count(*) FROM `+d.Table+` WHERE is_placeholder`).Scan(&n); err != nil {
			return nil, translate(err)
		}
		out[d.Name] = n
	}
	return out, nil
}
