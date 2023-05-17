package parse

import (
	"fmt"
	"image/color"
	"strings"

	"encoding/xml"
	"errors"
	"log"
	"math"

	"golang.org/x/image/colornames"
	"golang.org/x/image/math/fixed"

	"github.com/reactivego/svg"
	"github.com/reactivego/svg/matrix"
)

// svgCursor is used while parsing SVG files
type svgCursor struct {
	pathCursor
	icon                                    *SVG
	styleStack                              []PathStyle
	grad                                    *svg.Gradient
	inTitleText, inDescText, inGrad, inDefs bool
	currentDef                              []definition
}

// definition is used to store what's given in a def tag
type definition struct {
	ID, Tag string
	Attrs   []xml.Attr
}

// parseUnit converts a length with a unit into its value in 'px'
// percentage are supported, and refer to the current ViewBox
func (c *svgCursor) parseUnit(s string, asPerc percentageReference) (float64, error) {
	return resolveUnit(c.icon.ViewBox, s, asPerc)
}

func fToFixed(f float64) fixed.Int26_6 {
	return fixed.Int26_6(f * 64)
}

// treat the error according to the errorMode
func (c *svgCursor) handleError(originFmt string, args ...interface{}) error {
	formatted := fmt.Sprintf(originFmt, args...)
	switch c.errorMode {
	case svg.StrictErrorMode:
		return errors.New(formatted)
	case svg.WarnErrorMode:
		log.Println(formatted) // then return nil
	}
	return nil
}

func (c *svgCursor) readTransformAttr(m1 matrix.Matrix2D, k string) (matrix.Matrix2D, error) {
	ln := len(c.points)
	switch k {
	case "rotate":
		switch ln {
		case 1:
			m1 = m1.Rotate(c.points[0] * math.Pi / 180)
		case 3:
			m1 = m1.Translate(c.points[1], c.points[2]).
				Rotate(c.points[0]*math.Pi/180).
				Translate(-c.points[1], -c.points[2])
		default:
			return m1, ErrParamMismatch
		}
	case "translate":
		switch ln {
		case 1:
			m1 = m1.Translate(c.points[0], 0)
		case 2:
			m1 = m1.Translate(c.points[0], c.points[1])
		default:
			return m1, ErrParamMismatch
		}
	case "skewx":
		if ln == 1 {
			m1 = m1.SkewX(c.points[0] * math.Pi / 180)
		} else {
			return m1, ErrParamMismatch
		}
	case "skewy":
		if ln == 1 {
			m1 = m1.SkewY(c.points[0] * math.Pi / 180)
		} else {
			return m1, ErrParamMismatch
		}
	case "scale":
		// The scale(<x> [<y>]) transform function specifies a scale operation by x and y. If y is not provided, it is assumed to be equal to x.
		switch ln {
		case 1:
			m1 = m1.Scale(c.points[0], c.points[0])
		case 2:
			m1 = m1.Scale(c.points[0], c.points[1])
		default:
			return m1, ErrParamMismatch
		}
	case "matrix":
		if ln == 6 {
			m1 = m1.Mult(matrix.Matrix2D{
				A: c.points[0],
				B: c.points[1],
				C: c.points[2],
				D: c.points[3],
				E: c.points[4],
				F: c.points[5]})
		} else {
			return m1, ErrParamMismatch
		}
	default:
		return m1, ErrParamMismatch
	}
	return m1, nil
}

func (c *svgCursor) parseTransform(v string) (matrix.Matrix2D, error) {
	ts := strings.Split(v, ")")
	m1 := c.styleStack[len(c.styleStack)-1].Transform
	for _, t := range ts {
		t = strings.TrimSpace(t)
		if len(t) == 0 {
			continue
		}
		d := strings.Split(t, "(")
		if len(d) != 2 || len(d[1]) < 1 {
			return m1, ErrParamMismatch // badly formed transformation
		}
		err := c.getPoints(d[1])
		if err != nil {
			return m1, err
		}
		m1, err = c.readTransformAttr(m1, strings.ToLower(strings.TrimSpace(d[0])))
		if err != nil {
			return m1, err
		}
	}
	return m1, nil
}

