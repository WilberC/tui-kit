// Package frame is the surface every popup renders inside: a titled border, the
// rules that join its sides, the columns those rules open, and the footer that
// closes the detail. It clips every line to its own width, so a view can never
// wrap, whatever the window size.
package frame

import (
	"strings"

	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

// The frame's glyphs. The joints are what let a rule meet the sides and the
// columns: a rule painted without them floats inside the box.
const (
	// Vertical is the frame's side, and the rule between two columns.
	Vertical = "│"
	// Horizontal is the frame's edge and the body of a rule.
	Horizontal = "─"
	// TopLeft, TopRight, BottomLeft and BottomRight are the four corners.
	TopLeft     = "╭"
	TopRight    = "╮"
	BottomLeft  = "╰"
	BottomRight = "╯"
	// LeftTee and RightTee join a rule to the frame's sides.
	LeftTee  = "├"
	RightTee = "┤"
	// TopTee grows a column rule out of the opening rule.
	TopTee = "┬"
	// BottomTee carries a column rule into the closing rule.
	BottomTee = "┴"
)

// framePad is the cell of breathing room between a border and its content.
const framePad = 1

// Frame is the bordered surface a TUI renders inside: the title sits in the top
// border, every interior line is padded by one cell on each side, and the sides
// are joined by rules.
type Frame struct {
	// Title is the surface's name, drawn in the top border. It may be styled.
	Title string
	// Width is the whole surface, borders included.
	Width int
}

// ContentWidth is the width a row's content may use: the frame minus its two
// borders and the padding on each side.
func (f Frame) ContentWidth() int { return max(0, f.Width-2-2*framePad) }

// Top renders the top border with the title embedded in it.
func (f Frame) Top() string {
	inner := max(0, f.Width-2)
	if text.Width(f.Title) == 0 {
		return theme.Border.Render(TopLeft + strings.Repeat(Horizontal, inner) + TopRight)
	}
	fill := max(0, inner-3-text.Width(f.Title))
	return theme.Border.Render(TopLeft+Horizontal+" ") + f.Title + theme.Border.Render(" "+strings.Repeat(Horizontal, fill)+TopRight)
}

// Bottom renders the closing border.
func (f Frame) Bottom() string {
	return theme.Border.Render(BottomLeft + strings.Repeat(Horizontal, max(0, f.Width-2)) + BottomRight)
}

// Row renders one interior line: the side borders, one cell of padding, the
// content fitted to the width, one cell of padding.
func (f Frame) Row(content string) string {
	return theme.Border.Render(Vertical) + " " + text.Fit(content, f.ContentWidth()) + " " + theme.Border.Render(Vertical)
}

// Rule renders a horizontal rule joined to the frame's sides, so sections
// connect instead of floating inside the box. The joints carry the border role
// and the body the guide role, so the rule reads as structure, not as a border.
func (f Frame) Rule() string {
	return theme.Border.Render(LeftTee) + theme.Quiet.Render(strings.Repeat(Horizontal, max(0, f.Width-2))) + theme.Border.Render(RightTee)
}

// Render wraps plain rows into the frame: every row is fitted to the content
// width, so the caller may hand in unfinished text.
func (f Frame) Render(rows []string) string {
	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, f.Top())
	for _, row := range rows {
		lines = append(lines, f.Row(row))
	}
	lines = append(lines, f.Bottom())
	return strings.Join(lines, "\n")
}

// Placeholder renders the calm, centered message a frame shows when it has no
// rows to list: one line naming the state, one dim line naming what to do next.
func (f Frame) Placeholder(message, hint string) []string {
	rows := []string{""}
	rows = append(rows, text.Center(theme.Strong.Render(message), f.ContentWidth()))
	if hint != "" {
		rows = append(rows, text.Center(theme.Muted.Render(hint), f.ContentWidth()))
	}
	return append(rows, "")
}

// Layout lays a frame out: the top border, the rows the caller composed, and the
// closing border. The rows are used as they are, so a caller that composed its
// own rules keeps them.
func Layout(f Frame, rows ...string) string {
	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, f.Top())
	lines = append(lines, rows...)
	lines = append(lines, f.Bottom())
	return strings.Join(lines, "\n")
}

