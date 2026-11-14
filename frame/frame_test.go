package frame_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/frame"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

// truecolor forces the renderer to emit colour, so a test can look for a role in
// a rendered line, and restores the profile afterwards.
func truecolor(t *testing.T) {
	t.Helper()
	before := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(before) })
}

func TestFrameClipsEveryLineToItsWidth(t *testing.T) {
	f := frame.Frame{Title: theme.GoldBold.Render("hs") + theme.Quiet.Render(" › ") + theme.AccentBold.Render("sessions"), Width: 40}
	rows := []string{
		f.Row("short"),
		f.Rule(),
		f.Row(strings.Repeat("long ", 30)),
	}

	if got, want := text.Width(f.Top()), 40; got != want {
		t.Fatalf("top border width = %d, want %d (%q)", got, want, f.Top())
	}
	if got, want := text.Width(f.Bottom()), 40; got != want {
		t.Fatalf("bottom border width = %d, want %d", got, want)
	}
	if got, want := text.Width(rows[1]), 40; got != want {
		t.Fatalf("rule width = %d, want %d", got, want)
	}
	for _, row := range []string{rows[0], rows[2]} {
		if got, want := text.Width(row), 40; got != want {
			t.Fatalf("row width = %d, want %d (%q)", got, want, row)
		}
	}
	if !strings.Contains(rows[2], text.Ellipsis) {
		t.Fatalf("a clipped row does not end with an ellipsis: %q", rows[2])
	}
}

func TestFrameWithoutATitleStillCloses(t *testing.T) {
	f := frame.Frame{Width: 24}

	if got, want := text.Width(f.Top()), 24; got != want {
		t.Fatalf("top border width = %d, want %d (%q)", got, want, f.Top())
	}
	if strings.Contains(text.Plain(f.Top()), "  ") {
		t.Fatalf("a titleless top border left a gap where the title was: %q", text.Plain(f.Top()))
	}
}

func TestContentWidthIsTheRoomInsideThePadding(t *testing.T) {
	if got, want := (frame.Frame{Width: 40}).ContentWidth(), 36; got != want {
		t.Fatalf("ContentWidth = %d, want %d", got, want)
	}
	if got, want := (frame.Frame{Width: 2}).ContentWidth(), 0; got != want {
		t.Fatalf("ContentWidth of a degenerate frame = %d, want %d", got, want)
	}
}

func TestRuleJoinsTheSides(t *testing.T) {
	rule := (frame.Frame{Width: 30}).Rule()

	if plain := text.Plain(rule); !strings.HasPrefix(plain, frame.LeftTee) || !strings.HasSuffix(plain, frame.RightTee) {
		t.Fatalf("rule does not join the frame's sides: %q", plain)
	}
	if got, want := text.Width(rule), 30; got != want {
		t.Fatalf("rule width = %d, want %d", got, want)
	}
}

func TestColumnRuleJoinsEveryColumn(t *testing.T) {
	widths := frame.ColumnWidths(78, 2)
	f := frame.Frame{Width: 80}

	open := frame.ColumnRule(widths, frame.TopTee)
	closing := frame.ColumnRule(widths, frame.BottomTee)

	for name, rule := range map[string]string{"opening": open, "closing": closing} {
		if got, want := text.Width(rule), f.Width; got != want {
			t.Fatalf("%s column rule width = %d, want %d (%q)", name, got, want, text.Plain(rule))
		}
	}
	if got, want := strings.Count(text.Plain(open), frame.TopTee), 1; got != want {
		t.Fatalf("the opening rule has %d joints, want %d: %q", got, want, text.Plain(open))
	}
	if !strings.Contains(text.Plain(closing), frame.BottomTee) {
		t.Fatalf("the closing rule has no joint: %q", text.Plain(closing))
	}
	if plain := text.Plain(open); !strings.HasPrefix(plain, frame.LeftTee) || !strings.HasSuffix(plain, frame.RightTee) {
		t.Fatalf("the column rule floats inside the frame: %q", plain)
	}
}

