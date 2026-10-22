-- 0004 — footer shows the temporary domain while www.sunshinefacility.com is
-- being purchased. Only touches the row if nobody has edited it in the CMS.
UPDATE site_text SET value_en = 'sunshinefacility.sunshinefood.co.id', value_id = 'sunshinefacility.sunshinefood.co.id'
 WHERE key = 'footer.domain' AND value_en = 'www.sunshinefacility.com' AND value_id = 'www.sunshinefacility.com';
