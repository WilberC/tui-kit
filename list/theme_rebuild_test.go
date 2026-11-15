package list_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/theme"
)

func sgrOf(style lipgloss.Style) string {
	rendered := style.Render("x")
	if i := strings.Index(rendered, "x"); i > 0 {
		return rendered[:i]
	}
	return rendered
}

func withTrueColor(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
}

// The rows capture the theme's roles, so a theme installed after the package's
// init must reach them: Use rebuilds the sets every row paints with.
func TestListRolesFollowTheActiveTheme(t *testing.T) {
	withTrueColor(t)
	them := theme.TokyoNight()
	them.Text = "#abcdf0"
	them.Hover = "#010203"
	them.Selection = "#040506"
	them.Accent = "#778899"
	if err := theme.Use(them); err != nil {
		t.Fatalf("theme.Use: %v", err)
	}
	t.Cleanup(func() {
		if err := theme.Use(theme.TokyoNight()); err != nil {
			t.Fatalf("restore tokyo-night: %v", err)
		}
	})

	panel := list.RowRolesFor(list.State{})
	if !strings.Contains(sgrOf(panel.Name), "38;2;171;205;240") {
		t.Fatalf("the panel row kept the old text role: %q", sgrOf(panel.Name))
	}
	hover := list.RowRolesFor(list.State{Hovered: true})
	if !strings.Contains(sgrOf(hover.Space), "48;2;1;2;3") {
		t.Fatalf("the hovered row kept the old hover tint: %q", sgrOf(hover.Space))
	}
	if !strings.Contains(sgrOf(hover.Name), "38;2;171;205;240") {
		t.Fatalf("the hovered row kept the old text role: %q", sgrOf(hover.Name))
	}
	selection := list.RowRolesFor(list.State{Selected: true})
	if !strings.Contains(sgrOf(selection.Space), "48;2;4;5;6") {
		t.Fatalf("the selected row kept the old selection tint: %q", sgrOf(selection.Space))
	}
	if !strings.Contains(sgrOf(selection.Name), "38;2;171;205;240") {
		t.Fatalf("the selected row kept the old text role: %q", sgrOf(selection.Name))
	}
}
