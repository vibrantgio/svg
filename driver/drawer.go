package driver

import (
	"golang.org/x/image/math/fixed"

	"github.com/vibrantgio/svg"
)

// DrawerNG is the interface a rendering backend implements.
// A caller describes a single path by calling Clear, SetWinding, the path
// building methods (Start / Line / QuadBezier / CubeBezier / Close), then
// Fill and/or Stroke. Fill and Stroke may each consume the path, so backends
// that cannot re-use their internal path representation must buffer the
// ops internally and replay them on each call.
//
// Transformations are already applied to the points before they reach the
// drawer.
type DrawerNG interface {
	// Clear resets the internal state before starting a new path.
	Clear()

	// SetWinding selects the fill rule for the current path.
	// If useNonZeroWinding is false, the even-odd rule is used.
	SetWinding(useNonZeroWinding bool)

	// SetStrokeOptions parametrises the stroke style for the current path.
	SetStrokeOptions(options StrokeOptions)

	// Start begins a new sub-path at the given point.
	Start(a fixed.Point26_6)

	// Line adds a straight segment from the current point to b.
	Line(b fixed.Point26_6)

	// QuadBezier adds a quadratic bezier segment.
	QuadBezier(b, c fixed.Point26_6)

	// CubeBezier adds a cubic bezier segment.
	CubeBezier(b, c, d fixed.Point26_6)

	// Close closes the current sub-path back to its start point.
	Close()

	// Fill fills the accumulated path with the given colour.
	Fill(color svg.Pattern, opacity float64)

	// Stroke strokes the accumulated path with the given colour.
	Stroke(color svg.Pattern, opacity float64)
}

// FillAndStroker is an optional interface for backends that can fill and
// stroke a single path in one native operation (for example PDF's B
// operator). When a backend implements it, draw.Draw prefers FillAndStroke
// over separate Fill+Stroke calls.
type FillAndStroker interface {
	FillAndStroke(fillCol svg.Pattern, fillOp float64,
		strokeCol svg.Pattern, strokeOp float64)
}

// StrokeOptions parametrises a stroke.
type StrokeOptions struct {
	LineWidth fixed.Int26_6
	Dash      svg.DashOptions
	Join      svg.JoinOptions
}
