// Package text measures and fits text in terminal cells: widths that ignore
// escapes, truncation that ends an honest cut with an ellipsis, columns that keep
// their x, and the short forms a list column shows (a path, an age).
//
// It depends on no other tui-kit package, so anything may import it, and it
// knows nothing about colour: a caller decides which role a fitted string keeps.
package text

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// Ellipsis is how a clipped string ends: one character, so a cut reads as a cut
// and never as content.
const Ellipsis = "…"

// joinGap is the least air JoinDropRight leaves between its two sides: the right
// side never touches the text it explains.
const joinGap = 1

// The glyphs of the kit's typography. Each one is a single cell, so a glyph
// never shifts a column.
const (
	// Live is the glyph a live thing carries: a running session, a focused pane.
	Live = "●"
	// Idle is the glyph a quiescent thing carries.
	Idle = "○"
	// Mark is the glyph for state that outlives the selection: the default
	// session, the focused workspace. It is gold and quieter than the selection.
	Mark = "◆"
	// SelectMark caps the selected row. The tint carries the highlight; this one
	// cell keeps the row's left edge readable.
	SelectMark = "▌"
	// Cursor is the block cursor of the live filter query.
	Cursor = "▏"
	// Unset is the glyph an entry with no key of its own shows instead of an
	// empty chip.
	Unset = "–"
)

// Width is the visible width of s: ANSI escapes and styling cost no cells.
func Width(s string) int { return ansi.StringWidth(s) }

// Clip shortens s to width cells, ending with an ellipsis when it had to cut.
func Clip(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if Width(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, Ellipsis)
}

// ClipLeft shortens s from the start, keeping its tail: a path is identified by
// where it ends, not by where it starts.
func ClipLeft(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if Width(s) <= width {
		return s
	}
	drop := Width(s) - (width - Width(Ellipsis))
	return ansi.TruncateLeft(s, drop, Ellipsis)
}

// EllipsisPath shortens a path to width by keeping its trailing segments, the
// way a shell prompt shows the part that identifies it: "…/herdr/work.sock"
// instead of a cut in the middle of a directory name.
func EllipsisPath(path string, width int) string {
	if width <= 0 {
		return ""
	}
	if Width(path) <= width {
		return path
	}
	if width <= Width(Ellipsis)+1 {
		return ClipLeft(path, width)
	}

	segments := strings.Split(path, "/")
	tail := ""
	for i := len(segments) - 1; i >= 0; i-- {
		segment := segments[i]
		if segment == "" {
			continue
		}
		candidate := segment
		if tail != "" {
			candidate = segment + "/" + tail
		}
		if Width(Ellipsis+"/"+candidate) > width {
			break
		}
		tail = candidate
	}
	if tail == "" {
		return ClipLeft(path, width)
	}
	return Ellipsis + "/" + tail
}

// Fit clips s to width cells and pads it with plain spaces, so the caller can
// place the next column at a fixed x.
func Fit(s string, width int) string {
	clipped := Clip(s, width)
	if pad := width - Width(clipped); pad > 0 {
		return clipped + strings.Repeat(" ", pad)
	}
	return clipped
}

// PadLeft fits s into width cells against its right edge, so a column of numbers
// or a source name ends on one cell.
func PadLeft(s string, width int) string {
	clipped := Clip(s, width)
	if pad := width - Width(clipped); pad > 0 {
		return strings.Repeat(" ", pad) + clipped
	}
	return clipped
}

// Join places left at the start and right at the end of width cells. When both
// cannot fit, the left side is clipped first: the right side carries the short
// facts a narrow window must keep.
//
// Join lays down no gutter of its own: at an exact fit the two sides touch. That
// is the line every caller has rendered since the kit was extracted, so its bytes
// stay as they are in v0.4.0; a caller that wants one cell of air either puts it
// inside right (frame.Footer spends two) or uses JoinDropRight, which enforces it.
func Join(left, right string, width int) string {
	if gap := width - Width(left) - Width(right); gap >= 0 {
		return left + strings.Repeat(" ", gap) + right
	}
	left = Clip(left, max(0, width-Width(right)-1))
	gap := width - Width(left) - Width(right)
	if gap < 0 {
		return Clip(right, width)
	}
	return left + strings.Repeat(" ", gap) + right
}

// JoinDropRight places left at the start and right at the end of width cells,
// like Join, keeps at least one plain cell between the two sides, and gives up
// the right side first: when both do not fit with that gutter, right is dropped
// whole and left stays readable. It is the join for a line whose right side is a
// legend that may give way before the title it explains. Only when the left side
// alone does not fit is it clipped, and that cut ends with an ellipsis.
//
// The gutter is the kit's, so a product stops buying it by padding the right side
// with a leading space: the padded call drew exactly these bytes.
func JoinDropRight(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	if right == "" {
		return Clip(left, width)
	}
	if gap := width - Width(left) - Width(right); gap >= joinGap {
		return left + strings.Repeat(" ", gap) + right
	}
	return Clip(left, width)
}

// Center places s in the middle of width cells.
func Center(s string, width int) string {
	s = Clip(s, width)
	if pad := width - Width(s); pad > 0 {
		left := pad / 2
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
	}
	return s
}

// ShortPath replaces the user's home prefix with ~, the way the rest of the
// terminal shortens a path a user recognizes. It is capability-based: no path is
// assumed, and a path outside the home directory is returned unchanged.
func ShortPath(path string) string {
	if path == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" || home == string(filepath.Separator) {
		return path
	}
	if path == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(path, home+string(filepath.Separator)); ok {
		return "~/" + rest
	}
	return path
}

// RelativeTime renders how long ago t was, in the compact form a list column
// holds: just now, 5m ago, 2h ago, 3d ago.
func RelativeTime(now, t time.Time) string {
	if t.IsZero() {
		return ""
	}
	elapsed := now.Sub(t)
	switch {
	case elapsed < time.Minute:
		return "just now"
	case elapsed < time.Hour:
		return ago(int(elapsed.Minutes()), "m")
	case elapsed < 24*time.Hour:
		return ago(int(elapsed.Hours()), "h")
	case elapsed < 30*24*time.Hour:
		return ago(int(elapsed.Hours()/24), "d")
	case elapsed < 365*24*time.Hour:
		return ago(int(elapsed.Hours()/(24*30)), "mo")
	default:
		return ago(int(elapsed.Hours()/(24*365)), "y")
	}
}

func ago(value int, unit string) string {
	if value < 1 {
		value = 1
	}
	return itoa(value) + unit + " ago"
}

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

// Plain strips styling from s: a caller that paints a row it did not compose
// needs the row's text with none of its colours.
func Plain(s string) string { return ansi.Strip(s) }
