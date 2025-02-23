package dummy

import (
	"fmt"

	"golang.org/x/image/math/fixed"

	"github.com/reactivego/svg"
	"github.com/reactivego/svg/driver"
)

func NewDriver() driver.Driver {
	return &dummy{}
}

// Driver will given a parsed SVG document, implements how to draw it on screen.
// This requires a dummy implementation to perform the actual draw operations,
// such as a rasterizer to output .png images or a pdf writer.
type dummy struct {
	optimize       bool
	nonZeroWinding bool
	strokeOptions  driver.StrokeOptions
}

// SetupDrawers returns the backend painters, and
// will be called at the begining of every path.
// If the `willXXX` boolean is false, the returned drawer should be nil
// to avoid useless operations.
// When both booleans are true, one can assume that the exact same draw operations
// will be performed on the Filler first and then on the Stroker.
// This promise may enable the implementation to avoid duplicating filled and stroked paths.
func (drv *dummy) SetupDrawers(willFill, willStroke bool) (driver.Filler, driver.Stroker) {
	fmt.Printf("SetupDrawers willFill=%t willStroke=%t\n", willFill, willStroke)
	drv.optimize = willFill && willStroke
	switch {
	case drv.optimize:
		return drv, drv
	case willFill:
		return drv, nil
	case willStroke:
		return nil, drv
	default:
		return nil, nil
	}
}

// Clear must reset the internal state, before starting a new path painting
func (drv *dummy) Clear() {
	fmt.Println("Clear")
	drv.nonZeroWinding = true
	drv.strokeOptions = driver.StrokeOptions{}
}

// Start starts a new path at the given point.
func (drv *dummy) Start(a fixed.Point26_6) {
	fmt.Printf("Start a=%v\n", a)
}

// Line Adds a line for the current point to `b`
func (drv *dummy) Line(b fixed.Point26_6) {
	fmt.Printf("Line b=%v\n", b)
}

// QuadBezier adds a quadratic bezier curve to the path
func (drv *dummy) QuadBezier(b, c fixed.Point26_6) {
	// Add necessary logic to add a quadratic bezier curve
	fmt.Printf("QuadBezier b=%v c=%v\n", b, c)
}

// CubeBezier adds a cubic bezier curve to the path
func (drv *dummy) CubeBezier(b, c, d fixed.Point26_6) {
	// Add necessary logic to add a cubic bezier curve
	fmt.Printf("CubeBezier b=%v c=%v d=%v\n", b, c, d)
}

// Closes the path to the start point if `closeLoop` is true
func (drv *dummy) Stop(closeLoop bool) {
	fmt.Printf("Stop closeLoop=%t\n", closeLoop)
}

// Draw fills or strokes the accumulated path using the given color
func (drv *dummy) Draw(color svg.Pattern, opacity float64) {
	// Add necessary logic to fill or stroke the path using the given color and opacity
	fmt.Printf("Draw color=%v opacity=%v\n", color, opacity)
}

// Decide to use or not the "non-zero winding" rule for the current path
func (drv *dummy) SetWinding(useNonZeroWinding bool) {
	drv.nonZeroWinding = useNonZeroWinding
	fmt.Printf("SetWinding useNonZeroWinding=%t\n", useNonZeroWinding)
}

// Parametrize the stroking style for the current path
func (drv *dummy) SetStrokeOptions(options driver.StrokeOptions) {
	drv.strokeOptions = options
	fmt.Printf("SetStrokeOptions options=%v\n", options)
}
