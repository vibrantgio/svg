// Package gio implements a Gio-UI backend for SVG rendering.
//
// Current limitations (see PLAN.md Phase 5 for the work remaining):
//   - Stroke is a stub — stroked SVG elements do not render.
//   - Gradient fills are not implemented.
//   - Gio's clip.Outline uses the non-zero-winding rule; SetWinding is ignored.
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

// assert interface conformance
var _ driver.DrawerNG = (*Driver)(nil)

func NewDriver(ops *op.Ops) *Driver {
	return &Driver{Ops: ops}
}

// Driver is a DrawerNG that emits Gio clip and paint ops.
// Because gioui.org/op/clip.Path is single-use (consumed by clip.Outline /
// clip.Stroke), path operations are buffered and replayed into a fresh
// clip.Path for each of Fill and Stroke.
type Driver struct {
	Ops *op.Ops

	ops           []svg.Operation
	strokeOptions driver.StrokeOptions
}

func pt(p fixed.Point26_6) f32.Point {
	return f32.Point{X: float32(p.X) / 64, Y: float32(p.Y) / 64}
}

func (d *Driver) Clear() {
	d.ops = d.ops[:0]
	d.strokeOptions = driver.StrokeOptions{}
}

func (d *Driver) SetWinding(useNonZeroWinding bool) {}

func (d *Driver) SetStrokeOptions(options driver.StrokeOptions) {
	d.strokeOptions = options
}

func (d *Driver) Start(a fixed.Point26_6)            { d.ops = append(d.ops, svg.OpMoveTo(a)) }
func (d *Driver) Line(b fixed.Point26_6)             { d.ops = append(d.ops, svg.OpLineTo(b)) }
func (d *Driver) QuadBezier(b, c fixed.Point26_6)    { d.ops = append(d.ops, svg.OpQuadTo{b, c}) }
func (d *Driver) CubeBezier(b, c, e fixed.Point26_6) { d.ops = append(d.ops, svg.OpCubicTo{b, c, e}) }
func (d *Driver) Close()                             { d.ops = append(d.ops, svg.OpClose{}) }

// buildPath replays buffered ops into a fresh clip.Path, returning the
// resulting path spec.
func (d *Driver) buildPath() clip.PathSpec {
	var p clip.Path
	p.Begin(d.Ops)
	for _, op := range d.ops {
		switch o := op.(type) {
		case svg.OpMoveTo:
			p.MoveTo(pt(fixed.Point26_6(o)))
		case svg.OpLineTo:
			p.LineTo(pt(fixed.Point26_6(o)))
		case svg.OpQuadTo:
			p.QuadTo(pt(o[0]), pt(o[1]))
		case svg.OpCubicTo:
			p.CubeTo(pt(o[0]), pt(o[1]), pt(o[2]))
		case svg.OpClose:
			p.Close()
		}
	}
	return p.End()
}

func (d *Driver) Fill(col svg.Pattern, opacity float64) {
	switch c := col.(type) {
	case svg.PlainColor:
		shape := clip.Outline{Path: d.buildPath()}.Op()
		c.A = uint8(opacity * float64(c.A))
		paint.FillShape(d.Ops, color.NRGBA(c), shape)
	case svg.Gradient:
		// TODO (Phase 5): implement linear and radial gradients.
	}
}

// Stroke is a stub; see Phase 5.
func (d *Driver) Stroke(col svg.Pattern, opacity float64) {
	// TODO (Phase 5): replay ops into a fresh clip.Path and use clip.Stroke.
}
