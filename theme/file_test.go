package theme_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/theme"
)

const gruvboxJSON = `{
  "name": "gruvbox",
  "colors": {
    "text": "#ebdbb2",
    "accent": 196,
    "chip": ""
  }
}`

// A file is a partial override: every role it names replaces the built-in,
// every role it leaves untouched stays Tokyo Night.
func TestAFileOverridesOnlyTheRolesItNames(t *testing.T) {
	them, err := theme.LoadJSON([]byte(gruvboxJSON))
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	if them.Name != "gruvbox" {
		t.Fatalf("the file's theme is called %q, want \"gruvbox\"", them.Name)
	}
	if got := string(them.Text); got != "#ebdbb2" {
		t.Fatalf("the text role is %q, want #ebdbb2", got)
	}
	if got := string(them.Accent); got != "196" {
		t.Fatalf("the accent role is %q, want the index \"196\"", got)
	}
	if got := string(them.Chip); got != "" {
		t.Fatalf("the chip role is %q, want the terminal's default", got)
	}
	builtIn := theme.TokyoNight()
	for name, pair := range map[string][2]string{
		"muted":             {string(them.Muted), string(builtIn.Muted)},
		"guide":             {string(them.Guide), string(builtIn.Guide)},
		"border":            {string(them.Border), string(builtIn.Border)},
		"gold":              {string(them.Gold), string(builtIn.Gold)},
		"teal":              {string(them.Teal), string(builtIn.Teal)},
		"green":             {string(them.Green), string(builtIn.Green)},
		"red":               {string(them.Red), string(builtIn.Red)},
		"selection tint":    {string(them.Selection), string(builtIn.Selection)},
		"hover tint":        {string(them.Hover), string(builtIn.Hover)},
		"chip on tint":      {string(them.ChipOnSelection), string(builtIn.ChipOnSelection)},
		"muted on tint":     {string(them.MutedOnSelection), string(builtIn.MutedOnSelection)},
		"the guide on tint": {string(them.GuideOnSelection), string(builtIn.GuideOnSelection)},
	} {
		if got, want := pair[0], pair[1]; got != want {
			t.Fatalf("the %s role moved to %q without being named: want the built-in %q", name, got, want)
		}
	}
}

func TestALoadedThemeIsAValidTheme(t *testing.T) {
	them, err := theme.LoadJSON([]byte(gruvboxJSON))
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	if err := theme.Use(them); err != nil {
		t.Fatalf("theme.Use(loaded): %v", err)
	}
	t.Cleanup(func() { _ = theme.Use(theme.TokyoNight()) })
	if got := theme.Active().Name; got != "gruvbox" {
		t.Fatalf("the active theme is %q, want \"gruvbox\"", got)
	}
}

// A file with no colors block is the built-in palette under another name.
func TestAFileWithNoColorsIsTheBuiltInUnderAnotherName(t *testing.T) {
	them, err := theme.LoadJSON([]byte(`{"name": "tokyo-night-soft"}`))
	if err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	builtIn := theme.TokyoNight()
	if them.Text != builtIn.Text || them.Accent != builtIn.Accent {
		t.Fatalf("a colours-less file changed the palette: %+v", them)
	}
	if them.Name != "tokyo-night-soft" {
		t.Fatalf("the theme is called %q, want \"tokyo-night-soft\"", them.Name)
	}
}

func TestAnUnknownRoleIsAnError(t *testing.T) {
	_, err := theme.LoadJSON([]byte(`{"colors": {"acccent": "#ffffff"}}`))
	if err == nil || !strings.Contains(err.Error(), "acccent") {
		t.Fatalf("a mistyped role = %v, want an error naming it", err)
	}
}

func TestABrokenFileIsAnError(t *testing.T) {
	if _, err := theme.LoadJSON([]byte(`{"colors": `)); err == nil {
		t.Fatal("a broken JSON file loaded")
	}
}

func TestLoadFileReadsTheDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gruvbox.json")
	if err := os.WriteFile(path, []byte(gruvboxJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	them, err := theme.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if them.Name != "gruvbox" || string(them.Text) != "#ebdbb2" {
		t.Fatalf("LoadFile read %+v, want the gruvbox text role", them)
	}
	if _, err := theme.LoadFile(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("a missing file loaded")
	}
}
