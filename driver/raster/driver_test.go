package driver

import (
	"image"
	"strings"
	"testing"

	"github.com/srwiley/scanFT"

	"github.com/vibrantgio/svg/driver"
	"github.com/vibrantgio/svg/parser"
)

// TestFillRuleWinding renders a pentagram: a five-pointed star drawn as one
// closed subpath that visits every second vertex, so its five edges
// self-intersect. The centre pentagon has winding number two, so
// it is filled under the non-zero rule and hollow under even-odd. The parser
// used to invert fill-rule — evenodd set non-zero winding and vice versa —
// and this is the regression test for that: an explicit nonzero and an
// unstated rule must fill the centre, an explicit evenodd must not.
//
// The default rasterx.ScannerGV ignores SetWinding, so the star is rastered
// through scanFT.ScannerFT, which honours it.
func TestFillRuleWinding(t *testing.T) {
	for _, tc := range []struct {
		name         string
		attr         string // fill-rule attribute, or empty for unstated
		centreFilled bool
	}{
		{"nonzero", ` fill-rule="nonzero"`, true},
		{"evenodd", ` fill-rule="evenodd"`, false},
		{"unstated", ``, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			img := renderStar(t, tc.attr)
			// A point on an arm has winding number one and is filled
			// under either rule; if it is empty nothing rendered and
			// the centre assertion would be meaningless.
			if !opaqueAt(img, 50, 10) {
				t.Fatal("arm pixel (50,10) empty: the star did not render")
			}
			if got := opaqueAt(img, 50, 50); got != tc.centreFilled {
				t.Errorf("centre pixel (50,50) filled = %v, want %v", got, tc.centreFilled)
			}
		})
	}
}

// renderStar parses a 100x100 document holding one self-intersecting star
// path with the given fill-rule attribute (empty for none) and rasters it
// through a winding-aware scanner.
func renderStar(t *testing.T, fillRuleAttr string) *image.RGBA {
	t.Helper()
	doc := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">` +
		`<path d="M50,0 L79,91 L2,35 L98,35 L21,91 Z"` +
		fillRuleAttr + `/></svg>`
	icon, err := parser.NewParser(parser.StrictErrorMode).ParseStream(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("parse star document: %v", err)
	}
	const w, h = 100, 100
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := scanFT.NewScannerFT(w, h, scanFT.NewRGBAPainter(img))
	driver.Draw(NewDriver(w, h, scanner), icon, 1.0)
	return img
}

// opaqueAt reports whether a pixel is more opaque than not, so antialiasing
// at a shape boundary cannot flip an assertion.
func opaqueAt(img *image.RGBA, x, y int) bool {
	return img.RGBAAt(x, y).A >= 0x80
}
