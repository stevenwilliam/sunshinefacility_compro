# Handoff: Sunshine Facilities Group website + admin CMS

## Overview
A bilingual (EN/ID) company-profile website for Sunshine Facilities Group, an Indonesian integrated facility-services company and part of Lippo Indonesia. The audience is building owners, facility managers, and procurement/GA teams comparing vendors before requesting a quotation. There are 7 public pages plus an admin CMS through which all content and images are managed.

## About the design files
Everything in `design/` and `SFG Site (standalone).html` is a **design reference built in HTML**, a working prototype of the intended look and behaviour. It is not production code. Recreate it in the target stack (see `docs/99-steven-preference.md`) using that stack's patterns. Do not ship the `.dc.html` runtime (`support.js`).

- `SFG Site (standalone).html`: every public page in one offline file (hash routes `#/`, `#/services`, `#/service/<line>`, `#/industries`, `#/about`, `#/careers`, `#/contact`). Open it in any browser.
- `design/*.dc.html`: one file per page, the source of truth for exact values. Styles are inline. Copy text is in each file's `COPY = { en: {...}, id: {...} }` object. Opening them directly needs a local web server in `design/` (e.g. `npx serve design`).

## Fidelity
**High fidelity.** Colours, type, spacing, copy, motion and interactions are final, except the content flagged as placeholder under "Content status".

## Design tokens

### Colour
| Role | Hex |
|---|---|
| Charcoal (dark sections, primary text) | #1F211C |
| Charcoal deep (footer, hero edges) | #161813 |
| Dark surfaces / rules on dark | #2A2C26, #33352E, #3A3C35, #44463E, #4E5047 |
| White (page) | #FFFFFF |
| Cotton (alt section ground) | #F4F4F1 |
| Borders light → strong | #EFEFEA, #E6E6E1, #D6D6CF, #C4C4BC |
| Text secondary / muted | #45473F / #5A5C54 |
| Text on dark: body / muted / bright | #D6D7CE / #C9CABF / #EDEDE6 |
| Primary button (Strong Red) / hover | #C6241C / #A3181A, white text (5.7:1) |
| Accent on dark (Tangerine) | #FD572A |
| Disabled button fill | #8A8B83 |
| WhatsApp float / hover | #1E9E55 / #178A49 |
| WhatsApp contact card | #1E7A44 / hover #17663A |
| Success | #4A7023 (icons), #E3F0E5 bg + #1F6B34 text (pills) |

**Five business lines.** Each has a base colour, an ink colour used behind text, a shade used on marks, and an on-ink text colour:
| Line | key | base | ink (fields) | shade | text on ink |
|---|---|---|---|---|---|
| Food Services | food | #EA2915 | #C6241C | #C6241C | #FFFFFF |
| Digital & Operational Integration | digital | #28937A | #1F7360 | #1F7360 | #FFFFFF |
| Workforce Solutions | workforce | #5E9628 | #4A7023 | #4A7023 | #FFFFFF |
| Remote | remote | #FD572A | #FD572A | #E04E2B | #1F211C (5.0:1) |
| Facility Management | facility | #C6241C | #A3181A | #A3181A | #FFFFFF |

White on Signal Red #EA2915 (4.4:1) or Mint #28937A (3.8:1) fails AA for body text, which is why text always sits on the ink column.

