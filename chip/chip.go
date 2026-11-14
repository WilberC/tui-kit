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

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
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
}

// PageKey is the key that moves a hint bar to its next page. It is a named
// constant so every product displays the same key, and it is a non-printable
// one, so it never collides with the filter a key bar's screen reads. The kit
// renders it; the product wires the handler.
const PageKey = "tab"

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
}

// Render draws the page the bar shows: as many hints as fit, whole and in
// declaration order (the order is the priority). When the hints do not fit,
// every page ends with the paging chip — the PageKey slab and the muted
// position, `more 1/2` — and that chip is part of the width budget, so a line
// never exceeds width. When they all fit, the bar is one line with no paging
// chip, exactly as Hints always rendered it. A hint wider than the bar is
// clipped with the ellipsis; a hint is never dropped without the chip.
func (b HintBar) Render(width int, hints ...Hint) string {
	if width <= 0 || len(hints) == 0 {
		return ""
	}

	items := make([]string, len(hints))
	for index, hint := range hints {
		items[index] = Pair(theme.ChipText, hint.Key, hint.Label)
	}

	pages := paginate(width, items)
	page := pageIndex(b.Page, len(pages))
	line := join(indexes(pages[page], items))
	if len(pages) > 1 {
		line = join([]string{line, Pair(theme.ChipText, PageKey, paging(page, len(pages)))})
	}
	return text.Clip(line, width)
}

// Pages is how many pages the bar needs at this width: one when everything fits,
// and always at least one, so a caller may use it as a modulus.
func (b HintBar) Pages(width int, hints ...Hint) int {
	if width <= 0 || len(hints) == 0 {
		return 1
	}

	items := make([]string, len(hints))
	for index, hint := range hints {
		items[index] = Pair(theme.ChipText, hint.Key, hint.Label)
	}
	return len(paginate(width, items))
}

// Next moves the bar to its next page and wraps from the last page back to page
// 0. A bar that fits stays on page 0: there is nothing to page to.
func (b *HintBar) Next(width int, hints ...Hint) {
	pages := b.Pages(width, hints...)
	b.Page = (pageIndex(b.Page, pages) + 1) % pages
}

// Hints renders the first page of a hint bar. It is HintBar{}.Render, kept so
// the callers that show one page keep compiling; a product that wires PageKey
// should hold a HintBar and move it with Next.
func Hints(width int, hints ...Hint) string { return HintBar{}.Render(width, hints...) }

// paginate splits rendered hints into consecutive pages. When the whole bar fits
// it is one page with no paging chip; otherwise each page reserves the room the
// chip needs, so the bar never takes a second row.
func paginate(width int, items []string) [][]int {
	pages := split(width, items, 0)
	for round := 0; round < maxPagingRounds; round++ {
		if len(pages) == 1 && width <= pageWidth(pages[0], items) {
			return pages
		}
		next := split(width, items, reserve(len(pages)))
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
func reserve(pages int) int {
	if pages < 2 {
		return 0
	}
	return text.Width(Gutter) + text.Width(Pair(theme.ChipText, PageKey, paging(pages-1, pages)))
}

// paging is the paging chip's label: the current page, then the page count.
func paging(page, pages int) string {
	return "more " + strconv.Itoa(page+1) + "/" + strconv.Itoa(pages)
}

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
