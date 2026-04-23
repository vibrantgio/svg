package parser

import (
	"encoding/xml"
	"strings"

	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/matrix"
	"golang.org/x/image/math/fixed"
)

func init() {
	// avoids cyclical static declaration
	// called on package initialization
	elementFuncs["use"] = useF
}

type svgFunc func(c *svgCursor, attrs []xml.Attr) error

var elementFuncs = map[string]svgFunc{
	"svg":            svgF,
	"g":              gF,
	"line":           lineF,
	"stop":           stopF,
	"rect":           rectF,
	"circle":         circleF,
	"ellipse":        circleF, // circleF handles ellipse also
	"polyline":       polylineF,
	"polygon":        polygonF,
	"path":           pathF,
	"desc":           descF,
	"defs":           defsF,
	"title":          titleF,
	"linearGradient": linearGradientF,
	"radialGradient": radialGradientF,
}

func svgF(c *svgCursor, attrs []xml.Attr) error {
	c.icon.ViewBox.X = 0
	c.icon.ViewBox.Y = 0
	c.icon.ViewBox.W = 0
	c.icon.ViewBox.H = 0
	var width, height float64
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "viewBox":
			err = c.getPoints(attr.Value)
			if len(c.points) != 4 {
				return ErrParamMismatch
			}
			c.icon.ViewBox.X = c.points[0]
			c.icon.ViewBox.Y = c.points[1]
			c.icon.ViewBox.W = c.points[2]
			c.icon.ViewBox.H = c.points[3]
		case "width":
			c.icon.Width = attr.Value
			width, err = parseBasicFloat(attr.Value)
		case "height":
			c.icon.Height = attr.Value
			height, err = parseBasicFloat(attr.Value)
		}
		if err != nil {
			return err
		}
	}
	if c.icon.ViewBox.W == 0 {
		c.icon.ViewBox.W = width
	}
	if c.icon.ViewBox.H == 0 {
		c.icon.ViewBox.H = height
	}
	return nil
}

func gF(c *svgCursor, attrs []xml.Attr) error {
	// g does nothing but push the style
	return nil
}

func rectF(c *svgCursor, attrs []xml.Attr) error {
	var x, y, w, h, rx, ry float64
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "x":
			x, err = c.parseUnit(attr.Value, widthPercentage)
		case "y":
			y, err = c.parseUnit(attr.Value, heightPercentage)
		case "width":
			w, err = c.parseUnit(attr.Value, widthPercentage)
		case "height":
			h, err = c.parseUnit(attr.Value, heightPercentage)
		case "rx":
			rx, err = c.parseUnit(attr.Value, widthPercentage)
		case "ry":
			ry, err = c.parseUnit(attr.Value, heightPercentage)
		}
		if err != nil {
			return err
		}
	}
	if w == 0 || h == 0 {
		return nil
	}
	c.path.AddRoundRect(x+c.curX, y+c.curY, w+x+c.curX, h+y+c.curY, rx, ry, 0)
	return nil
}

func circleF(c *svgCursor, attrs []xml.Attr) error {
	var cx, cy, rx, ry float64
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "cx":
			cx, err = c.parseUnit(attr.Value, widthPercentage)
		case "cy":
			cy, err = c.parseUnit(attr.Value, heightPercentage)
		case "r":
			rx, err = c.parseUnit(attr.Value, diagPercentage)
			ry = rx
		case "rx":
			rx, err = c.parseUnit(attr.Value, widthPercentage)
		case "ry":
			ry, err = c.parseUnit(attr.Value, heightPercentage)
		}
		if err != nil {
			return err
		}
	}
	if rx == 0 || ry == 0 { // not drawn, but not an error
		return nil
	}
	c.ellipseAt(cx+c.curX, cy+c.curY, rx, ry)
	return nil
}

