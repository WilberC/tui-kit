// Package query asks the terminal for its own colours: the default foreground
// and background and the sixteen ANSI slots, in one batch, the way Pi does.
//
// The query is the kit's only I/O: it opens /dev/tty, writes one batch of OSC
// requests (10, 11 and 4 × 16) and reads the replies until the DA1 reply that
// closes the batch arrives or the timeout expires. A palette is adopted
// all-or-nothing — a terminal that answers only part of the batch yields no
// palette, so the caller falls back to its default theme. Output that is not
// a terminal is never written to: Colors returns ErrNotTTY instead.
package query

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/x/term"
)

// Palette is what the terminal reported about itself: the default foreground
// and background and the sixteen ANSI slots, every value a lowercase #rrggbb
// or empty when that answer never arrived.
type Palette struct {
	FG    string
	BG    string
	Slots [16]string
}

// ErrNotTTY says the query has no terminal to ask: no /dev/tty, or one that is
// not a terminal. Nothing is written in that case.
var ErrNotTTY = errors.New("query: not a terminal")

// defaultTimeout is how long the run waits for the terminal's replies. Pi
// uses the same value, and applies replies that arrive later by re-querying;
// here the caller simply falls back, so the wait stays short.
const defaultTimeout = 100 * time.Millisecond

type config struct {
	timeout time.Duration
	isTerm  func(uintptr) bool
	open    func() (*os.File, error)
}

// An Option changes how Colors talks to the terminal.
type Option func(*config)

// WithTimeout bounds the wait for the terminal's replies.
func WithTimeout(d time.Duration) Option {
	return func(c *config) { c.timeout = d }
}

func withIsTerm(f func(uintptr) bool) Option {
	return func(c *config) { c.isTerm = f }
}

func withOpen(f func() (*os.File, error)) Option {
	return func(c *config) { c.open = f }
}

// Colors asks the terminal for its palette. It opens /dev/tty — never stdout —
// puts it in raw mode for the duration of the query, and restores it on every
// exit path.
func Colors(opts ...Option) (Palette, error) {
	c := config{
		timeout: defaultTimeout,
		isTerm:  term.IsTerminal,
		open:    func() (*os.File, error) { return os.OpenFile("/dev/tty", os.O_RDWR, 0) },
	}
	for _, opt := range opts {
		opt(&c)
	}

	f, err := c.open()
	if err != nil {
		return Palette{}, fmt.Errorf("%w: %v", ErrNotTTY, err)
	}
	defer f.Close()
	if !c.isTerm(f.Fd()) {
		return Palette{}, fmt.Errorf("%w: /dev/tty is not a terminal", ErrNotTTY)
	}

	state, err := term.MakeRaw(f.Fd())
	if err != nil {
		return Palette{}, fmt.Errorf("query: raw mode: %w", err)
	}
	defer term.Restore(f.Fd(), state)

	return run(f, f, c.timeout)
}
