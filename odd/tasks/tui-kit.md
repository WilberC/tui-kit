# tui-kit

Public Go module (github.com/WilberC/tui-kit) holding the shared visual system of
herdr-suite and herdr-utils, extracted from their tested code, plus the design guide.

## Specs

- S1: "Podemos armar alguna libreria o algo similar con estilos o parametros que nos sirva para armar nuestras TUI? seria una libreria privada para seguir la linea visual y buenas practicas de diseño en TUI" — the owner then chose: a Go module in its own repo, public (so public release builds and `go install` fetch it without credentials), built after the redesign was committed by extracting the already-tested code.
- S2: Packages: `theme/` (Tokyo Night roles), `frame/` (frame, rules, footer), `chip/` (chips and hint bar), `list/` (rows, tree, selection), `text/` (fit, clip, ellipsis), and `DESIGN.md` (the visual guide).
- S3: The look stays exactly what the owner approved on 2026-11-13: tinted selection (bg #283457, accent ▌ edge, every element keeps its role), padded chip slabs, one blank row between groups, joined rules, 80 to 180 columns without wrapping.
- S4: herdr-suite and herdr-utils import the module with `go get` and drop their private copies; their screens render the same as before (byte-identical where the code moved unchanged).
- S5: Repository name: "tui-kit esta bien".

## Tasks

| ID | Specs | Route | Commit |
| --- | --- | --- | --- |
| T1 | S1-S3, S5 | worker | v0.1.0 |
| T2 | S4 (herdr-suite) | worker | pending |
| T3 | S4 (herdr-utils) | worker | pending |

## Log

- L1 (2026-11-13): "Muchisimo mejor!! Aunque hay unos pequeños detalles. [...] Podemos armar alguna libreria o algo similar con estilos o parametros que nos sirva para armar nuestras TUI? seria una libreria privada para seguir la linea visual y buenas practicas de diseño en TUI"
- L2 (2026-11-13): owner decisions: Go module in its own repo; public; after closing the selection fix.
- L3 (2026-11-14): "tui-kit esta bien" — repo created: github.com/WilberC/tui-kit (public, MIT).
- L4 (2026-11-14): T1 built. Module `github.com/WilberC/tui-kit` (go 1.27.3, lipgloss v1.1.0, x/ansi v0.10.1, termenv v0.16.0) with packages `theme/`, `text/`, `chip/`, `list/`, `frame/`, `spinner/`, extracted from herdr-suite `internal/ui` and herdr-utils `internal/which`; `DESIGN.md` is the canonical visual guide, `README.md` the entry point, `examples/demo` the visual check. Evidence: `go test ./...` 52 tests pass, `go vet ./...` clean, `gofmt -l .` empty; captures at 180/140/80 in /tmp/tui-kit-shots (demo-180/140/80.png) reviewed by eye; palette pinned in theme (no hex outside `theme/`, test-enforced); rendering at call time pinned by profile tests. Commit pending, parent owns it.
