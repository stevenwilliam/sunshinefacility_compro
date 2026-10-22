# 08 — Roadmap

## Phase 1 — this build

Public site (7 pages, EN/ID, hero ball, ribbons), forms with storage and
mail, full CMS with media, parameters, inbox, users, audit; seed of all
design content; tests; dev deployment on `claudedev`; production handbook.

## Before launch (content and infrastructure — Steven's list)

Real address, phone, WhatsApp, email, map pin; social links; testimonials;
management names and photos; real openings; confirmed vision and mission;
higher-resolution photos (field app, mining camp); a native Indonesian
proofread; SMTP relay + SPF/DKIM for `noreply@sunshinefacility.com`;
production server, DNS and TLS. All in `RUN-WHEN-BACK.md`.

## Phase 2 candidates (deliberately later)

| Item | Why later |
|---|---|
| `/id/` URLs with `hreflang` | Indonesian pages indexed separately (D20) |
| Prerendering per page | richer previews per service if needed |
| WhatsApp alerts via WAHA | the notify port is ready (D7) |
| Inbox status / notes / assignee | Steven chose read-only (D8) |
| 2FA, self-service password reset | D14 |
| CMS in Bahasa Indonesia | D18 |
| `/metrics` (Prometheus) | no monitoring stack yet |
| Tailwind 4 | clears dev-only build-tool advisories (12-security) |
| Flutter app | not in scope; the API is versioned for it |
