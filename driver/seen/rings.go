package seen

import (
	"math"
	"slices"

	"github.com/vibrantgio/seen/point"
)

// region is a filled outline together with the holes cut out of it.
// The outer ring is counter-clockwise, holes are clockwise.
type region struct {
	outer []point.Point
	holes [][]point.Point
	level int // containment depth, used to stack nested fills
}

// classify splits the closed sub-paths of one SVG path into filled regions
// and their holes according to the path's fill rule.
func classify(rings [][]point.Point, nonzero bool) []region {
	kept := rings[:0:0]
	for _, r := range rings {
		if len(r) >= 3 {
			kept = append(kept, r)
		}
	}
	rings = kept
	if len(rings) == 0 {
		return nil
	}

	n := len(rings)
	area := make([]float64, n)
	for i, r := range rings {
		area[i] = shoelace(r)
	}

	// containers[i] lists the rings whose interior contains ring i,
	// by even-odd test of a representative vertex.
	containers := make([][]int, n)
	for i := range rings {
		for j := range rings {
			if i != j && insideEvenOdd(rings[i][0], rings[j]) {
				containers[i] = append(containers[i], j)
			}
		}
	}

	// Decide per ring whether its inside is filled and its outside is not
	// (an outer boundary), or the reverse (a hole boundary).
	outer := make([]bool, n)
	hole := make([]bool, n)
	for i := range rings {
		var in, out bool
		if nonzero {
			w := sign(area[i])
			for _, j := range containers[i] {
				w += sign(area[j])
			}
			in, out = w != 0, w-sign(area[i]) != 0
		} else {
			depth := len(containers[i])
			in, out = depth%2 == 0, depth%2 == 1
		}
		outer[i], hole[i] = in && !out, out && !in
	}

	// Attach each hole to its innermost containing outer ring.
	index := make(map[int]int) // ring index -> regions index
	var regions []region
	for i := range rings {
		if !outer[i] {
			continue
		}
		o := rings[i]
		if area[i] < 0 {
			o = reversed(o)
		}
		index[i] = len(regions)
		regions = append(regions, region{outer: o, level: len(containers[i])})
	}
	for i := range rings {
		if !hole[i] {
			continue
		}
		parent := -1
		for _, j := range containers[i] {
			if !outer[j] {
				continue
			}
			if parent == -1 || math.Abs(area[j]) < math.Abs(area[parent]) {
				parent = j
			}
		}
		if parent == -1 {
			continue // malformed: hole without a filled parent
		}
		h := rings[i]
		if area[i] > 0 {
			h = reversed(h)
		}
		r := &regions[index[parent]]
		r.holes = append(r.holes, h)
	}
	return regions
}

// keyhole merges the holes into the outer ring with zero-width bridges,
// producing a single contour that fills correctly under any fill rule.
// Bridges connect nearest vertex pairs; with well-separated sibling holes
// (the usual case in icon art) bridges do not cross other holes.
func keyhole(outer []point.Point, holes [][]point.Point) []point.Point {
	merged := outer
	remaining := make([][]point.Point, len(holes))
	copy(remaining, holes)
	for len(remaining) > 0 {
		bestH, bestHi, bestOi, bestD := -1, 0, 0, math.MaxFloat64
		for h, ring := range remaining {
			for hi, hp := range ring {
				for oi, op := range merged {
					dx, dy := hp.X-op.X, hp.Y-op.Y
					if d := dx*dx + dy*dy; d < bestD {
						bestH, bestHi, bestOi, bestD = h, hi, oi, d
					}
				}
			}
		}
		merged = splice(merged, bestOi, remaining[bestH], bestHi)
		remaining = slices.Delete(remaining, bestH, bestH+1)
	}
	return merged
}

// splice inserts the full hole cycle into the outer ring at the bridge
// vertices, walking out to the hole, around it, and back.
func splice(outer []point.Point, oi int, hole []point.Point, hi int) []point.Point {
	out := make([]point.Point, 0, len(outer)+len(hole)+2)
	out = append(out, outer[:oi+1]...)
	out = append(out, hole[hi:]...)
	out = append(out, hole[:hi+1]...)
	out = append(out, outer[oi:]...)
	return out
}

// shoelace returns the signed area: positive for counter-clockwise rings
// in seen's y-up coordinates.
func shoelace(ring []point.Point) float64 {
	sum := 0.0
	for i, p := range ring {
		q := ring[(i+1)%len(ring)]
		sum += p.X*q.Y - q.X*p.Y
	}
	return sum / 2
}

// insideEvenOdd reports whether p lies inside ring by even-odd ray casting.
func insideEvenOdd(p point.Point, ring []point.Point) bool {
	in := false
	for i, a := range ring {
		b := ring[(i+1)%len(ring)]
		if (a.Y > p.Y) != (b.Y > p.Y) &&
			p.X < a.X+(p.Y-a.Y)/(b.Y-a.Y)*(b.X-a.X) {
			in = !in
		}
	}
	return in
}

func sign(a float64) int {
	switch {
	case a > 0:
		return 1
	case a < 0:
		return -1
	}
	return 0
}
