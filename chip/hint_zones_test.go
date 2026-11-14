package chip_test

import (
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
	"github.com/WilberC/tui-kit/zone"
)

// TestHintBarRecordsTheChipOfEveryHint pins S8 on the bar: every slab the bar
// draws is a zone at the x the user sees, the paging chip is the well-known
// PageZone, and no zone covers anything but a slab.
func TestHintBarRecordsTheChipOfEveryHint(t *testing.T) {
	hints := hintBar()
	const width, x, y = 60, 7, 3

	var bar chip.HintBar
	zones := zone.New()
	line := bar.RenderZones(zones, x, y, width, hints...)
	cells := []rune(text.Plain(line))
	plain := string(cells)

	if pages := bar.Pages(width, hints...); pages < 2 {
		t.Fatalf("Pages(%d) = %d, want a paged bar", width, pages)
	}

	drawn := 0
	for _, hint := range hints {
		slab := " " + hint.Key + " "
		if !strings.Contains(plain, slab) {
			continue // this hint is on another page
		}
		drawn++
		rect, ok := zones.Find(chip.HintZone(hint.Key))
		if !ok {
			t.Fatalf("the bar did not record %q: %v", chip.HintZone(hint.Key), zones.Zones())
		}
		if rect.Y != y {
			t.Fatalf("%q was recorded on row %d, want %d", hint.Key, rect.Y, y)
		}
		if got, want := rect.W, text.Width(slab); got != want {
			t.Fatalf("%q's slab is %d cells wide, want %d", hint.Key, got, want)
		}
		if rect.X < x || rect.Right() > x+width {
			t.Fatalf("%q's zone %v escapes the bar it was drawn in (%d..%d)", hint.Key, rect, x, x+width)
		}
		if got := string(cells[rect.X-x : rect.Right()-x]); got != slab {
			t.Fatalf("%q's zone covers %q, want the slab %q", hint.Key, got, slab)
		}
	}
	if drawn == 0 {
		t.Fatalf("the first page draws no hint at all: %q", plain)
	}

	page, ok := zones.Find(chip.PageZone)
	if !ok {
		t.Fatalf("the paging chip has no zone: %v", zones.Zones())
	}
	if got, want := string(cells[page.X-x:page.Right()-x]), " "+chip.PageKey+" "; got != want {
		t.Fatalf("the paging zone covers %q, want the slab %q", got, want)
	}

	for _, z := range zones.Zones() {
		if got := string(cells[z.Rect.X-x : z.Rect.Right()-x]); !strings.HasPrefix(got, " ") || !strings.HasSuffix(got, " ") {
			t.Fatalf("zone %q covers %q, which is not a chip slab", z.ID, got)
		}
	}
}

// TestAClickOnThePagingChipPagesTheBar pins the wiring a product writes: the
// paging chip's zone id is chip.PageZone, so a click on it is the same action as
// chip.PageKey, and the zone follows the page the bar shows.
func TestAClickOnThePagingChipPagesTheBar(t *testing.T) {
	hints := hintBar()
	const width = 60

	var bar chip.HintBar
	zones := zone.New()
	if !strings.Contains(text.Plain(bar.RenderZones(zones, 0, 0, width, hints...)), " "+chip.PageKey+" ") {
		t.Fatalf("the first page does not end with the paging chip")
	}
	page, ok := zones.Find(chip.PageZone)
	if !ok {
		t.Fatalf("the paging chip has no zone")
	}
	if id, ok := zones.Hit(page.X+2, page.Y); !ok || id != chip.PageZone {
		t.Fatalf("a click in the paging chip = %q/%v, want %q", id, ok, chip.PageZone)
	}
	if id, ok := zones.Hit(page.Right(), page.Y); ok {
		t.Fatalf("a click one cell past the paging chip answered %q", id)
	}

	first := bar.Render(width, hints...)
	bar.Next(width, hints...)
	if bar.Page != 1 {
		t.Fatalf("after the paging action Page = %d, want 1", bar.Page)
	}
	if bar.Render(width, hints...) == first {
		t.Fatalf("the bar did not page")
	}

	zones.Reset()
	bar.RenderZones(zones, 0, 0, width, hints...)
	next, ok := zones.Find(chip.PageZone)
	if !ok {
		t.Fatalf("the second page has no paging zone")
	}
	if id, ok := zones.Hit(next.X+2, next.Y); !ok || id != chip.PageZone {
		t.Fatalf("a click on the second page's paging chip = %q/%v", id, ok)
	}
}

