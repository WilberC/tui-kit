# DESIGN

The visual guide of tui-kit. This file is the canonical source both products
(herdr-suite and herdr-utils) follow; the Go packages are its executable form. If
a screen and this document disagree, the screen is wrong.

The rules exist because a TUI is read at a glance, often while the user is doing
something else. Everything here protects three things: **where the eye starts**
(the title, the selected row), **what a colour means** (live, failed, work in
progress), and **the grid** (nothing wraps, nothing floats, no column drifts).

---

## 1. Palette

Tokyo Night, roles only. No screen names a colour; a role is what a row asks for.
Fourteen roles, and they are all of them.

| Role | Value | Means | Never |
| --- | --- | --- | --- |
| text | `#c0caf5` | Prose: names, values, the primary text of a row. | Used for explanation. |
| muted | `#565f89` | Every secondary fact: workspace, counts, age, path, hint label. | Used for a primary fact. |
| guide | `#3b4261` | Structure that must recede: tree guides, separators, the body of a rule. | Used where a user must read it. |
| border | `#3d59a1` | The frame's structural blue: borders, corners, tees, joints, column rules. | Used for emphasis or for text. |
| accent | `#7aa2f7` | The product's voice: the surface's name, a filter marker, the selection's edge, a header label. | Used for state. |
| gold | `#e0af68` | State that outlives the selection: the default item, the group glyph, a group name, a status message. | Used for the selection. |
| teal | `#73daca` | Work in progress: the live view label, the spinner. | Used for success. |
| green | `#9ece6a` | Live things: a running session, the key that will run. | Used for decoration. |
| red | `#f7768e` | Failure, and nothing else. One red fact per screen at most. | Used for emphasis. |
| chip bg | `#2f3549` | The slab a token sits on. | Painted behind prose. |
| selection bg | `#283457` | The tint of the selected row. | Painted as a solid bar. |
| chip on selection | `#3d59a1` | The slab of a token sitting on the tint: one step brighter. | Reused off the tint. |
| muted on selection | `#737aa2` | Muted facts on the tint. | Used off the tint. |
| guide on selection | `#5f6b96` | The tree guide on the tint. | Used off the tint. |

The three accents carry the interface (accent, gold, teal); green means live, red
means failure; everything explanatory recedes to muted. A screen that needs a
fourth accent is a screen with too much on it.

Why the two blues: `border` is the frame's structure and stays one step under the
content; `accent` is the product's voice and marks what the user is looking at.
They are separate roles because a frame drawn in the accent's brightness competes
with the selected row.

## 2. Frame anatomy

Top to bottom, every surface is the same seven parts:

```
╭─ product › surface ─────────────────────────────────╮   top border + title
│ summary fact · another              KEYS · LIST · 4 │   header row
├─────────────────────────────────────────────────────┤   rule (no joints)
│ ❱ filter▏ query text     x after prefix  ctrl+x ... │   filter + legend row
├──────────────────┬──────────────────────────────────┤   column rule (joints)
│                  │                                  │   one blank row
│ ● GROUP  6       │  ● OTHER  3                      │   group header
│   ├─  chip  desc │    └─  chip  desc                │   entries
│                  │                                  │   (one blank row between groups)
├──────────────────┴──────────────────────────────────┤   column rule (joints)
│ detail · fact · source                 4 / 4 shown  │   footer line
│  enter run   ↑↓ move   type filter   esc close      │   hint bar
╰─────────────────────────────────────────────────────╯   bottom border
```

Rules of the anatomy:

- **The title lives in the top border**, never as a row. It is
  `gold bold name` + `guide › ` + `accent bold surface`, so the product is gold
  and the surface is blue at a fixed location on every screen.
- **One cell of padding** between a border and its content, on every row. Content
  never touches the frame.
- **The rule under the header has no joints** (├──┤). The columns have not opened
  yet, so a joint would announce a column that does not exist.
- **The column rules carry joints** (┬ above, ┴ below), so the vertical rules grow
  out of the frame. A column rule painted flat leaves the verticals floating.
- **One blank row** separates the opening column rule from the first group
  header, and one blank row separates two groups. Never two.
- **The footer is two lines, never three**: the detail line (what is selected) and
  the hint bar (what the keys do). A status message rides the detail line at the
  right edge; it never adds a row.

## 3. Chips

A chip is a **padded slab**, not a bracketed word:

```
[x]      ✗   brackets read as literal text
 x       ✗   a key with no surface is just a word
 x       ✓    bold label, one cell of padding each side, on #2f3549
```

Rules:

- The background carries the padding: `slab.Render(" " + label + " ")`.
- A label follows its chip after **exactly one plain space**.
- Chips that share a column share a **chip width** (`chip.Cell`), so every
  description starts on the same cell. The column is capped (14 cells) and never
  eats the description's minimum room.
