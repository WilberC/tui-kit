package frame_test

import (
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
