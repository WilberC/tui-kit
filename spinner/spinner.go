// Package spinner renders the work-in-progress tick of the kit: quadrant blocks,
// which every monospaced font carries, on the palette's working accent. The
// caller owns the tick, so a spinner costs nothing while nothing is happening.
package spinner

import "github.com/WilberC/tui-kit/theme"

// Frames is the cycle the surfaces animate while work is in flight.
var Frames = []string{"▖", "▘", "▝", "▗"}

// Render draws one frame of the cycle on the teal role. A frame index outside
// the cycle wraps, so a caller may count up without normalising it.
func Render(frame int) string {
	if len(Frames) == 0 {
		return ""
	}
	index := frame % len(Frames)
	if index < 0 {
		index += len(Frames)
	}
	return theme.Teal.Render(Frames[index])
}
