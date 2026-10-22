# 15 — Admin guide

For whoever runs the system (superadmin and server operator).

## People

- **Users** (CMS): add with name, email, role and a first password; share it directly. Change role or deactivate in the same drawer. Reset a password there — their sessions end at once.
- You cannot demote or deactivate yourself, and the last active superadmin cannot be demoted. Make someone else superadmin first.
- A locked account (5 wrong passwords) unlocks by itself after 15 minutes, or immediately when you reset its password.

## Settings that change behaviour (CMS → Parameters)

| Group | What | Who |
|---|---|---|
| Notifications | `quotation_to`, `career_to`, `certificate_to` (comma-separated), `from` | superadmin |
| Certificates | `reminder_days` | superadmin |
| Careers | `retention_months`, `cv_max_mb` | superadmin |
| Forms and spam protection | `min_fill_seconds`, `per_ip_per_hour`, `per_email_per_hour` | superadmin |
| Media uploads | `max_upload_mb`, `max_dimension_px` | superadmin |
| Sign-in security | `password_min_length`, `max_failed_logins`, `lockout_minutes` | superadmin |
| Analytics | `ga4_measurement_id` (empty = none) | superadmin |
| Search and link previews | `og_image` (a media slug) | superadmin |
| Contact details, social links, building types, locations, job types, brand colours | site content | editor |

Values are type-checked: a retention of "twelve" or a recipient of "itdept"
is refused with the reason. Option groups take new rows; other groups have
fixed keys (their delete button says so).

## Mail

- CMS → **Outgoing mail** shows every notification: Queued, Sent, or Failed ×n with the last error. Failures retry for about a week; the submission is always safe in the inbox.
- Production needs SPF and DKIM for the sender domain, or Gmail will file the mail as spam.

## Jobs

| Job | Timer | Run now |
|---|---|---|
| certificate reminders | 07:00 Jakarta | `cd /home/dev/projects/sfg && make job JOB=certificate-reminders` |
| application purge | 02:30 Jakarta | `make job JOB=purge-applications` |
| mail delivery | every 30 s inside the service | `make job JOB=deliver-mail` |

`systemctl list-timers 'sfg-*'` shows when each runs next.

## Audit

CMS → **Audit history**: every change, upload, deletion, CV download,
password change/reset, lockout and job deletion — who, when, from which IP,
and the before/after values. Searchable and exportable. It cannot be edited
or deleted, by anyone, including through the database.

## Break-glass

- **Nobody can sign in** (every superadmin locked out, deactivated or gone): on the server
  ```bash
  cd /home/dev/projects/sfg
  set -a; . /etc/sfg/sfg.env; set +a
  printf '%s\n' 'a new long password' | /home/dev/projects/sfg/bin/sfg reset-password someone@sunshinefacility.com
  ```
  It sets the password, unlocks and reactivates the account, ends its sessions, and writes an audit row (`system:reset-password`). Read the password from a file or `read -s` rather than typing it into shell history.
- **Site shows no images**: the object store is down or empty — `docker ps | grep sfg-minio`, then `make seed-assets` (idempotent).
- **Service will not start**: `journalctl -u sfg -n 50`. "migration(s) pending" → `make migrate-up`. "checksum" → a migration file was edited after it ran; restore the file from git.

## Month-end / periodic

- Check *Outgoing mail* for failures.
- Check the sidebar's placeholder count is 0 (before launch).
- Review certificates' dates.
- Export the audit history if you keep an archive.
