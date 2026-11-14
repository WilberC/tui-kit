package theme_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/theme"
)

// rgb reads a palette value as its three channels.
func rgb(t *testing.T, value string) [3]int {
	t.Helper()
	digits, ok := strings.CutPrefix(value, "#")
	if !ok || len(digits) != 6 {
		t.Fatalf("%q is not a palette colour", value)
	}
	var out [3]int
	for channel := 0; channel < 3; channel++ {
		parsed, err := strconv.ParseUint(digits[channel*2:channel*2+2], 16, 8)
		if err != nil {
			t.Fatalf("%q is not a palette colour: %v", value, err)
		}
		out[channel] = int(parsed)
	}
	return out
}

// TestHoverIsOneStepBetweenTheBaseAndTheSelection pins the hover tint of S8: one
// step above the terminal's own background and still fainter than the selection,
// so a hover previews the selection instead of competing with it. The base is the
// terminal's Tokyo Night background, which the kit itself never paints.
func TestHoverIsOneStepBetweenTheBaseAndTheSelection(t *testing.T) {
	const base = "#1a1b26"

	background := rgb(t, base)
	hover := rgb(t, string(theme.ColorHover))
	selection := rgb(t, string(theme.ColorSelection))

	for channel, name := range map[int]string{0: "red", 1: "green", 2: "blue"} {
		if background[channel] >= hover[channel] {
			t.Fatalf("the hover tint's %s channel is %d, no lighter than the base's %d", name, hover[channel], background[channel])
		}
		if hover[channel] >= selection[channel] {
			t.Fatalf("the hover tint's %s channel is %d, not fainter than the selection's %d", name, hover[channel], selection[channel])
		}
	}
}

// TestOnHoverKeepsTheRoleAndAddsTheFaintTint pins the composition the rows use:
// green stays green and gold stays gold on the hover tint, exactly as they do on
// the selection.
func TestOnHoverKeepsTheRoleAndAddsTheFaintTint(t *testing.T) {
	withColors(t)
	hovered := theme.OnHover(theme.Gold)

	if !strings.Contains(hovered.Render("x"), "38;2;224;175;104") {
		t.Fatalf("theme.OnHover lost the role's own colour: %q", hovered.Render("x"))
	}
	if !strings.Contains(hovered.Render("x"), "48;2;32;39;63") {
		t.Fatalf("theme.OnHover did not add the hover tint: %q", hovered.Render("x"))
	}
	if theme.OnHover(theme.Gold).Render("x") == theme.OnSelection(theme.Gold).Render("x") {
		t.Fatalf("the hover tint is the selection tint")
	}
}

// TestPaletteHoldsEveryRole pins the palette as data: a tool that reproduces the
// guide's colours must see every role in it. Two roles may share a value — the chip
// on the tint is the border's blue — so the list is by role, not by value.
func TestPaletteHoldsEveryRole(t *testing.T) {
	palette := theme.Palette()
	if got, want := len(palette), 15; got != want {
		t.Fatalf("theme.Palette holds %d colours, want the %d roles DESIGN.md lists", got, want)
	}
	seen := map[string]bool{}
	for _, colour := range palette {
		value := string(colour)
		if !strings.HasPrefix(value, "#") || len(value) != 7 {
			t.Fatalf("%q is not a palette colour", value)
		}
		seen[value] = true
	}
	for _, role := range []struct {
		name  string
		value lipgloss.Color
	}{
		{"text", theme.ColorText}, {"muted", theme.ColorMuted}, {"guide", theme.ColorGuide},
		{"border", theme.ColorBorder}, {"accent", theme.ColorAccent}, {"gold", theme.ColorGold},
		{"teal", theme.ColorTeal}, {"green", theme.ColorGreen}, {"red", theme.ColorRed},
		{"chip", theme.ColorChip}, {"selection", theme.ColorSelection}, {"hover", theme.ColorHover},
		{"chip on selection", theme.ColorChipOnSelection}, {"muted on selection", theme.ColorMutedOnSelection},
		{"guide on selection", theme.ColorGuideOnSelection},
	} {
		if !seen[string(role.value)] {
			t.Fatalf("theme.Palette is missing the %s role (%s)", role.name, role.value)
		}
	}
}
