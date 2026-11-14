// Package theme is the tui-kit palette: the Tokyo Night roles every surface
// paints with, plus the few compositions that are pure role arithmetic.
//
// Render code names a role, never a colour: the palette lives here, and only
// here. A role exports its colour, for a caller that must build a style of its
// own, and a ready style, so the two cannot drift. Every role is rendered at
// call time, so the profile in force when a line is drawn decides its colour.
package theme

import "github.com/charmbracelet/lipgloss"

// Palette roles. Three accents carry the interface (blue, gold, teal); green
// means live, red means failure, and everything explanatory recedes to muted.
const (
	// ColorText is prose: names, values, a row's primary text.
	ColorText lipgloss.Color = "#c0caf5"
	// ColorMuted is explanation: every secondary fact beside a primary one.
	ColorMuted lipgloss.Color = "#565f89"
	// ColorGuide is structure: tree guides, separators, quiet rules.
	ColorGuide lipgloss.Color = "#3b4261"
	// ColorBorder is the frame's structural blue, one step under the accent.
	ColorBorder lipgloss.Color = "#3d59a1"
	// ColorAccent is the product's voice: its name, a focused thing, a label.
	ColorAccent lipgloss.Color = "#7aa2f7"
	// ColorGold marks state that outlives the selection: the default item, the
	// group glyph, a group name.
	ColorGold lipgloss.Color = "#e0af68"
	// ColorTeal marks work in progress.
	ColorTeal lipgloss.Color = "#73daca"
	// ColorGreen marks live things.
	ColorGreen lipgloss.Color = "#9ece6a"
	// ColorRed marks failure.
	ColorRed lipgloss.Color = "#f7768e"
	// ColorChip is the dark surface a token sits on.
	ColorChip lipgloss.Color = "#2f3549"
	// ColorSelection is the tint a selected row gains instead of a painted bar.
	ColorSelection lipgloss.Color = "#283457"
	// ColorHover is the tint a hovered row gains: one step above the terminal's
	// own background and still fainter than the selection, so a hover previews
	// the selection instead of competing with it.
	ColorHover lipgloss.Color = "#21273f"
	// ColorChipOnSelection is the chip surface one step brighter, for a token
	// that sits on the tint.
	ColorChipOnSelection lipgloss.Color = "#3d59a1"
	// ColorMutedOnSelection lifts the muted role enough to read on the tint.
	ColorMutedOnSelection lipgloss.Color = "#737aa2"
	// ColorGuideOnSelection lifts the guide role so a tree guide stays readable
	// on the tint without competing with the row's facts.
	ColorGuideOnSelection lipgloss.Color = "#5f6b96"
)

// Palette is every role colour in one slice, for a tool that must reproduce the
// palette exactly — a capture, a preview, a dump — instead of naming the roles one
// by one. It is the list DESIGN.md §1 documents.
func Palette() []lipgloss.Color {
	return []lipgloss.Color{
		ColorText, ColorMuted, ColorGuide, ColorBorder, ColorAccent,
		ColorGold, ColorTeal, ColorGreen, ColorRed, ColorChip,
		ColorSelection, ColorHover, ColorChipOnSelection,
		ColorMutedOnSelection, ColorGuideOnSelection,
	}
}

// Roles, as styles a view composes with. Text is prose, Muted is explanation,
// Quiet is structure, Border is the frame, Accent is the product's voice, Gold
// marks state that outlives the selection, Teal marks work in progress, Green
// marks live things and Red marks failure.
var (
	Text       = lipgloss.NewStyle().Foreground(ColorText)
	Strong     = Text.Bold(true)
	Muted      = lipgloss.NewStyle().Foreground(ColorMuted)
	Quiet      = lipgloss.NewStyle().Foreground(ColorGuide)
	Border     = lipgloss.NewStyle().Foreground(ColorBorder)
	Accent     = lipgloss.NewStyle().Foreground(ColorAccent)
	AccentBold = Accent.Bold(true)
	Gold       = lipgloss.NewStyle().Foreground(ColorGold)
	GoldBold   = Gold.Bold(true)
	Teal       = lipgloss.NewStyle().Foreground(ColorTeal)
	TealBold   = Teal.Bold(true)
	Green      = lipgloss.NewStyle().Foreground(ColorGreen)
	GreenBold  = Green.Bold(true)
	Red        = lipgloss.NewStyle().Foreground(ColorRed)
	RedBold    = Red.Bold(true)
)

