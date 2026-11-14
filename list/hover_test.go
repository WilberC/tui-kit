package list_test

import (
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
	"github.com/WilberC/tui-kit/zone"
)

// hoverRow is one chip row painted with the roles under test.
func hoverRow(roles list.Roles, width int) string {
	return list.ChipRow(roles, true, chip.Cell(roles.Chip, "prefix+k", 9), roles.Name.Render("Command bar"), width)
}

// TestHoverTintsFainterThanTheSelection pins the hover role of S8: a hovered row
// previews the selection in a fainter tint, and it is not the selection — no
// accent edge and no selection tint.
func TestHoverTintsFainterThanTheSelection(t *testing.T) {
	truecolor(t)
	const width = 40

	hovered := hoverRow(list.RowRolesFor(list.State{Hovered: true}), width)
	selected := hoverRow(list.RowRolesFor(list.State{Selected: true}), width)
	panel := hoverRow(list.RowRoles(false), width)

	if !strings.Contains(hovered, sgr(theme.HoverBG)) {
		t.Fatalf("a hovered row carries no hover tint: %q", hovered)
	}
	if strings.Contains(hovered, sgr(theme.SelectionBG)) {
		t.Fatalf("a hovered row wears the selection tint: %q", hovered)
	}
	if !strings.Contains(selected, sgr(theme.SelectionBG)) {
		t.Fatalf("the selected row lost its tint: %q", selected)
	}
	if strings.Contains(selected, sgr(theme.HoverBG)) {
		t.Fatalf("the selected row wears the hover tint: %q", selected)
	}
	if strings.HasPrefix(text.Plain(hovered), text.SelectMark) {
		t.Fatalf("a hovered row carries the selection's edge: %q", text.Plain(hovered))
	}
	if !strings.HasPrefix(text.Plain(selected), text.SelectMark) {
		t.Fatalf("the selected row lost its edge: %q", text.Plain(selected))
	}
	for name, row := range map[string]string{"hovered": hovered, "selected": selected, "panel": panel} {
		if got, want := text.Width(row), width; got != want {
			t.Fatalf("the %s row is %d cells wide, want %d", name, got, want)
		}
	}
	if strings.Contains(panel, sgr(theme.HoverBG)) || strings.Contains(panel, sgr(theme.SelectionBG)) {
		t.Fatalf("a panel row is tinted: %q", panel)
	}
}

// TestSelectionWinsOverHover pins the order of the two states: a row that is both
// hovered and selected reads as the selection, never as half of each.
func TestSelectionWinsOverHover(t *testing.T) {
	truecolor(t)

	both := list.RowRolesFor(list.State{Selected: true, Hovered: true})
	row := hoverRow(both, 40)

	if !strings.Contains(row, sgr(theme.SelectionBG)) {
		t.Fatalf("a row that is selected and hovered lost the selection tint: %q", row)
	}
	if strings.Contains(row, sgr(theme.HoverBG)) {
		t.Fatalf("a row that is selected and hovered wears the hover tint: %q", row)
	}
	if !strings.HasPrefix(text.Plain(row), text.SelectMark) {
		t.Fatalf("a row that is selected and hovered lost the selection's edge: %q", text.Plain(row))
	}
	if !strings.Contains(row, sgr(theme.TextOnSelection)) {
		t.Fatalf("a row that is selected and hovered lost the selection's name role: %q", row)
	}
}

// TestAnEntryRecordsItsRowAndItsChip pins the list side of S8: the row registers
// the whole row and the chip slab, at the coordinates the caller placed it, and
// it draws exactly the row ChipRow draws.
func TestAnEntryRecordsItsRowAndTheChip(t *testing.T) {
	const width, x, y = 40, 12, 7
	roles := list.RowRoles(false)
	cell := chip.Cell(theme.ChipText, "enter", 6)

	zones := zone.New()
	entry := list.Entry{
		Roles:       roles,
		Last:        true,
		Cell:        cell,
		Description: "run",
		Width:       width,
		ID:          "row:enter",
		ChipID:      "chip:enter",
		Zones:       zones,
		X:           x,
		Y:           y,
	}
	row := entry.Render()

	if want := list.ChipRow(roles, true, cell, "run", width); row != want {
		t.Fatalf("an entry drew %q, want the ChipRow %q", text.Plain(row), text.Plain(want))
	}
	if got, want := text.Width(row), width; got != want {
		t.Fatalf("the entry is %d cells wide, want %d", got, want)
	}

	rowRect, ok := zones.Find("row:enter")
	if !ok || rowRect != (zone.Rect{X: x, Y: y, W: width, H: 1}) {
		t.Fatalf("the row zone = %v/%v, want the whole row at (%d, %d)", rowRect, ok, x, y)
	}
	chipRect, ok := zones.Find("chip:enter")
	if !ok || chipRect != (zone.Rect{X: x + 5, Y: y, W: text.Width(cell) - 1, H: 1}) {
		t.Fatalf("the chip zone = %v/%v, want the slab at (%d, %d)", chipRect, ok, x+5, y)
	}
	if id, ok := zones.Hit(x, y); !ok || id != "row:enter" {
		t.Fatalf("the row's first cell answered %q/%v", id, ok)
	}
	if id, ok := zones.Hit(x+5, y); !ok || id != "chip:enter" {
		t.Fatalf("the chip's first cell answered %q/%v", id, ok)
	}
	if id, ok := zones.Hit(x+width, y); ok {
		t.Fatalf("a cell past the row answered %q", id)
	}
	if id, ok := zones.Hit(x, y+1); ok {
		t.Fatalf("a row below the entry answered %q", id)
	}

	// An entry with no key of its own names no chip, so it registers no chip zone.
	quiet := list.Entry{
		Roles: roles, Last: true, Description: "Open the bar", Width: width,
		ID: "row:open", Zones: zone.New(), X: x, Y: y,
	}
	quiet.Render()
	if _, ok := quiet.Zones.Find("chip:enter"); ok {
		t.Fatalf("a keyless entry registered a chip zone")
	}
	if got := quiet.Zones.Len(); got != 1 {
		t.Fatalf("a keyless entry registered %d zones, want only its row", got)
	}
}

