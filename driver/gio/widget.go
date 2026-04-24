package gio

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"

	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/driver"
)

// IconWidget returns a widget that renders the given IconWidget data using a clip.Path.
// According to the IconWidget specification, default value when the preserveAspectRatio attribute
// is not specified is "xMidYMid meet". This means that the image is scaled to fit the viewport
// while preserving the aspect ratio. The image is centered in the viewport along the x and y axes.
func IconWidget(icon *svg.Icon, width, height unit.Dp, opacity float64) (layout.Widget, error) {
	return func(gtx layout.Context) layout.Dimensions {
		if width == 0 {
			width = unit.Dp(icon.ViewBox.W)
		}
		if height == 0 {
			height = unit.Dp(icon.ViewBox.H)
		}
		size := gtx.Constraints.Constrain(image.Pt(gtx.Dp(width), gtx.Dp(height)))
		icon.SetTarget(icon.ViewBox.AspectMeet(float64(size.X), float64(size.Y), 0.5, 0.5))
		drv := NewDriver(gtx.Ops)
		driver.Draw(drv, icon, opacity)
		return layout.Dimensions{Size: size}
	}, nil
}
