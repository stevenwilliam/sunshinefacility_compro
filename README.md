# Sunshine Facilities Group — website and CMS

The bilingual (EN/ID) company website for **www.sunshinefacility.com** (temporarily served at **sunshinefacility.sunshinefood.co.id**) and
the admin CMS that edits every text and image on it.

- **Public site:** Home (with the interactive 3D jigsaw hero ball), Services,
  Service Detail for the five business lines, Industries, About Us, Careers,
  Contact / Request a Quotation. Sticky header, EN/ID switch, floating
  WhatsApp, mobile-first from 360px, WCAG AA.
- **Forms:** quotation requests and job applications (with CV) are stored,
  emailed to the team through a retrying outbox, and readable in the CMS.
- **CMS** (`/admin`): pages, site text, services and their lists, industries,
  clients, certificates (expired ones cannot be published), careers,
  testimonials, management, lists, stats, media library, parameters; inbox;
  users and roles; audit history; outgoing mail. Every list searches and
  exports pipe-delimited CSV.

## Read in this order

1. `CLAUDE.md` — how this project is built, and where it departs from the standing preferences.
2. `docs/00-README-and-decisions.md` — the decision log.
3. `docs/02-business-rules.md` — the normative rules (BR-x.y); code and tests cite them.
4. `docs/02a-general-flow.md` — the flows, drawn.
5. `docs/PROGRESS.md` — what is built and what was actually run.
6. `docs/RUN-WHEN-BACK.md` — what needs Steven, or a machine that does not exist yet.

## Run it (development server)

```bash
cd /home/dev/projects/sfg
make web build migrate-up seed-assets seed-dev   # first time
make deploy                                      # after a change
make check                                       # fmt, vet, tests, contrast, diagrams, typecheck, audit
make e2e                                         # Playwright against the running site
```

Site: `http://<dev-server>:8097/` · CMS: `http://<dev-server>:8097/admin`.
Full setup: `docs/11-local-dev-setup.md`. Production: `docs/13-production-deployment-handbook.md`.

## Layout

```
cmd/api/            the one binary: serve, migrate, seed-assets, seed-dev, job
internal/domain/    pure business rules (access, certificate, forms, params, resource, …)
internal/app/       use-cases
internal/adapter/   http, postgres, storage (MinIO), notify (SMTP), media (WebP)
internal/platform/  config, errors, CSV, database, ids, logging, rate limit, sanitize, security
db/migrations/      schema, seed content, later text additions
scripts/            seed generator, contrast checker, mermaid checker
web/                React 18 + Vite + Tailwind; src/pages (site), src/admin (CMS), src/hero (three.js)
test/               integration + security suites; web/e2e for Playwright
deploy/             systemd units and timers, nginx vhost
docs/               the numbered doc set; docs/design/handoff is the approved design, verbatim
```
