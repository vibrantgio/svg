package svg

import "image/color"

// Color is enables you to differentiate between a valid and invalid (none) color.
type Color struct {
	valid bool
	color PlainColor
}

var InvalidColor = Color{}

func ValidColor(r, g, b, a uint8) Color {
	return Color{valid: true, color: PlainColor{R: r, G: g, B: b, A: a}}
}

func (o Color) AsColor() color.Color {
	if o.valid {
		return o.color
	}
	return nil
}

func (o Color) AsPattern() Pattern {
	if o.valid {
		return o.color
	}
	return nil
}
