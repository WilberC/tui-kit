// Command demo renders one tui-kit surface — a framed popup with groups, chips,
// a tinted selected row and a footer — at the width and height you ask for, so
// the look can be checked by eye or captured as an image. The pane is clamped to
// the recommended size range, so a tall or wide pane shows the clamped frame
// centred, with its footer pinned inside it.
//
//	go run ./examples/demo -w 140 -h 40
//	go run ./examples/demo -w 180 -o /tmp/demo-180.ansi
//	go run ./examples/demo -w 60 -h 30 -page 1   # a narrow bar pages its hints
//	go run ./examples/demo -w 140 -h 50          # a tall pane: the frame is centred
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/frame"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

// maxChipWidth and minDescWidth keep the chip column readable at 80 columns: the
// column never eats the description's room.
const (
	maxChipWidth = 14
	minDescWidth = 8
)

// entry is one listed row: a key, what it runs, and whether the entry declares a
// key of its own.
type entry struct {
	key         string
	description string
	unset       bool
}

// group is one column section: a name and the entries it holds.
type group struct {
	name    string
	entries []entry
}

// demo data: a full first group, two small ones, and a group of discovered
// actions that declare no key of their own.
var groups = []group{
	{name: "herdr", entries: []entry{
		{key: "prefix+k", description: "Command bar"},
		{key: "prefix+p", description: "Plugin manager"},
		{key: "prefix+i", description: "Keybinding guide"},
		{key: "prefix+.", description: "Edit the config"},
		{key: "prefix+g", description: "Git status in a popup"},
		{key: "prefix+r", description: "Restart the plugin runtime"},
	}},
	{name: "gentle", entries: []entry{
		{key: "ctrl+g", description: "Lazygit"},
		{key: "ctrl+f", description: "Fuzzy find files"},
		{key: "alt+g", description: "Git graph"},
	}},
	{name: "discovered", entries: []entry{
		{key: text.Unset, description: "Open the bar", unset: true},
		{key: text.Unset, description: "Restore the last session", unset: true},
		{key: text.Unset, description: "Run the health check", unset: true},
	}},
}

// demoHints are the footer's keys, in priority order: the bar pages them when a
// narrow window cannot hold them all.
var demoHints = []chip.Hint{
	{Key: "enter", Label: "run"},
	{Key: "↑↓", Label: "move"},
	{Key: "type", Label: "filter"},
	{Key: "ctrl+g", Label: "git status"},
	{Key: "alt+enter", Label: "create tab"},
	{Key: "esc", Label: "close"},
}

func main() {
	width := flag.Int("w", 140, "width of the popup in cells")
	height := flag.Int("h", 40, "height of the popup in rows")
	output := flag.String("o", "", "write the rendered ANSI to this file instead of stdout")
	plain := flag.Bool("plain", false, "render without colour, the way a dumb terminal sees it")
	page := flag.Int("page", 0, "hint bar page to draw; a page past the end wraps")
	flag.Parse()

	// The demo is also a capture tool: a pipe has no terminal, so force the
	// truecolor profile instead of letting lipgloss fall back to Ascii.
	lipgloss.SetColorProfile(termenv.TrueColor)
	if *plain {
		lipgloss.SetColorProfile(termenv.Ascii)
	}

	view := renderPage(*width, *height, *page)
	if *output == "" {
		fmt.Print(view)
		return
	}
	if err := os.WriteFile(*output, []byte(view), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "demo: write %s: %v\n", *output, err)
		os.Exit(1)
	}
}

// render draws the whole popup: one frame, two notes at the top, the columns the
// width earns, and the two footer lines.
func render(width, height int) string { return renderPage(width, height, 0) }

