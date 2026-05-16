# Ink Demo Parity Checklist

This document tracks tmux side-by-side comparisons between upstream Ink demos
and Goink ports. Paths are intentionally relative to the repository/workspace
roots.

## Tracking Rules

- Upstream paths are relative to the upstream Ink repository.
- Go paths are relative to this repository.
- Confirmed rows mean the current tmux comparison workflow accepted the result.
- Legacy rows imported from `docs/compare-e2e.md` are preserved as backlog hints
  unless they are also present in the confirmed table.
- Interactive comparisons use `ink-node` for upstream Ink and `ink-demo` for
  Goink.
- Upstream golden captures should prefer direct Node execution over `npm run`
  when possible, because `npm run` adds wrapper headers and extra spacing that
  are not Ink program output.
- For remaining comparisons, exact Node/Go diff matches can be accepted without
  a user-facing tmux review; only mismatches or ambiguous visual/runtime output
  need side-by-side user confirmation.

## Confirmed Current Parity

| Done | Upstream demo | Go demo | Parity status | Compared behavior |
| --- | --- | --- | --- | --- |
| [x] | `examples/hello-world.tsx` | `examples/hello-world-demo` | Visual parity accepted | Program output matched when upstream was run directly with Node to avoid `npm run` wrapper output. |
| [x] | `examples/aria` | `examples/aria-demo` | Visual and interaction parity accepted | Initial render and spacebar checkbox toggle matched; screen-reader metadata is represented in props but the visual comparison uses terminal output. |
| [x] | `examples/borders` | `examples/border-demo` | Visual parity accepted | Border glyphs, padding, margin, and the full upstream border style matrix matched. |
| [x] | `examples/box-backgrounds` | `examples/box-backgrounds-demo` | Visual and terminal history parity accepted | Background fills, inherited background colors, ANSI-256 RGB/hex downgrade, fullscreen clear behavior, and final prompt placement matched. |
| [x] | `examples/chat` | `examples/chat-demo` | Visual and interaction parity accepted | Initial render, message submit flow, Backspace/Delete editing, functional state updates, and Ctrl+C exit matched after removing the duplicate manual input loop. |
| [x] | `examples/input-demo.tsx` | `examples/input-demo` | Visual and interaction parity accepted | Initial render and sampled key handling matched after replacing the legacy raw-terminal demo with an upstream-shaped Ink app. |
| [x] | `examples/colored-text.tsx` | `examples/colored-text` | Visual parity accepted | Text, blank-line structure, colors, and bold styling matched after replacing manual ANSI strings with upstream-shaped `Text` props. |
| [x] | `examples/justify-content` | `examples/justify-content-demo` | Visual parity accepted | Static layout and text alignment matched. |
| [x] | `examples/select-input` | `examples/select-input-demo` | Visual and interaction parity accepted | Initial render, cursor movement, color rendering, and Ctrl+C exit behavior were compared. |
| [x] | `examples/static` | `examples/static-demo` | Visual parity accepted | Static output accumulation and completion screen matched. |
| [x] | `examples/table` | `examples/table-demo` | Visual parity accepted | Table output matched; Go behavior was adjusted to terminate like upstream. |
| [x] | `examples/terminal-resize` | `examples/terminal-resize-demo` | Visual and resize parity accepted | Width and height resize captures matched after layout fixes. |
| [x] | `examples/use-input` | `examples/use-input-demo` | Visual and interaction parity accepted | Arrow movement, quit behavior, final blank-line difference, and scenario fixture were reviewed. |
| [x] | `examples/use-focus` | `examples/use-focus-demo` | Interaction parity accepted | Tab, Shift+Tab, Escape focus reset, and Ctrl+C exit were compared. |
| [x] | `examples/use-focus-with-id` | `examples/use-focus-with-id-demo` | Interaction parity accepted | Number-key focus selection and focused state matched. |
| [x] | `examples/use-stdout` | `examples/use-stdout-demo` | Runtime output parity accepted | Periodic stdout writes and terminal dimensions matched. |
| [x] | `examples/use-stderr` | `examples/use-stderr-demo` | Runtime output parity accepted | Periodic stderr writes and managed output restoration matched. |
| [x] | `examples/counter` | `examples/counter` | Behavior parity accepted | Counter updates matched; numeric values are timing-dependent. |
| [x] | `examples/render-throttle` | `examples/render-throttle-demo` | Behavior parity accepted | Throttled render behavior matched; counter values are timing-dependent. |
| [x] | `examples/suspense` | `examples/suspense-demo` | Visual parity accepted | Loading and final `Hello World` frame matched. |
| [x] | `examples/use-transition` | `examples/use-transition-demo` | Visual and hook behavior parity accepted | Initial render and typed search result render matched after `UseTransition` support was added. |
| [x] | `examples/concurrent-suspense` | `examples/concurrent-suspense-demo` | Runtime parity accepted | Progressive loading and final data frames matched after `Suspense` support was added. |
| [x] | `examples/incremental-rendering` | `examples/incremental-rendering-demo` | Approximate visual parity accepted | Layout and live updates looked similar; exact frame comparison is noisy due to random data, time, and fast repaint flicker. |
| [x] | `examples/cursor-ime` | `examples/cursor-ime-demo` | Visual and IME interaction parity accepted | Initial render, Korean/wide-character cursor placement, grapheme deletion, mixed input, and Ctrl+C exit matched after removing the duplicate manual input loop. |
| [x] | `examples/jest` | `examples/jest-demo` | Runtime parity accepted | Static completed tests, running tests, summary, and final `Ran all test suites.` matched; pass/fail and order are random. |
| [x] | `examples/live-counter.tsx` | `examples/live-counter` | Runtime parity accepted | Counter layout and 100ms automatic updates matched after replacing the legacy render loop with an upstream-shaped Ink hook app; count values are timing-dependent. |
| [x] | `examples/subprocess-output` | `examples/subprocess-output-demo` | Runtime parity accepted | Child process output was stripped to plain text and last five lines matched structurally; child test results are random. |
| [x] | `examples/test-aria.tsx` | `examples/test-aria` | Visual and screen-reader parity accepted | Default terminal render matched; with `INK_SCREEN_READER=true`, accessible output drops layout gaps, prefers `aria-label`, and matches Node Ink's no-final-newline behavior. |
| [x] | `examples/test-display.tsx` | `examples/test-display` | Visual and runtime parity accepted | Initial render, blue background box, `display: none/flex` one-second toggle, and persistent lower box matched after replacing the fixed four-frame Go loop with an upstream-shaped hook app. |
| [x] | `examples/test-exit.tsx` | `examples/test-exit` | Visual and lifecycle parity accepted | Initial render, red heading, automatic `useApp().exit()` after two seconds, and prompt restoration matched after async `UseApp().Exit()` was wired to wake the mounted session. |
| [x] | `examples/test-flex-direction.tsx` | `examples/test-flex-direction` | Visual parity accepted | Row, column, row-reverse, column-reverse ordering and spacing matched. |
| [x] | `examples/test-align.tsx` | `examples/test-align` | Visual parity accepted | `alignItems`, `alignSelf`, vertical placement, and full-terminal-width border sizing matched after fixing default stdout viewport detection. |
| [x] | `examples/test-dimensions.tsx` | `examples/test-dimensions` | Visual parity accepted | Fixed width/height, percentage width, minWidth, minHeight, colors, and full-terminal-width sizing matched after rounding fractional layout dimensions like Yoga. |
| [x] | `examples/test-padding.tsx` | `examples/test-padding` | Visual parity accepted | Padding, paddingX/Y, marginLeft/marginTop, background fills, and border plus padding plus margin layout matched. |
| [x] | `examples/test-overflow.tsx` | `examples/test-overflow` | Visual parity accepted | `overflow="hidden"` clipping and default visible overflow through a bordered container matched. |
| [x] | `examples/test-wrap.tsx` | `examples/test-wrap` | Visual parity accepted | Flex wrap line breaks, item margins, blue background fills, and width-40 border layout matched. |
| [x] | `examples/test-text-wrap.tsx` | `examples/test-text-wrap` | Visual parity accepted | Wrap, truncate-end, truncate-middle ellipsis placement, and box interior spacing matched. |
| [x] | `examples/test-static.tsx` | `examples/test-static` | Runtime parity accepted | Static item accumulation, green check styling, final dynamic `All tests complete!` frame, and automatic completion matched after replacing the manual loop with a hook runtime. |
| [x] | `examples/test-focus.tsx` | `examples/test-focus` | Visual and interaction parity accepted | Initial unfocused render, Tab focus cycling, Shift+Tab reverse focus, focused border/text color, and JSX trailing-space box width matched after reshaping the Go fixture to upstream behavior. |
| [x] | `examples/test-input-field.tsx` | `examples/test-textinput` | Visual and interaction parity accepted | Initial input field, typed text, Backspace/Delete editing, Enter no-op behavior, blue cursor marker, Escape exit, and final value display matched. |
| [x] | `examples/test-select.tsx` | `examples/test-select` | Visual and interaction parity accepted | Initial selected item, arrow-key movement, upward wrap-around, green selected-row styling, and `q` exit behavior matched. |
| [x] | `examples/test-table.tsx` | `examples/test-table` | Visual parity accepted | Padding, title color, bold header cells, horizontal separator, column widths, and data alignment matched after replacing the Go widget table with the upstream-shaped Box/Text fixture. |
| [x] | `examples/test-resize.tsx` | `examples/test-resize` | Visual and lifecycle parity accepted | Current terminal size, cyan title, border/padding, resize hint text, and one-shot process exit matched after changing the Go fixture to `RenderOnce()`. |
| [x] | `test/gap.tsx` | `examples/test-gap` | Visual parity accepted | Gap wrapping, column gap, row gap blank-line placement, ANSI colors, and final prompt placement matched against an upstream-shaped Node render. |
| [x] | `tests/upstream` absolute-position cases | `examples/test-absolute` | Visual parity accepted | Absolute children overlay earlier cells, do not consume flow space, and respect margin offsets like upstream Yoga output. |
| [x] | `test/components.tsx` OSC 8 link cases | `examples/test-link` | Visual and raw ANSI parity accepted | OSC 8 hyperlink open/close sequences are preserved through `Text`; terminal display shows the same visible label while ANSI capture retains the same hyperlink escapes. |
| [x] | `test/components.tsx` Transform cases | `examples/test-transform` | Visual parity accepted | Nested transforms, text-node squashing, multi-line transforms, and transforms embedded inside `Text` matched upstream terminal output. |
| [x] | `src/components/ErrorOverview.tsx` baseline | `examples/test-error` | Visual parity accepted | Stackless ErrorOverview output, red background label, message text, blank-line padding, and final prompt placement matched upstream. |
| [x] | `test/screen-reader.tsx` role/state cases | `examples/test-screen-reader` | Screen-reader parity accepted | List/listitem narration, selected/checked/disabled/expanded/readonly states, multiselectable listbox output, and final no-newline prompt placement matched upstream. |
| [x] | `test/components.tsx` Newline and Spacer cases | `examples/test-newline-spacer` | Visual parity accepted | Single and counted `Newline`, horizontal `Spacer`, vertical `Spacer`, colors, blank-line placement, and final prompt placement matched upstream. |
| [x] | `test/measure-element.tsx` measureElement cases | `examples/test-measure-element` | Runtime parity accepted | Initial zero-width render and post-layout `Width: 100` measurement matched using captured debug writes with a fixed 100-column stdout. |
| [x] | `test/text-width.tsx` wide character cases | `examples/test-text-width` | Visual and terminal parity accepted | Fixed-width emoji boxes matched upstream, including ZWJ emoji and combining mark behavior after moving the ANSI styled render path to grapheme-cluster width accounting. |
| [x] | `test/measure-text.tsx` measureText constructor case | `examples/test-measure-text` | Runtime parity accepted | `measureText("constructor")` returned width 11 in both upstream Ink and Goink. |
| [x] | `test/flex.tsx` flex grow/shrink/basis cases | `examples/test-flex` | Visual parity accepted | `flexGrow`, `flexShrink`, numeric `flexBasis`, percentage `flexBasis`, and row/column basis blank-line output matched upstream. |
| [x] | `test/margin.tsx` margin cases | `examples/test-margin` | Visual parity accepted | `margin`, `marginX`, `marginY`, per-edge margins, nested margins, multiline text, wrapped text, and concurrent equivalents matched upstream. |
| [x] | `test/background.tsx` background color cases | `examples/test-background` | Raw ANSI parity accepted | Background inheritance, explicit overrides, clearing inherited background, RGB/hex/ANSI256 background codes, wide characters, emoji, filled boxes, borders, padding, centering, and column layouts matched upstream after ANSI token normalization. |
| [x] | `test/flex-justify-content.tsx` justifyContent cases | `examples/test-justify-content` | Visual and raw ANSI parity accepted | Row and column `justifyContent` center/end/between/evenly/around behavior, colored squashed text, blank-line placement, and the upstream failing-marked actual `space-around` output matched upstream. |
| [x] | `test/flex-align-items.tsx`, `test/flex-align-self.tsx` alignment cases | `examples/test-flex-align` | Visual parity accepted | Row/column `alignItems` and `alignSelf` center/end behavior, multi-text squashing, indentation, and blank-line placement matched upstream. |
| [x] | `test/width-height.tsx` width/height cases | `examples/test-width-height` | Visual parity accepted | Fixed and percent width/height, minWidth/minHeight, height clipping, upstream failing-marked `minWidth="50%"` actual output, and concurrent render equivalents matched upstream. |
| [x] | `test/text.tsx` Text cases | `examples/test-text` | Raw ANSI parity accepted | Empty/null children, color/background variants, dim/bold/inverse styling, constructor text, text-node remeasure equivalents, and concurrent render equivalents matched upstream after SGR token normalization. |
| [x] | `test/flex-wrap.tsx` flexWrap cases | `examples/test-flex-wrap` | Visual parity accepted | Row/column `nowrap`, `wrap`, and `wrap-reverse` behavior, including leading blank lines and indentation, matched upstream after whitespace token normalization. |
| [x] | `test/overflow.tsx` overflow clipping cases | `examples/test-overflow-cases` | Visual parity accepted | `overflowX`, `overflowY`, combined `overflow`, border clipping, negative margin intersections, nested overflow, out-of-bounds border writes, and concurrent representative cases matched upstream after whitespace token normalization. |
| [x] | `test/padding.tsx` padding cases | `examples/test-padding-cases` | Visual parity accepted | `padding`, directional padding, nested padding, multiline padding, wrapped text padding, and concurrent representative cases matched upstream after whitespace token normalization. |
| [x] | `test/flex-direction.tsx` flexDirection cases | `examples/test-flex-direction-cases` | Visual parity accepted | Row, row-reverse, column, column-reverse, column text-node squashing, and concurrent representative cases matched upstream after whitespace token normalization. |
| [x] | `test/borders.tsx` border cases | `tests/upstream` generated border cases | Automated parity accepted | 98 generated border cases passed, covering round/full-width/fit-content/fixed-size boxes, wide characters, emoji, variation-selector emoji, nested borders, hidden sides, per-side colors and dim colors, custom arrow borders, overflow clipping, and update/concurrent cases. |
| [x] | `test/components.tsx` component cases | `tests/upstream` generated component cases | Automated parity accepted | 315 generated Text/Newline/Spacer/Transform/Static/Render cases passed, covering basic text, variables, fragments, runtime text nesting errors, wrap/truncate, remeasure, link ANSI closure, raw-mode/runtime branches, CI frame behavior, static deltas, transform squashing, newline, spacer, and concurrent equivalents. |
| [x] | `tests/upstream` generated golden suite | `tests/upstream_parity_test.go` | Automated parity accepted | `go test ./tests -run 'TestUpstreamGoldenParity|TestUpstreamCoverageCounts' -count=1` passed for 784 generated upstream cases: 467 Box, 157 Text, 49 Transform, 38 Static, 34 Spacer, 33 Newline, 4 Render, and 2 Measure cases. |
| [x] | real project Ink usage goldens | `tests/project_upstream` | Automated parity accepted | `go test ./tests -run 'TestProjectUpstreamGoldenParity|TestProjectUpstreamCoverageCounts' -count=1` passed for 46 curated project cases: 32 `gemini-cli`, 11 `neovate-code`, 1 `shopify-cli`, 1 `tweakcc`, and 1 `nanocoder` case. |
| [x] | `tests/tui` PTY smoke scenarios | `tests/tui` | Runtime parity accepted | `timeout 120s go test ./tests/tui -count=1 -v` passed for static output, aria toggle, chat messages, cursor IME editing, use-input, select input, focus navigation, stdout/stderr restoration, table, and terminal resize scenarios. |

