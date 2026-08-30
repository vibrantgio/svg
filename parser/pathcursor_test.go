package parser

import (
	"strings"
	"testing"
)

// icon wraps a path d attribute in a minimal SVG document and parses it.
func parsePath(t *testing.T, d string) string {
	t.Helper()
	doc := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="` + d + `"/></svg>`
	icon, err := NewParser(StrictErrorMode).ParseStream(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("parse %q: %v", d, err)
	}
	if len(icon.Paths) != 1 {
		t.Fatalf("parse %q: %d paths, want 1", d, len(icon.Paths))
	}
	return icon.Paths[0].Path.ToSVGPath()
}

// TestChainedArcsMatchSplitArcs verifies every 7-value set of a chained arc
// command draws its own arc: one command carrying two arcs must produce the
// same path as the equivalent two single-arc commands.
func TestChainedArcsMatchSplitArcs(t *testing.T) {
	chained := parsePath(t, "M6 12a6 6 0 0 1 6-6 4 4 0 0 1 4 4")
	split := parsePath(t, "M6 12a6 6 0 0 1 6-6a4 4 0 0 1 4 4")
	if chained != split {
		t.Errorf("chained arcs:\n%s\nsplit arcs:\n%s\nwant identical paths", chained, split)
	}
}

// TestChainedArcsAdvance verifies the current point tracks each chained
// arc's own endpoint rather than sticking at the first one.
func TestChainedArcsAdvance(t *testing.T) {
	got := parsePath(t, "M6 12a6 6 0 0 1 6-6 4 4 0 0 1 4 4L0 0")
	if !strings.Contains(got, "L0.000,0.000") {
		t.Fatalf("path lost its trailing line: %s", got)
	}
	// The final arc endpoint precedes the line.
	if !strings.Contains(got, "16.000,10.000") {
		t.Errorf("path never reaches the second arc's endpoint 16,10: %s", got)
	}
}