// cellOf is the cell x where needle starts in a rendered row.
func cellOf(t *testing.T, row, needle string) int {
	t.Helper()
	plain := []rune(text.Plain(row))
	want := []rune(needle)
	for index := 0; index+len(want) <= len(plain); index++ {
		if string(plain[index:index+len(want)]) == needle {
			return index
		}
	}
	t.Fatalf("%q is not in %q", needle, text.Plain(row))
	return -1
}

// TestTheChipSlabStartsOnTheSameCellEverySurface pins what a column means: the
// selected, hovered and plain rows put the chip slab and the description on the
// same cells, so only the edge cell and the tint differ between them.
func TestTheChipSlabStartsOnTheSameCellEverySurface(t *testing.T) {
	truecolor(t)
	const width, chipWidth = 40, 8

	want, wantDescription := 0, 0
	for _, c := range []struct {
		name  string
		state list.State
	}{
		{name: "plain"},
		{name: "hovered", state: list.State{Hovered: true}},
		{name: "selected", state: list.State{Selected: true}},
	} {
		roles := list.RowRolesFor(c.state)
		row := list.ChipRow(roles, false, chip.Cell(roles.Chip, "prefix+k", chipWidth), "Command bar", width)

		slab := cellOf(t, row, "prefix+k")
		description := cellOf(t, row, "Command bar")
		if got, want := text.Width(row), width; got != want {
			t.Fatalf("the %s row is %d cells wide, want %d", c.name, got, want)
		}
		if c.name == "plain" {
			want, wantDescription = slab, description
			continue
		}
		if slab != want {
			t.Fatalf("the %s row's chip slab starts on cell %d, want %d as on the plain row", c.name, slab, want)
		}
		if description != wantDescription {
			t.Fatalf("the %s row's description starts on cell %d, want %d", c.name, description, wantDescription)
		}
	}
}

// TestAMarkedRowKeepsItsNeighboursColumns pins the marker's place: the gold ◆ of a
// default item follows the row's facts, after two cells of the row's own
// background, so a marked row puts its chip and its name on exactly the cells its
// neighbours use — a mark never becomes a second cursor in front of the row.
func TestAMarkedRowKeepsItsNeighboursColumns(t *testing.T) {
	truecolor(t)
	const width, chipWidth = 46, 8

	plainRoles := list.RowRoles(false)
	markedRoles := list.RowRoles(false)
	selectedRoles := list.RowRoles(true)

	plain := list.ChipRow(plainRoles, false,
		chip.Cell(plainRoles.Chip, "prefix+k", chipWidth), plainRoles.Name.Render("Command bar"), width)
	marked := list.ChipRow(markedRoles, true,
		chip.Cell(markedRoles.Chip, "prefix+i", chipWidth), markedRoles.Name.Render("Command bar")+list.Mark(markedRoles), width)
	selected := list.ChipRow(selectedRoles, true,
		chip.Cell(selectedRoles.Chip, "prefix+i", chipWidth), selectedRoles.Name.Render("Command bar")+list.Mark(selectedRoles), width)

	for name, row := range map[string]string{"marked": marked, "selected": selected} {
		if got, want := cellOf(t, row, "prefix+i"), cellOf(t, plain, "prefix+k"); got != want {
			t.Fatalf("the %s row's chip starts on cell %d, want %d", name, got, want)
		}
		if got, want := cellOf(t, row, "Command bar"), cellOf(t, plain, "Command bar"); got != want {
			t.Fatalf("the %s row's name starts on cell %d, want %d", name, got, want)
		}
		if got, want := text.Width(row), width; got != want {
			t.Fatalf("the %s row is %d cells wide, want %d", name, got, want)
		}
	}

	// The mark is last on the row, and on the tint it keeps the selection's gold.
	if got, want := cellOf(t, marked, text.Mark), cellOf(t, marked, "Command bar")+len("Command bar")+2; got != want {
		t.Fatalf("the mark starts on cell %d, want %d (two cells after the name)", got, want)
	}
	if !strings.Contains(selected, sgr(theme.OnSelection(theme.Gold))) {
		t.Fatalf("the mark lost its gold on the tint: %q", selected)
	}
	if strings.Contains(plain, text.Mark) {
		t.Fatalf("an unmarked row drew a mark: %q", text.Plain(plain))
	}
}
