# glyph-roles

A new `glyph` package (agent role glyphs for subagent rows, task status marks
for the todo vocabulary) with its DESIGN.md §6 entry, so herdr-sidecar's dock
can render per-role subagent icons and TODOS marks from the kit instead of
inventing its own. Branch `feat/glyph-roles`, tag `v0.9.0`.

## Specs

- S0 (owner rule, verbatim): "Every TUI I build follows github.com/WilberC/tui-kit
  and its DESIGN.md (local clone: `~/Projects/Personal/tui-kit`). In Go, import
  its packages (`theme`, `text`, `chip`, `list`, `frame`, `spinner`) instead of
  writing private styles; in another language, reproduce its palette roles and
  rules. A new visual element or behaviour goes into tui-kit first, with its
  tests and its `DESIGN.md` entry; the product then bumps the dependency."
- S1 (herdr-sidecar `odd/tasks/dock-integrations.md` S3, condensed): one glyph
  per subagent role family, one cell wide, between the status mark and the name;
  the mark keeps owning state. Shipped (plain) set: explore `◎` (U+25CE, alt
  U+F06E), verify `✓` (U+2713, alt U+F00C), worker `⚒` (U+2692, alt U+F0AD),
  judge `⚖` (U+2696, alt U+F0E3), fix `✚` (U+271A, alt U+F0FA), review `⌕`
  (U+2315, alt U+F002), unknown `–` muted (the existing "no value" glyph).
  Nerd Font alternates recorded, not shipped.
- S2 (herdr-sidecar `odd/tasks/dock-integrations.md` S2, condensed): todo
  status marks — pending `○` (muted), in_progress `◐` (live/WIP), blocked `⊘`
  (red, note shows as `└ waiting: <note>`), done folded into `✓ N done`,
  dropped `✕` (muted, excluded from the counts).

## Tasks

| ID | Specs | Route | Commit |
| --- | --- | --- | --- |
| T1 | S0-S2 (glyph package, DESIGN.md §6, README, tests) | worker | v0.9.0 |

## Log

- L1 (2026-11-14): T1 built (commit pending, parent owns it). New `glyph/`
  package: exported constants for the seven role glyphs (Explore, Verify,
  Worker, Judge, Fix, Review, Unknown — Unknown is the kit's existing muted
  "no value" `–`) and the five status marks (Pending `○`, InProgress `◐`,
  Blocked `⊘`, Done `✓`, Dropped `✕`); `Role(agentID string) string` maps an
  agent id to its family case-insensitively — a family word contained anywhere
  in the id names the family (explore, verify, worker, judge, fix), the
  `jd-judge-*`/`jd-fix-*` prefixes name theirs explicitly, `review-*` is
  prefix-only so a product named "reviewer-extra" is not read as a review row —
  and anything else answers the muted Unknown glyph. `Status(status string)
  string` maps the todo vocabulary case-insensitively, accepts the dropped
  aliases `cancelled`/`canceled`/`abandoned` and the hyphen/space forms of
  `in_progress`, and answers Unknown `–` for a status word it does not know (a
  status the kit cannot name has no mark). Width verification: every glyph
  measured one cell with `text.Width` (x/ansi, go-runewidth tables) before any
  table was written, including `⚒` U+2692 and `⚖` U+2696 from the Misc Symbols
  block — East-Asian Ambiguous, but one cell in the common monospace fonts — so
  no glyph needed replacing. Tests: `TestRoleMapsAgentIDsToTheirFamilyGlyph`
  (every rule incl. unknown and case-insensitivity, 27 cases),
  `TestStatusMapsTodoStatusesToTheirMark` (all five statuses, the three
  aliases, case-insensitivity, unknown, 14 cases) and
  `TestEveryGlyphIsOneCellWide`, which pins the one-cell rule for every
  exported glyph against the kit's own width function. The shared `✓` (verify
  role glyph and done status mark) was kept, judged unambiguous by column: a
  status mark lives in the row's status column and a role glyph in the role
  column, and a row never shows two of either — documented in DESIGN.md §6.
  DESIGN.md §6 gained the two tables ("Agent role glyphs" with the Nerd Font
  alternate column marked not shipped, "Task status marks") and the width
  paragraph; README's Packages table gained the `glyph/` row. Evidence:
  test-first — RED observed as `no non-test Go files in .../glyph` (build
  failure), then GREEN: `go vet ./...` clean, `go test ./...` all 15 packages
  pass (glyph adds 3 tests). Known follow-up for herdr-sidecar: bump to
  tui-kit v0.9.0 and render `glyph.Role(id)` between the row's status mark and
  the agent name, `glyph.Status(status)` in the TODOS block.
