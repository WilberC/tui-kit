package list_test

import (
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/zone"
)

// TestAnEntryWithNoCellKeepsItsZone pins the row helper for a list with no chip
// column: an entry whose Cell is empty draws exactly the row list.Row draws — the
// tree guide and the caller's body — and still records the whole row as a click
// target, so a session row needs no hand-measured zone beside the kit's.
func TestAnEntryWithNoCellKeepsItsZone(t *testing.T) {
	truecolor(t)
	const width, x, y = 46, 3, 9
	roles := list.RowRolesFor(list.State{Selected: true})
	body := roles.Name.Render("api") + "  " + roles.Muted.Render("herdr · 3 tabs")

	entry := list.Entry{
		Roles: roles, Last: false, Cell: "", Description: body, Width: width,
		ID: "row:api", ChipID: "chip:api", Zones: zone.New(), X: x, Y: y,
	}
	row := entry.Render()

	if want := list.SelectedRow(list.Row(roles, false, body), width); row != want {
		t.Fatalf("a cell-less entry drew %q, want the plain row %q", text.Plain(row), text.Plain(want))
	}
	if !strings.HasPrefix(text.Plain(row), text.SelectMark) {
		t.Fatalf("the cell-less selected entry lost its accent edge: %q", text.Plain(row))
	}
	if got, want := text.Width(row), width; got != want {
		t.Fatalf("the cell-less entry is %d cells wide, want %d", got, want)
	}

	rowRect, ok := entry.Zones.Find("row:api")
	if !ok || rowRect != (zone.Rect{X: x, Y: y, W: width, H: 1}) {
		t.Fatalf("the row zone = %v/%v, want the whole row at (%d, %d)", rowRect, ok, x, y)
	}
	if _, ok := entry.Zones.Find("chip:api"); ok {
		t.Fatalf("a row with no chip column registered a chip zone")
	}
	if id, ok := entry.Zones.Hit(x+width-1, y); !ok || id != "row:api" {
		t.Fatalf("the row's last cell answered %q/%v, want the row", id, ok)
	}

	// The plain draw is the same row list.Row renders, so a product that moves a
	// session row onto Entry does not change a byte of it.
	plain := list.RowRoles(false)
	quietEntry := list.Entry{Roles: plain, Last: true, Description: "docs  2 tabs", Width: width}
	if want := plain.Pad(list.Row(plain, true, "docs  2 tabs"), width); quietEntry.Render() != want {
		t.Fatalf("a cell-less panel entry drew %q, want %q", text.Plain(quietEntry.Render()), text.Plain(want))
	}
}

// TestATrailingColumnEndsOnTheRowEdge pins the table row: the trailing column is
// flush with the row's last cell, exactly TrailingGap cells of the row's own
// background separate it from the description, and that gap is tinted on the
// hover and on the selection, so no tint breaks before the trailing column.
func TestATrailingColumnEndsOnTheRowEdge(t *testing.T) {
	truecolor(t)
	const width, chipWidth = 44, 8
	const description, trailing = "Command bar", "herdr-suite"

	for _, c := range []struct {
		name  string
		state list.State
	}{
		{name: "plain"},
		{name: "hovered", state: list.State{Hovered: true}},
		{name: "selected", state: list.State{Selected: true}},
	} {
		roles := list.RowRolesFor(c.state)
		painted := roles.Name.Render(description)
		source := roles.Muted.Render(trailing)
		row := list.Entry{
			Roles:       roles,
			Last:        true,
			Cell:        chip.Cell(roles.Chip, "prefix+k", chipWidth),
			Description: painted,
			Trailing:    source,
			Width:       width,
		}.Render()

		if got, want := text.Width(row), width; got != want {
			t.Fatalf("the %s row is %d cells wide, want %d: %q", c.name, got, want, text.Plain(row))
		}
		trailStart := cellOf(t, row, trailing)
		if got, want := trailStart+text.Width(trailing), width; got != want {
			t.Fatalf("the %s row's trailing column ends on cell %d, want the row's last cell %d", c.name, got, want)
		}
		// The whole span between the two columns is the row's own background, never
		// plain space: this is what keeps the tint unbroken up to the trailing column.
		gap := trailStart - cellOf(t, row, description) - text.Width(description)
		if gap < list.TrailingGap {
			t.Fatalf("the %s row leaves %d cells between its columns, want at least %d", c.name, gap, list.TrailingGap)
		}
		// The spare room and the gutter are one run of the row's own background, so
		// the row carries a single tinted span between its columns.
		if want := painted + roles.Gap(gap) + source; !strings.Contains(row, want) {
			t.Fatalf("the %s row does not carry the row's own background between its columns as one span: %q", c.name, row)
		}
	}

	// On the panel surface the gap is plain, because that surface has no tint.
	roles := list.RowRoles(false)
	plainRow := list.Entry{
		Roles: roles, Last: false, Cell: chip.Cell(roles.Chip, "prefix+k", 8),
		Description: description, Trailing: trailing, Width: width,
	}.Render()
	gutter := plainRow[strings.Index(plainRow, description)+len(description) : strings.Index(plainRow, trailing)]
	if len(gutter) < list.TrailingGap || strings.Trim(gutter, " ") != "" {
		t.Fatalf("the panel row's gutter is not plain background: %q", plainRow)
	}
}

// TestATrailingColumnGivesTheDescriptionWay pins the narrow row: the trailing
// column keeps the edge, the description is clipped with an honest ellipsis, and
// the gap survives; past the point where even the trailing column fits, the row
// keeps the trailing column and never grows past its width.
func TestATrailingColumnGivesTheDescriptionWay(t *testing.T) {
	truecolor(t)
	const trailing = "herdr-suite"
	roles := list.RowRoles(true)
	long := roles.Name.Render(strings.Repeat("a description ", 4))

	row := list.Entry{
		Roles: roles, Last: true, Cell: chip.Cell(roles.Chip, "prefix+k", 8),
		Description: long, Trailing: roles.Muted.Render(trailing), Width: 30,
	}.Render()
	if got, want := text.Width(row), 30; got != want {
		t.Fatalf("the narrow row is %d cells wide, want %d", got, want)
	}
	start := cellOf(t, row, trailing)
	if got, want := start+text.Width(trailing), 30; got != want {
		t.Fatalf("the trailing column ends on cell %d, want 30", got)
	}
	plain := []rune(text.Plain(row))
	if got, want := string(plain[start-1-list.TrailingGap]), text.Ellipsis; got != want {
		t.Fatalf("the description was not clipped with an ellipsis before the gap: %q", text.Plain(row))
	}

	// No room for the description: the row keeps its trailing column, clipped.
	tiny := list.Entry{
		Roles: roles, Last: true, Description: long,
		Trailing: roles.Muted.Render(trailing), Width: 8,
	}.Render()
	if got, want := text.Width(tiny), 8; got != want {
		t.Fatalf("a row too narrow for both columns is %d cells wide, want %d: %q", got, want, text.Plain(tiny))
	}
	if strings.Contains(text.Plain(tiny), "description") {
		t.Fatalf("a row too narrow for both kept the description: %q", text.Plain(tiny))
	}
}
