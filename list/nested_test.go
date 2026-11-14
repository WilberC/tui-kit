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

// nestIndent is how far a nested row starts: its parent's continuation column, its
// own indent and guide, and the space after them.
const nestIndent = 10

// herdrChild composes a nested workspace row the way herdr-suite composes it by
// hand today: the parent's continuation column, the child's own tree guide, then
// the content the product painted.
func herdrChild(parentLast, last bool, content string) string {
	continuation := "  " + " " + "  "
	if !parentLast {
		continuation = "  " + theme.Quiet.Render("│") + "  "
	}
	return continuation + list.Guide(theme.Quiet, last) + " " + content
}

// nestedRow is the same composition written with a row's own roles, which is what
// a product gets when a nested row wears the hover or the selection: on the panel
// its styles paint no background, so it is herdrChild's string exactly.
func nestedRow(roles list.Roles, parentLast, last bool, content string) string {
	glyph := roles.Space.Render(" ")
	if !parentLast {
		glyph = roles.Guide.Render("│")
	}
	guide := roles.Guide.Render("├─")
	if last {
		guide = roles.Guide.Render("└─")
	}
	return roles.Space.Render("  ") + glyph + roles.Space.Render("  ") +
		roles.Space.Render("  ") + guide + roles.Space.Render(" ") + content
}

// TestANestedRowHangsUnderItsParentGuide pins the nested row: an entry with a
// parent draws its own guide under the parent's continuation column — the column
// keeps the parent's guide open when the parent has a sibling below it and stays
// blank when the parent closes the group — and its bytes are exactly the ones
// herdr-suite composes by hand for a nested workspace row, on the panel, under the
// pointer and on the selection.
func TestANestedRowHangsUnderItsParentGuide(t *testing.T) {
	truecolor(t)
	const width = 58

	for _, c := range []struct {
		name  string
		state list.State
	}{
		{name: "plain"},
		{name: "hovered", state: list.State{Hovered: true}},
		{name: "selected", state: list.State{Selected: true}},
	} {
		roles := list.RowRolesFor(c.state)
		content := roles.Name.Render("web") + "  " + roles.Muted.Render("3 tabs · 2 panes")
		for _, parentLast := range []bool{false, true} {
			for _, last := range []bool{false, true} {
				row := list.Entry{
					Roles:       roles,
					Last:        last,
					Parent:      &list.Parent{Last: parentLast},
					Description: content,
					Width:       width,
				}.Render()

				reference := nestedRow(roles, parentLast, last, content)
				want := roles.Pad(reference, width)
				if roles.Selected {
					want = list.SelectedRow(reference, width)
				}
				if row != want {
					t.Fatalf("the %s nested row (parent last %v, last %v) drew %q, want %q",
						c.name, parentLast, last, row, want)
				}
				if got := text.Width(row); got != width {
					t.Fatalf("the %s nested row is %d cells wide, want %d", c.name, got, width)
				}
				// A child starts five cells past its parent: the continuation column
				// is the parent's own indent and guide.
				if at := cellOf(t, row, "web"); at != nestIndent {
					t.Fatalf("the %s nested row starts its body on cell %d, want %d", c.name, at, nestIndent)
				}
				if open := strings.Contains(text.Plain(row), "│"); open == parentLast {
					t.Fatalf("the %s nested row (parent last %v) %s the parent's guide open: %q",
						c.name, parentLast, map[bool]string{true: "kept", false: "did not keep"}[open], text.Plain(row))
				}
			}
		}
	}

	// The panel state is herdr-suite's own composition, byte for byte: the frame
	// fits a row to its width, and the kit pads it instead, with plain cells.
	plain := list.RowRolesFor(list.State{})
	body := "web  3 tabs · 2 panes"
	for _, parentLast := range []bool{false, true} {
		for _, last := range []bool{false, true} {
			row := list.Entry{
				Roles: plain, Last: last, Parent: &list.Parent{Last: parentLast},
				Description: body, Width: width,
			}.Render()
			want := herdrChild(parentLast, last, body)
			want += strings.Repeat(" ", width-text.Width(want))
			if row != want {
				t.Fatalf("the plain nested row (parent last %v, last %v) drew %q, want herdr-suite's %q",
					parentLast, last, row, want)
			}
		}
	}
}

