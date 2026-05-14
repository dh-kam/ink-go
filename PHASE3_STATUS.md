# Phase 3 and Runtime Integration Status

Last updated: 2026-05-14

## Historical Scope

Phase 3 originally covered input handling, focus management, terminal
management, and live rendering. That scope is complete. The current source also
contains several post-Phase-3 additions: runtime sessions, managed output
restoration, screen-reader rendering, mouse dispatch, reconciler caching,
component widgets, and parity tooling.

## Completed Phase 3 Surface

### Input

- Keyboard parser for printable input, return, tab, shift-tab, escape,
  backspace/delete, arrows, home/end, page up/down, function keys, modifier
  cursor sequences, control letters, meta-prefixed variants, SS3 variants, and
  bracketed paste
- Hook-level input delivery through `UseInput`
- Mounted session input routing through `HandleInput` and stdin subscribers
- Cross-chunk bracketed paste buffering

Primary source areas:

- `pkg/input`
- `pkg/hooks/hooks.go`
- `pkg/ink/hooks_wrappers.go`
- `pkg/ink/session.go`

### Focus

- Focus manager primitives in `pkg/focus`
- Hook-level focus registration through `UseFocus` and `UseFocusOpts`
- Runtime focus controls through `UseFocusManager`
- Tab, shift-tab, escape blur, inactive targets, auto-focus replay, missing-id
  behavior, and registration-order handling

Primary source areas:

- `pkg/focus`
- `pkg/hooks/hooks.go`
- `pkg/ink/runtime_hooks.go`
- `pkg/ink/focus_runtime_test.go`

### Terminal and Managed Runtime

- Raw-mode reference counting for active input hooks
- Managed mount/unmount lifecycle with `Mount`, `MountWithOptions`,
  `RenderWithOptions`, `Rerender`, `Clear`, `Unmount`, and `WaitUntilExit`
- Cursor hide/show/reset handling, stdout/stderr restore cycles, resize
  handling, synchronized-update wrapping, and CI-aware behavior
- Static-output splitting and dynamic live-region rendering
- `MaxFPS` throttling and coalescing

Primary source areas:

- `pkg/terminal`
- `pkg/renderloop`
- `pkg/ink/session.go`
- `pkg/ink/managed_render.go`
- `pkg/ink/output_helpers.go`

### Mouse

- SGR 1006 mouse parsing and DECSET 1000/1006 mode toggles
- Legacy X10 mouse parsing
- Mounted app-scoped dispatch for `UseMouse`

Primary source areas:

- `pkg/input/mouse.go`
- `pkg/input/mouse_x10.go`
- `pkg/terminal/mouse.go`
- `pkg/hooks/mouse_hook.go`
- `pkg/ink/mouse_runtime.go`

## Post-Phase-3 Additions

- `UseEffect`, `UseMemo`, `UseCallback`, `UseRef`, `UseReducer`,
  `UseContext`, `UseTransition`, and `UseDeferredValue`
- `UseApp`, `UseStdin`, `UseStdout`, `UseStderr`, `UseCursor`,
  `UseIsScreenReaderEnabled`, and `UseAnnounce`
- DOM-like refs, `MeasureElement`, `MeasureElementPosition`, and `MeasureText`
- Screen-reader rendering with aria labels, states, roles, live regions, and
  static deltas
- `pkg/reconciler` diff/patch/tracker and `pkg/ink/render_cache.go`
- Component widgets: select, text input, confirm, multiselect, quick search,
  tabs, forms, form wizard, error overview, table, spinner, progress bar,
  alert, link, gradient, big text, syntax, image, and error boundary
- `cmd/tui-transcript`, `cmd/tui-compare`, and `internal/tuitest` for PTY-based
  runtime parity checks

## Validation

Relevant commands:

```bash
go test ./pkg/input ./pkg/focus ./pkg/terminal ./pkg/renderloop ./pkg/hooks ./pkg/ink
go test ./tests/tui -count=1
go test ./...
```

The TUI smoke suite currently covers upstream Node Ink vs Goink behavior for
interactive input, Ctrl+C shutdown, focus navigation, stdout/stderr restoration,
static output, tables, IME cursor behavior, and terminal resize sweeps.

## Remaining Runtime Risks

- Some host/PTY-specific timing behavior can still affect exact transcript
  captures.
- Directly-shrunk text-only sibling rows still lack the full iterative Yoga
  measure-and-redistribute pass.
- Real-project runtime coverage should continue expanding under
  `tests/project_upstream`.
