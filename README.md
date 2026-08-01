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

So this fork replaces the rasterizer with an interface. `parser` reads a
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

Outside ADR-001's tier table — a support library the design system consumes and
never depends on. The [organization page](https://github.com/vibrantgio) has the
full tier table.

Two repositories in the design system use it, and both use the same three
pieces — the root model, `parser` and `driver/gio`.
[prism](https://github.com/vibrantgio/prism)'s `icon` package holds an
`*svg.Icon` and its `icon/gallery` renders one;
[markdown](https://github.com/vibrantgio/markdown)'s `svgimage` provides inline
SVG images to a rendered document. Nothing here imports the design system except
the three demos under `driver/gio/example/`, which reach for the tier-0 leaves
`style` and `textdraw`.

```sh
go get github.com/vibrantgio/svg
go get github.com/vibrantgio/svg/driver/gio
```

Five modules, all Go 1.25.1: the root, plus `driver/gio`, `driver/pdf`,
`driver/raster` and `driver/seen`. `driver/gio` is on gioui.org v0.10.1 like the
rest of the organization. Nested-module tags carry the directory as a prefix —
`driver/gio/v0.0.7`, not `v0.0.7` — and versions are per module, not lockstep
with the root. See Status.

## Packages

Five packages in the root module, one per driver module.

| Package | |
| --- | --- |
| `svg` | The model, and nothing else — no parsing, no I/O, no constructor. `Icon` is a `ViewBox`, a `[]StyledPath` and a `matrix.Matrix2D`; `Icon.SetTarget(x, y, w, h)` fits it to a rectangle. `Path` is a `[]Operation` — `OpMoveTo`, `OpLineTo`, `OpQuadTo`, `OpCubicTo`, `OpClose`. `PathStyle` carries fill and line `Pattern`s (`PlainColor` or `Gradient`), widths, joins and dashes. |
| `svg/matrix` | `Matrix2D`, the 2-D affine transform: `Identity`, `Mult`, `Translate`, `Scale`, `Rotate`, `SkewX`, `SkewY`, `Invert`. |
| `svg/parser` | Document to icon. `NewParser(ErrorMode)` then `ParseStream(io.Reader)` or `ParseFile(name)`. `ErrorMode` is `IgnoreErrorMode` (the default), `WarnErrorMode` or `StrictErrorMode`. Shapes, transforms, gradients, `defs` and `use` are all reduced to styled paths here. |
| `svg/driver` | The backend contract. `DrawerNG` is what a driver implements; `FillAndStroker` is an optional extra for targets that can do both in one pass. `Draw(d DrawerNG, i *svg.Icon, opacity float64)` is the walker every consumer calls. |
| `svg/driver/dummy` | A `DrawerNG` that prints every call. The debugging backend, and the only driver that is not its own module. |
| `svg/driver/gio` | Gio. `NewDriver(ops *op.Ops)` for the op list, and `IconWidget(icon, width, height unit.Dp, opacity float64) layout.Widget` when you just want a widget. This is the driver the design system uses. |
| `svg/driver/raster` | `*image.RGBA`, via `srwiley/rasterx`. `NewDriver(w, h int, scanner rasterx.Scanner)`, or `RasterSVGIconToImage(io.Reader)` for the whole job. |
| `svg/driver/pdf` | A PDF content stream, via `benoitkugler/pdf`. `NewRenderer(cs)`, or `RenderSVGIconToPDF(icon io.Reader, pdfName string) error`. |
| `svg/driver/seen` | 3-D. `NewDrawer(options...)` collects the paths and `Object(kind)` hands back a `seen.Object` of extruded faces. Options: `Tolerance`, `Size`, `Depth`, `LayerOffset`, `Flat`. |

## Usage

Parse once, draw with a driver. The shortest real consumer is markdown's inline
image provider, `markdown/svgimage/svgimage.go:71`, which is the whole pipeline
in two calls:

```go
icon, err := parser.NewParser(parser.IgnoreErrorMode).ParseStream(f)
if err != nil {
	return nil, fmt.Errorf("svgimage: parse %q: %w", url, err)
}
return giodriver.IconWidget(icon, 0, 0, 1), nil
```

`IconWidget` with a zero width and height renders at the icon's own view box
size, constrained down with the aspect preserved. That is the ninety-percent
case and there is no reason to reach past it.

When you need the icon inside your own op list — composed with a clip, a
transform, or other paint operations — drive it yourself. `driver.Draw` takes
any `DrawerNG`, so changing one type changes the target:

```go
drv := giodriver.NewDriver(gtx.Ops)
icon.SetTarget(0, 0, float64(size.X), float64(size.Y))
driver.Draw(drv, icon, 1.0)
```

Parsing is not cheap and an icon is reusable once parsed, so parse at startup
and keep the `*svg.Icon`. `driver/gio/example/circles/main.go:38` does exactly
that against a `//go:embed`-ed document:

```go
parser := parser.NewParser(parser.WarnErrorMode)
widget := vsvg.IconWidget(try(parser.ParseStream(bytes.NewBuffer(circles_svg))), 0, 0, 1.0)
```

Note the `WarnErrorMode` there rather than the `IgnoreErrorMode` markdown uses.
With a document you control and can fix, warn; with a document that arrived from
outside, ignore, so that one malformed file does not take out the render.

## For coding assistants

Read the canonical guide before writing code against this module — the module
inventory with current tags, the application skeleton, MVU and rx semantics,
typography, and the pitfalls that are not guessable:

<https://raw.githubusercontent.com/vibrantgio/.github/master/llms.txt>

[`AGENTS.md`](./AGENTS.md) in this repository has the build and test commands.

## Status

Honest about what does not work yet. Every count below is measured.

- **`fill-rule` is parsed backwards.** `parser/svgcursor.go:133` reads

      curStyle.UseNonZeroWinding = strings.EqualFold(v, "evenodd")

  which sets non-zero winding when the document asks for `evenodd`, and even-odd
  when it asks for `nonzero`. `svg.DefaultStyle.UseNonZeroWinding` is `true` and
  correct, so only documents that *state* a fill rule are affected — but those
  are exactly the documents that state it because it matters. A self-intersecting
  or hole-punched path with an explicit `fill-rule` fills wrong.
- **The Gio driver ignores winding entirely**, which is why nothing in the
  design system has noticed the bug above. `driver/gio/driver.go:59` is
  `func (d *Driver) SetWinding(useNonZeroWinding bool) {}` — an empty body —
  because it paints through `clip.Outline`, which is non-zero only. The other
  three drivers do honour the flag: `raster` forwards it to the rasterx scanner,
  `pdf` and `seen` store it. So the inverted parse is invisible under Gio and
  live everywhere else.
- **`driver/seen` does not build.** `go build ./...` fails with
  `verifying github.com/vibrantgio/seen/context/gio@v0.0.7: checksum mismatch`
  on a stale `go.sum` pin. The other four modules build and vet clean. Known and
  already recorded; not this module's defect to fix.
- **The driver modules are pinned to a stale root.** The root is tagged
  `v0.0.8`, but `driver/gio`, `driver/pdf` and `driver/raster` all still
  `require github.com/vibrantgio/svg v0.0.7`; only `driver/seen` is on `v0.0.8`.
  A fix to the parser does not reach the drivers until each `go.mod` is bumped,
  and nothing enforces that.
- **The Gio driver is the least complete of the four, and it is the one
  everything uses.** Radial gradients are stubbed to their first stop
  (`driver.go:164`, with the `TODO` to prove it). Stroke cap, join and dash are
  dropped — `Stroke` reads `LineWidth` and nothing else. Linear gradients that
  are not two stops with pad spreading fall back to rasterizing into an
  off-screen `image.RGBA` with a per-pixel Go loop, inside the draw path, every
  frame.
- **Nine of eleven packages have no tests at all.** Only `parser` (21.7%
  coverage, two tests) and `driver/pdf` have any. Untested: the root model,
  `matrix`, `driver/dummy`, `driver/raster`, all four packages of `driver/gio` —
  and `driver` itself, the `Draw` walker every backend depends on and every
  consumer calls.
- **`driver/pdf`'s 97.4% coverage is not what it looks like.** The test renders
  all seventy-four SVGs under `driver/testdata` and never reads the output back.
  The repository says so in its own `.gitignore`: *"driver/pdf tests only write
  these, never read or compare them … Artifacts, not goldens."* It is near-total
  execution with near-zero assertion. There are no golden images anywhere in the
  repository and no `-update` flag; `driver/seen`'s render assertion is that the
  emitted markup is at least 2 000 bytes long.
- **A good deal of SVG is silently unsupported.** The element dispatch table has
  sixteen keys — `svg`, `g`, `line`, `stop`, `rect`, `circle`, `ellipse`,
  `polyline`, `polygon`, `path`, `desc`, `defs`, `title`, `linearGradient`,
  `radialGradient` and `use`. Everything else is an error escalated per
  `ErrorMode`, which under the default `IgnoreErrorMode` means dropped: `text`,
  `image`, `clipPath`, `mask`, `pattern`, `filter`, `symbol`, `marker`,
  `switch`, `<style>` and CSS classes. `url(...)` as a fill silently becomes
  opaque black (`parser/parsecolor.go:44`) whatever the error mode, and
  `rgba()`, `hsl()`, eight-digit hex and `currentColor` are not parsed at all.
- **`ErrZeroLengthID` can never be returned.** Both guards, at
  `parser/elementfuncs.go:247` and `:303`, are `if len(id) >= 0` — always true —
  so a zero-length gradient id is accepted instead of rejected.
- **`driver/raster` declares `package driver`**, so its import path
  `…/svg/driver/raster` yields the package name `driver` and collides with
  `…/svg/driver`. Any consumer of both must alias one of them.
- **Four of the five drivers have no consumer anywhere in the organization** —
  `driver/raster`, `driver/pdf`, `driver/dummy` and `driver/seen`, the last of
  which also cannot be built. The twelve SVGs under
  `driver/gio/example/testdata/` are referenced by no Go file in the org.
- **The fork's direct ancestor is credited only in prose.** `LICENSE` is
  srwiley's BSD-3-Clause, `Copyright (c) 2018, Steven R Wiley`, carried verbatim
  and unmodified since the first commit. benoitkugler — the fork this code
  actually descends from — appears in this README and in `AGENTS.md` and in no
  LICENSE or NOTICE file. `driver/testdata/LICENSE` separately carries the
  freepik/flaticon CC-BY-3.0 attribution for the test icons.
- **Several doc comments still describe the code they were copied from.**
  `Path.AddArc` and `AddRoundRect` are documented under their old unexported
  names; `parser.Parser`'s comment says "See the 'Draw' methods to use it" and
  `Parser` has no `Draw` method; `parser/doc.go` points readers at
  `oksvg/svgraster` and `oksvg/svgpdf`, which do not exist in this repository.
  All of it is visible in `go doc`.
