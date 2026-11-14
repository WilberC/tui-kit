// Command gallery renders every documented state of tui-kit to a PNG under
// docs/gallery, so DESIGN.md can show the look it describes instead of only
// naming it.
//
// One command regenerates the whole gallery:
//
//	go run ./examples/gallery
//	go run ./examples/gallery -o /tmp/gallery     # somewhere else
//	go run ./examples/gallery -only rows          # one state, while working on it
//
// Every state is drawn at a fixed size with the truecolor profile and handed to
// `freeze`, so an image only changes when the look changes: no clock, no random
// content, no terminal-dependent width. The three demo captures are produced by
// running examples/demo itself, so the image is that command's own output rather
// than a copy of it that could drift.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/frame"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/spinner"
	"github.com/WilberC/tui-kit/text"
	"github.com/WilberC/tui-kit/theme"
)

// The capture is rendered the same way every time: the Tokyo Night theme behind
// the kit's own palette, one font size, no window chrome and no shadow, so a
// pixel only changes when the look does.
var freezeArgs = []string{
	"-l", "ansi",
	"--theme", "tokyonight",
	"--font.size", "10",
	"--line-height", "1",
	"--margin", "0",
	"--padding", "8",
	"--shadow.blur", "0",
	"--border.radius", "0",
}

// state is one documented screen: a name that is also its file name, the fixed
// size it is captured at, and how the pane is produced.
type state struct {
	Name   string
	Width  int
	Height int
	// Draw renders the pane. A state without Draw is a capture of
	// examples/demo, run the way the demo runs itself.
	Draw func(width, height int) string
}

// states is the gallery, in the order DESIGN.md reads it. Each width is the
// narrowest that still shows the element: a capture is evidence, not wallpaper.
var states = []state{
	{Name: "frame-anatomy", Width: 140, Height: 19, Draw: frameAnatomy},
	{Name: "columns", Width: 160, Height: 13, Draw: columns},
	{Name: "chips", Width: 80, Height: 19, Draw: chips},
	{Name: "rows", Width: 80, Height: 14, Draw: rows},
	{Name: "table", Width: 88, Height: 14, Draw: table},
	{Name: "hints-page-1", Width: 60, Height: 10, Draw: func(w, h int) string { return hintBar(w, h, 0) }},
	{Name: "hints-page-2", Width: 60, Height: 10, Draw: func(w, h int) string { return hintBar(w, h, 1) }},
	{Name: "empty", Width: 60, Height: 9, Draw: empty},
	{Name: "zero-match", Width: 60, Height: 12, Draw: zeroMatch},
	{Name: "loading", Width: 60, Height: 10, Draw: loading},
	{Name: "demo-180x50", Width: 180, Height: 50},
	{Name: "demo-140x40", Width: 140, Height: 40},
	{Name: "demo-80x24", Width: 80, Height: 24},
}

func main() {
	output := flag.String("o", filepath.Join("docs", "gallery"), "directory the PNGs are written to")
	only := flag.String("only", "", "regenerate only the state with this name")
	flag.Parse()

	if err := run(*output, *only); err != nil {
		fmt.Fprintf(os.Stderr, "gallery: %v\n", err)
		os.Exit(1)
	}
}