// TestAHintBarThatFitsRecordsOnlyItsHints pins the other half of the paging
// rule: a bar with no paging chip has no paging zone, so a product's click
// handler cannot fire on a chip that is not drawn.
func TestAHintBarThatFitsRecordsOnlyItsHints(t *testing.T) {
	hints := hintBar()[:2]

	var bar chip.HintBar
	zones := zone.New()
	bar.RenderZones(zones, 0, 0, 200, hints...)

	if _, ok := zones.Find(chip.PageZone); ok {
		t.Fatalf("a bar that fits recorded a paging zone: %v", zones.Zones())
	}
	if got, want := zones.Len(), len(hints); got != want {
		t.Fatalf("the bar recorded %d zones, want one per hint (%d)", got, want)
	}
}

// TestHintBarStartsANewHintSetOnItsFirstPage pins the follow-up the products
// asked for: the bar notices that the hint set changed and starts over, so a
// screen drops its own reset code — and Page still names a page when a product
// wants one.
func TestHintBarStartsANewHintSetOnItsFirstPage(t *testing.T) {
	hints := hintBar()
	const width = 60

	var bar chip.HintBar
	bar.Next(width, hints...)
	if bar.Page != 1 {
		t.Fatalf("after Next Page = %d, want 1", bar.Page)
	}
	if page := text.Plain(bar.Render(width, hints...)); !strings.Contains(page, "more 2/") {
		t.Fatalf("the same set moved the bar: %q", page)
	}
	if bar.Page != 1 {
		t.Fatalf("rendering the same set moved the bar to page %d", bar.Page)
	}

	changed := append(append([]chip.Hint{}, hints...), chip.Hint{Key: "ctrl+r", Label: "refresh"})
	bar.Render(width, changed...)
	if bar.Page != 0 {
		t.Fatalf("a new hint set left the bar on page %d, want 0", bar.Page)
	}
	if page := text.Plain(bar.Render(width, changed...)); !strings.Contains(page, "more 1/") {
		t.Fatalf("the new set did not render its first page: %q", page)
	}

	explicit := chip.HintBar{Page: 1}
	if page := text.Plain(explicit.Render(width, changed...)); !strings.Contains(page, "more 2/") {
		t.Fatalf("the bar ignored the page it was given: %q", page)
	}
	if explicit.Page != 1 {
		t.Fatalf("the bar's first set reset an explicit Page to %d", explicit.Page)
	}
}

// TestAHintSlabKeepsAConfirmationGold pins the per-hint style: a confirmation's
// y keeps the gold slab of its own role while its neighbours keep the default,
// and the style costs no cells, so paging is unaffected.
func TestAHintSlabKeepsAConfirmationGold(t *testing.T) {
	truecolor(t)
	gold := theme.ChipGold
	hints := []chip.Hint{
		{Key: "y", Label: "yes", Slab: &gold},
		{Key: "n", Label: "no"},
	}

	var bar chip.HintBar
	line := bar.Render(80, hints...)
	want := chip.Pair(theme.ChipGold, "y", "yes") + chip.Gutter + chip.Pair(theme.ChipText, "n", "no")

	if line != want {
		t.Fatalf("a gold hint rendered %q, want %q", text.Plain(line), text.Plain(want))
	}
	if !strings.Contains(line, sgr(theme.ChipGold)) {
		t.Fatalf("the confirmation lost its gold slab: %q", line)
	}
	if !strings.Contains(line, sgr(theme.ChipText)) {
		t.Fatalf("the neighbouring hint lost the default slab: %q", line)
	}

	plainHints := []chip.Hint{{Key: "y", Label: "yes"}, {Key: "n", Label: "no"}}
	if got, want := bar.Pages(80, hints...), bar.Pages(80, plainHints...); got != want {
		t.Fatalf("a slab changed the page count: %d, want %d", got, want)
	}
}
