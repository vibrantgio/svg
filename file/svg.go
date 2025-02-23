package file

import "github.com/reactivego/svg"

// SVG holds data from parsed SVGs.
// See the `Draw` methods to use it.
type SVG struct {
	ViewBox      svg.ViewBox
	Titles       []string // Title elements collect here
	Descriptions []string // Description elements collect here
	Paths        []svg.StyledPath

	Width, Height string // top level width and height attributes

	Grads map[string]*svg.Gradient
	Defs  map[string][]svg.Definition
}

func NewSVG() *SVG {
	return &SVG{Defs: make(map[string][]svg.Definition), Grads: make(map[string]*svg.Gradient)}
}