func TestColumnWidthsSplitTheInnerWidth(t *testing.T) {
	for _, c := range []struct{ inner, count int }{{78, 1}, {78, 2}, {138, 3}} {
		widths := frame.ColumnWidths(c.inner, c.count)
		if len(widths) != c.count {
			t.Fatalf("ColumnWidths(%d, %d) returned %d widths", c.inner, c.count, len(widths))
		}
		rule := frame.ColumnRule(widths, frame.TopTee)
		if got, want := text.Width(rule), c.inner+2; got != want {
			t.Fatalf("columns of %d/%d span %d cells, want %d", c.inner, c.count, got, want)
		}
	}
}

func TestColumnCountFollowsTheResponsiveRules(t *testing.T) {
	cases := []struct {
		inner, want, count int
	}{
		{178, 4, 3},
		{148, 4, 3},
		{147, 4, 2},
		{98, 4, 2},
		{97, 4, 1},
		{178, 1, 1},
	}
	for _, c := range cases {
		if got := frame.ColumnCount(c.inner, c.want); got != c.count {
			t.Fatalf("ColumnCount(%d, %d) = %d, want %d", c.inner, c.want, got, c.count)
		}
	}
}

func TestGridKeepsEveryRowOnTheFrame(t *testing.T) {
	f := frame.Frame{Width: 60}
	widths := frame.ColumnWidths(f.Width-2, 2)
	blocks := [][]string{
		{"left one", "left two"},
		{"right one"},
	}

	lines := frame.Grid(f, blocks, widths, 3)
	if len(lines) != 3 {
		t.Fatalf("grid drew %d lines, want 3", len(lines))
	}
	for _, line := range lines {
		if got, want := text.Width(line), f.Width; got != want {
			t.Fatalf("grid line is %d cells wide, want %d (%q)", got, want, text.Plain(line))
		}
	}
	// The vertical rule between the columns sits on one cell on every line.
	for _, line := range lines[1:] {
		plain := text.Plain(line)
		if at := strings.Index(plain, frame.Vertical); at < 0 {
			t.Fatalf("a grid line has no column rule: %q", plain)
		} else if at != strings.Index(text.Plain(lines[0]), frame.Vertical) {
			t.Fatalf("the column rule moved to cell %d: %q", at, plain)
		}
	}
}

func TestLayoutClosesTheFrame(t *testing.T) {
	f := frame.Frame{Title: theme.AccentBold.Render("demo"), Width: 30}
	view := frame.Layout(f, f.Row("one"), f.Rule(), f.Row("two"))
	lines := strings.Split(view, "\n")

	if len(lines) != 5 {
		t.Fatalf("layout has %d lines, want 5", len(lines))
	}
	if !strings.HasPrefix(text.Plain(lines[0]), frame.TopLeft) || !strings.HasPrefix(text.Plain(lines[len(lines)-1]), frame.BottomLeft) {
		t.Fatalf("layout does not close the frame: %q", view)
	}
	for _, line := range lines {
		if got, want := text.Width(line), 30; got != want {
			t.Fatalf("layout line is %d cells wide, want %d: %q", got, want, line)
		}
	}
}

func TestPlaceholderCentersTheStateAndItsWayOut(t *testing.T) {
	f := frame.Frame{Width: 40}
	rows := f.Placeholder("Nothing here", "press enter to create")

	if len(rows) != 4 {
		t.Fatalf("placeholder has %d rows, want 4", len(rows))
	}
	if rows[0] != "" || rows[len(rows)-1] != "" {
		t.Fatalf("the placeholder does not breathe: %q", rows)
	}
	for _, row := range rows {
		if got, want := text.Width(row), f.ContentWidth(); got > want {
			t.Fatalf("placeholder row is %d cells wide, want at most %d: %q", got, want, row)
		}
	}
	if !strings.Contains(text.Plain(rows[1]), "Nothing here") || !strings.Contains(text.Plain(rows[2]), "press enter") {
		t.Fatalf("placeholder lost its two lines: %q", rows)
	}
}

