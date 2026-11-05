-- 0010 — the public contact email was typed as info@sunshinefacilites.com
-- (missing an "i") in the CMS on production. Only that exact value changes.
UPDATE parameters SET value_en = 'info@sunshinefacilities.com', value_id = 'info@sunshinefacilities.com', is_placeholder = false, updated_at = now()
 WHERE grp = 'contact' AND key = 'email' AND value_en = 'info@sunshinefacilites.com';
UPDATE parameters SET value_id = 'info@sunshinefacilities.com', updated_at = now()
 WHERE grp = 'contact' AND key = 'email' AND value_id = 'info@sunshinefacilites.com';
