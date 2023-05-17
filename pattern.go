package svg

// Pattern groups a basic color and a gradient pattern
// A nil value may by used to indicated that the function (fill or stroke) is off
type Pattern interface {
	IsPattern()
}
