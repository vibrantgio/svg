package file

import (
	"encoding/xml"
	"io"
	"os"

	"github.com/reactivego/svg"
	"golang.org/x/net/html/charset"
)

// ReadFromFile reads the icon from the named file.
// This only supports a sub-set of SVG, but this is enough to draw many icons.
// The errMode determines if the icon ignores, errors out, or logs a warning
// when it does not handle an element found in the icon file.
func ReadFromFile(filename string, errMode ErrorMode) (*SVG, error) {
	fin, errf := os.Open(filename)
	if errf != nil {
		return nil, errf
	}
	defer fin.Close()
	return ReadFromStream(fin, errMode)
}

// ReadFromStream reads the Document from the given io.Reader
// This only supports a sub-set of SVG, but is enough to draw many icons.
// errMode determines if the icon ignores, errors out, or logs a warning
// if it does not handle an element found in the icon file.
func ReadFromStream(stream io.Reader, errMode ErrorMode) (*SVG, error) {
	doc := NewSVG()
	cursor := &svgCursor{styleStack: []svg.PathStyle{svg.DefaultStyle}, icon: doc}
	cursor.errorMode = errMode
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
	return doc, nil
}
