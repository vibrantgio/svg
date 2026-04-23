package main

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func IMG(img image.Image) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max

		ib := img.Bounds()
		sz := ib.Size()

		ox := float32(0.0)
		oy := float32(0.0)
		sx := float32(size.X) / float32(sz.X)
		sy := float32(size.Y) / float32(sz.Y)

		cstack := clip.Rect{Max: size}.Op().Push(gtx.Ops)
		t := f32.NewAffine2D(sx, 0.0, ox, 0.0, sy, oy)
		tstack := op.Affine(t).Push(gtx.Ops)
		paint.NewImageOp(img).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		tstack.Pop()
		cstack.Pop()

		return layout.Dimensions{Size: size}
	}
}
