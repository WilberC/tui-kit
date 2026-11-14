// Package list is the listed surface: rows that keep their columns, the tree
// guide that places them, the section header that opens a group, and the tinted
// selection that lifts one row without repainting it.
package list

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
	"github.com/WilberC/tui-kit/zone"
)

// Roles is the set of roles one row paints with. A row is either on the panel
// surface, hovered, or selected, and both tints are tints: every element keeps
// the meaning its colour carries, with the muted and structural tones lifted so
// they still read on the darker background. The selection wins over the hover.
type Roles struct {
	// Selected says whether the row is the highlighted one.
	Selected bool
	// Hovered says whether the pointer is over the row and it is not selected.
	Hovered bool
	// Name is the row's primary text: the session, the layout, the workspace.
	Name lipgloss.Style
	// Muted is every secondary fact: workspace, counts, age, path.
	Muted lipgloss.Style
	// Guide is the tree structure that places the row.
	Guide lipgloss.Style
	// Space fills the cells of a row that carry no text, so a selected row's
	// tint reaches its edges.
	Space lipgloss.Style
	// Chip is the token surface a row may carry.
	Chip lipgloss.Style
	// ChipAccent is the token surface of a chip that carries the product's accent
	// instead of its text role — a keybinding guide's binding. It follows Chip's
	// two surfaces: its own slab on the panel and under the pointer, and the
	// brighter slab on the selection, so a product never overrides Roles.Chip by
	// hand to paint one column differently.
	ChipAccent lipgloss.Style
	// Gold, Green, Teal and Red keep their meaning on both surfaces.
	Gold  lipgloss.Style
	Green lipgloss.Style
	Teal  lipgloss.Style
	Red   lipgloss.Style
}

var (
	panelRoles = Roles{
		Name:       theme.Text,
		Muted:      theme.Muted,
		Guide:      theme.Quiet,
		Space:      lipgloss.NewStyle(),
		Chip:       theme.ChipText,
		ChipAccent: theme.ChipAccent,
		Gold:       theme.Gold,
		Green:      theme.Green,
		Teal:       theme.Teal,
		Red:        theme.Red,
	}
	hoverRoles = Roles{
		Hovered: true,
		Name:    theme.TextOnHover,
		Muted:   theme.MutedOnHover,
		Guide:   theme.GuideOnHover,
		Space:   theme.HoverBG,
		// A chip carries its own surface, so it steps onto the hover tint rather
		// than brightening like the selection's: a hover is a preview.
		Chip:       theme.ChipText,
		ChipAccent: theme.ChipAccent,
		Gold:       theme.OnHover(theme.Gold),
		Green:      theme.OnHover(theme.Green),
		Teal:       theme.OnHover(theme.Teal),
		Red:        theme.OnHover(theme.Red),
	}
	selectionRoles = Roles{
		Selected:   true,
		Name:       theme.TextOnSelection,
		Muted:      theme.MutedOnSelection,
		Guide:      theme.GuideOnSelection,
		Space:      theme.SelectionBG,
		Chip:       theme.ChipOnSelection,
		ChipAccent: theme.ChipAccentOnSelection,
		Gold:       theme.OnSelection(theme.Gold),
		Green:      theme.OnSelection(theme.Green),
		Teal:       theme.OnSelection(theme.Teal),
		Red:        theme.OnSelection(theme.Red),
	}
)

// State is how a row is marked: the highlighted row, the row under the pointer,
// or neither. Selecting wins, so a row that is both stays the selection.
type State struct {
	Selected bool
	Hovered  bool
}

// RowRolesFor returns the roles a row paints with: the panel surface, the faint
// hover tint, or the selection.
func RowRolesFor(state State) Roles {
	switch {
	case state.Selected:
		return selectionRoles
	case state.Hovered:
		return hoverRoles
	default:
		return panelRoles
	}
}

// RowRoles returns the roles a row paints with: the panel surface, or the
// selection variants when the row is the highlighted one. It is
// RowRolesFor(State{Selected: selected}), kept so the callers that only know the
// selection keep compiling.
func RowRoles(selected bool) Roles { return RowRolesFor(State{Selected: selected}) }

