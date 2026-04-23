package parser

import (
	"math"

	"github.com/vibrantgio/svg"
)

type percentageReference uint8

const (
	widthPercentage percentageReference = iota
	heightPercentage
	diagPercentage
)

// resolveUnit converts a length with a unit into its value in 'px' percentage
// are supported, and refer to the viewBox `asPerc` is only applied when `s`
// contains a percentage.
func resolveUnit(viewBox svg.ViewBox, s string, asPerc percentageReference) (float64, error) {
	value, isPercentage, err := parseUnit(s)
	if err != nil {
		return 0, err
	}
	if isPercentage {
		w, h := viewBox.W, viewBox.H
		switch asPerc {
		case widthPercentage:
			return value / 100 * w, nil
		case heightPercentage:
			return value / 100 * h, nil
		case diagPercentage:
			normalizedDiag := math.Sqrt(w*w+h*h) / root2
			return value / 100 * normalizedDiag, nil
		}
	}
	return value, nil
}

var root2 = math.Sqrt(2)
