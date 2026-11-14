package list_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

// truecolor forces the renderer to emit colour, so a test can look for a role in
// a row, and restores the profile afterwards.
func truecolor(t *testing.T) {
	t.Helper()
	before := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(before) })
}

// sgr is the escape sequence a style renders, so a test can look for a role in a
// row without depending on the colour arithmetic lipgloss uses.
func sgr(style lipgloss.Style) string {
	rendered := style.Render("x")
	if i := strings.Index(rendered, "x"); i > 0 {
		return rendered[:i]
	}
	return rendered
}

// TestSelectionTintsInsteadOfPainting pins the selection surface: the row gains
// a tint background and every element keeps the style that carries its meaning.
func TestSelectionTintsInsteadOfPainting(t *testing.T) {
	truecolor(t)
	roles := list.RowRoles(true)
	row := list.SelectedRow(list.Row(roles, false, roles.Name.Render("work")+"  "+roles.Muted.Render("3 tabs")), 30)

	for name, want := range map[string]string{
		"tint":  sgr(theme.SelectionBG),
		"name":  sgr(theme.TextOnSelection),
		"facts": sgr(theme.MutedOnSelection),
		"edge":  sgr(theme.SelectionEdge),
		"guide": sgr(theme.GuideOnSelection),
	} {
		if !strings.Contains(row, want) {
			t.Fatalf("the selected row does not carry the %s role (%q): %q", name, want, row)
		}
	}
	if got, want := text.Width(row), 30; got != want {
		t.Fatalf("selected row width = %d, want %d", got, want)
	}
}

// TestSelectionKeepsTheGuideAndTheColumns pins the two things the tint must not
// cost: the tree guide stays readable, and the row's columns keep their x.
func TestSelectionKeepsTheGuideAndTheColumns(t *testing.T) {
	truecolor(t)
	body := theme.Text.Render("api")

	plain := text.Plain(list.Row(list.RowRoles(false), false, body))
	selected := text.Plain(list.SelectedRow(list.Row(list.RowRoles(true), false, body), 24))

	if !strings.Contains(selected, "├─") {
		t.Fatalf("the selected row lost its tree guide: %q", selected)
	}
	// Columns are compared in cells, not bytes: the guide glyphs are multi-byte.
	at := func(row string) int { return text.Width(row[:strings.Index(row, "api")]) }
	if got, want := at(selected), at(plain); got != want {
		t.Fatalf("the selected row's body starts at cell %d, want %d", got, want)
	}
	if got, want := text.Width(selected), 24; got != want {
		t.Fatalf("the selected row is %d cells, want %d", got, want)
	}
}

// TestDefaultMarkKeepsGoldOnSelection pins that state which outlives the
// selection keeps its colour: the default marker is gold on the tint too.
func TestDefaultMarkKeepsGoldOnSelection(t *testing.T) {
	truecolor(t)
	mark := list.RowRoles(true).Gold.Render(text.Mark)

	if !strings.Contains(mark, sgr(theme.OnSelection(theme.Gold))) {
		t.Fatalf("the gold marker lost its colour on the tint: %q", mark)
	}
	if strings.Contains(mark, sgr(theme.Gold)) {
		t.Fatalf("the gold marker is rendered without the tint: %q", mark)
	}
}

// TestSelectedRowStartsAtTheBarsEdge pins the selection cap: it is the bar's own
// first cell, and the row keeps every other cell it had.
func TestSelectedRowStartsAtTheBarsEdge(t *testing.T) {
	row := "├─ work  ◆"
	selected := text.Plain(list.SelectedRow(row, 20))

	if got, want := text.Width(selected), 20; got != want {
		t.Fatalf("selected row width = %d, want %d", got, want)
	}
	if !strings.HasPrefix(selected, text.SelectMark) {
		t.Fatalf("the bar does not open with the cap: %q", selected)
	}
	if !strings.Contains(selected, "work  ◆") {
		t.Fatalf("the selected row lost its content: %q", selected)
	}
	if strings.Contains(selected, "├─") {
		t.Fatalf("the selection kept the guide glyph in the edge's cell: %q", selected)
	}
}

func TestSectionNamesTheGroupInSmallCaps(t *testing.T) {
	section := text.Plain(list.Section("project", 2))

	if section != "● PROJECT  2" {
		t.Fatalf("section = %q, want %q", section, "● PROJECT  2")
	}
}

func TestGuideClosesTheGroup(t *testing.T) {
	if got := text.Plain(list.Guide(theme.Quiet, true)); !strings.HasSuffix(got, "└─") {
		t.Fatalf("the last guide does not close the group: %q", got)
	}
	if got := text.Plain(list.Guide(theme.Quiet, false)); !strings.HasSuffix(got, "├─") {
		t.Fatalf("a middle guide does not continue the group: %q", got)
	}
}

