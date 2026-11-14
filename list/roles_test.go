package list_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/theme"
)

// TestTheAccentChipKeepsItsAccentOnEverySurface pins the accent chip role: a key
// that carries the product's accent (a guide's binding) keeps that accent on the
// panel, under the pointer and on the selection, where it steps onto the brighter
// slab like the plain chip does. No product overrides Roles.Chip by hand.
func TestTheAccentChipKeepsItsAccentOnEverySurface(t *testing.T) {
	truecolor(t)
	const width, chipWidth = 40, 8
	accent := strings.TrimSuffix(strings.TrimPrefix(sgr(theme.Accent), "\x1b["), "m")

	for _, c := range []struct {
		name  string
		state list.State
		slab  string
	}{
		{name: "plain", slab: sgr(theme.ChipAccent)},
		{name: "hovered", state: list.State{Hovered: true}, slab: sgr(theme.ChipAccent)},
		{name: "selected", state: list.State{Selected: true}, slab: sgr(theme.ChipAccentOnSelection)},
	} {
		roles := list.RowRolesFor(c.state)
		row := list.ChipRow(roles, false,
			chip.Cell(roles.ChipAccent, "prefix+k", chipWidth), roles.Name.Render("Command bar"), width)

		if !strings.Contains(row, c.slab) {
			t.Fatalf("the %s row's accent chip did not keep its slab (%q): %q", c.name, c.slab, row)
		}
		if !strings.Contains(row, accent) {
			t.Fatalf("the %s row's accent chip lost the accent role: %q", c.name, row)
		}
		if strings.Contains(row, sgr(theme.ChipText)) {
			t.Fatalf("the %s row's accent chip fell back to the plain text slab: %q", c.name, row)
		}
	}

	// The plain chip role is untouched, so every existing caller keeps its bytes.
	for name, c := range map[string]struct {
		roles list.Roles
		want  lipgloss.Style
	}{
		"panel":    {list.RowRoles(false), theme.ChipText},
		"hovered":  {list.RowRolesFor(list.State{Hovered: true}), theme.ChipText},
		"selected": {list.RowRoles(true), theme.ChipOnSelection},
	} {
		if got, want := sgr(c.roles.Chip), sgr(c.want); got != want {
			t.Fatalf("the %s row's plain chip role = %q, want %q", name, got, want)
		}
	}
}
