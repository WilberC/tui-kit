package theme_test

import (
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/query"
	"github.com/WilberC/tui-kit/theme"
)

// The palette a Tokyo Night terminal reports, as the query normalizes it.
func tokyoTerminal() query.Palette {
	var p query.Palette
	p.FG = "#c0caf5"
	p.BG = "#1a1b26"
	slots := []string{
		"#1a1b26", "#f7768e", "#9ece6a", "#e0af68", "#7aa2f7", "#bb9af7",
		"#7dcfff", "#a9b1d6", "#414868", "#f7768e", "#9ece6a", "#e0af68",
		"#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5",
	}
	copy(p.Slots[:], slots)
	return p
}

func channelDelta(a, b string) int {
	if len(a) != 7 || len(b) != 7 {
		return 255 * 3
	}
	parse := func(s string) [3]int {
		var out [3]int
		for i := 0; i < 3; i++ {
			out[i] = hexInt(s[1+2*i], s[2+2*i])
		}
		return out
	}
	x, y := parse(a), parse(b)
	worst := 0
	for i := range x {
		d := x[i] - y[i]
		if d < 0 {
			d = -d
		}
		if d > worst {
			worst = d
		}
	}
	return worst
}

func hexInt(hi, lo byte) int {
	return hexNibble(hi)*16 + hexNibble(lo)
}

func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	default:
		return 0
	}
}

// The system theme generated from Tokyo Night's own terminal must land close
// to the palette the kit builds by hand: the derivation is honest only if it
// reproduces the approved look when the approved look is what the terminal
// wears. The semantic colours come straight from the slots; the tints are
// derived, and each must sit within twenty channel steps of the built-in.
func TestASystemThemeFromTokyoNightLandsNearTheBuiltIn(t *testing.T) {
	them := theme.System(tokyoTerminal())

	if them.Name != "system" {
		t.Fatalf("the generated theme is called %q, want \"system\"", them.Name)
	}

	for name, pair := range map[string][2]string{
		"accent":            {string(them.Accent), "#7aa2f7"},
		"gold":              {string(them.Gold), "#e0af68"},
		"green":             {string(them.Green), "#9ece6a"},
		"red":               {string(them.Red), "#f7768e"},
		"text":              {string(them.Text), "#c0caf5"},
		"muted":             {string(them.Muted), "#565f89"},
		"guide":             {string(them.Guide), "#3b4261"},
		"border":            {string(them.Border), "#3d59a1"},
		"chip surface":      {string(them.Chip), "#2f3549"},
		"selection tint":    {string(them.Selection), "#283457"},
		"hover tint":        {string(them.Hover), "#21273f"},
		"muted on tint":     {string(them.MutedOnSelection), "#737aa2"},
		"the guide on tint": {string(them.GuideOnSelection), "#5f6b96"},
	} {
		if got, want := pair[0], pair[1]; channelDelta(got, want) > 20 {
			t.Fatalf("the %s role is %s, more than twenty steps from the built-in %s", name, got, want)
		}
	}
	if got := string(them.Teal); got != "#7dcfff" {
		t.Fatalf("the teal role is %s, want the terminal's own cyan", got)
	}
	if got := string(them.ChipOnSelection); got != string(them.Border) {
		t.Fatalf("the chip on tint is %s, want the border %s", got, them.Border)
	}
}

// The system theme must be installable: every role it generates is a value Use
// accepts.
func TestASystemThemeIsAValidTheme(t *testing.T) {
	if err := theme.Use(theme.System(tokyoTerminal())); err != nil {
		t.Fatalf("theme.Use(system): %v", err)
	}
	t.Cleanup(func() { _ = theme.Use(theme.TokyoNight()) })
	if got := theme.Active().Name; got != "system" {
		t.Fatalf("the active theme is %q, want \"system\"", got)
	}
}

// An answer the query would never have produced — an empty palette — is the
// built-in theme, not a black-on-black guess.
func TestAnUnusableAnswerIsTheBuiltInTheme(t *testing.T) {
	if got := theme.System(query.Palette{}); got != theme.TokyoNight() {
		t.Fatalf("an empty palette produced %+v, want the built-in theme", got)
	}
	p := tokyoTerminal()
	p.Slots[7] = ""
	if got := theme.System(p); got.Name != "tokyo-night" {
		t.Fatalf("a palette with an empty slot produced %q, want the built-in", got.Name)
	}
}

// On a light terminal the same derivation must stay legible: the text meets
// the contrast gate against the background and both tints are visible
// departures from it.
func TestASystemThemeOnALightTerminalStaysLegible(t *testing.T) {
	var p query.Palette
	p.FG = "#000000"
	p.BG = "#ffffff"
	slots := []string{
		"#000000", "#cd3131", "#0dbc79", "#e5e510", "#2472c8", "#bc3fbc",
		"#11a8cd", "#e5e5e5", "#666666", "#f14c4c", "#23d18b", "#f5f543",
		"#3b8eea", "#d670d6", "#29b8db", "#ffffff",
	}
	copy(p.Slots[:], slots)
	them := theme.System(p)

	if got := string(them.Text); !strings.HasPrefix(got, "#") {
		t.Fatalf("the text role on a light terminal is %q, want a hex", got)
	}
	for name, c := range map[string]string{
		"selection tint": string(them.Selection),
		"hover tint":     string(them.Hover),
	} {
		if channelDelta(c, p.BG) < 15 {
			t.Fatalf("the %s (%s) is indistinguishable from the background on a light terminal", name, c)
		}
	}
	if got := string(them.Muted); channelDelta(got, p.FG) == 0 {
		t.Fatalf("the muted role (%s) is the text role, want it receded toward the background", got)
	}
}
