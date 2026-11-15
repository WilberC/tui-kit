package query

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// queryBytes is the batch the run writes: the default foreground, the default
// background, the sixteen slots, and the DA1 request whose reply closes the
// read.
const queryBytes = "" +
	"\x1b]10;?\x07" +
	"\x1b]11;?\x07" +
	"\x1b]4;0;?\x07\x1b]4;1;?\x07\x1b]4;2;?\x07\x1b]4;3;?\x07" +
	"\x1b]4;4;?\x07\x1b]4;5;?\x07\x1b]4;6;?\x07\x1b]4;7;?\x07" +
	"\x1b]4;8;?\x07\x1b]4;9;?\x07\x1b]4;10;?\x07\x1b]4;11;?\x07" +
	"\x1b]4;12;?\x07\x1b]4;13;?\x07\x1b]4;14;?\x07\x1b]4;15;?\x07" +
	"\x1b[c"

// run writes the batch and gathers the replies until the DA1 sentinel arrives
// or the timeout expires, then judges the palette all-or-nothing.
func run(r io.Reader, w io.Writer, timeout time.Duration) (Palette, error) {
	if _, err := w.Write([]byte(queryBytes)); err != nil {
		return Palette{}, fmt.Errorf("query: write the batch: %w", err)
	}

	deadline := time.Now().Add(timeout)
	var buf []byte
	tmp := make([]byte, 512)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		if d, ok := r.(interface{ SetReadDeadline(time.Time) error }); ok {
			d.SetReadDeadline(deadline)
		}
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			if _, saw := parseReplies(buf); saw {
				break
			}
		}
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, io.EOF) {
				break
			}
			return Palette{}, fmt.Errorf("query: read the replies: %w", err)
		}
	}

	p, _ := parseReplies(buf)
	if !p.complete() {
		return Palette{}, fmt.Errorf("query: the terminal reported %s", p.describeMissing())
	}
	return p, nil
}

func (p Palette) complete() bool {
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

func (p Palette) describeMissing() string {
	missing := []string{}
	if p.FG == "" {
		missing = append(missing, "the default foreground")
	}
	if p.BG == "" {
		missing = append(missing, "the default background")
	}
	absent := []string{}
	for i, slot := range p.Slots {
		if slot == "" {
			absent = append(absent, strconv.Itoa(i))
		}
	}
	if len(absent) > 0 {
		missing = append(missing, "slot(s) "+strings.Join(absent, ", "))
	}
	if len(missing) == 0 {
		return "a complete palette"
	}
	return "no answer for " + strings.Join(missing, " and ")
}

// parseReplies walks one buffer of terminal output, applying every complete
// OSC reply it finds and reporting whether the DA1 reply has been seen. An
// unterminated reply at the end of the buffer is left for the next read.
func parseReplies(buf []byte) (Palette, bool) {
	var p Palette
	sawDA1 := false
	i := 0
	for i < len(buf) {
		if buf[i] != 0x1b || i+1 >= len(buf) {
			i++
			continue
		}
		switch buf[i+1] {
		case ']':
			end, termLen := findTerminator(buf, i+2)
			if end < 0 {
				return p, sawDA1
			}
			applyOSC(&p, string(buf[i+2:end]))
			i = end + termLen
		case '[':
			end := findFinal(buf, i+2)
			if end < 0 {
				return p, sawDA1
			}
			if isDA1(string(buf[i : end+1])) {
				// The sentinel closes the batch: nothing after it is applied.
				return p, true
			}
			i = end + 1
		default:
			i++
		}
	}
	return p, sawDA1
}

// findTerminator locates the BEL or ST that closes an OSC sequence.
func findTerminator(buf []byte, from int) (int, int) {
	for j := from; j < len(buf); j++ {
		if buf[j] == 0x07 {
			return j, 1
		}
		if buf[j] == 0x1b && j+1 < len(buf) && buf[j+1] == '\\' {
			return j, 2
		}
	}
	return -1, 0
}

// findFinal locates the final byte (0x40–0x7e) that closes a CSI sequence.
func findFinal(buf []byte, from int) int {
	for j := from; j < len(buf); j++ {
		if buf[j] >= 0x40 && buf[j] <= 0x7e {
			return j
		}
	}
	return -1
}

// isDA1 reports whether a CSI sequence is the device-attributes reply that
// closes the batch: ESC [ ? ... c.
func isDA1(seq string) bool {
	return strings.HasPrefix(seq, "\x1b[?") && strings.HasSuffix(seq, "c")
}

// applyOSC applies one OSC body — the part between CSI and its terminator —
// to the palette. Slot replies may arrive in any order.
func applyOSC(p *Palette, body string) {
	parts := strings.Split(body, ";")
	switch {
	case parts[0] == "10" && len(parts) >= 2:
		if v := normalizeRGB(parts[1]); v != "" {
			p.FG = v
		}
	case parts[0] == "11" && len(parts) >= 2:
		if v := normalizeRGB(parts[1]); v != "" {
			p.BG = v
		}
	case parts[0] == "4" && len(parts) >= 3:
		i, err := strconv.Atoi(parts[1])
		if err == nil && i >= 0 && i < 16 {
			if v := normalizeRGB(parts[2]); v != "" {
				p.Slots[i] = v
			}
		}
	}
}

// normalizeRGB turns an XTerm rgb: answer into a lowercase #rrggbb. Each
// component may carry one to four hex digits, and is rescaled to eight bits
// against the maximum of its own width (value * 255 / max), the way XTerm
// does. Anything else normalizes to empty.
func normalizeRGB(s string) string {
	rest, ok := strings.CutPrefix(s, "rgb:")
	if !ok {
		if rest, ok = strings.CutPrefix(s, "rgba:"); !ok {
			return ""
		}
	}
	groups := strings.Split(rest, "/")
	if len(groups) < 3 {
		return ""
	}
	var out strings.Builder
	out.WriteByte('#')
	for _, g := range groups[:3] {
		if g == "" || len(g) > 4 {
			return ""
		}
		v, err := strconv.ParseUint(g, 16, 64)
		if err != nil {
			return ""
		}
		max := uint64(1)<<(4*len(g)) - 1
		scaled := v * 255 / max
		fmt.Fprintf(&out, "%02x", scaled)
	}
	return out.String()
}