- An entry with **no key of its own** keeps the column with a dim `–` and no slab
  (`chip.UnsetCell`). A group of such entries must not read as a wall of empty
  boxes — and must never claim a key it does not have (`enter` is not "the key"
  of a discovered action).
- Hints are chips too (`chip.Hints`): the key is a slab, the label is muted
  prose, items are separated by **two plain spaces**. A hint that does not fit is
  **dropped whole**, from the end; the bar never wraps. The first hint is never
  dropped — at an extreme width it is clipped to the bar, because a footer with
  no keys at all is worse than one cut label.

## 4. Selection

The selection is a **tint**, not a painted bar:

```
▌ ├─  prefix+k  Command bar            ← #283457 behind the row, ▌ in cell 0
  ├─  prefix+p  Plugin manager         ← panel surface
```

- The row gains the background `#283457`; **every element keeps its own role**.
  The name becomes `text` bold on the tint, secondary facts become
  `muted on selection`, the tree guide becomes `guide on selection`.
- Green, gold, teal and red keep their meaning on the tint (`OnSelection`): a gold
  default marker stays gold, a green live dot stays green.
- The **accent `▌` occupies the row's first cell**, replacing the indent. Every
  other cell keeps its x, so the columns do not shift when the cursor moves.
- The tint runs to the row's full width. A tint that stops mid-row reads as a
  broken highlight.
- A chip on the selected row uses `chip on selection`, one step brighter, so the
  token still reads as a token on the darker background.
- **Exactly one row** carries the tint on a screen.
- The gold `◆` is for state that outlives the selection (the default item); it is
  quieter than the selection and is not a second cursor.

## 5. Rhythm and spacing

- **One blank row between groups.** Two blank rows read as "something is missing
  here"; none reads as one long list.
- **One blank row above the first group**, under the opening column rule.
- Gaps between columns belong to the row's own background: on a selected row they
  are tinted (`Roles.Gap`), never plain spaces.
- Separators are ` · ` in the guide role: one space on each side, never two, and
  never touching a chip's background.
- The right side of a line is where short, stable facts live (counts, page,
  summary). When a line must give something up, the **left** side is clipped
  first; the summary stays flush with the inner edge.
- Nothing ends a line with trailing spaces outside the frame's padding, and no
  row repeats a space before a separator.

## 6. Typography

One glyph, one meaning, one cell wide — so a glyph never shifts a column.

| Glyph | Meaning | Role |
| --- | --- | --- |
| `…` | an honest cut | current |
| `–` | no key / no value | muted |
| `·` | separator between two facts | guide |
| `›` | the product → surface cut in a title | guide |
| `●` | live (also the group glyph) | green / gold |
| `○` | idle | muted |
| `◆` | state that outlives the selection | gold |
| `▌` | the selected row's edge | accent on tint |
| `▏` | the live filter cursor | accent |
| `❱` | the filter marker | accent |
| `├─` `└─` | tree guide, middle / last of a group | guide |
| `▖▘▝▗` | work in progress | teal |

Tree guides **close** a group: the last entry of a group is `└─`, the others
`├─`. A list that trails off with `├─` looks unfinished.

## 7. Responsive rules

The frame owns its width. Whatever the terminal does, no line wraps and no line
exceeds the frame.

| Width | Columns | What is dropped first |
| --- | --- | --- |
| 180 | 3 when the surface holds three groups | nothing; the header may carry the page |
| 140 | 2 | a fourth column is never invented |
| 80 | 1 | the legend, then the longest hints; the first hint always survives |

- Column count follows the inner width (`frame.ColumnCount`): 3 from 148 inner
  cells, 2 from 98, 1 below — and never more than the surface has groups.
- Columns split the inner width evenly, one cell of rule between them
  (`frame.ColumnWidths` + `frame.ColumnRule`), so a rule — never a guess —
  decides where a column starts.
- A body shorter than its content **pages** rather than squeezing: the header
  names the page (`page 1 / 3`), and the selection decides which page is shown.
- A hint, a fact or a chip is dropped **whole**. The only thing that is ever cut
  is a single string, and a cut ends with `…`.

## 8. States

Every surface must have an answer for three states, and none of them is a blank
box.

**Empty** — nothing to list yet. One centred line naming the state, one dim line
naming the next step, one blank row above and below:

```
│                                                                  │
│                      Nothing here yet                            │
│                  press enter to create one                       │
│                                                                  │
```

**Zero match** — a filter that matches nothing. Name the query, offer the way
out, and never draw empty groups:

```
│                                                                  │
│            No binding matches “nothing-matches-this”             │
│                                                                  │
│     ctrl+u clears the filter · esc closes the guide              │
│                                                                  │
```