// Pad fits s into width cells, filling the rest with the row's own background,
// so a selected row's tint has no gaps between its columns.
func (r Roles) Pad(s string, width int) string {
	clipped := text.Clip(s, width)
	return clipped + r.Gap(width-text.Width(clipped))
}

// Gap is n cells of the row's own background: the gutter between two columns.
func (r Roles) Gap(n int) string {
	if n <= 0 {
		return ""
	}
	return r.Space.Render(strings.Repeat(" ", n))
}

// AlignRight places right at the end of width cells, filling the space between
// with the row's own background.
func (r Roles) AlignRight(left, right string, width int) string {
	gap := width - text.Width(left) - text.Width(right)
	if gap < 0 {
		left = text.Clip(left, max(0, width-text.Width(right)-1))
		gap = width - text.Width(left) - text.Width(right)
	}
	if gap < 0 {
		return text.Clip(right, width)
	}
	return left + r.Gap(gap) + right
}

// AgentState renders the agent state a workspace reports on these roles: live
// work is green, a worked-up agent is gold, an unknown state stays muted, and a
// failed agent is the one red thing in a list.
func (r Roles) AgentState(state string) string {
	switch lower(state) {
	case "":
		return ""
	case "working", "running", "busy":
		return r.Green.Render(state)
	case "idle", "done", "stopped":
		return r.Muted.Render(state)
	case "waiting", "blocked", "paused", "needs-input":
		return r.Gold.Render(state)
	case "error", "failed", "dead":
		return r.Red.Render(state)
	default:
		return r.Muted.Render(state)
	}
}

// Mark is the gold ◆ a row carries for state that outlives the selection: the
// default item, the focused workspace. It follows the row's facts after two cells
// of the row's own background, so a marked row never moves its chip column or its
// description, and a mark is never the first thing on a row: it is not a second
// cursor. On a tinted row it takes that row's roles, so the gold stays gold.
func Mark(roles Roles) string { return roles.Gap(2) + roles.Gold.Render(text.Mark) }

// StateGlyph is the glyph a running state renders with: green when live, muted
// when quiescent.
func StateGlyph(running bool) string {
	if running {
		return theme.Green.Render(text.Live)
	}
	return theme.Muted.Render(text.Idle)
}

// StateWord names a running state on the palette: live is green, quiescent is
// muted.
func StateWord(running bool) string {
	if running {
		return theme.Green.Render("running")
	}
	return theme.Muted.Render("stopped")
}

// Guide renders the tree guide of a list entry: two cells, joined to the last
// entry of its group so the list closes instead of trailing off.
func Guide(role lipgloss.Style, last bool) string {
	if last {
		return "  " + role.Render("└─")
	}
	return "  " + role.Render("├─")
}

// Row renders one listed entry: the indent and tree guide its position asks for,
// then the body the caller composed. On a selected row the indent is tinted, so
// the background reaches the row's edge, and the caller wraps the result in
// SelectedRow to add the accent edge. A row that hangs under another one is drawn
// through Entry.Parent, which prefixes this guide with the parent's continuation
// column.
func Row(roles Roles, last bool, body string) string {
	guide := "├─"
	if last {
		guide = "└─"
	}
	return roles.Space.Render("  ") + roles.Guide.Render(guide) + roles.Space.Render(" ") + body
}

// SelectedRow paints the selected row: the accent edge in the first cell, then
// the row itself, which keeps every element's colour, and the tint carried to
// the row's width so the highlight has no ragged end.
func SelectedRow(row string, width int) string {
	if width <= 0 {
		return ""
	}
	// The edge takes the row's first cell, so every column keeps its x: the
	// indent the row already spent becomes the mark.
	body := ansi.TruncateLeft(row, 1, "")
	if pad := width - 1 - text.Width(body); pad > 0 {
		body += theme.SelectionBG.Render(strings.Repeat(" ", pad))
	} else if text.Width(body) > width-1 {
		body = text.Clip(body, width-1)
	}
	return theme.SelectionEdge.Render(text.SelectMark) + body
}

