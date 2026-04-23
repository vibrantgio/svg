package parser

import (
	"encoding/xml"
	"io"
	"os"

	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/matrix"
	"golang.org/x/net/html/charset"
)

// Parser holds data from parsed SVGs.
// See the `Draw` methods to use it.
type Parser struct {
	ViewBox      svg.ViewBox
	Titles       []string // Title elements collect here
	Descriptions []string // Description elements collect here
	Paths        []svg.StyledPath

	Width, Height string // top level width and height attributes

	Grads map[string]*svg.Gradient
	Defs  map[string][]svg.Definition

	ErrorMode ErrorMode
}

// NewParser creates and SVG document parser that can parse a subset of SVG2.0.
// The errMode determines if the icon ignores, errors out, or logs a warning
// when it does not handle an element found in the SVG file.
func NewParser(errorMode ErrorMode) *Parser {
	return &Parser{
		Defs:      make(map[string][]svg.Definition),
		Grads:     make(map[string]*svg.Gradient),
		ErrorMode: errorMode}
}

// ParseStream reads the SVG document from the given io.Reader
// This only supports a sub-set of SVG, but is enough to draw many icons.
func (doc *Parser) ParseStream(stream io.Reader) (*svg.Icon, error) {
	cursor := &svgCursor{styleStack: []svg.PathStyle{svg.DefaultStyle}, icon: doc}
	cursor.errorMode = doc.ErrorMode
	decoder := xml.NewDecoder(stream)
	decoder.CharsetReader = charset.NewReaderLabel
	seenTag := false
	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				if !seenTag {
					return nil, ErrInvalidSvgXmlIcon
				}
				break
			}
			return nil, err
		}
		// Inspect the type of the XML token
		switch se := t.(type) {
		case xml.StartElement:
			seenTag = true
			// Reads all recognized style attributes from the start element
			// and places it on top of the styleStack
			err = cursor.pushStyle(se.Attr)
			if err != nil {
				return nil, err
			}
			err = cursor.readStartElement(se)
			if err != nil {
				return nil, err
			}
		case xml.EndElement:
			// pop style
			cursor.styleStack = cursor.styleStack[:len(cursor.styleStack)-1]
			switch se.Name.Local {
			case "g":
				if cursor.inDefs {
					cursor.currentDef = append(cursor.currentDef, svg.Definition{
						Tag: "endg",
					})
				}
			case "title":
				cursor.inTitleText = false
			case "desc":
				cursor.inDescText = false
			case "defs":
				if len(cursor.currentDef) > 0 {
					cursor.icon.Defs[cursor.currentDef[0].ID] = cursor.currentDef
					cursor.currentDef = make([]svg.Definition, 0)
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
	return &svg.Icon{ViewBox: doc.ViewBox, Paths: doc.Paths, Transform: matrix.Identity}, nil
}

// ParseFile reads the SVG document from the named file.
// This only supports a sub-set of SVG, but this is enough to draw many icons.
func (doc *Parser) ParseFile(filename string) (*svg.Icon, error) {
	if fin, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		defer fin.Close()
		return doc.ParseStream(fin)
	}
}
