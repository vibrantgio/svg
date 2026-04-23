package parser

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/vibrantgio/svg"
	"golang.org/x/image/colornames"
)

// Color is enables you to differentiate between a valid and invalid (none) color.
type Color struct {
	valid bool
	color svg.PlainColor
}

var InvalidColor = Color{}

func ValidColor(r, g, b, a uint8) Color {
	return Color{valid: true, color: svg.PlainColor{R: r, G: g, B: b, A: a}}
}

func (o Color) AsColor() color.Color {
	if o.valid {
		return o.color
	}
	return nil
}

func (o Color) AsPattern() svg.Pattern {
	if o.valid {
		return o.color
	}
	return nil
}

// ParseColor parses an SVG color string in all forms including all
// SVG1.1 names, obtained from the package golang.org/x/image/colornames.
func ParseColor(colorStr string) (Color, error) {
	colorStr = strings.ToLower(colorStr)

	// We are not handling urls and gradients and stuff at this point
	if strings.HasPrefix(colorStr, "url") {
		return ValidColor(0, 0, 0, 255), nil
	}

	// none signals that the function (fill or stroke) is off; not the same as black
	if colorStr == "none" {
		return InvalidColor, nil
	}

	// lookup named colors defined in the SVG 1.1 spec.
	if cn, ok := colornames.Map[colorStr]; ok {
		return ValidColor(cn.R, cn.G, cn.B, cn.A), nil
	}

	// read the rgb color string e.g. rgb(255, 0, 0)
	cStr := strings.TrimPrefix(colorStr, "rgb(")
	if cStr != colorStr {
		cStr := strings.TrimSuffix(cStr, ")")
		vals := strings.Split(cStr, ",")
		if len(vals) != 3 {
			return InvalidColor, ErrParamMismatch
		}
		var cvals [3]uint8
		for i, v := range vals {
			if strings.HasSuffix(v, "%") {
				v = strings.TrimSpace(strings.TrimSuffix(v, "%"))
				if n, err := strconv.Atoi(v); err != nil {
					return InvalidColor, err
				} else {
					cvals[i] = uint8((n * 0xFF) / 100)
				}
			} else {
				if n, err := strconv.Atoi(strings.TrimSpace(v)); err != nil {
					return InvalidColor, err
				} else {
					if n > 255 {
						n = 255
					}
					cvals[i] = uint8(n)
				}
			}
		}
		return ValidColor(cvals[0], cvals[1], cvals[2], 0xFF), nil
	}

	// read the SFG color string e.g. #FBD9BD
	if colorStr[0] == '#' {
		colorStr = strings.TrimPrefix(colorStr, "#")
		// SVG specs say duplicate characters in case of 3 digit hex number
		if len(colorStr) == 3 {
			colorStr = string([]byte{colorStr[0], colorStr[0],
				colorStr[1], colorStr[1], colorStr[2], colorStr[2]})
		}
		if len(colorStr) != 6 {
			return InvalidColor, ErrParamMismatch
		}
		var err error
		var r, g, b uint64
		if r, err = strconv.ParseUint(colorStr[0:2], 16, 8); err != nil {
			return InvalidColor, err
		}
		if g, err = strconv.ParseUint(colorStr[2:4], 16, 8); err != nil {
			return InvalidColor, err
		}
		if b, err = strconv.ParseUint(colorStr[4:6], 16, 8); err != nil {
			return InvalidColor, err
		}
		return ValidColor(uint8(r), uint8(g), uint8(b), 0xFF), nil
	}
	return InvalidColor, ErrParamMismatch
}
