package frame_test

import (
	"strings"
	"testing"
)

// TestBodyRowsForAnswersBeforeTheRowsExist pins the row count a product needs
// first: a header that names the body's page asks the surface how many body rows
// the edges it is about to draw leave, so no placeholder row stands in for it.
func TestBodyRowsForAnswersBeforeTheRowsExist(t *testing.T) {
	planned := bounded(80)
	planned.Top, planned.Bottom = nil, nil // nothing composed yet

	drawn := bounded(80)
	for _, height := range []int{20, 60, 12, 5} {
		if got, want := planned.BodyRowsFor(height, len(drawn.Top), len(drawn.Bottom)), drawn.BodyRows(height); got != want {
			t.Fatalf("BodyRowsFor(%d, %d, %d) = %d, want the drawn %d", height, len(drawn.Top), len(drawn.Bottom), got, want)
		}
	}

	// A header that grows by a row is a body row less, and the count says so
	// before any of the rows exists.
	if got, want := planned.BodyRowsFor(20, 1, 2), 13; got != want {
		t.Fatalf("BodyRowsFor(20, 1, 2) = %d, want %d", got, want)
	}
	if got, want := planned.BodyRowsFor(20, 2, 2), 12; got != want {
		t.Fatalf("BodyRowsFor(20, 2, 2) = %d, want %d", got, want)
	}
	// A footer too tall for the pane keeps its last rows and drops the header, so
	// the body has no room at all instead of a negative count.
	if got := planned.BodyRowsFor(5, 1, 3); got != 0 {
		t.Fatalf("BodyRowsFor(5, 1, 3) = %d, want 0", got)
	}
}

// TestBodyRowsForAgreesWithTheRenderedBody pins the other half: the number the
// header was built from is the number of content rows the surface actually shows
// once the edges are in place.
func TestBodyRowsForAgreesWithTheRenderedBody(t *testing.T) {
	surface := bounded(80)
	planned := surface.BodyRowsFor(20, 2, 2)

	surface.Top = surface.Rows(80, "HEADER", "STATUS")
	surface.Bottom = surface.Rows(80, "DETAIL", "HINTS")
	if got, want := surface.BodyRows(20), planned; got != want {
		t.Fatalf("the drawn body holds %d rows, want the planned %d", got, want)
	}

	body := make([]string, planned)
	for index := range body {
		body[index] = surface.Row(80, "row")
	}
	lines := strings.Split(surface.Render(80, 20, body), "\n")
	if got, want := len(lines), 20; got != want {
		t.Fatalf("the surface drew %d rows, want %d", got, want)
	}
	// Every planned row is inside the frame, and the last one sits just above the
	// body's closing padding row.
	if got := rowContent(lines[3+planned]); got != "row" {
		t.Fatalf("the last planned body row = %q, want a content row", got)
	}
	if got := rowContent(lines[4+planned]); got != "" {
		t.Fatalf("the body does not close with a blank row: %q", got)
	}
}

// TestBodyRowsKeepsThePlannedRowCount is the regression guard for the old call:
// BodyRows still answers from the rows the surface holds.
func TestBodyRowsKeepsThePlannedRowCount(t *testing.T) {
	surface := bounded(80)
	if got, want := surface.BodyRows(20), surface.BodyRowsFor(20, 1, 2); got != want {
		t.Fatalf("BodyRows(20) = %d, want BodyRowsFor(20, 1, 2) = %d", got, want)
	}
	if got, want := surface.BodyRows(20), 13; got != want {
		t.Fatalf("BodyRows(20) = %d, want %d", got, want)
	}
}