// renderPage draws the popup showing one page of the hint bar, so a narrow window
// can be captured on every page it holds. The pane is clamped to the recommended
// size range, so a pane taller than the frame leaves the frame centred instead of
// hugging its top, and the frame itself is always as tall as the clamped pane: the
// footer is pinned to its bottom edge and the body keeps one blank row at each end.
func renderPage(paneWidth, paneHeight, page int) string {
	size := frame.Recommended
	f := frame.Frame{
		Title: theme.GoldBold.Render("tui-kit") + theme.Quiet.Render(" › ") + theme.AccentBold.Render("demo"),
		Width: size.ClampWidth(paneWidth),
	}
	inner := f.Width - 2

	header := text.Join(
		theme.Muted.Render("3 groups")+theme.Dot()+theme.AccentBold.Render("12 bindings"),
		theme.TealBold.Render("KEYS")+theme.Dot()+theme.Muted.Render("LIST")+theme.Dot()+theme.Muted.Render("12 / 12 shown"),
		f.ContentWidth(),
	)

	filter := text.Join(
		theme.Accent.Render("❱ ")+theme.Muted.Render("Filter keys"+text.Ellipsis),
		chip.Legend(
			chip.Pair(theme.ChipText, "x", "after prefix"),
			chip.Pair(theme.ChipText, "ctrl+x", "direct"),
			theme.Muted.Render(text.Unset)+" "+theme.Muted.Render("unset"),
		),
		f.ContentWidth(),
	)

	count := frame.ColumnCount(inner, 2)
	widths := frame.ColumnWidths(inner, count)
	blocks := columnBlocks(widths)

	detail := frame.Footer(
		theme.AccentBold.Render("Command bar")+theme.Dot()+
			theme.Green.Render("prefix+k")+theme.Dot()+
			theme.Muted.Render("herdr"),
		theme.Muted.Render("plugin action"),
		f.ContentWidth(),
	)

	// Six hints do not fit below 80 columns, so a narrow window shows the paging
	// chip and the rest of the bar behind PageKey.
	hints := chip.HintBar{Page: page}.Render(f.ContentWidth(), demoHints...)

	surface := frame.Surface{
		Frame:   f,
		Size:    size,
		Columns: widths,
		Top: []string{
			f.Row(header),
			f.Rule(),
			f.Row(filter),
			frame.ColumnRule(widths, frame.TopTee),
		},
		Bottom: []string{
			frame.ColumnRule(widths, frame.BottomTee),
			f.Row(detail),
			f.Row(hints),
		},
	}

	// The body is exactly the rows Surface says it has, so the grid never decides
	// the frame's height; the surface pads it at both ends.
	body := frame.Grid(f, blocks, widths, surface.BodyRows(paneHeight))
	return surface.Popup(paneWidth, paneHeight, body)
}

// columnBlocks pours the groups into one block per column: the chip column is
// shared inside a column, the last entry of a group closes its guide, and one
// blank row separates two groups so the eye groups them without a rule.
func columnBlocks(widths []int) [][]string {
	blocks := make([][]string, len(widths))
	perColumn := (len(groups) + len(widths) - 1) / len(widths)
	position := 0 // flat index of the entry, so the demo can select the first one

	for index, columnWidth := range widths {
		start := index * perColumn
		end := min(start+perColumn, len(groups))
		if start >= len(groups) {
			blocks[index] = nil
			continue
		}

		labels := make([]string, 0, len(groups[start].entries))
		for _, g := range groups[start:end] {
			for _, e := range g.entries {
				labels = append(labels, e.key)
			}
		}
		chipWidth := min(chip.Width(labels...), maxChipWidth)
		chipWidth = max(chipWidth, min(chip.Width(text.Unset), columnWidth-4-minDescWidth))

		lines := []string{}
		for groupIndex, g := range groups[start:end] {
			if groupIndex > 0 {
				lines = append(lines, strings.Repeat(" ", columnWidth))
			}
			lines = append(lines, list.Section(g.name, len(g.entries)))
			for entryIndex, e := range g.entries {
				roles := list.RowRoles(position == 0)
				cell, description := chip.Cell(roles.Chip, e.key, chipWidth), roles.Name.Render(e.description)
				if e.unset {
					// An entry with no key keeps the column with a dim dash and no
					// slab, and its description recedes with it.
					cell, description = chip.UnsetCell(roles.Muted, e.key, chipWidth), roles.Muted.Render(e.description)
				}
				lines = append(lines, list.ChipRow(roles, entryIndex == len(g.entries)-1, cell, description, columnWidth))
				position++
			}
		}
		blocks[index] = lines
	}
	return blocks
}
