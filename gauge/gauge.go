// Package gauge draws micro-bars: one row of filled and light cells for a
// share of a scale (H), or a column filled from the bottom (V). No graphs,
// no animation — a gauge is the smallest thing a readout can lean on. The
// caller owns the semantics: it passes the value, the scale the full bar
// stands for, and paints the parts with whatever theme roles its thresholds
// call for.
package gauge

import "strings"

// The cells. Filled carries the value; Light carries the room left.
const (
	Filled = "█"
	Light  = "░"
)

// H splits a horizontal bar into its filled and light parts: value over
// scale, n cells wide. Both parts come back unpainted so the caller renders
// each with its own role (typically the filled part colored, the light part
// quiet). Scale 0 or a negative value yields an empty bar; a value beyond
// the scale is a full one.
func H(value, scale float64, n int) (filled, empty string) {
	f := share(value, scale, n)
	return strings.Repeat(Filled, f), strings.Repeat(Light, n-f)
}

// V draws a vertical bar: n rows, top-down, the column filled from the
// bottom — row 0 is the top of the bar. Each row is one cell. The value
// beyond the scale fills every row; an unreported value has no bar here
// (the caller renders absence its own way).
func V(value, scale float64, n int) []string {
	f := share(value, scale, n)
	rows := make([]string, n)
	for i := 0; i < n; i++ {
		if i < n-f {
			rows[i] = Light
		} else {
			rows[i] = Filled
		}
	}
	return rows
}

// share clamps value/scale into 0..n cells.
func share(value, scale float64, n int) int {
	if scale <= 0 || value <= 0 {
		return 0
	}
	f := int(value/scale*float64(n) + 0.5)
	if f > n {
		f = n
	}
	return f
}
