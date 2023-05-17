package svg

// Bounds defines a bounding box, such as a viewport or a path extent.
type Bounds struct{ X, Y, W, H float64 }

// AspectMeet positions and sizes the ViewBox inside the rectangle defined by
// the arguments, while maintaining the aspect ratio. The arguments dx,dy are
// the width and height of a rectangle (with X and Y both 0.0) and the alignment
// of the rectangle. Where {ax:0.0,ay:0.0} is left/top and {ax:1.0,ay:1.0}
// is right/bottom.
func (viewBox Bounds) AspectMeet(dx, dy, ax, ay float64) (X, Y, W, H float64) {
	aspect := viewBox.W / viewBox.H
	if dx/dy < aspect {
		h := dx / aspect
		return 0, (dy - h) * ay, dx, h
	} else {
		w := dy * aspect
		return (dx - w) * ax, 0, w, dy
	}
}
