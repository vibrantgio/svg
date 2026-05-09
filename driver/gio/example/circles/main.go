package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"

	"github.com/vibrantgio/style"
	vsvg "github.com/vibrantgio/svg/driver/gio"
	"github.com/vibrantgio/svg/parser"
	"github.com/vibrantgio/textdraw"
)

func main() {
	go Circles()
	app.Main()
}

//go:embed circles.svg
var circles_svg []byte

func Circles() {
	defer catch()

	window := new(app.Window)
	window.Option(
		app.Title("SVG - Circles"),
		app.Size(1000, 700))

	parser := parser.NewParser(parser.WarnErrorMode)
	widget := vsvg.IconWidget(try(parser.ParseStream(bytes.NewBuffer(circles_svg))), 0, 0, 1.0)

	ops := new(op.Ops)
	shaper := text.NewShaper(text.WithCollection(style.FontFaces()))
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			os.Exit(0)
		case app.FrameEvent:
			gtx := app.NewContext(ops, e)
			start := time.Now()

			layout.UniformInset(24).Layout(gtx, widget)

			msg := fmt.Sprintf("%v", time.Since(start).Round(time.Microsecond))
			text := textdraw.Text(shaper, style.H5, 0.0, 0.0, color.Black, msg)
			layout.UniformInset(12).Layout(gtx, text)
			e.Frame(gtx.Ops)
		}
	}
}
