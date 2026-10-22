# SFG — design reference

The working numbers for Sunshine Facilities Group. `docs/10-design-system.md`
carries the reasoning; this is the sheet to have open while building. Every
ratio below was measured by `scripts/contrast.py`, and that script **fails
when a ratio it computes is not written here** — the sheet and the arithmetic
cannot drift apart (`make contrast`).

**Source of truth for the look:** the approved design in
`docs/design/handoff/` (README + `design/*.dc.html`). Brand is the SFG
brand guide only — never the Nocturne system or any other theme.

## 1. Type

**Nohemi** variable (100–900), self-hosted as woff2 (`web/src/assets/fonts`),
licence for web use confirmed by Steven (Q26).

| Role | Size | Weight | Line height | Tracking |
| --- | --- | --- | --- | --- |
| Hero H1 | clamp(40px, 5.2vw, 74px) | 600 | .98 | −0.035em, balance |
| Page H1 | clamp(38px, 5.4vw, 72px) | 600 | 1 | −0.035em, balance |
| H2 | clamp(30px, 3.6vw, 48px) | 550 | 1.05 | −0.03em |
| H2 small | clamp(28px, 3.2vw, 42px) | 550 | 1.05 | −0.03em |
| CTA H2 | clamp(30px, 3.8vw, 52px) | 550 | 1 | −0.035em |
| Card title | 26px | 550 | 1.08 | −0.02em |
| Lead | clamp(17px, 1.4vw, 19px) | 400 | 1.55 | — |
| Eyebrow | 14px + 28×3 bar | 500 | — | — |
| Kicker label | 13px uppercase | 600 | — | .06em |
| Stat | clamp(34px, 3.6vw, 48px), 3px left rule #EA2915 | 550 | 1 | −0.03em |

Numbers, counters and times use `tabular-nums`. Base line height is the
browser's `normal` (Tailwind's 1.5 is overridden in `index.css`).

## 2. Palette

| Token | Hex | Use |
| --- | --- | --- |
| Charcoal | #1F211C | dark sections, primary text |
| Charcoal deep | #161813 | footer, hero edges |
| Dark surfaces | #2A2C26 #33352E #3A3C35 #44463E #4E5047 | rules on dark |
| Cotton | #F4F4F1 | alternate section ground |
| Borders | #EFEFEA #E6E6E1 #D6D6CF #C4C4BC | light → strong |
| Text | #45473F / #5A5C54 | secondary / muted |
| On dark | #D6D7CE / #C9CABF / #EDEDE6 | body / muted / bright |
| Strong Red | #C6241C, hover #A3181A | buttons, kickers, focus |
| Tangerine | #FD572A | accent and focus on dark |
| Disabled | #8A8B83 | disabled button fill |
| WhatsApp | #1E9E55 / #178A49; card #1E7A44 / #17663A | |

**Lines** (base · ink · shade · text on ink · label on white):
Food #EA2915 · #C6241C · #C6241C · #FFFFFF · #C6241C —
Digital #28937A · #1F7360 · #1F7360 · #FFFFFF · #1F7360 —
Workforce #5E9628 · #4A7023 · #4A7023 · #FFFFFF · #4A7023 —
Remote #FD572A · #FD572A · #E04E2B · #1F211C · #B53A17 —
Facility #C6241C · #A3181A · #A3181A · #FFFFFF · #A3181A.
The CMS refuses a line whose text-on-ink or label-on-white is below 4.5:1
(BR-11.1).

## 3. Measured contrast

