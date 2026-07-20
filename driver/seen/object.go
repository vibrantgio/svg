package seen

import (
	"math"

	core "github.com/vibrantgio/seen"
	"github.com/vibrantgio/seen/point"
	"github.com/vibrantgio/seen/shape"
)

// Object finalizes everything drawn so far and returns it as a single
// seen.Object: centered at the origin, y up, scaled so the longest x/y side
// measures Size, with paint-order layers spaced by LayerOffset (or extruded
// by Depth). Call it once, after driver.Draw.
func (d *Drawer) Object(kind string) core.Object {
	minX, minY, minZ := math.Inf(1), math.Inf(1), math.Inf(1)
	maxX, maxY, maxZ := math.Inf(-1), math.Inf(-1), math.Inf(-1)
	zstep := d.layerOffset
	if d.depth > 0 {
		zstep = d.depth
	}
	for i := range d.faces {
		for _, p := range d.faces[i].Points {
			minX, maxX = math.Min(minX, p.X), math.Max(maxX, p.X)
			minY, maxY = math.Min(minY, p.Y), math.Max(maxY, p.Y)
			z := p.Z * zstep
			minZ, maxZ = math.Min(minZ, z), math.Max(maxZ, z)
		}
	}
	if len(d.faces) == 0 || maxX-minX <= 0 && maxY-minY <= 0 {
		return shape.NewShapeWithFaces(kind, d.faces)
	}
	s := d.size / math.Max(maxX-minX, maxY-minY)
	cx, cy, cz := (minX+maxX)/2, (minY+maxY)/2, (minZ+maxZ)/2
	for i := range d.faces {
		pts := d.faces[i].Points
		for j, p := range pts {
			pts[j] = point.Point{X: (p.X - cx) * s, Y: (p.Y - cy) * s, Z: p.Z*zstep - cz}
		}
	}
	return shape.NewShapeWithFaces(kind, d.faces)
}
