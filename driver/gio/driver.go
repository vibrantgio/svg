package gio

import (
	"image/color"

	"golang.org/x/image/math/fixed"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/driver"
)

func NewDriver(ops *op.Ops) *Driver {
	return &Driver{Ops: ops}
}

// Driver will given a parsed SVG document, implements how to draw it on screen.
// This requires a Driver implementation to perform the actual draw operations,
// such as a rasterizer to output .png images or a pdf writer.
type Driver struct {
	fill, stroke bool

	Ops  *op.Ops
	Clip clip.Path
}

func Pt(p fixed.Point26_6) f32.Point {
	return f32.Point{X: float32(p.X) / 64, Y: float32(p.Y) / 64}
}

// SetupDrawers returns the backend painters, and
// will be called at the begining of every path.
// If the `willXXX` boolean is false, the returned drawer should be nil
// to avoid useless operations.
// When both booleans are true, one can assume that the exact same draw operations
// will be performed on the Filler first and then on the Stroker.
// This promise may enable the implementation to avoid duplicating filled and stroked paths.
func (drv *Driver) SetupDrawers(willFill, willStroke bool) (driver.Filler, driver.Stroker) {
	drv.fill = willFill
	drv.stroke = willStroke
	switch {
	case willFill && willStroke:
		return &filler{Driver: drv}, &stroker{Driver: drv}
	case willFill:
		return &filler{Driver: drv}, nil
	case willStroke:
		return nil, &stroker{Driver: drv}
	default:
		return nil, nil
	}
}

// Clear must reset the internal state, before starting a new path painting
func (drv *Driver) Clear() {
	//	drv.Ops.Reset()
	drv.Clip.Begin(drv.Ops)
}

// Start starts a new path at the given point.
func (drv *Driver) Start(a fixed.Point26_6) {
	drv.Clip.MoveTo(Pt(a))
}

// Line Adds a line for the current point to `b`
func (drv *Driver) Line(b fixed.Point26_6) {
	drv.Clip.LineTo(Pt(b))
}

// QuadBezier adds a quadratic bezier curve to the path
func (drv *Driver) QuadBezier(b, c fixed.Point26_6) {
	drv.Clip.QuadTo(Pt(b), Pt(c))
}

// CubeBezier adds a cubic bezier curve to the path
func (drv *Driver) CubeBezier(b, c, d fixed.Point26_6) {
	drv.Clip.CubeTo(Pt(b), Pt(c), Pt(d))
}

// Closes the path to the start point if `closeLoop` is true
func (drv *Driver) Stop(closeLoop bool) {
	if closeLoop {
		drv.Clip.Close()
	}
}

func (drv *Driver) Draw(color svg.Pattern, opacity float64) {}

type filler struct {
	*Driver
	nonZeroWinding bool
}

// Draw fills or strokes the accumulated path using the given color
func (drv *filler) Draw(col svg.Pattern, opacity float64) {
	switch c := col.(type) {
	case svg.PlainColor:
		shape := clip.Outline{Path: drv.Clip.End()}.Op()
		c.A = uint8(opacity * float64(c.A))
		paint.FillShape(drv.Ops, color.NRGBA(c), shape)
	case svg.Gradient:
		// TODO: implement gradient
	}
}

// Decide to use or not the "non-zero winding" rule for the current path
func (drv *filler) SetWinding(useNonZeroWinding bool) {
	drv.nonZeroWinding = useNonZeroWinding
}

type stroker struct {
	*Driver
	strokeOptions driver.StrokeOptions
}

// Clear must reset the internal state, before starting a new path painting
func (drv *stroker) Clear() {
}

// Start starts a new path at the given point.
func (drv *stroker) Start(a fixed.Point26_6) {
}

// Line Adds a line for the current point to `b`
func (drv *stroker) Line(b fixed.Point26_6) {
}

// QuadBezier adds a quadratic bezier curve to the path
func (drv *stroker) QuadBezier(b, c fixed.Point26_6) {
}

// CubeBezier adds a cubic bezier curve to the path
func (drv *stroker) CubeBezier(b, c, d fixed.Point26_6) {
}

// Closes the path to the start point if `closeLoop` is true
func (drv *stroker) Stop(closeLoop bool) {
}

func (drv *stroker) Draw(color svg.Pattern, opacity float64) {
}

// Parametrize the stroking style for the current path
func (drv *stroker) SetStrokeOptions(options driver.StrokeOptions) {
	drv.strokeOptions = options
}
