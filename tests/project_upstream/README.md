# Project Upstream Golden Tests

This directory contains project-based parity cases curated from real public Ink applications and compared against the original Ink implementation in `../ink`.

## Files

- `helpers.mjs`: small DSL helpers for building normalized case trees
- `cases.mjs`: source of truth for curated project-based case definitions
- `cases.json`: generated case definitions consumed by Go tests
- `generate_goldens.mjs`: regenerates `cases.json` and renders those cases with upstream Ink
- `goldens.json`: generated expected outputs
- `../project_upstream_parity_test.go`: Go test that renders the same cases through the port

## Regenerate Goldens

1. Install upstream dependencies:

```bash
(cd ../ink && npm install)
```

2. Regenerate the golden outputs:

```bash
node tests/project_upstream/generate_goldens.mjs
```

## Run Parity Tests

```bash
go test ./tests -run 'TestProjectUpstreamGoldenParity|TestProjectUpstreamCoverageCounts' -count=1
```

## Notes

- Keep paths relative in this directory and related docs.
- Add new cases to `cases.mjs`, then regenerate the generated JSON files.
- These cases are curated from real public Ink applications, not upstream Ink fixture titles.
- The current generated suite covers:
  - `gemini-cli`: 32
  - `neovate-code`: 11
  - `shopify-cli`: 1
  - `tweakcc`: 1
  - `nanocoder`: 1
- Each case should stay traceable to a real project source location through the research notes under `research/project-goldens/`.
- Cases may also set `mode: "error"` with `expectedError` to compare public runtime render failures, or `mode: "managed-frames"` with `frames`, `env`, and ordered substring expectations for managed multi-frame output.
- Cases may also set `screenReader: true` or `ansi: true` to compare those render modes against upstream debug output.
