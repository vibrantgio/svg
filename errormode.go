package svg

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
