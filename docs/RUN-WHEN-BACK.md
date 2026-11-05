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

### M0 — switch production to the real domains (D41), in this order

Everything else is staged in `~ven/sfg-release/` on 119.8.189.246. Only these steps need you (Cloudflare login and sudo).

1. **Origin certificates.** For each zone (`sunshinefacility.com`, `sunshinefacility.co.id`, `sunshinefacility.id`): Cloudflare → zone → SSL/TLS → Origin Server → **Create Certificate** (RSA, hostnames `*.zone` and `zone`, 15 years). On the server, in a real terminal: `sudo mkdir -p /etc/ssl/sfg`, then `sudo vi /etc/ssl/sfg/<zone>.pem` (paste the certificate) and `sudo vi /etc/ssl/sfg/<zone>.key` (paste the private key). That makes six files.
2. **Cloudflare SSL mode.** For each zone: SSL/TLS → Overview → **Full (strict)**.
3. **Run the cutover**: `ssh maxx-prod`, then `sudo bash ~/sfg-release/prod-switch-domain.sh`. It checks the certificates, makes backups (binary, env, nginx, database dump), installs the new build, migrates, sets `SFG_PUBLIC_BASE_URL`, swaps nginx (restores the old file if `nginx -t` fails), checks every host at the origin, and prints the undo commands. Expect `200` for www.sunshinefacility.com and `301` for the rest.
4. **DNS, straight after.** For each zone: DNS → set `@` and `www` to **A `119.8.189.246`, Proxied (orange cloud)**. Delete any other A/AAAA/CNAME for `@`/`www` that point at the old Sunshinefood server. Until then the old domain redirects to the old Sunshinefood page.
5. Tell me. I'll verify all seven hostnames from outside.


| # | What | Steps |
|---|---|---|
| M1 | **Verify from your laptop** — `curl` on the server does not cross the firewall | open `http://<dev-server-ip>:8097/` and `/admin`; if it spins: `sudo grep -a 'DPT=8097' /var/log/ufw.log` |
| M2 | Mail goes out through **Office 365** as `noreply@sunshinefood.co.id` (done). Check that sunshinefood.co.id has SPF `include:spf.protection.outlook.com` and Microsoft DKIM enabled, or notifications may land in spam. Now that www.sunshinefacility.com is bought, either keep this sender or add that domain to Microsoft 365. The public email typo `info@sunshinefacilites.com` is fixed to `info@sunshinefacilities.com` by migration 0010 (next deploy); `sunshinefacility.com` has no MX yet, so mail there would bounce | Microsoft 365 admin → Domains; Defender → DKIM |
| M3 | Production is **deployed** (D37). Keep Cloudflare SSL/TLS mode on **Full (strict)** on every zone (Origin Certificates for the three real domains, Let's Encrypt for the temporary one). Store the `ven`, root and CMS passwords in your password manager, then rotate the Office 365 password, which went through chat. Mirror `/var/backups/sfg` and `/srv/sfg-minio` off the server | Cloudflare → SSL/TLS; `/srv/README.md` on the server |
| M4 | GA4 `G-XNBZ88NTP9` is set by migration 0009 (goes live with the M0 cutover). **In GA, turn off** Data streams → stream → Enhanced measurement → *Page changes based on browser history events*, or every page after the first counts twice. Then the § 3 settings (retention, internal traffic) | `docs/17-google-analytics-handbook.md` §§ 2–3 |
| M5 | Upgrade the system **npm** (10.8 crashes on an optional peer) | `sudo npm install -g npm@11`, then delete `web/.npmrc` and confirm `npm install` works |
| M6 | Upgrade the server's shared **Go** to ≥ 1.26.10 (this module pins its own toolchain meanwhile) | your call — it affects every project on the box |

## 3. Left alone on purpose

| What | Why |
|---|---|
| WhatsApp alerts, inbox statuses, 2FA, self-service reset, `/id/` URLs, CMS in Indonesian | phase 2 (`08-roadmap.md`) |
| Tailwind 4 | dev-only advisories in Tailwind 3's watcher; runtime is clean |
| `/metrics` | no monitoring stack yet |
| Test data in the dev database | e2e runs leave quotation requests from "E2E Tester" / "Budi Santoso" in the dev inbox; harmless, deletable by nobody (read-only inbox) — they age out only if they are applications |