// Float places a rendered block in a window the way a popup opens: centered on
// both axes, with the block's own edges as the only boundary. A block wider than
// the window is left at the left edge rather than clipped, so a caller that
// measured its frame keeps the whole frame.
func Float(block string, width, height int) string {
	lines := strings.Split(block, "\n")
	blockWidth := 0
	for _, line := range lines {
		blockWidth = max(blockWidth, text.Width(line))
	}
	left := max(0, (width-blockWidth)/2)
	top := max(0, (height-len(lines))/2)

	padding := strings.Repeat(" ", left)
	var out strings.Builder
	for i := 0; i < top; i++ {
		out.WriteString("\n")
	}
	for i, line := range lines {
		if i > 0 {
			out.WriteString("\n")
		}
		if line != "" {
			out.WriteString(padding)
			out.WriteString(line)
		}
	}
	return out.String()
}

// Footer joins the detail of the selected item with its short summary: the
// summary sits flush with the inner edge, a plain gutter separates the two, and
// the left side is clipped first so the summary never moves.
func Footer(left, right string, width int) string {
	if right == "" {
		return text.Clip(left, width)
	}
	if text.Width(left)+2+text.Width(right) > width {
		left = text.Clip(left, max(0, width-text.Width(right)-2))
	}
	if gap := width - text.Width(left) - text.Width(right); gap > 0 {
		return left + strings.Repeat(" ", gap) + right
	}
	return right
}

// ColumnCount is how many columns a width earns: three when the terminal is
// wide, two in the middle, one otherwise, and never more than want.
func ColumnCount(innerWidth, want int) int {
	count := 1
	switch {
	case innerWidth >= 148:
		count = 3
	case innerWidth >= 98:
		count = 2
	}
	if want > 0 {
		count = min(count, want)
	}
	return count
}

// ColumnWidths splits the inner width into count columns with one cell of rule
// between them. Each column spends one cell of padding on each side of its
// content, so ColumnRule over the result spans the frame's whole width.
func ColumnWidths(innerWidth, count int) []int {
	if count < 1 {
		count = 1
	}
	total := innerWidth - 2*count - (count - 1)
	base, rest := total/count, total%count
	widths := make([]int, count)
	for index := range widths {
		widths[index] = max(0, base)
		if index < rest && base >= 0 {
			widths[index]++
		}
	}
	return widths
}

// ColumnRule opens or closes the columns: a joint above every column rule, so
// the vertical rules grow out of the frame instead of floating in it. Pass TopTee
// to open the columns and BottomTee to close them.
func ColumnRule(widths []int, junction string) string {
	var rule strings.Builder
	for index, width := range widths {
		rule.WriteString(strings.Repeat(Horizontal, max(0, width+2)))
		if index < len(widths)-1 {
			rule.WriteString(junction)
		}
	}
	return theme.Border.Render(LeftTee) + theme.Border.Render(rule.String()) + theme.Border.Render(RightTee)
}

// Grid joins column blocks row by row: every cell is fitted to its column, the
// rows are separated by the frame's vertical rule, and the result is exactly
// height lines tall. blocks[i] is the body of column i and may be shorter than
// height, in which case the column is padded with its own background.
func Grid(f Frame, blocks [][]string, widths []int, height int) []string {
	lines := height
	if lines <= 0 {
		for _, block := range blocks {
			lines = max(lines, len(block))
		}
	}
	out := make([]string, 0, lines)
	for line := 0; line < lines; line++ {
		var row strings.Builder
		row.WriteString(theme.Border.Render(Vertical))
		for index, width := range widths {
			content := ""
			if index < len(blocks) && line < len(blocks[index]) {
				content = blocks[index][line]
			}
			row.WriteString(" ")
			row.WriteString(text.Fit(content, width))
			row.WriteString(" ")
			if index < len(widths)-1 {
				row.WriteString(theme.Border.Render(Vertical))
			}
		}
		row.WriteString(theme.Border.Render(Vertical))
		out = append(out, row.String())
	}
	return out
}