func TestFloatCentersTheBlock(t *testing.T) {
	block := "╭─╮\n│x│\n╰─╯"
	placed := frame.Float(block, 20, 7)
	lines := strings.Split(placed, "\n")

	if len(lines) != 5 {
		t.Fatalf("floated block is %d rows tall, want 5 (%q)", len(lines), placed)
	}
	if strings.TrimSpace(lines[0]) != "" || strings.TrimSpace(lines[1]) != "" {
		t.Fatalf("the block is not centered vertically: %q", placed)
	}
	if want := strings.Repeat(" ", 8) + "╭─╮"; lines[2] != want {
		t.Fatalf("the block is not centered horizontally: %q, want %q", lines[2], want)
	}
	if got, want := text.Width(lines[2]), 11; got != want {
		t.Fatalf("a floated line is %d cells wide, want the indent plus the block (%d): %q", got, want, lines[2])
	}
	if got := frame.Float(block, 2, 1); got != block {
		t.Fatalf("a block wider than the window was clipped instead of kept: %q", got)
	}
}

func TestFooterKeepsTheSummaryFlush(t *testing.T) {
	footer := text.Plain(frame.Footer(theme.AccentBold.Render("work")+" "+theme.Muted.Render("running"), "4 workspaces", 40))

	if got, want := text.Width(footer), 40; got != want {
		t.Fatalf("footer width = %d, want %d", got, want)
	}
	if !strings.HasSuffix(footer, "4 workspaces") {
		t.Fatalf("footer does not end with the summary: %q", footer)
	}
	gutter := strings.TrimSuffix(footer, "4 workspaces")
	if len(gutter) < len("work running")+2 || strings.TrimSpace(gutter) != "work running" {
		t.Fatalf("footer has no plain gutter before the summary: %q", footer)
	}
	if strings.Contains(footer, " · ") || strings.Contains(footer, "│") {
		t.Fatalf("footer invented a separator: %q", footer)
	}

	long := text.Plain(frame.Footer(strings.Repeat("fact ", 12), "4 workspaces", 40))
	if got, want := text.Width(long), 40; got != want {
		t.Fatalf("a clipped footer is %d cells wide, want %d (%q)", got, want, long)
	}
	if !strings.HasSuffix(long, "4 workspaces") {
		t.Fatalf("clipping dropped the summary: %q", long)
	}
	if got, want := frame.Footer("only left", "", 40), "only left"; got != want {
		t.Fatalf("a footer without a summary = %q, want %q", got, want)
	}
}

// TestFrameFollowsTheProfileAtCallTime pins that no frame glyph is rendered once
// at init: the profile in force at draw time decides the colour.
func TestFrameFollowsTheProfileAtCallTime(t *testing.T) {
	before := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(before) })

	f := frame.Frame{Title: "demo", Width: 20}
	if strings.ContainsRune(f.Top(), 0x1b) {
		t.Fatalf("an Ascii frame carries escapes: %q", f.Top())
	}

	lipgloss.SetColorProfile(termenv.TrueColor)
	paint := strings.SplitN(theme.Border.Render("x"), "x", 2)[0]
	if !strings.Contains(f.Top(), paint) {
		t.Fatalf("the frame does not use the border role at draw time: %q", f.Top())
	}
	if !strings.Contains(f.Rule(), paint) {
		t.Fatalf("the rule does not use the border role at draw time: %q", f.Rule())
	}
}

// TestFrameSpendsNoColourOnWhatIsNotAFrameGlyph pins the frame's paint: the
// interior rule body is quiet, and the borders carry the border role.
func TestFrameSpendsNoColourOnWhatIsNotAFrameGlyph(t *testing.T) {
	truecolor(t)
	rule := (frame.Frame{Width: 20}).Rule()

	if !strings.Contains(rule, strings.SplitN(theme.Quiet.Render("x"), "x", 2)[0]) {
		t.Fatalf("the rule's body is not the guide role: %q", rule)
	}
	if !strings.Contains(rule, strings.SplitN(theme.Border.Render("x"), "x", 2)[0]) {
		t.Fatalf("the rule's joints are not the border role: %q", rule)
	}
}

// rowContent is what a framed row carries between its borders: the border glyphs
// and the padding taken away, so a blank row reads as blank.
func rowContent(line string) string {
	plain := strings.TrimSuffix(strings.TrimPrefix(text.Plain(line), frame.Vertical), frame.Vertical)
	return strings.TrimSpace(plain)
}

