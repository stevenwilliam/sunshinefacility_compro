-- 0006 — the office phone number is not shown on the public site. An empty
-- contact.phone hides it everywhere (footer, Home quote band, Contact page,
-- Organization JSON-LD); typing a number back in CMS → Parameters shows it again.
UPDATE parameters SET value_en = '', value_id = '', is_placeholder = false, updated_at = now()
 WHERE grp = 'contact' AND key = 'phone';
