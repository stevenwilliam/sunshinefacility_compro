# 10 — Design system

The numbers are in `.claude/skills/impeccable/design.md` (type scale,
palette, every measured contrast ratio, rules that are not taste). This
document is the reasoning and the component inventory. The approved design
is `design/handoff/`; the brand is the SFG brand guide only.

## Principles

1. **The design is the spec.** Values come from the `.dc.html` sources; where the README and a source file disagree, the source file's inline value wins (e.g. Home hero buttons are 8px radius).
2. **Measured, not eyeballed.** Page heights of every page were compared to the approved standalone design at 1440 and 360 px (`make shots`): within a few pixels on all pages except Careers (+135 px — the prototype omits its own "4 reasons to join" block; the build renders it as the README and COPY require).
3. **Text sits on the ink colour**, never the line base (white on Signal Red is 4.35:1, on Mint 3.79:1).
4. **Placeholders are visible** until real.

## Tokens

Tailwind theme (`web/tailwind.config.js`) carries the palette and the
breakpoints of the design (640, 760, 960, 1080, 1280). Component classes in
`web/src/index.css` carry the type scale, buttons, poster card, divider,
stats, form fields and the placeholder tag — component classes, so a single
utility cannot out-rank them by accident (impeccable §4). The CMS has its
own token set (`web/src/admin/admin.css`) with light and dark values.

## Components

| Component | File | Notes |
|---|---|---|
| Header | `components/Chrome.tsx` | sticky 72px, blur, nav ≥1080 else burger drawer (Esc closes), EN/ID pill (hit area 44px via pseudo element), CTA ≥640, skip link |
| Footer + divider strip + WhatsApp float | `components/Chrome.tsx` | placeholder tags beside contact details and social links |
| Signature ribbons | `components/Ribbon.tsx` | `patterns/signature-ribbon.html` transcribed exactly; reduced motion stops them |
| Poster card | `.poster*` in index.css | white clipped strip, ink field, grayscale photo masked by the line mark |
| Certificates | `components/Certificates.tsx` | Home cards + lightbox (portalled to body), About tiles, Food badges |
| Hero ball | `hero/scene.ts`, `hero/HeroBall.tsx` | port of the design's `makeHero()`; lazy; HUD, pill, labels, chips, detail panel |
| Forms | `pages/Contact.tsx`, `pages/Careers.tsx`, Home band | live "To send, add: …", field errors from the server in the visitor's language, honeypot |
| Icons | `lib/phosphor.tsx` (generated), `lib/icons.tsx` (editor-pickable set) | |

## Accessibility (WCAG 2.2 AA)

- Contrast: 47 pairings measured, all pass (design.md §3).
- Focus visible on every control: Strong Red on light, Tangerine on dark (Strong Red on Charcoal is only 2.84:1).
- One `h1` per page (e2e), headings in order, landmarks (header, nav, main, footer), skip link.
- Forms: real labels, `aria-invalid`, `aria-describedby` to the error, `aria-live` for the send status, `role=status` on success with focus moved to it, `role=alert` for failures.
- Dialogs: `role=dialog`, `aria-modal`, focus moved in and returned, Esc closes.
- Colour is never the only signal (pills carry text; expired dates also say "Expired").
- Reduced motion honoured (ribbons, ball spin and transitions, smooth scroll).
- Hit targets ≥ 44px.
- Decorative images are `aria-hidden`; content images have alt text in both languages from the media library.

## Deliberate differences from the prototype

| Difference | Why |
|---|---|
| Lightbox image height accounts for padding + caption | the prototype pushed Close below the fold on a 900px screen |
| Long placeholder tags wrap on phones | 431px tag overflowed a 360px screen |
| Chips under the ball hidden if WebGL fails | they would do nothing |
| About certificate tile meta = "Issued by …" for SLHS | one field drives Home and About (D34) |
| Home quick form continues to Contact | the README says it "links on"; the prototype only showed a thank-you (D28) |