// run renders every state and captures it.
func run(output, only string) error {
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	freeze, err := findFreeze()
	if err != nil {
		return err
	}

	dir := output
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// A pipe has no terminal, so ask for truecolor instead of letting lipgloss
	// fall back to Ascii: the capture must carry the palette.
	lipgloss.SetColorProfile(termenv.TrueColor)

	// The ANSI intermediates live outside the worktree and are removed on the way
	// out, so a gallery run leaves only the PNGs.
	scratch, err := os.MkdirTemp("", "tui-kit-gallery-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)

	captured := 0
	for _, s := range states {
		if only != "" && only != s.Name {
			continue
		}
		source := filepath.Join(scratch, s.Name+".ansi")
		if err := write(s, root, source); err != nil {
			return err
		}
		target := filepath.Join(dir, s.Name+".png")
		if err := capture(freeze, source, target); err != nil {
			return err
		}
		fmt.Printf("%-14s %3dx%-3d %s\n", s.Name, s.Width, s.Height, target)
		captured++
	}
	if captured == 0 {
		return fmt.Errorf("no state named %q; the gallery has %s", only, names())
	}
	return nil
}

// write puts one state's ANSI in a file.
func write(s state, root, path string) error {
	if s.Draw != nil {
		return os.WriteFile(path, []byte(s.Draw(s.Width, s.Height)), 0o644)
	}
	demo := exec.Command("go", "run", "./examples/demo",
		"-w", fmt.Sprint(s.Width), "-h", fmt.Sprint(s.Height), "-o", path)
	demo.Dir = root
	demo.Stdout, demo.Stderr = os.Stdout, os.Stderr
	if err := demo.Run(); err != nil {
		return fmt.Errorf("run examples/demo for %s: %w", s.Name, err)
	}
	return nil
}

// capture renders one ANSI file to a PNG and keeps the PNG small enough to commit.
func capture(freeze, source, target string) error {
	cmd := exec.Command(freeze, append(append([]string{}, freezeArgs...), "-o", target, source)...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("freeze %s: %w", filepath.Base(target), err)
	}
	return shrink(target)
}

// shrink re-encodes a capture as a palette PNG. Freeze writes a truecolor PNG
// whose antialiasing leaves a few hundred distinct colours; storing them as a
// palette is the same picture — the background stays the exact theme colour and
// every glyph keeps its edge — at a seventh of the size, which is what makes the
// committed gallery reasonable.
func shrink(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	picture, err := png.Decode(file)
	file.Close()
	if err != nil {
		return fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}

	counts := map[color.RGBA]int{}
	bounds := picture.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			counts[snapRole(rgba(picture.At(x, y)))]++
		}
	}

	// The role colours are seeded first and kept exactly, so a captured slab or tint
	// is the role itself rather than its nearest neighbour; then the colours that
	// carry the picture, so a capture with more than a palette's worth of blends
	// keeps its background and its glyph edges.
	type counted struct {
		colour color.RGBA
		count  int
	}
	ranked := make([]counted, 0, len(counts)+len(theme.Palette()))
	seeded := map[color.RGBA]bool{}
	for _, role := range theme.Palette() {
		colour := rgba(role)
		if counts[colour] == 0 {
			continue // a role this capture never drew
		}
		seeded[colour] = true
		ranked = append(ranked, counted{colour: colour, count: counts[colour]})
	}
	for colour, count := range counts {
		if seeded[colour] {
			continue
		}
		ranked = append(ranked, counted{colour: colour, count: count})
	}
	// A total order, so two runs write the same bytes: the palette is ranked by
	// how much of the picture a colour carries, then by the colour itself.
	sort.Slice(ranked, func(i, j int) bool {
		left, right := ranked[i], ranked[j]
		if left.count != right.count {
			return left.count > right.count
		}
		if left.colour.R != right.colour.R {
			return left.colour.R < right.colour.R
		}
		if left.colour.G != right.colour.G {
			return left.colour.G < right.colour.G
		}
		if left.colour.B != right.colour.B {
			return left.colour.B < right.colour.B
		}
		return left.colour.A < right.colour.A
	})
	palette := make(color.Palette, min(256, len(ranked)))
	for index := range palette {
		palette[index] = ranked[index].colour
	}

	shrunk := image.NewPaletted(bounds, palette)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			shrunk.SetColorIndex(x, y, uint8(nearest(palette, snapRole(rgba(picture.At(x, y))))))
		}
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(out, shrunk); err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}
	return nil
}

// snapRole returns the role colour a rendered colour is a rounding step away from,
// and the colour itself when it is not. Freeze's rasterizer rounds some channels
// down by one, and a guide that documents a palette must show that palette.
func snapRole(colour color.RGBA) color.RGBA {
	for _, role := range theme.Palette() {
		if want := rgba(role); near(colour, want) {
			return want
		}
	}
	return colour
}

