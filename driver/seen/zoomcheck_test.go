package seen

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	core "github.com/vibrantgio/seen"
	contextsvg "github.com/vibrantgio/seen/context/svg"
	"github.com/vibrantgio/seen/layer/bsort"
	"github.com/vibrantgio/seen/matrix"
	"github.com/vibrantgio/seen/quaternion"
	"github.com/vibrantgio/svg/driver"
	"github.com/vibrantgio/svg/parser"
)

// TestRenderZoomedOrbit checks the example's camera zoom: magnifying the
// camera's Norm must shrink the whole tableau uniformly, gap included.
func TestRenderZoomedOrbit(t *testing.T) {
	outDir := os.Getenv("SVGSEEN_OUT")
	if outDir == "" {
		outDir = t.TempDir()
	}
	scene := core.NewDefaultScene()
	for i, file := range []string{"island.svg", "cape.svg"} {
		icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseFile(
			filepath.Join("..", "testdata", "landscapeIcons", file))
		if err != nil {
			t.Fatalf("parse: %v", err)
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
		scene.Group.Add(object)
	}
	doc, err := contextsvg.NewSVG("out", 512, 512)
	if err != nil {
		t.Fatal(err)
	}
	context := contextsvg.NewContext(doc.GetElementById("out"),
		bsort.NewLayerForScene(scene))
	scene.Camera.SetRotation(quaternion.RotY(0.9).RotX(-0.2))
	for _, zoom := range []float64{1.0, 0.5} {
		scene.FitCenter(0, 0, 512, 512)
		scene.Camera.Norm = matrix.Scale(zoom, zoom, 1).Mul(scene.Camera.Norm)
		context.Render()
		name := filepath.Join(outDir, "zoom-"+map[float64]string{1: "100", 0.5: "050"}[zoom]+".svg")
		if err := doc.SaveToFile(name); err != nil {
			t.Fatal(err)
		}
		t.Logf("rendered %s", name)
	}
}
