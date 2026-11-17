// Package chip is the token surface of the interface: one slab per key, mode or
// short state word, the columns those slabs align in, and the hint bar built
// from them.
//
// A chip is a shape, not a bracketed word: the background carries one cell of
// padding on each side of the label, and a label always follows its chip after
// exactly one plain space. The hint bar pages when the hints do not fit, so a
// hint is never hidden without the paging chip that says more exist.
package chip

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
	"github.com/WilberC/tui-kit/zone"
)

// Gutter is what separates two chip pairs in a legend or a hint bar: two plain
// spaces, so no separator ever touches a chip's background.
const Gutter = "  "

// Chip renders one token on the chip surface: the slab carries one cell of
// padding on each side, so a chip is a shape rather than a bracketed word.
func Chip(role lipgloss.Style, label string) string {
	return role.Render(" " + label + " ")
}

// Pair is one chip and the dim label that explains it: the slab, then exactly one
// plain space, then the label in the muted role.
func Pair(slab lipgloss.Style, key, label string) string {
	return Chip(slab, key) + " " + theme.Muted.Render(label)
}

// Legend joins already rendered chip pairs with two plain spaces. It never
// inserts a glyph between them, so the gutter stays colourless.
func Legend(items ...string) string {
	return join(items)
}

// Width is the cell width of the widest label: the chip column a group of rows
// shares, so every description starts on the same cell.
func Width(labels ...string) int {
	width := 0
	for _, label := range labels {
		width = max(width, text.Width(label))
	}
	return width
}

// Cell renders one chip cell of a fixed column: the label is padded to width
// cells inside the slab, and one plain space follows the slab, so the
// description that comes next keeps its x on every row.
func Cell(slab lipgloss.Style, label string, width int) string {
	return slab.Render(" "+text.Fit(label, width)+" ") + " "
}

// UnsetCell renders the cell of an entry that declares no key of its own: no
// slab at all, a dim glyph padded into the same column, so a group of such
// entries never reads as a wall of empty boxes.
func UnsetCell(role lipgloss.Style, label string, width int) string {
	return role.Render(text.Fit(" "+label, width+3))
}

// Hint is one key/label pair of a hint bar.
type Hint struct {
	Key   string
	Label string
	// Slab is the chip surface the key renders on. A nil Slab is the bar's own
	// surface, theme.ChipText; it is set for the one hint that must keep a meaning
	// of its own — the gold y of a confirmation — so such a hint never repaints
	// its neighbours.
	Slab *lipgloss.Style
}

// slab is the chip surface the hint's key renders on.
func (h Hint) slab() lipgloss.Style {
	if h.Slab == nil {
		return theme.ChipText
	}
	return *h.Slab
}

// pair is the hint as the bar draws it: the slab, one plain space, the label.
func (h Hint) pair() string { return Pair(h.slab(), h.Key, h.Label) }

// PageKey is the key that moves a hint bar to its next page. It is a named
// constant so every product displays the same key, and it is a non-printable
// one, so it never collides with the filter a key bar's screen reads. The kit
// renders it; the product wires the handler.
const PageKey = "tab"

// PageZone is the well-known zone id of the paging chip. A click on it is the
// same action as PageKey: the product looks the id up and moves the bar on.
const PageZone = "chip:page"

// HintZone is the zone id of one hint's chip in the hint bar, so a product that
// lets a hint be clicked looks it up by its key.
func HintZone(key string) string { return "chip:" + key }

// maxPagingRounds bounds the search for a stable page count: the paging chip's
// label grows with the number of pages, so the reserve settles by iteration.
const maxPagingRounds = 16

// HintBar is a hint bar that pages instead of hiding what does not fit. It holds
// the page the bar shows; the hints themselves are passed at call time, so the
// bar costs nothing until a line is drawn. The zero value shows page 0.
type HintBar struct {
	// Page is the page to render, counted from zero. A page past the end folds
	// back into range, so a caller that keeps its own counter never renders an
	// empty bar.
	Page int

	// PagingLabel names the word of the paging chip, so a product may localise
	// it — the dock says `más`. The position stays `n/m` after whatever word
	// the bar is given. An empty PagingLabel is the default word, "more".
	PagingLabel string

	// seen is the identity of the hint set the bar last drew, and known says
	// whether it has drawn one at all: the first set keeps the page a caller asked
	// for (a HintBar{Page: 1}), and a set that differs from the one before starts
	// over on page 0.
	seen  string
	known bool
}

