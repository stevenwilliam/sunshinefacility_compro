# 02 — Business rules (normative)

Where this document and any other disagree on product logic, **this one is
right** and the other is stale. Code comments and test names cite these ids.
"Today" always means the date in **Asia/Jakarta**.

## BR-1 Content and language

- **BR-1.1** Every public string exists in English and Bahasa Indonesia. In the CMS a bilingual pair is required together or optional together (`resource_test: TestBR1_1_PairsAgree`). Interface strings live in `site_text`; every key the site uses must exist there (`TestCatalogueCoversEveryKey`).
- **BR-1.2** The visitor's language is remembered per browser (`localStorage['sfg-lang']`) and applies site-wide without a reload; `<html lang>` follows it. `?lang=id|en` in a URL sets and remembers it, so a shared link opens in its language.
- **BR-1.3** Only published rows reach the public site. The public bundle never carries recipients, thresholds, system parameters or anything secret (`TestBR1_3_BundleLeaksNothingPrivate`).

## BR-2 Placeholders

- **BR-2.1** A row or parameter flagged `is_placeholder` is shown on the public site with a visible **Placeholder** tag (dashed). Seeded placeholders: contact address, phone, WhatsApp, email and map pin; social links; testimonials; management; career openings; vision and mission. The CMS shows how many remain.
- **BR-2.2** The flag clears automatically when a **content** field of that row changes. Publishing, ordering and other non-content fields do not clear it.
- **BR-2.3** An editor may set or clear the flag explicitly; an explicit change in the same save wins over BR-2.2.
- **BR-2.4** Testimonials are seeded unpublished. While no testimonial is published, the Home page shows the unpublished **placeholder** testimonials as dashed placeholder cards. An unpublished testimonial that is no longer a placeholder is never sent to the site.

## BR-3 Certificates

- **BR-3.1** A certificate whose `valid_until` is before today is **Expired**. It is valid on its last day. No date = no expiry.
- **BR-3.2** An expired certificate cannot be published — refused by the API (with the date) and by a database trigger, so a direct SQL write is refused too. In the CMS the Published toggle is disabled with the reason.
- **BR-3.3** A published certificate that expires disappears from the public site on its own from the next day, with no write.
- **BR-3.4** A reminder email goes to `notify.certificate_to` when a published certificate is within `certificates.reminder_days` (default 60) of expiry — once per certificate per expiry date.
- **BR-3.5** The halal certificate (expired 25 Aug 2025) is seeded unpublished.

## BR-4 Media

- **BR-4.1** Uploads are identified by their bytes: PNG, JPEG or WebP; at most `media.max_upload_mb` (15) and `media.max_dimension_px` (8000) on either side, checked before decoding.
- **BR-4.2** Every upload is re-encoded to WebP renditions (photos 480/960/1600, certificates also 2400, logos 240/480), never upscaled; the original is kept privately. Public URLs carry the content hash and are cached as immutable.
- **BR-4.3** An image in use cannot be deleted; the refusal says how many items use it (FK RESTRICT in the database as well).
- **BR-4.4** SVG is never uploadable. The only SVGs are the five brand marks, which arrive with the seed and are served with a sandboxing CSP.

## BR-5 Quotation requests

- **BR-5.1** Required: full name, company, a valid email, phone (at least six digits; digits, spaces, `+ - ( ) .`), building type (one of the building-type parameters), location, at least one service (published line keys). Optional: area (digits only, 1–100,000,000 m²), message (≤ 2000). Limits are sent to the browser in the bundle so both sides use the same numbers.
- **BR-5.2** `?service=<line>` preselects that service. Submit stays disabled, with "To send, add: …", until the required fields are complete.
- **BR-5.3** A request is stored first, and its notification is written to the outbox in the same transaction; a worker sends it (retries 1 min, 5 min, 30 min, 2 h, 6 h, then daily; up to 10 attempts). A relay outage never loses a request.
- **BR-5.4** Anti-spam: a filled honeypot is answered as a success and stored nowhere; a form sent faster than `forms.min_fill_seconds` (3) is refused with a visible message; more than `forms.per_ip_per_hour` (10) per IP or `forms.per_email_per_hour` (5) per email in an hour is refused (429).
- **BR-5.5** The short form on Home does not submit; it carries name, company, building type and services to the Contact form.

## BR-6 Job applications

- **BR-6.1** Required: full name, phone, valid email, position (a published opening or "Open application"), CV. The CV is PDF, DOC or DOCX **by its bytes** (a DOCX must contain a Word document part), at most `careers.cv_max_mb` (5). The stored name is generated; the visitor's filename is kept only as a sanitised label.
- **BR-6.2** The application is stored and notified like BR-5.3. The email does **not** attach the CV (personal data stays in the CMS).
- **BR-6.3** Applications older than `careers.retention_months` (12) are deleted with their CVs by a daily job; each deletion is audited as `system:purge-applications`. Quotation requests are kept.

## BR-7 Inbox

- **BR-7.1** The inbox is read-only. Sales read quotation requests; HR read applications; superadmin both.
- **BR-7.2** A CV opens through a five-minute signed link; each open is audited.

## BR-8 Accounts and roles

- **BR-8.1** Roles and permissions (deny by default):

  | Permission | superadmin | editor | sales | hr |
  |---|---|---|---|---|
  | Content read/write, media | ✓ | ✓ | | |
  | Option and contact parameters | ✓ | ✓ | | |
  | System parameters (recipients, thresholds, analytics, security) | ✓ | | | |
  | Quotation requests | ✓ | | ✓ | |
  | Job applications and CVs | ✓ | | | ✓ |
  | Users, audit history, outgoing mail | ✓ | | | |

- **BR-8.2** With no users and a configured setup token, the first-run screen creates the first superadmin, once. Without the token it is closed. In development only, `sfg seed-dev` creates `admin@sfg.local`; it refuses to run in production.
- **BR-8.3** Sign-in errors are identical for an unknown email, a wrong password, a locked or a deactivated account. After `auth.max_failed_logins` (5) failures the account locks for `auth.lockout_minutes` (15). Sign-in is also rate limited per IP and per email.
- **BR-8.4** Nobody can demote or deactivate themselves, and the last active superadmin cannot be demoted or deactivated by anyone.
- **BR-8.5** Passwords are at least `auth.password_min_length` (12) characters. A password change or reset ends every session of that user.

## BR-9 Lists and exports

- **BR-9.1** Every list has a debounced search; filters where the data has a natural one.
- **BR-9.2** Every grid exports CSV with a **pipe** delimiter (RFC 4180 quoting, UTF-8 BOM), containing exactly the rows on screen (same search and filters), with references shown by name. Every cell starting `= + - @ TAB CR` is prefixed with `'`.
- **BR-9.3** Every disabled control states why, visibly (export with no rows, add where rows are fixed, delete of an image in use, publish of an expired certificate, save with missing fields, send with missing fields, clear with no filters).

## BR-10 Audit

- **BR-10.1** Every change in the CMS and every job deletion writes an audit row (actor, action, entity, before/after of changed fields, IP, user agent, request id) in the same transaction. The table is append-only, enforced by a trigger.

## BR-11 Brand colours

- **BR-11.1** A service line is refused unless its text-on-ink colour reaches 4.5:1 against its ink colour and its label colour reaches 4.5:1 on white. The CMS shows the live ratio.

## BR-12 Analytics

- **BR-12.1** GA4 loads only when `analytics.ga4_measurement_id` is set **and** the visitor accepted the consent banner. Never on `/admin`.
