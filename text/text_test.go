package text_test

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/text"
)

// The width contract is the first guarantee of the kit: whatever a caller hands
// in, a fitted string occupies exactly the cells it was given and a clipped one
// never exceeds them.

func TestFittingNeverLeavesTheTarget(t *testing.T) {
	cases := []struct {
		text  string
		width int
	}{
		{"work", 8},
		{"gentle-shell", 6},
		{lipgloss.NewStyle().Bold(true).Render("styled"), 3},
		{"", 4},
		{"日本語のセッション", 5},
		{"one", 0},
	}
	for _, c := range cases {
		if got := text.Width(text.Clip(c.text, c.width)); got > c.width {
			t.Fatalf("Clip(%q, %d) is %d cells wide", c.text, c.width, got)
		}
		for name, got := range map[string]string{
			"Fit":    text.Fit(c.text, c.width),
			"Center": text.Center(c.text, c.width),
		} {
			if width := text.Width(got); width != c.width {
				t.Fatalf("%s(%q, %d) is %d cells wide, want %d (%q)", name, c.text, c.width, width, c.width, got)
			}
		}
		if width := text.Width(text.PadLeft(c.text, c.width)); width != c.width {
			t.Fatalf("PadLeft(%q, %d) is %d cells wide, want %d", c.text, c.width, width, c.width)
		}
	}
}

func TestClipEndsAnHonestCutWithAnEllipsis(t *testing.T) {
	got := text.Clip("gentle-shell", 6)

	if !strings.HasSuffix(got, text.Ellipsis) {
		t.Fatalf("a clipped string does not end with an ellipsis: %q", got)
	}
	if full := text.Clip("work", 8); full != "work" {
		t.Fatalf("a string that fits was changed: %q", full)
	}
}

func TestClipLeftKeepsTheTailOfAPath(t *testing.T) {
	path := "/home/ana/.local/share/herdr-suite/layouts/work.json"
	clipped := text.ClipLeft(path, 24)

	if got, want := text.Width(clipped), 24; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if !strings.HasPrefix(clipped, text.Ellipsis) || !strings.HasSuffix(clipped, "work.json") {
		t.Fatalf("a clipped path kept the wrong end: %q", clipped)
	}
	if got := text.ClipLeft(path, 200); got != path {
		t.Fatalf("a path that fits was clipped: %q", got)
	}
}

func TestEllipsisPathKeepsTheSegmentsThatIdentifyIt(t *testing.T) {
	path := "/home/ana/.local/share/herdr-suite/layouts/work.json"
	short := text.EllipsisPath(path, 24)

	if got := text.Width(short); got > 24 {
		t.Fatalf("ellipsised path is %d cells wide, want at most 24 (%q)", got, short)
	}
	if !strings.HasPrefix(short, text.Ellipsis+"/") {
		t.Fatalf("ellipsised path does not open with the ellipsis segment: %q", short)
	}
	if !strings.HasSuffix(short, "work.json") {
		t.Fatalf("ellipsised path lost the segment that identifies it: %q", short)
	}
	if got := text.EllipsisPath(path, 200); got != path {
		t.Fatalf("a path that fits was changed: %q", got)
	}
	if got := text.Width(text.EllipsisPath(path, 2)); got > 2 {
		t.Fatalf("a two-cell budget produced %d cells: %q", got, text.EllipsisPath(path, 2))
	}
}

func TestJoinKeepsTheRightSideWhenBothCannotFit(t *testing.T) {
	got := text.Join("a very long left side indeed", "7 / 8 shown", 20)

	if width := text.Width(got); width != 20 {
		t.Fatalf("width = %d, want 20 (%q)", width, got)
	}
	if !strings.HasSuffix(got, "7 / 8 shown") {
		t.Fatalf("join dropped the right side: %q", got)
	}
	if got := text.Join("left", "right", 20); got != "left"+strings.Repeat(" ", 11)+"right" {
		t.Fatalf("join did not place the sides at the edges: %q", got)
	}
}

func TestShortPathReplacesTheHomePrefixWithATilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if got, want := text.ShortPath(home+"/work/api"), "~/work/api"; got != want {
		t.Fatalf("ShortPath = %q, want %q", got, want)
	}
	if got, want := text.ShortPath(home), "~"; got != want {
		t.Fatalf("ShortPath of the home directory = %q, want %q", got, want)
	}
	if got, want := text.ShortPath("/etc/herdr"), "/etc/herdr"; got != want {
		t.Fatalf("ShortPath changed a path outside home: %q", got)
	}
}

func TestRelativeTimeNamesTheAge(t *testing.T) {
	now := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		when time.Time
		want string
	}{
		{now.Add(-10 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5m ago"},
		{now.Add(-2 * time.Hour), "2h ago"},
		{now.Add(-26 * time.Hour), "1d ago"},
		{now.Add(-31 * 24 * time.Hour), "1mo ago"},
		{now.Add(-400 * 24 * time.Hour), "1y ago"},
		{time.Time{}, ""},
	}
	for _, c := range cases {
		if got := text.RelativeTime(now, c.when); got != c.want {
			t.Fatalf("RelativeTime(%v) = %q, want %q", c.when, got, c.want)
		}
	}
}

func TestPlainStripsStyling(t *testing.T) {
	styled := lipgloss.NewStyle().Bold(true).Render("work")

	if got, want := text.Plain(styled), "work"; got != want {
		t.Fatalf("Plain = %q, want %q", got, want)
	}
	if got, want := text.Width(styled), 4; got != want {
		t.Fatalf("Width counts escapes: %d, want %d", got, want)
	}
}

func TestEveryGlyphIsOneCell(t *testing.T) {
	for name, glyph := range map[string]string{
		"Ellipsis":   text.Ellipsis,
		"Live":       text.Live,
		"Idle":       text.Idle,
		"Mark":       text.Mark,
		"SelectMark": text.SelectMark,
		"Cursor":     text.Cursor,
		"Unset":      text.Unset,
	} {
		if width := text.Width(glyph); width != 1 {
			t.Fatalf("%s = %q is %d cells wide, want 1", name, glyph, width)
		}
	}
}
