package seen

import (
	"math"
	"strconv"

	"golang.org/x/image/math/fixed"

	"github.com/vibrantgio/seen/color"
	"github.com/vibrantgio/seen/face"
	"github.com/vibrantgio/seen/point"
	"github.com/vibrantgio/seen/shader"
	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/driver"
)

// Option configures a Drawer.
type Option func(*Drawer)

// Tolerance sets the maximum distance, in icon (viewBox) units, that a
// flattened bezier segment may deviate from the true curve. Default 1.
func Tolerance(t float64) Option { return func(d *Drawer) { d.tolerance = t } }

// Size sets the length of the longest x/y side of the produced object.
// Default 2, matching the 2x2x2 world of the seen examples.
func Size(s float64) Option { return func(d *Drawer) { d.size = s } }

// Depth extrudes the icon: the first path becomes a base slab of the given
// thickness (in the same units as Size) and every following path a thin
// plate raised above it, a stacked relief. Default 0 (flat faces separated
// by LayerOffset).
func Depth(depth float64) Option { return func(d *Drawer) { d.depth = depth } }

// LayerOffset sets the z spacing between successive paths when Depth is 0,
// in the same units as Size. It keeps SVG paint order stable under the
// painter's algorithm. Default Size * 0.0015.
func LayerOffset(o float64) Option { return func(d *Drawer) { d.layerOffset = o } }

// Flat renders faces in their exact SVG colors, ignoring scene lighting.
// By default faces get the standard material and are shaded by the scene.
func Flat() Option { return func(d *Drawer) { d.flat = true } }

// Drawer accumulates the paths of an svg.Icon as seen faces.
// Drive it with driver.Draw, then collect the geometry with Object.
type Drawer struct {
	tolerance   float64
	size        float64
	depth       float64
	layerOffset float64
	flat        bool

	// state of the styled path currently being drawn
	nonzero        bool
	stroke         driver.StrokeOptions
	open           []point.Point   // sub-path under construction
	rings          [][]point.Point // completed sub-paths
	pos            point.Point     // current point
	lastLo, lastHi int             // faces emitted by Fill for this path

	faces face.Faces
	layer int // paint-order index of the next path
	bias  int // extruded faces displaced so far, see faceBias
}

var _ driver.DrawerNG = (*Drawer)(nil)

func NewDrawer(options ...Option) *Drawer {
	d := &Drawer{tolerance: 1, size: 2, layerOffset: -1}
	for _, o := range options {
		o(d)
	}
	if d.layerOffset < 0 {
		d.layerOffset = d.size * 0.0015
	}
	return d
}

// pt converts an incoming fixed-point position to seen's y-up coordinates.
func (d *Drawer) pt(p fixed.Point26_6) point.Point {
	return point.Point{X: float64(p.X) / 64, Y: -float64(p.Y) / 64}
}

func (d *Drawer) Clear() {
	d.open, d.rings = nil, nil
	d.lastLo, d.lastHi = 0, 0
}

func (d *Drawer) SetWinding(useNonZeroWinding bool) { d.nonzero = useNonZeroWinding }

func (d *Drawer) SetStrokeOptions(options driver.StrokeOptions) { d.stroke = options }

func (d *Drawer) Start(a fixed.Point26_6) {
	d.flush()
	d.pos = d.pt(a)
	d.open = append(d.open, d.pos)
}

func (d *Drawer) Line(b fixed.Point26_6) {
	d.pos = d.pt(b)
	d.open = append(d.open, d.pos)
}

func (d *Drawer) QuadBezier(b, c fixed.Point26_6) {
	pb, pc := d.pt(b), d.pt(c)
	d.open = flattenQuad(d.open, d.pos, pb, pc, d.tolerance)
	d.pos = pc
}

func (d *Drawer) CubeBezier(b, c, e fixed.Point26_6) {
	pb, pc, pe := d.pt(b), d.pt(c), d.pt(e)
	d.open = flattenCubic(d.open, d.pos, pb, pc, pe, d.tolerance)
	d.pos = pe
}

func (d *Drawer) Close() { d.flush() }

// flush completes the sub-path under construction. Fill and Stroke also
// flush, closing any sub-path that had no explicit close command.
func (d *Drawer) flush() {
	if len(d.open) >= 3 {
		if first, last := d.open[0], d.open[len(d.open)-1]; nearlyEqual(first, last) {
			d.open = d.open[:len(d.open)-1]
		}
	}
	if len(d.open) >= 3 {
		d.rings = append(d.rings, simplify(d.open, d.tolerance))
	}
	d.open = nil
}

func (d *Drawer) Fill(pattern svg.Pattern, opacity float64) {
	d.flush()
	lo := len(d.faces)
	d.emit(materialWith(pattern, opacity, d.flat), nil, "")
	d.lastLo, d.lastHi = lo, len(d.faces)
}

func (d *Drawer) Stroke(pattern svg.Pattern, opacity float64) {
	d.flush()
	material := materialWith(pattern, opacity, true)
	width := strconv.FormatFloat(float64(d.stroke.LineWidth)/64, 'g', 4, 64)
	if d.lastHi > d.lastLo {
		// decorate the faces just emitted by Fill for this same path
		for i := d.lastLo; i < d.lastHi; i++ {
			d.faces[i].StrokeMaterial = material
			d.faces[i].Options["stroke-width"] = width
		}
		return
	}
	// stroke-only path: emit its regions with a transparent fill
	transparent := &shader.Material{Color: color.Color{}, Shader: shader.Flat}
	d.emit(transparent, material, width)
}

