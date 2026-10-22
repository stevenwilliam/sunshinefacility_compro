# 12 — Security

Target: OWASP ASVS v4 Level 2, every OWASP Top 10 (2021) category. Each
control names where it is implemented **and the test that proves it**.

## OWASP Top 10 map

| Category | Control | Where | Proven by |
|---|---|---|---|
| A01 Broken access control | Deny-by-default permission matrix; every admin route declares its permission; use-cases check again; parameters scoped by group; objects not visible to a role are 404 | `domain/access`, `adapter/http/server.go`, `app/content.go` | `TestBR8_1_Matrix`, `TestBR8_1_AuthorisationMatrix`, `TestIDORInbox`, `TestParametersTypedAndScoped` |
| A02 Cryptographic failures | argon2id (64 MiB, t=3, p=4); refresh tokens SHA-256 hashed; HS256 pinned; TLS 1.2+ at nginx; HSTS in production; Secure cookie in production | `platform/security`, nginx | `platform/security` tests, `TestJWTTamperingAndLogout` |
| A03 Injection | bound parameters everywhere; SQL identifiers only from `domain/resource` (pinned to plain identifiers by a test); allow-listed enums and keys; search wildcards escaped; mail headers refuse CR/LF; CSV formula guard; React escapes output, no `dangerouslySetInnerHTML`; JSON in `<script>` escaped by `json.Marshal` | `adapter/postgres`, `domain/resource`, `notify`, `csvexport` | `TestIdentifiersArePlain`, `TestBR5_FormAbuse`, `TestBR9_2_CSVExport` |
| A04 Insecure design | business invariants in the database (expired certificate trigger, FK RESTRICT, CHECKs, append-only audit); outbox so a relay outage loses nothing; abuse cases below | migrations | `TestBR3_2_…` (DB path), `TestBR4_3_…`, `TestBR10_1_AuditTrail`, `TestBR5_3_MailRetriedNotLost` |
| A05 Misconfiguration | CSP without `unsafe-inline`, `frame-ancestors 'none'`, nosniff, X-Frame-Options, Referrer-Policy, Permissions-Policy, COOP; config refuses a short JWT secret; SVG served with sandbox CSP; private bucket | `middleware.go`, `config` | `TestSecurityHeaders` |
| A06 Vulnerable components | govulncheck (clean for reachable code), staticcheck, gosec, `npm audit --omit=dev` (0); toolchain pinned 1.26.10 | `make audit`, `go.mod` | run 2026-10-22 (PROGRESS) |
| A07 Identification & auth | generic login errors; constant-cost path for unknown emails; lockout; per-IP and per-email rate limits (app + nginx); 15-min access JWT in memory; rotating refresh with reuse detection; jti revocation; sessions end on password change/reset; first-run setup token; no default admin in production | `app/auth.go` | `TestBR8_3_…`, `TestBR8_2_SetupOnce`, `TestJWTTamperingAndLogout`, `TestBR8_4_LastSuperadmin` |
| A08 Software & data integrity | migrations checksummed; uploads re-encoded; deterministic seed | `platform/database`, `adapter/media` | `TestBR4_1_Uploads` |
| A09 Logging & monitoring | structured logs with request id, no PII; append-only audit of every change, CV download, lockout and job deletion | `logging`, `audit_log` | `TestBR10_1_AuditTrail`, `TestBR6_3_RetentionPurge` |
| A10 SSRF | the server fetches nothing from user-supplied URLs; social links are validated `https://` and only rendered as links | `domain/params` | `params` tests |

## Abuse cases

| Abuse | Control |
|---|---|
| Form flooding / spam | honeypot (silent), minimum fill time, per-IP and per-email hourly limits from stored rows (survive restarts), nginx edge limit |
| Malicious file as CV | magic bytes (PDF/OLE/DOCX-with-word-part), size cap before full read, generated storage name, never served publicly, 5-minute signed link |
| Polyglot / script in an image | magic bytes, decode with a dimension cap, re-encode to WebP (original private), SVG refused |
| Account enumeration | one message for every failed sign-in; dummy hash on unknown email |
| Password spraying | per-email limiter + lockout + nginx limit |
| Stolen refresh token | rotation; reuse revokes all sessions of that user |
| Editor escalation | parameters' system groups (recipients, thresholds, analytics) need superadmin; users/audit superadmin-only |
| Locking everyone out | the last active superadmin cannot be demoted or deactivated; nobody can demote themselves |
| Publishing an expired certificate via SQL | database trigger |
| Formula injection in exports | every cell guarded |
| Header injection via form fields into mail | CR/LF refused in headers; subjects collapsed to one line |
| Fake data in search results | placeholder contact details excluded from JSON-LD |
| Personal data kept forever | automatic deletion after the retention period, audited |

## ASVS L2 notes

- V2 (authentication): length-based policy (≥ 12, parameter), no composition rules, no default credentials; 2FA not in phase 1 (D14).
- V3 (sessions): token lifetimes 15 min / 7 days; logout revokes both; SameSite=Strict + Bearer makes CSRF inapplicable.
- V5 (validation): server re-validates everything; normalise before validating; reject, never repair.
- V7 (errors/logging): typed errors, generic 500s, causes only in logs.
- V8 (data protection): CVs private and time-limited; PII excluded from logs and URLs; retention enforced.
- V12 (files): see abuse cases.
- V14 (config): secrets only in `/etc/sfg/sfg.env`; `.env.example` documents the surface.

## Known and accepted

| Item | Status |
|---|---|
| Dev-only npm advisories (braces/micromatch/postcss-selector-parser via Tailwind 3's file watcher) | build tool only, never shipped; runtime audit is clean; Tailwind 4 migration is phase 2 |
| 1 govulncheck advisory in a required module | not reachable from our code |
| Refresh cookie not `Secure` on the plain-HTTP dev port | production and any HTTPS request set it |
| No `/metrics` | phase 2 |

## Rotation

JWT secret: change `SFG_JWT_SECRET`, restart — every session ends.
Database password: `ALTER ROLE sfg_app PASSWORD …`, update the env, restart.
MinIO keys: recreate the container with new root credentials, update the env.
Setup token: blank it after the first superadmin exists.
