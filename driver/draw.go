package driver

import (
	"github.com/reactivego/svg"
	"github.com/reactivego/svg/matrix"
	"golang.org/x/image/math/fixed"
)

// DrawTransformed draws the compiled SvgPath into the driver while applying transform t.
func DrawTransformed(d Driver, paths []svg.StyledPath, t matrix.Matrix2D, opacity float64) {
	for _, path := range paths {
		m := path.Style.Transform
		path.Style.Transform = t.Mult(m)

		filler, stroker := d.SetupDrawers(path.Style.FillColor != nil, path.Style.LineColor != nil)
		if filler != nil { // nil color disable filling
			filler.Clear()
			filler.SetWinding(path.Style.UseNonZeroWinding)

			for _, op := range path.Path {
				DrawOperationTransformed(filler, op, path.Style.Transform)
			}
			filler.Stop(false)

			filler.Draw(path.Style.FillColor, path.Style.FillOpacity*opacity)
			filler.SetWinding(true) // default is true
		}

		if stroker != nil { // nil color disable lining
			stroker.Clear()

			lineGap := path.Style.Join.LineGap
			if lineGap == svg.NilGap {
				lineGap = svg.DefaultStyle.Join.LineGap
			}
			lineCap := path.Style.Join.TrailLineCap
			if lineCap == svg.NilCap {
				lineCap = svg.DefaultStyle.Join.TrailLineCap
			}
			leadLineCap := lineCap
			if path.Style.Join.LeadLineCap != svg.NilCap {
				leadLineCap = path.Style.Join.LeadLineCap
			}
			stroker.SetStrokeOptions(StrokeOptions{
				LineWidth: fixed.Int26_6(path.Style.LineWidth * 64),
				Join: svg.JoinOptions{
					MiterLimit:   path.Style.Join.MiterLimit,
					LineJoin:     path.Style.Join.LineJoin,
					LeadLineCap:  leadLineCap,
					TrailLineCap: lineCap,
					LineGap:      lineGap,
				},
				Dash: path.Style.Dash,
			})

			for _, op := range path.Path {
				DrawOperationTransformed(stroker, op, path.Style.Transform)
			}
			stroker.Stop(false)

			stroker.Draw(path.Style.LineColor, path.Style.LineOpacity*opacity)
		}

		// Restore untransformed matrix
		path.Style.Transform = m
	}
}

func DrawOperationTransformed(d Drawer, op svg.Operation, M matrix.Matrix2D) {
	switch op := op.(type) {
	// starts a new path at the given point.
	case svg.OpMoveTo:
		d.Stop(false) // implicit close if currently in path.
		// transform the operation `op` by applying `M`
		d.Start(M.TFixed(fixed.Point26_6(op)))
	// draw a line
	case svg.OpLineTo:
		// transform the operation `op` by applying `M`
		d.Line(M.TFixed(fixed.Point26_6(op)))
	// draw a quadratic bezier curve
	case svg.OpQuadTo:
		// transform the operation `op` by applying `M`
		b, c := M.TFixed(op[0]), M.TFixed(op[1])
		d.QuadBezier(b, c)
	// draw a cubic bezier curve
	case svg.OpCubicTo:
		// transform the operation `op` by applying `M`
		b, c, d_ := M.TFixed(op[0]), M.TFixed(op[1]), M.TFixed(op[2])
		d.CubeBezier(b, c, d_)
	case svg.OpClose:
		d.Stop(true)
	}
}