// near reports whether two colours differ by at most one step per channel.
func near(a, b color.RGBA) bool {
	for _, pair := range [][2]uint8{{a.R, b.R}, {a.G, b.G}, {a.B, b.B}, {a.A, b.A}} {
		low, high := min(pair[0], pair[1]), max(pair[0], pair[1])
		if high-low > 1 {
			return false
		}
	}
	return true
}

// rgba is a pixel as eight-bit channels, which is what a palette holds.
func rgba(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

// nearest finds the palette entry closest to a colour.
func nearest(palette color.Palette, want color.RGBA) int {
	best, bestDistance := 0, 1<<62
	for index, entry := range palette {
		have := rgba(entry)
		dr, dg, db, da := int(have.R)-int(want.R), int(have.G)-int(want.G), int(have.B)-int(want.B), int(have.A)-int(want.A)
		distance := dr*dr + dg*dg + db*db + da*da
		if distance < bestDistance {
			best, bestDistance = index, distance
		}
	}
	return best
}

// names lists the states the gallery knows, for the error a typo deserves.
func names() string {
	all := make([]string, len(states))
	for index, s := range states {
		all[index] = s.Name
	}
	return strings.Join(all, ", ")
}

// moduleRoot walks up from the working directory to the module root, so the
// command works from anywhere inside the repository.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no go.mod found: run the gallery from inside the tui-kit repository")
		}
		dir = parent
	}
}

// findFreeze locates the capture tool: the PATH first, then the owner's
// ~/.local/bin, so a trimmed PATH still regenerates the gallery.
func findFreeze() (string, error) {
	if path, err := exec.LookPath("freeze"); err == nil {
		return path, nil
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, ".local", "bin", "freeze")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", errors.New("freeze not found: it renders the ANSI captures to PNG " +
		"(go install github.com/charmbracelet/freeze@latest, or put it in ~/.local/bin)")
}

// screen starts a documented surface. A capture is the smallest frame that shows
// its state with the standard padding, so the pane IS the frame here: no clamping
// and no spare room below the content, because a guide image with half a pane of
// empty body would teach the opposite of DESIGN §8. A product still draws within
// frame.Recommended.
func screen(width int, product, name string) frame.Surface {
	return frame.Surface{
		Frame: frame.Frame{
			Title: theme.GoldBold.Render(product) + theme.Quiet.Render(" › ") + theme.AccentBold.Render(name),
			Width: width,
		},
	}
}

// chips is the token surface: a bare chip, a pair, a legend, a shared chip column,
// the dash an entry without a key keeps, and a chip on the selection tint.
func chips(width, height int) string {
	s := screen(width, "tui-kit", "chips")
	f := s.Frame
	content := f.ContentWidth()
	s.Top = []string{
		s.Row(width, text.Join(theme.Muted.Render("one slab per key"), theme.Muted.Render("7 kinds"), content)),
		f.Rule(),
	}

	chipWidth := chip.Width("prefix+k", "prefix+.", "enter")
	selected, panel := list.RowRoles(true), list.RowRoles(false)
	// One width for the three rows, so their descriptions start on one cell; the
	// row is the body's width, so the selection tint runs to the frame's padding.
	rowWidth := content
	body := s.Rows(width,
		chip.Chip(theme.ChipText, "enter")+" "+theme.Muted.Render("a chip is a padded slab, never [enter]"),
		chip.Pair(theme.ChipText, "ctrl+k", "a pair is a slab, one plain space, a label"),
		chip.Legend(
			chip.Pair(theme.ChipText, "x", "after prefix"),
			chip.Pair(theme.ChipText, "ctrl+x", "direct"),
		)+"   "+theme.Muted.Render("two plain spaces between items"),
		"",
		theme.Muted.Render("one chip column: every description starts on the same cell"),
		list.ChipRow(selected, false,
			chip.Cell(selected.Chip, "prefix+k", chipWidth),
			selected.Name.Render("the selected row's cell"),
			rowWidth),
		list.ChipRow(panel, false,
			chip.Cell(panel.Chip, "prefix+.", chipWidth),
			panel.Name.Render("an entry with its own key"),
			rowWidth),
		list.ChipRow(panel, true,
			chip.UnsetCell(panel.Muted, text.Unset, chipWidth),
			panel.Muted.Render("an entry with no key keeps the column"),
			rowWidth),
		"",
		theme.Muted.Render("a token on the tint uses ")+chip.Chip(theme.ChipOnSelection, "chip on selection"),
	)

	s.Bottom = s.Rows(width,
		"",
		frame.Footer(theme.AccentBold.Render("chips"), theme.Muted.Render("7 kinds"), content),
		"",
	)
	s.Bottom[0] = f.Rule()
	s.Bottom[2] = s.Row(width, chip.Hints(content,
		chip.Hint{Key: "esc", Label: "close"},
		chip.Hint{Key: "enter", Label: "copy"},
	))
	return s.Popup(width, height, body)
}

