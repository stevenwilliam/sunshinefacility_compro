Paste everything below the line into Claude Code (CLI), run from the folder that contains `design_handoff_sfg_website/`.

---

You are building the production website and admin CMS for **Sunshine Facilities Group (SFG)**, www.sunshinefacility.com.

**Read first, in this order, before writing any code:**
1. `design_handoff_sfg_website/docs/99-steven-preference.md`: my engineering DNA and working preferences. Follow it for stack, structure, questions, reporting and every UI rule.
2. `design_handoff_sfg_website/README.md`: the full design spec (pages, tokens, components, interactions, validation, CMS, assets).
3. Open `design_handoff_sfg_website/SFG Site (standalone).html` in a browser. It is the approved design, fully offline, all 7 public pages. Click through every page, both languages, the hero ball (click, drag, pick a panel) and both forms.
4. The per-page source designs are in `design_handoff_sfg_website/design/*.dc.html`. Read them for exact values: every style is inline, and copy text lives in the `COPY` objects in each file's logic class.

**What these files are:** HTML design references. Do not ship them or copy their runtime. Recreate the design in my stack (per the preferences file) at high fidelity: same layout, colours, type, spacing, copy, motion and behaviour.

**Scope:**
- Public site: Home, Services, Service Detail (one template, 5 lines), Industries, About Us, Careers, Contact / Request a Quotation. Bilingual EN/ID with the header switch, a sticky header with the Request a Quotation button, and a floating WhatsApp button on every page. Mobile-first from 360px, WCAG AA.
- Home hero: the interactive 3D jigsaw ball (three.js) exactly as specified in README → "Hero ball". Keep the performance rules: load it after first paint, pause when off-screen or the tab is hidden, cap the pixel ratio, honour reduced motion.
- Signature ribbon backgrounds ("Ste" behind the client list, "ven" behind Certifications + Testimonials): use `patterns/signature-ribbon.html` as-is.
- Backend + admin CMS: every text and image on the public site must be editable. The entities, fields, rules and screens are in README → "Admin CMS". The quotation and career forms must store submissions and notify by email, and the CMS needs a screen to view them.
- Seed the database with all content from the design files (both languages), the client logos, certificates and photos in `design/assets/`.

**Hard rules:**
- Brand is the SFG brand guide only: Nohemi font, Charcoal/Cotton base, the 5 line colours, Strong Red #C6241C buttons. Do NOT use the Nocturne design system or any other theme.
- The halal certificate expired 25 Aug 2025. Seed it unpublished, and enforce "expired certificates cannot be published".
- Placeholders must stay visibly marked until I replace them: address, phone/WhatsApp, testimonials, management names and photos, careers openings, vision & mission.
- Every list gets search. Every grid gets a pipe-delimited (`|`) CSV export. Every disabled control shows why it is disabled.

**How to work:** follow my preferences file. Ask every question you have in ONE batch up front, each with a proposed default, then build end to end without stopping. At the end, report what you verified and what you did not, plus any blockers.