| Foreground | Background | Use | Ratio | Needs | Result |
| --- | --- | --- | --- | --- | --- |
| `#1F211C` | `#FFFFFF` | Body text on white | **16.25:1** | 4.5 | PASS |
| `#45473F` | `#FFFFFF` | Secondary text on white | **9.43:1** | 4.5 | PASS |
| `#5A5C54` | `#FFFFFF` | Muted text / labels on white | **6.79:1** | 4.5 | PASS |
| `#5A5C54` | `#F4F4F1` | Muted text on Cotton | **6.16:1** | 4.5 | PASS |
| `#45473F` | `#F4F4F1` | Secondary text on Cotton | **8.56:1** | 4.5 | PASS |
| `#C6241C` | `#FFFFFF` | Kicker / link hover / placeholder tag (Strong Red) on white | **5.72:1** | 4.5 | PASS |
| `#A3181A` | `#FFFFFF` | Error text on white | **7.77:1** | 4.5 | PASS |
| `#FFFFFF` | `#C6241C` | Primary button label (white on Strong Red) | **5.72:1** | 4.5 | PASS |
| `#FFFFFF` | `#A3181A` | Primary button hover label | **7.77:1** | 4.5 | PASS |
| `#FFFFFF` | `#8A8B83` | Disabled button label (white on #8A8B83) | **3.44:1** | 3.0 | PASS |
| `#1F6B34` | `#E3F0E5` | Success pill | **5.56:1** | 4.5 | PASS |
| `#FFFFFF` | `#1E7A44` | WhatsApp contact card | **5.35:1** | 4.5 | PASS |
| `#FFFFFF` | `#1E9E55` | WhatsApp float icon (non-text) | **3.46:1** | 3.0 | PASS |
| `#C6241C` | `#FFFFFF` | Focus ring on light (non-text) | **5.72:1** | 3.0 | PASS |
| `#FFFFFF` | `#1F211C` | Headings on Charcoal | **16.25:1** | 4.5 | PASS |
| `#D6D7CE` | `#1F211C` | Body on Charcoal | **11.20:1** | 4.5 | PASS |
| `#C9CABF` | `#1F211C` | Muted on Charcoal | **9.81:1** | 4.5 | PASS |
| `#A8A99F` | `#1F211C` | Hero second sentence (#A8A99F, 40px+) on Charcoal | **6.84:1** | 3.0 | PASS |
| `#C9CABF` | `#161813` | Footer text on Charcoal deep | **10.80:1** | 4.5 | PASS |
| `#FD572A` | `#1F211C` | Tangerine kicker / icons / focus ring on Charcoal | **5.09:1** | 4.5 | PASS |
| `#FD572A` | `#161813` | Footer link hover / placeholder tag on Charcoal deep | **5.60:1** | 4.5 | PASS |
| `#C6241C` | `#1F211C` | Strong Red focus ring on Charcoal (NOT USED: fails 3:1) | **2.84:1** | — | n/a (not used for text) |
| `#FFFFFF` | `#C6241C` | Food: text on ink | **5.72:1** | 4.5 | PASS |
| `#FFFFFF` | `#1F7360` | Digital: text on ink | **5.71:1** | 4.5 | PASS |
| `#FFFFFF` | `#4A7023` | Workforce: text on ink | **5.78:1** | 4.5 | PASS |
| `#1F211C` | `#FD572A` | Remote: text on ink | **5.09:1** | 4.5 | PASS |
| `#FFFFFF` | `#A3181A` | Facility: text on ink | **7.77:1** | 4.5 | PASS |
| `#FFFFFF` | `#EA2915` | White on Signal Red base (NOT USED for text) | **4.35:1** | — | n/a (not used for text) |
| `#FFFFFF` | `#28937A` | White on Mint base (NOT USED for text) | **3.79:1** | — | n/a (not used for text) |
| `#B53A17` | `#FFFFFF` | Remote label on white (Services page) | **5.86:1** | 4.5 | PASS |
| `#1F7360` | `#FFFFFF` | Digital label on white | **5.71:1** | 4.5 | PASS |
| `#4A7023` | `#FFFFFF` | Workforce label on white | **5.78:1** | 4.5 | PASS |
| `#1F211C` | `#F4F4F1` | CMS text on page ground | **14.74:1** | 4.5 | PASS |
| `#5A5C54` | `#FFFFFF` | CMS faint text on surface | **6.79:1** | 4.5 | PASS |
| `#45473F` | `#EFEFEA` | CMS draft pill | **8.18:1** | 4.5 | PASS |
| `#A3181A` | `#FDEAE7` | CMS expired pill | **6.70:1** | 4.5 | PASS |
| `#C9CABF` | `#1F211C` | CMS sidebar link | **9.81:1** | 4.5 | PASS |
| `#A8A99F` | `#1F211C` | CMS sidebar group label | **6.84:1** | 4.5 | PASS |
| `#EDEDE6` | `#1F211C` | CMS dark: text on surface | **13.82:1** | 4.5 | PASS |
| `#C9CABF` | `#1F211C` | CMS dark: muted on surface | **9.81:1** | 4.5 | PASS |
| `#A8A99F` | `#1F211C` | CMS dark: faint on surface | **6.84:1** | 4.5 | PASS |
| `#EDEDE6` | `#2A2C26` | CMS dark: input text | **12.01:1** | 4.5 | PASS |
| `#FF8A73` | `#1F211C` | CMS dark: error / danger | **7.07:1** | 4.5 | PASS |
| `#8FD19E` | `#1E3A25` | CMS dark: success pill | **6.98:1** | 4.5 | PASS |
| `#FFB4A6` | `#4A1D1A` | CMS dark: expired pill | **8.32:1** | 4.5 | PASS |
| `#D6D7CE` | `#33352E` | CMS dark: draft pill | **8.57:1** | 4.5 | PASS |
| `#FD572A` | `#1F211C` | CMS dark: focus ring (non-text) | **5.09:1** | 3.0 | PASS |

Notes. White on Signal Red (4.35) and on Mint (3.79) fail AA, which is why
text always sits on the **ink** colour, never the base. Strong Red as a focus
ring on Charcoal is 2.84:1, so every dark section switches the ring to
Tangerine (`.on-dark`, index.css). The disabled button label is exempt
(WCAG 1.4.3, inactive controls) and still clears 3:1.

## 4. Rules that are not taste

- Radius 4px everywhere; pills 999px. (Home hero buttons and the quote form
  use 8px — that is the design's own value, kept.)
- Container: content 1280px max **plus** side padding clamp(20px,4vw,48px) —
  the design is content-box (measured: 1184 vs 1280 at 1440 px).
- Hit targets ≥ 44px (the language toggle extends its hit area with a pseudo
  element so it keeps the design's visual size).
- Focus ring 2px, offset 3px; Strong Red on light, Tangerine on dark.
- Breakpoints: burger < 1080, header CTA hidden < 640, hero side-by-side ≥ 960,
  mobile layouts < 760.
- Signature ribbon: `patterns/signature-ribbon.html` as-is
  (`web/src/components/Ribbon.tsx`); reduced motion stops it.
- Placeholders are visibly tagged (dashed Strong Red / Tangerine on dark).

## 5. Incidents in this project (candidates for the portable log)

| What nearly shipped | Cause |
| --- | --- |
| Certificate lightbox Close button unclickable — the next section painted over the overlay | the overlay was rendered inside a section with `position:relative; z-index:1`, a stacking context; its z-index 80 only counted inside it. Fixed with a portal. |
| A white button label would turn red on hover | a global `a:hover` at (0,1,1) out-ranks a component class at (0,1,0); there is no global link hover colour now |
| CMS search text under its icon, selects stretched | `.cms .inp` at (0,2,0) beat Tailwind utilities at (0,1,0); variants are component classes |
| Integration tests "expired" after 15 minutes of wall time | the test clock was fixed while tokens were verified against the real clock |
| Mail retry never became due in tests | back-off computed from the app clock, due-check against the database's `now()` |
