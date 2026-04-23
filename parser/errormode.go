package parser

import (
	"fmt"
	"log"
)

// ErrorMode is the for setting how the parser reacts to unparsed elements
type ErrorMode uint8

const (
	// IgnoreErrorMode skips unparsed SVG elements
	IgnoreErrorMode ErrorMode = iota
	// WarnErrorMode outputs a warning when an unparsed SVG element is found
	WarnErrorMode
	// StrictErrorMode causes a error when an unparsed SVG element is found
	StrictErrorMode
)

// treat the error according to the ErrorMode
func HandleError(mode ErrorMode, originFmt string, args ...any) error {
	formatted := fmt.Sprintf(originFmt, args...)
	switch mode {
	case StrictErrorMode:
		return Error(formatted)
	case WarnErrorMode:
		log.Println(formatted) // then return nil
	}
	return nil
}
