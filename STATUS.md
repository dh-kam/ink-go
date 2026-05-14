# Goink Development Status

Last updated: 2026-05-14

## Current State

The repository is an active Go port of TypeScript Ink. The source tree now
contains 95 Go packages from `go list ./...`, 72 example directories, generated
upstream parity fixtures, project-derived parity fixtures, and PTY-based TUI
smoke tests.

`go test ./...` passed on the current source before this documentation update.

## Implemented Surface

### Core Runtime

- Virtual DOM node model and element/text helpers (`pkg/vdom`)
- Public app API (`pkg/ink`) with `Render`, `RenderToString`, `NewApp`,
  `Mount`, `MountWithOptions`, `RenderWithOptions`, `Rerender`, `Clear`,
  `Unmount`, and `WaitUntilExit`
- Managed stdout/stderr output restoration, incremental rendering, throttled
  `MaxFPS` rendering, resize handling, static output splitting, cursor updates,
  and CI-aware output behavior
- Screen-reader render mode with aria labels, roles, state narration, hidden
  subtrees, live announcements, and static-output deltas
- DOM-like refs and measurement via `MeasureElement`, `MeasureElementPosition`,
  and `MeasureText`

### Components

Implemented component helpers include:

- Core layout/text: `Box`, `Text`, `Newline`, `Space`, `Spacer`, `Static`,
  `StaticItems`, `Transform`, `Border`
- Input and selection: `TextInput`, `PasswordInput`, `Select`, `MultiSelect`,
  `QuickSearch`, `Confirm`, `Tabs`
- Display widgets: `Spinner`, `ProgressBar`, `Table`, `Divider`, `Alert`,
  `Link`, `Gradient`, `BigText`, `Syntax`, `Image`
- Error and form flows: `ErrorBoundary`, `ErrorOverview`,
  `ErrorOverviewGroup`, `Form`, `FormState`, `FormWizard`

`BigText` currently includes block, tiny, shadow, outline, slim, and digital
fonts. `Syntax` currently includes Go, JSON, YAML, Markdown, Bash, Python, Rust,
SQL, JavaScript, and diff tokenizers.

### Hooks

Implemented hook surface includes:

- State and lifecycle: `UseState`, `UseEffect`, `UseMemo`, `UseCallback`,
  `UseRef`, `UseReducer`, `UseContext`
- Runtime: `UseApp`, `UseStdin`, `UseStdout`, `UseStderr`, `UseCursor`,
  `UseIsScreenReaderEnabled`, `UseAnnounce`
- Input and navigation: `UseInput`, `UseFocus`, `UseFocusOpts`,
  `UseFocusManager`, `UseMouse`
- Concurrent-style helpers: `UseTransition`, `UseDeferredValue`

### Layout, Rendering, and Input

- Pure-Go flexbox-style layout engine with grow/shrink, basis, min/max
  dimensions, gaps, wrapping, reverse directions, absolute positioning,
  clipping, margins, padding, and border-aware geometry
- ANSI renderer with foreground/background colors, modifiers, nested style
  inheritance, OSC 8 hyperlink preservation, full-area background fills, and
  per-side border colors/dim props
- Grapheme-cluster and wide-rune handling for wrapping, truncation, fixed-width
  boxes, emoji ZWJ sequences, combining marks, and variation selectors
- Keyboard parser for upstream-style escape variants, modifiers, bracketed
  paste, function keys, cursor keys, control keys, and paste chunks
- SGR 1006 and legacy X10 mouse parsing, terminal mouse mode toggles, and
  mounted app-scoped mouse dispatch

## Test and Parity Coverage

- `tests/upstream`: 784 generated cases from upstream Ink fixtures
  - Box: 467
  - Text: 157
  - Transform: 49
  - Static: 38
  - Newline: 33
  - Spacer: 34
  - Measure: 2
  - Render: 4
- `tests/project_upstream`: 22 project-derived cases
  - gemini-cli: 8
  - neovate-code: 11
  - shopify-cli: 1
  - tweakcc: 1
  - nanocoder: 1
- `tests/tui`: PTY transcript smoke tests for upstream Node Ink vs Goink
  scenarios such as background rendering, aria toggle, chat, IME input,
  select input, focus navigation, static output, stdout/stderr restoration,
  tables, and terminal resize sweeps
- `pkg/renderer`: snapshot and output-capture helpers for app-level tests
- `internal/tuitest`: scenario runner, transcript recorder, terminal screen
  projection, and manifest-based runtime launcher

## Remaining Risks

- Directly-shrunk text-only sibling rows do not yet implement the full
  iterative Yoga measure-and-redistribute pass.
- Real-project parity should keep expanding beyond the current curated
  `tests/project_upstream` cases.
- External `ink-*` ecosystem component compatibility is represented by local
  component ports and examples, but not exhaustively covered by generated
  upstream fixtures.

## Useful Commands

```bash
go test ./...
go test ./tests -run 'TestUpstreamGoldenParity|TestUpstreamCoverageCounts' -count=1
go test ./tests -run 'TestProjectUpstreamGoldenParity|TestProjectUpstreamCoverageCounts' -count=1
go test ./tests/tui -count=1
```
