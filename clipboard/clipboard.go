// Package clipboard writes a value to the user's clipboard through the terminal,
// without reading the mouse or the screen.
//
// A TUI that captures the mouse cannot also drag-select text: the terminal hands
// the drag to the program. The kit's answer (DESIGN.md §12) is not to
// reimplement selection — copying is central only to a chat transcript — but to
// offer one key that copies the one value the screen is about, written with OSC
// 52. This package is that sequence, and nothing else: it returns the bytes, so
// the caller decides where to write them.
package clipboard

import "encoding/base64"

// OSC52 wraps text in the sequence that sets the system clipboard (OSC 52,
// clipboard selection "c", the text in base64, closed by the string terminator).
// Writing it to the terminal makes the terminal copy the text where the terminal
// supports it; where it does not, the sequence is ignored and nothing breaks.
//
// The text travels through the terminal, so keep it to one short value: several
// terminals cap the payload (a few kilobytes at most) and drop the whole write
// past it. An empty text returns the empty string, so a screen with nothing to
// copy cannot clear the user's clipboard by accident.
func OSC52(text string) string {
	if text == "" {
		return ""
	}
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(text)) + "\x1b\\"
}
