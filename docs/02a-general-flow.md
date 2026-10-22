# 02a — The flows, drawn

The rules in `02` say what must be true; these diagrams show the shape.

## 1. The whole thing

What **flows** (solid) and what **constrains** (dotted).

```mermaid
flowchart LR
  Editor([Editor]) -->|edits| CMS[CMS /admin]
  CMS -->|writes, audited| DB[(PostgreSQL)]
  CMS -->|uploads, re-encoded| S3[(MinIO, private)]
  DB -->|published rows| Bundle[Public bundle]
  Params[[Parameters]] -.->|recipients, limits, contact| Bundle
  Today[[Today in Jakarta]] -.->|hides expired certificates| Bundle
  Bundle -->|embedded in HTML| Site[Public site]
  S3 -->|/media/id/hash-w.webp| Site
  Visitor([Visitor]) --> Site
  Site -->|quotation / application| API[API]
  API -->|stored + outbox, one transaction| DB
  Worker[Mail worker] -->|reads outbox| DB
  Worker -->|SMTP| Team([Sales / HR inbox email])
  Sales([Sales]) -->|reads| Inbox[CMS inbox]
  HR([HR]) -->|reads, opens CV| Inbox
  Inbox --> DB
```

## 2. Setup order, and what happens if a box is skipped

```mermaid
flowchart TD
  A[Create DB + role] --> B[sfg migrate up]
  B --> C[sfg seed-assets design/assets]
  C --> D[First superadmin: setup screen or seed-dev]
  D --> E[Superadmin sets notify.* recipients and sender]
  E --> F[Editors replace placeholders]
  F --> G[Launch]
```

| Skipped | What you see |
|---|---|
| `migrate up` | the service refuses to start: "migration(s) pending" |
| `seed-assets` | the site renders with **no images** — rows exist, files do not (no broken images: rows without a file are left out) |
| first superadmin | `/admin` shows the setup screen; without `SFG_SETUP_TOKEN` it is closed and nobody can sign in |
| `notify.*` | submissions are stored, mail goes to the seeded address (itdept) |
| placeholders | the site shows dashed **Placeholder** tags; the CMS sidebar counts them |

## 3. A certificate's life

```mermaid
stateDiagram-v2
  [*] --> Draft: created (published = false)
  Draft --> Published: publish (date in future or none)
  Published --> Draft: unpublish
  Published --> HiddenExpired: day after valid_until (no write)
  HiddenExpired --> Expired: shown as Expired in CMS
  Draft --> Expired: valid_until passes
  Expired --> Draft: new valid_until (renewal)
  Expired --> Expired: publish REFUSED (API 422, DB trigger)
```

## 4. A quotation request

```mermaid
sequenceDiagram
  participant V as Visitor
  participant S as Site (browser)
  participant A as API
  participant D as Database
  participant W as Mail worker
  V->>S: fills form
  S->>S: live check, "To send, add: …"
  S->>A: POST /public/quotation (+ honeypot, elapsed)
  A->>A: honeypot? fill time? rate limits?
  A->>A: validate (server is the authority)
  A->>D: one transaction - insert request, insert outbox mail
  A-->>S: 200 + first name
  loop every 30 s
    W->>D: claim due mail (FOR UPDATE SKIP LOCKED)
    W->>W: SMTP send
    W->>D: sent, or back-off and last error
  end
```

## 5. The refusal paths

```mermaid
flowchart TD
  R[Request] --> H{Honeypot filled?}
  H -- yes --> OK200[200, stored nowhere · BR-5.4]
  H -- no --> T{Faster than min fill?}
  T -- yes --> E422a[422 too_fast · BR-5.4]
  T -- no --> L{Over IP or email limit?}
  L -- yes --> E429[429 · BR-5.4]
  L -- no --> V{Valid?}
  V -- no --> E422b[422 per field · BR-5.1]
  V -- yes --> Store[Stored + outbox]
  C[CMS write] --> P{Permission?}
  P -- no --> F403[403 · BR-8.1]
  P -- yes --> X{Rule}
  X -- expired cert publish --> E422c[422 with date · BR-3.2]
  X -- colour below 4.5:1 --> E422d[422 with ratio · BR-11.1]
  X -- delete image in use --> E409[409 with count · BR-4.3]
  X -- delete fixed key --> E409b[409 · BR-9.3]
  X -- ok --> W2[Write + audit · BR-10.1]
```

## 6. Where a number on the site comes from

```mermaid
flowchart LR
  Stats[(stats rows, scope=home)] --> Bundle
  Facts[(service_facts per line)] --> Bundle
  Text[(site_text)] --> Bundle
  Bundle --> Page["Rendered in the visitor's language: value of en or id"]
```

The "10+ years" on Home and on About are **different rows** (`stats.scope`),
so changing one does not change the other — as in the design, where the
labels differ.
