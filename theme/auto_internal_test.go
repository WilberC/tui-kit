package theme

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/WilberC/tui-kit/query"
)

// Auto's resolution order is the decided policy, so it is pinned here with an
// injected terminal: a named setting wins, a failing one falls back to the
// built-in, and an empty setting asks the terminal — whose failure is also
// the built-in.
func TestAutoPrefersTheNamedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gruvbox.json"), []byte(`{"name": "gruvbox"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	them := Auto("gruvbox", dir)
	if them.Name != "gruvbox" {
		t.Fatalf("Auto produced %q, want the file's own theme", them.Name)
	}
}

func TestAFailingSettingFallsBackToTheBuiltIn(t *testing.T) {
	them := Auto("anything", t.TempDir())
	if them != TokyoNight() {
		t.Fatalf("Auto with no file to load produced %+v, want the built-in", them)
	}
}

func TestAutoWithNoSettingAsksTheTerminal(t *testing.T) {
	p := TokyoTerminalPalette()
	them := auto("", t.TempDir(), func() (query.Palette, error) { return p, nil })
	if them.Name != "system" {
		t.Fatalf("Auto with an empty setting produced %q, want \"system\"", them.Name)
	}
	if them != System(p) {
		t.Fatalf("Auto built %+v, want System's %+v", them, System(p))
	}
}

func TestAutoFallsBackWhenTheTerminalCannotBeAsked(t *testing.T) {
	them := auto("", t.TempDir(), func() (query.Palette, error) {
		return query.Palette{}, errors.New("no tty")
	})
	if them != TokyoNight() {
		t.Fatalf("Auto with an unanswerable terminal produced %+v, want the built-in", them)
	}
}

// TokyoTerminalPalette is the palette a Tokyo Night terminal reports; the
// external system_test.go builds its own copy for its golden test.
func TokyoTerminalPalette() query.Palette {
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
