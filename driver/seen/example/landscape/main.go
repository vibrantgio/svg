// Command landscape shows two SVG landscape icons converted to extruded 3D
// medallions with the svg/driver/seen backend, mounted back to back with a
// gap like a two-faced medal. The camera orbits the pair, so every 180
// degrees the other icon faces you. Drag to rotate, scroll to zoom.
package main

import (
	"bytes"
	_ "embed"
	"math"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/vibrantgio/seen"
	"github.com/vibrantgio/seen/context/gio"
	"github.com/vibrantgio/seen/drag"
	"github.com/vibrantgio/seen/layer/bsort"
	"github.com/vibrantgio/seen/matrix"
	"github.com/vibrantgio/seen/quaternion"
	"github.com/vibrantgio/seen/zoom"
	"github.com/vibrantgio/svg/driver"
	svgseen "github.com/vibrantgio/svg/driver/seen"
	"github.com/vibrantgio/svg/parser"
)

//go:embed island.svg
var islandSVG []byte

//go:embed cape.svg
var capeSVG []byte

func main() {
	go landscape()
	app.Main()
}

func landscape() {
	width, height := unit.Dp(800), unit.Dp(800)
	window := new(app.Window)
	window.Option(app.Title("Seen - spinning SVG landscapes"), app.Size(width, height))

	size := float64(height)

	objects := []seen.Object{
		object(islandSVG, "island", svgseen.Depth(0.1)),
		object(capeSVG, "cape", svgseen.Depth(0.1)),
	}
	// Back to back with a gap, like a two-faced medal: the island faces
	// the camera's home position, the cape faces the other way, so the
	// orbit shows the other icon every 180 degrees. Both poses are set
	// once, before the first render, so the bsort BSP tree never rebuilds.
	gap := size * 0.08
	for _, o := range objects {
		o.SetScale(size*0.3, size*0.3, size*0.3)
	}
	objects[0].SetTranslation(0, 0, gap)
	objects[1].SetTranslation(0, 0, -gap)
	objects[1].SetRotation(quaternion.RotY(math.Pi))

	// Backface culling stays on: extruded icons are closed solids with
	// outward normals, and culling is what keeps the coincident walls of
	// abutting SVG shapes from fighting (only the viewer-facing one of
	// each pair renders). Flat faces carry their own ShowBackfaces flag.
	scene := seen.NewDefaultScene()
	for _, o := range objects {
		scene.Group.Add(o)
	}

	// The scene is static and the CAMERA orbits: that is what makes the
	// bsort layer cheap and exact. Its splitting BSP delivers true
	// painter's order from any eye — a depth sort cannot order a big
	// background cap against small raised plates once the icon tilts —
	// and the world-space tree is built once, because spinning happens in
	// the camera, never in the model matrices.
	layer := bsort.NewLayerForScene(scene)
	context := gio.NewContext(window, layer)

	rotate := func(dy, dx float64) {
		camera := &scene.Camera
		camera.SetRotation(quaternion.RotY(dy).RotX(dx).Mul(camera.Rotation()))
	}

	context.Animate().OnBefore(func(t, dt time.Duration) {
		rotate(float64(dt.Milliseconds())*7e-4, 0)
	}).Start()

	context.Drag(drag.Inertia(true)).On(func(e drag.Event) {
		rotate(e.Dx/150, e.Dy/150)
		context.Render()
	})

	// Zoom in the camera, not the objects: scaling objects about their own
	// origins leaves the back-to-back gap fixed in world units, so the
	// medallions drift apart when zooming out — and every object change
	// would rebuild the bsort tree. Magnifying the camera's view
	// normalization keeps the world static.
	zoomFactor := 1.0
	context.Zoom().On(func(e zoom.Event) {
		zoomFactor *= e.Zoom
	})

	widget := gio.Widget(context, func(w, h unit.Dp) {
		scene.FitCenter(0, 0, float64(w), float64(h))
		scene.Camera.Norm = matrix.Scale(zoomFactor, zoomFactor, 1).Mul(scene.Camera.Norm)
	})

	ops := new(op.Ops)
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			os.Exit(0)
		case app.FrameEvent:
			gtx := app.NewContext(ops, e)
			widget(gtx)
			e.Frame(ops)
		}
	}
}

// object converts SVG markup to a centered seen object of unit-ish size.
func object(markup []byte, kind string, options ...svgseen.Option) seen.Object {
	icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseStream(bytes.NewReader(markup))
	if err != nil {
		panic(err)
	}
	drawer := svgseen.NewDrawer(options...)
	driver.Draw(drawer, icon, 1.0)
	return drawer.Object(kind)
}
