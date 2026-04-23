package svg

import (
	"image/color"

	"github.com/vibrantgio/svg/matrix"
	"golang.org/x/image/math/fixed"
)

// Gradient holds a description of an SVG 2.0 gradient
type Gradient struct {
	Direction GradDirection
	Stops     []GradStop
	Bounds    ViewBox
	Matrix    matrix.Matrix2D
	Spread    SpreadMethod
	Units     GradUnits
}

// GradDirection is either Radial or Linear
type GradDirection interface {
	Type() GradType
	Params() []float64
}

type GradType byte

const (
	Linear GradType = iota
	Radial
)

// LinearParams contains the linear gradient parameters x1, y1, x2, y2
type LinearParams [4]float64

func (LinearParams) Type() GradType      { return Linear }
func (p LinearParams) Params() []float64 { return p[:] }

// RadialParams contains the radial gradiant parameters cx, cy, fx, fy, r, fr
type RadialParams [6]float64

func (RadialParams) Type() GradType      { return Radial }
func (p RadialParams) Params() []float64 { return p[:] }

// GradUnits is the type for gradient units
type GradUnits byte

// SVG bounds paremater constants
const (
	ObjectBoundingBox GradUnits = iota
	UserSpaceOnUse
)

// SpreadMethod is the type for spread parameters
type SpreadMethod byte

// SVG spread parameter constants
const (
	PadSpread SpreadMethod = iota
	ReflectSpread
	RepeatSpread
)

// GradStop represents a stop in the SVG 2.0 gradient specification
type GradStop struct {
	StopColor color.Color
	Offset    float64
	Opacity   float64
}

func (Gradient) IsPattern() {
}

// ApplyPathExtent use the given path extent to adjust the bounding box,
// if required by `Units`.
// The `Params` field is not modified, but a matrix accounting for both the
// bounding box and the gradient matrix is returned
func (g *Gradient) ApplyPathExtent(extent fixed.Rectangle26_6) matrix.Matrix2D {
	if g.Units == ObjectBoundingBox {
		mnx, mny := float64(extent.Min.X)/64, float64(extent.Min.Y)/64
		mxx, mxy := float64(extent.Max.X)/64, float64(extent.Max.Y)/64
		g.Bounds.X, g.Bounds.Y = mnx, mny
		g.Bounds.W, g.Bounds.H = mxx-mnx, mxy-mny

		// units in Params are fraction, so we apply bounds
		return matrix.Identity.Scale(g.Bounds.W, g.Bounds.H).Mult(g.Matrix)
	}
	// units in Params are already scaled to the view box
	// just return the gradient matrix
	return g.Matrix
}