// rowIndent is how many cells the tree guide and its indents take before a row's
// body starts: two for the indent, two for the guide, one after it.
const rowIndent = 5

// nestIndent is where a nested row's own guide starts: its parent's continuation
// column, which is the parent's indent and guide, then the row's own indent and
// guide. A child therefore starts five cells past its parent.
const nestIndent = 2 * rowIndent

// guideOpen is the tree guide a nested row keeps open under its parent: the column
// continues the parent's guide instead of leaving a gap in it.
const guideOpen = "│"

// TrailingGap is the least room a row keeps between its description and its
// trailing column: two cells of the row's own background, so the last two columns
// of a table never touch and the selection or hover tint runs through the gutter.
const TrailingGap = 2

// Parent is the row a nested entry hangs under: the parent's place in its group.
// A nested entry draws its own guide under the parent's continuation column, which
// keeps the parent's guide open when the parent has a sibling below it and stays
// blank when the parent closes the group — the shape a workspace list nested under
// its session has.
type Parent struct {
	// Last says whether the parent closes its group. A last parent leaves the
	// continuation column blank, so the child's own guide closes the branch.
	Last bool
}

// Entry is one listed row: the roles it paints with, its place in its group, its
// columns, and where it goes in a zone map. Rendering it records the row and its
// chip slab as zones, which is what lets a product act on a click without
// measuring the layout itself.
//
// Entry covers both kinds of list row. With a Cell it is ChipRow plus the two zone
// ids, so the rows that need no zone keep calling ChipRow. With an empty Cell it is
// the row helper for a list that has no chip column: the entry draws exactly what
// Row draws, still records the whole row under ID, and records no chip zone — a
// session row lists itself without measuring a zone beside the kit's. With a Parent
// it is a nested row: its own guide hangs under the parent's continuation column,
// and it records its own zone like any other row.
type Entry struct {
	Roles Roles
	Last  bool
	// Parent is the row this entry hangs under, or nil for a top-level row. Set it
	// and the entry draws its own guide under the parent's continuation column, so
	// a nested row keeps the tree closed without the product composing the indent
	// by hand: the entry records its own zone at the same x and width the whole row
	// covers, continuation column included.
	Parent *Parent
	// Cell is the row's first column: a chip slab (see chip.Cell), or empty for a
	// row with no chip column.
	Cell string
	// Description is everything after the cell: the row's facts, in the roles the
	// caller chose.
	Description string
	// Trailing is the row's last column, right-aligned to the row's edge: the
	// source of a table, a count, a mode. Entry keeps TrailingGap cells of the
	// row's own background between it and the description and clips the
	// description, so the column never drifts and the tint never breaks. Empty
	// means the row has no trailing column.
	Trailing string
	// TrailingWidth is the width of the trailing column when the caller reserves
	// one. Zero keeps the column as wide as its text and right-aligns it against
	// the row's edge. A positive width makes the column exactly that wide with its
	// right edge on the row's edge and left-aligns the text inside it, so every
	// row's column starts on the same cell — the cell a table's header names — and
	// the cells the text does not fill are the row's own background, so the
	// selection or hover tint runs to the edge.
	TrailingWidth int
	Width         int
	// ID names the whole row — indent and columns, so a click anywhere on it
	// selects it — in Zones. An empty ID records no row zone.
	ID string
	// ChipID names the chip slab of the cell: the cell minus the plain space that
	// follows it. Leave it empty for an entry with no key of its own
	// (chip.UnsetCell draws no slab), so no click lands on a token that is not
	// there.
	ChipID string
	// Zones is the map the entry records itself in, in that map's coordinates. A
	// nil map records nothing.
	Zones *zone.Map
	// X and Y are where the row's first cell goes in the map's coordinates.
	X int
	Y int
}

