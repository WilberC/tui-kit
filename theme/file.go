package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/query"
)

// LoadFile reads a custom theme from a JSON file:
//
//	{
//	  "name": "gruvbox",
//	  "colors": { "text": "#ebdbb2", "accent": 196, "chip": "" }
//	}
//
// A file is a partial override over the built-in Tokyo Night palette: every
// role it names replaces the built-in one, every role it leaves untouched
// stays Tokyo Night, and a role set to "" is the terminal's own default. A
// role's value is a hex string, a decimal index (quoted or not), or "" — the
// same values Use accepts. A role the file names that the palette does not
// have is an error, so a typo is caught instead of silently ignored.
func LoadFile(path string) (Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, fmt.Errorf("theme: %w", err)
	}
	them, err := LoadJSON(data)
	if err != nil {
		return Theme{}, fmt.Errorf("theme: %s: %w", path, err)
	}
	return them, nil
}

// LoadJSON parses a custom theme out of JSON bytes, with LoadFile's rules.
func LoadJSON(data []byte) (Theme, error) {
	them := TokyoNight()

	var file struct {
		Name   string         `json:"name"`
		Colors map[string]any `json:"colors"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return Theme{}, fmt.Errorf("broken JSON: %w", err)
	}
	them.Name = file.Name

	for key, raw := range file.Colors {
		set, ok := roleByName[key]
		if !ok {
			return Theme{}, fmt.Errorf("unknown role %q: the palette names text, muted, guide, border, accent, gold, teal, green, red, chip, selection, hover, chipOnSelection, mutedOnSelection and guideOnSelection", key)
		}
		value, err := roleValue(raw)
		if err != nil {
			return Theme{}, fmt.Errorf("the %s role: %w", key, err)
		}
		set(&them, value)
	}

	if err := them.validate(); err != nil {
		return Theme{}, err
	}
	return them, nil
}

// roleByName names the palette's roles and the field each one fills.
var roleByName = map[string]func(*Theme, lipgloss.Color){
	"text":             func(t *Theme, c lipgloss.Color) { t.Text = c },
	"muted":            func(t *Theme, c lipgloss.Color) { t.Muted = c },
	"guide":            func(t *Theme, c lipgloss.Color) { t.Guide = c },
	"border":           func(t *Theme, c lipgloss.Color) { t.Border = c },
	"accent":           func(t *Theme, c lipgloss.Color) { t.Accent = c },
	"gold":             func(t *Theme, c lipgloss.Color) { t.Gold = c },
	"teal":             func(t *Theme, c lipgloss.Color) { t.Teal = c },
	"green":            func(t *Theme, c lipgloss.Color) { t.Green = c },
	"red":              func(t *Theme, c lipgloss.Color) { t.Red = c },
	"chip":             func(t *Theme, c lipgloss.Color) { t.Chip = c },
	"selection":        func(t *Theme, c lipgloss.Color) { t.Selection = c },
	"hover":            func(t *Theme, c lipgloss.Color) { t.Hover = c },
	"chipOnSelection":  func(t *Theme, c lipgloss.Color) { t.ChipOnSelection = c },
	"mutedOnSelection": func(t *Theme, c lipgloss.Color) { t.MutedOnSelection = c },
	"guideOnSelection": func(t *Theme, c lipgloss.Color) { t.GuideOnSelection = c },
}

// roleValue accepts what Use accepts, plus an unquoted JSON number for an
// index: a role in a file may be written 196 or "196".
func roleValue(raw any) (lipgloss.Color, error) {
	switch v := raw.(type) {
	case string:
		return lipgloss.Color(v), nil
	case float64:
		if v != float64(int(v)) || v < 0 || v > 255 {
			return "", fmt.Errorf("%v: a colour is a hex, a 0-255 index or empty", v)
		}
		return lipgloss.Color(strconv.Itoa(int(v))), nil
	default:
		return "", fmt.Errorf("%v: a colour is a hex, a 0-255 index or empty", raw)
	}
}

// Auto resolves the theme the decided policy picks: a named setting loads
// <dir>/<setting>.json — the custom theme's file, in the directory the
// product keeps its themes in — an empty setting asks the terminal for its
// own colours and generates the system theme from the answer, and whatever
// fails along the way is the built-in Tokyo Night palette. The kit looks the
// same after every failure as it does without the call.
func Auto(setting, dir string) Theme {
	return auto(setting, dir, func() (query.Palette, error) { return query.Colors() })
}

func auto(setting, dir string, colors func() (query.Palette, error)) Theme {
	if setting != "" {
		if them, err := LoadFile(filepath.Join(dir, setting+".json")); err == nil {
			return them
		}
		return TokyoNight()
	}
	if p, err := colors(); err == nil {
		return System(p)
	}
	return TokyoNight()
}
