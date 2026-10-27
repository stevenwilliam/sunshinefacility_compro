# 00 — Index and decision log

## The doc set

| # | Document | What it answers |
|---|---|---|
| 00 | this file | index, decision log, open questions |
| 01 | `01-PRD.md` | what we are building and for whom |
| 02 | `02-business-rules.md` | **normative** rules, BR-x.y, cited by code and tests |
| 02a | `02a-general-flow.md` | the flows, drawn |
| 03 | `03-data-model.md` | schema, ERD, constraints |
| 04 | `04-api-specification.md` + `openapi.yaml` | the REST contract |
| 05 | `05-architecture-and-nfr.md` | architecture, performance, observability |
| 06 | `06-domain-operations.md` | content operations, jobs, mail, retention |
| 07 | `07-test-plan.md` | test strategy and what each suite proves |
| 08 | `08-roadmap.md` | phase 1 and what is deliberately later |
| 09 | `09-deployment.md` | how the dev server runs it; backups, rollback |
| 10 | `10-design-system.md` | tokens, type, components, accessibility |
| 11 | `11-local-dev-setup.md` | everyday commands |
| 12 | `12-security.md` | ASVS L2 / Top 10 map, abuse cases, proving tests |
| 13a | `13a-development-server-preparation.md` | dev server handbook |
| 13 | `13-production-deployment-handbook.md` | empty machine → running site |
| 14 | `user_guide/user-guide.md` + `.pdf` | **the** CMS user guide, illustrated, kept current by `make user-guide` (`14-user-guide.md` points there) |
| 15 | `15-admin-guide.md` | for whoever runs it |
| 16 | `16-uat-scenario-handbook.md` | scenarios to sign it off |
| — | `PROGRESS.md`, `RUN-WHEN-BACK.md` | what ran; what is waiting |
| — | `design/handoff/` | the approved design, verbatim |
| — | `question-batch-2026-10-22.md` | the answered question batch |

## Decision log

Dated, with the documents each touched. A superseded decision is marked,
never deleted.

