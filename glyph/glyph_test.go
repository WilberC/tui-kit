package glyph_test

import (
	"testing"

	"github.com/WilberC/tui-kit/glyph"
	"github.com/WilberC/tui-kit/text"
)

func TestRoleMapsAgentIDsToTheirFamilyGlyph(t *testing.T) {
	cases := []struct {
		agentID string
		want    string
	}{
		// explore: suffix or contained.
		{"gentle-ai-explore", glyph.Explore},
		{"pi-explore", glyph.Explore},
		{"explorer", glyph.Explore},
		{"map-explore-2", glyph.Explore},
		// verify.
		{"gentle-ai-verify", glyph.Verify},
		{"verify-spec", glyph.Verify},
		// worker.
		{"gentle-ai-worker", glyph.Worker},
		{"worker-mwmmo5ju", glyph.Worker},
		// judge: prefix jd-judge or contained.
		{"jd-judge-a", glyph.Judge},
		{"jd-judge-b", glyph.Judge},
		{"the-judge", glyph.Judge},
		// fix: prefix jd-fix or contained.
		{"jd-fix-1", glyph.Fix},
		{"jd-fix-12", glyph.Fix},
		{"quickfixer", glyph.Fix},
		// review: prefix only.
		{"review-1", glyph.Review},
		{"review-spec", glyph.Review},
		// unknown: no family word anywhere.
		{"gentle-ai", glyph.Unknown},
		{"orchestrator", glyph.Unknown},
		{"", glyph.Unknown},
		// case-insensitive, in every family.
		{"GENTLE-AI-EXPLORE", glyph.Explore},
		{"JD-JUDGE-C", glyph.Judge},
		{"Worker-7", glyph.Worker},
		{"REVIEW-3", glyph.Review},
		{"Verify-Spec", glyph.Verify},
		{"jd-FIX-2", glyph.Fix},
	}
	for _, c := range cases {
		if got := glyph.Role(c.agentID); got != c.want {
			t.Errorf("Role(%q) = %q, want %q", c.agentID, got, c.want)
		}
	}
}

func TestStatusMapsTodoStatusesToTheirMark(t *testing.T) {
	cases := []struct {
		status string
		want   string
	}{
		{"pending", glyph.Pending},
		{"in_progress", glyph.InProgress},
		{"blocked", glyph.Blocked},
		{"done", glyph.Done},
		{"dropped", glyph.Dropped},
		// Aliases of dropped.
		{"cancelled", glyph.Dropped},
		{"canceled", glyph.Dropped},
		{"abandoned", glyph.Dropped},
		// Case-insensitive.
		{"PENDING", glyph.Pending},
		{"In_Progress", glyph.InProgress},
		{"DONE", glyph.Done},
		{"Cancelled", glyph.Dropped},
		// An unknown status has no mark.
		{"archived", glyph.Unknown},
		{"", glyph.Unknown},
	}
	for _, c := range cases {
		if got := glyph.Status(c.status); got != c.want {
			t.Errorf("Status(%q) = %q, want %q", c.status, got, c.want)
		}
	}
}

// One glyph, one meaning, one cell wide: every glyph this package exports
// measures exactly one cell with the kit's own width function, so a glyph can
// never shift a column.
func TestEveryGlyphIsOneCellWide(t *testing.T) {
	glyphs := map[string]string{
		"Explore":    glyph.Explore,
		"Verify":     glyph.Verify,
		"Worker":     glyph.Worker,
		"Judge":      glyph.Judge,
		"Fix":        glyph.Fix,
		"Review":     glyph.Review,
		"Unknown":    glyph.Unknown,
		"Pending":    glyph.Pending,
		"InProgress": glyph.InProgress,
		"Blocked":    glyph.Blocked,
		"Done":       glyph.Done,
		"Dropped":    glyph.Dropped,
	}
	for name, g := range glyphs {
		if width := text.Width(g); width != 1 {
			t.Errorf("glyph %s = %q (U+%04X) measures %d cells, want 1",
				name, g, []rune(g)[0], width)
		}
	}
}