// rows is the listed surface with both marks: the selection in its tint with the
// accent edge, and the row the pointer is over in the fainter hover tint.
func rows(width, height int) string {
	s := screen(width, "herdr", "rows")
	f := s.Frame
	content := f.ContentWidth()
	s.Top = []string{
		s.Row(width, text.Join(
			theme.Muted.Render("selection")+theme.Dot()+theme.Muted.Render("hover")+theme.Dot()+theme.Muted.Render("panel"),
			theme.TealBold.Render("LIST")+theme.Dot()+theme.Muted.Render("1 / 5 selected"),
			content,
		)),
		f.Rule(),
	}

	// Two cells of room for the gold mark the default row carries before its cell.
	chipWidth := chip.Width("prefix+k", "prefix+p", "prefix+i")
	entries := []struct {
		key   string
		name  string
		state list.State
		mark  string
	}{
		{key: "prefix+k", name: "Command bar", state: list.State{Selected: true}},
		{key: "prefix+p", name: "Plugin manager", state: list.State{Hovered: true}},
		{key: "prefix+i", name: "Keybinding guide", mark: "default"},
		{key: "prefix+g", name: "Git status", mark: "live"},
		{key: "prefix+r", name: "Restart the runtime", mark: "failed"},
	}
	body := []string{}
	for index, e := range entries {
		roles := list.RowRolesFor(e.state)
		cell := chip.Cell(roles.Chip, e.key, chipWidth)
		description := roles.Name.Render(e.name)
		switch e.mark {
		case "default":
			// State that outlives the selection follows the row's facts, so the
			// mark never moves the chip column. It is not a second cursor.
			description += list.Mark(roles)
		case "live":
			description += "  " + roles.Green.Render(text.Live+" running")
		case "failed":
			description += "  " + roles.Red.Render(text.Idle+" failed")
		}
		body = append(body, s.Row(width, list.ChipRow(roles, index == len(entries)-1, cell, description, content)))
	}
	body = append(body, s.Row(width, ""))

	s.Bottom = s.Rows(width,
		"",
		frame.Footer(theme.AccentBold.Render("Command bar")+
			theme.Dot()+theme.Green.Render("prefix+k"), theme.Muted.Render("5 / 5 shown"), content),
		"",
	)
	s.Bottom[0] = f.Rule()
	s.Bottom[2] = s.Row(width, chip.Hints(content,
		chip.Hint{Key: "enter", Label: "run"},
		chip.Hint{Key: "↑↓", Label: "move"},
		chip.Hint{Key: "esc", Label: "close"},
	))
	return s.Popup(width, height, body)
}