// framedRows renders the body contents the way a product hands them in: already
// framed, so a row's own rules and columns survive.
func framedRows(surface frame.Surface, contents ...string) []string {
	rows := make([]string, len(contents))
	for index, content := range contents {
		rows[index] = surface.Frame.Row(content)
	}
	return rows
}

// bounded is the surface the size tests size: a header row, a two-line footer and
// narrow bounds, so a pane both larger and smaller than the range is exercised.
// The rows are framed for the clamped width, the way a product composes them.
func bounded(paneWidth int) frame.Surface {
	size := frame.Size{MinWidth: 40, MaxWidth: 120, MinHeight: 12, MaxHeight: 20}
	f := frame.Frame{Title: "demo", Width: size.ClampWidth(paneWidth)}
	return frame.Surface{
		Frame:  f,
		Size:   size,
		Top:    []string{f.Row("HEADER")},
		Bottom: []string{f.Row("DETAIL"), f.Row("HINTS")},
	}
}

func TestSizeClampsBothWays(t *testing.T) {
	size := frame.Size{MinWidth: 60, MaxWidth: 160, MinHeight: 16, MaxHeight: 40}

	cases := []struct{ paneWidth, paneHeight, width, height int }{
		{200, 60, 160, 40}, // past the maximum: the frame stops growing
		{100, 30, 100, 30}, // inside the range: the frame is the pane
		{60, 16, 60, 16},   // exactly the minimum
		{160, 40, 160, 40}, // exactly the maximum
		{40, 10, 40, 10},   // a smaller pane wins: a frame never exceeds its pane
	}
	for _, c := range cases {
		if got := size.ClampWidth(c.paneWidth); got != c.width {
			t.Fatalf("ClampWidth(%d) = %d, want %d", c.paneWidth, got, c.width)
		}
		if got := size.ClampHeight(c.paneHeight); got != c.height {
			t.Fatalf("ClampHeight(%d) = %d, want %d", c.paneHeight, got, c.height)
		}
		width, height := size.Clamp(c.paneWidth, c.paneHeight)
		if width != c.width || height != c.height {
			t.Fatalf("Clamp(%d, %d) = (%d, %d), want (%d, %d)", c.paneWidth, c.paneHeight, width, height, c.width, c.height)
		}
	}

	var none frame.Size
	if got := none.ClampWidth(42); got != 42 {
		t.Fatalf("an unset Size clamped the width to %d, want the pane's 42", got)
	}
	if got := none.ClampHeight(17); got != 17 {
		t.Fatalf("an unset Size clamped the height to %d, want the pane's 17", got)
	}
}

// TestSurfaceIsAsTallAsTheClampedPane pins the fill rule: the frame's height is
// the clamped pane height, never the height of its content.
func TestSurfaceIsAsTallAsTheClampedPane(t *testing.T) {
	for _, pane := range []int{5, 11, 12, 16, 20, 40} {
		surface := bounded(80)
		bodies := map[string][]string{
			"no body":        nil,
			"one row":        framedRows(surface, "one"),
			"three rows":     framedRows(surface, "one", "two", "three"),
			"a hundred rows": framedRows(surface, make([]string, 100)...),
		}
		want := surface.Height(pane)
		for name, body := range bodies {
			lines := strings.Split(surface.Render(80, pane, body), "\n")
			if len(lines) != want {
				t.Fatalf("pane %d with %s: the frame is %d rows tall, want %d", pane, name, len(lines), want)
			}
		}
	}
}

// TestSurfacePinsTheFooterToTheLastInnerRows pins the anchoring rule: whatever the
// body holds, the footer keeps its rows against the closing border.
func TestSurfacePinsTheFooterToTheLastInnerRows(t *testing.T) {
	surface := bounded(80)
	bodies := [][]string{
		nil,
		framedRows(surface, "one"),
		framedRows(surface, "one", "two", "three", "four", "five", "six", "seven", "eight"),
	}
	for _, body := range bodies {
		lines := strings.Split(surface.Render(80, 20, body), "\n")

		if got := rowContent(lines[1]); got != "HEADER" {
			t.Fatalf("%d body rows: the header is not under the top border: %q", len(body), got)
		}
		if got := rowContent(lines[len(lines)-3]); got != "DETAIL" {
			t.Fatalf("%d body rows: the detail line is not the last footer row: %q", len(body), got)
		}
		if got := rowContent(lines[len(lines)-2]); got != "HINTS" {
			t.Fatalf("%d body rows: the hint bar is not pinned above the closing border: %q", len(body), got)
		}
		if !strings.HasPrefix(text.Plain(lines[len(lines)-1]), frame.BottomLeft) {
			t.Fatalf("%d body rows: the frame does not close after the footer: %q", len(body), text.Plain(lines[len(lines)-1]))
		}
	}
}

