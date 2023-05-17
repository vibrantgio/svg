package parse

import (
	"github.com/reactivego/svg"
	"github.com/reactivego/svg/matrix"
)

// PathStyle holds the state of the SVG style
type PathStyle struct {
	FillOpacity       float64
	LineOpacity       float64
	LineWidth         float64
	UseNonZeroWinding bool

	Join      svg.JoinOptions
	Dash      svg.DashOptions
	FillColor svg.Pattern // either PlainColor or Gradient
	LineColor svg.Pattern // either PlainColor or Gradient

	Transform matrix.Matrix2D // current transform
}