func lineF(c *svgCursor, attrs []xml.Attr) error {
	var x1, x2, y1, y2 float64
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "x1":
			x1, err = c.parseUnit(attr.Value, widthPercentage)
		case "x2":
			x2, err = c.parseUnit(attr.Value, widthPercentage)
		case "y1":
			y1, err = c.parseUnit(attr.Value, heightPercentage)
		case "y2":
			y2, err = c.parseUnit(attr.Value, heightPercentage)
		}
		if err != nil {
			return err
		}
	}
	c.path.Start(fixed.Point26_6{
		X: fixed.Int26_6((x1 + c.curX) * 64),
		Y: fixed.Int26_6((y1 + c.curY) * 64),
	})
	c.path.Line(fixed.Point26_6{
		X: fixed.Int26_6((x2 + c.curX) * 64),
		Y: fixed.Int26_6((y2 + c.curY) * 64),
	})
	return nil
}

func polylineF(c *svgCursor, attrs []xml.Attr) error {
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "points":
			err = c.getPoints(attr.Value)
			if len(c.points)%2 != 0 {
				return ErrPolygonHasOddNumberOfPoints
			}
		}
		if err != nil {
			return err
		}
	}
	if len(c.points) > 4 {
		c.path.Start(fixed.Point26_6{
			X: fixed.Int26_6((c.points[0] + c.curX) * 64),
			Y: fixed.Int26_6((c.points[1] + c.curY) * 64),
		})
		for i := 2; i < len(c.points)-1; i += 2 {
			c.path.Line(fixed.Point26_6{
				X: fixed.Int26_6((c.points[i] + c.curX) * 64),
				Y: fixed.Int26_6((c.points[i+1] + c.curY) * 64),
			})
		}
	}
	return nil
}

func polygonF(c *svgCursor, attrs []xml.Attr) error {
	err := polylineF(c, attrs)
	if len(c.points) > 4 {
		c.path.Stop(true)
	}
	return err
}

func pathF(c *svgCursor, attrs []xml.Attr) error {
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "d":
			err = c.compilePath(attr.Value)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func descF(c *svgCursor, attrs []xml.Attr) error {
	c.inDescText = true
	c.icon.Descriptions = append(c.icon.Descriptions, "")
	return nil
}

func titleF(c *svgCursor, attrs []xml.Attr) error {
	c.inTitleText = true
	c.icon.Titles = append(c.icon.Titles, "")
	return nil
}

func defsF(c *svgCursor, attrs []xml.Attr) error {
	c.inDefs = true
	return nil
}

func linearGradientF(c *svgCursor, attrs []xml.Attr) error {
	var err error
	c.inGrad = true
	// interpretation of percentage in gradient parameters depends
	// on gradientUnits: we first store the string values
	// and resolve them in a second pass
	parameterStrings := [4]string{"0%", "0%", "100%", "0"} // default value
	c.grad = &svg.Gradient{Bounds: c.icon.ViewBox, Matrix: matrix.Identity}
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "id":
			id := attr.Value
			if len(id) >= 0 {
				c.icon.Grads[id] = c.grad
			} else {
				return ErrZeroLengthID
			}
		case "x1":
			parameterStrings[0] = attr.Value
		case "y1":
			parameterStrings[1] = attr.Value
		case "x2":
			parameterStrings[2] = attr.Value
		case "y2":
			parameterStrings[3] = attr.Value
		default:
			err = c.readGradAttr(attr)
		}
		if err != nil {
			return err
		}
	}
	// now we can resolve percentages
	bbox := svg.ViewBox{W: 1, H: 1} // default is ObjectBoundingBox
	if c.grad.Units == svg.UserSpaceOnUse {
		bbox = c.grad.Bounds
	}
	var params svg.LinearParams
	params[0], err = resolveUnit(bbox, parameterStrings[0], widthPercentage)
	if err != nil {
		return err
	}
	params[1], err = resolveUnit(bbox, parameterStrings[1], heightPercentage)
	if err != nil {
		return err
	}
	params[2], err = resolveUnit(bbox, parameterStrings[2], widthPercentage)
	if err != nil {
		return err
	}
	params[3], err = resolveUnit(bbox, parameterStrings[3], heightPercentage)
	if err != nil {
		return err
	}
	c.grad.Direction = params
	return nil
}

