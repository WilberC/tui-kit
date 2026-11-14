package main

import (
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/frame"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
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
		var bar chip.HintBar
		pages := bar.Pages(width-2, demoHints...)
		if pages < 2 {
			t.Fatalf("%d wide: the hint bar does not page (%d pages)", width, pages)
		}
		for page := 0; page < pages; page++ {
			lines := strings.Split(renderPage(width, 30, page).pane, "\n")
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
	lines := strings.Split(renderPage(paneWidth, paneHeight, 0).pane, "\n")

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

// TestTheDemoRecordsWhatAClickCanHit pins S8 end to end and prints the map the
// render recorded, so the mouse wiring can be read without a mouse: every listed
// row answers from its own cells, the paging chip answers as chip.PageZone, and
// no zone escapes the frame it was drawn in.
func TestTheDemoRecordsWhatAClickCanHit(t *testing.T) {
	// The narrow window cannot hold all twelve entries: the body stops at
	// BodyRows, and the rows it never drew are not clickable either.
	for _, window := range []struct {
		width, height int
		rows          int
	}{{180, 50, 12}, {80, 24, 9}} {
		view := renderPage(window.width, window.height, 0)
		if view.zones.Len() == 0 {
			t.Fatalf("%dx%d: the render recorded no zone", window.width, window.height)
		}
		for _, z := range view.zones.Zones() {
			t.Logf("%dx%d: %-12s %3d,%3d %3dx%d", window.width, window.height, z.ID, z.Rect.X, z.Rect.Y, z.Rect.W, z.Rect.H)
		}

		frameWidth := frame.Recommended.ClampWidth(window.width)
		frameHeight := frame.Recommended.ClampHeight(window.height)
		left := max(0, (window.width-frameWidth)/2)
		top := max(0, (window.height-frameHeight)/2)
		for _, z := range view.zones.Zones() {
			if z.Rect.X < left+2 || z.Rect.Y < top+1 ||
				z.Rect.Right() > left+frameWidth-2 || z.Rect.Bottom() > top+frameHeight-1 {
				t.Fatalf("%dx%d: zone %q %v is not inside the frame", window.width, window.height, z.ID, z.Rect)
			}
		}

		rows := 0
		for _, z := range view.zones.Zones() {
			if _, ok := rowIndex(z.ID); !ok {
				continue
			}
			rows++
			if id, ok := view.zones.Hit(z.Rect.X, z.Rect.Y); !ok || id != z.ID {
				t.Fatalf("%dx%d: the first cell of %q answered %q/%v", window.width, window.height, z.ID, id, ok)
			}
		}
		if rows != window.rows {
			t.Fatalf("%dx%d: %d rows are clickable, want %d", window.width, window.height, rows, window.rows)
		}

		if zones := view.zones.Len(); zones < rows {
			t.Fatalf("%dx%d: %d zones for %d rows", window.width, window.height, zones, rows)
		}
	}

	// A bar that fits draws no paging chip, so it records no paging zone: a
	// handler wired to it can never fire on a chip that is not there.
	if _, ok := renderPage(180, 50, 0).zones.Find(chip.PageZone); ok {
		t.Fatalf("the wide bar recorded a paging chip it does not draw")
	}
	for _, width := range []int{80, 60} {
		page, ok := renderPage(width, 24, 0).zones.Find(chip.PageZone)
		if !ok {
			t.Fatalf("%d wide: the paging chip has no zone", width)
		}
		if id, ok := renderPage(width, 24, 0).zones.Hit(page.X+2, page.Y); !ok || id != chip.PageZone {
			t.Fatalf("%d wide: a click on the paging chip answered %q/%v", width, id, ok)
		}
	}
}

// TestTheDemoHoversARowNextToTheSelection pins what the hover capture shows: the
// row beside the selected one wears the faint hover tint and no accent edge, so
// both tints can be judged side by side.
func TestTheDemoHoversARowNextToTheSelection(t *testing.T) {
	before := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(before) })

	hover := strings.SplitN(theme.HoverBG.Render("x"), "x", 2)[0]
	selection := strings.SplitN(theme.SelectionBG.Render("x"), "x", 2)[0]

	lines := strings.Split(renderAt(140, 40, 0, hoveredRow).pane, "\n")
	hovered, selected := 0, 0
	for _, line := range lines {
		if strings.Contains(line, hover) {
			hovered++
			if strings.HasPrefix(text.Plain(line), "│ "+text.SelectMark) {
				t.Fatalf("the hovered row carries the selection's edge: %q", text.Plain(line))
			}
		}
		if strings.HasPrefix(text.Plain(line), "│ "+text.SelectMark) {
			selected++
		}
	}
	if hovered != 1 {
		t.Fatalf("%d rows carry the hover tint, want exactly 1", hovered)
	}
	if selected != 1 {
		t.Fatalf("%d rows carry the selection, want exactly 1", selected)
	}
	if strings.Count(text.Plain(strings.Join(lines, "\n")), text.SelectMark) != 1 {
		t.Fatalf("the hover added a second accent edge")
	}
	if hover == selection {
		t.Fatalf("the hover tint and the selection tint are the same escape")
	}
}

