package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/WilberC/tui-kit/text"
)

// galleryDir is where the captures live, relative to this package.
func galleryDir() string { return filepath.Join("..", "..", "docs", "gallery") }

// TestEveryStateHasACapture pins the gallery's promise: every state the generator
// knows is a file the documentation can show. It needs no freeze, so a missing
// image fails the suite instead of a reviewer's eye.
func TestEveryStateHasACapture(t *testing.T) {
	for _, s := range states {
		path := filepath.Join(galleryDir(), s.Name+".png")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("state %q has no capture: run `go run ./examples/gallery` (%v)", s.Name, err)
		}
		if info.Size() == 0 {
			t.Fatalf("state %q has an empty capture: %s", s.Name, path)
		}
	}
}

// TestTheGalleryAndDESIGNAgree pins both directions between the generator, the
// images on disk and the guide: DESIGN.md shows every capture, every link in
// DESIGN.md resolves, and no image rots in the directory unshown.
func TestTheGalleryAndDESIGNAgree(t *testing.T) {
	design, err := os.ReadFile(filepath.Join("..", "..", "DESIGN.md"))
	if err != nil {
		t.Fatalf("read DESIGN.md: %v", err)
	}

	reference := regexp.MustCompile(`docs/gallery/([a-z0-9-]+)\.png`)
	shown := map[string]bool{}
	for _, match := range reference.FindAllStringSubmatch(string(design), -1) {
		shown[match[1]] = true
	}

	known := map[string]bool{}
	for _, s := range states {
		known[s.Name] = true
		if !shown[s.Name] {
			t.Fatalf("DESIGN.md does not show %s.png", s.Name)
		}
	}
	for name := range shown {
		if !known[name] {
			t.Fatalf("DESIGN.md shows %s.png, which the gallery does not generate: %s", name, names())
		}
	}

	entries, err := os.ReadDir(galleryDir())
	if err != nil {
		t.Fatalf("read the gallery: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".png") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".png")
		if !shown[name] {
			t.Fatalf("docs/gallery/%s is not shown by DESIGN.md", entry.Name())
		}
	}
}

// TestEveryDrawnStateIsExactlyItsCaptureSize pins the other half of the size
// discipline: a drawn state IS its frame — no centring, no clamped spare rows — so
// the guide never shows a body the size of half a pane.
func TestEveryDrawnStateIsExactlyItsCaptureSize(t *testing.T) {
	for _, s := range states {
		if s.Draw == nil {
			continue // the demo captures measure the demo, not the gallery
		}
		lines := strings.Split(s.Draw(s.Width, s.Height), "\n")
		if len(lines) != s.Height {
			t.Fatalf("state %q draws %d rows, want its %d", s.Name, len(lines), s.Height)
		}
		for index, line := range lines {
			if width := widthOf(line); width != s.Width {
				t.Fatalf("state %q line %d is %d cells wide, want %d: %q", s.Name, index, width, s.Width, line)
			}
		}
	}
}

// widthOf is the cell width of a rendered line.
func widthOf(line string) int { return len([]rune(text.Plain(line))) }

// TestEveryStateIsASaneCapture pins the size discipline: a capture is evidence,
// not wallpaper, so no state asks for a pane the kit's own bounds do not offer
// and every state has a distinct name.
func TestEveryStateIsASaneCapture(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range states {
		if seen[s.Name] {
			t.Fatalf("two states are named %q", s.Name)
		}
		seen[s.Name] = true
		if s.Width <= 0 || s.Height <= 0 {
			t.Fatalf("state %q has no size", s.Name)
		}
		if s.Width > 180 || s.Height > 50 {
			t.Fatalf("state %q is %dx%d, wider or taller than the widest documented pane (180x50)", s.Name, s.Width, s.Height)
		}
	}
}
