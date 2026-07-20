package seen

import (
	"math"
	"slices"

	"github.com/vibrantgio/seen/point"
)

// simplify reduces a closed ring with Douglas-Peucker, dropping vertices
// that deviate less than eps from the chords around them. Flattening emits
// densely spaced, nearly collinear points on shallow curves; every vertex
// dropped here also saves a wall face when extruding.
func simplify(ring []point.Point, eps float64) []point.Point {
	if len(ring) <= 4 || eps <= 0 {
		return ring
	}
	// deduplicate consecutive equal points
	dst := ring[:1]
	for _, p := range ring[1:] {
		if !nearlyEqual(p, dst[len(dst)-1]) {
			dst = append(dst, p)
		}
	}
	ring = dst
	if len(ring) <= 4 {
		return ring
	}
	// treat the ring as an open polyline that returns to its start;
	// both endpoints are kept, so the seam vertex always survives
	closed := append(slices.Clone(ring), ring[0])
	keep := make([]bool, len(closed))
	keep[0], keep[len(closed)-1] = true, true
	douglasPeucker(closed, 0, len(closed)-1, eps, keep)
	out := make([]point.Point, 0, len(ring))
	for i, p := range closed[:len(closed)-1] {
		if keep[i] {
			out = append(out, p)
		}
	}
	if len(out) < 3 {
		return ring
	}
	return out
}

// distToSegment returns the distance from p to the segment (a, b).
func distToSegment(p, a, b point.Point) float64 {
	dx, dy := b.X-a.X, b.Y-a.Y
	l2 := dx*dx + dy*dy
	u := 0.0
	if l2 > 0 {
		u = max(0, min(1, ((p.X-a.X)*dx+(p.Y-a.Y)*dy)/l2))
	}
	return math.Hypot(p.X-(a.X+u*dx), p.Y-(a.Y+u*dy))
}

func douglasPeucker(pts []point.Point, lo, hi int, eps float64, keep []bool) {
	if hi-lo < 2 {
		return
	}
	split, max := -1, eps
	for i := lo + 1; i < hi; i++ {
		if d := distToSegment(pts[i], pts[lo], pts[hi]); d >= max {
			split, max = i, d
		}
	}
	if split < 0 {
		return
	}
	keep[split] = true
	douglasPeucker(pts, lo, split, eps, keep)
	douglasPeucker(pts, split, hi, eps, keep)
}
