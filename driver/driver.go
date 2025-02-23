package driver

// Driver will given a parsed SVG document, implement how to draw it on screen.
// This requires a driver implementation to perform the actual draw operations,
// such as a rasterizer to output .png images or a pdf writer to output .pdf
// documents.
type Driver interface {
	// SetupDrawers returns the backend painters, and will be called at the begining
	// of every path. If the `willXXX` boolean is false, the returned drawer should
	// be nil to avoid useless operations. When both booleans are true, one can assume
	// that the exact same draw operations will be performed on the Filler first and
	// then on the Stroker. This promise may enable the implementation to avoid
	// duplicating filled and stroked paths.
	SetupDrawers(willFill, willStroke bool) (Filler, Stroker)
}
