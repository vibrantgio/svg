package pdf

import (
	"io"

	"github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/model"
	"github.com/vibrantgio/svg"
	"github.com/vibrantgio/svg/driver"
	"github.com/vibrantgio/svg/parser"
	"golang.org/x/image/math/fixed"
)

// assert interface conformance
var (
	_ driver.DrawerNG       = (*Renderer)(nil)
	_ driver.FillAndStroker = (*Renderer)(nil)
)

// Renderer is a DrawerNG that writes SVG operations into a PDF content
// stream. It also implements FillAndStroker so that combined fill+stroke
// paths use PDF's native B operator (and emit the path only once).
type Renderer struct {
	pather
	fillOpacityStates   map[float64]*model.GraphicState
	strokeOpacityStates map[float64]*model.GraphicState
	useNonZeroWinding   bool
}

// pather writes path-construction operators into the content stream and
// keeps the bounding box up-to-date, which is needed when painting
// gradients defined in object-bounding-box units.
type pather struct {
	pdf         *contentstream.GraphicStream
	boundingBox BoundingBox
}

func saveApperanceToFile(ap *contentstream.GraphicStream, filename string) error {
	var (
		doc  model.Document
		page model.PageObject
	)
	ap.ApplyToPageObject(&page, true)
	doc.Catalog.Pages.Kids = append(doc.Catalog.Pages.Kids, &page)
	return doc.WriteFile(filename, nil)
}

// RenderSVGIconToPDF reads the given icon and renders it
// into the given file.
func RenderSVGIconToPDF(icon io.Reader, pdfName string) error {
	parsedIcon, err := parser.NewParser(parser.WarnErrorMode).ParseStream(icon)
	if err != nil {
		return err
	}
	ap := contentstream.NewGraphicStream(model.Rectangle{Urx: 595.28, Ury: 841.89})
	renderer := NewRenderer(&ap)
	ap.Ops(
		contentstream.OpSave{},
		contentstream.OpConcat{Matrix: model.Matrix{1, 0, 0, -1, 0, 841.89}},
	)
	driver.Draw(renderer, parsedIcon, 1.0)
	ap.Ops(contentstream.OpRestore{})

	return saveApperanceToFile(&ap, pdfName)
}

// NewRenderer return a renderer which will
// write to the given `pdf`.
func NewRenderer(cs *contentstream.GraphicStream) *Renderer {
	return &Renderer{
		pather:              pather{pdf: cs},
		fillOpacityStates:   make(map[float64]*model.GraphicState),
		strokeOpacityStates: make(map[float64]*model.GraphicState),
	}
}

func fixedTof(a fixed.Point26_6) (model.Fl, model.Fl) {
	return model.Fl(a.X) / 64, model.Fl(a.Y) / 64
}

func fToFixed(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
}

func (p *pather) Clear() {
	p.boundingBox = BoundingBox{}
}

func (p *pather) Start(a fixed.Point26_6) {
	x, y := fixedTof(a)
	p.pdf.Ops(contentstream.OpMoveTo{X: x, Y: y})
	p.boundingBox.Start(a)
}

func (p *pather) Line(b fixed.Point26_6) {
	x, y := fixedTof(b)
	p.pdf.Ops(contentstream.OpLineTo{X: x, Y: y})
	p.boundingBox.Line(b)
}

func (p *pather) QuadBezier(b fixed.Point26_6, c fixed.Point26_6) {
	cx, cy := fixedTof(b)
	x, y := fixedTof(c)
	p.pdf.Ops(contentstream.OpCurveTo1{X2: cx, Y2: cy, X3: x, Y3: y})
	p.boundingBox.QuadBezier(b, c)
}

func (p *pather) CubeBezier(b fixed.Point26_6, c fixed.Point26_6, d fixed.Point26_6) {
	cx0, cy0 := fixedTof(b)
	cx1, cy1 := fixedTof(c)
	x, y := fixedTof(d)
	p.pdf.Ops(contentstream.OpCubicTo{X1: cx0, Y1: cy0, X2: cx1, Y2: cy1, X3: x, Y3: y})
	p.boundingBox.CubeBezier(b, c, d)
}

// Stop emits a close-path operator if closeLoop is true.
// Used by the bounding-box tests; Renderer uses Close instead.
func (p *pather) Stop(closeLoop bool) {
	if closeLoop {
		p.pdf.Ops(contentstream.OpClosePath{})
	}
}

