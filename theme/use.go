package theme

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Theme is one resolution of the palette: a name and the fifteen roles every
// surface paints with. The built-in resolution is TokyoNight, the palette
// DESIGN.md §1 documents; Use installs another one, a system theme generated
// from the terminal's own colours or a custom one loaded from a file, and
// every style in the kit renders with it from the next render on.
//
// A role value is a `#rgb`/`#rrggbb` hex, a 256-colour index written as a
// decimal string, or the empty string: the empty role is the terminal's own
// default, and renders no colour escape at all.
type Theme struct {
	// Name identifies the theme in a picker, a log, a dump.
	Name string
	// Text is prose: names, values, a row's primary text.
	Text lipgloss.Color
	// Muted is explanation: every secondary fact beside a primary one.
	Muted lipgloss.Color
	// Guide is structure: tree guides, separators, quiet rules.
	Guide lipgloss.Color
	// Border is the frame's structural blue, one step under the accent.
	Border lipgloss.Color
	// Accent is the product's voice: its name, a focused thing, a label.
	Accent lipgloss.Color
	// Gold marks state that outlives the selection.
	Gold lipgloss.Color
	// Teal marks work in progress.
	Teal lipgloss.Color
	// Green marks live things.
	Green lipgloss.Color
	// Red marks failure.
	Red lipgloss.Color
	// Chip is the dark surface a token sits on.
	Chip lipgloss.Color
	// Selection is the tint a selected row gains instead of a painted bar.
	Selection lipgloss.Color
	// Hover is the tint a hovered row gains, fainter than the selection.
	Hover lipgloss.Color
	// ChipOnSelection is the chip surface one step brighter, on the tint.
	ChipOnSelection lipgloss.Color
	// MutedOnSelection lifts the muted role enough to read on the tint.
	MutedOnSelection lipgloss.Color
	// GuideOnSelection lifts the guide role so a tree guide stays readable.
	GuideOnSelection lipgloss.Color
}

// TokyoNight is the built-in theme: the approved palette the kit ships with,
// the one every capture in DESIGN.md was taken with, and the fallback when
// the terminal cannot be asked for its colours.
func TokyoNight() Theme {
	return Theme{
		Name:             "tokyo-night",
		Text:             ColorText,
		Muted:            ColorMuted,
		Guide:            ColorGuide,
		Border:           ColorBorder,
		Accent:           ColorAccent,
		Gold:             ColorGold,
		Teal:             ColorTeal,
		Green:            ColorGreen,
		Red:              ColorRed,
		Chip:             ColorChip,
		Selection:        ColorSelection,
		Hover:            ColorHover,
		ChipOnSelection:  ColorChipOnSelection,
		MutedOnSelection: ColorMutedOnSelection,
		GuideOnSelection: ColorGuideOnSelection,
	}
}

var (
	mu       sync.Mutex
	active   = TokyoNight()
	useHooks []func()
)

// Active is the theme every style in the kit renders with.
func Active() Theme {
	mu.Lock()
	defer mu.Unlock()
	return active
}

// Use makes them the theme every style in the kit renders with, rebuilding the
// role styles and every surface that captured roles at init. A theme whose
// roles are neither a hex, a 256 index nor the empty string is rejected, and
// the previous theme stays in force. Use is safe to call once at start-up and
// again when terminal colours arrive late; it is not safe to call from the
// middle of a render.
func Use(them Theme) error {
	if err := them.validate(); err != nil {
		return err
	}
	if them.Name == "" {
		them.Name = "custom"
	}
	mu.Lock()
	active = them
	rebuild()
	hooks := make([]func(), len(useHooks))
	copy(hooks, useHooks)
	mu.Unlock()
	for _, hook := range hooks {
		hook()
	}
	return nil
}

// OnUse registers hook to run after every successful Use, in registration
// order. A package that builds its surfaces from the roles at init registers
// here, so a theme installed after its init still reaches them.
func OnUse(hook func()) {
	mu.Lock()
	defer mu.Unlock()
	useHooks = append(useHooks, hook)
}

