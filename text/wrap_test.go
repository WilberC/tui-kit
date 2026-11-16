package text

import (
	"strings"
	"testing"
)

func TestWrapKeepsShortLinesAndFits(t *testing.T) {
	got := Wrap("cpu 32%", 40)
	if len(got) != 1 || got[0] != Fit("cpu 32%", 40) {
		t.Fatalf("a fitting line returns one fitted row, got %q", got)
	}
}

func TestWrapFoldsOnWordsAndPreservesEveryWord(t *testing.T) {
	got := Wrap("host cpu 32% mem 61% load 0.15", 15)
	if testing.Verbose() {
		t.Logf("folded: %q", got)
	}
	joined := strings.Join(got, " ")
	for _, word := range []string{"host", "cpu", "32%", "mem", "61%", "load", "0.15"} {
		if !strings.Contains(joined, word) {
			t.Fatalf("wrap must keep every word, got %q", joined)
		}
	}
	for _, line := range got {
		if w := Width(strings.TrimRight(line, " ")); w > 15 {
			t.Fatalf("line wider than width: %q (%d)", line, w)
		}
	}
	// every line is fitted to the same width
	for _, line := range got {
		if Width(line) != 15 {
			t.Fatalf("lines must be fitted to width, got %q (%d)", line, Width(line))
		}
	}
}

func TestWrapPreservesANSIStyles(t *testing.T) {
	styled := "\x1b[38;2;247;118;142mhost 32%\x1b[0m long tail here"
	got := Wrap(styled, 12)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, "38;2;247;118;142") {
		t.Fatalf("the style must survive the fold, got %q", joined)
	}
	for _, word := range []string{"long", "tail", "here"} {
		if !strings.Contains(joined, word) {
			t.Fatalf("the content must survive the fold, missing %q: %q", word, joined)
		}
	}
}

func TestWrapRowsFoldsEveryRow(t *testing.T) {
	got := WrapRows([]string{"short", "a much longer row that will fold", "ok"}, 10)
	if len(got) < 4 {
		t.Fatalf("the long row folds, got %d rows: %q", len(got), got)
	}
	if got[0] != "short     " {
		t.Fatalf("short rows stay first and fitted, got %q", got[0])
	}
}

func TestWrapWidthZeroReturnsTheLine(t *testing.T) {
	got := Wrap("anything", 0)
	if len(got) != 1 || got[0] != "anything" {
		t.Fatalf("width 0 folds nothing, got %q", got)
	}
}
