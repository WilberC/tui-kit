package gauge

import (
	"strings"
	"testing"
)

func TestHSplitsFilledAndLight(t *testing.T) {
	filled, empty := H(32.5, 100, 8)
	if filled != "███" || empty != "░░░░░" {
		t.Fatalf("32.5%% of 100 over 8 cells = 3 filled, got %q %q", filled, empty)
	}
}

func TestHClampsBeyondTheScale(t *testing.T) {
	filled, empty := H(150, 100, 8)
	if filled != strings.Repeat(Filled, 8) || empty != "" {
		t.Fatalf("beyond the scale is a full bar, got %q %q", filled, empty)
	}
	filled, empty = H(-5, 100, 8)
	if filled != "" || empty != strings.Repeat(Light, 8) {
		t.Fatalf("negative is an empty bar, got %q %q", filled, empty)
	}
}

func TestHScaleZeroIsEmpty(t *testing.T) {
	filled, empty := H(50, 0, 4)
	if filled != "" || empty != strings.Repeat(Light, 4) {
		t.Fatalf("no scale, no bar: %q %q", filled, empty)
	}
}

func TestVTopDownFilledFromTheBottom(t *testing.T) {
	rows := V(25, 100, 4) // 25% of 4 cells = 1 filled
	if len(rows) != 4 {
		t.Fatalf("one row per cell, got %d", len(rows))
	}
	if rows[0] != Light || rows[1] != Light || rows[2] != Light || rows[3] != Filled {
		t.Fatalf("25%% fills the bottom row only: %v", rows)
	}
	rows = V(75, 100, 4) // 3 filled
	if rows[0] != Light || rows[1] != Filled || rows[2] != Filled || rows[3] != Filled {
		t.Fatalf("75%% fills the bottom three: %v", rows)
	}
}

func TestVFillsCompletelyBeyondTheScale(t *testing.T) {
	rows := V(200, 100, 4)
	for i, r := range rows {
		if r != Filled {
			t.Fatalf("row %d must be filled: %v", i, rows)
		}
	}
}

func TestPartsComposeIntoTheOldBar(t *testing.T) {
	filled, empty := H(61.4, 100, 8)
	if filled+empty != "█████░░░" {
		t.Fatalf("the parts join into the classic bar, got %q", filled+empty)
	}
}