// The selection surface. It is a tint, never a fill: the row keeps every element
// it has when it is not selected, lifted so each role still reads on the darker
// background.
var (
	// SelectionBG tints a selected row instead of painting over it, so every
	// element on the row keeps its own colour.
	SelectionBG = lipgloss.NewStyle().Background(ColorSelection)
	// SelectionEdge is the selected row's first cell: the accent mark on the
	// tint, never a fill over the row's content.
	SelectionEdge = SelectionBG.Foreground(ColorAccent).Bold(true)
	// TextOnSelection is the selected row's primary text: the text role, bold,
	// on the tint.
	TextOnSelection = SelectionBG.Foreground(ColorText).Bold(true)
	// MutedOnSelection is every secondary fact on the tint.
	MutedOnSelection = SelectionBG.Foreground(ColorMutedOnSelection)
	// GuideOnSelection is the tree structure on the tint.
	GuideOnSelection = SelectionBG.Foreground(ColorGuideOnSelection)
	// ChipOnSelection is the chip slab one step brighter, for a token that sits
	// on the tint.
	ChipOnSelection = lipgloss.NewStyle().
			Background(ColorChipOnSelection).
			Foreground(ColorText).
			Bold(true)
	// ChipAccentOnSelection is the accent chip's slab on the tint: the accent
	// label keeps its colour on the brighter token surface.
	ChipAccentOnSelection = lipgloss.NewStyle().
				Background(ColorChipOnSelection).
				Foreground(ColorAccent).
				Bold(true)
)

// Chips are the tokens of the interface: keys, modes and short state words. One
// dark surface, one role for the label, so a chip is always a chip.
var (
	ChipText   = chipSurface(ColorText)
	ChipAccent = chipSurface(ColorAccent)
	ChipGold   = chipSurface(ColorGold)
	ChipTeal   = chipSurface(ColorTeal)
	ChipGreen  = chipSurface(ColorGreen)
	ChipRed    = chipSurface(ColorRed)
)

// OnSelection lifts a role onto the selection tint. The role keeps its colour
// and gains the background, so green, gold, teal and red mean the same thing on
// a selected row as off it.
func OnSelection(role lipgloss.Style) lipgloss.Style {
	return role.Background(ColorSelection)
}

// The hover surface. A hovered row is where the pointer is, not where the user
// is: it wears the selection's composition in a fainter tint, so it reads as a
// preview of what a click would select. It never carries the accent edge.
var (
	// HoverBG tints a hovered row, one step above the panel and fainter than the
	// selection.
	HoverBG = lipgloss.NewStyle().Background(ColorHover)
	// TextOnHover is a hovered row's primary text.
	TextOnHover = HoverBG.Foreground(ColorText).Bold(true)
	// MutedOnHover is every secondary fact on the hover tint.
	MutedOnHover = HoverBG.Foreground(ColorMuted)
	// GuideOnHover is the tree structure on the hover tint.
	GuideOnHover = HoverBG.Foreground(ColorGuide)
)

// OnHover lifts a role onto the hover tint, the way OnSelection lifts it onto the
// selection: state keeps its meaning while the pointer is over it.
func OnHover(role lipgloss.Style) lipgloss.Style {
	return role.Background(ColorHover)
}

func chipSurface(fg lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(ColorChip).
		Foreground(fg).
		Bold(true)
}

// Dot separates two facts: a space on each side, in the guide role, so it
// recedes between the facts instead of competing with them.
func Dot() string { return Quiet.Render(" · ") }

// Label renders a dim label followed by its value, the label/value pair a header
// or a footer is built from.
func Label(label, value string, valueStyle lipgloss.Style) string {
	return Muted.Render(label) + " " + valueStyle.Render(value)
}
