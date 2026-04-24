// Package dummy is a debugging driver that logs every draw call.
// It does not render anything.
package dummy

import (
	"fmt"

	"golang.org/x/image/math/fixed"

	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/driver"
)

// assert interface conformance
var _ driver.DrawerNG = (*Dummy)(nil)

func NewDriver() *Dummy {
	return &Dummy{}
}

// Dummy is a DrawerNG that logs each call for debugging.
type Dummy struct {
	nonZeroWinding bool
	strokeOptions  driver.StrokeOptions
}

func (drv *Dummy) Clear() {
	fmt.Println("Clear")
	drv.nonZeroWinding = true
	drv.strokeOptions = driver.StrokeOptions{}
}

func (drv *Dummy) SetWinding(useNonZeroWinding bool) {
	drv.nonZeroWinding = useNonZeroWinding
	fmt.Printf("SetWinding useNonZeroWinding=%t\n", useNonZeroWinding)
}

func (drv *Dummy) SetStrokeOptions(options driver.StrokeOptions) {
	drv.strokeOptions = options
	fmt.Printf("SetStrokeOptions options=%v\n", options)
}

func (drv *Dummy) Start(a fixed.Point26_6) {
	fmt.Printf("Start a=%v\n", a)
}

func (drv *Dummy) Line(b fixed.Point26_6) {
	fmt.Printf("Line b=%v\n", b)
}

func (drv *Dummy) QuadBezier(b, c fixed.Point26_6) {
	fmt.Printf("QuadBezier b=%v c=%v\n", b, c)
}

func (drv *Dummy) CubeBezier(b, c, d fixed.Point26_6) {
	fmt.Printf("CubeBezier b=%v c=%v d=%v\n", b, c, d)
}

func (drv *Dummy) Close() {
	fmt.Println("Close")
}

func (drv *Dummy) Fill(color svg.Pattern, opacity float64) {
	fmt.Printf("Fill color=%v opacity=%v\n", color, opacity)
}

func (drv *Dummy) Stroke(color svg.Pattern, opacity float64) {
	fmt.Printf("Stroke color=%v opacity=%v\n", color, opacity)
}
