package theme

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/WilberC/tui-kit/query"
)

// System builds a theme from the colours the terminal reported about itself:
// the semantic colours come from the ANSI slots (the brighter of each pair),
// the text is the terminal's foreground once it clears the contrast gate
// against its own background, and every surface role is a mix of the
// background toward the text or the accent, in the same proportions the
// built-in Tokyo Night palette sits at — feeding System the palette a Tokyo
// Night terminal reports lands within twenty channel steps of the hand-built
// palette, which is the golden test system_test.go pins.
//
// An answer the query would never have produced — a missing foreground,
// background or slot — is not guessed at: it is the built-in theme, so a
// caller that falls back on error and a caller that installs System directly
// end in the same place.
func System(p query.Palette) Theme {
	them := TokyoNight()

	if !usable(p) {
		return them
	}
	them.Name = "system"
	bg, bgOK := parseHex(p.BG)
	fg, fgOK := parseHex(p.FG)
	if !bgOK || !fgOK {
		return them
	}

	text, textOK := fg, true
	if contrast(fg, bg) < 4.5 {
		if readable, ok := pushToContrast(fg, bg); ok {
			text = readable
		} else {
			textOK = false
		}
	}

	muted := mix(bg, text, 0.40)
	guide := mix(bg, text, 0.22)
	accent, accentOK := parseHex(brighter(p.Slots[4], p.Slots[12]))
	accentRGB := accent
	if !accentOK {
		accentRGB = mix(bg, text, 0.35)
	}

	if textOK {
		them.Text = lipgloss.Color(hexOf(text))
	} else {
		them.Text = ""
	}
	them.Muted = lipgloss.Color(hexOf(muted))
	them.Guide = lipgloss.Color(hexOf(guide))
	them.Border = lipgloss.Color(hexOf(mix(bg, accentRGB, 0.55)))
	them.Accent = lipgloss.Color(hexOf(accentRGB))
	them.Gold = lipgloss.Color(brighter(p.Slots[3], p.Slots[11]))
	them.Teal = lipgloss.Color(brighter(p.Slots[6], p.Slots[14]))
	them.Green = lipgloss.Color(brighter(p.Slots[2], p.Slots[10]))
	them.Red = lipgloss.Color(brighter(p.Slots[1], p.Slots[9]))
	them.Chip = lipgloss.Color(hexOf(mix(bg, text, 0.15)))
	them.Selection = lipgloss.Color(hexOf(mix(bg, accentRGB, 0.19)))
	them.Hover = lipgloss.Color(hexOf(mix(bg, text, 0.08)))
	them.ChipOnSelection = them.Border
	them.MutedOnSelection = lipgloss.Color(hexOf(mix(muted, text, 0.25)))
	them.GuideOnSelection = lipgloss.Color(hexOf(mix(guide, text, 0.31)))
	return them
}

// usable mirrors the query's all-or-nothing rule: the derivation never works
// from a half answer.
func usable(p query.Palette) bool {
	if p.FG == "" || p.BG == "" {
		return false
	}
	for _, slot := range p.Slots {
		if slot == "" {
			return false
		}
	}
	return true
}

// brighter picks the more luminous of a slot's two ANSI entries, so a dark
// base slot never darkens a theme whose bright variant is the vivid one.
func brighter(a, b string) string {
	ca, caOK := parseHex(a)
	cb, cbOK := parseHex(b)
	switch {
	case !caOK:
		return b
	case !cbOK:
		return a
	case luminance(ca) >= luminance(cb):
		return a
	default:
		return b
	}
}

// pushToContrast moves the foreground away from the background — toward white
// on a dark one, toward black on a light one — until the text clears the
// 4.5:1 gate, and reports failure when twenty steps are not enough.
func pushToContrast(fg, bg rgb) (rgb, bool) {
	target := rgb{255, 255, 255}
	if luminance(bg) > 0.5 {
		target = rgb{0, 0, 0}
	}
	out := fg
	for i := 0; i < 20 && contrast(out, bg) < 4.5; i++ {
		out = mix(out, target, 0.15)
	}
	return out, contrast(out, bg) >= 4.5
}

type rgb [3]float64

func parseHex(s string) (rgb, bool) {
	if !strings.HasPrefix(s, "#") || len(s) != 7 {
		return rgb{}, false
	}
	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return rgb{}, false
	}
	return rgb{float64(v >> 16 & 0xff), float64(v >> 8 & 0xff), float64(v & 0xff)}, true
}

func hexOf(c rgb) string {
	return fmt.Sprintf("#%02x%02x%02x", channelByte(c[0]), channelByte(c[1]), channelByte(c[2]))
}

func channelByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}

// mix walks from a toward b by f, in sRGB — the proportions the derivation
// uses were measured on the built-in palette, so a linear walk in display
// space is exactly what the golden test judges.
func mix(a, b rgb, f float64) rgb {
	return rgb{a[0] + f*(b[0]-a[0]), a[1] + f*(b[1]-a[1]), a[2] + f*(b[2]-a[2])}
}

// luminance is the WCAG relative luminance, and contrast the WCAG ratio the
// legibility gates use.
func luminance(c rgb) float64 {
	lin := func(v float64) float64 {
		x := v / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c[0]) + 0.7152*lin(c[1]) + 0.0722*lin(c[2])
}

func contrast(a, b rgb) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}
