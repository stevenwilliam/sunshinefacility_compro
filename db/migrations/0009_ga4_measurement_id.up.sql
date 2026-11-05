-- 0009 — Google Analytics 4 switched on with the property's measurement ID
-- (D42). Only fills an empty value, so an ID typed in the CMS is kept.
-- The banner still asks each visitor first (BR-12.1).
UPDATE parameters SET value_en = 'G-XNBZ88NTP9', value_id = 'G-XNBZ88NTP9', is_placeholder = false, updated_at = now()
 WHERE grp = 'analytics' AND key = 'ga4_measurement_id' AND value_en = '';
