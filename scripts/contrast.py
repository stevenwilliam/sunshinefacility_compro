#!/usr/bin/env python3
"""Measure the WCAG contrast of every text pairing the SFG site and CMS use.

impeccable §3: contrast is calculated, not eyeballed. This script is the
calculator, and .claude/skills/impeccable/design.md §3 is its recorded output.
It FAILS when:
  - a pairing that carries body text is below 4.5:1 (large/UI below 3:1), or
  - design.md does not carry the measured ratio for a pairing (so the sheet
    and the arithmetic can never silently disagree).

    python3 -I scripts/contrast.py            # check
    python3 -I scripts/contrast.py --table    # print the markdown table
    python3 -I scripts/contrast.py '#FFFFFF' '#C6241C'
"""

from __future__ import annotations

import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
DESIGN = ROOT / ".claude" / "skills" / "impeccable" / "design.md"

# (foreground, background, where, minimum). 4.5 = body text; 3.0 = large text
# (24px+, or 18.66px bold) and non-text UI such as focus rings and borders.
PAIRS = [
    # Light ground
    ("#1F211C", "#FFFFFF", "Body text on white", 4.5),
    ("#45473F", "#FFFFFF", "Secondary text on white", 4.5),
    ("#5A5C54", "#FFFFFF", "Muted text / labels on white", 4.5),
    ("#5A5C54", "#F4F4F1", "Muted text on Cotton", 4.5),
    ("#45473F", "#F4F4F1", "Secondary text on Cotton", 4.5),
    ("#C6241C", "#FFFFFF", "Kicker / link hover / placeholder tag (Strong Red) on white", 4.5),
    ("#A3181A", "#FFFFFF", "Error text on white", 4.5),
    ("#FFFFFF", "#C6241C", "Primary button label (white on Strong Red)", 4.5),
    ("#FFFFFF", "#A3181A", "Primary button hover label", 4.5),
    ("#FFFFFF", "#8A8B83", "Disabled button label (white on #8A8B83)", 3.0),
    ("#1F6B34", "#E3F0E5", "Success pill", 4.5),
    ("#FFFFFF", "#1E7A44", "WhatsApp contact card", 4.5),
    ("#FFFFFF", "#1E9E55", "WhatsApp float icon (non-text)", 3.0),
    ("#C6241C", "#FFFFFF", "Focus ring on light (non-text)", 3.0),
    # Dark ground
    ("#FFFFFF", "#1F211C", "Headings on Charcoal", 4.5),
    ("#D6D7CE", "#1F211C", "Body on Charcoal", 4.5),
    ("#C9CABF", "#1F211C", "Muted on Charcoal", 4.5),
    ("#A8A99F", "#1F211C", "Hero second sentence (#A8A99F, 40px+) on Charcoal", 3.0),
    ("#C9CABF", "#161813", "Footer text on Charcoal deep", 4.5),
    ("#FD572A", "#1F211C", "Tangerine kicker / icons / focus ring on Charcoal", 4.5),
    ("#FD572A", "#161813", "Footer link hover / placeholder tag on Charcoal deep", 4.5),
    ("#C6241C", "#1F211C", "Strong Red focus ring on Charcoal (NOT USED: fails 3:1)", 0.0),
    # Service-line ink fields (text always sits on the ink, never the base)
    ("#FFFFFF", "#C6241C", "Food: text on ink", 4.5),
    ("#FFFFFF", "#1F7360", "Digital: text on ink", 4.5),
    ("#FFFFFF", "#4A7023", "Workforce: text on ink", 4.5),
    ("#1F211C", "#FD572A", "Remote: text on ink", 4.5),
    ("#FFFFFF", "#A3181A", "Facility: text on ink", 4.5),
    ("#FFFFFF", "#EA2915", "White on Signal Red base (NOT USED for text)", 0.0),
    ("#FFFFFF", "#28937A", "White on Mint base (NOT USED for text)", 0.0),
    ("#B53A17", "#FFFFFF", "Remote label on white (Services page)", 4.5),
    ("#1F7360", "#FFFFFF", "Digital label on white", 4.5),
    ("#4A7023", "#FFFFFF", "Workforce label on white", 4.5),
    # CMS light
    ("#1F211C", "#F4F4F1", "CMS text on page ground", 4.5),
    ("#5A5C54", "#FFFFFF", "CMS faint text on surface", 4.5),
    ("#45473F", "#EFEFEA", "CMS draft pill", 4.5),
    ("#A3181A", "#FDEAE7", "CMS expired pill", 4.5),
    ("#C9CABF", "#1F211C", "CMS sidebar link", 4.5),
    ("#A8A99F", "#1F211C", "CMS sidebar group label", 4.5),
    # CMS dark
    ("#EDEDE6", "#1F211C", "CMS dark: text on surface", 4.5),
    ("#C9CABF", "#1F211C", "CMS dark: muted on surface", 4.5),
    ("#A8A99F", "#1F211C", "CMS dark: faint on surface", 4.5),
    ("#EDEDE6", "#2A2C26", "CMS dark: input text", 4.5),
    ("#FF8A73", "#1F211C", "CMS dark: error / danger", 4.5),
    ("#8FD19E", "#1E3A25", "CMS dark: success pill", 4.5),
    ("#FFB4A6", "#4A1D1A", "CMS dark: expired pill", 4.5),
    ("#D6D7CE", "#33352E", "CMS dark: draft pill", 4.5),
    ("#FD572A", "#1F211C", "CMS dark: focus ring (non-text)", 3.0),
]


