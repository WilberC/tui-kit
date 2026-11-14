package text_test

import (
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/text"
)

// TestJoinDropRightGivesWayBeforeTheLeft pins the legend join: when both sides do
// not fit, the right side is dropped whole and the left stays readable — the
// mirror of Join, for a line whose right side is the part that can go.
func TestJoinDropRightGivesWayBeforeTheLeft(t *testing.T) {
	const title, legend = "Keybinding guide", "x after prefix · ctrl+x direct"

	full := text.JoinDropRight(title, legend, 60)
	if want := title + strings.Repeat(" ", 60-text.Width(title)-text.Width(legend)) + legend; full != want {
		t.Fatalf("JoinDropRight that fits = %q, want %q", full, want)
	}
	for _, width := range []int{20, 24, text.Width(title)} {
		if got := text.JoinDropRight(title, legend, width); got != title {
			t.Fatalf("at %d cells the legend did not give way whole: %q", width, got)
		}
	}
	if got, want := text.JoinDropRight("left", "right", 20), "left"+strings.Repeat(" ", 11)+"right"; got != want {
		t.Fatalf("JoinDropRight did not place the sides at the edges: %q", got)
	}
}

// TestJoinDropRightClipsTheLeftLast pins the only case where the left is cut: it
// does not fit on its own, so it ends with an honest ellipsis instead of pushing
// the line past the width.
func TestJoinDropRightClipsTheLeftLast(t *testing.T) {
	const title, legend = "Keybinding guide", "x after prefix"

	tiny := text.JoinDropRight(title, legend, 8)
	if got, want := text.Width(tiny), 8; got != want {
		t.Fatalf("JoinDropRight at 8 cells is %d cells wide: %q", got, tiny)
	}
	if !strings.HasSuffix(tiny, text.Ellipsis) {
		t.Fatalf("the cut left side does not end with an ellipsis: %q", tiny)
	}
	if strings.Contains(tiny, "right") {
		t.Fatalf("the dropped legend came back: %q", tiny)
	}
}

// TestJoinDropRightHandlesAnEmptySide keeps the helper safe where a legend is
// simply absent: no trailing padding, no leading gap.
func TestJoinDropRightHandlesAnEmptySide(t *testing.T) {
	if got, want := text.JoinDropRight("title", "", 20), "title"; got != want {
		t.Fatalf("JoinDropRight with no right side = %q, want %q", got, want)
	}
	if got := text.JoinDropRight("title", "right", 0); got != "" {
		t.Fatalf("JoinDropRight at no width = %q, want the empty string", got)
	}
	if got := text.JoinDropRight("a very long title indeed", "", 5); got != "a ve"+text.Ellipsis {
		t.Fatalf("JoinDropRight did not clip the lone left side: %q", got)
	}
}
