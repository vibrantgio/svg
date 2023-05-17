package parse

type Error string

func (e Error) Error() string { return string(e) }

var (
	ErrParamMismatch  = Error("param mismatch")
	ErrCommandUnknown = Error("unknown command")
	ErrZeroLengthID   = Error("zero length id")
)