def parse_hex(s: str) -> tuple[int, int, int]:
    s = s.strip().lstrip("#")
    if len(s) != 6:
        raise ValueError(f"not a #RRGGBB colour: {s!r}")
    return int(s[0:2], 16), int(s[2:4], 16), int(s[4:6], 16)


def _ch(c: int) -> float:
    v = c / 255.0
    return v / 12.92 if v <= 0.04045 else ((v + 0.055) / 1.055) ** 2.4


def luminance(hex_: str) -> float:
    r, g, b = parse_hex(hex_)
    return 0.2126 * _ch(r) + 0.7152 * _ch(g) + 0.0722 * _ch(b)


def ratio(a: str, b: str) -> float:
    la, lb = luminance(a), luminance(b)
    hi, lo = max(la, lb), min(la, lb)
    return (hi + 0.05) / (lo + 0.05)


def table() -> str:
    rows = ["| Foreground | Background | Use | Ratio | Needs | Result |", "| --- | --- | --- | --- | --- | --- |"]
    for fg, bg, use, need in PAIRS:
        r = ratio(fg, bg)
        res = "n/a (not used for text)" if need == 0 else ("PASS" if r >= need else "FAIL")
        rows.append(f"| `{fg}` | `{bg}` | {use} | **{r:.2f}:1** | {need if need else '—'} | {res} |")
    return "\n".join(rows)


def main(argv: list[str]) -> int:
    if len(argv) == 3:
        print(f"{ratio(argv[1], argv[2]):.2f}:1")
        return 0
    if len(argv) == 2 and argv[1] == "--table":
        print(table())
        return 0
    failures = []
    sheet = DESIGN.read_text(encoding="utf-8") if DESIGN.exists() else ""
    for fg, bg, use, need in PAIRS:
        r = ratio(fg, bg)
        if need and r < need:
            failures.append(f"FAIL {fg} on {bg} ({use}): {r:.2f}:1 < {need}:1")
        token = f"**{r:.2f}:1**"
        if f"`{fg}` | `{bg}` | {use} | {token}" not in sheet:
            failures.append(f"design.md does not record {fg} on {bg} ({use}) as {token}")
    if failures:
        print("\n".join(failures))
        print(f"\n{len(failures)} problem(s). Regenerate the table with --table and paste it into design.md §3.")
        return 1
    print(f"contrast: all {len(PAIRS)} pairings measured, pass, and match design.md")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
