# 06 — Domain operations

## Content

- **Where each thing on the site is edited** — `14-user-guide.md` has the page-by-page map. In short: headings and paragraphs of a section → *Pages*; every label, button and message → *Site text*; the five lines → *Services* (+ tags, figures, "what's included"); sectors → *Industries*; logos → *Clients*; *Certificates*; openings → *Careers*; quotes → *Testimonials*; people → *Management*; small lists (Why Sunshine, journey, mission, values, ecosystem, …) → *Lists*; figure strips → *Stats*; phone/address/hours/map/social → *Parameters*.
- A change is live on the next page load (the bundle cache is invalidated on every write).
- **Placeholders**: the sidebar shows how many remain per section. Editing a placeholder's content clears its tag (BR-2.2); untick "Placeholder" to clear it without editing, or tick it to mark something as provisional.

## Certificates

- Renewal: upload the new first page in *Media*, open the certificate, choose the new image, set the new *Valid until*, publish.
- The reminder (BR-3.4) arrives `certificates.reminder_days` before expiry, once per expiry date.
- On the day after expiry the certificate leaves the site without anyone acting.

## Submissions and mail

- Each request/application is stored, then mailed via the outbox. *Outgoing mail* shows Queued / Sent / Failed ×n with the last error.
- Retry schedule: 1 min, 5 min, 30 min, 2 h, 6 h, then daily, up to 10 tries. After that the row stays failed for a person to see; the submission itself is safe in the inbox.
- Recipients: `notify.quotation_to`, `notify.career_to`, `notify.certificate_to` (comma-separated). Sender `notify.from`.

## Scheduled jobs

| Job | When | What | Re-run safe |
|---|---|---|---|
| mail worker | every 30 s inside `sfg serve` | sends due outbox mail (`FOR UPDATE SKIP LOCKED`) | yes |
| `certificate-reminders` | 07:00 Asia/Jakarta (timer) | queues reminders; then delivers | yes (ledger) |
| `purge-applications` | 02:30 Asia/Jakarta (timer) | deletes applications + CVs older than `careers.retention_months`, audits each | yes |

Run by hand: `make job JOB=certificate-reminders` (or `deliver-mail`, `purge-applications`).

## Media

- Upload as Photo, Logo, Certificate or Brand; the kind sets the renditions.
- Source photos from the company profile are small (380–1855 px wide); the
  pipeline never upscales, so the hero/poster photos are only as sharp as
  their sources. Better photos are on Steven's list (RUN-WHEN-BACK).
- Alt text (EN/ID) is per image, in the media library.

## Parameters that change behaviour

| Key | Default | Effect |
|---|---|---|
| `forms.min_fill_seconds` | 3 | faster submissions refused (0 turns the check off) |
| `forms.per_ip_per_hour` / `per_email_per_hour` | 10 / 5 | submission limits |
| `careers.cv_max_mb` | 5 | largest CV |
| `careers.retention_months` | 12 | application deletion age |
| `certificates.reminder_days` | 60 | reminder lead time |
| `media.max_upload_mb` / `max_dimension_px` | 15 / 8000 | upload limits |
| `auth.password_min_length` / `max_failed_logins` / `lockout_minutes` | 12 / 5 / 15 | sign-in security |
| `analytics.ga4_measurement_id` | empty | empty = no analytics at all |
| `seo.og_image` | photo/city-aerial | link-preview image (by media slug) |
