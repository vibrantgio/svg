package seen

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	core "github.com/vibrantgio/seen"
	contextsvg "github.com/vibrantgio/seen/context/svg"
	"github.com/vibrantgio/seen/layer/bsort"
	"github.com/vibrantgio/seen/point"
	"github.com/vibrantgio/seen/quaternion"
	"github.com/vibrantgio/svg/driver"
	"github.com/vibrantgio/svg/parser"
)

func square(cx, cy, r float64) []point.Point {
	return []point.Point{
		{X: cx - r, Y: cy - r}, {X: cx + r, Y: cy - r},
		{X: cx + r, Y: cy + r}, {X: cx - r, Y: cy + r},
	}
}

func TestFlattenCubicStaysWithinTolerance(t *testing.T) {
	a := point.Point{X: 0, Y: 0}
	b := point.Point{X: 40, Y: 100}
	c := point.Point{X: 60, Y: -100}
	e := point.Point{X: 100, Y: 0}
	const tol = 0.5
	pts := flattenCubic([]point.Point{a}, a, b, c, e, tol)
	if len(pts) < 8 {
		t.Fatalf("expected a wiggly cubic to need many segments, got %d points", len(pts))
	}
	if last := pts[len(pts)-1]; last != e {
		t.Errorf("end point not exact: %v", last)
	}
	// every curve sample must lie within tol of the polyline
	for i := 0; i <= 200; i++ {
		u := float64(i) / 200
		p := cubicAt(a, b, c, e, u)
		best := math.Inf(1)
		for j := 0; j+1 < len(pts); j++ {
			best = math.Min(best, distToSegment(p, pts[j], pts[j+1]))
		}
		if best > tol*1.01 {
			t.Fatalf("curve point at t=%.3f deviates %.3f > tol %.2f", u, best, tol)
		}
	}
}

func cubicAt(a, b, c, e point.Point, u float64) point.Point {
	v := 1 - u
	return point.Point{
		X: v*v*v*a.X + 3*v*v*u*b.X + 3*v*u*u*c.X + u*u*u*e.X,
		Y: v*v*v*a.Y + 3*v*v*u*b.Y + 3*v*u*u*c.Y + u*u*u*e.Y,
	}
}

func TestClassifyDonutEvenOdd(t *testing.T) {
	outer := square(0, 0, 10)
	hole := square(0, 0, 4)
	regions := classify([][]point.Point{outer, hole}, false)
	if len(regions) != 1 {
		t.Fatalf("expected 1 region, got %d", len(regions))
	}
	if len(regions[0].holes) != 1 {
		t.Fatalf("expected 1 hole, got %d", len(regions[0].holes))
	}
	if shoelace(regions[0].outer) <= 0 {
		t.Error("outer ring should be counter-clockwise")
	}
	if shoelace(regions[0].holes[0]) >= 0 {
		t.Error("hole ring should be clockwise")
	}
}

func TestClassifyDonutNonZero(t *testing.T) {
	outer := square(0, 0, 10)         // CCW (positive area)
	hole := reversed(square(0, 0, 4)) // CW: opposite winding cuts a hole
	solid := square(0, 0, 2)          // CCW again: fills inside the hole
	regions := classify([][]point.Point{outer, hole, solid}, true)
	if len(regions) != 2 {
		t.Fatalf("expected 2 regions (ring and center), got %d", len(regions))
	}
	var holes int
	for _, r := range regions {
		holes += len(r.holes)
	}
	if holes != 1 {
		t.Errorf("expected 1 hole in total, got %d", holes)
	}
}

func TestKeyholePunchesHole(t *testing.T) {
	outer := square(0, 0, 10)
	hole := reversed(square(0, 0, 4)) // CW, as classify produces
	merged := keyhole(outer, [][]point.Point{hole})
	if want := len(outer) + len(hole) + 2; len(merged) != want {
		t.Fatalf("merged ring has %d points, want %d", len(merged), want)
	}
	if insideEvenOdd(point.Point{X: 0, Y: 0}, merged) {
		t.Error("center of the hole should be outside the keyholed ring")
	}
	if !insideEvenOdd(point.Point{X: 7, Y: 0}, merged) {
		t.Error("point in the band should be inside the keyholed ring")
	}
}

