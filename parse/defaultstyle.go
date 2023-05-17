package parse

import (
	"github.com/reactivego/svg"
	"github.com/reactivego/svg/matrix"
)

// DefaultStyle sets the default PathStyle to fill black, winding rule,
// full opacity, no stroke, ButtCap line end and Bevel line connect.
var DefaultStyle = PathStyle{
	FillOpacity:       1.0,
	LineOpacity:       1.0,
	LineWidth:         2.0,
	UseNonZeroWinding: true,
	Join: svg.JoinOptions{
		MiterLimit:   fToFixed(4.),
		LineJoin:     svg.Bevel,
		TrailLineCap: svg.ButtCap,
	},
	FillColor: svg.PlainColor{A: 255},
	Transform: matrix.Identity,
}