// Render draws the entry and records its zones.
func (e Entry) Render() string {
	if e.Width <= 0 {
		return ""
	}
	row := e.continuation() + Row(e.Roles, e.Last, e.body())
	if e.Roles.Selected {
		row = SelectedRow(row, e.Width)
	} else {
		row = e.Roles.Pad(row, e.Width)
	}

	if e.ID != "" {
		e.Zones.Add(e.ID, e.X, e.Y, e.Width, 1)
	}
	if e.ChipID != "" && text.Width(e.Cell) > 1 {
		e.Zones.Add(e.ChipID, e.X+e.indent(), e.Y, text.Width(e.Cell)-1, 1)
	}
	return row
}

// continuation is the parent's guide column of a nested entry: the cells its parent
// spent before its own body, drawn as the parent's own background, with the
// parent's guide kept open — or blank when the parent closes its group. A top-level
// entry has no continuation.
func (e Entry) continuation() string {
	if e.Parent == nil {
		return ""
	}
	glyph := e.Roles.Space.Render(" ")
	if !e.Parent.Last {
		glyph = e.Roles.Guide.Render(guideOpen)
	}
	return e.Roles.Space.Render("  ") + glyph + e.Roles.Space.Render("  ")
}

// indent is how many cells the entry spends before its body: its own tree guide,
// or the parent's continuation column and its own guide.
func (e Entry) indent() int {
	if e.Parent != nil {
		return nestIndent
	}
	return rowIndent
}

// body composes the cells the tree guide places: the chip cell, the description,
// and the trailing column flush with the row's right edge. The gutter between the
// description and the trailing column is the row's own background, so a tint runs
// through it, and the description gives way first — the trailing column is the
// short, stable fact a table keeps.
func (e Entry) body() string {
	content := e.Cell + e.Description
	if e.Trailing == "" {
		return content
	}
	available := e.Width - e.indent()
	if available <= 0 {
		// The row is narrower than its own indent: there is no column to align, so
		// the description keeps what the frame can show.
		return content
	}
	column := e.trailingColumnWidth()
	if room := available - column - TrailingGap; room >= 0 {
		content = text.Clip(content, room)
		// The spare room and the gutter are one run of the row's own background, so
		// the row carries a single tinted span between its columns — the bytes a
		// hand-aligned column writes.
		return content + e.Roles.Gap(room-text.Width(content)+TrailingGap) + e.trailingColumn(column)
	}
	// The row is too narrow for both columns: the trailing column keeps the edge
	// and the description gives way, the way a line's right side stays flush.
	return e.Roles.Pad(e.trailingColumn(available), available)
}

// trailingColumnWidth is the room the trailing column takes: its own text, or the
// width the caller reserved.
func (e Entry) trailingColumnWidth() int {
	if e.TrailingWidth > 0 {
		return e.TrailingWidth
	}
	return text.Width(e.Trailing)
}

// trailingColumn draws the trailing column at width cells: the text left-aligned
// inside the caller's reserved width, and the rest of the column the row's own
// background, so the column ends on the row's edge and the tint reaches it.
func (e Entry) trailingColumn(width int) string {
	if e.TrailingWidth <= 0 {
		return e.Trailing
	}
	content := text.Clip(e.Trailing, width)
	return content + e.Roles.Gap(width-text.Width(content))
}

// ChipRow renders one listed entry whose first column is a chip cell: the tree
// guide, the cell (see chip.Cell), then the description. On the selection tint
// the row's first cell becomes the accent edge and the tint runs to width, so
// the description keeps the x it has on the panel.
func ChipRow(roles Roles, last bool, cell, description string, width int) string {
	return Entry{Roles: roles, Last: last, Cell: cell, Description: description, Width: width}.Render()
}

// Section renders a group header: the filled gold glyph that opens a group, its
// name in small caps and how many entries it holds.
func Section(label string, count int) string {
	return theme.Gold.Render(text.Live) + " " + theme.GoldBold.Render(strings.ToUpper(label)) + "  " + theme.Muted.Render(itoa(count))
}

// lower normalises a reported state word for the role switch.
func lower(word string) string { return strings.ToLower(word) }

func itoa(value int) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var out [20]byte
	i := len(out)
	for value > 0 {
		i--
		out[i] = digits[value%10]
		value /= 10
	}
	return string(out[i:])
}
