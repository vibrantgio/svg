package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"

	"github.com/vibrantgio/style"
	vsvg "github.com/vibrantgio/svg/driver/gio"
	"github.com/vibrantgio/svg/parser"
	vtext "github.com/vibrantgio/text"
)

func main() {
	go Circles()
	app.Main()
}

//go:embed circles.svg
var circles_svg []byte

func Circles() {
	defer catch()

	window := app.NewWindow(
		app.Title("SVG - Circles"),
		app.Size(1000, 700))

	parser := parser.NewParser(parser.WarnErrorMode)
	widget := try(vsvg.IconWidget(try(parser.ParseStream(bytes.NewBuffer(circles_svg))), 0, 0, 1.0))

	ops := new(op.Ops)
	shaper := text.NewShaper(style.FontFaces())
	for event := range window.Events() {
		if frame, ok := event.(system.FrameEvent); ok {
			gtx := layout.NewContext(ops, frame)
			start := time.Now()

			layout.UniformInset(24).Layout(gtx, widget)

			msg := fmt.Sprintf("%v", time.Since(start).Round(time.Microsecond))
			text := vtext.Text(shaper, style.H5, 0.0, 0.0, color.Black, msg)
			layout.UniformInset(12).Layout(gtx, text)
			frame.Frame(ops)
		}
	}

	os.Exit(0)
}
