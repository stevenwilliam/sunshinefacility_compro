-- Development only: production is forward-only (CLAUDE.md §6).
DROP TABLE IF EXISTS notification_outbox, career_applications, quotation_requests,
    stats, content_items, management, testimonials, career_openings,
    certificate_reminders, certificates, clients, industry_lines, industries,
    service_groups, service_facts, service_items, service_lines,
    page_sections, site_text, parameters, media,
    audit_log, revoked_jti, refresh_tokens, users CASCADE;
DROP FUNCTION IF EXISTS certificates_refuse_expired_publish();
DROP FUNCTION IF EXISTS audit_log_append_only();
