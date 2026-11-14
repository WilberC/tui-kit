package frame_test

import (
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/frame"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
	"github.com/WilberC/tui-kit/zone"
)

// runes is one rendered line as cells: every glyph the kit draws is one cell
// wide, so the rune at index i is the cell at pane x i.
func runes(t *testing.T, line string) []rune {
	t.Helper()
	plain := []rune(text.Plain(line))
	if width := text.Width(line); width != len(plain) {
		t.Fatalf("a %d-cell line is %d runes wide: %q", width, len(plain), text.Plain(line))
	}
	return plain
}

// zonedSurface is the shape a product renders: a header, a rule, a filter row,
// the opening column rule, a two-column body whose second column holds a chip
// row, the closing rule and a two-line footer. The map it returns is in
// frame-content coordinates: x from the first content cell, y from the first row
// under the top border.
func zonedSurface(t *testing.T, paneWidth, paneHeight int) (frame.Surface, []string, *zone.Map) {
	t.Helper()
	size := frame.Recommended
	f := frame.Frame{
		Title: theme.GoldBold.Render("demo"),
		Width: size.ClampWidth(paneWidth),
	}
	widths := frame.ColumnWidths(f.Width-2, 2)
	surface := frame.Surface{
		Frame:   f,
		Size:    size,
		Columns: widths,
		Top: []string{
			f.Row("HEADER"),
			f.Rule(),
			f.Row("FILTER"),
			frame.ColumnRule(widths, frame.TopTee),
		},
		Bottom: []string{
			frame.ColumnRule(widths, frame.BottomTee),
			f.Row("DETAIL"),
			f.Row("HINTS"),
		},
	}

	const column = 1
	zones := zone.New()
	row := list.Entry{
		Roles:       list.RowRoles(true),
		Last:        true,
		Cell:        chip.Cell(theme.ChipText, "enter", 5),
		Description: theme.Text.Render("run"),
		Width:       widths[column],
		ID:          "row:enter",
		ChipID:      "chip:enter",
		Zones:       zones,
		X:           surface.ColumnX()[column],
		Y:           surface.BodyY(0),
	}
	blocks := [][]string{{"left"}, {row.Render()}}
	body := frame.Grid(f, blocks, widths, surface.BodyRows(paneHeight))
	return surface, body, zones
}

// TestSurfaceAnswersTheCellItDrewARowIn is the promise of S8: the kit records
// where it drew each clickable element and answers which one a pane cell hit, in
// a centred, padded, columned surface at both captured window sizes. Nothing is
// left for the product to measure.
func TestSurfaceAnswersTheCellItDrewARowIn(t *testing.T) {
	for _, window := range []struct{ width, height int }{{180, 50}, {80, 24}} {
		surface, body, zones := zonedSurface(t, window.width, window.height)
		lines := strings.Split(surface.PopupZones(window.width, window.height, body, zones), "\n")

		frameWidth := frame.Recommended.ClampWidth(window.width)
		frameHeight := frame.Recommended.ClampHeight(window.height)
		left := max(0, (window.width-frameWidth)/2)
		top := max(0, (window.height-frameHeight)/2)

		// The row sits in the second column's content, on the body's first content
		// line: the centring, the border, the header rows and the body's top
		// padding row all come before it.
		columnX := left + 2 + surface.ColumnX()[1]
		rowY := top + 1 + surface.BodyY(0)
		cells := runes(t, lines[rowY])

		if got, want := string(cells[left]), frame.Vertical; got != want {
			t.Fatalf("%dx%d: the frame does not open at the centring offset %d: %q", window.width, window.height, left, got)
		}
		if got, want := string(cells[columnX]), text.SelectMark; got != want {
			t.Fatalf("%dx%d: the selected row does not start at cell %d: %q", window.width, window.height, columnX, string(cells[columnX:min(len(cells), columnX+8)]))
		}
		if got, want := string(cells[columnX+5:columnX+12]), " enter "; got != want {
			t.Fatalf("%dx%d: the chip is not at cell %d: %q", window.width, window.height, columnX+5, got)
		}

		cases := []struct {
			name string
			x, y int
			id   string
			ok   bool
		}{
			{"the row's first cell", columnX, rowY, "row:enter", true},
			{"the chip's first cell", columnX + 5, rowY, "chip:enter", true},
			{"the chip's last cell", columnX + 11, rowY, "chip:enter", true},
			{"the cell after the chip", columnX + 12, rowY, "row:enter", true},
			{"the padding row above the body", columnX, rowY - 1, "", false},
			{"below the frame", left, top + frameHeight, "", false},
			{"left of the frame", left - 1, rowY, "", false},
		}
		for _, c := range cases {
			id, ok := zones.Hit(c.x, c.y)
			if id != c.id || ok != c.ok {
				t.Fatalf("%dx%d: Hit(%d, %d) on %s = %q/%v, want %q/%v",
					window.width, window.height, c.x, c.y, c.name, id, ok, c.id, c.ok)
			}
		}
	}
}

