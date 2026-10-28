package postgres

import (
	"context"
	"encoding/json"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
)

// PublicSnapshot reads every published row the site needs, in one
// read-only transaction so the bundle is a consistent picture.
func (s *Store) PublicSnapshot(ctx context.Context) (*app.Snapshot, error) {
	snap := &app.Snapshot{Media: map[string]app.MediaRecord{}}
	err := s.InTx(ctx, func(st app.Store) error {
		tx := st.(*Store)
		db := tx.q(ctx)
		if err := db.Exec(`SET TRANSACTION READ ONLY`).Error; err != nil {
			return err
		}
		each := func(query string, fn func(scan func(...any) error) error) error {
			rows, err := db.Raw(query).Rows()
			if err != nil {
				return translate(err)
			}
			defer rows.Close()
			for rows.Next() {
				if err := fn(rows.Scan); err != nil {
					return err
				}
			}
			return rows.Err()
		}

		if err := each(`SELECT key, value_en, value_id, is_placeholder FROM site_text`, func(sc func(...any) error) error {
			var r app.TextRow
			if err := sc(&r.Key, &r.EN, &r.ID, &r.Placeholder); err != nil {
				return err
			}
			snap.Text = append(snap.Text, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT page, section, title_en, title_id, body_en, body_id, COALESCE(image_id::text, ''), published, is_placeholder
			FROM page_sections ORDER BY page, sort`, func(sc func(...any) error) error {
			var r app.SectionRow
			if err := sc(&r.Page, &r.Section, &r.TitleEN, &r.TitleID, &r.BodyEN, &r.BodyID, &r.ImageID, &r.Published, &r.Placeholder); err != nil {
				return err
			}
			snap.Sections = append(snap.Sections, r)
			return nil
		}); err != nil {
			return err
		}
		ps, err := tx.AllParameters(ctx)
		if err != nil {
			return err
		}
		snap.Params = ps
		if err := each(`SELECT `+mediaCols+` FROM media m WHERE m.hash IS NOT NULL`, func(sc func(...any) error) error {
			r, err := scanMedia(scanner(sc), false)
			if err != nil {
				return err
			}
			snap.Media[r.ID] = *r
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT id::text, key, sort, title_en, title_id, short_en, short_id, label_en, label_id, summary_en, summary_id,
			description_en, description_id, lead_en, lead_id, color_base, color_ink, color_shade, color_on_ink, color_label,
			COALESCE(mark_id::text, ''), COALESCE(photo_id::text, ''), COALESCE(map_image_id::text, ''), show_map, show_certs
			FROM service_lines WHERE published ORDER BY sort, title_en`, func(sc func(...any) error) error {
			var r app.LineRow
			if err := sc(&r.ID, &r.Key, &r.Sort, &r.TitleEN, &r.TitleID, &r.ShortEN, &r.ShortID, &r.LabelEN, &r.LabelID,
				&r.SummaryEN, &r.SummaryID, &r.DescriptionEN, &r.DescriptionID, &r.LeadEN, &r.LeadID,
				&r.Base, &r.Ink, &r.Shade, &r.OnInk, &r.Label, &r.MarkID, &r.PhotoID, &r.MapImageID, &r.ShowMap, &r.ShowCerts); err != nil {
				return err
			}
			snap.Lines = append(snap.Lines, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT line_id::text, kind, text_en, text_id FROM service_items WHERE published ORDER BY line_id, kind, sort`, func(sc func(...any) error) error {
			var r app.ItemRow
			if err := sc(&r.LineID, &r.Kind, &r.EN, &r.ID); err != nil {
				return err
			}
			snap.Items = append(snap.Items, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT line_id::text, value_en, value_id, label_en, label_id FROM service_facts WHERE published ORDER BY line_id, sort`, func(sc func(...any) error) error {
			var r app.FactRow
			if err := sc(&r.LineID, &r.ValueEN, &r.ValueID, &r.LabelEN, &r.LabelID); err != nil {
				return err
			}
			snap.Facts = append(snap.Facts, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT line_id::text, title_en, title_id, body_en, body_id, items_en, items_id FROM service_groups WHERE published ORDER BY line_id, sort`, func(sc func(...any) error) error {
			var r app.GroupRow
			if err := sc(&r.LineID, &r.TitleEN, &r.TitleID, &r.BodyEN, &r.BodyID, &r.ItemsEN, &r.ItemsID); err != nil {
				return err
			}
			snap.Groups = append(snap.Groups, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT i.id::text, i.key, i.icon, i.title_en, i.title_id, i.body_en, i.body_id,
			COALESCE((SELECT json_agg(l.line_id::text ORDER BY l.sort) FROM industry_lines l JOIN service_lines sl ON sl.id = l.line_id AND sl.published WHERE l.industry_id = i.id)::text, '[]')
			FROM industries i WHERE i.published ORDER BY i.sort, i.title_en`, func(sc func(...any) error) error {
			var r app.IndustryRow
			var lines string
			if err := sc(&r.ID, &r.Key, &r.Icon, &r.TitleEN, &r.TitleID, &r.BodyEN, &r.BodyID, &lines); err != nil {
				return err
			}
			if err := json.Unmarshal([]byte(lines), &r.LineIDs); err != nil {
				return err
			}
			snap.Industries = append(snap.Industries, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT name, COALESCE(logo_id::text, ''), COALESCE(industry_id::text, '') FROM clients WHERE published ORDER BY sort, name`, func(sc func(...any) error) error {
			var r app.ClientRow
			if err := sc(&r.Name, &r.LogoID, &r.IndustryID); err != nil {
				return err
			}
			snap.Clients = append(snap.Clients, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT id::text, name, issuer, COALESCE(to_char(valid_until, 'YYYY-MM-DD'), ''), COALESCE(image_id::text, ''), description_en, description_id, published
			FROM certificates WHERE published ORDER BY sort, name`, func(sc func(...any) error) error {
			var r app.CertRow
			if err := sc(&r.ID, &r.Name, &r.Issuer, &r.ValidUntil, &r.ImageID, &r.DescriptionEN, &r.DescriptionID, &r.Published); err != nil {
				return err
			}
			snap.Certificates = append(snap.Certificates, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT list_key, icon, title_en, title_id, body_en, body_id, meta_en, meta_id, color, is_placeholder
			FROM content_items WHERE published ORDER BY list_key, sort`, func(sc func(...any) error) error {
			var r app.ListRow
			if err := sc(&r.ListKey, &r.Icon, &r.TitleEN, &r.TitleID, &r.BodyEN, &r.BodyID, &r.MetaEN, &r.MetaID, &r.Color, &r.Placeholder); err != nil {
				return err
			}
			snap.Lists = append(snap.Lists, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT scope, value_en, value_id, label_en, label_id FROM stats WHERE published ORDER BY scope, sort`, func(sc func(...any) error) error {
			var r app.StatRow
			if err := sc(&r.Scope, &r.ValueEN, &r.ValueID, &r.LabelEN, &r.LabelID); err != nil {
				return err
			}
			snap.Stats = append(snap.Stats, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT id::text, title_en, title_id, line_id::text, location_key, type_key, COALESCE(to_char(closing_date, 'YYYY-MM-DD'), ''),
			summary_en, summary_id, requirements_en, requirements_id, linkedin_url, is_placeholder
			FROM career_openings WHERE published ORDER BY sort, title_en`, func(sc func(...any) error) error {
			var r app.OpeningRow
			if err := sc(&r.ID, &r.TitleEN, &r.TitleID, &r.LineID, &r.LocationKey, &r.TypeKey, &r.ClosingDate,
				&r.SummaryEN, &r.SummaryID, &r.RequirementsEN, &r.RequirementsID, &r.LinkedInURL, &r.Placeholder); err != nil {
				return err
			}
			snap.Openings = append(snap.Openings, r)
			return nil
		}); err != nil {
			return err
		}
		if err := each(`SELECT quote_en, quote_id, name, role, company, published, is_placeholder FROM testimonials
			WHERE published OR is_placeholder ORDER BY sort, name`, func(sc func(...any) error) error {
			var r app.TestimonialRow
			if err := sc(&r.QuoteEN, &r.QuoteID, &r.Name, &r.Role, &r.Company, &r.Published, &r.Placeholder); err != nil {
				return err
			}
			snap.Testimonials = append(snap.Testimonials, r)
			return nil
		}); err != nil {
			return err
		}
		return each(`SELECT name, role_en, role_id, COALESCE(photo_id::text, ''), is_placeholder FROM management WHERE published ORDER BY sort, name`, func(sc func(...any) error) error {
			var r app.PersonRow
			if err := sc(&r.Name, &r.RoleEN, &r.RoleID, &r.PhotoID, &r.Placeholder); err != nil {
				return err
			}
			snap.Management = append(snap.Management, r)
			return nil
		})
	})
	return snap, err
}

type scanner func(...any) error

func (f scanner) Scan(dest ...any) error { return f(dest...) }
