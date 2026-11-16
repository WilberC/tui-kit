package text

import (
	"strings"

	"github.com/muesli/reflow/wordwrap"
)

// Wrap folds s on words into width-wide lines, so a long row continues to
// the next line instead of hiding behind an ellipsis: Clip loses content,
// Wrap keeps it. ANSI escapes ride through reflow unchanged, so every colour
// survives the fold; each returned line is already fitted to width (padded),
// so the caller can place the next column without measuring again. A line
// that fits returns as a single fitted line; width 0 folds nothing.
func Wrap(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	if Width(s) <= width {
		return []string{Fit(s, width)}
	}
	folded := wordwrap.String(s, width)
	parts := strings.Split(folded, "\n")
	for i, p := range parts {
		parts[i] = Fit(strings.TrimRight(p, " "), width)
	}
	return parts
}

// WrapRows folds every row of a list into width-wide lines.
func WrapRows(rows []string, width int) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, Wrap(r, width)...)
	}
	return out
}