| # | Date | Decision | By | Docs |
|---|---|---|---|---|
| D1 | 2026-10-22 | Repository `stevenwilliam/sunshinefacility_compro`, branch `main`; auto-commit and push | Steven (Q1) | CLAUDE, 11 |
| D2 | 2026-10-22 | The handoff is kept verbatim in `docs/design/handoff/`; the zip is ignored | Steven (Q2) | 00 |
| D3 | 2026-10-22 | nginx `:8097` → Go `127.0.0.1:8098`; ufw for 192.168.88.0/24 and 172.16.0.0/24 | Steven (Q3) | 09, 13a |
| D4 | 2026-10-22 | Native PostgreSQL 18: `sfg`, `sfg_test` | Steven (Q4) | 03, 11 |
| D5 | 2026-10-22 | Own `sfg-minio` (9012/9013), private bucket; public images through the Go server under content-hashed URLs, CVs by 5-minute presigned links | Steven (Q5) | 05, 12 |
| D6 | 2026-10-22 | Quotations and applications both emailed to `itdept.sfg@gmail.com` (parameters); no auto-reply | Steven (Q6) | 02 BR-5/6, 06 |
| D7 | 2026-10-22 | Email only; WhatsApp alerts deferred behind the notify port | Steven (Q7) | 08 |
| D8 | 2026-10-22 | Inbox is read-only (no status/notes) | Steven (Q8) | 02 BR-7 |
| D9 | 2026-10-22 | Applications and CVs auto-deleted after N months (parameter, default 12); quotations kept | Steven (Q9); N and scope by default | 02 BR-6.3, 06 |
| D10 | 2026-10-22 | Anti-spam: rate limit per IP and email, honeypot, minimum fill time; no CAPTCHA | Steven (Q10) | 02 BR-5.4, 12 |
| D11 | 2026-10-22 | Repeated lists of identical shape share `content_items` with an allow-listed `list_key`; each still has its own CMS screen. Departs from Q13's "table per list" for eight identical shapes. | Claude (flagged) | 03, CLAUDE |
| D12 | 2026-10-22 | One industry↔line mapping drives both the Industries chips and the Service Detail "industries we serve" chips. The design files disagreed (e.g. Digital listed Industrial on one page, not the other); the Industries page's mapping was taken. | Claude (flagged) | 02, 03 |
| D13 | 2026-10-22 | Four roles: superadmin, editor, sales, hr | Steven (Q11) | 02 BR-8 |
| D14 | 2026-10-22 | Password login (argon2id), first-run setup with a token, dev-only seeded admin, superadmin resets, no 2FA | Steven (Q12) | 02 BR-8, 12 |
| D15 | 2026-10-22 | Table per list + Site text screen for every interface string | Steven (Q13); see D11 | 03 |
| D16 | 2026-10-22 | Placeholder flag auto-clears on a **content** edit; publishing/order do not count; an explicit change wins | Steven (Q14) + Claude's narrowing | 02 BR-2 |
| D17 | 2026-10-22 | Expired certificates: DB + API refuse; public auto-hide; reminder email N days before (default 60) | Steven (Q15); N by default | 02 BR-3 |
| D18 | 2026-10-22 | CMS in English only | Steven (Q16) | 14 |
| D19 | 2026-10-22 | Public site brand theme only; CMS light + dark | Steven (Q19) | 10, CLAUDE |
| D20 | 2026-10-22 | SEO: server-injected head per route; English URLs `/services/:line`; `?line=` 301; `?lang=id` | Steven (Q17) | 04, 05 |
| D21 | 2026-10-22 | Google Maps keyless embed; coordinates are placeholder parameters | Steven (Q20) | 02, 12 |
| D22 | 2026-10-22 | Uploads re-encoded to WebP at 480/960/1600 (certificates also 2400; logos 240/480), never upscaled | Steven (Q21) | 02 BR-4 |
| D23 | 2026-10-22 | GA4 optional: measurement-ID parameter, consent banner (default denied), never on /tytb | Steven (Q22) | 02 BR-12, 12 |
| D24 | 2026-10-22 | Production: handbook only, nothing deployed | Steven (Q23) | 13 |
| D25 | 2026-10-22 | Footer social links shown with placeholder URLs, tagged | Steven (Q24) | 02 BR-2 |
| D26 | 2026-10-22 | Sender `noreply@sunshinefacility.com` (parameter) | Steven (Q25) | 06 |
| D27 | 2026-10-22 | Nohemi web licence confirmed; self-hosted woff2 | Steven (Q26) | 10 |
| D28 | 2026-10-22 | Quotation form on Home carries its fields to the Contact form instead of submitting a partial request | Claude (README §8 "links on to the full Contact form") | 02 BR-5.5 |
| D29 | 2026-10-22 | Home testimonials: while none is published, the unpublished **placeholder** quotes are shown as dashed cards; a real unpublished quote never leaves the server | Claude | 02 BR-2.4 |
| D30 | 2026-10-22 | Phosphor icons: a generated module with only the two weights used (−170 kB on first load) instead of the vendored icon font | Claude (performance) | 05, 10 |
| D31 | 2026-10-22 | `web/.npmrc` `legacy-peer-deps=true`: npm 10.8/10.9 crash resolving an optional peer; every needed peer is explicit | Claude | 11, RUN-WHEN-BACK |
| D32 | 2026-10-22 | Go toolchain pinned to 1.26.10 in `go.mod` (stdlib advisories); the server's shared Go is untouched | Claude | 12 |
| D33 | 2026-10-22 | Container content width is 1280px + padding (the design is content-box), base line-height `normal` — both measured against the approved design | Claude | 10 |
| D34 | 2026-10-22 | Certificate meta: "Valid until Mon YYYY" when dated, else "Issued by {issuer}" — on About this replaces the design's "Sertifikat Laik Higiene Sanitasi" line for SLHS, so both pages read the same field | Claude (flagged) | 10 |
| D35 | 2026-10-22 | Contact-detail placeholders are excluded from the JSON-LD, so a fake phone number never reaches search engines | Claude | 05, 12 |
| D36 | 2026-10-22 | Temporary domain `sunshinefacility.sunshinefood.co.id` until www.sunshinefacility.com is bought; host comes only from `SFG_PUBLIC_BASE_URL` (footer text: site text `footer.domain`, migration 0004) | Steven | 09, 13 |
| D37 | 2026-10-22 | Production deployed to 119.8.189.246 (supersedes D24): shared multi-project host, Cloudflare in front (ufw lets only Cloudflare reach 80/443), binary built on the dev server and copied over, mail via Office 365 `noreply@sunshinefood.co.id` (AUTH LOGIN support added to the mailer) | Steven | 13 |
| D38 | 2026-10-22 | CMS moved from `/admin` to **`/tytb`** (`AdminPath` in `internal/adapter/http/spa.go`, `ADMIN` in `web/src/admin/paths.ts`); `/admin` is a plain 404, and robots.txt names neither path (CMS pages carry `noindex`). The API stays at `/api/v1/admin` | Steven | 02a, 14 |
| D39 | 2026-10-22 | CMS user guide lives in `docs/user_guide/` as Markdown + PDF with real screenshots; regenerated by `make user-guide` and enforced by `user-guide-check` in `make check` whenever the CMS changes | Steven | 14, CLAUDE §7 |
| D40 | 2026-10-27 | Copywriting review (`SFG_Website_Copywriting_Review_1.xlsx`, tab "Yang ini") shipped as migration **0005**, each UPDATE guarded by the text it replaces so CMS edits survive. Yellow rows had been done in the CMS (EN only); 0005 adds their Indonesian and the two that were missed (buttons → "Discuss Your Requirements", contact intro). Remote keeps its line and chip label "Remote" but is titled "Remote-Managed, On-Site Delivered"; mission = the review's recommended set; industry descriptions rewritten to its rule. New contact-form fields (row 32) deferred | Steven | 02 |

## Open questions

None blocking. Items waiting on Steven are in `RUN-WHEN-BACK.md` (real
contact details, social URLs, photos, testimonials, management, vision and
mission, SMTP relay and DNS, production server).
