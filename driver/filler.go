package driver

type Filler interface {
	Drawer

	// Decide to use or not the "non-zero winding" rule for the current path
	SetWinding(useNonZeroWinding bool)
}
