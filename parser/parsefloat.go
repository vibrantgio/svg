package parser

import "strings"

func parseBasicFloat(s string) (float64, error) {
	value, _, err := parseUnit(s)
	return value, err
}

func parseFraction(v string) (f float64, err error) {
	v = strings.TrimSpace(v)
	d := 1.0
	if strings.HasSuffix(v, "%") {
		d = 100
		v = strings.TrimSuffix(v, "%")
	}
	f, err = parseBasicFloat(v)
	f /= d
	// Fractions are not clamped: percentages outside [0,1] are legal here.
	return
}
