package file

type Error string

func (e Error) Error() string { return string(e) }

var (
	ErrInvalidSvgXmlIcon                         = Error("invalid svg xml icon")
	ErrPolygonHasOddNumberOfPoints               = Error("polygon has odd number of points")
	ErrOnlyUseTagsWithHrefIsSupported            = Error("only use tags with href is supported")
	ErrOnlyTheIdCssSelectorIsSupported           = Error("only the ID CSS selector is supported")
	ErrHrefIdInUseStatementWasNotFoundInSavedDef = Error("href ID in use statement was not found in saved defs")
	ErrParamMismatch                             = Error("param mismatch")
	ErrCommandUnknown                            = Error("unknown command")
	ErrZeroLengthID                              = Error("zero length id")
)
