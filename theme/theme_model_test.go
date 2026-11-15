package theme_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/theme"
)

func mustUse(t *testing.T, them theme.Theme) {
	t.Helper()
	if err := theme.Use(them); err != nil {
		t.Fatalf("theme.Use(%s): %v", them.Name, err)
	}
	t.Cleanup(func() {
		if err := theme.Use(theme.TokyoNight()); err != nil {
			t.Fatalf("restore tokyo-night: %v", err)
		}
	})
}

func TestTokyoNightIsTheBuiltInTheme(t *testing.T) {
	them := theme.TokyoNight()
	if them.Name != "tokyo-night" {
		t.Fatalf("the built-in theme is %q, want \"tokyo-night\"", them.Name)
	}
	cases := map[string]struct {
		got  string
		want string
	}{
		"text":              {string(them.Text), string(theme.ColorText)},
		"muted":             {string(them.Muted), string(theme.ColorMuted)},
		"guide":             {string(them.Guide), string(theme.ColorGuide)},
		"border":            {string(them.Border), string(theme.ColorBorder)},
		"accent":            {string(them.Accent), string(theme.ColorAccent)},
		"gold":              {string(them.Gold), string(theme.ColorGold)},
		"teal":              {string(them.Teal), string(theme.ColorTeal)},
		"green":             {string(them.Green), string(theme.ColorGreen)},
		"red":               {string(them.Red), string(theme.ColorRed)},
		"chip surface":      {string(them.Chip), string(theme.ColorChip)},
		"selection tint":    {string(them.Selection), string(theme.ColorSelection)},
		"hover tint":        {string(them.Hover), string(theme.ColorHover)},
		"chip on tint":      {string(them.ChipOnSelection), string(theme.ColorChipOnSelection)},
		"muted on tint":     {string(them.MutedOnSelection), string(theme.ColorMutedOnSelection)},
		"the guide on tint": {string(them.GuideOnSelection), string(theme.ColorGuideOnSelection)},
	}
	for name, c := range cases {
		if c.got != c.want {
			t.Fatalf("the %s role of the built-in theme is %q, want %q", name, c.got, c.want)
		}
	}
}

func TestActiveIsTokyoNightBeforeAnyUse(t *testing.T) {
	if got := theme.Active(); got != theme.TokyoNight() {
		t.Fatalf("the active theme is %+v, want the built-in", got)
	}
}

func TestUseRecolorsTheStyles(t *testing.T) {
	withColors(t)
	them := theme.TokyoNight()
	them.Name = "test"
	them.Text = "#abcdf0"
	them.Hover = "#010203"
	them.Muted = "196"
	them.Guide = ""
	mustUse(t, them)

	if got := theme.Active(); got.Text != "#abcdf0" || got.Hover != "#010203" {
		t.Fatalf("the active theme did not become the applied one: %+v", got)
	}
	for name, style := range map[string]lipgloss.Style{
		"text":          theme.Text,
		"text on hover": theme.TextOnHover,
		"text on tint":  theme.TextOnSelection,
		"strong":        theme.Strong,
		"chip text":     theme.ChipText,
		"chip on tint":  theme.ChipOnSelection,
	} {
		rendered := sgr(style)
		if !strings.Contains(rendered, "38;2;171;205;240") {
			t.Fatalf("the %s style kept the old text colour: %q", name, rendered)
		}
	}
	if !strings.Contains(sgr(theme.ChipAccent), "38;2;121;162;247") {
		t.Fatalf("the accent chip lost its own role: %q", sgr(theme.ChipAccent))
	}
	if !strings.Contains(sgr(theme.HoverBG), "48;2;1;2;3") {
		t.Fatalf("the hover surface lost the applied tint: %q", sgr(theme.HoverBG))
	}
	if !strings.Contains(sgr(theme.Muted), "38;5;196") {
		t.Fatalf("the muted style lost the 256 index: %q", sgr(theme.Muted))
	}
	if strings.Contains(sgr(theme.Quiet), "38") {
		t.Fatalf("the terminal default guide role carries a colour escape: %q", sgr(theme.Quiet))
	}
}

func TestUseRejectsAnInvalidThemeAndKeepsTheOldOne(t *testing.T) {
	withColors(t)
	bad := theme.TokyoNight()
	bad.Text = "#12g45z"
	if err := theme.Use(bad); err == nil {
		t.Fatal("theme.Use accepted a colour that is neither a hex, an index nor empty")
	}
	if got := theme.Active(); got != theme.TokyoNight() {
		t.Fatalf("a rejected Use still changed the theme: %+v", got)
	}
	if !strings.Contains(sgr(theme.Text), "38;2;192;202;245") {
		t.Fatalf("a rejected Use still changed the styles: %q", sgr(theme.Text))
	}
}

func TestUseRejectsAnOutOfRangeIndex(t *testing.T) {
	them := theme.TokyoNight()
	them.Text = "256"
	if err := theme.Use(them); err == nil {
		t.Fatal("theme.Use accepted the index 256")
	}
	them.Text = "-1"
	if err := theme.Use(them); err == nil {
		t.Fatal("theme.Use accepted the index -1")
	}
}

func TestUseRejectsABrokenHex(t *testing.T) {
	for _, c := range []string{"#12345", "#1234567", "#", "123456"} {
		them := theme.TokyoNight()
		them.Text = lipgloss.Color(c)
		if err := theme.Use(them); err == nil {
			t.Fatalf("theme.Use accepted the colour %q", c)
		}
	}
}

func TestAnUnnamedThemeIsCalledCustom(t *testing.T) {
	them := theme.TokyoNight()
	them.Name = ""
	mustUse(t, them)
	if got := theme.Active().Name; got != "custom" {
		t.Fatalf("the unnamed theme is called %q, want \"custom\"", got)
	}
}

func TestAHookRegisteredOnUseRunsOnEveryUse(t *testing.T) {
	runs := 0
	theme.OnUse(func() { runs++ })
	if runs != 0 {
		t.Fatalf("OnUse ran its hook at registration time")
	}
	mustUse(t, theme.TokyoNight())
	if runs != 1 {
		t.Fatalf("the hook ran %d times after one Use, want 1", runs)
	}
	mustUse(t, theme.TokyoNight())
	if runs != 2 {
		t.Fatalf("the hook ran %d times after two Uses, want 2", runs)
	}
}