// TestSurfaceKeepsEveryZoneInsideItsFrame pins the guarantee that makes the map
// trustworthy: a zone the surface did not draw is gone, and a zone it half drew
// keeps only the cells the user can see.
func TestSurfaceKeepsEveryZoneInsideItsFrame(t *testing.T) {
	const paneWidth, paneHeight = 180, 50
	surface, body, zones := zonedSurface(t, paneWidth, paneHeight)

	contentWidth := surface.Frame.ContentWidth()
	zones.Add("far", 400, 90, 10, 2)
	zones.Add("hanging", contentWidth-2, surface.BodyY(0), 12, 1)

	surface.PopupZones(paneWidth, paneHeight, body, zones)

	frameWidth := frame.Recommended.ClampWidth(paneWidth)
	frameHeight := frame.Recommended.ClampHeight(paneHeight)
	left := (paneWidth - frameWidth) / 2
	top := (paneHeight - frameHeight) / 2
	interior := frameHeight - 2

	for _, z := range zones.Zones() {
		if z.Rect.X < left+2 || z.Rect.Y < top+1 ||
			z.Rect.Right() > left+2+contentWidth || z.Rect.Bottom() > top+1+interior {
			t.Fatalf("zone %q %v is outside the frame's content (%d cells wide, %d rows tall) at (%d, %d)",
				z.ID, z.Rect, contentWidth, interior, left, top)
		}
	}
	if _, ok := zones.Find("far"); ok {
		t.Fatalf("a zone the surface never drew survived the render")
	}
	hanging, ok := zones.Find("hanging")
	if !ok || hanging.W != 2 || hanging.Right() != left+2+contentWidth {
		t.Fatalf("the hanging zone was not cut at the frame's edge: %v/%v", hanging, ok)
	}
	if id, ok := zones.Hit(left+2+contentWidth-1, top+1+surface.BodyY(0)); !ok || id != "hanging" {
		t.Fatalf("the visible part of the cut zone does not answer: %q/%v", id, ok)
	}
	if id, ok := zones.Hit(0, 0); ok {
		t.Fatalf("a cell outside the frame answered %q", id)
	}
}

// TestColumnXMatchesTheRenderedColumns pins the column offsets the kit records
// against the cells the grid actually draws: a column that reported the wrong x
// would hand the product a zone one pad cell off.
func TestColumnXMatchesTheRenderedColumns(t *testing.T) {
	for _, paneWidth := range []int{180, 140, 80} {
		surface, body, _ := zonedSurface(t, paneWidth, 30)
		// A grid row opens with the frame's border and one padding cell, so the
		// column content starts at index 2 of the line.
		cells := runes(t, body[0])[1+1:]
		offsets := surface.ColumnX()

		if got, want := len(offsets), len(surface.Columns); got != want {
			t.Fatalf("%d wide: ColumnX has %d entries, want %d", paneWidth, got, want)
		}
		if offsets[0] != 0 {
			t.Fatalf("%d wide: the first column starts at content x %d, want 0", paneWidth, offsets[0])
		}
		if !strings.HasPrefix(string(cells[offsets[0]:]), "left") {
			t.Fatalf("%d wide: the first column does not start at content x 0: %q", paneWidth, string(cells[:8]))
		}
		for index := 1; index < len(offsets); index++ {
			if got, want := string(cells[offsets[index]-2]), frame.Vertical; got != want {
				t.Fatalf("%d wide: column %d reports content x %d, but the separator is not two cells before it: %q",
					paneWidth, index, offsets[index], string(cells[max(0, offsets[index]-4):offsets[index]+4]))
			}
			if got, want := string(cells[offsets[index]-1]), " "; got != want {
				t.Fatalf("%d wide: column %d does not open with its padding cell: %q", paneWidth, index, got)
			}
		}
	}
}

