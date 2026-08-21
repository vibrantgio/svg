# AGENTS.md — svg

An SVG parser and renderer for Go, forked from `benoitkugler/oksvg`:
`parser` reads a document into an `svg.Icon` — a view box, a transform, and
styled paths built out of `Operation` values — and a driver, implementing
`driver.DrawerNG`, puts that icon somewhere. Each rendering driver is a
module of its own so that a caller taking one does not take the others'
dependencies; the Modules paragraph below lists which they are, measured
from the tree rather than typed here. `driver/gio` is the one a Gio caller
wants, and it adds `IconWidget`. `driver/dummy` is the exception that is
not a module: it logs every draw call and renders nothing, so it needs
nothing the parser does not already have.

**Layer.** Outside ADR-001's tier table: a support library, which the rule
binds in one direction only — every tier may import it, and it may import
nothing in the table itself. Its root module imports nothing else in the
organization. Its nested modules `svg/driver/gio` and `svg/driver/seen` add
`font`, `seen`, `seen/context/gio`, `style` and `textdraw` — those edges
are theirs and not the root module's. That direction is measured rather
than typed — `scripts/check-layers.sh --edges` reports the graph and
`scripts/sync-agents.sh` renders these sentences from it — so correcting
them here changes nothing. The other direction is measured too and
deliberately not written down: the gate checks the graph both ways, but a
public API's consumers are unknowable, so this file says what its module
needs and never who needs it.

**Read the canonical guide before you write code against this module.** It is
the organization's one agent guide — the module inventory with current tags,
the application skeleton, the MVU loop and rx semantics, typography, and the
pitfalls that are not guessable. It lives exactly once, in `vibrantgio/workbench` —
the repository that showcases building applications with Vibrant Gio —
and this file links it rather than copying it:

    https://raw.githubusercontent.com/vibrantgio/workbench/master/llms.txt

**Modules.** `github.com/vibrantgio/svg` at the repository root, and four
nested modules: `driver/gio/` (`github.com/vibrantgio/svg/driver/gio`),
`driver/pdf/` (`github.com/vibrantgio/svg/driver/pdf`), `driver/raster/`
(`github.com/vibrantgio/svg/driver/raster`), `driver/seen/`
(`github.com/vibrantgio/svg/driver/seen`). Nested-module tags carry the
directory as a prefix — `driver/gio/v0.0.9`, not `v0.0.9`.

**Build and test.** From the repository root, and again inside each nested
module directory — `./...` does not cross a module boundary:

    go build ./... && go test ./...
