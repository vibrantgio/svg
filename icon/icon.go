package icon

import (
	"io"
	"os"

	"github.com/reactivego/svg"
	"github.com/reactivego/svg/draw"
	"github.com/reactivego/svg/matrix"
	"github.com/reactivego/svg/parse"
)

// ReadFrom reads the icon from the named file.
// This only supports a sub-set of SVG, but this is enough to draw many icons.
// The errMode determines if the icon ignores, errors out, or logs a warning
// when it does not handle an element found in the icon file.
func ReadFrom(filename string, errMode svg.ErrorMode) (svg.Icon, error) {
	fin, errf := os.Open(filename)
	if errf != nil {
		return nil, errf
	}
	defer fin.Close()
	return ReadFromStream(fin, errMode)
}

// ReadFromStream reads the Document from the given io.Reader
// This only supports a sub-set of SVG, but is enough to draw many icons.
// errMode determines if the icon ignores, errors out, or logs a warning
// if it does not handle an element found in the icon file.
func ReadFromStream(stream io.Reader, errMode svg.ErrorMode) (svg.Icon, error) {
	svg := parse.NewSVG()
	if err := svg.ReadFromStream(stream, errMode); err != nil {
		return nil, err
	}
	return &icon{Svg: svg, Transform: matrix.Identity}, nil
}

type icon struct {
	Svg       *parse.SVG
	Transform matrix.Matrix2D
}

func (i *icon) ViewBox() svg.Bounds {
	return i.Svg.ViewBox
}

// SetTarget sets the Transform matrix to draw within the bounds of the rectangle arguments
func (i *icon) SetTarget(x, y, w, h float64) {
	scaleW := w / i.Svg.ViewBox.W
	scaleH := h / i.Svg.ViewBox.H
	i.Transform = matrix.Identity.Translate(x-i.Svg.ViewBox.X*scaleW, y-i.Svg.ViewBox.Y*scaleH).Scale(scaleW, scaleH)
}

// Draw the compiled SVG icon into the driver `d`.
// `opacity` is composed (mutliplied) with the eventual
// <stroke-opacity> and <fill-opacity> style attributes.
// All elements should be contained by the Bounds rectangle of the SvgIcon:
// see `SetTarget` method.
func (i *icon) Draw(d svg.Driver, opacity float64) {
	draw.DrawPathsTransformed(d, i.Svg.Paths, i.Transform, opacity)
}
