package clipboard_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/clipboard"
)

// TestOSC52IsTheSequenceThatSetsTheClipboard pins what a product's copy key
// writes: OSC 52, the clipboard selection, the text in base64 and the string
// terminator — one sequence, no newline, nothing to strip.
func TestOSC52IsTheSequenceThatSetsTheClipboard(t *testing.T) {
	for _, value := range []string{"prefix+k", "Command bar", "├─ herdr · 3 tabs", "mañana ✓", "a"} {
		got := clipboard.OSC52(value)
		want := "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(value)) + "\x1b\\"

		if got != want {
			t.Fatalf("OSC52(%q) = %q, want %q", value, got, want)
		}
		if strings.ContainsAny(got, "\n\r") {
			t.Fatalf("OSC52(%q) carries a line break: %q", value, got)
		}
		payload := strings.TrimSuffix(strings.TrimPrefix(got, "\x1b]52;c;"), "\x1b\\")
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			t.Fatalf("OSC52(%q) does not carry base64: %v", value, err)
		}
		if string(decoded) != value {
			t.Fatalf("OSC52(%q) round-trips to %q", value, decoded)
		}
	}
	if clipboard.OSC52("x") != clipboard.OSC52("x") {
		t.Fatalf("OSC52 is not deterministic")
	}
}

// TestOSC52OfNothingIsNothing pins the empty case: a product with no value must
// not clear the user's clipboard with an accidental empty write.
func TestOSC52OfNothingIsNothing(t *testing.T) {
	if got := clipboard.OSC52(""); got != "" {
		t.Fatalf("OSC52 with no text = %q, want the empty string", got)
	}
}