func (r *Renderer) Clear() {
	r.pather.Clear()
	r.useNonZeroWinding = true
}

func (r *Renderer) Close() {
	r.pdf.Ops(contentstream.OpClosePath{})
}

func (r *Renderer) SetWinding(useNonZeroWinding bool) {
	r.useNonZeroWinding = useNonZeroWinding
}

func (r *Renderer) SetStrokeOptions(options driver.StrokeOptions) {
	var capStyle, joinStyle uint8
	switch options.Join.TrailLineCap {
	case svg.ButtCap:
		capStyle = 0
	case svg.RoundCap:
		capStyle = 1
	case svg.SquareCap:
		capStyle = 2
	}
	switch options.Join.LineJoin {
	case svg.Bevel:
		joinStyle = 2
	case svg.Miter:
		joinStyle = 0
	case svg.Round:
		joinStyle = 1
	}

	dash := make([]model.Fl, len(options.Dash.Dash))
	for i, v := range options.Dash.Dash {
		dash[i] = model.Fl(v)
	}
	r.pdf.Ops(
		contentstream.OpSetDash{Dash: model.DashPattern{
			Array: dash,
			Phase: model.Fl(options.Dash.DashOffset),
		}},
		contentstream.OpSetLineWidth{W: model.Fl(options.LineWidth) / 64},
		contentstream.OpSetLineCap{Style: capStyle},
		contentstream.OpSetLineJoin{Style: joinStyle},
		contentstream.OpSetMiterLimit{Limit: model.Fl(options.Join.MiterLimit) / 64},
	)
}

// applyFillColor sets the fill colour and opacity graphic state for the
// next paint operator. Returns the effective opacity (already multiplied
// into the alpha graphic state). Gradients are not yet supported.
// TODO: support gradient
func (r *Renderer) applyFillColor(color svg.Pattern, opacity float64) {
	switch color := color.(type) {
	case svg.PlainColor:
		r.pdf.SetColorFill(color)
		opacity *= float64(color.A) / 255.
		gs, ok := r.fillOpacityStates[opacity]
		if !ok {
			gs = &model.GraphicState{Ca: model.ObjFloat(opacity), BM: []model.Name{"Normal"}}
			r.fillOpacityStates[opacity] = gs
		}
		name := r.pdf.AddExtGState(gs)
		r.pdf.Ops(contentstream.OpSetExtGState{Dict: name})
	case svg.Gradient:
	}
}

// TODO: support gradient
func (r *Renderer) applyStrokeColor(color svg.Pattern, opacity float64) {
	switch color := color.(type) {
	case svg.PlainColor:
		r.pdf.SetColorStroke(color)
		opacity *= float64(color.A) / 255.
		gs, ok := r.strokeOpacityStates[opacity]
		if !ok {
			gs = &model.GraphicState{CA: model.ObjFloat(opacity), BM: []model.Name{"Normal"}}
			r.strokeOpacityStates[opacity] = gs
		}
		name := r.pdf.AddExtGState(gs)
		r.pdf.Ops(contentstream.OpSetExtGState{Dict: name})
	}
}

func (r *Renderer) Fill(color svg.Pattern, opacity float64) {
	r.applyFillColor(color, opacity)
	if r.useNonZeroWinding {
		r.pdf.Ops(contentstream.OpFill{})
	} else {
		r.pdf.Ops(contentstream.OpEOFill{})
	}
}

func (r *Renderer) Stroke(color svg.Pattern, opacity float64) {
	r.applyStrokeColor(color, opacity)
	r.pdf.Ops(contentstream.OpStroke{})
}

// FillAndStroke paints the current path with PDF's B operator (or B* for
// the even-odd rule), avoiding a second emission of the path ops.
func (r *Renderer) FillAndStroke(fillCol svg.Pattern, fillOp float64,
	strokeCol svg.Pattern, strokeOp float64) {
	r.applyFillColor(fillCol, fillOp)
	r.applyStrokeColor(strokeCol, strokeOp)
	if r.useNonZeroWinding {
		r.pdf.Ops(contentstream.OpFillStroke{})
	} else {
		r.pdf.Ops(contentstream.OpEOFillStroke{})
	}
}
