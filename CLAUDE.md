# sfg — engineering & product DNA

This file is the contract for how this project is built. Read it first, every
session, before touching code or docs. Where it conflicts with a habit, this
file wins. Where it conflicts with `docs/02-business-rules.md` on *product*
logic, that document wins — this file governs *how* we build, not *what* the
product does.

It is generated from **`docs/99-steven-preference.md` §3–§9 — Steven's
portable engineering DNA, and a source of this file** — plus the brief in
`docs/design/handoff/CLAUDE_CODE_PROMPT.md` and the answered question batch
in `docs/question-batch-2026-10-22.md`. Those are the sources; this is the
local application of them. When they disagree, this file wins — it is the
newer, more specific decision — and it says below exactly where it departs.

---

## 1. What this is

**Owner:** Steven ("ven") · itdept.sfg@gmail.com
**Repository:** `git@github.com:stevenwilliam/sunshinefacility_compro.git`, branch `main`
**Started:** 2026-10-22

The public company-profile website for **Sunshine Facilities Group**
(www.sunshinefacility.com) — an Indonesian integrated facility-services
company, part of Lippo Indonesia — and the **admin CMS** that edits every
text and image on it. Bilingual **EN / ID**. Seven public pages (Home,
Services, Service Detail ×5 lines, Industries, About Us, Careers, Contact /
Request a Quotation), a three.js hero ball, two public forms (quotation, job
application) that are stored and emailed, and a read-only CMS inbox.

The design is approved and high fidelity: `docs/design/handoff/` (README,
`SFG Site (standalone).html`, `design/*.dc.html`). Recreate it; never ship
its runtime.

### Deviations from the standing preference, and why

| Deviation | Reason |
|---|---|
| **Public site is light-only** (99 §8 asks for light and dark). The CMS has both. | The approved design is one brand theme (Charcoal/Cotton). Decision Q19 / D19. |
| **Two languages, not three** (impeccable §8 says "all three languages"). | The site is EN + ID. The CMS is English-only, catalogue-backed (Q16). |
| **No `/id/` URL prefix** (99 §13 implies indexable language variants). | Language follows the design (per browser) plus `?lang=id`. Indonesian pages are not indexed separately until phase 2 (Q17). |
| **Client validation is not generated from an OpenAPI schema** (99 §7 "same rules, one source"). | The server sends its limits in the public bundle (`rules`) and its field definitions to the CMS (`/admin/schema`); both sides read those, not two hand copies. The OpenAPI document is written, not generated (D23). |
| **Repeated lists share one table** (`content_items` with an allow-listed `list_key`) instead of a table each (Q13 said "table per list"). | Eight lists have an identical shape; eight identical tables add nothing. Each list still has its own CMS screen and filter (D11). |
| **No money** in this product. | The integer-money rules of 99 §6 have nothing to apply to. |
| **Inbox is read-only** — no status, notes or assignee. | Decision Q8. Follow-up happens outside the CMS. |
| **Applications are deleted automatically** after `careers.retention_months` (99 §8 "nothing automated cancels a booking"). | Not a booking: personal data past its purpose (UU PDP). Decision Q9, audited per row. |

Everything else in `99-steven-preference.md` stands.

---

## 2. Architecture — non-negotiable

Hexagonal. Dependencies point **inward only**: `adapter → app → domain`;
`platform` is available to all. `domain` imports no framework, no driver, no
`net/http`, no SQL.

```
cmd/api/main.go           serve · migrate up|down|status · seed-assets · seed-dev · job <name>
internal/
  domain/                 pure rules, exhaustively unit-tested, no I/O
    access/  calendar/  certificate/  contrast/  filetype/  forms/  params/  placeholder/  resource/
  app/                    use-cases: auth, content (generic CMS), media, public bundle, submissions, jobs, audit
  adapter/
    http/                 gin router, middleware, handlers, SPA + SEO head injection
    postgres/             repositories, raw SQL with placeholders
    storage/              MinIO (private bucket)
    notify/               SMTP + log mailer
    media/                WebP re-encoding
  platform/               apierror config csvexport database id logging ratelimit sanitize security
db/migrations/            NNNN_name.up.sql + .down.sql, go:embed, checksummed
web/                      React 18 + Vite + TypeScript + Tailwind (public site + /admin), embedded in the binary
test/                     integration, authorisation and security suites (real PostgreSQL)
deploy/                   systemd units + timers, nginx vhost
```

