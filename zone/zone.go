// Package zone records where a surface drew each clickable element and answers
// which element a cell hit, so a product never recomputes the layout the kit
// already knows.
//
// A screen builds a Map while it renders: the frame translates it from its own
// content coordinates into the pane's, and the product asks Hit with the cell a
// click landed on. Nothing here draws, reads the terminal or writes anything, so
// a map costs one slice per frame and can be inspected in a test.
//
// The kit records only what the rules allow to be clicked — a chip and a row,
// never loose text — and every action a zone carries is reachable by keyboard
// too. A map is a convenience, not a second interface.
package zone

// Rect is a rectangle of terminal cells: X and Y are its top-left cell and W and
// H its size in cells.
type Rect struct {
	X int
	Y int
	W int
	H int
}

// Empty reports whether the rect covers no cell at all.
func (r Rect) Empty() bool { return r.W <= 0 || r.H <= 0 }

// Contains reports whether the cell (x, y) is inside the rect. The edges are
// half-open, so the cell at X+W is already the next element.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Right is the first cell past the rect's right edge.
func (r Rect) Right() int { return r.X + r.W }

// Bottom is the first row past the rect's bottom edge.
func (r Rect) Bottom() int { return r.Y + r.H }

// Intersect is the part of the rect that lies inside other, which covers nothing
// when the two do not overlap.
func (r Rect) Intersect(other Rect) Rect {
	left, top := max(r.X, other.X), max(r.Y, other.Y)
	right, bottom := min(r.Right(), other.Right()), min(r.Bottom(), other.Bottom())
	if right <= left || bottom <= top {
		return Rect{}
	}
	return Rect{X: left, Y: top, W: right - left, H: bottom - top}
}

// Zone is one named clickable rectangle: the id a product acts on, and where the
// kit drew it.
type Zone struct {
	ID   string
	Rect Rect
}

// Map is the zones one render recorded, in the order they were drawn. The last
// zone drawn at a cell wins, because the last one drawn is the one on top.
//
// A Map is not safe for concurrent use; a screen builds its own per frame. Every
// method tolerates a nil map, so a screen with no mouse support passes nil and
// the kit silently records nothing.
type Map struct {
	zones []Zone
}

// New returns an empty map. A zero Map is also usable, so New is only a habit.
func New() *Map { return &Map{} }

// Add records a zone with a positive area. A rect with no area covers no cell, so
// it is not recorded at all: a click can never land on it.
func (m *Map) Add(id string, x, y, w, h int) {
	m.AddRect(id, Rect{X: x, Y: y, W: w, H: h})
}

// AddRect records a zone with a positive area.
func (m *Map) AddRect(id string, rect Rect) {
	if m == nil || rect.Empty() {
		return
	}
	m.zones = append(m.zones, Zone{ID: id, Rect: rect})
}

// Hit answers which element a cell hit: the id of the last zone drawn there, and
// whether any zone was drawn there at all.
func (m *Map) Hit(x, y int) (string, bool) {
	zone, ok := m.Lookup(x, y)
	return zone.ID, ok
}

// Lookup is Hit with the whole zone, for a product that reacts to where the click
// landed inside it.
func (m *Map) Lookup(x, y int) (Zone, bool) {
	if m == nil {
		return Zone{}, false
	}
	for index := len(m.zones) - 1; index >= 0; index-- {
		if m.zones[index].Rect.Contains(x, y) {
			return m.zones[index], true
		}
	}
	return Zone{}, false
}

// Find is the rect a named zone was drawn in, for a product that must place
// something next to it (a tooltip, a marker).
func (m *Map) Find(id string) (Rect, bool) {
	if m == nil {
		return Rect{}, false
	}
	for index := len(m.zones) - 1; index >= 0; index-- {
		if m.zones[index].ID == id {
			return m.zones[index].Rect, true
		}
	}
	return Rect{}, false
}

// Zones is every zone in draw order. The returned slice is the map's own; a
// caller that keeps it after the next render must copy it.
func (m *Map) Zones() []Zone {
	if m == nil {
		return nil
	}
	return m.zones
}

// Len is how many zones the map holds.
func (m *Map) Len() int {
	if m == nil {
		return 0
	}
	return len(m.zones)
}

// Reset forgets every zone, so a screen can reuse the map it holds instead of
// allocating one per frame.
func (m *Map) Reset() {
	if m == nil {
		return
	}
	m.zones = m.zones[:0]
}

// Offset moves every zone, which is how a surface that knows where its content
// landed turns body coordinates into pane coordinates.
func (m *Map) Offset(dx, dy int) {
	if m == nil {
		return
	}
	for index := range m.zones {
		m.zones[index].Rect.X += dx
		m.zones[index].Rect.Y += dy
	}
}

// Clip keeps only the cells inside bounds: a zone the surface half drew keeps its
// visible part, and a zone it did not draw at all disappears instead of
// answering a click on nothing.
func (m *Map) Clip(bounds Rect) {
	if m == nil {
		return
	}
	kept := m.zones[:0]
	for _, zone := range m.zones {
		if visible := zone.Rect.Intersect(bounds); !visible.Empty() {
			zone.Rect = visible
			kept = append(kept, zone)
		}
	}
	m.zones = kept
}

// Append records every zone of other after the ones already here, which is how
// nested surfaces compose: the inner map keeps its draw order below the outer.
func (m *Map) Append(other *Map) {
	if m == nil || other == nil {
		return
	}
	m.zones = append(m.zones, other.zones...)
}
