# tui-kit

The shared visual system of my terminal UIs, extracted from the code that already
ships in [herdr-suite](https://github.com/WilberC/herdr-suite) and
[herdr-utils](https://github.com/WilberC/herdr-utils): one Tokyo Night palette,
one frame, one chip, one selection, one footer.

The point is that two binaries read as one product instead of two screens that
happen to share a terminal. Render code names a **role**, never a colour, and the
frame clips every line to its own width, so a view can never wrap — at 180
columns or at 80.

`DESIGN.md` is the canonical visual guide both products follow; this package is
its executable form.

## Install

```sh
go get github.com/WilberC/tui-kit@latest
```

Requires Go 1.27.3 or newer.

## Usage

```go
package main

import (
	"fmt"

	"github.com/WilberC/tui-kit/chip"
	"github.com/WilberC/tui-kit/frame"
	"github.com/WilberC/tui-kit/list"
	"github.com/WilberC/tui-kit/theme"
)

func main() {
	f := frame.Frame{
		Title: theme.GoldBold.Render("my-tui") + theme.Quiet.Render(" › ") + theme.AccentBold.Render("list"),
		Width: 60,
	}
	roles := list.RowRoles(true) // the highlighted row
	cell := chip.Cell(roles.Chip, "ctrl+k", 8)
	row := list.ChipRow(roles, true, cell, roles.Name.Render("Open the command bar"), f.ContentWidth())

	fmt.Print(frame.Layout(f,
		f.Row(theme.Muted.Render("3 bindings")),
		f.Rule(),
		f.Row(list.Section("keys", 3)),
		f.Row(row),
		f.Row(frame.Footer(theme.AccentBold.Render("Command bar"), "3 / 3 shown", f.ContentWidth())),
		f.Row(chip.Hints(f.ContentWidth(), chip.Hint{Key: "esc", Label: "close"})),
	))
}
```

`examples/demo` renders a full popup with groups, chips, a tinted row and a
footer at any size, so the look can be checked by eye:

```sh
go run ./examples/demo -w 140 -h 40
go run ./examples/demo -w 180 -o /tmp/demo.ansi   # piped, keeps truecolor
```

## Packages

| Package | Holds | Depends on |
| --- | --- | --- |
| `theme/` | The Tokyo Night palette, the role styles, and the few compositions that are pure role arithmetic (`Dot`, `Label`, `OnSelection`). The only place a hex value may live. | lipgloss |
| `text/` | Cell widths, clipping and ellipsis, `Fit`/`PadLeft`/`Join`/`Center`, short paths, relative ages, the typographic glyphs. | x/ansi |
| `frame/` | `Frame` (titled border, rules, rows), `Layout`, `Float`, `Footer`, and the joined column grid (`ColumnRule`, `ColumnWidths`, `Grid`). | theme, text |
| `chip/` | `Chip`, `Pair`, `Cell`/`UnsetCell` chip columns, `Legend`, and the `Hints` bar that drops whole hints instead of wrapping. | theme, text |
| `list/` | `Roles`/`RowRoles`, the tree `Guide`, `Row`, `SelectedRow`, `ChipRow`, `Section`, agent states. | theme, text |
| `spinner/` | The four-frame work-in-progress tick; the caller owns the tick. | theme |

The dependency graph runs one way: `theme` and `text` at the bottom, then `chip`,
`list` and `spinner`, then `frame`. Nothing imports upward, and `text` knows
nothing about colour.

Every role and glyph is rendered at call time — never as a package-level
pre-rendered string — so the colour profile in force when a line is drawn decides
its colour.

## Development

```sh
go test ./...
go vet ./...
gofmt -l .
```

Screenshots for the visual check:

```sh
mkdir -p /tmp/tui-kit-shots
go run ./examples/demo -w 140 -h 40 -o /tmp/tui-kit-shots/demo-140.ansi
freeze -l ansi --theme tokyonight -o /tmp/tui-kit-shots/demo-140.png /tmp/tui-kit-shots/demo-140.ansi
```

## License

MIT — see [LICENSE](LICENSE).
