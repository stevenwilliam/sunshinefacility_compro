-- 0011 — the public contact email belongs on the bought domain:
-- info@sunshinefacility.com (Steven, 2026-11-05). Replaces 0010's
-- correction and the original typo, EN and ID independently; any other
-- value (a later CMS edit) is kept.
UPDATE parameters SET value_en = 'info@sunshinefacility.com', is_placeholder = false, updated_at = now()
 WHERE grp = 'contact' AND key = 'email' AND value_en IN ('info@sunshinefacilities.com', 'info@sunshinefacilites.com');
UPDATE parameters SET value_id = 'info@sunshinefacility.com', is_placeholder = false, updated_at = now()
 WHERE grp = 'contact' AND key = 'email' AND value_id IN ('info@sunshinefacilities.com', 'info@sunshinefacilites.com');
