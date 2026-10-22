# 04 — API specification

Base path `/api/v1`. JSON in and out. The machine-readable contract is
`docs/openapi.yaml`; this page is the reading version.

## 1. Conventions

- Responses never cache (`Cache-Control: no-store`) except the public bundle (`no-cache` + `ETag`).
- Every response carries `X-Request-Id`; send your own to correlate.
- Timestamps are RFC 3339 UTC. Dates are `YYYY-MM-DD`.
- Lists accept `q` (search, every word must match), `f.<column>=` filters the resource declares, `limit`, `offset`, and `format=csv` (pipe-delimited export of the same rows).

## 2. Error model

```json
{ "error": { "code": "VALIDATION", "message": "Please correct the highlighted fields.", "fields": { "email": "email" } } }
```

| Code | HTTP | When |
|---|---|---|
| VALIDATION | 422 | bad input; `fields` maps field → reason. Public forms send machine codes (`required`, `too_long`, `email`, `phone`, `not_allowed`, `digits`, `out_of_range`, `control`, `file_type`, `file_size`; `_form`: `too_fast`/`invalid`) which the site renders in the visitor's language; the CMS gets English sentences |
| UNAUTHENTICATED | 401 | no/expired/revoked token; every sign-in failure (one message) |
| FORBIDDEN | 403 | role lacks the permission |
| NOT_FOUND | 404 | missing, malformed id, or not visible to the role |
| CONFLICT | 409 | duplicate, fixed rows, last superadmin |
| IN_USE | 409 | delete of something still referenced (message says how many) |
| CERTIFICATE_EXPIRED | 422 | database refused publishing (the API normally answers VALIDATION first) |
| RATE_LIMITED | 429 | with `Retry-After` |
| TOO_LARGE / UNSUPPORTED_TYPE | 413 / 415 | uploads |
| SETUP_CLOSED | 404 | setup already done or not enabled |
| INTERNAL | 500 | logged with the cause; the client sees a generic message |

## 3. Authentication

- `POST /auth/login {email, password}` → `{accessToken, accessExpiresAt, user}` + `sfg_rt` cookie (httpOnly, SameSite=Strict, Path=/api/v1/auth, Secure in production).
- `POST /auth/refresh` (cookie) → new access token, rotated cookie. Reusing a rotated refresh token revokes all the user's sessions.
- `POST /auth/logout` → revokes the access token's `jti` and the refresh token.
- `GET /auth/setup` → `{setupNeeded}`; `POST /auth/setup {token, email, name, password}` once.
- `GET /me`; `POST /me/password {current, password}`.
- Access token: HS256 JWT, 15 minutes, `Authorization: Bearer`. Permissions are read from the database on every request.

## 4. Public

| Method | Path | Notes |
|---|---|---|
| GET | `/public/bundle` | everything the site renders, both languages; ETag |
| POST | `/public/quotation` | JSON (BR-5); body also `website` (honeypot) and `elapsedMs` |
| POST | `/public/application` | multipart: fields + `cv` file (BR-6) |
| GET | `/media/{id}/{hash12}-{w}.webp`, `/media/{id}/{hash12}.svg` | immutable, 1 year |
| GET | `/robots.txt`, `/sitemap.xml` | |
| GET | any page | HTML with per-route head (title, description, canonical, OG, Twitter, JSON-LD) and the bundle embedded; unknown routes 404 with `noindex`; `?line=` on `/services` 301 → `/services/{line}` |

## 5. Admin (authenticated; permission per route)

| Method | Path | Permission |
|---|---|---|
| GET | `/admin/schema` — resource definitions, icons, parameter groups | content.read or params.content |
| GET | `/admin/options` — picker options | same |
| GET | `/admin/placeholders` — remaining per resource | content.read |
| GET/POST | `/admin/r/{resource}` | read: content.read (parameters: params.content); write: the resource's `writePerm` |
| GET/PUT/DELETE | `/admin/r/{resource}/{id}` | same |
| GET | `/admin/media` (`q`, `kind`, `format=csv`) | content.read |
| POST | `/admin/media` multipart `file`, `kind` | media.write |
| PUT/DELETE | `/admin/media/{id}` | media.write |
| GET | `/admin/quotations`, `/admin/quotations/{id}` | quotations.read |
| GET | `/admin/applications`, `/admin/applications/{id}` | applications.read |
| POST | `/admin/applications/{id}/cv` → `{url}` (5 min) | applications.read |
| GET/POST | `/admin/users`; PUT `/admin/users/{id}`; POST `/admin/users/{id}/password` | users.manage |
| GET | `/admin/audit` (`q`, `f.action`, `f.entity`) | audit.read |
| GET | `/admin/outbox` | audit.read |

Resources: `pages, text, services, service-items, service-facts,
service-groups, industries, clients, certificates, careers, testimonials,
management, lists, stats, parameters`. Their fields, types and limits are in
`internal/domain/resource/defs.go` and returned by `/admin/schema`.

## 6. Idempotency

No endpoint moves money or reserves capacity, so `Idempotency-Key` is not
implemented. A double-submitted form creates two rows; the per-email limit
(BR-5.4) bounds it.

## 7. Health

`GET /healthz` (process up) and `GET /readyz` (database and object store reachable).
