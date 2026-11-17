package chip_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
	"github.com/WilberC/tui-kit/zone"
)

// TestThePagingChipReadsMoreByDefault pins the default: a bar with no
// PagingLabel renders exactly the paging chip it always rendered, `more n/m`,
// byte for byte — spelled out or left empty it is the same line.
func TestThePagingChipReadsMoreByDefault(t *testing.T) {
	hints := hintBar()
	const width = 40

	var bar chip.HintBar
	pages := bar.Pages(width, hints...)
	if pages < 2 {
		t.Fatalf("Pages(%d) = %d, want a paged bar", width, pages)
	}
	for page := 0; page < pages; page++ {
		bar.Page = page
		plain := text.Plain(bar.Render(width, hints...))
		want := " tab  more " + strconv.Itoa(page+1) + "/" + strconv.Itoa(pages)
		if !strings.HasSuffix(plain, want) {
			t.Fatalf("page %d does not end with %q: %q", page, want, plain)
		}
	}

	spelled := chip.HintBar{PagingLabel: "more"}
	for page := 0; page < pages; page++ {
		bar.Page, spelled.Page = page, page
		if got := bar.Render(width, hints...); got != spelled.Render(width, hints...) {
			t.Fatalf("page %d: the empty label renders %q, want the byte-identical spelled %q",
				page, text.Plain(got), text.Plain(spelled.Render(width, hints...)))
		}
	}
}

// TestThePagingLabelIsUsedWhenDrawing pins the option: a bar carries the word
// its product asked for — `más` for the dock — in every page, and the tab slab
// stays the same chip it always was.
func TestThePagingLabelIsUsedWhenDrawing(t *testing.T) {
	truecolor(t)
	hints := hintBar()
	const width = 40

	bar := chip.HintBar{PagingLabel: "más"}
	pages := bar.Pages(width, hints...)
	if pages < 2 {
		t.Fatalf("Pages(%d) = %d, want a paged bar", width, pages)
	}
	if strings.Contains(text.Plain(bar.Render(width, hints...)), " more ") {
		t.Fatalf("the bar kept the default word: %q", text.Plain(bar.Render(width, hints...)))
	}
	for page := 0; page < pages; page++ {
		bar.Page = page
		line := bar.Render(width, hints...)
		want := " tab  más " + strconv.Itoa(page+1) + "/" + strconv.Itoa(pages)
		if plain := text.Plain(line); !strings.HasSuffix(plain, want) {
			t.Fatalf("page %d does not end with %q: %q", page, want, plain)
		}
		if !strings.Contains(line, sgr(theme.ChipText)) {
			t.Fatalf("page %d lost the tab slab: %q", page, line)
		}
	}
}

// TestThePagingLabelStaysInTheWidthBudget pins the promise the paging chip
// makes: with any label, at any width, no line exceeds the bar's width — and a
// page wide enough for the whole chip shows the position complete, never cut.
func TestThePagingLabelStaysInTheWidthBudget(t *testing.T) {
	hints := hintBar()
	bar := chip.HintBar{PagingLabel: "más aún"}
	const word = "más aún"
	// Page 0 always holds the first hint; when the width admits the smallest
	// line a paged bar can draw — the first hint's pair, the gutter and the
	// whole paging chip — the chip must be budgeted in, never cut.
	firstPair := text.Width(chip.Pair(theme.ChipText, hints[0].Key, hints[0].Label))

	fits := 0
	for width := 10; width <= 80; width++ {
		pages := bar.Pages(width, hints...)
		for page := 0; page < pages; page++ {
			bar.Page = page
			line := bar.Render(width, hints...)
			if got := text.Width(line); got > width {
				t.Fatalf("width %d page %d: the line is %d cells wide: %q",
					width, page, got, text.Plain(line))
			}
			position := strconv.Itoa(page+1) + "/" + strconv.Itoa(pages)
			chipWidth := text.Width(chip.Pair(theme.ChipText, chip.PageKey, word+" "+position))
			if page == 0 && width >= firstPair+text.Width(chip.Gutter)+chipWidth {
				want := " " + chip.PageKey + "  " + word + " " + position
				if plain := text.Plain(line); !strings.HasSuffix(plain, want) {
					t.Fatalf("width %d: the paging chip was cut instead of budgeted: %q",
						width, plain)
				}
				fits++
			}
		}
	}
	if fits == 0 {
		t.Fatal("no width showed the whole paging chip, so the test checks nothing")
	}
}

// TestThePagingLabelKeepsThePageZone pins that a localised word changes
// nothing about where a click lands: the paging chip's slab is still recorded
// under chip.PageZone, covering exactly the ` tab ` cells.
func TestThePagingLabelKeepsThePageZone(t *testing.T) {
	hints := hintBar()
	const width = 40

	bar := chip.HintBar{PagingLabel: "más"}
	zones := zone.New()
	cells := []rune(text.Plain(bar.RenderZones(zones, 0, 0, width, hints...)))

	page, ok := zones.Find(chip.PageZone)
	if !ok {
		t.Fatalf("the labelled bar has no paging zone: %v", zones.Zones())
	}
	if got := string(cells[page.X:page.Right()]); got != " "+chip.PageKey+" " {
		t.Fatalf("the paging zone covers %q, want the tab slab %q", got, " "+chip.PageKey+" ")
	}
	if id, ok := zones.Hit(page.X+2, page.Y); !ok || id != chip.PageZone {
		t.Fatalf("a click inside the labelled paging chip = %q/%v, want %q", id, ok, chip.PageZone)
	}
}
