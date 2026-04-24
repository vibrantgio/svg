// Implements a raster backend to render SVG images,
// by wrapping github.com/srwiley/rasterx.
package driver

import (
	"image"
	"io"

	"golang.org/x/image/math/fixed"

	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/driver"
	"github.com/vibrantgio/svg/parser"

	"github.com/srwiley/rasterx"
)

// assert interface conformance
var _ driver.DrawerNG = (*Driver)(nil)

// Driver is a DrawerNG that rasterises into a rasterx scanner.
// Path ops are buffered and replayed into the rasterx Filler on Fill and
// into the rasterx Dasher on Stroke, since each consumes its own path state.
type Driver struct {
	dasher *rasterx.Dasher
	ops    []svg.Operation
}

// NewDriver returns a renderer with default values,
// which will raster into `scanner`.
func NewDriver(width, height int, scanner rasterx.Scanner) *Driver {
	return &Driver{dasher: rasterx.NewDasher(width, height, scanner)}
}

func (d *Driver) Clear() {
	d.ops = d.ops[:0]
}

// SetWinding forwards the rule to the rasterx scanner. Note that the
// default scanner (rasterx.ScannerGV) does not support the even-odd rule
// and treats this as a no-op; scanFT-style scanners honour it.
func (d *Driver) SetWinding(useNonZeroWinding bool) {
	d.dasher.Scanner.SetWinding(useNonZeroWinding)
}

func (d *Driver) SetStrokeOptions(options driver.StrokeOptions) {
	d.dasher.SetStroke(
		options.LineWidth, options.Join.MiterLimit, capToFunc[options.Join.LeadLineCap],
		capToFunc[options.Join.TrailLineCap], gapToFunc[options.Join.LineGap],
		joinToJoin[options.Join.LineJoin], options.Dash.Dash, options.Dash.DashOffset,
	)
}

func (d *Driver) Start(a fixed.Point26_6)              { d.ops = append(d.ops, svg.OpMoveTo(a)) }
func (d *Driver) Line(b fixed.Point26_6)               { d.ops = append(d.ops, svg.OpLineTo(b)) }
func (d *Driver) QuadBezier(b, c fixed.Point26_6)      { d.ops = append(d.ops, svg.OpQuadTo{b, c}) }
func (d *Driver) CubeBezier(b, c, e fixed.Point26_6)   { d.ops = append(d.ops, svg.OpCubicTo{b, c, e}) }
func (d *Driver) Close()                               { d.ops = append(d.ops, svg.OpClose{}) }

// pather is the subset of rasterx.Filler / rasterx.Dasher used when replaying
// buffered path ops.
type pather interface {
	Start(a fixed.Point26_6)
	Line(b fixed.Point26_6)
	QuadBezier(b, c fixed.Point26_6)
	CubeBezier(b, c, d fixed.Point26_6)
	Stop(isClosed bool)
	Clear()
}

func (d *Driver) replay(p pather) {
	p.Clear()
	started := false
	for _, op := range d.ops {
		switch o := op.(type) {
		case svg.OpMoveTo:
			if started {
				p.Stop(false)
			}
			p.Start(fixed.Point26_6(o))
			started = true
		case svg.OpLineTo:
			p.Line(fixed.Point26_6(o))
		case svg.OpQuadTo:
			p.QuadBezier(o[0], o[1])
		case svg.OpCubicTo:
			p.CubeBezier(o[0], o[1], o[2])
		case svg.OpClose:
			p.Stop(true)
			started = false
		}
	}
	if started {
		p.Stop(false)
	}
}

func (d *Driver) Fill(color svg.Pattern, opacity float64) {
	f := &d.dasher.Filler
	d.replay(f)
	setColorFromPattern(color, opacity, f.Scanner)
	f.Draw()
}

func (d *Driver) Stroke(color svg.Pattern, opacity float64) {
	d.replay(d.dasher)
	setColorFromPattern(color, opacity, d.dasher.Scanner)
	d.dasher.Draw()
}

// RasterSVGIconToImage uses a default scanner rasterx.ScannerGV instance to renderer the
// icon into an image and return it.
func RasterSVGIconToImage(data io.Reader) (*image.RGBA, error) {
	svg := parser.NewParser(parser.WarnErrorMode)
	parsedIcon, err := svg.ParseStream(data)
	if err != nil {
		return nil, err
	}
	vb := parsedIcon.ViewBox
	w, h := int(vb.W), int(vb.H)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	renderer := NewDriver(w, h, scanner)
	driver.Draw(renderer, parsedIcon, 1.0)
	return img, nil
}

func toRasterxGradient(grad svg.Gradient) rasterx.Gradient {
	var (
		points   [5]float64
		isRadial bool
	)
	switch grad.Direction.Type() {
	case svg.Linear:
		dir := grad.Direction.Params()
		points[0], points[1], points[2], points[3] = dir[0], dir[1], dir[2], dir[3]
		isRadial = false
	case svg.Radial:
		dir := grad.Direction.Params()
		points[0], points[1], points[2], points[3], points[4], _ = dir[0], dir[1], dir[2], dir[3], dir[4], dir[5] // in rasterx fr is ignored
		isRadial = true
	}
	stops := make([]rasterx.GradStop, len(grad.Stops))
	for i := range grad.Stops {
		stops[i] = rasterx.GradStop(grad.Stops[i])
	}
	return rasterx.Gradient{
		Points:   points,
		Stops:    stops,
		Bounds:   grad.Bounds,
		Matrix:   rasterx.Matrix2D(grad.Matrix),
		Spread:   rasterx.SpreadMethod(grad.Spread),
		Units:    rasterx.GradientUnits(grad.Units),
		IsRadial: isRadial,
	}
}

// resolve gradient color
func setColorFromPattern(color svg.Pattern, opacity float64, scanner rasterx.Scanner) {
	switch color := color.(type) {
	case svg.PlainColor:
		scanner.SetColor(rasterx.ApplyOpacity(color, opacity))
	case svg.Gradient:
		_ = color.ApplyPathExtent(scanner.GetPathExtent())
		rasterxGradient := toRasterxGradient(color)
		scanner.SetColor(rasterxGradient.GetColorFunction(opacity))
	}
}

var (
	joinToJoin = [...]rasterx.JoinMode{
		svg.Round:     rasterx.Round,
		svg.Bevel:     rasterx.Bevel,
		svg.Miter:     rasterx.Miter,
		svg.MiterClip: rasterx.MiterClip,
		svg.Arc:       rasterx.Arc,
		svg.ArcClip:   rasterx.ArcClip,
	}

	capToFunc = [...]rasterx.CapFunc{
		svg.ButtCap:      rasterx.ButtCap,
		svg.SquareCap:    rasterx.SquareCap,
		svg.RoundCap:     rasterx.RoundCap,
		svg.CubicCap:     rasterx.CubicCap,
		svg.QuadraticCap: rasterx.QuadraticCap,
	}

	gapToFunc = [...]rasterx.GapFunc{
		svg.FlatGap:      rasterx.FlatGap,
		svg.RoundGap:     rasterx.RoundGap,
		svg.CubicGap:     rasterx.CubicGap,
		svg.QuadraticGap: rasterx.QuadraticGap,
	}
)
