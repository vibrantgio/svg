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
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"

	"github.com/reactivego/gio"
	"github.com/reactivego/gio/style"
	"github.com/reactivego/gio/svg"
	"github.com/vibrantgio/svg/parser"
	svggio "github.com/vibrantgio/svg/driver/gio"

	primitives "github.com/fogleman/primitive/primitive"
	"github.com/nfnt/resize"
)

const save_to_output_png = false

func main() {
	go Berries()
	app.Main()
}

//go:embed berries.jpg
var berries_jpg []byte

func Berries() {
	defer catch()

	window := app.NewWindow(
		app.Title("SVG - Primitive"),
		app.Size(1000, 700))

	// Decode the input image
	input := try(jpeg.Decode(bytes.NewBuffer(berries_jpg)))
	thumbnail := resize.Thumbnail(256, 256, input, resize.Bilinear)

	// Initialize the primitive model
	bg := primitives.MakeColor(primitives.AverageImageColor(thumbnail))
	model := primitives.NewModel(thumbnail, bg, 1024, runtime.NumCPU())

	// Run the algorithm for a specified number of steps
	const steps = 1000
	const save_every = 20
	poison := must(make(chan struct{}))
	widgets := must(make(chan layout.Widget, 8))
	go func(poison chan struct{}) {
		defer close(widgets)
		for i := 0; i < steps/save_every; i++ {
			for i := 0; i < save_every; i++ {
				select {
				case <-poison:
					return
				default:
					model.Step(primitives.ShapeTypeTriangle, 128, 0)
				}
			}
			par := parser.NewParser(parser.WarnErrorMode)
			if icon, err := par.ParseStream(bytes.NewBufferString(model.SVG())); err == nil {
				if widget, err := svggio.IconWidget(icon, 128, 128, 1.0); err == nil {
					select {
					case <-poison:
						return
					case widgets <- widget:
					}
				}
			}
		}
	}(poison)

	if save_to_output_png {
		output := model.Context.Image()
		noerr(primitives.SavePNG("output.png", output))
	}

	shaper := must(text.NewShaper(style.FontFaces()))

	ops := new(op.Ops)
	var icon layout.Widget
	for event := range window.Events() {
		if frame, ok := event.(system.FrameEvent); ok {
			gtx := layout.NewContext(ops, frame)
			start := time.Now()

			// Receive the next widget
			select {
			case i, received := <-widgets:
				if received {
					icon = i
					op.InvalidateOp{}.Add(ops)
				}
			default:
				op.InvalidateOp{At: time.Now().Add(250 * time.Millisecond)}.Add(ops)
			}

			if icon != nil {
				layout.UniformInset(24).Layout(gtx, icon)
			}

			msg := fmt.Sprintf("%v", time.Since(start).Round(time.Microsecond))
			text := gio.Text(shaper, style.H5, 0.0, 0.0, color.Black, msg)
			layout.UniformInset(12).Layout(gtx, text)
			frame.Frame(ops)
		}
	}

	close(poison)
	for range widgets {
		fmt.Println("sinking an icon")
	}

	os.Exit(0)
}
