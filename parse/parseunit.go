package parse

import (
	"strconv"
)

// convert the unit to pixels. Return true if it is a %
func parseUnit(s string) (float64, bool, error) {
	unit, value := findUnit(s)
	out, err := strconv.ParseFloat(value, 64)
	return out * toPx[unit], unit == Perc, err
}

var toPx = [...]float64{Px: 1, Cm: 96. / 2.54, Mm: 9.6 / 2.54, Pt: 96. / 72., In: 96., Q: 96. / 40. / 2.54, Pc: 96. / 6., Perc: 1}
