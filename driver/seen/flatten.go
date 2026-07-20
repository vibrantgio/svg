package seen

import (
	"math"

	"github.com/vibrantgio/seen/point"
)

const maxFlattenDepth = 16

// flattenQuad appends line segments approximating the quadratic bezier
// (a, b, c) to dst so no point deviates more than tol from the curve.
// The end point c is always appended, exactly.
func flattenQuad(dst []point.Point, a, b, c point.Point, tol float64) []point.Point {
	return flattenQuadRec(dst, a, b, c, tol, 0)
}

func flattenQuadRec(dst []point.Point, a, b, c point.Point, tol float64, depth int) []point.Point {
	// max deviation of a quadratic from its chord is |a - 2b + c| / 4
	if depth >= maxFlattenDepth || math.Hypot(a.X-2*b.X+c.X, a.Y-2*b.Y+c.Y)/4 <= tol {
		return append(dst, c)
	}
	ab, bc := mid(a, b), mid(b, c)
	m := mid(ab, bc)
	dst = flattenQuadRec(dst, a, ab, m, tol, depth+1)
	return flattenQuadRec(dst, m, bc, c, tol, depth+1)
}

// flattenCubic appends line segments approximating the cubic bezier
// (a, b, c, e) to dst so no point deviates more than tol from the curve.
// The end point e is always appended, exactly.
func flattenCubic(dst []point.Point, a, b, c, e point.Point, tol float64) []point.Point {
	return flattenCubicRec(dst, a, b, c, e, tol, 0)
}

func flattenCubicRec(dst []point.Point, a, b, c, e point.Point, tol float64, depth int) []point.Point {
	// B''(t) = 6((1-t)(a-2b+c) + t(b-2c+e)), and the deviation of a curve
	// from its chord is at most max|B''|/8, so 3/4 * max(d1, d2) is safe.
	d1 := math.Hypot(a.X-2*b.X+c.X, a.Y-2*b.Y+c.Y)
	d2 := math.Hypot(b.X-2*c.X+e.X, b.Y-2*c.Y+e.Y)
	if depth >= maxFlattenDepth || 3*math.Max(d1, d2)/4 <= tol {
		return append(dst, e)
	}
	ab, bc, ce := mid(a, b), mid(b, c), mid(c, e)
	abc, bce := mid(ab, bc), mid(bc, ce)
	m := mid(abc, bce)
	dst = flattenCubicRec(dst, a, ab, abc, m, tol, depth+1)
	return flattenCubicRec(dst, m, bce, ce, e, tol, depth+1)
}

func mid(a, b point.Point) point.Point {
	return point.Point{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
}
