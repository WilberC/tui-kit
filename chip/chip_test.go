package chip_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

// truecolor forces the renderer to emit colour, so a test can look for a role in
// a rendered string, and restores the profile afterwards.
func truecolor(t *testing.T) {
	t.Helper()
	before := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(before) })
}

// sgr is the escape sequence a style renders, so a test can look for a role
// without depending on the colour arithmetic lipgloss uses.
func sgr(style lipgloss.Style) string {
	rendered := style.Render("x")
	if i := strings.Index(rendered, "x"); i > 0 {
		return rendered[:i]
	}
	return rendered
}

func TestChipIsAPaddedSlab(t *testing.T) {
	got := text.Plain(chip.Chip(theme.ChipText, "enter"))

	if got != " enter " {
		t.Fatalf("chip = %q, want %q", got, " enter ")
	}
	if strings.ContainsAny(got, "[]") {
		t.Fatalf("chip carries brackets: %q", got)
	}
}

func TestPairPutsOnePlainSpaceBeforeTheLabel(t *testing.T) {
	got := chip.Pair(theme.ChipText, "enter", "run")

	if want := chip.Chip(theme.ChipText, "enter") + " " + theme.Muted.Render("run"); got != want {
		t.Fatalf("pair = %q, want %q", got, want)
	}
}

func TestHintsDropWholeHintsInsteadOfWrapping(t *testing.T) {
	hints := []chip.Hint{
		{Key: "↑↓", Label: "move"},
		{Key: "enter", Label: "attach"},
		{Key: "alt+enter", Label: "create"},
	}

	wide := chip.Hints(80, hints...)
	for _, hint := range hints {
		if !strings.Contains(wide, " "+hint.Key+" ") {
			t.Fatalf("a wide bar dropped %q: %q", hint.Key, wide)
		}
	}
	if width := text.Width(wide); width > 80 {
		t.Fatalf("hint bar width = %d, want at most 80", width)
	}

	narrow := chip.Hints(20, hints...)
	if width := text.Width(narrow); width > 20 {
		t.Fatalf("narrow hint bar width = %d, want at most 20 (%q)", width, narrow)
	}
	if strings.Contains(narrow, " alt+enter ") {
		t.Fatalf("the narrowest bar kept the last hint instead of dropping it: %q", narrow)
	}
	if !strings.Contains(narrow, " ↑↓ ") {
		t.Fatalf("the narrowest bar dropped the first hint: %q", narrow)
	}
	if width := text.Width(chip.Hints(1, hints...)); width > 1 {
		t.Fatalf("a one-cell bar is %d cells wide: %q", width, chip.Hints(1, hints...))
	}
}

func TestLegendJoinsItemsWithPlainSpaces(t *testing.T) {
	first := chip.Pair(theme.ChipText, "x", "after prefix")
	second := chip.Chip(theme.ChipText, "ctrl+x") + " " + theme.Muted.Render("direct")
	legend := chip.Legend(first, second)

	if want := first + chip.Gutter + second; legend != want {
		t.Fatalf("legend = %q, want %q", legend, want)
	}
	if strings.Contains(legend, " · ") || strings.Contains(legend, "│") {
		t.Fatalf("a separator touches the chips' background: %q", legend)
	}
	if got, want := text.Width(chip.Gutter), 2; got != want {
		t.Fatalf("the legend gutter is %d cells wide, want %d", got, want)
	}
}

func TestCellPadsEveryRowToTheSameColumn(t *testing.T) {
	width := 9
	short := chip.Cell(theme.ChipText, "x", width)
	long := chip.Cell(theme.ChipText, "ctrl+shift+u", width)

	if got, want := text.Width(short), width+3; got != want {
		t.Fatalf("short cell is %d cells wide, want %d (%q)", got, want, short)
	}
	if got, want := text.Width(long), width+3; got != want {
		t.Fatalf("long cell is %d cells wide, want %d (%q)", got, want, long)
	}
	if !strings.HasSuffix(text.Plain(short), "  ") {
		t.Fatalf("a cell does not end with the plain space that separates the description: %q", text.Plain(short))
	}
	if text.Plain(short)[0] != ' ' {
		t.Fatalf("a cell does not open with the chip's padding cell: %q", text.Plain(short))
	}
}

func TestUnsetCellKeepsTheColumnWithoutASlab(t *testing.T) {
	truecolor(t)
	width := 9
	unset := chip.UnsetCell(theme.Muted, text.Unset, width)

	if got, want := text.Width(unset), width+3; got != want {
		t.Fatalf("unset cell is %d cells wide, want %d (%q)", got, want, unset)
	}
	if !strings.Contains(text.Plain(unset), text.Unset) {
		t.Fatalf("the unset cell lost its dash: %q", unset)
	}
	if strings.Contains(unset, sgr(theme.ChipText)) {
		t.Fatalf("the unset cell carries a chip slab, so the group reads as empty boxes: %q", unset)
	}
}

func TestWidthIsTheWidestLabel(t *testing.T) {
	if got, want := chip.Width("x", "ctrl+shift+u", "enter"), len("ctrl+shift+u"); got != want {
		t.Fatalf("Width = %d, want %d", got, want)
	}
	if got, want := chip.Width(), 0; got != want {
		t.Fatalf("Width of no labels = %d, want %d", got, want)
	}
}

// TestChipFollowsTheProfileAtCallTime pins the bug the products hit: a chip
// rendered once at init kept the Ascii profile and lost its colour forever.
func TestChipFollowsTheProfileAtCallTime(t *testing.T) {
	before := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(before) })

	plain := chip.Chip(theme.ChipText, "enter")
	if strings.ContainsRune(plain, 0x1b) {
		t.Fatalf("an Ascii chip carries escapes: %q", plain)
	}

	lipgloss.SetColorProfile(termenv.TrueColor)
	if coloured := chip.Chip(theme.ChipText, "enter"); !strings.ContainsRune(coloured, 0x1b) {
		t.Fatalf("the chip stayed plain after the profile changed: %q", coloured)
	}
}

func TestChipOnSelectionIsOneStepBrighter(t *testing.T) {
	truecolor(t)
	panel := chip.Chip(theme.ChipText, "x")
	tinted := chip.Chip(theme.ChipOnSelection, "x")

	if panel == tinted {
		t.Fatalf("the chip looks the same on both surfaces: %q", tinted)
	}
	if !strings.Contains(tinted, sgr(theme.ChipOnSelection)) {
		t.Fatalf("the chip on the selection does not use its surface: %q", tinted)
	}
	if !strings.Contains(panel, sgr(theme.ChipText)) {
		t.Fatalf("the panel chip does not use its surface: %q", panel)
	}
}
