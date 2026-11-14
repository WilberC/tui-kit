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

// TestJoinDropRightKeepsOneCellBetweenTheSides pins the gutter rule: the two sides
// never touch, even when they fit exactly, so a product stops padding the right
// side with a leading space to buy that cell. The result is the same line the
// padded call used to draw, at every width.
func TestJoinDropRightKeepsOneCellBetweenTheSides(t *testing.T) {
	const left, right = "left", "right"

	// Exactly enough room for both, no room for the gutter: the right side gives
	// way whole, exactly as it did when the caller padded it with a space.
	if got, want := text.JoinDropRight(left, right, 9), left; got != want {
		t.Fatalf("JoinDropRight at an exact fit = %q, want the trimmed %q", got, want)
	}
	// One spare cell is the gutter, not padding around it.
	if got, want := text.JoinDropRight(left, right, 10), left+" "+right; got != want {
		t.Fatalf("JoinDropRight with one spare cell = %q, want %q", got, want)
	}
	if got, want := text.JoinDropRight(left, right, 14), left+strings.Repeat(" ", 5)+right; got != want {
		t.Fatalf("JoinDropRight = %q, want %q", got, want)
	}
	// A right side that must give way still gives way whole, and the left is only
	// clipped when it alone does not fit.
	if got, want := text.JoinDropRight(left, right, 8), left; got != want {
		t.Fatalf("JoinDropRight at 8 cells = %q, want %q", got, want)
	}
	if got, want := text.JoinDropRight(left, right, 3), "le"+text.Ellipsis; got != want {
		t.Fatalf("JoinDropRight at 3 cells = %q, want the honest cut %q", got, want)
	}
}

// TestJoinKeepsItsExactFitOutput pins why only JoinDropRight gained the gutter:
// Join's bytes must not move in v0.4.0, so an exact fit still closes the line with
// the two sides touching. A caller that wants air passes it inside the right side.
func TestJoinKeepsItsExactFitOutput(t *testing.T) {
	if got, want := text.Join("left", "right", 9), "leftright"; got != want {
		t.Fatalf("Join at an exact fit = %q, want its unchanged %q", got, want)
	}
	if got, want := text.Join("left", "right", 14), "left"+strings.Repeat(" ", 5)+"right"; got != want {
		t.Fatalf("Join = %q, want %q", got, want)
	}
}