// TestAgentStateUsesTheStateRoles pins the one place a row may carry green, gold
// or red: the agent state, which never renames what it was given.
func TestAgentStateUsesTheStateRoles(t *testing.T) {
	truecolor(t)
	roles := list.RowRoles(false)

	for state, want := range map[string]lipgloss.Style{
		"":        {},
		"working": theme.Green,
		"idle":    theme.Muted,
		"waiting": theme.Gold,
		"error":   theme.Red,
		"unknown": theme.Muted,
	} {
		got := roles.AgentState(state)
		if got == "" {
			if state != "" {
				t.Fatalf("AgentState(%q) rendered nothing", state)
			}
			continue
		}
		if !strings.Contains(got, sgr(want)) {
			t.Fatalf("AgentState(%q) = %q, want the role %q", state, got, sgr(want))
		}
	}
}

func TestStateWordsUseTheStateRoles(t *testing.T) {
	truecolor(t)

	if got, want := text.Plain(list.StateGlyph(true)), text.Live; got != want {
		t.Fatalf("a running state glyph = %q, want %q", got, want)
	}
	if got, want := text.Plain(list.StateGlyph(false)), text.Idle; got != want {
		t.Fatalf("a stopped state glyph = %q, want %q", got, want)
	}
	if !strings.Contains(list.StateWord(true), sgr(theme.Green)) {
		t.Fatalf("a running state is not green: %q", list.StateWord(true))
	}
	if !strings.Contains(list.StateWord(false), sgr(theme.Muted)) {
		t.Fatalf("a stopped state is not muted: %q", list.StateWord(false))
	}
}

func TestRolesFillTheirOwnSurface(t *testing.T) {
	truecolor(t)
	selected := list.RowRoles(true)

	if got, want := text.Width(selected.Pad("work", 12)), 12; got != want {
		t.Fatalf("Pad is %d cells wide, want %d", got, want)
	}
	if !strings.Contains(selected.Pad("work", 12), sgr(theme.SelectionBG)) {
		t.Fatalf("the selected row's padding is not the tint: %q", selected.Pad("work", 12))
	}
	if got, want := text.Width(selected.AlignRight("a", "b", 10)), 10; got != want {
		t.Fatalf("AlignRight is %d cells wide, want %d", got, want)
	}
	if got, want := text.Plain(selected.AlignRight("a", "b", 10)), "a"+strings.Repeat(" ", 8)+"b"; got != want {
		t.Fatalf("AlignRight = %q, want %q", got, want)
	}
	if got := list.RowRoles(false).Gap(0); got != "" {
		t.Fatalf("a zero gap rendered %q", got)
	}
}

// TestChipRowTintsOverTheChipAndKeepsTheColumns pins the selection on a row whose
// first column is a chip: the slab is one step brighter, the tint runs to the
// row's width, and the description keeps its x.
func TestChipRowTintsOverTheChipAndKeepsTheColumns(t *testing.T) {
	truecolor(t)
	const width, chipWidth = 40, 9
	description := "Command bar"

	plain := list.ChipRow(list.RowRoles(false), false, chip.Cell(theme.ChipText, "prefix+k", chipWidth), description, width)
	selected := list.ChipRow(list.RowRoles(true), false, chip.Cell(theme.ChipOnSelection, "prefix+k", chipWidth), description, width)

	if got, want := text.Width(selected), width; got != want {
		t.Fatalf("the selected chip row is %d cells wide, want %d", got, want)
	}
	if !strings.HasPrefix(text.Plain(selected), text.SelectMark) {
		t.Fatalf("the selected chip row has no accent edge: %q", text.Plain(selected))
	}
	if !strings.Contains(selected, sgr(theme.ChipOnSelection)) {
		t.Fatalf("the chip slab did not step up on the tint: %q", selected)
	}
	if !strings.Contains(selected, sgr(theme.SelectionBG)) {
		t.Fatalf("the chip row is not tinted past the chip: %q", selected)
	}
	at := func(row string) int {
		plain := text.Plain(row)
		return text.Width(plain[:strings.Index(plain, description)])
	}
	if got, want := at(selected), at(plain); got != want {
		t.Fatalf("the description starts at cell %d on the tint and %d on the panel", got, want)
	}
	if strings.Contains(selected, sgr(theme.ChipText)) {
		t.Fatalf("the selected chip row also draws the panel slab: %q", selected)
	}
}