### Type
The only font is **Nohemi** (variable, weights 100–900, `design/assets/fonts/Nohemi-VF.ttf`).
- Hero H1: `clamp(40px, 5.2vw, 74px)`, weight 600, line-height 0.98, letter-spacing −0.035em, `text-wrap: balance`
- Page H1: `clamp(38px, 5.4vw, 72px)`, 600, lh 1, −0.035em
- H2: `clamp(30px, 3.6vw, 48px)`, 550, lh 1.05, −0.03em (smaller H2s: `clamp(28px, 3.2vw, 42px)`)
- Card title: 26px / 550 / lh 1.08 / −0.02em
- Lead: `clamp(17px, 1.4vw, 19px)`, lh 1.55. Body 15–17px, lh 1.5–1.55.
- Eyebrow: 14px / 500 with a 28×3px bar before it (#C6241C on light, #FD572A on dark)
- Kicker label: 13px / 600 / uppercase / letter-spacing 0.06em
- Stats: `clamp(34px, 3.6vw, 48px)`, 550, −0.03em, with a 3px left border in #EA2915
- Use `font-variant-numeric: tabular-nums` on numbers, counters and times.

### Layout
- Container max-width 1280px. Side padding `clamp(20px, 4vw, 48px)`. Section vertical padding `clamp(56px, 7vw, 104px)`.
- Radius 4px everywhere. Pills and chips use 999px.
- Shadows: cards `0 10px 24px -10px rgba(31,33,28,.35)`; modals `0 30px 60px -20px rgba(0,0,0,.6)`.
- Hit targets are at least 44px. Focus ring: `2px solid #C6241C`, offset 3px.
- Breakpoints: inline header nav at ≥1080px (burger below that); header CTA hidden below 640px; hero side-by-side at ≥960px; mobile layouts below 760px.

## Brand patterns

**Poster card** (service cards, hero detail panel, Service Detail hero, Careers hero, About poster):
1. A white header strip with `clip-path: polygon(0 0, 100% 0, 100% 55%, 0 100%)` (the hero versions use 45%). It holds the line's mark (`assets/brand/sg-<line>.svg`) at 56–96px and a `01`–`05` counter in #5A5C54.
2. Below it, a colour field in the line's ink colour, with the title and content in its on-ink colour.
3. In the bottom-right corner, a **grayscale** photo masked by the line's mark SVG: `mask: url(sg-<line>.svg) center/contain no-repeat`, width about 66%, offset −5% right and bottom, `filter: grayscale(1)`.

**Colour divider strip** (above Why Sunshine and above the footer): a 5-column grid, height `clamp(64px, 9vw, 128px)`. Each tile is filled with a line's base colour. Inside, the line's mark is masked in its shade colour at left 28%, top −38% and width 120%, so it bleeds off the tile.

**Signature ribbon** (`patterns/signature-ribbon.html`, copy it exactly): a faint cursive word with two translucent gradient ribbons travelling along it on a 34s loop. Use "Ste" (still) behind the stats and client logos, and "ven" (floating) behind Certifications and Testimonials. The tiles above it get a solid white fill so it only shows in the gaps. Respect reduced motion.

## Global UI
- **Header** (sticky, 72px, `rgba(255,255,255,.96)` with blur, bottom border #E6E6E1): logo `sfg-logo-primary.png` at 34px tall, then nav (Home, Services, Industries, About Us, Careers, Contact). The active item has a 2px #C6241C underline and weight 550. On the right sit the EN/ID pill toggle (active state #1F211C with white text) and the Request a Quotation button. Below 1080px the nav becomes a burger that opens a full-width drawer with 20px links and the CTA.
- **Language:** stored in `localStorage['sfg-lang']`, applied site-wide without a reload, and `<html lang>` is updated. All copy exists in both languages in the design files.
- **Footer:** the colour divider strip, then a #161813 block with four columns: white logo, address and "Part of Lippo Indonesia." · Contact (phone, WhatsApp, email) · Quick links · Follow us (Instagram, LinkedIn, Facebook, YouTube as 44px outlined icon buttons). A bottom bar shows "© 2026 Sunshine Facilities Group" and www.sunshinefacility.com.
- **Floating WhatsApp button** on every page: 60px circle, #1E9E55, fixed `right/bottom: clamp(16px, 2vw, 28px)`.
- **Icons:** Phosphor (regular + fill), vendored in `design/assets/vendor/phosphor/`.

## Screens

### 1. Home (`design/SFG Home v2.dc.html`)
1. **Hero** (Charcoal radial gradient, height `clamp(600px, 100svh − 72px, 900px)` on desktop). The left column (max-width 540px) holds the eyebrow "Integrated facility services · Indonesia", the H1 "Clean, well-run buildings. *Every shift, on record.*" (second sentence in #A8A99F, weight 400), the lead, two buttons (Request a Quotation, solid red; Our Services, white outline), and the line "HACCP · ISO 22000:2018 · SLHS certified". The **hero ball** sits on the right at about 72% of the width. Below 960px the text stacks first, with the ball in a 460px area under it.
2. **Trust bar:** 4 stats (10+ years, 18+ cities, 1,000+ staff, 5M+ meals a year), then "Trusted by" and a grid of 19 client logos (118px minimum columns, 68px tiles). The logos are grayscale at 72% opacity and switch to colour on hover. The "Ste" ribbon sits behind.
3. **Services:** kicker, H2 "Five business lines, one team you can call.", intro, 5 poster cards (min 340px, 440px tall; title plus 4 sub-items; "Explore →"), and a 6th dark card "Not sure which services your building needs? → Book a site survey".
4. Colour divider strip.
5. **Why Sunshine** (Charcoal): 4 numbered points (trained supervised staff, scheduled preventive maintenance, digital reporting and proof of work, 24/7 response). On the right, a mock **work-order card** rotated −1.2°: WO number, "Completed" pill, site/task/crew/time list, before and after photos with timestamps, signature, and "GPS ✓". Its caption says it is sample data.
6. **Certifications:** 3 cards (ISO 22000:2018, HACCP CXC 1-1969, SLHS). Each has the top of its certificate image cropped in a 260px area, a description and the validity date. Clicking opens a full-size lightbox, closed by the backdrop, the Close button or Esc.
7. **Testimonials:** 3 dashed placeholder cards marked "Placeholder · replace with real quotes". The "ven" ribbon sits behind sections 6 and 7.
8. **Quotation band** (Charcoal): H2 "Tell us about your building.", WhatsApp and phone links, and a short form (name, company, building type, service chips) that links on to the full Contact form.
9. Footer.

### 2. Services (`Services.dc.html`)
A Charcoal hero with breadcrumb and H1 "Five business lines. One partner on site." Then 5 alternating rows, each a poster card (max 520px) beside the title, description, 6 sub-items in a grid with line-coloured squares, and "View service details →" underlined in the line colour. Then a Cotton band "Take one line or combine several." with 3 points (one contract, one site lead, one monthly report), and a CTA row with Request a Quotation and WhatsApp.

### 3. Service Detail (`Service Detail.dc.html`), one template for all 5 lines
Selected by route parameter (`?line=food|digital|workforce|remote|facility`; the default is facility).
- Breadcrumb, then a large poster hero (min-height `clamp(520px, 60vw, 620px)`) with the H1, lead, a white button (Request a Quotation) and an outlined WhatsApp button.
- A line switcher: pill links to the 5 lines, with the current one filled #1F211C.
- Key figures, where the line has them. Food: 5M+ meals, 80+ dietitians, 200+ cooks, 30+ projects. Workforce: 1,000+ / 18+. Remote: 5 hubs, 60 points, Timika. Facility: 18+, 10+, 24/7. Digital has none.
- "What's included": group cards, with either a list of sub-items or a paragraph, from the company profile.
- Remote only: a coverage map (`supply-map.png`). Food only: certification badges.
- A dark "Every job, on record." band with 4 proof points, a list of related industries as chips, and a CTA.

### 4. Industries (`Industries.dc.html`)
A Charcoal hero. Then 7 rows (Healthcare, Corporate offices, Retail & malls, Industrial & mining, Hospitality & F&B, Education, Residential towers). Each row has an icon, title, description, "Services we usually run" chips linking to Service Detail, and client logo tiles. Residential has no clients, so it shows a dashed "Talk to us" link. The page ends with a CTA.

### 5. About Us (`About.dc.html`)
- Intro and a "2015" poster in the Food colour.
- Stats strip and the story (two paragraphs from the company profile).
- A dark journey timeline: 2015 → 2017–2020 → 2021–2024 → 2025 → Today, each with a 3px top border in a line colour.
- Vision & mission, marked "Draft", and 3 values.
- The Lippo ecosystem: 7 sector cards on Cotton.
- Certifications: 3 cards.
- Management: 4 placeholder people with dashed photo slots.
- A dark CTA with Request a Quotation and Work with us.

### 6. Careers (`Careers.dc.html`)
- A poster hero in the Workforce colour, then 4 reasons to join.
- **Openings** (marked "Sample openings"): a search box plus Service line, Location and Type selects; a "Showing n of m" count; and a "Clear filters" button, disabled with "No filters applied" when nothing is set. Each row expands to show a summary, requirements and "Apply for this role", which prefills the form and scrolls to it. An empty state appears when nothing matches.
- **Apply form:** see the validation rules below. A success state follows sending.

### 7. Contact / Request a Quotation (`Contact.dc.html`)
- A Charcoal hero with H1 "Request a quotation."
- The form, with fields full name, company, work email, phone, building type, area size (m²), location, services needed (5 line chips) and message.
- A side column with a WhatsApp card, office/phone/email/hours, and a map with an "Open in maps" link.
- `?service=<line>` preselects a service chip. A success state follows sending.

### 8. Admin CMS (`Admin CMS.dc.html`)
A sidebar (#1F211C) on desktop and a section select on mobile. Sections: **Pages, Services, Clients, Certificates, Careers, Testimonials, Media library, Parameters.**
- Every list: a search box, an extra filter where relevant (parameter group, media type), a "Showing n of m" count, **Export CSV with a `|` delimiter** (disabled with the reason when there are no rows), Add, and per-row Edit/Delete.
- **Edit drawer** (580px, from the right): per-field types are text, textarea, select (options come from Parameters, or from Services for careers), image (picked from the Media library through a searchable picker), date, number and a Published toggle. Required fields are marked `*`. Save stays disabled with "To save, fill in: …" until they are complete. Delete asks for confirmation ("This cannot be undone.").
- **Rules:**
  - A certificate whose `validUntil` is before today shows as **Expired** and cannot be published; the toggle is disabled with the reason.
  - A media item in use cannot be deleted; the tooltip says how many items use it.
  - Seeded testimonials are unpublished.
- **Entities and fields:**
  - Page section: page, section, title EN/ID, body EN/ID, image, published, updated.
  - Service line: key, order, title EN/ID, description EN/ID, colour, photo, published.
  - Client: name, logo, industry, order, published.
  - Certificate: name, issuer, validUntil, image, published.
  - Career opening: title EN/ID, line, location, type, closing date, summary EN/ID, published.
  - Testimonial: quote EN/ID, name, role, company, published.
  - Media: url, name, type (Photo, Logo, Certificate, Brand).
  - Parameter: group, key, value EN, value ID. Groups are building types, industries, locations, job types, brand colours, contact details and stats.
- **Also needed (not designed yet):** login, roles, an inbox for quotation and career submissions, and audit history.

## Hero ball (Home), behaviour to reproduce
Built with three.js 0.160, which loads after first paint. Reference implementation: `makeHero()` in `SFG Home v2.dc.html`.

**Geometry**
- 5 interlocking **jigsaw pieces**, one per line, each a curved extruded slice of a sphere: radius 3, latitude span 0.84π, depth 0.22.
- The knob shapes come from a tab curve on each seam between neighbouring pieces.
- Each piece has a morph target from curved (ball) to flat (panel).

**Ball state**
- The photo faces are tinted with the line's base colour, and each piece's side walls are in its ink colour.
- It auto-rotates slowly, and the pieces gently "breathe" outward.
- Glowing amber seams pulse, with an additive halo sprite behind.
- A thin glowing outline ring and a ping ring sit around it, plus a "Click to open · or drag to rotate" pill under it.
- On hover everything brightens and the cursor becomes a pointer.

**Click → split** (about 2.1s, Iron Man style)
1. A shockwave ring fires.
2. Each piece flashes and pops out (0–12% of its own timeline), then bursts outward while tumbling to a scatter point about RING × 1.9 away (12–50%).
3. It flies in to its ring slot with a back-ease overshoot (50–100%), flattening as it goes, and flashes again on locking.
4. On locking, the face switches to the **poster texture**: white header strip, mark and masked photo.
5. In the centre, a sun core (icosahedron) appears with two rotating torus rings, rays to each panel, and a point light.

Pieces start in sequence, staggered by about 0.4 of the timeline across the five.

**Split state**
- The ring of 5 flat panels has radius 5.4 and labels that follow each panel.
- Drag to rotate (yaw 0.0065 rad per px with inertia; pitch is clamped to −0.55…0.6 and is off on touch).
- Clicking a panel or a chip brings it to the front: it moves toward the camera by RING × 0.55, scales to 1.3×, and the others dim by 60%.
- A poster-style **detail panel** opens (left side on desktop, bottom sheet on mobile) with prev/next/close buttons, "What's included" and both CTAs. Esc closes it, and the hero text fades out meanwhile.
- "Reassemble" reverses the split.

**Performance**
- Pixel ratio capped at 1.75.
- The render loop pauses when the hero is off-screen (IntersectionObserver) or the tab is hidden.
- Reduced motion means no auto-spin and an instant transition.
- On a WebGL or import failure the hero text still works.

## Forms and validation
- **Contact:** required fields are name, company, a valid email, phone, building type, location, and at least one service. Area accepts digits only. Submit stays disabled (#8A8B83) and live text reads "To send, add: …". On success: "Thank you, {firstName}. Our team will contact you within one working day…".
- **Careers:** required fields are name, phone, a valid email, position and a CV (PDF/DOC/DOCX, 5MB maximum, with an inline error if larger). There is an "Open application" option.
- Both forms must be stored server-side, emailed to the team, and visible in the CMS.

## Content status (do not publish as final)
- **Placeholder:** address ("Lorem ipsum dolor sit amet 12, Tangerang…"), phone +62 21 1234 5678, WhatsApp +62 1234 5678, email hello@sunshinefacility.com, testimonials, management names and photos, careers openings, and the map pin.
- **Draft wording to confirm:** vision and mission, the longer service bullet points, the client-to-industry mapping, and the certificate issuers.
- **The halal certificate expired on 25 Aug 2025:** seed it unpublished.
- ISO 22000 is valid to Sep 2027 and HACCP to Aug 2027, with the exact days assumed to be the end of the month. SLHS shows no expiry date.
- The Bahasa Indonesia copy was drafted from the English and needs a native proofread.

## Assets (`design/assets/`)
- `brand/`: `sfg-logo-primary.png`, `sfg-logo-white.png`, `sfg-logo-black.png` (1200px wide, from the client's files), plus `sg-food|digital|workforce|remote|facility.svg`, the 5 line marks redrawn as vectors from the brand guide.
- `clients/`: 19 client logos with transparent backgrounds, sliced from the client's logo sheet.
- `certificates/`: the first page of each certificate PDF (iso-22000, haccp-cxc, slhs, halal).
- `photos/`: photos taken from the company-profile deck. Digital and Remote use stand-in photos (mep-panel, technical); real field-app and mining-camp photos are wanted.
- `fonts/Nohemi-VF.ttf`: the licensed font supplied by the client.
- `vendor/`: three.module.js 0.160 and Phosphor icons (regular + fill).
- `docs/SFG-company-profile-2026.pptx`: the source company profile.

## Files
- `CLAUDE_CODE_PROMPT.md`: the prompt to paste into Claude Code.
- `SFG Site (standalone).html`: the offline preview of all public pages.
- `design/`: per-page design sources (`SFG Home v2`, `Services`, `Service Detail`, `Industries`, `About`, `Careers`, `Contact`, `Admin CMS`, shared `SiteHeader` and `SiteFooter`, `SFG Site` (combined), `Site Map` (all pages at 1440 and 360)).
- `patterns/signature-ribbon.html`: the exact ribbon SVG and CSS.
- `docs/99-steven-preference.md`: the owner's engineering and working preferences.
