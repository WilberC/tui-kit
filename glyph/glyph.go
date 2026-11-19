// Package glyph names the per-kind and per-state glyphs a dock renders on
// agent and todo rows: one glyph, one meaning, one cell wide.
//
// Two families live here. The role glyphs answer what a subagent *is* (its
// family of work) and sit between the row's status mark and the name — the
// mark owns state, the glyph owns kind. The status marks answer where a todo
// item stands in gentle-shell's todo vocabulary. Every exported glyph is
// pinned to one cell by a test against the kit's own width function
// (text.Width), so a glyph can never shift a column.
//
// The shipped set is plain Unicode, one cell wide in the common monospace
// fonts. Nerd Font alternates are recorded in DESIGN.md §6 for later; none is
// shipped.
package glyph

import "strings"

// Role glyphs: the family of work a subagent row is doing. The Unknown glyph
// is the kit's existing "no value" glyph, rendered muted.
const (
	Explore = "◎" // U+25CE · explore  (Nerd Font alternate U+F06E, not shipped)
	Verify  = "✓" // U+2713 · verify   (Nerd Font alternate U+F00C, not shipped)
	Worker  = "⚒" // U+2692 · worker   (Nerd Font alternate U+F0AD, not shipped)
	Judge   = "⚖" // U+2696 · judge    (Nerd Font alternate U+F0E3, not shipped)
	Fix     = "✚" // U+271A · fix      (Nerd Font alternate U+F0FA, not shipped)
	Review  = "⌕" // U+2315 · review   (Nerd Font alternate U+F002, not shipped)
	Unknown = "–" // U+2013 · no value; the "no key / no value" glyph, muted
)

// Status marks: where a todo item stands. Done folds into a header count
// (`✓ N done`) rather than a per-row mark, and Dropped is excluded from the
// counts entirely.
const (
	Pending    = "○" // muted, matches idle
	InProgress = "◐" // live / work in progress
	Blocked    = "⊘" // red
	Done       = "✓" // green, usually folded into a count
	Dropped    = "✕" // muted, excluded from the counts
)

// Role maps an agent id to its family glyph.
//
// The id is matched case-insensitively: a word contained anywhere in the id
// names the family (gentle-ai-explore, *-verify, *-worker, the-judge,
// quickfixer), the jd- prefixes name their family explicitly (jd-judge-*,
// jd-fix-*), and review is matched by prefix only (review-*), so a product
// named "reviewer-extra" is not read as a review row. Anything else is the
// muted Unknown glyph.
func Role(agentID string) string {
	id := strings.ToLower(agentID)
	switch {
	case strings.Contains(id, "explore"):
		return Explore
	case strings.Contains(id, "verify"):
		return Verify
	case strings.Contains(id, "worker"):
		return Worker
	case strings.HasPrefix(id, "jd-judge"), strings.Contains(id, "judge"):
		return Judge
	case strings.HasPrefix(id, "jd-fix"), strings.Contains(id, "fix"):
		return Fix
	case strings.HasPrefix(id, "review-"):
		return Review
	default:
		return Unknown
	}
}

// Status maps a todo status word to its mark, case-insensitively. The
// dropped status answers to its aliases cancelled, canceled and abandoned.
// Hyphen and space forms of in_progress are accepted (in-progress, in
// progress). An unknown status has no mark: it answers the muted Unknown
// glyph, the kit's "no value" glyph.
func Status(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pending":
		return Pending
	case "in_progress", "in-progress", "in progress":
		return InProgress
	case "blocked":
		return Blocked
	case "done":
		return Done
	case "dropped", "cancelled", "canceled", "abandoned":
		return Dropped
	default:
		return Unknown
	}
}
