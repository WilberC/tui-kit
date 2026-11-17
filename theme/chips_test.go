package theme_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/theme"
)

// filled is the composition a filled chip must equal: the accent (or another
// role) as the slab's background, the label colour as the foreground, bold like
// every other chip.
func filled(bg, fg lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Background(bg).Foreground(fg).Bold(true)
}

// TestChipAccentFillFillsTheSlabWithTheAccent pins the one place a chip is
// filled with the accent: the navbar's active tab. The slab carries the accent
// as its background — not the chip surface — and the label sits in the chip
// surface colour (the dark slab colour), so the fill reads as a filled slab and
// never needs a new palette role.
func TestChipAccentFillFillsTheSlabWithTheAccent(t *testing.T) {
	withColors(t)

	if got, want := sgr(theme.ChipAccentFill), sgr(filled(theme.ColorAccent, theme.ColorChip)); got != want {
		t.Fatalf("ChipAccentFill renders %q, want the accent fill %q", got, want)
	}
	if strings.Contains(theme.ChipAccentFill.Render("x"), sgr(theme.ChipText)) {
		t.Fatalf("ChipAccentFill is a quiet slab, not a filled one: %q", theme.ChipAccentFill.Render("x"))
	}
}

// TestChipMutedIsTheChipSurfaceWithTheMutedLabel pins the off/disabled chip:
// the same dark slab as every other chip, with the muted label and the same
// bold, so it never reads as prose or as a slab of its own invention.
func TestChipMutedIsTheChipSurfaceWithTheMutedLabel(t *testing.T) {
	withColors(t)

	if got, want := sgr(theme.ChipMuted), sgr(filled(theme.ColorChip, theme.ColorMuted)); got != want {
		t.Fatalf("ChipMuted renders %q, want the chip slab with the muted label %q", got, want)
	}
	if strings.Contains(theme.ChipMuted.Render("x"), sgr(theme.ChipText)) {
		t.Fatalf("ChipMuted carries the text label instead of the muted one: %q", theme.ChipMuted.Render("x"))
	}
}

// TestTheFilledAndMutedChipsFollowUse pins that both chips are rebuilt by
// theme.Use like every other composition: a theme with another accent and
// another chip surface changes both, and Tokyo Night puts them back.
func TestTheFilledAndMutedChipsFollowUse(t *testing.T) {
	withColors(t)
	other := theme.TokyoNight()
	other.Name = "test"
	other.Accent = "#eeeeee"
	other.Chip = "#111111"
	if err := theme.Use(other); err != nil {
		t.Fatalf("theme.Use(test): %v", err)
	}
	t.Cleanup(func() { _ = theme.Use(theme.TokyoNight()) })

	if got, want := sgr(theme.ChipAccentFill), sgr(filled(other.Accent, other.Chip)); got != want {
		t.Fatalf("the filled chip did not follow Use: %q, want %q", got, want)
	}
	if got, want := sgr(theme.ChipMuted), sgr(filled(other.Chip, other.Muted)); got != want {
		t.Fatalf("the muted chip did not follow Use: %q, want %q", got, want)
	}

	if err := theme.Use(theme.TokyoNight()); err != nil {
		t.Fatalf("theme.Use(tokyo-night): %v", err)
	}
	if got, want := sgr(theme.ChipAccentFill), sgr(filled(theme.ColorAccent, theme.ColorChip)); got != want {
		t.Fatalf("the built-in theme did not put the accent fill back: %q, want %q", got, want)
	}
	if got, want := sgr(theme.ChipMuted), sgr(filled(theme.ColorChip, theme.ColorMuted)); got != want {
		t.Fatalf("the built-in theme did not put the chip surface back: %q, want %q", got, want)
	}
}
