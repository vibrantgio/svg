// Package gio implements a Gio-UI backend for SVG rendering.
//
// Current limitations:
//   - Stroke line cap and line join are ignored — Gio's clip.Stroke does
//     not expose them. Every stroke renders as butt cap / miter join.
//   - Stroke dashing is ignored; dashed strokes render as solid strokes.
//   - Linear gradients with exactly two stops and pad spread render via
//     paint.LinearGradientOp. All other linear gradients are rasterised
//     into an off-screen image.
//   - Radial gradients are stubbed: the fill is a solid colour taken from
//     the first stop.
//   - Gio's clip.Outline implements the non-zero-winding rule only;
//     SetWinding's argument is ignored.
package gio

import (
	"image"
	"image/color"
	"math"

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

// pathExtent returns the axis-aligned bounding box of buffered ops,
// including off-curve control points (conservative).
func (d *Driver) pathExtent() fixed.Rectangle26_6 {
	var r fixed.Rectangle26_6
	first := true
	update := func(p fixed.Point26_6) {
		if first {
			r.Min, r.Max = p, p
			first = false
			return
		}
		if p.X < r.Min.X {
			r.Min.X = p.X
		}
		if p.Y < r.Min.Y {
			r.Min.Y = p.Y
		}
		if p.X > r.Max.X {
			r.Max.X = p.X
		}
		if p.Y > r.Max.Y {
			r.Max.Y = p.Y
		}
	}
	for _, o := range d.ops {
		switch o := o.(type) {
		case svg.OpMoveTo:
			update(fixed.Point26_6(o))
		case svg.OpLineTo:
			update(fixed.Point26_6(o))
		case svg.OpQuadTo:
			update(o[0])
			update(o[1])
		case svg.OpCubicTo:
			update(o[0])
			update(o[1])
			update(o[2])
		}
	}
	return r
}

func (d *Driver) Fill(col svg.Pattern, opacity float64) {
	shape := clip.Outline{Path: d.buildPath()}.Op()
	d.paintPattern(col, opacity, shape)
}

func (d *Driver) Stroke(col svg.Pattern, opacity float64) {
	shape := clip.Stroke{
		Path:  d.buildPath(),
		Width: float32(d.strokeOptions.LineWidth) / 64,
	}.Op()
	d.paintPattern(col, opacity, shape)
}

// paintPattern paints the given clip shape with a solid colour or gradient.
func (d *Driver) paintPattern(col svg.Pattern, opacity float64, shape clip.Op) {
	switch c := col.(type) {
	case svg.PlainColor:
		c.A = uint8(opacity * float64(c.A))
		paint.FillShape(d.Ops, color.NRGBA(c), shape)
	case svg.Gradient:
		d.paintGradient(c, opacity, shape)
	}
}

func (d *Driver) paintGradient(g svg.Gradient, opacity float64, shape clip.Op) {
	switch g.Direction.Type() {
	case svg.Linear:
		d.paintLinearGradient(g, opacity, shape)
	case svg.Radial:
		// TODO: proper radial gradient rendering. Stub with first stop.
		if len(g.Stops) > 0 {
			paint.FillShape(d.Ops, stopNRGBA(g.Stops[0], opacity), shape)
		}
	}
}

func (d *Driver) paintLinearGradient(g svg.Gradient, opacity float64, shape clip.Op) {
	if len(g.Stops) == 0 {
		return
	}
	if len(g.Stops) == 1 {
		paint.FillShape(d.Ops, stopNRGBA(g.Stops[0], opacity), shape)
		return
	}

	params := g.Direction.Params() // [x1, y1, x2, y2]
	extent := d.pathExtent()
	// ApplyPathExtent mutates g.Bounds when Units == ObjectBoundingBox,
	// and returns a scale+matrix that maps gradient-local params into the
	// bounding-box-relative offset (still needs +Bounds.Min to land in
	// pixel space).
	m := g.ApplyPathExtent(extent)
	s1x, s1y := m.Transform(params[0], params[1])
	s2x, s2y := m.Transform(params[2], params[3])
	if g.Units == svg.ObjectBoundingBox {
		s1x += g.Bounds.X
		s1y += g.Bounds.Y
		s2x += g.Bounds.X
		s2y += g.Bounds.Y
	}
	stop1 := f32.Point{X: float32(s1x), Y: float32(s1y)}
	stop2 := f32.Point{X: float32(s2x), Y: float32(s2y)}

	// Fast path: Gio's LinearGradientOp pads endpoint colours, which matches
	// SVG PadSpread. Only usable for exactly two stops.
	if len(g.Stops) == 2 && g.Spread == svg.PadSpread {
		defer shape.Push(d.Ops).Pop()
		paint.LinearGradientOp{
			Stop1:  stop1,
			Color1: stopNRGBA(g.Stops[0], opacity),
			Stop2:  stop2,
			Color2: stopNRGBA(g.Stops[1], opacity),
		}.Add(d.Ops)
		paint.PaintOp{}.Add(d.Ops)
		return
	}

	// Fallback: rasterise the gradient into an image sized to the bounding
	// box, then paint it through the clip shape.
	minX := math.Floor(float64(extent.Min.X) / 64)
	minY := math.Floor(float64(extent.Min.Y) / 64)
	maxX := math.Ceil(float64(extent.Max.X) / 64)
	maxY := math.Ceil(float64(extent.Max.Y) / 64)
	w := int(maxX - minX)
	h := int(maxY - minY)
	if w <= 0 || h <= 0 {
		return
	}
	img := rasterizeLinearGradient(g, opacity, minX, minY, w, h, stop1, stop2)

	// Clip first (path is in absolute pixel space), then offset so the
	// image lands at the bounding box origin.
	defer shape.Push(d.Ops).Pop()
	defer op.Offset(image.Pt(int(minX), int(minY))).Push(d.Ops).Pop()
	paint.NewImageOp(img).Add(d.Ops)
	paint.PaintOp{}.Add(d.Ops)
}

// rasterizeLinearGradient paints a multi-stop or non-pad linear gradient into
// an RGBA image sized to (w, h) pixels at offset (ox, oy) in absolute pixel
// space. stop1/stop2 are the gradient endpoints in absolute pixel space.
func rasterizeLinearGradient(g svg.Gradient, opacity, ox, oy float64, w, h int, stop1, stop2 f32.Point) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	dx := float64(stop2.X - stop1.X)
	dy := float64(stop2.Y - stop1.Y)
	lenSq := dx*dx + dy*dy
	if lenSq == 0 {
		c := stopNRGBA(g.Stops[0], opacity)
		fill := color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				img.SetRGBA(x, y, fill)
			}
		}
		return img
	}
	for y := 0; y < h; y++ {
		py := oy + float64(y) + 0.5
		for x := 0; x < w; x++ {
			px := ox + float64(x) + 0.5
			t := ((px-float64(stop1.X))*dx + (py-float64(stop1.Y))*dy) / lenSq
			t = applySpread(t, g.Spread)
			c := sampleStops(g.Stops, t, opacity)
			// Gio's ImageOp expects premultiplied alpha (color.RGBA).
			a := uint32(c.A)
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(uint32(c.R) * a / 0xff),
				G: uint8(uint32(c.G) * a / 0xff),
				B: uint8(uint32(c.B) * a / 0xff),
				A: c.A,
			})
		}
	}
	return img
}

