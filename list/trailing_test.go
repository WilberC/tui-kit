package list_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/list"
)

// cell is one rendered cell: the rune drawn there and the background in force at
// it ("none" for the terminal's own background).
type cell struct {
	r  rune
	bg string
}

// cells splits a rendered row into its cells and the background each carries, so a
// test can check that a tint is unbroken instead of trusting its own composition.
func cells(row string) []cell {
	out := []cell{}
	bg := "none"
	for i := 0; i < len(row); {
		if strings.HasPrefix(row[i:], "\x1b[") {
			end := strings.IndexByte(row[i:], 'm')
			if end < 0 {
				break
			}
			bg = nextBackground(bg, row[i+2:i+end])
			i += end + 1
			continue
		}
		r := []rune(row[i:])[0]
		out = append(out, cell{r: r, bg: bg})
		i += len(string(r))
	}
	return out
}

// nextBackground is the background a sequence leaves in force: a full reset returns
// to the terminal's own, a truecolor background sets one, and anything else keeps
// what was there.
func nextBackground(bg, params string) string {
	fields := strings.Split(params, ";")
	for i := 0; i < len(fields); i++ {
		switch fields[i] {
		case "0":
			bg = "none"
		case "48":
			if i+4 < len(fields) && fields[i+1] == "2" {
				return "48;2;" + strings.Join(fields[i+2:i+5], ";")
			}
		}
	}
	return bg
}

// tint is the background a style paints a cell with, in the form cells reports.
func tint(style lipgloss.Style) string {
	rendered := style.Render("x")
	start := strings.Index(rendered, "\x1b[")
	if start < 0 {
		return "none"
	}
	end := strings.IndexByte(rendered, 'm')
	if end < 0 || end < start {
		return "none"
	}
	at := strings.Index(rendered[start:end], "48;2;")
	if at < 0 {
		return "none"
	}
	parts := strings.Split(rendered[start+at+len("48;2;"):end], ";")
	if len(parts) < 3 {
		return "none"
	}
	return "48;2;" + strings.Join(parts[:3], ";")
}

// TestTheTrailingColumnEndsOnTheRowEdgeWithItsTint pins the table's right edge: the
// trailing column's last cell is the row's last cell, and on the hover and the
// selection every cell from the column's first to that last cell carries the row's
// tint, so no letter sits outside the band.
func TestTheTrailingColumnEndsOnTheRowEdgeWithItsTint(t *testing.T) {
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
		row := list.Entry{
			Roles:       roles,
			Last:        true,
			Cell:        chip.Cell(roles.Chip, "prefix+k", chipWidth),
			Description: roles.Name.Render(description),
			Trailing:    roles.Muted.Render(trailing),
			Width:       width,
		}.Render()

		drawn := cells(row)
		if got, want := len(drawn), width; got != want {
			t.Fatalf("the %s row is %d cells wide, want %d", c.name, got, want)
		}
		if got, want := string(drawn[len(drawn)-1].r), "e"; got != want {
			t.Fatalf("the %s row's last cell holds %q, want the trailing column's last letter", c.name, got)
		}

		want := tint(roles.Space)
		start := cellOf(t, row, trailing)
		for index := start; index < len(drawn); index++ {
			if got := drawn[index].bg; got != want {
				t.Fatalf("the %s row's trailing cell %d has background %q, want the row's own %q (row %q)",
					c.name, index, got, want, string(row))
			}
		}
	}
}

// TestAReservedTrailingColumnStartsOnTheSameCell pins a table's last column: when
// the caller reserves a width, every row's trailing text starts on the same cell —
// the cell its header names — the column ends on the row's edge, and the cells the
// text leaves empty are the row's own background, so the tint still reaches the
// edge.
func TestAReservedTrailingColumnStartsOnTheSameCell(t *testing.T) {
	truecolor(t)
	const width, chipWidth, column = 44, 8, 12

	for _, c := range []struct {
		name  string
		state list.State
	}{
		{name: "plain"},
		{name: "hovered", state: list.State{Hovered: true}},
		{name: "selected", state: list.State{Selected: true}},
	} {
		roles := list.RowRolesFor(c.state)
		start := -1
		for _, source := range []string{"herdr-suite", "custom", "herdr-utils"} {
			row := list.Entry{
				Roles:         roles,
				Last:          true,
				Cell:          chip.Cell(roles.Chip, "prefix+k", chipWidth),
				Description:   roles.Name.Render("Command bar"),
				Trailing:      roles.Muted.Render(source),
				TrailingWidth: column,
				Width:         width,
			}.Render()

			drawn := cells(row)
			if got, want := len(drawn), width; got != want {
				t.Fatalf("the %s row's %q column is %d cells wide, want %d", c.name, source, got, want)
			}
			at := cellOf(t, row, source)
			if start < 0 {
				start = at
				if got, want := start, width-column; got != want {
					t.Fatalf("the %s row's column starts on cell %d, want the reserved cell %d", c.name, got, want)
				}
			}
			if at != start {
				t.Fatalf("the %s row's %q starts on cell %d, want the column's cell %d", c.name, source, at, start)
			}

			// The column always ends on the row's edge, and whatever the text leaves
			// empty is the row's own background.
			want := tint(roles.Space)
			for index := start; index < len(drawn); index++ {
				if got := drawn[index].bg; got != want {
					t.Fatalf("the %s row's column cell %d has background %q, want the row's own %q",
						c.name, index, got, want)
				}
			}
		}
	}
}
