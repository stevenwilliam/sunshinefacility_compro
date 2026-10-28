-- 0007 — an opening may link to its LinkedIn job post (optional). The check
-- keeps the column to https LinkedIn links, so the public "View on LinkedIn"
-- link can never point elsewhere.
ALTER TABLE career_openings ADD COLUMN linkedin_url text NOT NULL DEFAULT ''
    CONSTRAINT career_openings_linkedin_url_check
    CHECK (linkedin_url = '' OR linkedin_url ~ '^https://([a-z]{2,3}\.)?linkedin\.com/[^[:space:]]*$');

INSERT INTO site_text (key, value_en, value_id) VALUES
 ('careers.linkedin', 'View on LinkedIn', 'Lihat di LinkedIn');
