# Run when back

Three lists, kept apart on purpose.

## 1. Needs Steven (content and decisions)

| # | What | Where it goes |
|---|---|---|
| S1 | Real **address, phone, WhatsApp, email** | CMS → Parameters → Contact details (tags clear when you change them) |
| S2 | **Map pin** (latitude, longitude of the office) | Parameters → Contact details → map_lat / map_lng |
| S3 | **Social links** (Instagram, LinkedIn, Facebook, YouTube) | Parameters → Social links |
| S4 | Real **testimonials** (quote EN/ID, name, role, company), then publish | CMS → Testimonials |
| S5 | **Management** names, roles, photos | CMS → Management (+ Media for the photos) |
| S6 | Real **career openings** (the eight are samples) | CMS → Careers |
| S7 | Confirm **vision & mission** wording | Pages → About / vision; Lists → About · Mission |
| S8 | Higher-resolution **photos** (field app, mining camp; the deck's photos are 380–1855px) | Media → upload, then pick them in Services / Pages |
| S9 | Native **Bahasa Indonesia proofread** (copy was drafted from English) | Site text, Pages, Lists |
| S10 | Confirm **certificate issuers** and the **client-to-industry mapping** (both marked draft in the handoff) | Certificates; Clients / Industries |
| S11 | Renewed **halal certificate** (expired 25 Aug 2025) if it should return | Media + Certificates |
| S12 | Confirm decisions taken by default: retention **12 months** (Q9), reminder **60 days** (Q15) | Parameters, or tell me |
| S13 | Review the flagged deviations D11, D12, D34 in `00-README-and-decisions.md` | one word each |
| S14 | Whether to sync the incident rows in `.claude/skills/impeccable/design.md` §5 into the portable `99-steven-preference.md` log (would touch every project — your standing rule) | yes/no |

## 2. Needs a machine or access that does not exist yet

| # | What | Steps |
|---|---|---|
| M1 | **Verify from your laptop** — `curl` on the server does not cross the firewall | open `http://<dev-server-ip>:8097/` and `/admin`; if it spins: `sudo grep -a 'DPT=8097' /var/log/ufw.log` |
| M2 | **SMTP relay** for production + **SPF/DKIM** for `noreply@sunshinefacility.com` | relay credentials into `/etc/sfg/sfg.env` (`SFG_SMTP_*`); publish SPF (`v=spf1 include:<relay> ~all`) and the relay's DKIM record on sunshinefacility.com |
| M3 | **Production server, DNS, TLS** | `13-production-deployment-handbook.md`, top to bottom |
| M4 | GA4 measurement ID (only if you want analytics) | Parameters → Analytics → `G-…` |
| M5 | Upgrade the system **npm** (10.8 crashes on an optional peer) | `sudo npm install -g npm@11`, then delete `web/.npmrc` and confirm `npm install` works |
| M6 | Upgrade the server's shared **Go** to ≥ 1.26.10 (this module pins its own toolchain meanwhile) | your call — it affects every project on the box |

## 3. Left alone on purpose

| What | Why |
|---|---|
| Production deployment | D24: handbook only |
| WhatsApp alerts, inbox statuses, 2FA, self-service reset, `/id/` URLs, CMS in Indonesian | phase 2 (`08-roadmap.md`) |
| Tailwind 4 | dev-only advisories in Tailwind 3's watcher; runtime is clean |
| `/metrics` | no monitoring stack yet |
| Test data in the dev database | e2e runs leave quotation requests from "E2E Tester" / "Budi Santoso" in the dev inbox; harmless, deletable by nobody (read-only inbox) — they age out only if they are applications |
