package svg

type Icon interface {
	ViewBox() Bounds
	SetTarget(x, y, w, h float64)
	Draw(d Driver, opacity float64)
}