**Loading** — work in flight. A teal spinner (`spinner.Render`) on the line that
describes the work; the caller owns the tick, so the screen costs nothing while
nothing is happening. Never a full-screen "loading" placeholder.

## 9. Anti-patterns we fixed

These are real regressions from the two products. Do not bring them back.

| Anti-pattern | Why it failed | The rule now |
| --- | --- | --- |
| Solid selection bar (`bg` over the whole row, one foreground) | Every colour on the row died; state stopped meaning anything. | A tint; each element keeps its role. |
| Selection without the `▌` edge | The eye could not find where the row starts. | The accent edge in cell 0. |
| Selection that painted the tree guide flat | The guide disappeared and the row looked unparented. | `guide on selection` lifts it instead. |
| Chip slab in the panel colour on the tint | The token sank into the selection. | `chip on selection`, one step brighter. |
| Bracketed chips (`[enter]`) | Reads as literal text, not as a token. | Padded slab, no brackets. |
| Chip label glued to the slab (`enterrun`) | No separation between token and meaning. | Exactly one plain space, then the label. |
| Caret key names (`^K`) | Not what the key is called, not searchable. | `ctrl+k`. |
| Separators painted with a chip background (` · ` inside the slab) | The gutter became a second chip. | Two plain spaces between items. |
| Rules without joints (floating rules) | The columns looked disconnected from the frame. | `┬` / `┴` joints. |
| A rule before the columns opened carrying joints | Announced a column that did not exist. | Plain `├──┤` under the header. |
| Multiple blank rows between groups | Read as missing content. | Exactly one. |
| Selected row's indent leaked as plain spaces | The tint had a gap at the left edge. | The row's own background in every gap. |
| Strings rendered once at init | Kept the Ascii profile the process started with and lost colour forever. | Render at call time, always. |
| Left side shrinking the summary on a narrow window | The count jumped around while the user filtered. | Clip the left side first; the summary stays flush. |
| Wrapping a hint bar to a second line | The frame grew a row and pushed the footer. | Drop whole hints. |
| Empty groups drawn under a zero-match filter | The screen looked broken rather than empty. | One calm message and the way out. |

## 10. Migration map

The private copies these packages were extracted from, for the screens that are
still on them:

| Old (herdr-suite `internal/ui`) | New |
| --- | --- |
| `ui.Frame`, `ui.Layout`, `ui.Float`, `ui.Footer` | `frame.` |
| `ui.Section`, `ui.Guide`, `ui.Row`, `ui.SelectedRow`, `ui.Roles`, `ui.RowRoles` | `list.` |
| `ui.Chip`, `ui.Hint`, `ui.Hints` | `chip.` |
| `ui.Width`, `ui.Clip`, `ui.ClipLeft`, `ui.EllipsisPath`, `ui.Fit`, `ui.Join`, `ui.Center`, `ui.ShortPath`, `ui.RelativeTime`, `ui.Plain` | `text.` |
| `ui.Live`, `ui.Idle`, `ui.Mark`, `ui.SelectMark`, `ui.Cursor`, `ui.Ellipsis` | `text.` |
| `ui.Dot`, `ui.Label` | `theme.` |
| `ui.StateGlyph`, `ui.StateWord` | `list.` |
| `ui.Text` … `ui.ChipRed`, `ui.OnSelection` | `theme.` |
| `ui.Spinner`, `ui.SpinnerFrames` | `spinner.Render`, `spinner.Frames` |

| Old (herdr-utils `internal/which`) | New |
| --- | --- |
| `truncate`, `padPlain`, `padStyled`, `fitStyled` | `text.Clip`, `text.Fit` |
| `padLeft` | `text.PadLeft` |
| `joinSides` | `text.Join` / `frame.Footer` |
| `topBorder`, `bottomBorder`, `panelLine` | `frame.Frame.Top`, `.Bottom`, `.Row` |
| `plainRule` | `frame.Frame.Rule` |
| `columnRule`, `columnWidths`, `columnCount`, `assemble` | `frame.ColumnRule`, `frame.ColumnWidths`, `frame.ColumnCount`, `frame.Grid` |
| `chipPair`, `chipCell`, `hintsLine`, legend items | `chip.Pair`, `chip.Cell`/`chip.UnsetCell`, `chip.Hints`, `chip.Legend` |
| `styleSelectEdge` / `styleSelectChip` / `styleSelectText` / `styleSelectMuted` | `theme.SelectionEdge`, `theme.ChipOnSelection`, `theme.TextOnSelection`, `theme.MutedOnSelection` |
| `roleText` … `roleChipBG`, `styleText` … `styleSelectMuted` | `theme.ColorText` … `theme.ChipRed` |
