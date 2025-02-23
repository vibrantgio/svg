package file

import (
	"strconv"
	"strings"
)

type unit uint8

// Absolute units supported.
const (
	Px unit = iota
	Cm
	Mm
	Pt
	In
	Q
	Pc
	Perc // Special case : percentage (%) relative to the viewbox
)

// convert the unit to pixels. Return true if it is a %
func parseUnit(s string) (float64, bool, error) {
	unit, value := findUnit(s)
	out, err := strconv.ParseFloat(value, 64)
	return out * toPx[unit], unit == Perc, err
}

var toPx = [...]float64{Px: 1, Cm: 96. / 2.54, Mm: 9.6 / 2.54, Pt: 96. / 72., In: 96., Q: 96. / 40. / 2.54, Pc: 96. / 6., Perc: 1}

// look for an absolute unit, or nothing (considered as pixels)
// % is also supported
func findUnit(s string) (u unit, value string) {
	s = strings.TrimSpace(s)
	for u, suffix := range absoluteUnits {
		if strings.HasSuffix(s, suffix) {
			valueS := strings.TrimSpace(strings.TrimSuffix(s, suffix))
			return unit(u), valueS
		}
	}
	return Px, s
}

var absoluteUnits = [...]string{Px: "px", Cm: "cm", Mm: "mm", Pt: "pt", In: "in", Q: "Q", Pc: "pc", Perc: "%"}