// TestSurfacePadsTheBodyAtBothEnds pins the padding rule: the body always opens
// and closes with a blank row, and the room it does not use stays inside it.
func TestSurfacePadsTheBodyAtBothEnds(t *testing.T) {
	surface := bounded(80)
	bodies := [][]string{
		nil,
		framedRows(surface, "one"),
		framedRows(surface, "one", "two", "three"),
	}
	for _, body := range bodies {
		lines := strings.Split(surface.Render(80, 20, body), "\n")

		if got := rowContent(lines[2]); got != "" {
			t.Fatalf("%d body rows: the body does not open with a blank row: %q", len(body), got)
		}
		if got := rowContent(lines[len(lines)-4]); got != "" {
			t.Fatalf("%d body rows: the body does not close with a blank row: %q", len(body), got)
		}
		for index, want := range body {
			if got := rowContent(lines[3+index]); got != rowContent(want) {
				t.Fatalf("%d body rows: row %d = %q, want %q", len(body), index, got, rowContent(want))
			}
		}
		if got := rowContent(lines[3+len(body)]); got != "" {
			t.Fatalf("%d body rows: the spare room is not below the content: %q", len(body), got)
		}
	}
}

// TestSurfaceShowsExactlyTheBodyRows pins the contract a product pages with: the
// body holds the count BodyRows answers, and anything past it is dropped.
func TestSurfaceShowsExactlyTheBodyRows(t *testing.T) {
	surface := bounded(80) // 20 - 2 borders - 1 header - 2 footer - 2 padding = 13
	if got, want := surface.BodyRows(20), 13; got != want {
		t.Fatalf("BodyRows(20) = %d, want %d", got, want)
	}
	if got, want := surface.BodyRows(60), 13; got != want {
		t.Fatalf("BodyRows(60) = %d, want the clamped %d", got, want)
	}
	if got, want := surface.BodyRows(12), 5; got != want {
		t.Fatalf("BodyRows(12) = %d, want %d", got, want)
	}
	if got := surface.BodyRows(5); got != 0 {
		t.Fatalf("BodyRows(5) = %d, want 0 when the pane cannot hold a body", got)
	}
	if got, want := surface.Height(60), 20; got != want {
		t.Fatalf("Height(60) = %d, want the clamped %d", got, want)
	}

	numbers := make([]string, 0, 100)
	for index := 1; index <= 100; index++ {
		numbers = append(numbers, strconv.Itoa(index))
	}
	body := framedRows(surface, numbers...)
	lines := strings.Split(surface.Render(80, 20, body), "\n")
	room := surface.BodyRows(20)
	if got, want := rowContent(lines[2+room]), strconv.Itoa(room); got != want {
		t.Fatalf("the last visible body row = %q, want %q", got, want)
	}
}

// TestSurfaceNeverExceedsThePaneWidth pins the width rule: every frame line is
// the clamped width, so no line ever exceeds the pane.
func TestSurfaceNeverExceedsThePaneWidth(t *testing.T) {
	for _, pane := range []int{30, 40, 80, 120, 200} {
		surface := bounded(pane)
		width := surface.Width(pane)
		view := surface.Render(pane, 20, framedRows(surface, "a row"))
		for _, line := range strings.Split(view, "\n") {
			if got := text.Width(line); got != width {
				t.Fatalf("pane %d: a frame line is %d cells wide, want the clamped %d", pane, got, width)
			}
			if got := text.Width(line); got > pane {
				t.Fatalf("pane %d: a frame line is %d cells wide, past the pane", pane, got)
			}
		}
	}
}

