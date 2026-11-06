-- 0015 — the sample work-order card (WO-2610-0412) beside "Why Sunshine" on
-- Home is removed (owner's request, 2026-11-06). Its text and its two photo
-- sections go too, so the CMS lists nothing that no longer shows. The photos
-- stay in the media library.
DELETE FROM site_text WHERE key LIKE 'home.wo.%';
DELETE FROM page_sections WHERE page = 'home' AND section IN ('wo_before', 'wo_after');
