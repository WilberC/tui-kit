// Package chip is the token surface of the interface: one slab per key, mode or
// short state word, the columns those slabs align in, and the hint bar built
// from them.
//
// A chip is a shape, not a bracketed word: the background carries one cell of
// padding on each side of the label, and a label always follows its chip after
// exactly one plain space.
package chip

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

// Gutter is what separates two chip pairs in a legend or a hint bar: two plain
// spaces, so no separator ever touches a chip's background.
const Gutter = "  "

// Chip renders one token on the chip surface: the slab carries one cell of
// padding on each side, so a chip is a shape rather than a bracketed word.
func Chip(role lipgloss.Style, label string) string {
	return role.Render(" " + label + " ")
}

// Pair is one chip and the dim label that explains it: the slab, then exactly one
// plain space, then the label in the muted role.
func Pair(slab lipgloss.Style, key, label string) string {
	return Chip(slab, key) + " " + theme.Muted.Render(label)
}

// Legend joins already rendered chip pairs with two plain spaces. It never
// inserts a glyph between them, so the gutter stays colourless.
func Legend(items ...string) string {
	return join(items)
}

// Width is the cell width of the widest label: the chip column a group of rows
// shares, so every description starts on the same cell.
func Width(labels ...string) int {
	width := 0
	for _, label := range labels {
		width = max(width, text.Width(label))
	}
	return width
}

// Cell renders one chip cell of a fixed column: the label is padded to width
// cells inside the slab, and one plain space follows the slab, so the
// description that comes next keeps its x on every row.
func Cell(slab lipgloss.Style, label string, width int) string {
	return slab.Render(" "+text.Fit(label, width)+" ") + " "
}

// UnsetCell renders the cell of an entry that declares no key of its own: no
// slab at all, a dim glyph padded into the same column, so a group of such
// entries never reads as a wall of empty boxes.
func UnsetCell(role lipgloss.Style, label string, width int) string {
	return role.Render(text.Fit(" "+label, width+3))
}

// Hint is one key/label pair of a hint bar.
type Hint struct {
	Key   string
	Label string
}

// Hints lays out as many hints as fit in width. A hint that does not fit is
// dropped whole, from the end, so a narrow window loses detail instead of
// wrapping the footer.
func Hints(width int, hints ...Hint) string {
	line := ""
	for _, hint := range hints {
		next := Chip(theme.ChipText, hint.Key) + " " + theme.Muted.Render(hint.Label)
		if line != "" {
			next = Gutter + next
		}
		if text.Width(line)+text.Width(next) > width {
			break
		}
		line += next
	}
	if line == "" && len(hints) > 0 {
		return text.Clip(Chip(theme.ChipText, hints[0].Key)+" "+theme.Muted.Render(hints[0].Label), width)
	}
	return line
}

func join(items []string) string {
	var out string
	for _, item := range items {
		if item == "" {
			continue
		}
		if out != "" {
			out += Gutter
		}
		out += item
	}
	return out
}
