package main

import "image"

// Meet will scale viewbox to fit inside the viewport, preserving aspect ratio.
// The ax,ay parameters in the range [0,1] specify the alignment of the viewbox
// within the viewport, with e.g. ax=0.5, ay=0.5 centering the viewbox.
func Meet(viewport, viewbox image.Point, ax, ay float64) image.Rectangle {
	vpx, vpy := float64(viewport.X), float64(viewport.Y)
	vbx, vby := float64(viewbox.X), float64(viewbox.Y)
	aspect := vbx / vby
	if vpx/vpy < aspect {
		h := vpx / aspect
		return image.Rect(0, int((vpy-h)*ay), int(vpx), int(h))
	} else {
		w := vpy * aspect
		return image.Rect(int((vpx-w)*ax), 0, int(w), int(vpy))
	}
}
