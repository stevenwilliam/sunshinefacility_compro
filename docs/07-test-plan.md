# 07 — Test plan

## Layers

| Layer | Where | Runs against | Command |
|---|---|---|---|
| Domain unit tests | `internal/domain/**_test.go` | nothing (pure) | `go test ./internal/...` |
| Platform + adapter unit tests | `internal/platform`, `internal/adapter/media` | nothing / in-memory | same |
| Integration + security | `test/*.go` | real PostgreSQL (`sfg_test`), real router, migrations and seed; fakes only for object store and SMTP | `make test` |
| End-to-end | `web/e2e/*.spec.ts` | the running dev site through nginx | `make e2e` |
| Visual comparison | `web/scripts/shots.mjs` | running site **and** the approved standalone design | `make shots` |
| Contrast | `scripts/contrast.py` | design.md | `make contrast` |
| Diagrams | `scripts/mermaid-check` | every mermaid block in docs/ | `make diagrams` |
| Static/security | `go vet`, staticcheck, gosec, govulncheck, `npm audit --omit=dev`, `tsc` | code | `make check` |

Each integration test starts from a **fresh schema** (drop, migrate, seed):
order dependence is a defect, so `make test-shuffle` must pass too.

## What the suites prove (BR → test)

| Rule | Test |
|---|---|
| BR-1.1 pairs agree | `TestBR1_1_PairsAgree` |
| BR-1.1 catalogue covers every key | `TestCatalogueCoversEveryKey` (oracle: the migrated DB) |
| BR-1.3 no leaks in bundle | `TestBR1_3_BundleLeaksNothingPrivate` |
| BR-2.2/2.3 placeholder flag | `TestBR2_*` (domain + `TestBR2_PlaceholderFlag`) |
| BR-2.4 testimonial placeholders | `TestBR2_4_TestimonialPlaceholdersInBundle` |
| BR-3.1–3.4 certificates | domain `TestBR3_*`; `TestBR3_2_ExpiredCertificateCannotBePublished` (API + DB), `TestBR3_3_…`, `TestBR3_4_ReminderOncePerExpiry` |
| BR-4.1–4.4 media | `TestBR4_1_Images`, `TestBR4_1_Uploads`, `TestBR4_2_Renditions`, `TestBR4_3_MediaInUseCannotBeDeleted` |
| BR-5 quotations | domain `TestBR5_1_*`; `TestBR5_FormAbuse`; `TestBR5_3_MailRetriedNotLost` |
| BR-6 applications | domain `TestBR6_1_*`; `TestBR6_1_CVChecks`; `TestBR6_3_RetentionPurge` |
| BR-7 inbox | `TestIDORInbox` |
| BR-8 accounts | `TestBR8_1_Matrix`, `TestBR8_1_AuthorisationMatrix`, `TestBR8_2_SetupOnce`, `TestBR8_3_LoginGenericAndLockout`, `TestBR8_4_LastSuperadmin`, `TestJWTTamperingAndLogout` |
| BR-9.2 CSV | `TestBR9_2_CSVExport`, csvexport unit tests |
| BR-10.1 audit | `TestBR10_1_AuditTrail` |
| BR-11.1 contrast | `TestBR11_DesignRatios`, `TestBR11_1_ContrastRefused` |
| SEO | `TestSEOHeadPerRoute`, `TestNotFoundAndRedirects` |
| Icons agree | `TestIconListsAgree` |
| Headers | `TestSecurityHeaders` |

## End-to-end scenarios

Pages render with one `h1` and no console errors; no horizontal scroll at
360px; language switch and `?lang=id`; certificate lightbox (Esc, Close);
placeholders tagged; contact form disabled-with-reason then sent; Home quick
form carries to Contact; careers filters, clear-disabled reason, apply
prefill, oversized CV refused; service detail switcher and Remote map; CMS
edit appears on site; expired certificate toggle disabled with reason; list
search and export-disabled reason; media in use.

## Manual / visual QA checklist

- Compare `make shots` output against the design at 1440 and 360.
- Hero ball under Xvfb (headless Chromium on this VM has no WebGL): idle, click → split, pick a panel, Esc, reassemble.
- Keyboard only: skip link, header, burger drawer (Esc), forms, lightbox, CMS drawer (Esc), confirm dialog.
- Reduced motion: ribbons still, ball does not spin.
- Screen reader spot-check of form errors (announced via `aria-live`/`role=alert`).
