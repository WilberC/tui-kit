package theme_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

// The palette is the approved look, so it is pinned here as values, not as a
// description: a role that moves is a redesign and must be reviewed as one.
func TestThePaletteIsTheApprovedTokyonightSet(t *testing.T) {
	cases := map[string]struct {
		got  lipgloss.Color
		want string
	}{
		"text":              {theme.ColorText, "#c0caf5"},
		"muted":             {theme.ColorMuted, "#565f89"},
		"guide":             {theme.ColorGuide, "#3b4261"},
		"border":            {theme.ColorBorder, "#3d59a1"},
		"accent":            {theme.ColorAccent, "#7aa2f7"},
		"gold":              {theme.ColorGold, "#e0af68"},
		"teal":              {theme.ColorTeal, "#73daca"},
		"green":             {theme.ColorGreen, "#9ece6a"},
		"red":               {theme.ColorRed, "#f7768e"},
		"chip surface":      {theme.ColorChip, "#2f3549"},
		"selection tint":    {theme.ColorSelection, "#283457"},
		"chip on tint":      {theme.ColorChipOnSelection, "#3d59a1"},
		"muted on tint":     {theme.ColorMutedOnSelection, "#737aa2"},
		"the guide on tint": {theme.ColorGuideOnSelection, "#5f6b96"},
	}
	for name, c := range cases {
		if string(c.got) != c.want {
			t.Fatalf("the %s role is %q, want %q", name, c.got, c.want)
		}
	}
}

// TestNoColourLivesOutsideTheTheme pins the rule that keeps the palette in one
// place: render code names a role, never a hex literal.
func TestNoColourLivesOutsideTheTheme(t *testing.T) {
	hex := regexp.MustCompile(`#[0-9a-fA-F]{6}\b`)
	root := ".."
	checked := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "theme", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		checked++
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if match := hex.Find(source); match != nil {
			t.Fatalf("%s names the colour %q; a role belongs in theme/", path, match)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the module: %v", err)
	}
	if checked == 0 {
		t.Fatal("the hex scan checked no file, so it cannot judge the rule")
	}
}

func sgr(style lipgloss.Style) string {
	rendered := style.Render("x")
	if i := strings.Index(rendered, "x"); i > 0 {
		return rendered[:i]
	}
	return rendered
}

func withColors(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
}

func TestOnSelectionKeepsTheRoleAndAddsTheTint(t *testing.T) {
	withColors(t)
	lifted := sgr(theme.OnSelection(theme.Gold))

	if !strings.Contains(lifted, "38;2;224;175;104") {
		t.Fatalf("theme.OnSelection lost the role's own colour: %q", lifted)
	}
	if !strings.Contains(lifted, `48;2;40;52;87`) {
		t.Fatalf("theme.OnSelection did not add the selection tint: %q", lifted)
	}
}

func TestDotIsAQuietSeparator(t *testing.T) {
	if got, want := text.Plain(theme.Dot()), " · "; got != want {
		t.Fatalf("theme.Dot = %q, want %q", got, want)
	}
	withColors(t)
	if !strings.Contains(theme.Dot(), sgr(theme.Quiet)) {
		t.Fatalf("the separator does not carry the guide role: %q", theme.Dot())
	}
}

func TestLabelPairsADimNameWithItsValue(t *testing.T) {
	withColors(t)
	label := theme.Label("workspace", "work", theme.AccentBold)

	if got, want := text.Plain(label), "workspace work"; got != want {
		t.Fatalf("theme.Label = %q, want %q", got, want)
	}
	if !strings.Contains(label, sgr(theme.Muted)) || !strings.Contains(label, sgr(theme.AccentBold)) {
		t.Fatalf("theme.Label lost one of its two roles: %q", label)
	}
}

// TestRolesAreRenderedAtCallTime pins that no role string is frozen at init.
func TestRolesAreRenderedAtCallTime(t *testing.T) {
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })

	if strings.ContainsRune(theme.Text.Render("x"), 0x1b) {
		t.Fatalf("an Ascii role carries escapes: %q", theme.Text.Render("x"))
	}
	lipgloss.SetColorProfile(termenv.TrueColor)
	if !strings.ContainsRune(theme.Text.Render("x"), 0x1b) {
		t.Fatalf("the role stayed plain after the profile changed: %q", theme.Text.Render("x"))
	}
}
