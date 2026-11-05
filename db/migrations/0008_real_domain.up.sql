-- 0008 — www.sunshinefacility.com is bought (with .co.id and .id, which
-- redirect to it). The footer shows the real domain again, undoing 0004.
-- Only touches the row if nobody has edited it in the CMS since.
UPDATE site_text SET value_en = 'www.sunshinefacility.com', value_id = 'www.sunshinefacility.com'
 WHERE key = 'footer.domain' AND value_en = 'sunshinefacility.sunshinefood.co.id' AND value_id = 'sunshinefacility.sunshinefood.co.id';
