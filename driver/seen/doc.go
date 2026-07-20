// Package seen implements a rendering backend for github.com/vibrantgio/svg
// that produces 3D geometry for github.com/vibrantgio/seen instead of pixels.
//
// The Drawer implements driver.DrawerNG. driver.Draw walks a parsed svg.Icon
// and feeds transformed path segments to the drawer, which flattens curves,
// resolves fill rules and holes, and accumulates one seen face per filled
// region. Object returns the result as a seen.Object centered at the origin,
// ready to be added to a scene and rotated like any other shape.
//
// Successive SVG paths are offset slightly along the z axis so the painter's
// algorithm reproduces SVG paint order; with Depth set, each path becomes an
// extruded slab instead, stacking into a relief.
//
// Because this package shares its name with the library it targets,
// importers typically alias it:
//
//	import (
//		"github.com/vibrantgio/seen"
//		"github.com/vibrantgio/svg/driver"
//		svgseen "github.com/vibrantgio/svg/driver/seen"
//	)
//
//	drawer := svgseen.NewDrawer(svgseen.Depth(0.06))
//	driver.Draw(drawer, icon, 1.0)
//	object := drawer.Object("logo")
//
// Gradients are approximated by the average color of their stops. Strokes on
// filled paths become the faces' stroke material; stroke-only paths are
// emitted with a transparent fill.
package seen
