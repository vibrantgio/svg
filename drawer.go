package svg

import "golang.org/x/image/math/fixed"

// Drawer knows how to do the actual draw operations
// but doesn't need any SVG kwowledge
// In particular, tranformations matrix are already applied to the points
// before sending them to the Drawer.
type Drawer interface {
	// Clear must reset the internal state, before starting a new path painting
	Clear()

	// Start starts a new path at the given point.
	Start(a fixed.Point26_6)

	// Line Adds a line for the current point to `b`
	Line(b fixed.Point26_6)

	// QuadBezier adds a quadratic bezier curve to the path
	QuadBezier(b, c fixed.Point26_6)

	// CubeBezier adds a cubic bezier curve to the path
	CubeBezier(b, c, d fixed.Point26_6)

	// Closes the path to the start point if `closeLoop` is true
	Stop(closeLoop bool)

	// Draw fills or strokes the accumulated path using the given color
	Draw(color Pattern, opacity float64)
}

// DrawerNG is the interface for specifying paths and then filling and/or stroking them.
// This is a simpler interface to use instead of the Driver, Drawer, Filler and Stroker interfaces.
type DrawerNG interface {
	// Clear must reset the internal state, before starting a new path painting
	Clear()

	// Decide to use or not the "non-zero winding" fill rule for the current path
	SetWinding(useNonZeroWinding bool)

	// Parametrize the stroking style for the current path
	SetStrokeOptions(options StrokeOptions)

	// Start starts a new path at the given point.
	Start(a fixed.Point26_6)

	// Line Adds a line for the current point to `b`
	Line(b fixed.Point26_6)

	// QuadBezier adds a quadratic bezier curve to the path
	QuadBezier(b, c fixed.Point26_6)

	// CubeBezier adds a cubic bezier curve to the path
	CubeBezier(b, c, d fixed.Point26_6)

	// Closes the path to the start point
	Close()

	// Fill fills the accumulated path using the given color
	Fill(color Pattern, opacity float64)

	// Stroke will stroke the accumulated path using the given color
	Stroke(color Pattern, opacity float64)
}
