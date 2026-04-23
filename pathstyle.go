package svg

import (
	"github.com/vibrantgio/svg/matrix"
	"golang.org/x/image/math/fixed"
)

// PathStyle holds the state of the SVG style
type PathStyle struct {
	FillOpacity       float64
	LineOpacity       float64
	LineWidth         float64
	UseNonZeroWinding bool

	Join      JoinOptions
	Dash      DashOptions
	FillColor Pattern // either PlainColor or Gradient
	LineColor Pattern // either PlainColor or Gradient

	Transform matrix.Matrix2D // current transform
}

type DashOptions struct {
	Dash       []float64 // values for the dash pattern (nil or an empty slice for no dashes)
	DashOffset float64   // starting offset into the dash array
}

type JoinOptions struct {
	MiterLimit   fixed.Int26_6 // he miter cutoff value for miter, arc, miterclip and arcClip joinModes
	LineJoin     JoinMode      // JoinMode for curve segments
	TrailLineCap CapMode       // capping functions for leading and trailing line ends. If one is nil, the other function is used at both ends.

	LeadLineCap CapMode // not part of the standard specification
	LineGap     GapMode // not part of the standard specification. determines how a gap on the convex side of two lines joining is filled
}

// JoinMode type to specify how segments join.
type JoinMode uint8

// JoinMode constants determine how stroke segments bridge the gap at a join
// ArcClip mode is like MiterClip applied to arcs, and is not part of the SVG2.0
// standard.
const (
	Arc JoinMode = iota // New in SVG2
	Round
	Bevel
	Miter
	MiterClip // New in SVG2
	ArcClip   // Like MiterClip applied to arcs, and is not part of the SVG2.0 standard.
)

func (s JoinMode) String() string {
	switch s {
	case Round:
		return "Round"
	case Bevel:
		return "Bevel"
	case Miter:
		return "Miter"
	case MiterClip:
		return "MiterClip"
	case Arc:
		return "Arc"
	case ArcClip:
		return "ArcClip"
	default:
		return "<unknown JoinMode>"
	}
}

// CapMode defines how to draw caps on the ends of lines
type CapMode uint8

const (
	NilCap CapMode = iota // default value
	ButtCap
	SquareCap
	RoundCap
	CubicCap     // Not part of the SVG2.0 standard.
	QuadraticCap // Not part of the SVG2.0 standard.
)

func (c CapMode) String() string {
	switch c {
	case NilCap:
		return "NilCap"
	case ButtCap:
		return "ButtCap"
	case SquareCap:
		return "SquareCap"
	case RoundCap:
		return "RoundCap"
	case CubicCap:
		return "CubicCap"
	case QuadraticCap:
		return "QuadraticCap"
	default:
		return "<unknown CapMode>"
	}
}

// GapMode defines how to bridge gaps when the miter limit is exceeded,
// and is not part of the SVG2.0 standard.
type GapMode uint8

const (
	NilGap GapMode = iota
	FlatGap
	RoundGap
	CubicGap
	QuadraticGap
)

func (g GapMode) String() string {
	switch g {
	case NilGap:
		return "NilGap"
	case FlatGap:
		return "FlatGap"
	case RoundGap:
		return "RoundGap"
	case CubicGap:
		return "CubicGap"
	case QuadraticGap:
		return "QuadraticGap"
	default:
		return "<unknown GapMode>"
	}
}