func TestLandscapeIconsProduceCenteredGeometry(t *testing.T) {
	for _, name := range []string{"cape.svg", "island.svg"} {
		t.Run(name, func(t *testing.T) {
			icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseFile(
				filepath.Join("..", "testdata", "landscapeIcons", name))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			d := NewDrawer()
			driver.Draw(d, icon, 1.0)
			object := d.Object(name)
			faces := object.Faces()
			if len(faces) == 0 {
				t.Fatal("no faces produced")
			}
			minX, minY := math.Inf(1), math.Inf(1)
			maxX, maxY := math.Inf(-1), math.Inf(-1)
			for _, f := range faces {
				if len(f.Points) < 3 {
					t.Fatalf("degenerate face with %d points", len(f.Points))
				}
				if f.FillMaterial == nil {
					t.Fatal("face without fill material")
				}
				for _, p := range f.Points {
					minX, maxX = math.Min(minX, p.X), math.Max(maxX, p.X)
					minY, maxY = math.Min(minY, p.Y), math.Max(maxY, p.Y)
				}
			}
			if s := math.Max(maxX-minX, maxY-minY); math.Abs(s-2) > 1e-9 {
				t.Errorf("longest side is %f, want 2", s)
			}
			if cx, cy := (minX+maxX)/2, (minY+maxY)/2; math.Abs(cx) > 1e-9 || math.Abs(cy) > 1e-9 {
				t.Errorf("not centered: (%f, %f)", cx, cy)
			}
			t.Logf("%s: %d faces", name, len(faces))
		})
	}
}

// TestRenderBackToBackOrbit renders the example's two-faced-medal
// arrangement — both icons extruded, mounted back to back with a gap —
// from four camera orbit angles. Set SVGSEEN_OUT to keep the files.
func TestRenderBackToBackOrbit(t *testing.T) {
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
	scene.FitCenter(0, 0, 512, 512)
	doc, err := contextsvg.NewSVG("out", 512, 512)
	if err != nil {
		t.Fatal(err)
	}
	context := contextsvg.NewContext(doc.GetElementById("out"),
		bsort.NewLayerForScene(scene))
	for _, angle := range []float64{0, 0.9, math.Pi, math.Pi + 0.9} {
		scene.Camera.SetRotation(quaternion.RotY(angle).RotX(-0.2))
		context.Render()
		name := filepath.Join(outDir, fmt.Sprintf("combo-%03.0f.svg", angle*180/math.Pi))
		if err := doc.SaveToFile(name); err != nil {
			t.Fatal(err)
		}
		t.Logf("rendered %s", name)
	}
}

// TestRenderLandscapeIconHeadless renders the converted cape icon through
// seen's SVG backend and checks real polygons come out the other end.
// Set SVGSEEN_OUT to a directory to keep the rendered files for inspection.
func TestRenderLandscapeIconHeadless(t *testing.T) {
	outDir := os.Getenv("SVGSEEN_OUT")
	if outDir == "" {
		outDir = t.TempDir()
	}
	for _, tc := range []struct {
		file   string
		depth  float64
		suffix string
		spin   float64
	}{
		{"cape.svg", 0, "-flat", 0.6},
		{"island.svg", 0, "-flat", 0.6},
		{"cape.svg", 0.06, "-extruded", 0.6},
		{"island.svg", 0.06, "-extruded", 0.6},
		{"island.svg", 0.06, "-back", 0.6 + math.Pi},
	} {
		icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseFile(
			filepath.Join("..", "testdata", "landscapeIcons", tc.file))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		options := []Option{}
		if tc.depth > 0 {
			options = append(options, Depth(tc.depth))
		}
		d := NewDrawer(options...)
		driver.Draw(d, icon, 1.0)
		object := d.Object(tc.file)
		object.SetRotation(quaternion.RotY(tc.spin).RotX(-0.35))
		object.SetScale(150, 150, 150)

		const width, height = 512, 512
		doc, err := contextsvg.NewSVG("out", width, height)
		if err != nil {
			t.Fatal(err)
		}
		// culling stays on: if winding or normals were wrong, faces would
		// vanish and the size assertion below would catch it
		scene := core.NewDefaultScene()
		scene.Group.Add(object)
		scene.FitCenter(0, 0, width, height)
		context := contextsvg.NewContext(doc.GetElementById("out"),
			bsort.NewLayerForScene(scene))
		context.Render()

		name := filepath.Join(outDir, tc.file[:len(tc.file)-4]+tc.suffix+".svg")
		if err := doc.SaveToFile(name); err != nil {
			t.Fatal(err)
		}
		markup, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if len(markup) < 2000 {
			t.Errorf("%s: suspiciously small render (%d bytes)", name, len(markup))
		}
		t.Logf("rendered %s (%d bytes)", name, len(markup))
	}
}
