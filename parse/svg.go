// Provides parsing and rendering of SVG images.
// SVG files are parsed into an abstract representation,
// which can then be consumed by painting drivers.
// See for example oksvg/svgraster or oksvg/svgpdf .
package parse

import (
	"encoding/xml"
	"errors"
	"io"

	"github.com/reactivego/svg"
	"golang.org/x/net/html/charset"
)

// SVG holds data from parsed SVGs.
// See the `Draw` methods to use it.
type SVG struct {
	ViewBox      svg.Bounds
	Titles       []string // Title elements collect here
	Descriptions []string // Description elements collect here
	Paths        []StyledPath

	Width, Height string // top level width and height attributes

	grads map[string]*svg.Gradient
	defs  map[string][]definition
}

func NewSVG() *SVG {
	return &SVG{defs: make(map[string][]definition), grads: make(map[string]*svg.Gradient)}
}

// ReadFromStream reads the Icon from the given io.Reader
// This only supports a sub-set of SVG, but
// is enough to draw many icons. errMode determines if the icon ignores, errors out, or logs a warning
// if it does not handle an element found in the icon file.
func (doc *SVG) ReadFromStream(stream io.Reader, errMode svg.ErrorMode) error {
	cursor := &svgCursor{styleStack: []PathStyle{DefaultStyle}, icon: doc}
	cursor.errorMode = errMode
	decoder := xml.NewDecoder(stream)
	decoder.CharsetReader = charset.NewReaderLabel
	seenTag := false
	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				if !seenTag {
					return errors.New("invalid svg xml icon")
				}
				break
			}
			return err
		}
		// Inspect the type of the XML token
		switch se := t.(type) {
		case xml.StartElement:
			seenTag = true
			// Reads all recognized style attributes from the start element
			// and places it on top of the styleStack
			err = cursor.pushStyle(se.Attr)
			if err != nil {
				return err
			}
			err = cursor.readStartElement(se)
			if err != nil {
				return err
			}
		case xml.EndElement:
			// pop style
			cursor.styleStack = cursor.styleStack[:len(cursor.styleStack)-1]
			switch se.Name.Local {
			case "g":
				if cursor.inDefs {
					cursor.currentDef = append(cursor.currentDef, definition{
						Tag: "endg",
					})
				}
			case "title":
				cursor.inTitleText = false
			case "desc":
				cursor.inDescText = false
			case "defs":
				if len(cursor.currentDef) > 0 {
					cursor.icon.defs[cursor.currentDef[0].ID] = cursor.currentDef
					cursor.currentDef = make([]definition, 0)
				}
				cursor.inDefs = false
			case "radialGradient", "linearGradient":
				cursor.inGrad = false
			}
		case xml.CharData:
			if cursor.inTitleText {
				doc.Titles[len(doc.Titles)-1] += string(se)
			}
			if cursor.inDescText {
				doc.Descriptions[len(doc.Descriptions)-1] += string(se)
			}
		}
	}
	return nil
}
