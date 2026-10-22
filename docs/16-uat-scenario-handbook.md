# 16 — UAT scenario handbook

For the person signing the site off, alone, without a developer. Work top to
bottom: later scenarios build on earlier ones. ⛔ marks a refusal — the
system must say **no**, and say why.

**Site:** http://<dev-server>:8097/ · **CMS:** http://<dev-server>:8097/admin
**Accounts:** `admin@sfg.local` (superadmin; password in the access table of
the build report). Scenario A1 creates the other roles.

## A. Accounts

| # | Who | Do | Expect |
|---|---|---|---|
| A1 | superadmin | Users → Add user, three times: editor@, sales@, hr@ (role each) | three users listed, Active |
| A2 ⛔ | anyone | Sign in with a wrong password | "Email or password is incorrect." |
| A3 ⛔ | anyone | Sign in with an email that has no account | exactly the same message as A2 |
| A4 ⛔ | editor | Five wrong passwords, then the right one | still refused with the same message; after 15 minutes the right one works |
| A5 ⛔ | superadmin | Users → yourself → change role to Editor | the radio is greyed out: "You cannot remove your own superadmin access" |
| A6 | sales | Sign in | sees only Quotation requests (and Your account) |
| A7 | hr | Sign in | sees only Job applications |

## B. Public site

| # | Do | Expect |
|---|---|---|
| B1 | Open Home on a laptop | header, hero with the 3D ball loading after the text, logos, five service cards, Why Sunshine, certificates, dashed placeholder testimonials, quote form, footer, green WhatsApp button |
| B2 | Click the ball | it splits into five panels around a glowing core; drag rotates it |
| B3 | Click a panel or a chip | the panel comes forward, a details card opens with prev/next/close; Esc closes it; Reassemble puts the ball back |
| B4 | Press ID in the header | the whole site turns Indonesian without reloading, and stays Indonesian on other pages and after a refresh |
| B5 | Open `/about?lang=id` in a new private window | opens in Indonesian |
| B6 | Phone (or narrow window, 360px) | burger menu, everything readable, nothing scrolls sideways |
| B7 | Home → click a certificate | it opens large; Esc closes; Close closes; clicking outside closes |
| B8 ⛔ | Look for the halal certificate on Home, About, Food Services | it is not there (expired 25 Aug 2025) |
| B9 | Look at footer, Contact, About management, Careers, vision & mission | each provisional item carries a visible "Placeholder" / "Contoh" / "Draft" tag |
| B10 | Paste the Home link into WhatsApp | the preview shows the title, description and a picture |

## C. Forms

| # | Do | Expect |
|---|---|---|
| C1 ⛔ | Contact: leave everything empty | "Send request" is grey; under it "To send, add: full name, company, …" |
| C2 ⛔ | Type "5,000" in Area | only the digits stay (5000) |
| C3 | Fill everything, send | "Thank you, <first name>. Our team will contact you within one working day…" |
| C4 | Check the notification inbox | an email "Quotation request: <company> (<location>)" from noreply@sunshinefacility.com |
| C5 | CMS (sales) → Quotation requests | C3 is at the top; click shows all fields |
| C6 | Service Detail → Request a Quotation | Contact opens with that service already ticked |
| C7 | Home quote band: fill name, company, type, a service → Request a Quotation | Contact opens with those already filled |
| C8 ⛔ | Careers: attach a 6 MB PDF | red note "This file is over 5 MB…"; Send stays grey |
| C9 ⛔ | Careers: attach a .pdf that is really a picture or text file | after Send: "This file type is not accepted…" |
| C10 | Careers: open an opening → Apply for this role → fill, attach a real PDF, send | "Application received"; HR sees it; Open CV downloads it |
| C11 ⛔ | Send six quotation requests from the same email within an hour | the sixth is refused: "Too many requests…" |

## D. Content editing

| # | Who | Do | Expect |
|---|---|---|---|
| D1 | editor | Site text → search "Trusted by" → change English → Save | Home shows the new words at once |
| D2 | editor | Testimonials → open one → change the quote → publish | the Placeholder tag on that quote is gone; Home shows the real quote |
| D3 ⛔ | editor | Certificates → Halal → try Published | the switch is grey: "Can't publish: this certificate expired on 25 Aug 2025…" |
| D4 ⛔ | editor | Services → Remote → Text on ink = #FFFFFF → Save | refused: the ratio is shown and it needs 4.5:1 |
| D5 ⛔ | editor | Media → bin on a client logo | grey, tooltip "used in 1 item. Replace it there first." |
| D6 ⛔ | editor | Media → upload an SVG | refused: "Upload a PNG, JPEG or WebP image…" |
| D7 | editor | Media → upload a PNG photo; use it in Management for one person | the photo appears on About |
| D8 ⛔ | editor | Pages → Add | grey, with "Sections are fixed by the page layout…" |
| D9 ⛔ | editor | Parameters → Notifications | not visible (superadmin only) |
| D10 ⛔ | superadmin | Parameters → Careers → retention_months = "twelve" | refused: "must be a whole number" |
| D11 | anyone | Any list → search → Export CSV | file opens in Excel with columns split on `|`, only the searched rows |
| D12 ⛔ | anyone | Search for nonsense → Export CSV | grey, "Nothing to export: no rows match the current search." |

## E. Records

| # | Who | Do | Expect |
|---|---|---|---|
| E1 | superadmin | Audit history | D1–D7 are there with who, when, before and after |
| E2 | superadmin | Outgoing mail | the C4 mail is "Sent" |

## Sign-off

| Area | Pass / Fail | Notes |
|---|---|---|
| A Accounts | | |
| B Public site | | |
| C Forms | | |
| D Content | | |
| E Records | | |

Signed: ____________________ Date: ____________

**What a failure means.** A failed ⛔ scenario blocks launch (a control that
does not refuse is not a control). A failed visual detail may be deferred —
only with a named owner and a date written next to it.
