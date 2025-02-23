package svg

import "encoding/xml"

// Definition is used to store what's given in a def tag
type Definition struct {
	ID, Tag string
	Attrs   []xml.Attr
}
