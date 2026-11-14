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
)

// Roles is the set of roles one row paints with. A row is either on the panel
// surface or selected, and the selection is a tint: every element keeps the
// meaning its colour carries, with the muted and structural tones lifted so they
// still read on the darker background.
type Roles struct {
	// Selected says whether the row is the highlighted one.
	Selected bool
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
	// Gold, Green, Teal and Red keep their meaning on both surfaces.
	Gold  lipgloss.Style
	Green lipgloss.Style
	Teal  lipgloss.Style
	Red   lipgloss.Style
}

var (
	panelRoles = Roles{
		Name:  theme.Text,
		Muted: theme.Muted,
		Guide: theme.Quiet,
		Space: lipgloss.NewStyle(),
		Chip:  theme.ChipText,
		Gold:  theme.Gold,
		Green: theme.Green,
		Teal:  theme.Teal,
		Red:   theme.Red,
	}
	selectionRoles = Roles{
		Selected: true,
		Name:     theme.TextOnSelection,
		Muted:    theme.MutedOnSelection,
		Guide:    theme.GuideOnSelection,
		Space:    theme.SelectionBG,
		Chip:     theme.ChipOnSelection,
		Gold:     theme.OnSelection(theme.Gold),
		Green:    theme.OnSelection(theme.Green),
		Teal:     theme.OnSelection(theme.Teal),
		Red:      theme.OnSelection(theme.Red),
	}
)

// RowRoles returns the roles a row paints with: the panel surface, or the
// selection variants when the row is the highlighted one.
func RowRoles(selected bool) Roles {
	if selected {
		return selectionRoles
	}
	return panelRoles
}

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
// SelectedRow to add the accent edge.
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

// ChipRow renders one listed entry whose first column is a chip cell: the tree
// guide, the cell (see chip.Cell), then the description. On the selection tint
// the row's first cell becomes the accent edge and the tint runs to width, so
// the description keeps the x it has on the panel.
func ChipRow(roles Roles, last bool, cell, description string, width int) string {
	if width <= 0 {
		return ""
	}
	row := Row(roles, last, cell+description)
	if roles.Selected {
		return SelectedRow(row, width)
	}
	return roles.Pad(row, width)
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
