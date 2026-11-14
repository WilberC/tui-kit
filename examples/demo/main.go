// Command demo renders one tui-kit surface — a framed popup with groups, chips,
// a selected row, the row the mouse hovers beside it, and a footer — at the width
// and height you ask for, so the look can be checked by eye or captured as an
// image. The pane is clamped to the recommended size range, so a tall or wide
// pane shows the clamped frame centred, with its footer pinned inside it.
//
// It is also the mouse example: every render records where it drew each row, each
// chip and the paging chip, so -zones prints what a click can hit and -at asks
// what one cell hits, without the demo measuring its own layout.
//
//	go run ./examples/demo -w 140 -h 40
//	go run ./examples/demo -w 180 -o /tmp/demo-180.ansi
//	go run ./examples/demo -w 60 -h 30 -page 1   # a narrow bar pages its hints
//	go run ./examples/demo -w 140 -h 50          # a tall pane: the frame is centred
//	go run ./examples/demo -w 80 -h 24 -zones     # what is clickable, and where
//	go run ./examples/demo -w 80 -h 24 -at 30,12  # what a click there hits
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/frame"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
	"github.com/WilberC/tui-kit/zone"
)

// maxChipWidth and minDescWidth keep the chip column readable at 80 columns: the
// column never eats the description's room.
const (
	maxChipWidth = 14
	minDescWidth = 8
)

// selectedRow is the row the demo highlights, and hoveredRow the row the mouse is
// over: next to the selection, so both tints can be judged side by side.
const (
	selectedRow = 0
	hoveredRow  = 1
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

// view is one rendered demo: the pane it paints and the cells its clickable
// elements landed on.
type view struct {
	pane  string
	zones *zone.Map
}

// rowZone is the id of one listed row, and chipZone the id of its chip slab: a
// product names its own rows this way and maps the id to an action.
func rowZone(index int) string  { return "row:" + strconv.Itoa(index) }
func chipZone(index int) string { return "chip:" + strconv.Itoa(index) }

// rowIndex is the entry index a row's zone id names.
func rowIndex(id string) (int, bool) {
	rest, ok := strings.CutPrefix(id, "row:")
	if !ok {
		return 0, false
	}
	index, err := strconv.Atoi(rest)
	return index, err == nil
}

func main() {
	width := flag.Int("w", 140, "width of the popup in cells")
	height := flag.Int("h", 40, "height of the popup in rows")
	output := flag.String("o", "", "write the rendered ANSI to this file instead of stdout")
	plain := flag.Bool("plain", false, "render without colour, the way a dumb terminal sees it")
	page := flag.Int("page", 0, "hint bar page to draw; a page past the end wraps")
	hover := flag.Int("hover", hoveredRow, "row index to draw as the mouse's hover, -1 for none")
	printZones := flag.Bool("zones", false, "print every clickable zone the render recorded")
	at := flag.String("at", "", "print the zone a click at x,y hits")
	flag.Parse()

	// The demo is also a capture tool: a pipe has no terminal, so force the
	// truecolor profile instead of letting lipgloss fall back to Ascii.
	lipgloss.SetColorProfile(termenv.TrueColor)
	if *plain {
		lipgloss.SetColorProfile(termenv.Ascii)
	}

	demo := renderAt(*width, *height, *page, *hover)
	if *printZones {
		writeZones(os.Stdout, demo.zones)
	}
	if *at != "" {
		x, y, err := parseCell(*at)
		if err != nil {
			fmt.Fprintf(os.Stderr, "demo: %v\n", err)
			os.Exit(2)
		}
		if id, ok := demo.zones.Hit(x, y); ok {
			fmt.Fprintf(os.Stdout, "%d,%d → %s%s\n", x, y, id, action(id))
		} else {
			fmt.Fprintf(os.Stdout, "%d,%d → nothing clickable\n", x, y)
		}
	}
	if *output == "" {
		fmt.Print(demo.pane)
		return
	}
	if err := os.WriteFile(*output, []byte(demo.pane), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "demo: write %s: %v\n", *output, err)
		os.Exit(1)
	}
}

