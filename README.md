# svg

An SVG parser and renderer for Go, for
[Vibrant Gio](https://github.com/vibrantgio), a design system for native desktop
applications on macOS, Windows and Linux, written in pure Go on
[Gio](https://gioui.org). Forked from
[benoitkugler/oksvg](https://github.com/benoitkugler/oksvg), itself a heavily
modified fork of [srwiley/oksvg](https://github.com/srwiley/oksvg).

The problem this fork exists to solve is that oksvg parses and rasterizes in one
motion: it reads a document and paints pixels. That is exactly wrong for a
retained-mode UI toolkit, where the thing you want is not a bitmap but a
sequence of drawing operations you can put into an op list, transform, clip and
re-emit every frame. Rasterizing to an image first costs a copy per frame, loses
resolution independence, and pins the output to one backend.

So this fork puts an interface where oksvg has a rasterizer. `parser` reads a
document into an `svg.Icon` — a view box, a transform, and styled paths built
out of `Operation` values in 26.6 fixed point — and that is all it does. The
`driver` package walks an icon and calls a `DrawerNG`. What the drawer does with
the calls is its business: `driver/gio` turns them into `clip` and `paint`
operations on an `op.Ops`, `driver/raster` into an `*image.RGBA`, `driver/pdf`
into a PDF content stream, and `driver/seen` into 3-D geometry — extruded faces
in a [seen](https://github.com/vibrantgio/seen) scene rather than pixels at all.
One parse, four targets.

Each driver is its own module, which is the point of the split: the parser is
the thing everything depends on, and it stays free of Gio, of rasterx, of a PDF
writer and of a scene graph. The root module's only dependencies are
`golang.org/x/image` and `golang.org/x/net`.

## Where it sits

Outside the tier table — a support library. The
[organization page](https://github.com/vibrantgio) has the full tier table.

Nothing here imports the design system except the three demos under
`driver/gio/example/`, which import `style` and `textdraw`.

```sh
go get github.com/vibrantgio/svg
go get github.com/vibrantgio/svg/driver/gio
```

Five modules, all Go 1.25.1: the root, plus `driver/gio`, `driver/pdf`,
`driver/raster` and `driver/seen`. `driver/gio` is on gioui.org v0.10.2.
Nested-module tags carry the directory as a prefix — `driver/gio/vX.Y.Z` —
and versions are per module, not lockstep with the root.

## Packages

| Package | |
| --- | --- |
| `svg` | The model, and nothing else — no parsing, no I/O, no constructor. `Icon` is a `ViewBox`, a `[]StyledPath` and a `matrix.Matrix2D`; `Icon.SetTarget(x, y, w, h)` fits it to a rectangle. `Path` is a `[]Operation` — `OpMoveTo`, `OpLineTo`, `OpQuadTo`, `OpCubicTo`, `OpClose`. `PathStyle` carries fill and line `Pattern`s (`PlainColor` or `Gradient`), widths, joins and dashes. |
| `svg/matrix` | `Matrix2D`, the 2-D affine transform: `Identity`, `Mult`, `Translate`, `Scale`, `Rotate`, `SkewX`, `SkewY`, `Invert`. |
| `svg/parser` | Document to icon. `NewParser(ErrorMode)` then `ParseStream(io.Reader)` or `ParseFile(name)`. `ErrorMode` is `IgnoreErrorMode` (the default), `WarnErrorMode` or `StrictErrorMode`. Shapes, transforms, gradients, `defs` and `use` are all reduced to styled paths here. |
| `svg/driver` | The backend contract. `DrawerNG` is what a driver implements; `FillAndStroker` is an optional extra for targets that can do both in one pass. `Draw(d DrawerNG, i *svg.Icon, opacity float64)` is the walker that drives any of them. |
| `svg/driver/dummy` | A `DrawerNG` that prints every call. The debugging backend, and the only driver that is not its own module. |
| `svg/driver/gio` | Gio. `NewDriver(ops *op.Ops)` for the op list, and `IconWidget(icon, width, height unit.Dp, opacity float64) layout.Widget` when you just want a component. |
| `svg/driver/raster` | `*image.RGBA`, via `srwiley/rasterx`. `NewDriver(w, h int, scanner rasterx.Scanner)`, or `RasterSVGIconToImage(io.Reader)` for the whole job. |
| `svg/driver/pdf` | A PDF content stream, via `benoitkugler/pdf`. `NewRenderer(cs)`, or `RenderSVGIconToPDF(icon io.Reader, pdfName string) error`. |
| `svg/driver/seen` | 3-D. `NewDrawer(options...)` collects the paths and `Object(kind)` hands back a `seen.Object` of extruded faces. Options: `Tolerance`, `Size`, `Depth`, `LayerOffset`, `Flat`. |

## Usage

Parse once, draw with a driver. The whole pipeline is two calls:

```go
icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseStream(f)
if err != nil {
	return nil, fmt.Errorf("parse %q: %w", url, err)
}
return giodriver.IconWidget(icon, 0, 0, 1), nil
```

`IconWidget` with a zero width and height renders at the icon's own view box
size, constrained down with the aspect preserved. That is the common case.

When you need the icon inside your own op list — composed with a clip, a
transform, or other paint operations — drive it yourself. `driver.Draw` takes
any `DrawerNG`, so changing one type changes the target:

```go
drv := giodriver.NewDriver(gtx.Ops)
icon.SetTarget(0, 0, float64(size.X), float64(size.Y))
driver.Draw(drv, icon, 1.0)
```

Parsing is not cheap and an icon is reusable once parsed, so parse at startup
and keep the `*svg.Icon`, for example from a `//go:embed`-ed document:

```go
parser := parser.NewParser(parser.WarnErrorMode)
component := vsvg.IconWidget(try(parser.ParseStream(bytes.NewBuffer(circles_svg))), 0, 0, 1.0)
```

With a document you control and can fix, use `WarnErrorMode`; with a document
that arrived from outside, use `IgnoreErrorMode`, so that one malformed file
does not take out the render.

## For coding assistants

Read the org guide before you write code against this module: the plan
root's [`AGENTS.md`](https://github.com/vibrantgio/organisation/blob/master/AGENTS.md).
The build and test commands are reached from that file.

## Status

The gaps in what ships.

- **The Gio driver ignores winding.** The parser keeps non-zero winding unless
  the document states `evenodd`, matching the SVG initial value, and
  `driver/raster`'s `TestFillRuleWinding` pins the three cases — `nonzero`,
  `evenodd` and unstated. The Gio driver drops the flag: its `SetWinding` has
  an empty body, because it paints through `clip.Outline`, which is non-zero
  only. The other three drivers honour the flag: `raster` forwards it to the
  rasterx scanner, `pdf` and `seen` store it. So an `evenodd` document renders
  wrong under Gio — not from the parser, but because the backend has no
  even-odd rule.
- **The Gio driver is the least complete of the four.** Radial gradients are
  stubbed to their first stop (a `TODO` in `paintGradient` says so). Stroke
  cap, join and dash are dropped — `Stroke` reads `LineWidth` and nothing
  else. Linear gradients that are not two stops with pad spreading fall back
  to rasterizing into an off-screen `image.RGBA` with a per-pixel Go loop,
  inside the draw path, on every draw.
- **Most packages have no tests.** Only `parser`, `driver/pdf`,
  `driver/raster` and `driver/seen` have any. Untested: the root model,
  `matrix`, `driver`, `driver/dummy` and `driver/gio` with its examples —
  `driver` being the `Draw` walker every backend depends on.
- **`driver/pdf`'s tests assert little.** They render the SVGs under
  `driver/testdata` and never read the output back; the repository's
  `.gitignore` says so: *"driver/pdf tests only write these, never read or
  compare them … Artifacts, not goldens."* There are no golden images in the
  repository and no `-update` flag.
- **A good deal of SVG is silently unsupported.** The element dispatch table
  knows `svg`, `g`, `line`, `stop`, `rect`, `circle`, `ellipse`, `polyline`,
  `polygon`, `path`, `desc`, `defs`, `title`, `linearGradient`,
  `radialGradient` and `use`. Everything else is an error escalated per
  `ErrorMode`, which under the default `IgnoreErrorMode` means dropped: `text`,
  `image`, `clipPath`, `mask`, `pattern`, `filter`, `symbol`, `marker`,
  `switch`, `<style>` and CSS classes. `url(...)` as a fill becomes opaque
  black in `ParseColor` whatever the error mode, and `rgba()`, `hsl()`,
  eight-digit hex and `currentColor` are not parsed at all.
- **`ErrZeroLengthID` can never be returned.** Both guards, in
  `linearGradientF` and `radialGradientF`, are `if len(id) >= 0` — always
  true — so a zero-length gradient id is accepted instead of rejected.
- **`driver/raster` declares `package driver`**, so its import path
  `…/svg/driver/raster` yields the package name `driver` and collides with
  `…/svg/driver`. Code importing both must alias one of them.
- **The SVGs under `driver/gio/example/testdata/` are referenced by no Go file
  in this repository.**
- **The fork's direct ancestor is credited only in prose.** `LICENSE` is
  srwiley's BSD-3-Clause, `Copyright (c) 2018, Steven R Wiley`. benoitkugler —
  the fork this code descends from — appears in this README and in no LICENSE
  or NOTICE file. `driver/testdata/LICENSE` separately carries the
  freepik/flaticon CC-BY-3.0 attribution for the test icons.
- **Several doc comments describe code that is not in this repository.**
  `Path.AddArc` and `AddRoundRect` are documented under unexported names
  (`addArc`, `addRoundRect`); `parser.Parser`'s comment says "See the `Draw`
  methods to use it" and `Parser` has no `Draw` method; `parser/doc.go`
  points readers at `oksvg/svgraster` and `oksvg/svgpdf`, which are not in
  this repository. All of it is visible in `go doc`.