// TestSurfaceRowFramesBodyRows pins the helper that stops an unframed row from
// slipping into a body: it frames at the clamped width, clips like any row, and
// leaves the surface's own frame untouched.
func TestSurfaceRowFramesBodyRows(t *testing.T) {
	surface, _, _ := zonedSurface(t, 140, 40)
	before := surface.Frame.Width

	row := surface.Row(180, "one body row")
	want := frame.Frame{Title: surface.Frame.Title, Width: frame.Recommended.ClampWidth(180)}.Row("one body row")
	if row != want {
		t.Fatalf("Row(180) = %q, want %q", text.Plain(row), text.Plain(want))
	}
	if got, want := text.Width(row), 160; got != want {
		t.Fatalf("Row(180) is %d cells wide, want the clamped %d", got, want)
	}
	if plain := text.Plain(row); !strings.HasPrefix(plain, frame.Vertical) || !strings.HasSuffix(plain, frame.Vertical) {
		t.Fatalf("a body row is not framed: %q", plain)
	}
	if got, want := surface.Frame.Width, before; got != want {
		t.Fatalf("Row rewrote the surface's frame width to %d, want %d", got, want)
	}

	rows := surface.Rows(80, "a", strings.Repeat("long ", 40))
	if len(rows) != 2 {
		t.Fatalf("Rows(80, two contents) returned %d rows", len(rows))
	}
	for index, framed := range rows {
		if got, want := text.Width(framed), surface.Width(80); got != want {
			t.Fatalf("row %d is %d cells wide, want %d", index, got, want)
		}
	}
	if !strings.Contains(rows[1], text.Ellipsis) {
		t.Fatalf("a long body row was not clipped: %q", text.Plain(rows[1]))
	}
}

// TestBodyYAndBottomYNameTheRowsTheSurfaceDraws pins the two coordinates a
// product registers a body row and a footer row with: they must land on the lines
// the surface renders, with the body's padding row and the footer's pinned rows
// already counted.
func TestBodyYAndBottomYNameTheRowsTheSurfaceDraws(t *testing.T) {
	const paneWidth, paneHeight = 140, 40
	surface, body, zones := zonedSurface(t, paneWidth, paneHeight)

	footerY := surface.BottomY(paneHeight, 0)
	zones.Add("footer", 0, footerY, 6, 1)
	lines := strings.Split(surface.RenderZones(paneWidth, paneHeight, body, zones), "\n")

	if got := rowContent(lines[surface.BodyY(0)+1]); !strings.Contains(got, "left") {
		t.Fatalf("BodyY(0) does not name the first body content row: %q", got)
	}
	if got := rowContent(lines[surface.BodyY(0)]); strings.ReplaceAll(got, frame.Vertical, "") != "" {
		t.Fatalf("the row above the first body content row is not the padding row: %q", got)
	}
	if got := text.Plain(lines[footerY+1]); !strings.Contains(got, frame.BottomTee) {
		t.Fatalf("BottomY(0) does not name the closing column rule: %q", got)
	}
	if id, ok := zones.Hit(2, footerY+1); !ok || id != "footer" {
		t.Fatalf("a footer zone does not answer without the centring offset: %q/%v", id, ok)
	}
}

// TestRenderWithoutZonesIsTheSameView keeps the call the products already make
// byte-identical: recording zones is something a product opts into.
func TestRenderWithoutZonesIsTheSameView(t *testing.T) {
	surface, body, _ := zonedSurface(t, 140, 40)

	want := surface.Render(140, 40, body)
	if got := surface.RenderZones(140, 40, body, nil); got != want {
		t.Fatalf("RenderZones with no map drew a different view")
	}
	if got := surface.Popup(140, 50, body); got != surface.PopupZones(140, 50, body, nil) {
		t.Fatalf("PopupZones with no map drew a different popup")
	}
}

