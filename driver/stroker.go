package driver

import (
	"golang.org/x/image/math/fixed"

	"github.com/vibrantgio/svg"
)

type Stroker interface {
	Drawer

	// Parametrize the stroking style for the current path
	SetStrokeOptions(options StrokeOptions)
}

type StrokeOptions struct {
	LineWidth fixed.Int26_6 // width of the line
	Dash      svg.DashOptions
	Join      svg.JoinOptions
}
