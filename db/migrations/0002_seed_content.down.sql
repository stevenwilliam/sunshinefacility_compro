-- Development only: production is forward-only (CLAUDE.md §6).
DELETE FROM management; DELETE FROM testimonials; DELETE FROM career_openings;
DELETE FROM stats; DELETE FROM content_items; DELETE FROM certificates;
DELETE FROM clients; DELETE FROM industry_lines; DELETE FROM industries;
DELETE FROM service_groups; DELETE FROM service_facts; DELETE FROM service_items;
DELETE FROM service_lines; DELETE FROM page_sections; DELETE FROM site_text;
DELETE FROM parameters; DELETE FROM media WHERE slug IS NOT NULL;
