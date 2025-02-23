package driver

import (
	"github.com/reactivego/svg"
	"github.com/reactivego/svg/matrix"
)

type Icon struct {
	ViewBox   svg.ViewBox
	Paths     []svg.StyledPath
	Transform matrix.Matrix2D
}

func NewIcon(viewbox svg.ViewBox, paths []svg.StyledPath) *Icon {
	return &Icon{ViewBox: viewbox, Paths: paths, Transform: matrix.Identity}
}

// SetTarget sets the Transform matrix to draw within the bounds of the rectangle arguments
func (i *Icon) SetTarget(x, y, w, h float64) {
	scaleW := w / i.ViewBox.W
	scaleH := h / i.ViewBox.H
	i.Transform = matrix.Identity.Translate(x-i.ViewBox.X*scaleW, y-i.ViewBox.Y*scaleH).Scale(scaleW, scaleH)
}

// Draw the compiled SVG icon into the driver `d`.
// `opacity` is composed (multiplied) with the eventual
// <stroke-opacity> and <fill-opacity> style attributes.
// All elements should be contained by the Bounds rectangle of the SvgIcon:
// see `SetTarget` method.
func (i *Icon) Draw(d Driver, opacity float64) {
	DrawTransformed(d, i.Paths, i.Transform, opacity)
}
