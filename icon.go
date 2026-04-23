package svg

import (
	"github.com/vibrantgio/svg/matrix"
)

type Icon struct {
	ViewBox   ViewBox
	Paths     []StyledPath
	Transform matrix.Matrix2D
}

// SetTarget sets the Transform matrix to draw within the bounds of the rectangle arguments
func (i *Icon) SetTarget(x, y, w, h float64) {
	scaleW := w / i.ViewBox.W
	scaleH := h / i.ViewBox.H
	i.Transform = matrix.Identity.Translate(x-i.ViewBox.X*scaleW, y-i.ViewBox.Y*scaleH).Scale(scaleW, scaleH)
}