// sync records the hint set the bar is drawing and starts a new set on its first
// page. A screen whose hints grow or shrink (a confirmation gains its y, a key
// disappears) therefore needs no reset of its own: the bar notices it.
func (b *HintBar) sync(hints []Hint) {
	key := setKey(hints)
	if b.known && key == b.seen {
		return
	}
	if b.known {
		b.Page = 0
	}
	b.seen, b.known = key, true
}

// setKey names a set of hints, so the bar compares the set and not the width it
// happens to be drawn at or the page it is on.
func setKey(hints []Hint) string {
	var key strings.Builder
	for _, hint := range hints {
		key.WriteString(hint.Key)
		key.WriteByte(0)
		key.WriteString(hint.Label)
		key.WriteByte(0)
	}
	return key.String()
}

// Render draws the page the bar shows: as many hints as fit, whole and in
// declaration order (the order is the priority). When the hints do not fit,
// every page ends with the paging chip — the PageKey slab and the muted
// position, `<PagingLabel> n/m` ("more 1/2" by default) — and that chip is
// part of the width budget, so a line
// never exceeds width. When they all fit, the bar is one line with no paging
// chip, exactly as Hints always rendered it. A hint wider than the bar is
// clipped with the ellipsis; a hint is never dropped without the chip.
func (b *HintBar) Render(width int, hints ...Hint) string {
	return b.RenderZones(nil, 0, 0, width, hints...)
}

// RenderZones is Render with the bar recorded in a zone map: each hint's slab
// under HintZone(key) and the paging chip's slab under PageZone, at (x, y) in the
// map's coordinates. A product answers a click from the cell alone, without
// measuring the bar, and a click on the paging chip is the same action as
// PageKey.
func (b *HintBar) RenderZones(zones *zone.Map, x, y, width int, hints ...Hint) string {
	if b == nil {
		return ""
	}
	b.sync(hints)
	if width <= 0 || len(hints) == 0 {
		return ""
	}

	drawn := b.layout(width, hints)
	for _, chip := range drawn.chips {
		zones.Add(chip.id, x+chip.x, y, min(chip.w, width-chip.x), 1)
	}
	return text.Clip(drawn.line, width)
}

// bar is the line the bar draws and the chips it drew in it, so a render can
// record the cells a click can land on.
type bar struct {
	line  string
	chips []drawnChip
}

// drawnChip is one chip of the bar: what a click there means, and where the bar
// drew it.
type drawnChip struct {
	id   string
	x, w int
}

// layout builds the page the bar shows, chip by chip, so the zones and the line
// cannot disagree about an x.
func (b *HintBar) layout(width int, hints []Hint) bar {
	rendered := pairs(hints)
	pages := paginate(width, b.pagingWord(), rendered)
	page := pageIndex(b.Page, len(pages))

	out := bar{}
	for position, index := range pages[page] {
		if position > 0 {
			out.line += Gutter
		}
		slab := Chip(hints[index].slab(), hints[index].Key)
		out.chips = append(out.chips, drawnChip{id: HintZone(hints[index].Key), x: text.Width(out.line), w: text.Width(slab)})
		out.line += rendered[index]
	}
	if len(pages) > 1 {
		if len(pages[page]) > 0 {
			out.line += Gutter
		}
		out.chips = append(out.chips, drawnChip{
			id: PageZone,
			x:  text.Width(out.line),
			w:  text.Width(Chip(theme.ChipText, PageKey)),
		})
		out.line += Pair(theme.ChipText, PageKey, paging(b.pagingWord(), page, len(pages)))
	}
	return out
}

// pairs renders every hint of the set, so paging and drawing measure the same
// strings.
func pairs(hints []Hint) []string {
	out := make([]string, len(hints))
	for index, hint := range hints {
		out[index] = hint.pair()
	}
	return out
}

