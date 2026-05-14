# Goink

React-style terminal UI primitives for Go, ported from the TypeScript
[Ink](https://github.com/vadimdemedes/ink) model.

The module path is currently:

```bash
go get github.com/dh-kam/ink-go
```

## Status

Goink has moved well past the original MVP. The current tree includes the
core component model, a flexbox-style layout renderer, runtime mounting,
input/focus hooks, mouse support, screen-reader rendering, snapshot utilities,
TUI transcript tooling, and upstream parity suites.

Current source-backed highlights:

- Virtual DOM and component helpers in `pkg/vdom` and `pkg/components`
- Layout, ANSI styling, wide-rune and grapheme-aware rendering
- Mounted runtime with managed stdout/stderr output, rerendering, cleanup, and
  throttled rendering
- Hooks for state, effects, memoization, refs, reducer, context, input, focus,
  mouse, transition, deferred values, app/stdin/stdout/stderr, cursor, and
  screen-reader state
- Interactive components including text input, select, multiselect, quick
  search, confirm, tabs, forms, form wizard, table, progress bar, spinner,
  error boundary, error overview, links, gradients, big text, syntax
  highlighting, and image rendering
- Reconciler and renderer test helpers for snapshots, fake stdin, and
  stdout/stderr frame capture
- 784 generated upstream Ink parity cases plus 22 project-derived parity cases
- 72 example applications and fixtures under `examples/`

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("Hello, Goink!", vdom.Props{"color": "cyan", "bold": true}),
		components.Text("Component trees render to terminal output."),
	)
}

func main() {
	fmt.Print(ink.Render(App))
}
```

For a managed interactive app:

```go
package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	count, setCount := ink.UseState(0)
	app := ink.UseApp()
	ink.UseInput(func(input string, key ink.InputKey) {
		if key.Return {
			setCount(count.(int) + 1)
		}
		if input == "q" || key.Escape {
			app.Exit()
		}
	})

	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Text(fmt.Sprintf("Count: %d", count)),
		components.Text("Press Enter to increment, q or Esc to quit.", vdom.Props{"dimColor": true}),
	)
}

func main() {
	instance, err := ink.Mount(App)
	if err != nil {
		panic(err)
	}
	if err := instance.WaitUntilExit(); err != nil {
		panic(err)
	}
}
```

## Examples

Representative examples:

```bash
go run ./examples/hello-world-demo
go run ./examples/border-demo
go run ./examples/select-input-demo
go run ./examples/chat-demo
go run ./examples/table-demo
go run ./examples/wizard-demo
go run ./examples/widgets-gallery
```

The example tree also contains parity fixtures for upstream Ink behavior such
as focus, input, static output, terminal resize, screen-reader output, wrapping,
overflow, absolute positioning, OSC 8 links, and IME cursor handling.

## Packages

- `pkg/ink`: public app/runtime API, managed sessions, hooks wrappers, output
  helpers, measurement, announcer, suspense, and render cache
- `pkg/components`: public component helpers and higher-level widgets
- `pkg/layout`: pure-Go flexbox-style layout calculations
- `pkg/styles`: ANSI color/style helpers and color parsing
- `pkg/input`: keyboard, SGR 1006 mouse, and X10 mouse parsing
- `pkg/focus`: focus manager primitives
- `pkg/context`: generic context provider/consumer support
- `pkg/reconciler`: vdom diff, patches, and tracker
- `pkg/renderer`: test renderer, snapshots, fake stdin, stdout/stderr capture
- `pkg/renderloop`: lower-level render loop utilities
- `internal/renderer`: layout/ANSI/screen-reader renderer implementation
- `internal/tuitest`: scenario runner, PTY transcript tools, terminal screen
  projection, and golden assertions
- `cmd/tui-transcript` and `cmd/tui-compare`: CLI tools for runtime parity
  transcript capture and comparison

## Testing

```bash
go test ./...
```

Focused parity suites:

```bash
go test ./tests -run 'TestUpstreamGoldenParity|TestUpstreamCoverageCounts' -count=1
go test ./tests -run 'TestProjectUpstreamGoldenParity|TestProjectUpstreamCoverageCounts' -count=1
go test ./tests/tui -count=1
```

Regenerate upstream goldens after editing `tests/upstream/cases.mjs`:

```bash
node tests/upstream/generate_goldens.mjs
```

Regenerate project-based goldens after editing `tests/project_upstream/cases.mjs`:

```bash
node tests/project_upstream/generate_goldens.mjs
```

## Parity Notes

The generated upstream parity skip list is currently empty. Remaining work is
mostly breadth and edge-case depth rather than missing core APIs: more
real-project runtime patterns, more terminal edge cases, more external
`ink-*` component compatibility targets, and a closer iterative Yoga
shrink-and-redistribute pass for directly-shrunk text-only sibling rows.

## License

MIT

Inspired by [Ink](https://github.com/vadimdemedes/ink) by Vadim Demedes.
