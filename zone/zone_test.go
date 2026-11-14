package zone_test

import (
	"testing"

	"github.com/WilberC/tui-kit/zone"
)

// TestHitReturnsTheTopmostZoneAtTheCell pins the question a product asks: which
// element did a pane cell hit? The element drawn last wins where two overlap,
// because the last one drawn is the one the user sees.
func TestHitReturnsTheTopmostZoneAtTheCell(t *testing.T) {
	zones := zone.New()
	zones.Add("row:0", 0, 0, 10, 1)
	zones.Add("chip:enter", 2, 0, 7, 1)

	cases := []struct {
		x, y int
		id   string
		ok   bool
	}{
		{0, 0, "row:0", true},      // before the chip: the row
		{1, 0, "row:0", true},      // the row's last cell before the chip
		{2, 0, "chip:enter", true}, // the chip was drawn after the row
		{8, 0, "chip:enter", true}, // the chip's last cell
		{9, 0, "row:0", true},      // after the chip: the row again
		{10, 0, "", false},         // one cell past the row
		{0, 1, "", false},          // one row below it
		{-1, 0, "", false},         // left of the pane
	}
	for _, c := range cases {
		id, ok := zones.Hit(c.x, c.y)
		if id != c.id || ok != c.ok {
			t.Fatalf("Hit(%d, %d) = %q/%v, want %q/%v", c.x, c.y, id, ok, c.id, c.ok)
		}
	}
}

// TestARectCoversItsOwnCellsAndNothingElse pins the half-open edges a hit test
// lives on: the cell at x+w is already the next element.
func TestARectCoversItsOwnCellsAndNothingElse(t *testing.T) {
	rect := zone.Rect{X: 2, Y: 3, W: 4, H: 2}

	for _, inside := range []struct{ x, y int }{{2, 3}, {5, 3}, {2, 4}, {5, 4}} {
		if !rect.Contains(inside.x, inside.y) {
			t.Fatalf("%v does not contain its own cell (%d, %d)", rect, inside.x, inside.y)
		}
	}
	for _, outside := range []struct{ x, y int }{{1, 3}, {6, 3}, {2, 2}, {2, 5}} {
		if rect.Contains(outside.x, outside.y) {
			t.Fatalf("%v contains (%d, %d), which is outside it", rect, outside.x, outside.y)
		}
	}
	if got, want := rect.Right(), 6; got != want {
		t.Fatalf("Right = %d, want %d", got, want)
	}
	if got, want := rect.Bottom(), 5; got != want {
		t.Fatalf("Bottom = %d, want %d", got, want)
	}
}

// TestAZoneWithNoAreaIsNeverRegistered pins the degenerate cases a clipped row
// produces: there is nothing to click, so nothing is stored.
func TestAZoneWithNoAreaIsNeverRegistered(t *testing.T) {
	zones := zone.New()
	zones.Add("empty-width", 0, 0, 0, 1)
	zones.Add("empty-height", 0, 0, 5, 0)
	zones.Add("negative", 0, 0, -3, 1)

	if got := zones.Len(); got != 0 {
		t.Fatalf("the map holds %d zones, want none: %v", got, zones.Zones())
	}
	if _, ok := zones.Hit(0, 0); ok {
		t.Fatalf("a hit landed on a zone with no area")
	}
}

// TestANilMapAnswersEverythingWithoutPanicking pins why every kit helper takes a
// map it may not need: a screen with no mouse support passes nil.
func TestANilMapAnswersEverythingWithoutPanicking(t *testing.T) {
	var zones *zone.Map

	zones.Add("row:0", 0, 0, 10, 1)
	zones.AddRect("row:1", zone.Rect{X: 0, Y: 1, W: 10, H: 1})
	zones.Offset(3, 4)
	zones.Clip(zone.Rect{X: 0, Y: 0, W: 1, H: 1})
	zones.Append(zone.New())
	zones.Reset()

	if id, ok := zones.Hit(0, 0); id != "" || ok {
		t.Fatalf("a nil map answered %q/%v", id, ok)
	}
	if _, ok := zones.Lookup(0, 0); ok {
		t.Fatalf("a nil map looked a zone up")
	}
	if _, ok := zones.Find("row:0"); ok {
		t.Fatalf("a nil map found a zone")
	}
	if got := zones.Len(); got != 0 {
		t.Fatalf("a nil map holds %d zones", got)
	}
	if got := zones.Zones(); len(got) != 0 {
		t.Fatalf("a nil map returned %d zones", len(got))
	}
}