// TestTheDemoPagesBehindTheSameChipItRecords pins the paging wiring: the zone the
// bar records says which page it is on, and the next page moves the paging chip's
// label without moving the key.
func TestTheDemoPagesBehindTheSameChipItRecords(t *testing.T) {
	first := renderPage(80, 24, 0)
	second := renderPage(80, 24, 1)

	page, ok := first.zones.Find(chip.PageZone)
	if !ok {
		t.Fatalf("the first page records no paging zone")
	}
	if id, ok := first.zones.Hit(page.X, page.Y); !ok || id != chip.PageZone {
		t.Fatalf("a click at the paging zone's first cell answered %q/%v", id, ok)
	}
	lines := strings.Split(first.pane, "\n")
	bar := text.Plain(lines[len(lines)-2])
	if !strings.Contains(bar, "more 1/") {
		t.Fatalf("the first page does not say it is the first: %q", bar)
	}
	lines = strings.Split(second.pane, "\n")
	if bar := text.Plain(lines[len(lines)-2]); !strings.Contains(bar, "more 2/") {
		t.Fatalf("the second page does not say it is the second: %q", bar)
	}
	if _, ok := second.zones.Find(chip.PageZone); !ok {
		t.Fatalf("the second page records no paging zone")
	}
}

// TestTheDemoFramesItsRules pins the regression the visual review caught: a rule
// is a whole line of the frame and is placed as it is, never handed to Row/Rows.
// A double-framed rule starts one cell in, grows a vertical tick and ends with an
// ellipsis instead of its joint.
func TestTheDemoFramesItsRules(t *testing.T) {
	for _, window := range []struct{ width, height int }{{180, 50}, {140, 40}, {80, 24}, {60, 30}} {
		view := renderPage(window.width, window.height, 0)
		frameWidth := frame.Recommended.ClampWidth(window.width)
		lines := strings.Split(view.pane, "\n")

		header, filter := -1, -1
		for index, line := range lines {
			plain := text.Plain(line)
			if header < 0 && strings.Contains(plain, "12 bindings") {
				header = index
			}
			if filter < 0 && strings.Contains(plain, "Filter keys") {
				filter = index
			}
		}
		if header < 0 || filter < 0 {
			t.Fatalf("%dx%d: the demo draws no header or no filter row", window.width, window.height)
		}

		// The rule under the header and the column rule under the filter are
		// whole lines: they join both borders, span the frame and are never cut.
		for name, index := range map[string]int{"the rule under the header": header + 1, "the column rule": filter + 1} {
			plain := strings.TrimLeft(text.Plain(lines[index]), " ")
			if !strings.HasPrefix(plain, frame.LeftTee) || !strings.HasSuffix(plain, frame.RightTee) {
				t.Fatalf("%dx%d: %s does not join both borders: %q", window.width, window.height, name, plain)
			}
			if got, want := text.Width(plain), frameWidth; got != want {
				t.Fatalf("%dx%d: %s is %d cells wide, want the frame's %d: %q", window.width, window.height, name, got, want, plain)
			}
			if strings.Contains(plain, text.Ellipsis) {
				t.Fatalf("%dx%d: %s was clipped: %q", window.width, window.height, name, plain)
			}
			body := strings.TrimSuffix(strings.TrimPrefix(plain, frame.LeftTee), frame.RightTee)
			body = strings.NewReplacer(frame.TopTee, "", frame.BottomTee, "", frame.Horizontal, "").Replace(body)
			if body != "" {
				t.Fatalf("%dx%d: %s carries %q, so a vertical tick runs through it: %q",
					window.width, window.height, name, body, plain)
			}
		}

		// No framed line anywhere in the popup ends with a cut against its border.
		for index, line := range lines {
			plain := text.Plain(line)
			for _, end := range []string{text.Ellipsis + frame.Vertical, text.Ellipsis + frame.RightTee, text.Ellipsis + frame.BottomRight} {
				if strings.Contains(plain, end) {
					t.Fatalf("%dx%d: line %d is cut against its border: %q", window.width, window.height, index, plain)
				}
			}
		}
	}
}

// TestTheDemoKeepsItsWholeLegend pins the filter row's legend at the widths the
// captures use: the legend only gives way whole, and at 140 columns every item it
// showed in v0.2.0 still fits.
func TestTheDemoKeepsItsWholeLegend(t *testing.T) {
	for _, width := range []int{180, 140} {
		lines := strings.Split(renderPage(width, 40, 0).pane, "\n")
		filter := ""
		for _, line := range lines {
			if strings.Contains(text.Plain(line), "Filter keys") {
				filter = text.Plain(line)
			}
		}
		if filter == "" {
			t.Fatalf("%d wide: the demo draws no filter row", width)
		}
		for _, item := range []string{"after prefix", "direct", "unset"} {
			if !strings.Contains(filter, item) {
				t.Fatalf("%d wide: the legend dropped %q: %q", width, item, filter)
			}
		}
		if strings.Contains(filter, text.Ellipsis+" ") && strings.Contains(filter, "unse"+text.Ellipsis) {
			t.Fatalf("%d wide: the legend was cut instead of dropped whole: %q", width, filter)
		}
	}
}