// table is the LIST look: a row whose last column is right-aligned to the row's
// edge, with the selection and the pointer side by side so both tints can be judged
// through the gutter that separates the description from the trailing column. The
// chips carry the accent role, so the guide's bindings are the same token on the
// panel, under the pointer and on the selection.
func table(width, height int) string {
	s := screen(width, "hu", "which")
	f := s.Frame
	content := f.ContentWidth()
	s.Top = []string{
		s.Row(width, text.Join(
			theme.Muted.Render("trailing column"),
			theme.TealBold.Render("LIST")+theme.Dot()+theme.Muted.Render("4 rows"),
			content,
		)),
		f.Rule(),
	}

	chipWidth := chip.Width("prefix+k", "ctrl+g")
	// The source column is reserved, so every value and its header start on the same
	// cell; the spare cell after the longest value is the row's own background, so
	// the tint runs to the row's edge behind the column.
	sourceWidth := chip.Width("herdr-suite", "herdr-utils", "custom") + 1
	// The names ride the same cells their values do: list.Entry spends five cells on
	// the tree guide, and the chip column's padding places the label one cell later.
	header := text.Fit(strings.Repeat(" ", 5)+chip.Cell(theme.Muted, "BINDING", chipWidth)+theme.Muted.Render("ENTRY"), content-sourceWidth)
	header += theme.Muted.Render("SOURCE") + strings.Repeat(" ", sourceWidth-text.Width("SOURCE"))

	rows := []struct {
		key         string
		description string
		source      string
		unset       bool
		state       list.State
	}{
		{key: "prefix+k", description: "Command bar", source: "herdr-suite", state: list.State{Selected: true}},
		{key: "prefix+p", description: "Plugin manager", source: "herdr-suite", state: list.State{Hovered: true}},
		{key: "ctrl+g", description: "Lazygit", source: "custom"},
		{key: text.Unset, description: "Open the bar", source: "herdr-utils", unset: true},
	}

	body := []string{
		s.Row(width, header),
		s.Row(width, theme.Quiet.Render(strings.Repeat(frame.Horizontal, content))),
	}
	for index, row := range rows {
		roles := list.RowRolesFor(row.state)
		cell, description := chip.Cell(roles.ChipAccent, row.key, chipWidth), roles.Name.Render(row.description)
		if row.unset {
			cell, description = chip.UnsetCell(roles.Muted, row.key, chipWidth), roles.Muted.Render(row.description)
		}
		body = append(body, s.Row(width, list.Entry{
			Roles:         roles,
			Last:          index == len(rows)-1,
			Cell:          cell,
			Description:   description,
			Trailing:      roles.Muted.Render(row.source),
			TrailingWidth: sourceWidth,
			Width:         content,
		}.Render()))
	}

	s.Bottom = s.Rows(width,
		"",
		frame.Footer(
			theme.AccentBold.Render("Command bar")+theme.Dot()+theme.Green.Render("prefix+k"),
			theme.Muted.Render("4 / 4 shown"),
			content,
		),
	)
	s.Bottom[0] = f.Rule()
	return s.Popup(width, height, body)
}

// hintBar is the paging rule: a bar that cannot hold every hint ends with the
// paging chip and shows the rest on the next page, without taking a second row.
func hintBar(width, height, page int) string {
	s := screen(width, "tui-kit", "hint bar")
	f := s.Frame
	content := f.ContentWidth()
	s.Top = []string{
		s.Row(width, text.Join(theme.Muted.Render("6 hints"), theme.Muted.Render("60 cells"), content)),
		f.Rule(),
	}
	bar := chip.HintBar{Page: page}
	pages := bar.Pages(content, galleryHints...)
	body := s.Rows(width, text.Join(
		theme.Muted.Render("the body keeps its rows:"),
		theme.Muted.Render("page "+fmt.Sprint(page+1)+" of "+fmt.Sprint(pages)),
		content,
	))
	s.Bottom = s.Rows(width,
		"",
		frame.Footer(theme.AccentBold.Render("page "+fmt.Sprint(page+1)),
			theme.Muted.Render("tab moves on and wraps"), content),
		"",
	)
	s.Bottom[0] = f.Rule()
	s.Bottom[2] = s.Row(width, bar.Render(content, galleryHints...))
	return s.Popup(width, height, body)
}

// galleryHints is the hint set the paging captures page through.
var galleryHints = []chip.Hint{
	{Key: "enter", Label: "run"},
	{Key: "↑↓", Label: "move"},
	{Key: "type", Label: "filter"},
	{Key: "ctrl+g", Label: "git status"},
	{Key: "alt+enter", Label: "create tab"},
	{Key: "esc", Label: "close"},
}

