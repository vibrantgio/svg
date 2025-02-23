package svg

import (
	"github.com/reactivego/svg/matrix"
	"golang.org/x/image/math/fixed"
)

// matrixAdder add points to path after applying a  matrix M to all points
type matrixAdder struct {
	path *Path
	M    matrix.Matrix2D
}

// Reset sets the matrix M to identity
func (t *matrixAdder) Reset() {
	t.M = matrix.Identity
}

// Start starts a new path
func (t *matrixAdder) Start(a fixed.Point26_6) {
	t.path.Start(t.M.TFixed(a))
}

// Line adds a linear segment to the current curve.
func (t *matrixAdder) Line(b fixed.Point26_6) {
	t.path.Line(t.M.TFixed(b))
}

// QuadBezier adds a quadratic segment to the current curve.
func (t *matrixAdder) QuadBezier(b, c fixed.Point26_6) {
	t.path.QuadBezier(t.M.TFixed(b), t.M.TFixed(c))
}

// CubeBezier adds a cubic segment to the current curve.
func (t *matrixAdder) CubeBezier(b, c, d fixed.Point26_6) {
	t.path.CubeBezier(t.M.TFixed(b), t.M.TFixed(c), t.M.TFixed(d))
}
