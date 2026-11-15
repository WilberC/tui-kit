# tui-kit: terminal-inherited theming (v0.5.0)

Kit-wide theming: the Tokyo Night palette stays the built-in default; the kit
learns to inherit colors from the terminal (query OSC), to generate a system
theme from what the terminal reports, and to load custom themes from JSON —
following the model Pi implements. Existing callers compile unchanged; the
default look renders byte-identical.

## Specs

- S1: "Quiero que lo haga como habiamos visto que lo hace pi, mantener el tema
  que tiene ahora mismo como default, pero a parte tambien heredar e incluso
  poder agregar temas custom" — three sources of color: the built-in Tokyo
  Night palette (default), a theme generated from the terminal's own colors,
  and custom themes addable by the user.
- S2: fallback when "el terminal no se puede consultar (salida piperizada,
  terminal sin OSC 11)" is "Tokyo Night (paleta actual)" — the kit looks
  exactly as today, deterministically, with what is already captured.
- S3: gallery stays "Capturas Tokyo Night canónicas": existing captures remain
  byte-identical (sha256), the inherited mode is documented separately, the
  gallery stays deterministic.
- S4: the Pi model to mirror, verified in the installed pi source:
  - One query batch: OSC 10 (default fg) + OSC 11 (default bg) + OSC 4 for
    indices 0–15, terminated by a DA1 request (`\x1b[c`) as end-of-replies
    sentinel; 100 ms timeout; late replies may still be applied; the palette
    is adopted only when all 16 slots arrive.
  - A theme value accepts `#hex`, a 256 index, or `""`; `""` means "use the
    terminal's default" (no SGR emitted).
  - Resolution order: product override → custom theme file → built-in →
    system (generated); invalid/unloadable falls back to the default.
  - Escapes/theme are resolved once per theme build, not per frame; a theme
    change rebuilds derived state and marks the UI dirty.
- S5: enforcement carried over: no hex outside `theme/` is test-enforced;
  rendering is pinned by profile tests at call time; theme files are data,
  not code.
- S6: source compatibility: every existing exported symbol keeps its name and
  type; with no `Use` call the rendered bytes are identical to v0.4.x
  (gallery sha256 unchanged is the evidence).

## Tasks

- T1 (S1, S5, S6): `theme.Theme` struct (15 roles + name), `TokyoNight()`,
  `Use`/`Active` with validation, rebuild of the package style vars, and a
  hook registry so packages that captured roles at init (list) rebuild on
  `Use`. Inline.
- T2 (S4): `query` package: batch OSC 10/11/4 + DA1 sentinel, raw mode, 100 ms
  default timeout, `#rrggbb` normalization, all-or-nothing palette, `ErrNotTTY`
  when the output is not a terminal. Worker (background), package `query/`
  only.
- T3 (S1, S4): `theme.System(p query.Palette)`: derive the 15 roles from the
  reported fg/bg/slots (mixes toward the background, semantic colors from the
  ANSI slots, contrast gate for text); golden test feeds Tokyo Night's own
  terminal colors and must land within tolerance of the built-in palette.
  Inline, after T2.
- T4 (S1): custom themes: `theme.LoadFile(path)` (JSON, partial override over
  Tokyo Night, `""` = terminal default) and `theme.Auto(setting, dir)` pinning
  the decided resolution order with its fallback. Inline.
- T5 (S3, S5, S6): gallery pins `theme.Use(theme.TokyoNight())`, captures
  byte-identical; DESIGN.md documents the theme model (§1 note + new section)
  and README names the new surface with the migration table. Inline, last.

## Log

- L3 (2026-11-15): T1 committed 0d9df6b. Evidence: `go test ./...` green,
  `go vet ./...` clean, `gofmt -l .` empty. Test-first: RED as build failures
  (undefined `theme.Theme`/`Use`/`TokyoNight`/`Active`), two test-side bugs
  fixed (HoverBG is a background-only surface; the accent edge lives in
  SelectionEdge, not Name), then GREEN. Also learned: termenv renders
  #7aa2f7 as 38;2;121;162;247 (not 122) — pinned empirically.
- L1 (2026-11-15): "Continuemos con el analisis de hacer tui-kit en color
  heredado del terminal. Necesitas respuestas de mi parte? o puedes terminar
  todo ?" — analysis phase. Owner decisions recorded: default Tokyo Night +
  terminal inheritance + custom themes like pi (S1); fallback Tokyo Night
  (S2); canonical Tokyo Night captures (S3). Pi's theme system mapped with
  path:line evidence (system theme generated, not a file; query batch with
  DA1 sentinel; JSON themes in a per-name file; resolution order; precomputed
  escapes; no query in non-TTY).
- L2 (2026-11-15): Owner: "para eso tenemos el ejemplo de lo que hace Pi para
  agarrar buenas practicas de ahi y mejorar por lo que no tenemos que hacer
  todo desde 0" — authorization to implement the full scope mirroring pi.
  Feature doc created; branch `feat/terminal-theme`.
