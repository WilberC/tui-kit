package zone_test

import (
	"testing"

	"github.com/WilberC/tui-kit/zone"
)

// TestToggleIDWrapsAStateName pins the contract a product relies on: the id a
// toggle control's cells are recorded under is the state name behind the
// toggle: prefix, so a click handler asks IsToggle and gets the state back.
func TestToggleIDWrapsAStateName(t *testing.T) {
	if got, want := zone.ToggleID("agent.gpu"), "toggle:agent.gpu"; got != want {
		t.Fatalf("ToggleID(%q) = %q, want %q", "agent.gpu", got, want)
	}

	name, ok := zone.IsToggle(zone.ToggleID("agent.gpu"))
	if !ok || name != "agent.gpu" {
		t.Fatalf("IsToggle(ToggleID(%q)) = %q/%v, want the state name back", "agent.gpu", name, ok)
	}

	zones := zone.New()
	zones.Add(zone.ToggleID("agent.gpu"), 2, 0, 6, 1)
	id, hit := zones.Hit(4, 0)
	if !hit || id != "toggle:agent.gpu" {
		t.Fatalf("Hit on a recorded toggle = %q/%v, want %q", id, hit, "toggle:agent.gpu")
	}
	if name, ok := zone.IsToggle(id); !ok || name != "agent.gpu" {
		t.Fatalf("IsToggle on a hit id = %q/%v, want the state name", name, ok)
	}
}

// TestIsToggleRejectsNonToggleIDs pins that only the exact toggle: prefix is a
// toggle: rows, chips and near misses keep their own meaning and answer false.
func TestIsToggleRejectsNonToggleIDs(t *testing.T) {
	for _, id := range []string{"row:0", "chip:page", "toggleX:gpu", "xxtoggle:gpu", "Toggle:gpu", "toggle"} {
		if name, ok := zone.IsToggle(id); ok || name != "" {
			t.Fatalf("IsToggle(%q) = %q/%v, want not a toggle", id, name, ok)
		}
	}
}

// TestAnEmptyIDIsNeverAToggle pins the degenerate case: an empty id names no
// state, so ToggleID wraps nothing and the bare prefix is not a toggle either.
func TestAnEmptyIDIsNeverAToggle(t *testing.T) {
	if got := zone.ToggleID(""); got != "" {
		t.Fatalf("ToggleID(%q) = %q, want no id at all", "", got)
	}
	if name, ok := zone.IsToggle(""); ok || name != "" {
		t.Fatalf("IsToggle(%q) = %q/%v, want not a toggle", "", name, ok)
	}
	if name, ok := zone.IsToggle("toggle:"); ok || name != "" {
		t.Fatalf("IsToggle of the bare prefix = %q/%v, want not a toggle", name, ok)
	}
}
