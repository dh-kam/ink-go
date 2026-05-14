# Goink Current Status Report

Last updated: 2026-05-14

## Summary

Goink is no longer only a Phase 1/2 experiment. The current source implements a
broad React/Ink-style terminal UI runtime in Go, with core rendering, managed
sessions, hooks, input/focus/mouse handling, screen-reader support, a large
component set, and multiple parity harnesses.

The historical Phase 1/2/3 milestones are complete. Current work is centered on
upstream parity breadth, real-project fixture coverage, terminal edge cases, and
component ecosystem compatibility.

## Source-Backed Feature Set

### Runtime and Rendering

- `pkg/ink` exposes one-shot rendering, app instances, mounted sessions,
  managed render instances, app/stdin/stdout/stderr/cursor hooks, measurement,
  announcer support, suspense helpers, and render-section caching.
- `internal/renderer` handles flex layout output, ANSI output, screen-reader
  output, background fill, border draw order, overflow clipping, cell-level
  diffs, wide-rune continuation cells, and grapheme-cluster-safe text layout.
- `pkg/layout` implements the pure-Go layout model used by the renderer.
- `pkg/reconciler` provides diff/patch operations and a tracker for identical
  tree short-circuiting.

### Public Components

The component package contains:

- `Box`, `Text`, `Newline`, `Space`, `Spacer`, `Static`, `StaticItems`,
  `Transform`, `Border`
- `TextInput`, `PasswordInput`, `Select`, `MultiSelect`, `QuickSearch`,
  `Confirm`, `Tabs`
- `Spinner`, `ProgressBar`, `Table`, `Divider`, `Alert`
- `Link`, `Gradient`, `BigText`, `Syntax`, `Image`
- `ErrorBoundary`, `ErrorOverview`, `ErrorOverviewGroup`
- `Form`, `FormState`, `FormWizard`

`BigText` includes block, tiny, shadow, outline, slim, and digital fonts.
`Syntax` includes Go, JSON, YAML, Markdown, Bash, Python, Rust, SQL,
JavaScript, and diff.

### Hooks and Input

The hook surface includes `UseState`, `UseEffect`, `UseMemo`, `UseCallback`,
`UseRef`, `UseReducer`, `UseContext`, `UseInput`, `UseFocus`, `UseFocusOpts`,
`UseFocusManager`, `UseMouse`, `UseTransition`, `UseDeferredValue`, `UseApp`,
`UseStdin`, `UseStdout`, `UseStderr`, `UseCursor`,
`UseIsScreenReaderEnabled`, and `UseAnnounce`.

The input layer parses keyboard escape sequences, modifier variants, bracketed
paste, function keys, control keys, SGR 1006 mouse events, and legacy X10 mouse
frames.

## Verification Surface

- Full repository test command: `go test ./...`
- Generated upstream Ink parity: 784 cases in `tests/upstream`
- Project-derived parity: 22 cases in `tests/project_upstream`
- PTY runtime parity smoke tests in `tests/tui`
- Example-level golden/scenario tests under `examples/*/testdata`
- Snapshot and fake-stdin utilities in `pkg/renderer`

## What Changed Since The Old Phase 2 Report

The old Phase 2 final report only described colors, layout, and basic
components. The current source now includes:

- Runtime mounting and managed render lifecycle
- Input, focus, mouse, cursor, stdout/stderr, app, context, reducer, effect,
  memo, callback, ref, transition, and deferred-value hooks
- Screen-reader rendering and aria/live-region support
- Reconciler, render cache, incremental/cell-level output paths
- TUI transcript tooling and manifest-based upstream comparisons
- Full widget set for forms, lists, tables, progress, error display, links,
  gradients, big text, syntax highlighting, and images

## Remaining Work

- Implement the full iterative Yoga shrink-and-redistribute behavior for
  directly-shrunk text-only sibling rows.
- Expand `tests/project_upstream` with more real public Ink application states.
- Add more terminal edge cases around uncommon escape sequences, resize timing,
  and host-specific PTY behavior.
- Continue matching external `ink-*` component packages where local equivalents
  exist.

## Current Verdict

The framework is usable for substantial terminal UI experiments and has a deep
test/parity harness. It should still be treated as an active porting workspace,
not a final upstream-compatible replacement.