func (c *svgCursor) readStyleAttr(curStyle *PathStyle, k, v string) error {
	switch k {
	case "fill":
		gradient, ok := c.readGradURL(v, curStyle.FillColor)
		if ok {
			curStyle.FillColor = gradient
			break
		}
		optCol, err := ParseColor(v)
		curStyle.FillColor = optCol.AsPattern()
		return err
	case "stroke":
		gradient, ok := c.readGradURL(v, curStyle.LineColor)
		if ok {
			curStyle.LineColor = gradient
			break
		}
		optCol, errc := ParseColor(v)
		if errc != nil {
			return errc
		}
		curStyle.LineColor = optCol.AsPattern()
	case "stroke-linegap":
		switch v {
		case "flat":
			curStyle.Join.LineGap = svg.FlatGap
		case "round":
			curStyle.Join.LineGap = svg.RoundGap
		case "cubic":
			curStyle.Join.LineGap = svg.CubicGap
		case "quadratic":
			curStyle.Join.LineGap = svg.QuadraticGap
		default:
			return c.handleError("unsupported value '%s' for <stroke-linegap>", v)
		}
	case "stroke-leadlinecap":
		switch v {
		case "butt":
			curStyle.Join.LeadLineCap = svg.ButtCap
		case "round":
			curStyle.Join.LeadLineCap = svg.RoundCap
		case "square":
			curStyle.Join.LeadLineCap = svg.SquareCap
		case "cubic":
			curStyle.Join.LeadLineCap = svg.CubicCap
		case "quadratic":
			curStyle.Join.LeadLineCap = svg.QuadraticCap
		default:
			return c.handleError("unsupported value '%s' for <stroke-leadlinecap>", v)
		}
	case "stroke-linecap":
		switch v {
		case "butt":
			curStyle.Join.TrailLineCap = svg.ButtCap
		case "round":
			curStyle.Join.TrailLineCap = svg.RoundCap
		case "square":
			curStyle.Join.TrailLineCap = svg.SquareCap
		case "cubic":
			curStyle.Join.TrailLineCap = svg.CubicCap
		case "quadratic":
			curStyle.Join.TrailLineCap = svg.QuadraticCap
		default:
			return c.handleError("unsupported value '%s' for <stroke-linecap>", v)
		}
	case "stroke-linejoin":
		switch v {
		case "miter":
			curStyle.Join.LineJoin = svg.Miter
		case "miter-clip":
			curStyle.Join.LineJoin = svg.MiterClip
		case "arc-clip":
			curStyle.Join.LineJoin = svg.ArcClip
		case "round":
			curStyle.Join.LineJoin = svg.Round
		case "arc":
			curStyle.Join.LineJoin = svg.Arc
		case "bevel":
			curStyle.Join.LineJoin = svg.Bevel
		default:
			return c.handleError("unsupported value '%s' for <stroke-linejoin>", v)
		}
	case "stroke-miterlimit":
		mLimit, err := parseBasicFloat(v)
		if err != nil {
			return err
		}
		curStyle.Join.MiterLimit = fToFixed(mLimit)
	case "stroke-width":
		width, err := c.parseUnit(v, widthPercentage)
		if err != nil {
			return err
		}
		curStyle.LineWidth = width
	case "stroke-dashoffset":
		dashOffset, err := c.parseUnit(v, diagPercentage)
		if err != nil {
			return err
		}
		curStyle.Dash.DashOffset = dashOffset
	case "stroke-dasharray":
		if v != "none" {
			dashes := splitOnCommaOrSpace(v)
			dList := make([]float64, len(dashes))
			for i, dstr := range dashes {
				d, err := c.parseUnit(strings.TrimSpace(dstr), diagPercentage)
				if err != nil {
					return err
				}
				dList[i] = d
			}
			curStyle.Dash.Dash = dList
			break
		}
	case "opacity", "stroke-opacity", "fill-opacity":
		op, err := parseBasicFloat(v)
		if err != nil {
			return err
		}
		if k != "stroke-opacity" {
			curStyle.FillOpacity *= op
		}
		if k != "fill-opacity" {
			curStyle.LineOpacity *= op
		}
	case "transform":
		m, err := c.parseTransform(v)
		if err != nil {
			return err
		}
		curStyle.Transform = m
	}
	return nil
}

// pushStyle parses the style element, and push it on the style stack. Only color and opacity are supported
// for fill. Note that this parses both the contents of a style attribute plus
// direct fill and opacity attributes.
func (c *svgCursor) pushStyle(attrs []xml.Attr) error {
	var pairs []string
	for _, attr := range attrs {
		switch strings.ToLower(attr.Name.Local) {
		case "style":
			pairs = append(pairs, strings.Split(attr.Value, ";")...)
		default:
			pairs = append(pairs, attr.Name.Local+":"+attr.Value)
		}
	}
	// Make a copy of the top style
	curStyle := c.styleStack[len(c.styleStack)-1]
	for _, pair := range pairs {
		kv := strings.Split(pair, ":")
		if len(kv) >= 2 {
			k := strings.ToLower(kv[0])
			k = strings.TrimSpace(k)
			v := strings.TrimSpace(kv[1])
			err := c.readStyleAttr(&curStyle, k, v)
			if err != nil {
				return err
			}
		}
	}
	c.styleStack = append(c.styleStack, curStyle) // Push style onto stack
	return nil
}