// emit converts the buffered rings of the current path into faces at the
// current paint-order layer. Layers use unit z spacing here; Object scales
// z to LayerOffset or Depth when the geometry is finalized.
func (d *Drawer) emit(fill, stroke *shader.Material, width string) {
	regions := classify(d.rings, d.nonzero)
	if len(regions) == 0 {
		return
	}
	for _, r := range regions {
		outline := keyhole(r.outer, r.holes)
		// nested same-winding fills overlap their parent: nudge them up
		z := float64(d.layer) + 0.05*float64(r.level)
		if d.depth == 0 {
			f := d.newFace(offsetZ(outline, z), fill, stroke, width)
			f.ShowBackfaces = true
			d.faces = append(d.faces, f)
			continue
		}
		// Extruded, the first path becomes a base slab of unit thickness
		// (scaled to Depth later) and every following path a thin plate
		// raised above it. Plates make the paths' depth ranges disjoint,
		// which is what lets a barycenter depth sort order faces of
		// different paths correctly from any angle: a big background cap
		// has a single unrepresentative barycenter, so overlapping depth
		// ranges would let small raised features ghost through it.
		const plate = 0.08
		var zb, zf float64
		if z == 0 {
			zb, zf = 0, 1 // base slab
		} else {
			zb = 1 + math.Max(0, z-1)*plate
			zf = zb + plate
		}
		d.faces = append(d.faces, d.newFace(offsetZ(outline, zf+d.faceBias()), fill, stroke, width))
		d.faces = append(d.faces, d.newFace(offsetZ(reversed(outline), zb+d.faceBias()), fill, nil, ""))
		walls := append([][]point.Point{r.outer}, r.holes...)
		for _, ring := range walls {
			for i := range ring {
				j := (i + 1) % len(ring)
				bias := d.faceBias()
				quad := point.Points{
					{X: ring[i].X, Y: ring[i].Y, Z: zf + bias},
					{X: ring[i].X, Y: ring[i].Y, Z: zb + bias},
					{X: ring[j].X, Y: ring[j].Y, Z: zb + bias},
					{X: ring[j].X, Y: ring[j].Y, Z: zf + bias},
				}
				d.faces = append(d.faces, d.newFace(quad, fill, nil, ""))
			}
		}
	}
	d.layer++
}

// faceBias returns a strictly increasing hair of z displacement, one step
// per extruded face. Abutting SVG shapes share edges, so different paths
// emit wall quads in exactly the same place; without the displacement their
// sort depths tie and trade places while rotating, which shows as
// z-fighting. The step is far below the 0.02 paint-order lift, so it can
// never flip intended layer order.
func (d *Drawer) faceBias() float64 {
	d.bias++
	return float64(d.bias) * 2e-5
}

func (d *Drawer) newFace(points point.Points, fill, stroke *shader.Material, width string) face.Face {
	f := face.FaceWith(points)
	f.FillMaterial = fill
	if stroke != nil {
		f.StrokeMaterial = stroke
		f.Options["stroke-width"] = width
	}
	return f
}

// materialWith maps an SVG paint pattern to a seen material. Gradients are
// approximated by the average color of their stops.
func materialWith(pattern svg.Pattern, opacity float64, flat bool) *shader.Material {
	c := color.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}
	switch p := pattern.(type) {
	case svg.PlainColor:
		c = color.Color{
			R: float64(p.R) / 255, G: float64(p.G) / 255,
			B: float64(p.B) / 255, A: float64(p.A) / 255,
		}
	case svg.Gradient:
		c = averageStops(p)
	case *svg.Gradient:
		c = averageStops(*p)
	}
	c.A *= opacity
	m, _ := shader.NewMaterialWith(c)
	if flat {
		m.Shader = shader.Flat
	}
	return m
}

func averageStops(g svg.Gradient) color.Color {
	if len(g.Stops) == 0 {
		return color.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}
	}
	var c color.Color
	for _, stop := range g.Stops {
		r, gr, b, a := stop.StopColor.RGBA()
		c.R += float64(r>>8) / 255
		c.G += float64(gr>>8) / 255
		c.B += float64(b>>8) / 255
		c.A += float64(a>>8) / 255 * stop.Opacity
	}
	n := float64(len(g.Stops))
	return color.Color{R: c.R / n, G: c.G / n, B: c.B / n, A: c.A / n}
}

func offsetZ(ring []point.Point, z float64) point.Points {
	out := make(point.Points, len(ring))
	for i, p := range ring {
		out[i] = point.Point{X: p.X, Y: p.Y, Z: z}
	}
	return out
}

func reversed(ring []point.Point) []point.Point {
	out := make([]point.Point, len(ring))
	for i, p := range ring {
		out[len(ring)-1-i] = p
	}
	return out
}

func nearlyEqual(a, b point.Point) bool {
	const eps = 1e-9
	dx, dy := a.X-b.X, a.Y-b.Y
	return dx > -eps && dx < eps && dy > -eps && dy < eps
}
