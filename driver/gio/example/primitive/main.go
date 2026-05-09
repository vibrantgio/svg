package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"os"
	"runtime"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"

	"github.com/vibrantgio/style"
	vsvg "github.com/vibrantgio/svg/driver/gio"
	"github.com/vibrantgio/svg/parser"
	"github.com/vibrantgio/textdraw"

	"github.com/fogleman/primitive/primitive"
	"github.com/nfnt/resize"
)

func main() {
	go Primitive()
	app.Main()
}

//go:embed ipace.jpeg
var ipace_jpeg []byte

func Primitive() {
	defer catch()

	window := new(app.Window)
	window.Option(
		app.Title("SVG - Primitive"),
		app.Size(1000, 700))

	// Decode the input image
	input := try(jpeg.Decode(bytes.NewBuffer(ipace_jpeg)))
	thumbnail := resize.Thumbnail(256, 256, input, resize.Bilinear)

	// Initialize the primitive model
	bg := primitive.MakeColor(primitive.AverageImageColor(thumbnail))
	model := primitive.NewModel(thumbnail, bg, 1024, runtime.NumCPU())

	shaper := text.NewShaper(text.WithCollection(style.FontFaces()))

	// Run the algorithm for a specified number of steps
	const steps = 1000
	const save_every = 20
	widgets := make(chan layout.Widget, 8)
	go func() {
		for i := 0; i < steps/save_every; i++ {
			for i := 0; i < save_every; i++ {
				model.Step(primitive.ShapeTypeTriangle, 128, 0)
			}
			if icon, err := parser.NewParser(parser.WarnErrorMode).ParseStream(bytes.NewBufferString(model.SVG())); err == nil {
				widgets <- vsvg.IconWidget(icon, 128, 128, 1.0)
			}
		}
		close(widgets)
	}()

	ops := new(op.Ops)
	var icon layout.Widget
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			os.Exit(0)
		case app.FrameEvent:
			gtx := app.NewContext(ops, e)
			start := time.Now()

			// Receive the next SVG icon
			select {
			case i, received := <-widgets:
				if received {
					icon = i
					gtx.Execute(op.InvalidateCmd{})
				}
			default:
				gtx.Execute(op.InvalidateCmd{At: time.Now().Add(250 * time.Millisecond)})
			}

			if icon != nil {
				layout.UniformInset(24).Layout(gtx, icon)
			}

			msg := fmt.Sprintf("%v", time.Since(start).Round(time.Microsecond))
			text := textdraw.Text(shaper, style.H5, 0.0, 0.0, color.Black, msg)
			layout.UniformInset(12).Layout(gtx, text)
			e.Frame(gtx.Ops)
		}
	}
}
