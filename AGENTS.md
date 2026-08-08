# AGENTS.md — svg

An SVG parser and renderer for Go, forked from `benoitkugler/oksvg`:
`parser` reads a document into an `svg.Icon` — a view box, a transform, and
styled paths built out of `Operation` values — and a driver puts that icon
somewhere. The four drivers are separate modules so their dependencies stay
out of the parser: `driver/gio` (a Gio `op.Ops` driver and `IconWidget`),
`driver/raster`, `driver/pdf` and `driver/seen`.

**Layer.** Outside ADR-001's tier table: a support library the design
system consumes and never depends on. prism's `icon` package holds an
`svg.Icon`, its `icon/gallery` renders one through `svg/parser` and
`svg/driver/gio`, and markdown's `svgimage` uses that same pair for inline
images. Nothing here imports the design system except the three demo
programs under `driver/gio/example/`, which use the tier-0 leaves `style`
and `textdraw`; `driver/seen` is the one place svg depends on another
support library.

**Read the canonical guide before you write code against this module.** It is
the organization's one agent guide — the module inventory with current tags,
the application skeleton, the MVU loop and rx semantics, typography, and the
pitfalls that are not guessable. It lives exactly once, in `vibrantgio/.github`,
and this file links it rather than copying it:

    https://raw.githubusercontent.com/vibrantgio/.github/master/llms.txt

**Modules.** `github.com/vibrantgio/svg` at the repository root, and four
nested modules: `driver/gio/` (`github.com/vibrantgio/svg/driver/gio`),
`driver/pdf/` (`github.com/vibrantgio/svg/driver/pdf`), `driver/raster/`
(`github.com/vibrantgio/svg/driver/raster`), `driver/seen/`
(`github.com/vibrantgio/svg/driver/seen`). Nested-module tags carry the
directory as a prefix — `driver/gio/v0.0.9`, not `v0.0.9`.

**Build and test.** From the repository root, and again inside each nested
module directory — `./...` does not cross a module boundary:

    go build ./... && go test ./...

**`driver/seen` does not build from a clean checkout**, and did not before
this file existed. Its `go.sum` pins `github.com/vibrantgio/seen/context/gio
v0.0.7` to a hash that no published form of that module produces, so the build
stops with a checksum mismatch before compiling anything. `workbench/launcher`
is stuck on the identical line.

Nothing local is missing and no push closes it: the tag on GitHub, the module
proxy and a `GOPROXY=direct` fetch all agree with one another and all disagree
with `go.sum`, which records content that was never published. Dropping the
two `seen/context/gio v0.0.7` lines and re-running `go mod tidy` restores the
build — `go mod tidy` on its own cannot, because it verifies before it
rewrites. Do that deliberately, in a change that says so, rather than as a
side effect of unrelated work.

The root module and the other three drivers are green.
