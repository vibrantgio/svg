package parse

import (
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
