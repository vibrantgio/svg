package svg

import (
	"github.com/reactivego/svg/matrix"
	"golang.org/x/image/math/fixed"
)

// DefaultStyle sets the default PathStyle to fill black, winding rule,
// full opacity, no stroke, ButtCap line end and Bevel line connect.
var DefaultStyle = PathStyle{
	FillOpacity:       1.0,
	LineOpacity:       1.0,
	LineWidth:         2.0,
	UseNonZeroWinding: true,
	Join: JoinOptions{
		MiterLimit:   fixed.Int26_6(4. * 64),
		LineJoin:     Bevel,
		TrailLineCap: ButtCap,
	},
	FillColor: PlainColor{A: 255},
	Transform: matrix.Identity,
}