// TestPopupCentresTheClampedFrame pins what a product draws: the clamped frame
// sits in the middle of the pane instead of hugging its top edge.
func TestPopupCentresTheClampedFrame(t *testing.T) {
	const paneWidth, paneHeight = 200, 50
	surface := bounded(paneWidth)
	lines := strings.Split(surface.Popup(paneWidth, paneHeight, framedRows(surface, "one", "two")), "\n")

	frameHeight, frameWidth := surface.Height(paneHeight), surface.Width(paneWidth)
	top, left := (paneHeight-frameHeight)/2, (paneWidth-frameWidth)/2
	if len(lines) != top+frameHeight {
		t.Fatalf("the popup is %d rows tall, want %d", len(lines), top+frameHeight)
	}
	for index := 0; index < top; index++ {
		if strings.TrimSpace(lines[index]) != "" {
			t.Fatalf("row %d above the frame is not blank: %q", index, lines[index])
		}
	}
	for index := top; index < len(lines); index++ {
		if got, want := text.Width(lines[index]), left+frameWidth; got != want {
			t.Fatalf("row %d is %d cells wide, want the indent plus the frame (%d)", index, got, want)
		}
	}
}

// separatorCells is the cell where every column rule sits on a rendered line: the
// frame's two sides and the rules between the columns.
func separatorCells(line string) []int {
	cells := []int{}
	cell := 0
	for _, glyph := range text.Plain(line) {
		if string(glyph) == frame.Vertical {
			cells = append(cells, cell)
		}
		cell += text.Width(string(glyph))
	}
	return cells
}

// cellOf is the cell the first glyph sits on, or -1.
func cellOf(line string, wanted rune) int {
	cell := 0
	for _, glyph := range text.Plain(line) {
		if glyph == wanted {
			return cell
		}
		cell += text.Width(string(glyph))
	}
	return -1
}

// TestSurfaceKeepsTheColumnRuleThroughTheBody pins the grid contract: in a
// two-column surface every interior body row — padding and spare rows included —
// carries the separator in the same cell, so the column rule runs from the opening
// joint to the closing one without a break.
func TestSurfaceKeepsTheColumnRuleThroughTheBody(t *testing.T) {
	size := frame.Size{MinWidth: 40, MaxWidth: 120, MinHeight: 12, MaxHeight: 20}
	f := frame.Frame{Title: "demo", Width: 80}
	widths := frame.ColumnWidths(f.Width-2, 2)
	surface := frame.Surface{
		Frame:   f,
		Size:    size,
		Columns: widths,
		Top:     []string{f.Row("HEADER"), frame.ColumnRule(widths, frame.TopTee)},
		Bottom:  []string{frame.ColumnRule(widths, frame.BottomTee), f.Row("FOOTER")},
	}
	blocks := [][]string{{"left one", "left two"}, {"right one"}}
	lines := strings.Split(surface.Render(80, 20, frame.Grid(f, blocks, widths, surface.BodyRows(20))), "\n")

	joint := cellOf(lines[2], '┬')
	if joint < 0 {
		t.Fatalf("the opening rule carries no joint: %q", text.Plain(lines[2]))
	}
	want := []int{0, joint, f.Width - 1}

	// Interior: one header row, the opening rule, the body, the closing rule, one
	// footer row.
	body := lines[3 : len(lines)-3]
	if len(body) != surface.BodyRows(20)+2 {
		t.Fatalf("the body region is %d rows, want %d", len(body), surface.BodyRows(20)+2)
	}
	for index, line := range body {
		got := separatorCells(line)
		if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
			t.Fatalf("body row %d has its separators at %v, want %v: %q", index, got, want, text.Plain(line))
		}
	}
	if got := cellOf(lines[len(lines)-3], '┴'); got != joint {
		t.Fatalf("the closing joint is at cell %d, want the column's %d", got, joint)
	}

	// A plain body keeps plain blank rows: no separator in the middle.
	plain := bounded(80)
	plainLines := strings.Split(plain.Render(80, 20, framedRows(plain, "one")), "\n")
	if got := separatorCells(plainLines[2]); len(got) != 2 {
		t.Fatalf("a plain body grew a column separator: %q", text.Plain(plainLines[2]))
	}
}
