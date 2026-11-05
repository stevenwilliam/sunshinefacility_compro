# 17 — Google Analytics handbook

From no Google Analytics at all to visitor numbers in GA4, for
sunshinefacility.com. About 20 minutes. You need the Google account that
should own the data (`itdept.sfg@gmail.com`) and a **superadmin** login to
the CMS.

## What the site already does (nothing to code)

- **Switched on (D42).** The parameter `analytics.ga4_measurement_id` is
  **`G-YQ3537VDCW`** (migration 0013; 0009 had set an earlier stream's
  `G-XNBZ88NTP9`). Empty it and no Google script loads and no banner shows
  (BR-12.1, D23).
- **Consent first.** Once an ID is set, each visitor sees a bilingual
  banner (*Accept* / *Decline*). Google's script loads only after
  *Accept*. Ads storage, ad user data and ad personalisation are always
  denied. The choice is remembered in the browser (`localStorage`
  key `sfg-consent`).
- **Page views only.** The site is a single-page app, so it sends its own
  `page_view` on every navigation, with the full URL (so `?lang=id` shows). Form submissions (quotation, job
  application) are **not** tracked as events (see § 6).
- **Never in the CMS.** `/tytb` never loads analytics.
- The CSP already allows Google's tag and collection hosts.

Numbers will be **lower than real traffic**: visitors who decline, or who
ignore the banner, are never counted. That is the intended trade-off.

## 1. Create the GA4 property

1. Open <https://analytics.google.com>, signed in as `itdept.sfg@gmail.com`.
2. **Admin** (gear, bottom left) → **Create** → **Account**.
   - Account name: `Sunshine Facilities Group`.
   - Data-sharing settings: untick everything you don't need. *Technical
     support* is enough.
3. **Property**:
   - Property name: `sunshinefacility.com`
   - Reporting time zone: **Indonesia — (GMT+07:00) Jakarta**
   - Currency: **Indonesian Rupiah (IDR)**
4. Business details: industry *Business & Industrial Markets*, size *Large*.
   Objective: **Generate leads**. Then **Create** and accept the terms.

## 2. Add the web data stream

1. Choose platform **Web**.
2. Website URL: `https://sunshinefacility.com`, stream name
   `SFG website`.
3. **Enhanced measurement → gear icon.** Turn **off** *Page changes based
   on browser history events*. The site already sends a `page_view` on
   every navigation, so leaving this on **counts each page twice**
   (measured on 2026-11-05 on the first stream: `/services` arrived twice). The
   other options (scrolls, outbound clicks, file downloads, site search,
   form interactions, video) can stay on.
4. **Create stream.** Copy the **Measurement ID** from the top right. It
   looks like `G-AB12CD34EF`.

Ignore Google's "install the Google tag" instructions. The site already
has the tag. Never paste a `<script>` snippet anywhere.

## 3. Property settings worth changing once

All under **Admin**, in the property column:

| Setting | Where | Set to | Why |
|---|---|---|---|
| Data retention | Data collection and modification → Data retention | **14 months** | default is 2 months; exploration reports need longer |
| Google signals | Data collection and modification → Data collection | **off** | ad personalisation is denied anyway; keeps it simple under UU PDP |
| Granular location and device data | same page | on | city / device reports |
| Internal traffic | Data streams → SFG website → Configure tag settings → Define internal traffic | rule `traffic_type = internal`, **IP address equals** the office's public IP (search "what is my ip" from the office) | your own staff don't count as visitors |
| Data filter | Data collection and modification → Data filters → *Internal Traffic* | first **Testing**, after a week **Active** | applies the rule above |
| Users | Admin → Account access management → **+** | add colleagues as *Viewer* (or *Analyst*) | never share the owner login |

## 4. Switch it on in the CMS

Already done for `G-YQ3537VDCW` (migrations 0009 and 0013). Use these steps only to change the ID to another property.

1. Sign in to `https://sunshinefacility.com/tytb` as a **superadmin**.
2. **Parameters** → filter group **Analytics** → `ga4_measurement_id`.
3. Paste the Measurement ID (`G-…`) and save. Anything that does not look
   like `G-` plus 4–16 capitals and digits is refused, with the reason
   shown.

The change is live at once (no deploy, no restart) and appears in
**Audit history**.

## 5. Check it works

1. Open the site in a **private / incognito window** (so no earlier choice
   is remembered). The consent banner appears at the bottom.
2. Click **Accept**, then visit two or three pages.
3. In GA: **Reports → Realtime**. You should appear within about 30
   seconds, with the pages you opened.
4. For detail: install Google's *Tag Assistant* browser extension, or in
   developer tools → Network, filter `collect`. Each page should send one
   request to `…google-analytics.com/g/collect` with `en=page_view`. If you
   see **two** per page, step 2.3 was missed.
5. Click **Decline** in another private window. Nothing should go to
   Google.

To see the banner again in your own browser: developer tools →
Application → Local storage → delete `sfg-consent`.

## 6. Reading the numbers

| Question | Report |
|---|---|
| How many visitors, from where? | Reports → **Acquisition → Traffic acquisition** |
| Which pages are read? | Reports → **Engagement → Pages and screens** |
| Indonesian or English? | Pages and screens; pages opened with `?lang=id` carry it in *Page location* |
| Which cities / devices? | Reports → **User → Demographics / Tech** |
| Right now? | Reports → **Realtime** |

**Not available yet:** counting quotation requests or job applications as
*key events* (conversions). The real counts are in the CMS Inbox anyway.
Adding a `generate_lead` event on a successful quotation (and a
`job_application` event) is a small code change. Ask for it if you want
it in GA.

## 7. Search Console (recommended, free)

This shows which Google searches lead to the site.

1. <https://search.google.com/search-console> → **Add property** →
   **Domain** → `sunshinefacility.com`.
2. Google gives you a **TXT record**. In Cloudflare → `sunshinefacility.com`
   → DNS → **Add record**: type TXT, name `@`, content the value Google
   gave. Click Verify.
3. **Sitemaps** → submit `https://sunshinefacility.com/sitemap.xml`.
4. In GA: Admin → Product links → **Search Console links** → Link.

The `.co.id` and `.id` domains only redirect, so they don't need their own
property.

## 8. Switching it off

Empty `ga4_measurement_id` in CMS → Parameters. The script and the banner
disappear for every visitor on their next page load. Data already
collected stays in GA until its retention period ends. To delete it:
GA Admin → property → *Move to trash*.

## Known gaps (privacy)

- A visitor who chose *Decline* or *Accept* cannot change their mind from
  the site. The only way is to clear browser storage. UU PDP expects
  consent to be withdrawable. The fix is a small "Cookie settings" link in
  the footer that reopens the banner.
- The site has no privacy-policy page that the banner could link to.

Both are product decisions. They are listed here so they are not
forgotten before analytics is switched on.
