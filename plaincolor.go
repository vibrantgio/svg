package svg

import "image/color"

type PlainColor color.NRGBA

func (PlainColor) IsPattern() {
}

func (c PlainColor) RGBA() (r, g, b, a uint32) {
	return color.NRGBA(c).RGBA()
}
