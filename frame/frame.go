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

// BlankRow is one body row with the column separators and no content. A gridded
// body's padding and spare rows are built with it, so the column rule runs
// unbroken from the opening joint to the closing one.
func BlankRow(f Frame, widths []int) string {
	if len(widths) == 0 {
		return f.Row("")
	}
	return Grid(f, nil, widths, 1)[0]
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

// BodyPadding is the blank row a body always keeps at its top and at its bottom,
// so its content never touches the rule above it or the footer below it. It is
// what makes two sections distinguishable at a glance.
const BodyPadding = 1

// Size is the range a frame is drawn within, whatever the pane offers. A zero
// bound is no bound, so an unset Size draws the frame at the pane's own size.
type Size struct {
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// Recommended is the size range a product popup should use, and the range its
// herdr manifest should be able to offer:
//
//   - 60 wide: the narrowest width that still shows the first page of the hint
//     bar, a chip column and a description without cutting either.
//   - 160 wide: past that the eye cannot follow one footer across the pane, and
//     longer facts stop helping.
//   - 16 tall: the borders, the header and its rule, the filter row, the column
//     rules, the two footer rows and a three-row body.
//   - 40 tall: a whole screen of groups; past that the footer drifts away from
//     the content it describes.
//
// A pane smaller than the range wins: a frame never exceeds the pane showing it.
var Recommended = Size{MinWidth: 60, MaxWidth: 160, MinHeight: 16, MaxHeight: 40}

// ClampWidth is the frame width for a pane width.
func (s Size) ClampWidth(pane int) int { return clampSize(pane, s.MinWidth, s.MaxWidth) }

// ClampHeight is the frame height for a pane height.
func (s Size) ClampHeight(pane int) int { return clampSize(pane, s.MinHeight, s.MaxHeight) }

// Clamp is both bounds at once.
func (s Size) Clamp(width, height int) (int, int) {
	return s.ClampWidth(width), s.ClampHeight(height)
}

// clampSize holds pane inside min..max and never past the pane itself: the
// minimum of too small a pane is the pane, because a frame that overflows its
// pane is worse than a cramped one.
func clampSize(pane, minimum, maximum int) int {
	size := pane
	if minimum > 0 && size < minimum {
		size = minimum
	}
	if maximum > 0 && size > maximum {
		size = maximum
	}
	if size > pane {
		size = pane
	}
	return max(0, size)
}

// Surface is one screen inside its bounds: the frame it is drawn in, the size
// range it stays within, and the rows pinned to its edges. Top sits directly under
// the top border, Bottom directly above the closing border, and the body takes
// what is between them, always opening and closing with one blank row.
//
// The frame is as tall as the clamped pane, never as tall as its content, so the
// footer is pinned to the bottom edge instead of the pane staying empty below it.
// Top, Bottom and the body are framed lines: the product renders them with the
// frame it already sized (ask Width and Height first, or hand in a Frame whose
// Width is Width(paneWidth)), so its own rules and columns survive.
//
// When the body is a column grid, set Columns to its widths: the padding and spare
// rows are drawn as grid rows too, so a column rule runs unbroken from the opening
// joint to the closing one instead of breaking on a plain row.
type Surface struct {
	Frame  Frame
	Size   Size
	Top    []string
	Bottom []string
	// Columns are the body's column widths when the body is a grid; nil means a
	// plain body.
	Columns []int
}

// Width is the frame width for a pane width.
func (s Surface) Width(paneWidth int) int { return s.Size.ClampWidth(paneWidth) }

// Height is the frame height for a pane height.
func (s Surface) Height(paneHeight int) int { return s.Size.ClampHeight(paneHeight) }

// BodyRows is how many content rows the body gets for a pane height: the clamped
// height minus the two borders, the Top and Bottom rows and the body's two
// padding rows. A product pages or scrolls its content to fit this count.
func (s Surface) BodyRows(paneHeight int) int {
	return max(0, s.Height(paneHeight)-2-len(s.Top)-len(s.Bottom)-2*BodyPadding)
}

// Render composes the frame at the clamped size: Top under the top border, the
// body, then Bottom against the closing border. The frame is always as tall as
// the clamped pane, so extra room lands inside the body below its content and
// never below the footer. The body's first and last rows are blank; body rows
// past the room are dropped, so a caller asks BodyRows and pages.
func (s Surface) Render(paneWidth, paneHeight int, body []string) string {
	f := s.Frame
	f.Width = s.Width(paneWidth)

	top, bottom := s.Top, s.Bottom
	interior := max(0, s.Height(paneHeight)-2)
	if len(top)+len(bottom) > interior {
		// A pane too short for both edges: the footer keeps the last rows, the
		// header keeps whatever is left.
		if len(bottom) > interior {
			bottom = bottom[len(bottom)-interior:]
		}
		top = top[:max(0, interior-len(bottom))]
	}

	rows := make([]string, 0, interior)
	rows = append(rows, top...)
	rows = append(rows, bodyBlock(f, s.Columns, body, interior-len(top)-len(bottom))...)
	rows = append(rows, bottom...)
	return Layout(f, rows...)
}

// Popup renders the surface and centres it in the pane, so a frame narrower or
// shorter than its pane sits in the middle of it instead of hugging its top.
func (s Surface) Popup(paneWidth, paneHeight int, body []string) string {
	return Float(s.Render(paneWidth, paneHeight, body), paneWidth, paneHeight)
}

// bodyBlock builds the body's rows: a blank row, the content, the room that is
// left, a blank row — every one of them framed, so a padding row reads as empty
// body rather than as a hole in the frame. In a gridded body the blank rows keep
// the column separators, so the column rule does not break around them. The spare
// room sits below the content, so content stays just under the header while the
// body still ends with its padding row.
func bodyBlock(f Frame, columns []int, content []string, room int) []string {
	block := make([]string, max(0, room))
	blank := BlankRow(f, columns)
	for index := range block {
		block[index] = blank
	}
	if room >= 2*BodyPadding {
		visible := content
		if len(visible) > room-2*BodyPadding {
			visible = visible[:room-2*BodyPadding]
		}
		copy(block[BodyPadding:], visible)
	}
	return block
}
