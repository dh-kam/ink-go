# Upstream Golden Tests

This directory contains a parity harness that compares the Go port against the original Ink implementation in `../ink`.

## Files

- `cases.mjs`: source of truth for the shared case definitions
- `cases.json`: generated case definitions consumed by Go tests
- `generate_goldens.mjs`: regenerates `cases.json` and renders those cases with upstream Ink
- `goldens.json`: generated expected outputs
- `../upstream_parity_test.go`: Go test that renders the same cases through the port

## Regenerate Goldens

1. Install upstream dependencies:

```bash
(cd ../ink && npm install)
```

2. Regenerate the golden outputs:

```bash
node tests/upstream/generate_goldens.mjs
```

## Run Parity Tests

```bash
go test ./tests -run 'TestUpstreamGoldenParity|TestUpstreamCoverageCounts' -count=1
```

## Notes

- Keep paths relative in this directory and related docs.
- Add new cases to `cases.mjs`, then regenerate the generated JSON files.
- Prefer upstream behaviors already covered by `../ink/test` when choosing new cases.
- The current suite generates 784 parity cases:
  - `Text`: 157
  - `Newline`: 33
  - `Spacer`: 34
  - `Transform`: 49
  - `Box`: 467
  - `Static`: 38
  - `Measure`: 2
  - `Render`: 4
- Cases may also set `mode: "error"` with `expectedError` to compare public runtime render failures, or `mode: "managed-frames"` with `frames`, `env`, and ordered substring expectations for managed multi-frame output.
- Recent additions cover more direct upstream accessibility/layout cases, including border draw order with negative margins, clipping inside bordered overflow containers, flex grow/shrink/basis and row/column `justifyContent` upstream aliases, combining-mark grapheme and ZWJ emoji width, screen-reader Static runtime deltas, Node-generated runtime `measureElement` basic and throttled goldens, Node-generated `render()` maxFps throttle plus TTY `bsu`/`esu` synchronized-update goldens, OSC 8 hyperlink preservation in plain and ANSI output, an explicit `textWrap: "truncate-end"` tagged case, `screen-reader` nested same-role wrappers, `Transform` screen-reader accessibility labels, empty-content role/state narration spacing, ordered multi-state narration, the upstream newline/wrapped-text padding and margin cases, the remaining `width-height.tsx` `minWidth="50%"` case, a direct `Static` plus screen-reader parent-role case, the remaining `borders.tsx` fit-content colorful multi-node border case, a direct alias for the `borders.tsx` variation-selector emoji fit-content round-border fixture, direct `undefined` / `null` / single-empty-text child cases from `text.tsx` and `components.tsx`, the direct `text with component` / `text with fragment` / `fragment` cases from `components.tsx`, the direct `text with variable` / `number` cases from `components.tsx`, direct aliases for `Transform` children/squashing fixtures plus the direct `Newline` and `Spacer` component fixtures from `components.tsx`, direct aliases for the upstream `text.tsx` color/background/inversion fixtures, direct aliases for the upstream `screen-reader.tsx` baseline plus aria-state/multiline/listbox/select-input cases, direct aliases for the upstream `components.tsx` basic text plus `width-height.tsx` width/min-height cases, a broader block of direct `borders.tsx` aliases for single-node, multi-node, and nested round-border fixtures, the remaining direct `components.tsx` text aliases for variable/component/fragment children, wrap behavior, truncation, and empty-text handling, plus exact-name aliases for the upstream `display.tsx`, `gap.tsx`, `flex-direction.tsx`, `flex-wrap.tsx`, `padding.tsx`, `margin.tsx`, `text-width.tsx`, `screen-reader.tsx`, and `width-height.tsx` fixture titles where the rendered output is already covered, along with the remaining parity-safe concurrent exact-title aliases from `components.tsx`, `text.tsx`, `width-height.tsx`, and `borders.tsx` where the final rendered output is identical, the exact-title alias for `gap - concurrent`, and the last failure/runtime titles now handled directly in the parity harness through `mode: "error"`, `mode: "managed-frames"`, runtime measurement modes, and runtime render modes.
- Cases may also set `screenReader: true` or `ansi: true` to compare those render modes against upstream debug output.
- The Go parity harness compares combined `Static` plus dynamic output, matching upstream debug render behavior more closely than a plain layout-only render.