// Pages is how many pages the bar needs at this width: one when everything fits,
// and always at least one, so a caller may use it as a modulus.
func (b *HintBar) Pages(width int, hints ...Hint) int {
	if width <= 0 || len(hints) == 0 {
		return 1
	}
	return len(paginate(width, b.pagingWord(), pairs(hints)))
}

// Next moves the bar to its next page and wraps from the last page back to page
// 0. A bar that fits stays on page 0: there is nothing to page to.
func (b *HintBar) Next(width int, hints ...Hint) {
	if b == nil {
		return
	}
	b.sync(hints)
	pages := b.Pages(width, hints...)
	b.Page = (pageIndex(b.Page, pages) + 1) % pages
}

// Hints renders the first page of a hint bar. It is HintBar{}.Render, kept so
// the callers that show one page keep compiling; a product that wires PageKey
// should hold a HintBar and move it with Next.
func Hints(width int, hints ...Hint) string {
	var bar HintBar
	return bar.Render(width, hints...)
}

// paginate splits rendered hints into consecutive pages. When the whole bar fits
// it is one page with no paging chip; otherwise each page reserves the room the
// chip needs, so the bar never takes a second row.
func paginate(width int, word string, items []string) [][]int {
	pages := split(width, items, 0)
	for round := 0; round < maxPagingRounds; round++ {
		if len(pages) == 1 && width <= pageWidth(pages[0], items) {
			return pages
		}
		next := split(width, items, reserve(len(pages), word))
		if len(next) == len(pages) {
			return next
		}
		pages = next
	}
	return pages
}

// split fills pages greedily, in order, keeping each hint whole: a hint joins the
// current page while the page plus the reserve still fits, and the page is at
// least one hint long, so an oversized hint occupies a page of its own and is
// clipped rather than looping.
func split(width int, items []string, reserve int) [][]int {
	pages := [][]int{}
	page := []int{}
	used := 0
	for index, item := range items {
		itemWidth := text.Width(item)
		if len(page) > 0 {
			itemWidth += text.Width(Gutter)
			if used+itemWidth+reserve > width {
				pages = append(pages, page)
				page = []int{index}
				used = text.Width(item)
				continue
			}
		}
		page = append(page, index)
		used += itemWidth
	}
	if len(page) > 0 {
		pages = append(pages, page)
	}
	return pages
}

// reserve is the room one paging chip takes away from a page: the gutter that
// separates it from the last hint plus the chip of the widest label it can carry.
func reserve(pages int, word string) int {
	if pages < 2 {
		return 0
	}
	return text.Width(Gutter) + text.Width(Pair(theme.ChipText, PageKey, paging(word, pages-1, pages)))
}

// paging is the paging chip's label: its word — the default is "more" — then
// the current page and the page count.
func paging(word string, page, pages int) string {
	return word + " " + strconv.Itoa(page+1) + "/" + strconv.Itoa(pages)
}

// pagingWord is the word the paging chip carries: the product's label, or the
// default "more" when none was set.
func (b *HintBar) pagingWord() string {
	if b.PagingLabel == "" {
		return defaultPagingWord
	}
	return b.PagingLabel
}

// defaultPagingWord is the label an empty PagingLabel renders: the word the
// bar has always shown.
const defaultPagingWord = "more"

// pageIndex folds a page into range, so a caller's own counter can never render
// an empty bar.
func pageIndex(page, pages int) int {
	if pages <= 1 {
		return 0
	}
	return ((page % pages) + pages) % pages
}

// indexes picks the rendered hints of one page, in order.
func indexes(page []int, items []string) []string {
	out := make([]string, len(page))
	for position, index := range page {
		out[position] = items[index]
	}
	return out
}

// pageWidth is the width of one page's hints with their gutters.
func pageWidth(page []int, items []string) int {
	width := 0
	for position, index := range page {
		if position > 0 {
			width += text.Width(Gutter)
		}
		width += text.Width(items[index])
	}
	return width
}

func join(items []string) string {
	var out string
	for _, item := range items {
		if item == "" {
			continue
		}
		if out != "" {
			out += Gutter
		}
		out += item
	}
	return out
}