func radialGradientF(c *svgCursor, attrs []xml.Attr) error {
	c.inGrad = true
	c.grad = &svg.Gradient{Bounds: c.icon.ViewBox, Matrix: matrix.Identity}
	var setFx, setFy bool
	var err error
	parameterStrings := [6]string{"50%", "50%", "50%", "50%", "50%", "50%"} // default values
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "id":
			id := attr.Value
			if len(id) >= 0 {
				c.icon.Grads[id] = c.grad
			} else {
				return ErrZeroLengthID
			}
		case "cx":
			parameterStrings[0] = attr.Value
		case "cy":
			parameterStrings[1] = attr.Value
		case "fx":
			setFx = true
			parameterStrings[2] = attr.Value
		case "fy":
			setFy = true
			parameterStrings[3] = attr.Value
		case "r":
			parameterStrings[4] = attr.Value
		case "fr":
			parameterStrings[5] = attr.Value
		default:
			err = c.readGradAttr(attr)
		}
		if err != nil {
			return err
		}
	}
	if !setFx { // set fx to cx by default
		parameterStrings[2] = parameterStrings[0]
	}
	if !setFy { // set fy to cy by default
		parameterStrings[3] = parameterStrings[1]
	}

	// now we can resolve percentages
	bbox := svg.ViewBox{W: 1, H: 1} // default is ObjectBoundingBox
	if c.grad.Units == svg.UserSpaceOnUse {
		bbox = c.grad.Bounds
	}
	var params svg.RadialParams
	params[0], err = resolveUnit(bbox, parameterStrings[0], widthPercentage)
	if err != nil {
		return err
	}
	params[1], err = resolveUnit(bbox, parameterStrings[1], heightPercentage)
	if err != nil {
		return err
	}
	params[2], err = resolveUnit(bbox, parameterStrings[2], widthPercentage)
	if err != nil {
		return err
	}
	params[3], err = resolveUnit(bbox, parameterStrings[3], heightPercentage)
	if err != nil {
		return err
	}
	params[4], err = resolveUnit(bbox, parameterStrings[4], diagPercentage)
	if err != nil {
		return err
	}
	params[5], err = resolveUnit(bbox, parameterStrings[5], diagPercentage)
	if err != nil {
		return err
	}

	c.grad.Direction = params
	return nil
}

func stopF(c *svgCursor, attrs []xml.Attr) error {
	var err error
	if c.inGrad {
		stop := svg.GradStop{Opacity: 1.0}
		for _, attr := range attrs {
			switch attr.Name.Local {
			case "offset":
				stop.Offset, err = parseFraction(attr.Value)
			case "stop-color":
				// todo: add current color inherit
				var c Color
				c, err = ParseColor(attr.Value)
				stop.StopColor = c.AsColor()
			case "stop-opacity":
				stop.Opacity, err = parseBasicFloat(attr.Value)
			}
			if err != nil {
				return err
			}
		}
		c.grad.Stops = append(c.grad.Stops, stop)
	}
	return nil
}

func useF(c *svgCursor, attrs []xml.Attr) error {
	var (
		href string
		x, y float64
		err  error
	)
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "href":
			href = attr.Value
		case "x":
			x, err = c.parseUnit(attr.Value, widthPercentage)
		case "y":
			y, err = c.parseUnit(attr.Value, heightPercentage)
		}
		if err != nil {
			return err
		}
	}
	c.curX, c.curY = x, y
	defer func() {
		c.curX, c.curY = 0, 0
	}()
	if href == "" {
		return ErrOnlyUseTagsWithHrefIsSupported
	}
	if !strings.HasPrefix(href, "#") {
		return ErrOnlyTheIdCssSelectorIsSupported
	}
	defs, ok := c.icon.Defs[href[1:]]
	if !ok {
		return ErrHrefIdInUseStatementWasNotFoundInSavedDef
	}
	for _, def := range defs {
		if def.Tag == "endg" {
			// pop style
			c.styleStack = c.styleStack[:len(c.styleStack)-1]
			continue
		}
		if err = c.pushStyle(def.Attrs); err != nil {
			return err
		}
		df, ok := elementFuncs[def.Tag]
		if !ok {
			errStr := "Cannot process svg element " + def.Tag
			return HandleError(c.errorMode, errStr)
		}
		if err := df(c, def.Attrs); err != nil {
			return err
		}
		if def.Tag != "g" {
			// pop style
			c.styleStack = c.styleStack[:len(c.styleStack)-1]
		}
	}
	return nil
}