// frameAnatomy is the whole surface: header, rule, filter with its legend, the
// column rule, two columns of groups, the closing rule and the two footer lines.
func frameAnatomy(width, height int) string {
	s := screen(width, "herdr", "sessions")
	f := s.Frame
	inner := f.Width - 2
	content := f.ContentWidth()
	widths := frame.ColumnWidths(inner, frame.ColumnCount(inner, 2))
	s.Columns = widths

	s.Top = append([]string{
		s.Row(width, text.Join(
			theme.Muted.Render("3 groups")+theme.Dot()+theme.AccentBold.Render("8 sessions"),
			theme.TealBold.Render("SESSIONS")+theme.Dot()+theme.Muted.Render("WORKSPACES"),
			content,
		)),
		f.Rule(),
		s.Row(width, text.Join(
			theme.Accent.Render("❱ ")+theme.Muted.Render("Filter sessions"+text.Ellipsis),
			chip.Legend(chip.Pair(theme.ChipText, "x", "after prefix")),
			content,
		)),
	}, frame.ColumnRule(widths, frame.TopTee))

	groups := []struct {
		name string
		rows []string
	}{
		{name: "herdr", rows: []string{"api", "web", "docs"}},
		{name: "work", rows: []string{"payments", "reports"}},
		{name: "personal", rows: []string{"tui-kit", "dotfiles", "notes"}},
	}

	perColumn := (len(groups) + len(widths) - 1) / len(widths)
	blocks := make([][]string, len(widths))
	position := 0
	for index, columnWidth := range widths {
		start := index * perColumn
		end := min(start+perColumn, len(groups))
		if start >= len(groups) {
			blocks[index] = nil
			continue
		}
		chipWidth := chip.Width("herdr", "work", "personal")
		// Grid frames every cell, so a block holds the cell's content, not a row.
		lines := []string{}
		for groupIndex, group := range groups[start:end] {
			if groupIndex > 0 {
				lines = append(lines, strings.Repeat(" ", columnWidth))
			}
			lines = append(lines, list.Section(group.name, len(group.rows)))
			for rowIndex, name := range group.rows {
				roles := list.RowRoles(position == 0)
				cell := chip.Cell(roles.Chip, group.name, chipWidth)
				lines = append(lines, list.ChipRow(roles, rowIndex == len(group.rows)-1,
					cell, roles.Name.Render(name), columnWidth))
				position++
			}
		}
		blocks[index] = lines
	}

	s.Bottom = s.Rows(width,
		"",
		frame.Footer(
			theme.AccentBold.Render("api")+theme.Dot()+theme.Green.Render("working"),
			theme.Muted.Render("3 / 8 · 1 running"),
			content,
		),
		"",
	)
	s.Bottom[0] = frame.ColumnRule(widths, frame.BottomTee)
	s.Bottom[2] = s.Row(width, chip.Hints(content,
		chip.Hint{Key: "enter", Label: "attach"},
		chip.Hint{Key: "esc", Label: "close"},
	))
	body := frame.Grid(f, blocks, widths, s.BodyRows(height))
	return s.Popup(width, height, body)
}

