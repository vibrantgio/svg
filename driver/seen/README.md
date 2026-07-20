# svg/driver/seen

A rendering backend for `github.com/vibrantgio/svg` that produces 3D
geometry for [seen](https://github.com/vibrantgio/seen) instead of pixels:
parse an SVG, draw it through this driver, and get a `seen.Object` you can
add to a scene, light, spin, and drag like any other shape.

```go
import (
	"github.com/vibrantgio/svg/driver"
	svgseen "github.com/vibrantgio/svg/driver/seen"
	"github.com/vibrantgio/svg/parser"
)

icon, _ := parser.NewParser(parser.IgnoreErrorMode).ParseFile("logo.svg")
drawer := svgseen.NewDrawer(svgseen.Depth(0.1))
driver.Draw(drawer, icon, 1.0)
object := drawer.Object("logo") // a seen.Object, centered, y up, 2 units wide
```

Curves are flattened within `Tolerance`, holes are cut by splicing hole
contours into their outers with zero-width keyhole bridges, and successive
SVG paths are offset a hair along z (`LayerOffset`) so the painter's
algorithm reproduces SVG paint order. With `Depth` the icon becomes a single
extruded slab — a solid, medallion-like object. `Flat` keeps the exact SVG
colors instead of scene lighting. Gradients are approximated by the average
of their stops; strokes on filled paths become the faces' stroke material.

To spin a converted icon, keep the object still and orbit the camera
(`scene.Camera.SetRotation(...)` — camera rotation orbits the world origin)
and render with a `layer/bsort` layer. Its splitting BSP gives exact
painter's order from any angle — a per-face depth sort cannot order a big
background cap against small raised features once the icon tilts — and its
world-space tree is built once, since spinning happens in the camera, never
in the model matrices. Leave backface culling on: extruded icons are closed
solids with outward normals, and culling is what keeps the coincident walls
of abutting SVG shapes from fighting. Contours are Douglas-Peucker
simplified within `Tolerance` after flattening, which also keeps the wall
face count down when extruding.

See `example/landscape` for two extruded landscape icons mounted back to
back like a two-faced medal, with the camera orbiting the pair.
