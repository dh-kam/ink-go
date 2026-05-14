# TUI Parity Tests

This area verifies runtime behavior by running the same scenario against multiple app bindings.

The core rule is:

- Scenario files describe what to do.
- Runtime manifests describe how to launch each implementation.
- Transcripts record what actually happened in a PTY.
- Comparisons should prefer `screen` mode, which projects raw PTY output into a terminal screen before comparing.
- Scenario `capture.mode: screen` uses the same screen projection for golden fixtures.
- Tmux preview is only for human observation and must not decide pass/fail.

Typical local flow:

```sh
mkdir -p .tmp/tui

go run ./cmd/tui-transcript \
  -scenario examples/box-backgrounds-demo/testdata/box-backgrounds.scenario.yaml \
  -manifest tests/tui/runtimes.yaml \
  -runtime upstream-ink-node > .tmp/tui/node.json

go run ./cmd/tui-transcript \
  -scenario examples/box-backgrounds-demo/testdata/box-backgrounds.scenario.yaml \
  -manifest tests/tui/runtimes.yaml \
  -runtime goink-go > .tmp/tui/go.json

go run ./cmd/tui-compare \
  -left .tmp/tui/node.json \
  -right .tmp/tui/go.json \
  -mode screen
```

The smoke tests skip the Node parity case when the sibling upstream checkout is unavailable. Keep scenario paths, manifest paths, and manifest `cwd` values relative so the repository can be moved without editing fixtures.

Input can be declared as text, a named key, or raw bytes with `input.hex`. Raw byte input is used for terminal cases where a PTY read can split a UTF-8 rune across multiple writes. If a step intentionally produces no new output, mark the expectation as `sameAsPrevious` so transcript capture does not wait for a frame that should not exist.

Current smoke coverage:

- `box-backgrounds-demo/static`: upstream Node Ink vs Go screen parity for ANSI-sensitive static output.
- `aria-demo/toggle`: upstream Node Ink vs Go screen parity for interactive input and Ctrl+C shutdown.
- `aria-demo/toggle`: Go-only Ctrl+C shutdown check, so local shutdown regressions are still caught when upstream Node Ink is unavailable.
- `chat-demo/messages`: upstream Node Ink vs Go screen parity for typed messages, message submission, and Ctrl+C shutdown.
- `cursor-ime-demo/korean-editing`: upstream Node Ink vs Go screen parity for split UTF-8 Korean input, wide-character cursor position, backspace, and Ctrl+C shutdown.
- `use-input-demo/max-then-quit`: upstream Node Ink vs Go plain parity for arrow-key movement and trailing blank row preservation on `q` exit.
- `select-input-demo/navigation`: upstream Node Ink vs Go screen parity for wrapped arrow-key navigation and Ctrl+C shutdown.
- `select-input-demo/wrap-down`: upstream Node Ink vs Go screen parity for 8th/9th down-arrow PTY screen projection fixtures and Ctrl+C shutdown.
- `use-focus-demo/navigation`: upstream Node Ink vs Go screen parity for Tab, Shift+Tab, Esc reset, wrap navigation, and Ctrl+C shutdown.
- `static-demo/complete`: upstream Node Ink vs Go screen parity for 10 completed static items and final dynamic count.
- `use-focus-with-id-demo/navigation`: upstream Node Ink vs Go screen parity for direct ID focus, Tab, Shift+Tab, Esc reset, wrap navigation, and Ctrl+C shutdown.
- `use-stdout-demo/two-writes`: upstream Node Ink vs Go screen parity for external stdout writes, restored dimensions UI, and Ctrl+C shutdown.
- `use-stderr-demo/two-writes`: upstream Node Ink vs Go screen parity for external stderr writes, restored UI, and Ctrl+C shutdown.
- `table-demo/static`: seeded upstream Node Ink vs Go screen parity for percentage-width table layout and static process exit.
- `terminal-resize-demo/resize-and-input`: upstream Node Ink vs Go screen parity for SIGWINCH resize wrapping, typed input, enter clear, and Ctrl+C shutdown.
- `terminal-resize-demo/width-sweep`: upstream Node Ink vs Go screen parity while resizing from 100 columns down to 20 columns in 5-column steps.
- `terminal-resize-demo/height-sweep`: upstream Node Ink vs Go screen parity while resizing from 20 rows down to 4 rows in 1-row steps.