// splitOnCommaOrSpace returns a list of strings after splitting the input on comma and space delimiters
func splitOnCommaOrSpace(s string) []string {
	return strings.FieldsFunc(s,
		func(r rune) bool {
			return r == ',' || r == ' '
		})
}

func (c *svgCursor) readStartElement(se xml.StartElement) (err error) {
	var skipDef bool
	if se.Name.Local == "radialGradient" || se.Name.Local == "linearGradient" || c.inGrad {
		skipDef = true
	}
	if c.inDefs && !skipDef {
		ID := ""
		for _, attr := range se.Attr {
			if attr.Name.Local == "id" {
				ID = attr.Value
			}
		}
		if ID != "" && len(c.currentDef) > 0 {
			c.icon.defs[c.currentDef[0].ID] = c.currentDef
			c.currentDef = make([]definition, 0)
		}
		c.currentDef = append(c.currentDef, definition{
			ID:    ID,
			Tag:   se.Name.Local,
			Attrs: se.Attr,
		})
		return nil
	}
	df, ok := elementFuncs[se.Name.Local]
	if !ok {
		errStr := "Cannot process svg element " + se.Name.Local
		switch c.errorMode {
		case svg.StrictErrorMode:
			return errors.New(errStr)
		case svg.WarnErrorMode:
			log.Println(errStr)
		}
		return nil
	}
	err = df(c, se.Attr)

	if len(c.path) > 0 {
		//The cursor parsed a path from the xml element
		pathCopy := append(Path{}, c.path...)
		c.icon.Paths = append(c.icon.Paths,
			StyledPath{Path: pathCopy, Style: c.styleStack[len(c.styleStack)-1]})
		c.path = c.path[:0]
	}
	return
}

// readGradURL reads an SVG format gradient url
// Since the context of the gradient can affect the colors
// the current fill or line color is passed in and used in
// the case of a nil stopClor value
func (c *svgCursor) readGradURL(v string, defaultColor svg.Pattern) (grad svg.Gradient, ok bool) {
	if strings.HasPrefix(v, "url(") && strings.HasSuffix(v, ")") {
		urlStr := strings.TrimSpace(v[4 : len(v)-1])
		if strings.HasPrefix(urlStr, "#") {
			var g *svg.Gradient
			g, ok = c.icon.grads[urlStr[1:]]
			if ok {
				grad = localizeGradIfStopClrNil(g, defaultColor)
			}
		}
	}
	return
}

// readGradAttr reads an SVG gradient attribute
func (c *svgCursor) readGradAttr(attr xml.Attr) (err error) {
	switch attr.Name.Local {
	case "gradientTransform":
		c.grad.Matrix, err = c.parseTransform(attr.Value)
	case "gradientUnits":
		switch strings.TrimSpace(attr.Value) {
		case "userSpaceOnUse":
			c.grad.Units = svg.UserSpaceOnUse
		case "objectBoundingBox":
			c.grad.Units = svg.ObjectBoundingBox
		}
	case "spreadMethod":
		switch strings.TrimSpace(attr.Value) {
		case "pad":
			c.grad.Spread = svg.PadSpread
		case "reflect":
			c.grad.Spread = svg.ReflectSpread
		case "repeat":
			c.grad.Spread = svg.RepeatSpread
		}
	}
	return
}

func localizeGradIfStopClrNil(g *svg.Gradient, defaultColor svg.Pattern) svg.Gradient {
	grad := *g
	for _, s := range grad.Stops {
		if s.StopColor == nil { // This means we need copy the gradient's Stop slice
			// and fill in the default color

			// Copy the stops
			stops := append([]svg.GradStop{}, grad.Stops...)
			grad.Stops = stops
			// Use the background color when a stop color is nil
			clr := getColor(defaultColor)
			for i, s := range stops {
				if s.StopColor == nil {
					grad.Stops[i].StopColor = clr
				}
			}
			break // Only need to do this once
		}
	}
	return grad
}

// getColor is a helper function to get the background color
// if readGradUrl needs it.
func getColor(clr svg.Pattern) color.Color {
	switch c := clr.(type) {
	case svg.Gradient: // This is a bit lazy but oh well
		for _, s := range c.Stops {
			if s.StopColor != nil {
				return s.StopColor
			}
		}
	case svg.PlainColor:
		return c
	}
	return colornames.Black
}
