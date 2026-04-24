package driver

import (
	"golang.org/x/image/math/fixed"

	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/matrix"
)

// Draw renders the compiled SVG icon into the drawer d.
// opacity is composed (multiplied) with the <stroke-opacity> and
// <fill-opacity> style attributes. All elements should be contained by the
// icon's ViewBox: see Icon.SetTarget.
func Draw(d DrawerNG, i *svg.Icon, opacity float64) {
	emit := func(op svg.Operation, M matrix.Matrix2D) {
		switch op := op.(type) {
		case svg.OpMoveTo:
			d.Start(M.TFixed(fixed.Point26_6(op)))
		case svg.OpLineTo:
			d.Line(M.TFixed(fixed.Point26_6(op)))
		case svg.OpQuadTo:
			d.QuadBezier(M.TFixed(op[0]), M.TFixed(op[1]))
		case svg.OpCubicTo:
			d.CubeBezier(M.TFixed(op[0]), M.TFixed(op[1]), M.TFixed(op[2]))
		case svg.OpClose:
			d.Close()
		}
	}

	fas, _ := d.(FillAndStroker)

	for _, path := range i.Paths {
		willFill := path.Style.FillColor != nil
		willStroke := path.Style.LineColor != nil
		if !willFill && !willStroke {
			continue
		}

		m := path.Style.Transform
		path.Style.Transform = i.Transform.Mult(m)

		d.Clear()
		d.SetWinding(path.Style.UseNonZeroWinding)

		if willStroke {
			lineGap := path.Style.Join.LineGap
			if lineGap == svg.NilGap {
				lineGap = svg.DefaultStyle.Join.LineGap
			}
			trailLineCap := path.Style.Join.TrailLineCap
			if trailLineCap == svg.NilCap {
				trailLineCap = svg.DefaultStyle.Join.TrailLineCap
			}
			leadLineCap := trailLineCap
			if path.Style.Join.LeadLineCap != svg.NilCap {
				leadLineCap = path.Style.Join.LeadLineCap
			}
			d.SetStrokeOptions(StrokeOptions{
				LineWidth: fixed.Int26_6(path.Style.LineWidth * 64),
				Join: svg.JoinOptions{
					MiterLimit:   path.Style.Join.MiterLimit,
					LineJoin:     path.Style.Join.LineJoin,
					LeadLineCap:  leadLineCap,
					TrailLineCap: trailLineCap,
					LineGap:      lineGap,
				},
				Dash: path.Style.Dash,
			})
		}

		for _, op := range path.Path {
			emit(op, path.Style.Transform)
		}

		switch {
		case willFill && willStroke && fas != nil:
			fas.FillAndStroke(
				path.Style.FillColor, path.Style.FillOpacity*opacity,
				path.Style.LineColor, path.Style.LineOpacity*opacity,
			)
		default:
			if willFill {
				d.Fill(path.Style.FillColor, path.Style.FillOpacity*opacity)
			}
			if willStroke {
				d.Stroke(path.Style.LineColor, path.Style.LineOpacity*opacity)
			}
		}

		path.Style.Transform = m
	}
}
