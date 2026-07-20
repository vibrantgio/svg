package seen

import (
	"math"
	"path/filepath"
	"testing"

	core "github.com/vibrantgio/seen"
	contextsvg "github.com/vibrantgio/seen/context/svg"
	"github.com/vibrantgio/seen/layer"
	"github.com/vibrantgio/seen/layer/bsort"
	"github.com/vibrantgio/seen/layer/nsort"
	"github.com/vibrantgio/seen/layer/zsort"
	"github.com/vibrantgio/seen/quaternion"
	"github.com/vibrantgio/svg/driver"
	"github.com/vibrantgio/svg/parser"
)

// benchRender spins the converted icon through a full render each iteration,
// the same work a Gio frame does during animation.
func benchRender(b *testing.B, file string, depth float64, newLayer func(*core.Scene) layer.Layer) {
	icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseFile(
		filepath.Join("..", "testdata", "landscapeIcons", file))
	if err != nil {
		b.Fatal(err)
	}
	options := []Option{}
	if depth > 0 {
		options = append(options, Depth(depth))
	}
	d := NewDrawer(options...)
	driver.Draw(d, icon, 1.0)
	object := d.Object(file)
	object.SetScale(150, 150, 150)
	b.Logf("%s depth=%.2f: %d faces", file, depth, len(object.Faces()))

	scene := core.NewDefaultScene()
	scene.ShowBackfaces = true
	scene.Group.Add(object)
	scene.FitCenter(0, 0, 512, 512)
	doc, err := contextsvg.NewSVG("out", 512, 512)
	if err != nil {
		b.Fatal(err)
	}
	context := contextsvg.NewContext(doc.GetElementById("out"), newLayer(scene))

	for i := 0; b.Loop(); i++ {
		object.SetRotation(quaternion.RotY(float64(i) * 0.03))
		context.Render()
	}
}

func BenchmarkIslandExtrudedNsort(b *testing.B) {
	benchRender(b, "island.svg", 0.1, nsort.NewLayerForScene)
}

// BenchmarkBackToBackBsortOrbit measures the example's real workload: both
// extruded medallions back to back, static world, orbiting camera.
func BenchmarkBackToBackBsortOrbit(b *testing.B) {
	scene := core.NewDefaultScene()
	total := 0
	for i, file := range []string{"island.svg", "cape.svg"} {
		icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseFile(
			filepath.Join("..", "testdata", "landscapeIcons", file))
		if err != nil {
			b.Fatal(err)
		}
		d := NewDrawer(Depth(0.1))
		driver.Draw(d, icon, 1.0)
		object := d.Object(file)
		object.SetScale(150, 150, 150)
		if i == 0 {
			object.SetTranslation(0, 0, 40)
		} else {
			object.SetTranslation(0, 0, -40)
			object.SetRotation(quaternion.RotY(math.Pi))
		}
		total += len(object.Faces())
		scene.Group.Add(object)
	}
	b.Logf("%d faces", total)
	scene.FitCenter(0, 0, 512, 512)
	doc, err := contextsvg.NewSVG("out", 512, 512)
	if err != nil {
		b.Fatal(err)
	}
	context := contextsvg.NewContext(doc.GetElementById("out"),
		bsort.NewLayerForScene(scene))
	for i := 0; b.Loop(); i++ {
		scene.Camera.SetRotation(quaternion.RotY(float64(i) * 0.03))
		context.Render()
	}
}

// BenchmarkIslandExtrudedBsortOrbit is the example's configuration: static
// world, orbiting camera, so the BSP tree is built once and reused.
func BenchmarkIslandExtrudedBsortOrbit(b *testing.B) {
	icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseFile(
		filepath.Join("..", "testdata", "landscapeIcons", "island.svg"))
	if err != nil {
		b.Fatal(err)
	}
	d := NewDrawer(Depth(0.1))
	driver.Draw(d, icon, 1.0)
	object := d.Object("island.svg")
	object.SetScale(150, 150, 150)

	scene := core.NewDefaultScene()
	scene.Group.Add(object)
	scene.FitCenter(0, 0, 512, 512)
	doc, err := contextsvg.NewSVG("out", 512, 512)
	if err != nil {
		b.Fatal(err)
	}
	context := contextsvg.NewContext(doc.GetElementById("out"),
		bsort.NewLayerForScene(scene))

	for i := 0; b.Loop(); i++ {
		scene.Camera.SetRotation(quaternion.RotY(float64(i) * 0.03))
		context.Render()
	}
}

func BenchmarkIslandExtrudedZsort(b *testing.B) {
	benchRender(b, "island.svg", 0.1, zsort.NewLayerForScene)
}

func BenchmarkIslandFlatNsort(b *testing.B) {
	benchRender(b, "island.svg", 0, nsort.NewLayerForScene)
}

func BenchmarkIslandFlatZsort(b *testing.B) {
	benchRender(b, "island.svg", 0, zsort.NewLayerForScene)
}

func BenchmarkCapeExtrudedZsort(b *testing.B) {
	benchRender(b, "cape.svg", 0.1, zsort.NewLayerForScene)
}