**`internal/domain/resource` is the single source** for every CMS entity:
fields, types, limits, content-vs-not, who may write, whether rows can be
added. Validation, SQL identifiers, CMS forms and CSV columns all read it.

---

## 3. Stack

Go 1.26 (toolchain pinned to 1.26.10 in `go.mod`), gin, gorm + pgx (Raw/Exec
only), golang-jwt v5, google/uuid v7, minio-go v7, gen2brain/webp (pure Go),
x/image. PostgreSQL 18 native. MinIO and mailpit in Docker (satellites only).
Frontend: React **18** (not 19), Vite 8, TypeScript 5.9, Tailwind 3, three.js
**0.160.0**, Phosphor icons generated into `web/src/lib/phosphor.tsx`
(`make icons`). **Never a PWA.**

## 4. Database conventions

- Migrations are the source of truth; numbered, paired, forward-only in
  production, checksummed (an edited applied migration refuses to boot).
- **The database enforces the invariants**: expired certificates cannot be
  published (trigger), media in use cannot be deleted (FK RESTRICT), roles
  are a CHECK, audit is append-only (trigger), openings reference existing
  location/type parameters (composite FK).
- UUIDv7 keys. `timestamptz` in UTC; business dates in **Asia/Jakarta**
  explicitly (`domain/calendar`).
- Seed content is migration `0002`, **generated** from
  `scripts/seed/content.py` (`make seed-sql`), byte-deterministic. Later text
  additions are new migrations (`0003…`), never edits to `0002`.

## 5. Security posture

ASVS L2 / OWASP Top 10, mapped control-by-control with the test that proves
it in `docs/12-security.md`. Deny-by-default (`domain/access`); every admin
route declares its permission; argon2id; 15-minute access JWT in memory +
rotating hashed refresh cookie (SameSite=Strict, httpOnly) with reuse
detection; jti revocation on logout; generic login errors and lockout;
validation on both sides with server-side allow-lists; magic-byte uploads,
re-encoded; private bucket; CSP without `unsafe-inline`; pipe CSV with
formula guard; structured logs without PII; append-only audit.

## 6. Product & UI conventions (99 §8, applied)

- **Search on every list. Pipe-delimited CSV on every grid**, honouring the
  search and filters on screen. **Every disabled control says why**, visibly.
- **Placeholders stay visibly tagged** until the content is real; a content
  edit clears the flag (BR-2).
- Every configurable value is a **parameter** (CMS → Parameters), typed by
  `domain/params`; system groups are superadmin-only.
- WCAG AA measured (`make contrast`, design.md §3). Mobile-first from 360px.
- All public text comes from the catalogue (`site_text`) or content tables —
  no inline strings; a guard test fails on a key the catalogue lacks.

## 7. Working agreement

Steven's standing rules (99 §1–§2) apply: one batch of questions up front
(delivered one at a time — see memory), then build to the end; auto-commit
and push to `main` with conventional commits; docs updated in the same commit;
the truth about what was verified; anything needing an interactive terminal
goes in `docs/RUN-WHEN-BACK.md`; **every reply ends with the access table**
(URL, backend URL, admin user, admin password — dev credentials only).

## 8. Infrastructure (dev server `claudedev`)

| Thing | Where |
|---|---|
| Config | `/etc/sfg/sfg.env` (root:dev 0640) |
| Service | `sfg.service` → `127.0.0.1:8098`; timers `sfg-certificate-reminders`, `sfg-purge-applications` |
| Public | nginx `:8097` (allowlist: LAN 192.168.88.0/24, VMware host 172.16.0.0/24), ufw rules for both |
| Database | native PostgreSQL 18: `sfg`, `sfg_test` (role `sfg_app`) |
| Objects | Docker `sfg-minio` on 127.0.0.1:9012 (console 9013), bucket `sfg` |
| Mail | shared mailpit 127.0.0.1:1025 (UI :8025) |

`make help` lists every command. `make check` gates a commit.