## Pending Current Reconfirmation

| Done | Upstream demo | Go demo | Status |
| --- | --- | --- | --- |

## Legacy E2E Matrix Imported From `docs/compare-e2e.md`

The previous file tracked a broader feature checklist under older example names.
Rows below are merged for continuity, but are not automatically treated as
current confirmed parity when the example path or implementation target differs.

| Legacy done | Feature category | Upstream target | Legacy Go target | Current merge note |
| --- | --- | --- | --- | --- |
| [x] | Basic Rendering | `examples/hello-world.tsx` | `examples/hello` | Superseded by `examples/hello-world-demo`, which is now covered by the current confirmed table. |
| [x] | State and Timer | `examples/counter` | `examples/counter` | Already covered by the current confirmed table. |
| [x] | Borders | `examples/borders` | `examples/border-demo` | Covered by the current confirmed table. |
| [x] | Flexbox Layout | `examples/justify-content` | `examples/layout` | Already covered by `examples/justify-content-demo` in the current confirmed table. |
| [x] | Colors and Styles | `examples/colored-text.tsx` | `examples/colored-text` | Covered by the current confirmed table after the Go demo was reshaped to match upstream. |
| [x] | User Input | `examples/use-input` | `examples/input-demo` | Current confirmed parity is tracked under `examples/use-input-demo`; `examples/input-demo.tsx` is also covered by the current confirmed table. |
| [x] | Complex Demo | `examples/chat` | `examples/demo` | Covered by the current confirmed table with the newer `examples/chat-demo` target. |
| [x] | Live Updates | `examples/live-counter.tsx` | `examples/live-counter` | Covered by the current confirmed table after the Go demo was reshaped to match upstream. |
| [x] | Padding and Margin | `examples/test-padding.tsx` | `examples/test-padding` | Covered by the current confirmed table. |
| [x] | Flex Direction | `examples/test-flex-direction.tsx` | `examples/test-flex-direction` | Covered by the current confirmed table. |
| [x] | Align Items and Self | `examples/test-align.tsx` | `examples/test-align` | Covered by the current confirmed table after fixing default stdout viewport detection. |
| [x] | Width and Height | `examples/test-dimensions.tsx` | `examples/test-dimensions` | Covered by the current confirmed table after rounding fractional layout dimensions like Yoga. |
| [x] | Flex Wrap | `examples/test-wrap.tsx` | `examples/test-wrap` | Covered by the current confirmed table. |
| [x] | Overflow | `examples/test-overflow.tsx` | `examples/test-overflow` | Covered by the current confirmed table. |
| [x] | Text Wrap and Truncation | `examples/test-text-wrap.tsx` | `examples/test-text-wrap` | Covered by the current confirmed table. |
| [x] | Static Output | `examples/test-static.tsx` | `examples/test-static` | Covered by the current confirmed table after replacing the manual Go loop with an upstream-shaped hook runtime. |
| [x] | Cursor Helpers | `test/cursor.tsx`, `test/cursor-helpers.tsx` | `pkg/ink` cursor/session tests | Covered by upstream-shaped tests for cursor helper escape sequences, typed input cursor movement, trailing-space cursor-only updates, child unmount clearing, Suspense fallback cursor isolation, and stdout/stderr hook cursor restoration. |
| [x] | Display None | `examples/test-display.tsx` | `examples/test-display` | Covered by the current confirmed table after converting the Go example to an upstream-shaped hook runtime. |
| [x] | Manual Exit | `examples/test-exit.tsx` | `examples/test-exit` | Covered by the current confirmed table after fixing async `UseApp().Exit()` lifecycle wake-up. |
| [x] | Focus Management | `examples/test-focus.tsx` | `examples/test-focus` | Covered by the current confirmed table after reshaping the Go fixture to upstream behavior. |
| [x] | Screen Reader | `examples/test-aria.tsx` | `examples/test-aria` | Covered by the current confirmed table for both default visual output and `INK_SCREEN_READER=true` accessible output. |
| [x] | Terminal Resize | `examples/test-resize.tsx` | `examples/test-resize` | Covered by the current confirmed table after changing the Go fixture to one-shot `RenderOnce()` behavior. |
| [x] | Select Input | `examples/test-select.tsx` | `examples/test-select` | Covered by the current confirmed table. |
| [x] | Text Input | `examples/test-input-field.tsx` | `examples/test-textinput` | Covered by the current confirmed table after matching Enter no-op behavior. |
| [x] | Table | `examples/test-table.tsx` | `examples/test-table` | Covered by the current confirmed table after replacing the Go widget table with the upstream-shaped Box/Text fixture. |
| N/A | Spinner | `examples/test-spinner.tsx` | `pkg/components.Spinner` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Progress Bar | `examples/test-progress.tsx` | `pkg/components.ProgressBar` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| [x] | Hyperlinks | `test/components.tsx` OSC 8 link cases | `examples/test-link` | Covered by the current confirmed table using upstream OSC 8 preservation behavior. |
| N/A | Images | `examples/test-image.tsx` | `examples/syntax-image-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Gradients | `examples/test-gradient.tsx` | `examples/link-gradient-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Big Text | `examples/test-bigtext.tsx` | `examples/bigtext-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Syntax Highlighting | `examples/test-syntax.tsx` | `examples/syntax-image-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Confirm Dialog | `examples/test-confirm.tsx` | `examples/confirm-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Alert Dialog | `examples/test-alert.tsx` | `examples/alert-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Tabs | `examples/test-tabs.tsx` | `examples/tab-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Forms | `examples/test-form.tsx` | `examples/form-demo`, `examples/wizard-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| N/A | Multi Select | `examples/test-multiselect.tsx` | `examples/multiselect-demo` | No matching upstream Ink file exists in the current checkout; treat this as a Goink extension backlog item, not upstream parity. |
| [x] | Box Gap | `test/gap.tsx` | `examples/test-gap` | Covered by the current confirmed table using upstream `gap.tsx` cases rendered as a terminal comparison fixture. |
| [x] | Absolute Position | `tests/upstream` absolute-position cases | `examples/test-absolute` | Covered by the current confirmed table using upstream absolute-position parity cases rendered as a terminal comparison fixture. |
| N/A | Z-Index | `examples/test-zindex.tsx` | `examples/test-zindex` | No matching upstream Ink file exists in the current checkout; treat this as a stale legacy backlog item, not upstream parity. |
| N/A | Unmount Cleanup | `examples/test-cleanup.tsx` | `examples/test-cleanup` | No matching upstream Ink file exists in the current checkout; treat this as a stale legacy backlog item, not upstream parity. |
| [x] | Error Boundary | `src/components/ErrorOverview.tsx` baseline | `examples/test-error` | Covered by the current confirmed table for deterministic stackless ErrorOverview output; Go panic stack rendering remains language-specific. |
