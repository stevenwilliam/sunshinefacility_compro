-- 0014 — remove the job vacancies (owner's request, 2026-11-06): every posted
-- opening and every seeded sample. An unpublished, non-sample opening (a real
-- draft written in the CMS) is kept. Applications keep their position text;
-- their opening_id becomes NULL (FK ON DELETE SET NULL).
-- Also the text the Careers page shows when no position is posted.
DELETE FROM career_openings WHERE published OR is_placeholder;

INSERT INTO site_text (key, value_en, value_id) VALUES
 ('careers.none', 'There are no open positions right now. You can still send us an open application below.',
                  'Saat ini belum ada posisi yang dibuka. Anda tetap dapat mengirim lamaran terbuka di bawah.');
