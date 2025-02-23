package svg

import (
	"fmt"

	"golang.org/x/image/math/fixed"
)

// Operation groups the different SVG commands
type Operation interface {
	// SVG text representation of the command
	fmt.Stringer
}

// OpMoveTo moves the current point.
type OpMoveTo fixed.Point26_6

func (op OpMoveTo) String() string {
	return fmt.Sprintf("M%4.3f,%4.3f", float32(op.X)/64, float32(op.Y)/64)
}

// OpLineTo draws a line from the current point,
// and updates it.
type OpLineTo fixed.Point26_6

func (op OpLineTo) String() string {
	return fmt.Sprintf("L%4.3f,%4.3f", float32(op.X)/64, float32(op.Y)/64)
}

// OpQuadTo draws a quadratic Bezier curve from the current point,
// and updates it.
type OpQuadTo [2]fixed.Point26_6

func (op OpQuadTo) String() string {
	return fmt.Sprintf("Q%4.3f,%4.3f,%4.3f,%4.3f", float32(op[0].X)/64, float32(op[0].Y)/64,
		float32(op[1].X)/64, float32(op[1].Y)/64)
}

// OpCubicTo draws a cubic Bezier curve from the current point,
// and updates it.
type OpCubicTo [3]fixed.Point26_6

func (op OpCubicTo) String() string {
	return "C" + fmt.Sprintf("C%4.3f,%4.3f,%4.3f,%4.3f,%4.3f,%4.3f", float32(op[0].X)/64, float32(op[0].Y)/64,
		float32(op[1].X)/64, float32(op[1].Y)/64, float32(op[2].X)/64, float32(op[2].Y)/64)
}

// OpClose close the current path.
type OpClose struct{}

func (op OpClose) String() string {
	return "Z"
}