// render draws the whole popup on its first page, hovering the row beside the
// selection.
func render(width, height int) string { return renderAt(width, height, 0, hoveredRow).pane }

// renderPage draws the popup showing one page of the hint bar, so a narrow window
// can be captured on every page it holds. The pane is clamped to the recommended
// size range, so a pane taller than the frame leaves the frame centred instead of
// hugging its top, and the frame itself is always as tall as the clamped pane: the
// footer is pinned to its bottom edge and the body keeps one blank row at each end.
func renderPage(paneWidth, paneHeight, page int) view {
	return renderAt(paneWidth, paneHeight, page, hoveredRow)
}

// renderAt draws one view and records the zones it drew, which is the whole
// mouse contract: the surface knows the geometry, so the product does not.
func renderAt(paneWidth, paneHeight, page, hover int) view {
	zones := zone.New()
	size := frame.Recommended
	surface := frame.Surface{
		Frame: frame.Frame{
			Title: theme.GoldBold.Render("tui-kit") + theme.Quiet.Render(" › ") + theme.AccentBold.Render("demo"),
			Width: size.ClampWidth(paneWidth),
		},
		Size: size,
	}
	f := surface.At(paneWidth)
	inner := f.Width - 2
	content := f.ContentWidth()

	header := text.Join(
		theme.Muted.Render("3 groups")+theme.Dot()+theme.AccentBold.Render("12 bindings"),
		theme.TealBold.Render("KEYS")+theme.Dot()+theme.Muted.Render("LIST")+theme.Dot()+theme.Muted.Render("12 / 12 shown"),
		content,
	)

	// The filter keeps its cells and the legend gives way whole, so a narrow window
	// never shows half a key.
	filter := text.JoinDropRight(
		theme.Accent.Render("❱ ")+theme.Muted.Render("Filter keys"+text.Ellipsis),
		chip.Legend(
			chip.Pair(theme.ChipText, "x", "after prefix"),
			chip.Pair(theme.ChipText, "ctrl+x", "direct"),
			theme.Muted.Render(text.Unset)+" "+theme.Muted.Render("unset"),
		),
		content,
	)

	count := frame.ColumnCount(inner, 2)
	widths := frame.ColumnWidths(inner, count)
	surface.Columns = widths

	detail := frame.Footer(
		theme.AccentBold.Render("Command bar")+theme.Dot()+
			theme.Green.Render("prefix+k")+theme.Dot()+
			theme.Muted.Render("herdr"),
		theme.Muted.Render("plugin action"),
		content,
	)

	// Six hints do not fit below 80 columns, so a narrow window shows the paging
	// chip and the rest of the bar behind PageKey. A fresh bar keeps the page it is
	// given, so -page asks for one and a screen whose hints change still resets
	// itself.
	bar := chip.HintBar{Page: page}
	// Content rows are framed; a rule is a whole line of its own and is placed as
	// it is, so it joins both borders instead of being framed inside them.
	surface.Top = []string{
		surface.Row(paneWidth, header),
		f.Rule(),
		surface.Row(paneWidth, filter),
		frame.ColumnRule(widths, frame.TopTee),
	}
	surface.Bottom = surface.Rows(paneWidth, "", detail, "")
	surface.Bottom[0] = frame.ColumnRule(widths, frame.BottomTee)
	// The row exists first, so the bar knows the content row its chips land on.
	surface.Bottom[2] = f.Row(bar.RenderZones(zones, 0, surface.BottomY(paneHeight, 2), content, demoHints...))

	blocks := columnBlocks(surface, paneHeight, hover, zones)
	body := frame.Grid(f, blocks, widths, surface.BodyRows(paneHeight))
	return view{pane: surface.PopupZones(paneWidth, paneHeight, body, zones), zones: zones}
}

