package main

import (
	"strconv"
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/frame"
	"github.com/WilberC/tui-kit/text"
)

// TestTheDemoFitsItsWindow pins the surface contract the screenshots rely on:
// the frame is exactly the clamped pane size, it is centred when the pane is
// larger, and no line ever exceeds the pane it is drawn in — at the captured
// sizes and at every width between the narrowest terminal and the widest one.
func TestTheDemoFitsItsWindow(t *testing.T) {
	windows := []struct{ width, height int }{
		{180, 50},
		{140, 40},
		{140, 50},
		{80, 30},
		{80, 24},
	}
	for width := 40; width <= 200; width++ {
		windows = append(windows, struct{ width, height int }{width, 30})
	}
	for _, window := range windows {
		view := render(window.width, window.height)
		lines := strings.Split(view, "\n")

		frameWidth := frame.Recommended.ClampWidth(window.width)
		frameHeight := frame.Recommended.ClampHeight(window.height)
		top := max(0, (window.height-frameHeight)/2)
		left := max(0, (window.width-frameWidth)/2)

		if len(lines) != top+frameHeight {
			t.Fatalf("%dx%d: the popup is %d rows tall, want %d", window.width, window.height, len(lines), top+frameHeight)
		}
		for index, line := range lines {
			if width := text.Width(line); width > window.width {
				t.Fatalf("%dx%d: line %d is %d cells wide: %s", window.width, window.height, index, width, text.Plain(line))
			}
			if index < top {
				if strings.TrimSpace(line) != "" {
					t.Fatalf("%dx%d: the row above the frame is not blank: %q", window.width, window.height, line)
				}
				continue
			}
			if width := text.Width(line); width != left+frameWidth {
				t.Fatalf("%dx%d: frame line %d is %d cells wide, want %d", window.width, window.height, index, width, left+frameWidth)
			}
		}
	}
}

// TestTheDemoPagesItsHintsWhenNarrow pins what the narrow captures are meant to
// show: below 80 columns the hint bar ends with the paging chip instead of
// silently hiding keys, and every page it holds stays inside the frame.
func TestTheDemoPagesItsHintsWhenNarrow(t *testing.T) {
	for _, width := range []int{80, 60} {
		pages := (chip.HintBar{}).Pages(width-2, demoHints...)
		if pages < 2 {
			t.Fatalf("%d wide: the hint bar does not page (%d pages)", width, pages)
		}
		for page := 0; page < pages; page++ {
			lines := strings.Split(renderPage(width, 30, page), "\n")
			bar := text.Plain(lines[len(lines)-2])
			want := " " + chip.PageKey + "  more " + strconv.Itoa(page+1) + "/" + strconv.Itoa(pages) + " "
			if !strings.Contains(bar, want) {
				t.Fatalf("%d wide page %d: the hint bar does not carry %q:\n%s", width, page, want, bar)
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

// rowContent is what a framed row carries between its borders: the border and
// column glyphs and the padding taken away, so a padding row reads as blank even
// when it keeps the column separators.
func rowContent(line string) string {
	plain := strings.TrimSuffix(strings.TrimPrefix(text.Plain(line), "│"), "│")
	return strings.TrimSpace(strings.ReplaceAll(plain, "│", ""))
}

// TestTheDemoPinsItsFooterAndPadsItsBody pins what the sizing captures are meant
// to show on a tall pane: the frame is as tall as the clamped pane, the footer is
// its last inner rows, and the body keeps one blank row at each end.
func TestTheDemoPinsItsFooterAndPadsItsBody(t *testing.T) {
	const paneWidth, paneHeight = 140, 50
	lines := strings.Split(renderPage(paneWidth, paneHeight, 0), "\n")

	frameHeight := frame.Recommended.ClampHeight(paneHeight)
	top := (paneHeight - frameHeight) / 2
	if len(lines)-top != frameHeight {
		t.Fatalf("the frame is %d rows tall, want the clamped %d", len(lines)-top, frameHeight)
	}
	framed := lines[top:]
	last := len(framed) - 1

	if !strings.HasPrefix(text.Plain(framed[last]), "╰") {
		t.Fatalf("the frame does not close at its last row: %q", text.Plain(framed[last]))
	}
	if got := rowContent(framed[last-1]); !strings.Contains(got, "enter") {
		t.Fatalf("the hint bar is not the last inner row: %q", got)
	}
	if got := rowContent(framed[last-2]); !strings.Contains(got, "plugin action") {
		t.Fatalf("the detail line is not above the hint bar: %q", got)
	}
	if got := rowContent(framed[last-3]); !strings.Contains(got, "┴") {
		t.Fatalf("the closing column rule is not above the footer: %q", got)
	}
	if got := rowContent(framed[last-4]); got != "" {
		t.Fatalf("the body does not close with a blank row: %q", got)
	}
	if got := rowContent(framed[5]); got != "" {
		t.Fatalf("the body does not open with a blank row: %q", got)
	}
}
