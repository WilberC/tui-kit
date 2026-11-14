package main

import (
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/text"
)

// TestTheDemoFillsItsWindowExactly pins the surface contract the screenshots
// rely on: every line occupies the width it was asked for, and the popup is
// exactly as tall as it was asked to be — at the three captured sizes and at
// every width between the narrowest terminal and the widest one.
func TestTheDemoFillsItsWindowExactly(t *testing.T) {
	windows := []struct{ width, height int }{
		{180, 50},
		{140, 40},
		{80, 30},
	}
	for width := 40; width <= 200; width++ {
		windows = append(windows, struct{ width, height int }{width, 30})
	}
	for _, window := range windows {
		view := render(window.width, window.height)
		lines := strings.Split(view, "\n")

		if len(lines) != window.height {
			t.Fatalf("%dx%d: the popup is %d rows tall", window.width, window.height, len(lines))
		}
		for index, line := range lines {
			if width := text.Width(line); width != window.width {
				t.Fatalf("%dx%d: line %d is %d cells wide:\n%s", window.width, window.height, index, width, text.Plain(line))
			}
		}
	}
}

// TestTheDemoCarriesTheApprovedDetails pins what the screenshots are meant to
// show: one tinted row with the accent edge, a padded chip, and the footer's
// summary flush against the inner edge.
func TestTheDemoCarriesTheApprovedDetails(t *testing.T) {
	view := render(140, 40)
	lines := strings.Split(view, "\n")
	plain := text.Plain(view)

	selected := 0
	for _, line := range lines {
		if strings.HasPrefix(text.Plain(line), "│ "+text.SelectMark) {
			selected++
		}
	}
	if selected != 1 {
		t.Fatalf("%d rows carry the accent edge, want exactly 1", selected)
	}
	if !strings.Contains(plain, " enter ") || strings.Contains(plain, "[enter]") {
		t.Fatalf("the hint bar does not draw a padded chip: %q", plain)
	}
	if !strings.HasSuffix(strings.TrimSuffix(text.Plain(lines[len(lines)-3]), " │"), "plugin action") {
		t.Fatalf("the footer's summary is not flush with the inner edge: %q", text.Plain(lines[len(lines)-3]))
	}
}