// columnBlocks pours the groups into one block per column: the chip column is
// shared inside a column, the last entry of a group closes its guide, and one
// blank row separates two groups so the eye groups them without a rule. Every
// entry records the zones its row and its chip were drawn in.
func columnBlocks(surface frame.Surface, paneHeight, hover int, zones *zone.Map) [][]string {
	widths := surface.Columns
	blocks := make([][]string, len(widths))
	perColumn := (len(groups) + len(widths) - 1) / len(widths)
	position := 0 // flat index of the entry, so the demo can select and hover one
	// A body the pane cannot hold is cut, never squeezed: a row past the room is
	// not drawn, so it is not recorded and cannot be clicked.
	limit := surface.BodyRows(paneHeight)

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

		x := surface.ColumnX()[index]
		lines := []string{}
		for groupIndex, g := range groups[start:end] {
			if len(lines) >= limit {
				break
			}
			if groupIndex > 0 && len(lines)+1 < limit {
				lines = append(lines, strings.Repeat(" ", columnWidth))
			}
			if len(lines)+2 > limit {
				// No room for the header and one entry under it: a group header
				// with nothing below reads as a broken list, so neither is drawn.
				break
			}
			lines = append(lines, list.Section(g.name, len(g.entries)))
			for entryIndex, e := range g.entries {
				if len(lines) >= limit {
					break
				}
				state := list.State{}
				switch position {
				case selectedRow:
					state.Selected = true
				case hover:
					state.Hovered = true
				}
				roles := list.RowRolesFor(state)

				cell, description := chip.Cell(roles.Chip, e.key, chipWidth), roles.Name.Render(e.description)
				if e.unset {
					// An entry with no key keeps the column with a dim dash and no
					// slab, and its description recedes with it.
					cell, description = chip.UnsetCell(roles.Muted, e.key, chipWidth), roles.Muted.Render(e.description)
				}
				entry := list.Entry{
					Roles:       roles,
					Last:        entryIndex == len(g.entries)-1,
					Cell:        cell,
					Description: description,
					Width:       columnWidth,
					ID:          rowZone(position),
					Zones:       zones,
					X:           x,
					Y:           surface.BodyY(len(lines)),
				}
				if !e.unset {
					// A row with no key has no slab, so it has no chip to click.
					entry.ChipID = chipZone(position)
				}
				lines = append(lines, entry.Render())
				position++
			}
		}
		blocks[index] = lines
	}
	return blocks
}

// action is the wiring a product writes once per screen: the id a click produced,
// mapped to what it does. A click selects a row; enter or a double click runs it;
// a click on the paging chip pages the bar, exactly like chip.PageKey.
func action(id string) string {
	if index, ok := rowIndex(id); ok {
		return fmt.Sprintf(" (selects entry %d)", index)
	}
	if id == chip.PageZone {
		return " (next hint page, the same as " + chip.PageKey + ")"
	}
	return " (a chip; the demo wires no action of its own)"
}

// writeZones prints what the render recorded, one zone per line, so the mouse
// wiring can be read without a mouse.
func writeZones(w io.Writer, zones *zone.Map) {
	for _, z := range zones.Zones() {
		fmt.Fprintf(w, "%-14s %3d,%3d %3dx%d\n", z.ID, z.Rect.X, z.Rect.Y, z.Rect.W, z.Rect.H)
	}
}

// parseCell reads the x,y of a click from a flag.
func parseCell(value string) (x, y int, err error) {
	left, right, ok := strings.Cut(value, ",")
	if !ok {
		return 0, 0, fmt.Errorf("-%q wants x,y (for example 30,12)", "at")
	}
	if x, err = strconv.Atoi(strings.TrimSpace(left)); err != nil {
		return 0, 0, fmt.Errorf("-%s: %w", "at", err)
	}
	if y, err = strconv.Atoi(strings.TrimSpace(right)); err != nil {
		return 0, 0, fmt.Errorf("-%s: %w", "at", err)
	}
	return x, y, nil
}