// TestATopLevelRowKeepsNoContinuation pins the other side of the field: an entry
// with no parent draws exactly the row it drew before the field existed, so a
// list that never nests pays nothing for it.
func TestATopLevelRowKeepsNoContinuation(t *testing.T) {
	truecolor(t)
	const width = 46
	for _, state := range []list.State{{}, {Hovered: true}, {Selected: true}} {
		roles := list.RowRolesFor(state)
		body := roles.Name.Render("api") + "  " + roles.Muted.Render("herdr · 3 tabs")
		entry := list.Entry{Roles: roles, Last: false, Description: body, Width: width}
		if got, want := entry.Render(), list.ChipRow(roles, false, "", body, width); got != want {
			t.Fatalf("a row with no parent drew %q, want %q", got, want)
		}
	}
}

// TestANestedRowRecordsItsOwnZone pins the nested row's click target: the whole
// row is recorded at the caller's x and y — the continuation column included, so a
// click on the guide acts on the row just like a click on its text — and a chip
// column moves ten cells in, not five.
func TestANestedRowRecordsItsOwnZone(t *testing.T) {
	truecolor(t)
	const width, x, y = 60, 4, 7

	zones := zone.New()
	row := list.Entry{
		Roles:       list.RowRolesFor(list.State{Selected: true}),
		Last:        true,
		Parent:      &list.Parent{Last: false},
		Description: "web",
		Width:       width,
		ID:          "child:1",
		Zones:       zones,
		X:           x,
		Y:           y,
	}.Render()

	if rect, ok := zones.Find("child:1"); !ok || rect != (zone.Rect{X: x, Y: y, W: width, H: 1}) {
		t.Fatalf("the nested row zone = %v/%v, want the whole row at (%d, %d)", rect, ok, x, y)
	}
	for _, offset := range []int{0, 1, nestIndent - 1, nestIndent, width - 1} {
		if id, ok := zones.Hit(x+offset, y); !ok || id != "child:1" {
			t.Fatalf("the nested row's cell %d answered %q/%v, want the nested row", offset, id, ok)
		}
	}
	if got := text.Width(row); got != width {
		t.Fatalf("the nested row is %d cells wide, want %d", got, width)
	}

	// The chip column follows the nest: a slab starts ten cells in, and the zone it
	// records is where it was drawn.
	chips := zone.New()
	list.Entry{
		Roles:       list.RowRoles(false),
		Last:        true,
		Parent:      &list.Parent{Last: false},
		Cell:        chip.Cell(theme.ChipText, "prefix+k", 8),
		ChipID:      "chip:1",
		Description: "web",
		Width:       width,
		ID:          "child:1",
		Zones:       chips,
		X:           x,
		Y:           y,
	}.Render()
	if id, ok := chips.Hit(x+rowIndent, y); !ok || id != "child:1" {
		t.Fatalf("the fifth cell of a nested row answered %q/%v, want the row: the chip starts ten cells in", id, ok)
	}
	if id, ok := chips.Hit(x+nestIndent, y); !ok || id != "chip:1" {
		t.Fatalf("the nested row's chip cell answered %q/%v, want the chip", id, ok)
	}
}

// TestANestedRowKeepsItsTrailingColumnOnTheRowEdge pins the two together: a nested
// row's trailing column still ends on the row's last cell, so the nest eats the
// description's room and never the column's.
func TestANestedRowKeepsItsTrailingColumnOnTheRowEdge(t *testing.T) {
	truecolor(t)
	const width = 60
	const trailing = "2d ago"

	for _, state := range []list.State{{}, {Hovered: true}, {Selected: true}} {
		roles := list.RowRolesFor(state)
		row := list.Entry{
			Roles:       roles,
			Last:        true,
			Parent:      &list.Parent{Last: false},
			Description: roles.Name.Render(strings.Repeat("a long description ", 3)),
			Trailing:    roles.Muted.Render(trailing),
			Width:       width,
		}.Render()

		if got := text.Width(row); got != width {
			t.Fatalf("the nested row with a trailing column is %d cells wide, want %d", got, width)
		}
		start := cellOf(t, row, trailing)
		if start < nestIndent {
			t.Fatalf("the trailing column starts on cell %d, inside the nest's %d", start, nestIndent)
		}
		if got, want := start+text.Width(trailing), width; got != want {
			t.Fatalf("the nested row's trailing column ends on cell %d, want %d", got, want)
		}
	}
}

// rowIndent is how many cells a top-level row spends before its body.
const rowIndent = 5
