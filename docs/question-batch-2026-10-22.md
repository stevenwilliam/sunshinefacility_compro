# Question batch — answered 2026-10-22

Asked one at a time; answers by Steven unless marked *decided by default*.
These become D1… in `00-README-and-decisions.md` when the doc set is written.

| # | Topic | Decision |
|---|---|---|
| 1 | Repo | `git@github.com:stevenwilliam/sunshinefacility_compro.git`, branch `main` (repo exists, empty; SSH auth OK as stevenwilliam) |
| 2 | Handoff | Copied verbatim to `docs/design/handoff/`; zip gitignored |
| 3 | Ports | nginx `:8097` → Go API `127.0.0.1:8098`; ufw for 192.168.88.0/24 + 172.16.0.1 in RUN-WHEN-BACK |
| 4 | Database | Native PostgreSQL 18: `sfg` + `sfg_test` |
| 5 | Storage | Own `sfg-minio` (127.0.0.1:9012/9013), private bucket; public images via Go `/media/<hash>` immutable; CVs via presigned URL |
| 6 | Email to | Quotations + applications → `itdept.sfg@gmail.com` (parameters); mailpit in dev |
| 6b | Auto-reply | None |
| 7 | WhatsApp alerts | No; email only, behind notify port |
| 8 | Inbox | **Read-only list** (view + search + CSV); no status workflow |
| 9 | CV retention | **Auto-delete** applications + CVs older than N months; N = parameter, default 12 *(N and scope decided by default: applications only, quotations kept)*; audited |
| 10 | Anti-spam | Rate limit per IP/email + honeypot + minimum fill time; no CAPTCHA |
| 11 | Roles | superadmin, editor, sales (quotations), hr (applications/CVs) |
| 12 | Login | argon2id; prod first-run setup, no default admin; dev-only seed `admin@sfg.local`; superadmin resets; no self-reset, no 2FA |
| 13 | Content model | Table per repeated list (EN/ID) + CMS screen each; "Site text" screen for all UI strings |
| 14 | Placeholders | Flag + visible public tag; **auto-clears when a content field is edited** (not on publish/order-only changes); CMS shows remaining count; testimonials seeded unpublished → home shows dashed placeholder cards |
| 15 | Certificates | DB + API refuse publishing expired; public auto-hides expired (Asia/Jakarta); **plus expiry reminder email** N days before (parameter, default 60 *decided by default*) |
| 16 | CMS language | English only (catalogue-backed) |
| 17 | SEO | SPA; Go injects per-route title/description/OG/canonical/JSON-LD; English URLs `/services/:line` (`?line=` 301); localStorage lang + `?lang=id`; no `/id/` prefix (phase 2) |
| 18 | Browser | Install Playwright + Chromium |
| 19 | Dark mode | Public site brand theme only (deviation logged); CMS light + dark |
| 20 | Map | **Google Maps keyless embed**; coordinates parameter, placeholder |
| 21 | Images | Magic-byte check, re-encode WebP 480/960/1600, keep original |
| 22 | Analytics | **GA4**; measurement ID parameter (empty = no script); consent mode default denied + bilingual banner; never on /admin |
| 23 | Production | Handbook + systemd/nginx files only; no deploy |
| 24 | Social links | **Placeholder links shown**, all 4 icons, parameters marked placeholder |
| 25 | Sender | `noreply@sunshinefacility.com` parameter (Steven); SPF/DKIM in RUN-WHEN-BACK |
| 26 | Nohemi | Web licence confirmed by Steven; self-host woff2 |

Preferences in force: `docs/99-steven-preference.md` (copied from /home/share 2026-10-22; adds English-identifiers rule and the end-of-reply access table).
