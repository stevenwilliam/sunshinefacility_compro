-- 0013 — the GA4 stream that reports for sunshinefacility.com is
-- G-YQ3537VDCW (Steven, 2026-11-05); 0009 used an earlier stream's ID.
-- Only replaces that earlier ID, so a value set in the CMS is kept.
UPDATE parameters SET value_en = 'G-YQ3537VDCW', value_id = 'G-YQ3537VDCW', is_placeholder = false, updated_at = now()
 WHERE grp = 'analytics' AND key = 'ga4_measurement_id' AND value_en = 'G-XNBZ88NTP9';