// columns is the grid: the column rules grow out of the frame, the vertical rules
// run through the padding rows, and each column keeps its own chip width.
func columns(width, height int) string {
	s := screen(width, "herdr", "grid")
	f := s.Frame
	inner := f.Width - 2
	widths := frame.ColumnWidths(inner, frame.ColumnCount(inner, 3))
	s.Columns = widths
	content := f.ContentWidth()
	s.Top = append([]string{
		s.Row(width, text.Join(theme.Muted.Render("three columns"),
			theme.TealBold.Render("GRID")+theme.Dot()+theme.Muted.Render("9 rows"),
			content)),
		f.Rule(),
	}, frame.ColumnRule(widths, frame.TopTee))

	titles := []string{"herdr", "work", "personal"}
	names := [][]string{
		{"api", "web", "docs"},
		{"payments", "reports", "invoices"},
		{"tui-kit", "dotfiles", "notes"},
	}
	blocks := make([][]string, len(widths))
	for index, columnWidth := range widths {
		chipWidth := chip.Width("prefix+k", "herdr")
		// Grid frames every cell, so a block holds the cell's content, not a row.
		lines := []string{list.Section(titles[index], len(names[index]))}
		for rowIndex, name := range names[index] {
			roles := list.RowRoles(index == 0 && rowIndex == 0)
			lines = append(lines, list.ChipRow(roles, rowIndex == len(names[index])-1,
				chip.Cell(roles.Chip, "prefix+k", chipWidth), roles.Name.Render(name), columnWidth))
		}
		blocks[index] = lines
	}

	s.Bottom = s.Rows(width,
		"",
		frame.Footer(theme.AccentBold.Render("api"), theme.Muted.Render("9 / 9 shown"), content),
	)
	s.Bottom[0] = frame.ColumnRule(widths, frame.BottomTee)
	body := frame.Grid(f, blocks, widths, s.BodyRows(height))
	return s.Popup(width, height, body)
}

// empty is the state before anything exists: what the surface is, and the one key
// that changes it.
func empty(width, height int) string {
	s := screen(width, "herdr", "sessions")
	f := s.Frame
	content := f.ContentWidth()
	placeholder := f.Placeholder("No sessions yet", "press prefix+enter to start one")
	s.Bottom = s.Rows(width,
		"",
		frame.Footer(theme.Muted.Render("nothing to list"), theme.Muted.Render("0 / 0"), content),
		"",
	)
	s.Bottom[0] = f.Rule()
	s.Bottom[2] = s.Row(width, chip.Hints(content,
		chip.Hint{Key: "prefix+enter", Label: "start a session"},
		chip.Hint{Key: "esc", Label: "close"},
	))
	// The surface pads the body itself, so the placeholder's own outer blank rows
	// are dropped and its two lines are framed like any other body row.
	return s.Popup(width, height, s.Rows(width, placeholder[1:len(placeholder)-1]...))
}

// zeroMatch is the filter that matches nothing: the query, and the way out.
func zeroMatch(width, height int) string {
	s := screen(width, "herdr", "sessions")
	f := s.Frame
	content := f.ContentWidth()
	s.Top = []string{
		s.Row(width, theme.Accent.Render("❱ ")+theme.Muted.Render("nothing-matches-this")+theme.Accent.Render(text.Cursor)),
		f.Rule(),
	}
	body := s.Rows(width,
		text.Center(theme.Strong.Render("No session matches “nothing-matches-this”"), content),
		"",
		text.Center(theme.Muted.Render("ctrl+u clears the filter · esc closes"), content),
	)
	s.Bottom = s.Rows(width,
		"",
		frame.Footer(theme.Muted.Render("0 matches"), theme.Muted.Render("0 / 8"), content),
		"",
	)
	s.Bottom[0] = f.Rule()
	s.Bottom[2] = s.Row(width, chip.Hints(content,
		chip.Hint{Key: "ctrl+u", Label: "clear"},
		chip.Hint{Key: "esc", Label: "close"},
	))
	return s.Popup(width, height, body)
}

// loading is work in flight: one teal tick on the line that describes it, and
// nothing else pretending to be content.
func loading(width, height int) string {
	s := screen(width, "hu", "which")
	f := s.Frame
	content := f.ContentWidth()
	body := s.Rows(width,
		text.Center(spinner.Render(1)+" "+theme.Muted.Render("looking for commands in your path"), content),
		"",
		text.Center(theme.Muted.Render("esc cancels"), content),
	)
	s.Bottom = s.Rows(width,
		"",
		frame.Footer(theme.Muted.Render("scanning"), theme.Muted.Render("…"), content),
		"",
	)
	s.Bottom[0] = f.Rule()
	s.Bottom[2] = s.Row(width, chip.Hints(content, chip.Hint{Key: "esc", Label: "cancel"}))
	return s.Popup(width, height, body)
}
