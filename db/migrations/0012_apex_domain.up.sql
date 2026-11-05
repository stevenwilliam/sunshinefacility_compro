-- 0012 — the website is sunshinefacility.com, without www (D43). The footer
-- follows; only touches the row if nobody has edited it in the CMS since.
UPDATE site_text SET value_en = 'sunshinefacility.com', value_id = 'sunshinefacility.com'
 WHERE key = 'footer.domain' AND value_en = 'www.sunshinefacility.com' AND value_id = 'www.sunshinefacility.com';
