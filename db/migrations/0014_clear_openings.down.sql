-- The deleted openings are not restored (the pre-update dump holds them).
DELETE FROM site_text WHERE key = 'careers.none';