// TestRowPassesThroughAnAlreadyFramedLine pins the guard that keeps a rule a
// rule: a whole line of this frame handed back to Row comes back unchanged, so no
// caller can frame it twice and lose its joints to an ellipsis.
func TestRowPassesThroughAnAlreadyFramedLine(t *testing.T) {
	f := frame.Frame{Title: "demo", Width: 40}

	for name, line := range map[string]string{
		"a rule":          f.Rule(),
		"the top border":  f.Top(),
		"the bottom edge": f.Bottom(),
		"an interior row": f.Row("content"),
	} {
		if got := f.Row(line); got != line {
			t.Fatalf("Row framed %s a second time: %q, want %q", name, text.Plain(got), text.Plain(line))
		}
	}

	// Content that is not a frame line is still framed, unchanged behaviour.
	if got, want := text.Width(f.Row("hello")), 40; got != want {
		t.Fatalf("a plain row is %d cells wide, want %d", got, want)
	}
	// A line of the wrong width is content, and is fitted as content.
	short := frame.Frame{Width: 20}.Rule()
	if got := text.Width(f.Row(short)); got != 40 {
		t.Fatalf("a rule of another width is %d cells wide, want the frame's 40", got)
	}
}

// TestARuleThroughRowsJoinsBothBorders pins the bug the demo shipped: a rule
// handed to Rows stayed a rule — width, joints — and never grew an ellipsis.
func TestARuleThroughRowsJoinsBothBorders(t *testing.T) {
	surface, _, _ := zonedSurface(t, 140, 30)
	f := surface.At(140)

	rows := surface.Rows(140, "HEADER", f.Rule(), "FILTER")
	if len(rows) != 3 {
		t.Fatalf("Rows returned %d rows, want 3", len(rows))
	}
	rule := text.Plain(rows[1])
	if !strings.HasPrefix(rule, frame.LeftTee) || !strings.HasSuffix(rule, frame.RightTee) {
		t.Fatalf("the rule does not join both borders: %q", rule)
	}
	if strings.Contains(rule, text.Ellipsis) {
		t.Fatalf("the rule was clipped: %q", rule)
	}
	if got, want := text.Width(rows[1]), f.Width; got != want {
		t.Fatalf("the rule is %d cells wide, want %d", got, want)
	}
	if strings.Contains(strings.Trim(rule, "├┤"), frame.Vertical) {
		t.Fatalf("the rule carries a vertical glyph inside it: %q", rule)
	}
	for _, row := range []string{rows[0], rows[2]} {
		if plain := text.Plain(row); !strings.HasPrefix(plain, frame.Vertical) || !strings.HasSuffix(plain, frame.Vertical) {
			t.Fatalf("a content row is not framed: %q", plain)
		}
	}
}

// TestTheBodyPaddingCarriesNoGuideGlyph pins the padding row: one blank body row
// at each end holds no tree guide and no chip, whatever the rows around it carry.
// A guide tick in the padding row reads as a row that starts one line early.
func TestTheBodyPaddingCarriesNoGuideGlyph(t *testing.T) {
	size := frame.Size{}
	f := frame.Frame{Title: "demo", Width: 60}
	surface := frame.Surface{
		Frame:  f,
		Size:   size,
		Top:    []string{f.Row("HEADER"), f.Rule()},
		Bottom: []string{f.Rule(), f.Row("DETAIL"), f.Row("HINTS")},
	}
	body := surface.Rows(60,
		list.Row(list.RowRoles(true), false, "the selected row carries the accent edge"),
		list.Row(list.RowRoles(false), true, "the last row closes its group"),
	)
	lines := strings.Split(surface.Render(60, 12, body), "\n")

	// The body's first content line carries the row, so the line above it is the
	// padding row; the bottom padding row is the line before the footer's rule.
	for _, index := range []int{surface.BodyY(0), surface.BodyTop() + len(body) + 1} {
		content := rowContent(lines[index])
		if content != "" {
			t.Fatalf("body padding row %d carries content: %q", index, content)
		}
		for _, glyph := range []string{"├", "└", "─", text.SelectMark} {
			if strings.Contains(text.Plain(lines[index]), glyph) {
				t.Fatalf("body padding row %d carries %q: %q", index, glyph, text.Plain(lines[index]))
			}
		}
	}
	if got := rowContent(lines[surface.BodyY(0)+1]); !strings.Contains(got, "selected row") {
		t.Fatalf("the first body row does not follow the padding row: %q", got)
	}
}
