# 01 — Product requirements

## Problem

Building owners, facility managers and procurement/GA teams compare vendors
before asking for a quotation. Sunshine Facilities Group (SFG) — food,
facility, workforce, remote-site and digital services, part of Lippo
Indonesia — needs a credible bilingual site that explains the five business
lines, shows the clients and audited certificates, and turns visits into
quotation requests and job applications. Marketing needs to change every
word and image without a developer.

## Personas

| Persona | Wants | Where |
|---|---|---|
| **Facility / GA manager** (visitor) | What SFG does, for buildings like mine; proof; a fast way to ask for a price | public site, Contact |
| **Job seeker** (visitor) | Openings, what is required, a simple way to apply with a CV | Careers |
| **Editor** (SFG marketing) | Change text and images in both languages safely | CMS: content, media |
| **Sales** | See quotation requests and call back | CMS inbox |
| **HR** | See applications, open CVs | CMS inbox |
| **Superadmin** (SFG IT) | Users, parameters, audit, mail delivery | CMS administration |

## Scope — phase 1

1. Public site, seven pages, high fidelity to the approved design, EN/ID, sticky header with quotation CTA, floating WhatsApp, mobile-first from 360px, WCAG AA.
2. Home hero: the interactive three.js jigsaw ball (load after first paint, pause off-screen/hidden, pixel ratio cap 1.75, reduced motion).
3. Signature ribbons ("Ste" behind client logos, "ven" behind certificates + testimonials).
4. Quotation and career forms: validated on both sides, stored, emailed (retrying outbox), visible in the CMS.
5. CMS covering every text and image; media library; parameters; inbox; users/roles; audit; outgoing mail.
6. Seed: every string of the design in both languages; client logos, certificates and photos.
7. Rules: expired certificates cannot be published; placeholders stay visibly marked; every list has search; every grid exports pipe CSV; every disabled control says why.
8. SEO baseline: per-route title/description, Open Graph, canonical, robots, sitemap, Organization JSON-LD, `<html lang>`.

## Out of scope (phase 1)

WhatsApp alerts (D7), CMS in Indonesian (D18), `/id/` URLs (D20), status
workflow in the inbox (D8), 2FA and self-service password reset (D14),
production deployment (D24), a dark public theme (D19).

## Requirements, testable

| Id | Requirement | Verified by |
|---|---|---|
| R1 | All seven pages render at 1440 and 360 px with no horizontal scroll | e2e `360 px`, screenshot comparison |
| R2 | Language switch applies without reload, persists, updates `<html lang>`; `?lang=id` works | e2e |
| R3 | Every public string is editable in the CMS | catalogue guard test; e2e CMS edit |
| R4 | Expired certificate cannot be published (API and DB) and disappears from the site | integration BR-3.2, BR-3.3; e2e |
| R5 | Forms validated server-side; stored; emailed; bots dropped | integration BR-5, BR-6; e2e |
| R6 | Every CMS list searchable and exportable as pipe CSV honouring the search | integration BR-9.2; e2e |
| R7 | Roles see only what they may | integration BR-8.1 matrix |
| R8 | Placeholders visibly tagged until changed | e2e; integration BR-2 |
| R9 | Hero ball as specified, with its performance rules | visual run under Xvfb (PROGRESS) |
| R10 | Link previews show the right title and image without JavaScript | integration SEO test (curl-shaped) |

## Success measures

Quotation requests per month; share of visitors reaching Contact; time from
request to first call (outside the system); placeholders remaining = 0 at
launch (CMS sidebar counter).