func applySpread(t float64, spread svg.SpreadMethod) float64 {
	switch spread {
	case svg.ReflectSpread:
		f := math.Floor(t)
		frac := t - f
		if int(math.Abs(f))%2 != 0 {
			frac = 1 - frac
		}
		return frac
	case svg.RepeatSpread:
		return t - math.Floor(t)
	default:
		if t < 0 {
			return 0
		}
		if t > 1 {
			return 1
		}
		return t
	}
}

func sampleStops(stops []svg.GradStop, t, opacity float64) color.NRGBA {
	if t <= stops[0].Offset {
		return stopNRGBA(stops[0], opacity)
	}
	last := stops[len(stops)-1]
	if t >= last.Offset {
		return stopNRGBA(last, opacity)
	}
	for i := 0; i < len(stops)-1; i++ {
		a, b := stops[i], stops[i+1]
		if t >= a.Offset && t <= b.Offset {
			span := b.Offset - a.Offset
			if span == 0 {
				return stopNRGBA(b, opacity)
			}
			u := (t - a.Offset) / span
			ca, cb := stopNRGBA(a, opacity), stopNRGBA(b, opacity)
			return color.NRGBA{
				R: uint8(float64(ca.R)*(1-u) + float64(cb.R)*u),
				G: uint8(float64(ca.G)*(1-u) + float64(cb.G)*u),
				B: uint8(float64(ca.B)*(1-u) + float64(cb.B)*u),
				A: uint8(float64(ca.A)*(1-u) + float64(cb.A)*u),
			}
		}
	}
	return stopNRGBA(last, opacity)
}

func stopNRGBA(s svg.GradStop, opacity float64) color.NRGBA {
	c := color.NRGBAModel.Convert(s.StopColor).(color.NRGBA)
	c.A = uint8(float64(c.A) * s.Opacity * opacity)
	return c
}