// TestOffsetMovesTheWholeMap pins the translation the frame applies once it knows
// where its content landed in the pane.
func TestOffsetMovesTheWholeMap(t *testing.T) {
	zones := zone.New()
	zones.Add("row:0", 0, 0, 10, 1)
	zones.Add("chip:enter", 2, 0, 5, 1)

	zones.Offset(7, 3)

	if id, ok := zones.Hit(7, 3); !ok || id != "row:0" {
		t.Fatalf("after Offset(7, 3), Hit(7, 3) = %q/%v, want the row", id, ok)
	}
	if id, ok := zones.Hit(9, 3); !ok || id != "chip:enter" {
		t.Fatalf("the chip did not move with the map: %q/%v", id, ok)
	}
	if _, ok := zones.Hit(0, 0); ok {
		t.Fatalf("the map answered its old cell after the offset")
	}
	rect, ok := zones.Find("row:0")
	if !ok || rect.X != 7 || rect.Y != 3 {
		t.Fatalf("Find after Offset = %v/%v, want x 7 y 3", rect, ok)
	}
}

// TestClipKeepsWhatTheFrameShows pins the guarantee that makes hit testing safe:
// a zone half outside the surface keeps only the part the user can see, and a
// zone the surface never draws disappears instead of answering a click.
func TestClipKeepsWhatTheFrameShows(t *testing.T) {
	zones := zone.New()
	zones.Add("inside", 1, 1, 3, 1)
	zones.Add("half", 8, 1, 6, 1)
	zones.Add("gone", 20, 1, 4, 1)
	zones.Add("tall", 0, 4, 4, 10)

	zones.Clip(zone.Rect{X: 0, Y: 0, W: 10, H: 6})

	if rect, ok := zones.Find("inside"); !ok || rect.W != 3 {
		t.Fatalf("an inside zone was touched: %v/%v", rect, ok)
	}
	half, ok := zones.Find("half")
	if !ok || half.W != 2 || half.Right() != 10 {
		t.Fatalf("the half-visible zone was not cut at the texture's edge: %v/%v", half, ok)
	}
	if _, ok := zones.Find("gone"); ok {
		t.Fatalf("a zone outside the surface survived the clip")
	}
	tall, ok := zones.Find("tall")
	if !ok || tall.H != 2 || tall.Bottom() != 6 {
		t.Fatalf("the tall zone was not cut at the bottom edge: %v/%v", tall, ok)
	}
	if id, ok := zones.Hit(9, 1); !ok || id != "half" {
		t.Fatalf("the visible half of the cut zone does not answer: %q/%v", id, ok)
	}
}

// TestZonesKeepsTheDrawOrder pins the parts a caller inspects: the order is the
// order the surface drew them in, which is what makes Hit's last-wins rule
// predictable.
func TestZonesKeepsTheDrawOrder(t *testing.T) {
	other := zone.New()
	other.Add("third", 0, 2, 1, 1)

	zones := zone.New()
	zones.Add("first", 0, 0, 1, 1)
	zones.Add("second", 0, 1, 1, 1)
	zones.Append(other)

	ids := make([]string, 0, 3)
	for _, z := range zones.Zones() {
		ids = append(ids, z.ID)
		if z.Rect.Empty() {
			t.Fatalf("zone %q lost its rect", z.ID)
		}
	}
	if len(ids) != 3 || ids[0] != "first" || ids[1] != "second" || ids[2] != "third" {
		t.Fatalf("Zones() = %v, want the draw order", ids)
	}

	zones.Reset()
	if got := zones.Len(); got != 0 {
		t.Fatalf("Reset left %d zones", got)
	}
}