// rebuild rebuilds every role style from the active theme. It runs at init
// and again inside Use, already under the lock.
func rebuild() {
	Text = lipgloss.NewStyle().Foreground(active.Text)
	Strong = Text.Bold(true)
	Muted = lipgloss.NewStyle().Foreground(active.Muted)
	Quiet = lipgloss.NewStyle().Foreground(active.Guide)
	Border = lipgloss.NewStyle().Foreground(active.Border)
	Accent = lipgloss.NewStyle().Foreground(active.Accent)
	AccentBold = Accent.Bold(true)
	Gold = lipgloss.NewStyle().Foreground(active.Gold)
	GoldBold = Gold.Bold(true)
	Teal = lipgloss.NewStyle().Foreground(active.Teal)
	TealBold = Teal.Bold(true)
	Green = lipgloss.NewStyle().Foreground(active.Green)
	GreenBold = Green.Bold(true)
	Red = lipgloss.NewStyle().Foreground(active.Red)
	RedBold = Red.Bold(true)

	SelectionBG = lipgloss.NewStyle().Background(active.Selection)
	SelectionEdge = SelectionBG.Foreground(active.Accent).Bold(true)
	TextOnSelection = SelectionBG.Foreground(active.Text).Bold(true)
	MutedOnSelection = SelectionBG.Foreground(active.MutedOnSelection)
	GuideOnSelection = SelectionBG.Foreground(active.GuideOnSelection)
	ChipOnSelection = lipgloss.NewStyle().
		Background(active.ChipOnSelection).
		Foreground(active.Text).
		Bold(true)
	ChipAccentOnSelection = lipgloss.NewStyle().
		Background(active.ChipOnSelection).
		Foreground(active.Accent).
		Bold(true)

	ChipText = chipSurface(active.Chip, active.Text)
	ChipAccent = chipSurface(active.Chip, active.Accent)
	ChipGold = chipSurface(active.Chip, active.Gold)
	ChipTeal = chipSurface(active.Chip, active.Teal)
	ChipGreen = chipSurface(active.Chip, active.Green)
	ChipRed = chipSurface(active.Chip, active.Red)

	HoverBG = lipgloss.NewStyle().Background(active.Hover)
	TextOnHover = HoverBG.Foreground(active.Text).Bold(true)
	MutedOnHover = HoverBG.Foreground(active.Muted)
	GuideOnHover = HoverBG.Foreground(active.Guide)
}

func init() { rebuild() }

func (t Theme) validate() error {
	for name, c := range map[string]lipgloss.Color{
		"text":              t.Text,
		"muted":             t.Muted,
		"guide":             t.Guide,
		"border":            t.Border,
		"accent":            t.Accent,
		"gold":              t.Gold,
		"teal":              t.Teal,
		"green":             t.Green,
		"red":               t.Red,
		"chip surface":      t.Chip,
		"selection tint":    t.Selection,
		"hover tint":        t.Hover,
		"chip on tint":      t.ChipOnSelection,
		"muted on tint":     t.MutedOnSelection,
		"the guide on tint": t.GuideOnSelection,
	} {
		if !validColor(c) {
			return fmt.Errorf("the %s role is %q: a role is a #rgb or #rrggbb hex, a 0-255 index, or empty for the terminal's own default", name, string(c))
		}
	}
	return nil
}

// validColor accepts what the palette promises: empty (the terminal's own
// default), a #rgb or #rrggbb hex, or a decimal 0-255 index.
func validColor(c lipgloss.Color) bool {
	s := string(c)
	if s == "" {
		return true
	}
	if strings.HasPrefix(s, "#") {
		digits := s[1:]
		if len(digits) != 3 && len(digits) != 6 {
			return false
		}
		for _, r := range digits {
			if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
				return false
			}
		}
		return true
	}
	v, err := strconv.Atoi(s)
	return err == nil && v >= 0 && v <= 255
}
