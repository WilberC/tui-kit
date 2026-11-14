package spinner_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/spinner"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

func TestRenderCyclesTheFrames(t *testing.T) {
	if len(spinner.Frames) < 2 {
		t.Fatalf("the spinner needs a cycle, got %d frames", len(spinner.Frames))
	}
	for index, want := range spinner.Frames {
		if got := text.Plain(spinner.Render(index)); got != want {
			t.Fatalf("Render(%d) = %q, want %q", index, got, want)
		}
	}
	if got, want := spinner.Render(0), spinner.Render(len(spinner.Frames)); got != want {
		t.Fatalf("the cycle does not wrap: %q and %q", got, want)
	}
	for _, frame := range spinner.Frames {
		if got, want := text.Width(frame), 1; got != want {
			t.Fatalf("frame %q is %d cells wide, want %d", frame, got, want)
		}
	}
}

func TestRenderUsesTheWorkingAccent(t *testing.T) {
	before := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(before) })

	paint := strings.SplitN(theme.Teal.Render("x"), "x", 2)[0]
	if !strings.Contains(spinner.Render(1), paint) {
		t.Fatalf("the spinner is not on the working accent: %q", spinner.Render(1))
	}
}
