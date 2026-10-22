# Progress

✅ done and verified by running it · 🟡 built, partly verified · ⬜ not started.
A ✅ is earned by running the gate, on the date shown.

## Gates run on 2026-10-22

| Gate | Command | Result |
|---|---|---|
| Format | `make fmt` | gofmt clean |
| Vet | `go vet ./...` | clean |
| Unit + integration + security | `make test` (fresh `sfg_test` schema per test) | **all pass** — 12 packages + `test/` (28 integration/security tests) |
| Order independence | `go test ./test/ -shuffle=on` | pass |
| End-to-end | `make e2e` (Playwright, Chromium, against the deployed site on :8097) | **14 / 14 pass** |
| Contrast | `make contrast` | 47 pairings, all pass, all recorded in design.md |
| Diagrams | `make diagrams` | 8 mermaid diagrams across 21 documents parse (the check was seen failing on 3 broken diagrams first) |
| TypeScript | `tsc -b --noEmit` | clean |
| Runtime npm audit | `npm audit --omit=dev` | 0 vulnerabilities |
| govulncheck | `govulncheck ./...` | 0 reachable (toolchain 1.26.10) |
| staticcheck / gosec | | clean (justified exceptions annotated) |
| Visual | `make shots` — 16 pages × site/design at 1440 and 360 | page heights within a few px of the approved design except Careers (+135 px: the prototype omits its own "reasons" block); spot-checked by eye |
| Hero ball | headed Chromium under Xvfb + software GL | idle ball, split, focus panel, labels, chips, Esc — screenshots match the design's states |
| Live forms | quotation + application against the running service | stored; both mails delivered to mailpit from `noreply@sunshinefacility.com` |

## Modules

| Area | State | Verified by |
|---|---|---|
| Schema, constraints, triggers | ✅ | SQL refusal tests (expired publish, audit delete), integration suite |
| Seed: all design copy EN/ID, 42 media, parameters | ✅ | migrations applied on dev + test; bundle inspected |
| Seed assets into MinIO (WebP renditions) | ✅ | `sfg seed-assets` — 42 files processed |
| Public site: 7 pages, EN/ID, header, footer, WhatsApp, ribbons | ✅ | e2e, screenshots |
| Hero ball | ✅ visually; 🟡 reduced-motion path not exercised in a browser | Xvfb run |
| Forms: validation both sides, storage, outbox, anti-spam | ✅ | integration + e2e + live mailpit |
| SEO head, robots, sitemap, JSON-LD, redirects, 404 | ✅ | `TestSEOHeadPerRoute`, `TestNotFoundAndRedirects`, curl |
| CMS: all resources, media, parameters, inbox, users, audit, outbox, account | ✅ for the paths in e2e + integration; 🟡 every screen clicked by hand only for certificates, services, media, inbox | e2e (4 CMS specs), integration |
| Roles and permissions | ✅ | authorisation matrix test (10 routes × 5 callers) |
| Jobs: mail worker, certificate reminders, retention purge | ✅ | integration tests; timers installed and enabled |
| Break-glass `sfg reset-password` | ✅ | `TestBreakGlassReset` |
| CMS dark theme | 🟡 | tokens written and measured; not screenshotted in dark mode |
| GA4 + consent | 🟡 | code path only — no measurement ID set, so the banner never shows (by design) |
| Dev deployment (systemd, timers, nginx :8097, ufw) | ✅ on the server; 🟡 **not verified from another machine** | `curl` on the box only (see RUN-WHEN-BACK) |
| Production (119.8.189.246, behind Cloudflare) | ✅ deployed 2026-10-22 (D37) | from outside: HTTPS 200 on `/`, `/admin`, `/robots.txt`; HSTS + CSP present; CMS login sets a `Secure; HttpOnly; SameSite=Strict` cookie; a test quotation stored and mailed through Office 365 (sent, no error); a backup dump restored into a scratch DB |
| Docs set 00–16, OpenAPI, CLAUDE, README | ✅ written | diagrams parse; OpenAPI parses (26 paths) |

## Defects found by the suites and fixed

Lightbox covered by the next section (stacking context → portal);
placeholder tag overflowing at 360px; delete of a parameter in use returned
500 (`restrict_violation` 23001 unmapped); mail retry compared against the
database clock; CMS inputs out-ranking utilities (specificity); content width
1184 vs 1280 and line-height 1.5 vs normal against the design; gorm
expanding `[]string` into a row constructor; three diagrams that did not parse.
