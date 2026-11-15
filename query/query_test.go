package query

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The batch mirrors pi's: default fg, default bg, the sixteen slots, and the
// DA1 request as the end-of-replies sentinel.
func TestTheQueryAsksForEverythingInOneBatch(t *testing.T) {
	var want strings.Builder
	want.WriteString("\x1b]10;?\x07")
	want.WriteString("\x1b]11;?\x07")
	for i := 0; i < 16; i++ {
		want.WriteString("\x1b]4;" + strconv.Itoa(i) + ";?\x07")
	}
	want.WriteString("\x1b[c")
	if got := queryBytes; got != want.String() {
		t.Fatalf("the query batch is %q, want %q", got, want.String())
	}
}

// A terminal that answers everything, with BEL terminators and four-digit
// components, produces a complete palette.
func TestFullRepliesBuildACompletePalette(t *testing.T) {
	var in strings.Builder
	in.WriteString("\x1b]10;rgb:c8ca/f5f5/c0c0\x07")
	in.WriteString("\x1b]11;rgb:1a1a/1b1b/2626\x07")
	for i := 0; i < 16; i++ {
		in.WriteString("\x1b]4;" + strconv.Itoa(i) + ";rgb:aaaa/bbbb/cccc\x07")
	}
	in.WriteString("\x1b[?62;22c")

	p, err := run(strings.NewReader(in.String()), &strings.Builder{}, time.Second)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if p.FG != "#c8f5c0" || p.BG != "#1a1b26" {
		t.Fatalf("the defaults came back as %q/%q, want #c8f5c0/#1a1b26", p.FG, p.BG)
	}
	for i, slot := range p.Slots {
		if slot != "#aabbcc" {
			t.Fatalf("slot %d is %q, want #aabbcc", i, slot)
		}
	}
}

func TestSTTerminatorsParseLikeBEL(t *testing.T) {
	var in strings.Builder
	in.WriteString("\x1b]10;rgb:c8ca/f5f5/c0c0\x1b\\")
	in.WriteString("\x1b]11;rgb:1a1a/1b1b/2626\x1b\\")
	for i := 0; i < 16; i++ {
		in.WriteString("\x1b]4;" + strconv.Itoa(i) + ";rgb:aaaa/bbbb/cccc\x1b\\")
	}
	in.WriteString("\x1b[?1;2c")

	if _, err := run(strings.NewReader(in.String()), &strings.Builder{}, time.Second); err != nil {
		t.Fatalf("run: %v", err)
	}
}

// The slots may arrive in any order — the sentinel is what ends the read, not
// the order of the answers before it.
func TestRepliesInAnyOrderStillBuildThePalette(t *testing.T) {
	var in strings.Builder
	for i := 15; i >= 0; i-- {
		in.WriteString("\x1b]4;" + strconv.Itoa(i) + ";rgb:aaaa/bbbb/cccc\x07")
	}
	in.WriteString("\x1b]11;rgb:1a1a/1b1b/2626\x07")
	in.WriteString("\x1b]10;rgb:c8ca/f5f5/c0c0\x07")
	in.WriteString("\x1b[?62;22c")

	if _, err := run(strings.NewReader(in.String()), &strings.Builder{}, time.Second); err != nil {
		t.Fatalf("run: %v", err)
	}
}

// The sentinel ends the read wherever it appears, and nothing after it is
// applied: a batch closed before the colours is an incomplete palette.
func TestASentinelThatArrivesFirstClosesAnEmptyBatch(t *testing.T) {
	p, err := run(strings.NewReader("\x1b[?62;22c"), &strings.Builder{}, time.Second)
	if err == nil {
		t.Fatal("the run adopted a palette closed before any colour")
	}
	if p.FG != "" || p.BG != "" {
		t.Fatalf("a batch closed by its sentinel still returned colours: %+v", p)
	}
}

// Component widths the spec allows: 1, 2, 3 and 4 hex digits, rescaled the way
// XTerm does (value * 255 / max-of-width).
func TestComponentWidthsNormalizeToEightBits(t *testing.T) {
	cases := map[string]string{
		"rgb:c8ca/f5f5/c0c0": "#c8f5c0",
		"rgb:c8/ca/f5":       "#c8caf5",
		"rgb:c/a/f":          "#ccaaff",
		"rgb:c8c8/f5f5/c0c0": "#c8f5c0",
	}
	for in, want := range cases {
		if got := normalizeRGB(in); got != want {
			t.Fatalf("normalizeRGB(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeRejectsWhatIsNotAnRGBValue(t *testing.T) {
	for _, in := range []string{"", "rgb:", "rgb:zz/aa/bb", "#c8caf5", "rgb:c8/ca"} {
		if got := normalizeRGB(in); got != "" {
			t.Fatalf("normalizeRGB(%q) = %q, want empty", in, got)
		}
	}
}

func TestASilentTerminalTimesOut(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	_, err = run(r, &strings.Builder{}, 30*time.Millisecond)
	if err == nil {
		t.Fatal("a silent terminal produced a palette")
	}
}

func TestAnIncompletePaletteIsRejected(t *testing.T) {
	var in strings.Builder
	in.WriteString("\x1b]10;rgb:c8ca/f5f5/c0c0\x07")
	in.WriteString("\x1b]11;rgb:1a1a/1b1b/2626\x07")
	for i := 0; i < 15; i++ {
		in.WriteString("\x1b]4;" + strconv.Itoa(i) + ";rgb:aaaa/bbbb/cccc\x07")
	}
	in.WriteString("\x1b[?62;22c")

	p, err := run(strings.NewReader(in.String()), &strings.Builder{}, time.Second)
	if err == nil {
		t.Fatal("a palette with fifteen slots was adopted")
	}
	if p.FG != "" || p.BG != "" {
		t.Fatalf("a rejected run still returned colours: %+v", p)
	}
}

func TestAnAnsweredPaletteStopsTheRead(t *testing.T) {
	var in strings.Builder
	in.WriteString("\x1b]10;rgb:c8ca/f5f5/c0c0\x07")
	in.WriteString("\x1b]11;rgb:1a1a/1b1b/2626\x07")
	for i := 0; i < 16; i++ {
		in.WriteString("\x1b]4;" + strconv.Itoa(i) + ";rgb:aaaa/bbbb/cccc\x07")
	}
	in.WriteString("\x1b[?62;22c")
	// After the sentinel the reader would hand out junk forever; the run must
	// never ask it again.
	reader := strings.NewReader(in.String() + "\x1b]10;rgb:zzzz\x07")

	p, err := run(reader, &strings.Builder{}, time.Second)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if p.FG != "#c8f5c0" {
		t.Fatalf("the late junk overwrote the foreground: %q", p.FG)
	}
}

func TestColorsWithoutATerminalIsErrNotTTY(t *testing.T) {
	_, err := Colors(withOpen(func() (*os.File, error) { return nil, os.ErrPermission }))
	if !errors.Is(err, ErrNotTTY) {
		t.Fatalf("Colors without a tty = %v, want ErrNotTTY", err)
	}
}

func TestColorsOnANonTerminalFileIsErrNotTTY(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	_, err = Colors(withOpen(func() (*os.File, error) { return w, nil }), withIsTerm(func(uintptr) bool { return false }))
	if !errors.Is(err, ErrNotTTY) {
		t.Fatalf("Colors on a pipe = %v, want ErrNotTTY", err)
	}
}
