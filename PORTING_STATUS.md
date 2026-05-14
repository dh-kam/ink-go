# Goink Porting Status

Reference source: `../ink`

This repository is the active Go port workspace for the TypeScript Ink project.

## Current Baseline

The codebase was imported from the local Ink reference workspace under `../ink` on 2026-04-18.

Verified on 2026-05-14:

- `go test ./...` passes
- local reference source under `../ink` is readable

## Ported Surface

### Components

- `Text`
- `Box`
- `Newline`
- `Static`
- `Spacer` (basic Ink-compatible form via `flexGrow`)
- `Transform` (basic text-output transform support)
- `Border` (single, double, rounded, bold)
- `TextInput` / `PasswordInput` (controlled, with `TextInputState` controller)
- `ProgressBar`
- `Spinner`
- `Table`
- `Select` + `SelectState` (controlled list with windowed scrolling)
- `Divider` (horizontal rule with optional centered title)
- `Alert` (info / success / warning / error variants with colored icons)
- `Confirm` (yes/no prompt with controlled state and configurable accept/reject keys)
- `MultiSelect` (multi-pick list with windowed scrolling and toggle controller)
- `Tabs` + `TabsState` (focusable tab strip with `Next` / `Prev` / `SetActive` controller)
- `QuickSearch` (filter-as-you-type list with fuzzy matching and selection callback)
- `Link` (OSC 8 hyperlink wrapper around inline `Text`)
- `Gradient` (per-character RGB interpolation across one or more stops)
- `BigText` (figlet-style headline with `Block`, `Tiny`, `Shadow`, `Outline`, `Slim`, and `Digital` built-in fonts)
- `Syntax` (inline source highlighter with `Go`, `JSON`, `YAML`, `Markdown`, `Bash`, `Python`, `Rust`, `SQL`, `JavaScript`, and `Diff` tokenizers)
- `Image` (ANSI half-block image rendering for raw RGBA data)
- `ErrorBoundary` (panic-recovery wrapper with default red bordered fallback and `OnError` hook)
- `ErrorOverview` / `ErrorOverviewGroup` (grouped validation/runtime errors with stack/source context)
- `Form` (multi-field controller with per-field validation, focus traversal, and submit/cancel callbacks)
- `FormWizard` (multi-step form controller with validation gates and aggregated errors)

### Hooks / State

- `UseState`
- `UseInput`
- `UseFocus`
- `UseEffect`
- `UseMemo`
- `UseCallback`
- `UseRef`
- `UseReducer` (generic Redux-style state, stable dispatch identity)
- `UseContext` (generic provider/consumer via `pkg/context`)
- `UseMouse` (SGR 1006 mouse events with multi-subscriber fan-out)
- `UseTransition`
- `UseDeferredValue`
- `UseApp`
- `UseStdin`
- `UseStdout`
- `UseStderr`
- `UseFocusManager`
- `UseCursor`
- `UseIsScreenReaderEnabled`
- `UseAnnounce`
- `MeasureElement`
- `MeasureElementPosition`
- `MeasureText`
- `DOMElement` (current minimal Go-level ref handle)

### Runtime / Infra

- virtual DOM
- reconciler — `pkg/reconciler.Diff` + `Patch`/`ApplyAll` with LIS-based keyed children diff and a `Tracker` that short-circuits identical-tree renders
- simple renderer
- flex layout engine
- input handling — keyboard plus SGR 1006 and legacy X10 mouse routing wired into the mounted session input loop, with `routeMouseInput` delivering parsed `MouseEvent`s to subscribers scoped to the mounted app instance
- terminal mouse mode (DECSET 1000/1006 enable/disable)
- legacy X10 mouse parser (`pkg/input.ParseX10Mouse` / `IsX10MouseSequence`) for older terminals that only emit `\x1b[M` six-byte frames
- focus manager
- render loop
- testing utilities — `pkg/renderer.Render` / `Instance` capture every frame, `MatchSnapshot` golden-file comparison with `UPDATE_SNAPSHOTS=1`, optional fake stdin via `WithStdin()` exposing `Instance.Stdin()` + `SubscribeInput(...)` for driving `UseInput` / `UseMouse` without a real TTY, plus opt-in `WithStdoutCapture()` / `WithStderrCapture()` writers exposing `StdoutFrames()` / `StderrFrames()` for ink-testing-library-style out-of-band output assertions

## Compatibility Notes

The port is not a drop-in replacement for TypeScript Ink yet.

Implemented recently:

- text-like nodes now occupy layout width correctly
- `Spacer` now works in row/column layouts through `flexGrow`
- `Transform` can modify rendered text output in the renderer
- upstream parity harness compares this port against goldens generated from `../ink`
- upstream parity coverage is now 784 cases total, including 467 `Box` cases, 157 `Text` cases, 49 `Transform` cases, 38 `Static` cases, 33 `Newline` cases, 34 `Spacer` cases, 2 runtime `Measure` cases, and 4 runtime `Render` cases
- component helpers now expose richer Ink-like builders for nested `Text`, repeated `Newline`, and item-rendered `Static`
- `Box` parity now also covers `margin`, `alignItems`, and `justifyContent="space-evenly"`
- `Box` parity now also covers `display="none"`, `row-reverse` / `column-reverse`, `justifyContent="space-around"`, and multi-text `alignSelf` row cases
- `Box` parity now also covers `gap`, row/column `flexWrap="wrap"`, row/column `flexWrap="wrap-reverse"`, and the upstream non-wrap overflow behavior for column flex containers
- `Box` parity now also supports axis-specific `rowGap` / `columnGap`, including distinct main-axis and wrapped-line spacing
- `Box` parity now also covers fixed and percentage `width` / `height`, percentage `minWidth`, numeric `minWidth` / `minHeight`, and height-based text clipping from upstream `width-height.tsx`
- plain buffer rendering now tracks wide-rune continuation cells, so fixed-width boxes no longer insert extra spaces before adjacent siblings when rendering emoji or other double-width glyphs
- plain and ANSI renderers now also preserve zero-width variation selectors and combining marks on the previous visible cell, which fixes fit-content border layout for emoji sequences such as `🌡️⚠️✅`
- upstream golden parity now also covers screen-reader mode for `aria-label`, `aria-hidden`, `aria-role`, accessibility state narration, nested row/column structure, and listbox/listitem option output
- upstream golden parity now also covers ANSI debug output for background inheritance, background override and clear transitions, full-area background fill, border color and dim props, and the child-style-preserving `borderDimColor` regression from upstream `borders.tsx`
- upstream golden parity now also covers more cases from `components.tsx`, including inline count text, ignored empty sibling text in column layout, transform squashing with split text children, and static output followed by margin-separated dynamic content
- upstream golden parity now also covers the remaining static `text.tsx` style formats, including ANSI standard/hex/rgb/ansi256 foreground and background colors, `dimColor + bold`, dimmed colored text, and inverse text rendering
- the Go parity harness now combines rendered static and dynamic sections when comparing against upstream debug output, which matches upstream `Static` behavior more closely than comparing only the plain layout render
- `Box` parity now also covers baseline `alignSelf` and `flexBasis` behavior, including percentage `flexBasis` in both row and column containers
- `Box` parity now also covers baseline `flexShrink`, including upstream `dont shrink` and `shrink equally` row behavior
- `Box` parity now also covers baseline no-border `overflowX`, `overflowY`, and `overflow="hidden"` clipping for row, multiline text, and negative-margin intersection cases
- `Box` parity now also covers baseline `borderStyle="round"` rendering for fit-content, fixed-width, padding, and wrapped-text cases
- `Box` parity now also covers baseline hidden border sides and `overflowX` / `overflowY` clipping when the overflow container itself has a round border
- `Box` parity now also covers full-width root border boxes and custom object `borderStyle` glyph sets
- `Box` parity now also covers baseline background-driven geometry for fixed-size, bordered, padded, centered, and column-layout boxes in non-color debug rendering
- `Box` parity now also covers more overflow-with-border cases, including child bordered overflow, bordered overflow containers with multiple rows, top-edge intersection, and clipped overflow blocks with preserved trailing rows
- `Box` parity now also covers more overflow-with-border cases from upstream `overflow.tsx`, including horizontal multi-box clipping, empty bordered viewports when content is fully above or below the clip window, bottom-edge vertical intersections, and bordered mixed-box overflow containers with preserved trailing padding
- `Box` parity now also covers the matching plain overflow edge cases, including content fully before or after the horizontal clip window, right-edge intersections, and the vertical above/below/bottom-intersection cases without borders
- `Box` parity now also covers plain `overflow="hidden"` corner intersections on all four edges plus nested overflow containers
- `Box` parity now also covers the upstream out-of-bounds border write safety case when the rendered box exceeds the available terminal width
- `Box` parity now also covers more default `flexShrink` combinations, including text-text rows, mixed text-box rows, bordered mixed rows, and the cases where Yoga-style rounded overlap changes the final rendered columns
- `Box` parity now also covers more three-item default `flexShrink` rows, including text-text-text, box-text-box, text-box-box, box-box-text, spaced text triples, and mixed rows with a wider middle child
- `Text` nodes now use the upstream-like default `flexShrink` baseline, with width-aware remeasurement after flex shrink and parity coverage for clipped multi-text bordered overflow rows
- `Box` nodes now also use the upstream default `flexShrink` baseline, with parity coverage for default-shrinking sibling boxes in both plain and bordered rows
- basic app/runtime hooks are now available through `AppOptions`, `UseApp`, `UseStdin`, `UseStdout`, `UseStderr`, `UseCursor`, and `UseIsScreenReaderEnabled`
- `UseFocusManager` is now available with app-scoped `enable`, `disable`, `focus`, `focusNext`, and `focusPrevious` controls
- render boundaries now clear runtime hook context correctly, and cursor state no longer leaks across renders when `UseCursor` stops being used
- `UseEffect` now runs after the render pass instead of during component evaluation, which allows effect code to read ref-backed layout data from the current frame
- mounted sessions now perform a follow-up rerender when an effect updates state after commit, which closes the main gap needed for `measureElement`-driven state updates
- mounted sessions now also dispatch `UseInput` callbacks through `HandleInput` and compatible `stdin.SubscribeInput(...)` streams, with state updates and focus-tab navigation triggering rerenders
- `UseInput` now accepts both the legacy `(input, keys)` callback form and an Ink-like `(input string, key InputKey)` form with boolean key flags
- `UseInput` now also supports the upstream-style `isActive` option, raw-mode refcounting across multiple active hooks, and fuller modifier parsing for pasted input, `meta`, `ctrl`, `shift-tab`, `meta+arrow`, `ctrl+arrow`, and option-return sequences
- `UseFocus` now supports inactive focus targets that stay in registration order, skip correctly during `tab` / `shift-tab` navigation, enable raw mode only for active hooks, and start `focusPrevious()` from the last active target when nothing is focused
- mounted focus runtime now also mirrors upstream `Esc` behavior by blurring the currently focused component and rerendering the live tree when focus management is enabled
- mounted focus runtime tests now also cover `shift-tab` wraparound, and focus-manager tests now lock the upstream unregister semantics where re-registering an already mounted component does not reapply `autoFocus`
- a session-based runtime API now exists through `Mount` and `MountWithOptions`, with `Rerender`, `Clear`, `Unmount`, and `WaitUntilExit`
- standard render output now tracks cursor-only updates, restores the terminal cursor on TTY unmount, and preserves direct `UseStdout().Write()` output ahead of the managed Ink block
- `RenderWithOptions` now reuses mounted instances per stdout target, matching upstream `render()` more closely
- `pkg/ink` now exports wrappers for `UseInput`, `UseFocus`, `UseEffect`, `UseMemo`, `UseCallback`, and `UseRef`
- `RenderOptions` now supports `Debug` and `IncrementalRendering`, and the output layer has basic parity for append-only debug mode plus surgical incremental updates
- log-update parity tests now also cover no-trailing-newline transitions, repeated clear behavior, and rendering down to a single empty line in incremental mode
- session runtime now exposes `Sync`, `WriteStdout`, and `WriteStderr`, with parity coverage for cursor replay, cursor reset after `sync()`, and restoring the managed Ink block after out-of-band stdout/stderr writes
- runtime rendering now splits `Static` output from the managed dynamic block, so appended static items are written above the live region without redrawing older static lines
- `Static` append-only output now tracks per-root item counts, so replacements do not replay old static content and fullscreen redraws can rebuild the accumulated static history
- `RenderOptions` now also supports `MaxFPS`, with throttled rerender coalescing, unmount flush/cancel behavior, and TTY synchronized write wrapping for deferred frames
- throttled rerenders now precompute frames, allowing `Static` appends and exit-triggering renders to bypass the write throttle and cancel stale pending frames
- managed render instances now auto-resize from compatible stdout implementations, rerender immediately on resize, clear before width shrink, and clean up resize listeners on unmount
- screen-reader mode now uses a dedicated plain-text render path with basic `aria-label`, `aria-hidden`, `aria-role`, `aria-state`, `display="none"`, and static-output support
- standard TTY rerenders now use fullscreen `clearTerminal` recovery when the previous interactive frame filled the viewport, while still avoiding fullscreen clears when only `Static` content exceeds the viewport
- standard and incremental render paths now normalize non-fullscreen output with a trailing newline, while keeping internal logical output separate from the rendered terminal buffer state
- runtime tests now also cover fullscreen renders without an extra bottom newline plus standard/fullscreen transitions from populated output down to an empty frame
- runtime tests now also cover reconciler-style rerenders for child updates, text-node updates, append/insert/remove flows, keyed child reorder, and replacing a styled nested child with plain text
- runtime tests now also cover nested text growth rerenders where a child `Text` node is appended inside an existing `Text` tree and the measured output width must expand from `abc` to `abcx`
- managed render tests now also cover `nil -> component` transitions for `MeasureElement`, including the effect-driven follow-up render that converges from `Width: 0` to the measured width and the same flow under a throttled managed render instance
- runtime tests now also cover `OnRender` callback delivery for the initial render, a manual rerender, and an internal state-driven rerender triggered through `UseInput`
- input normalization and mounted `UseInput` runtime tests now also cover the remaining upstream cursor-navigation key matrix for `pageUp`, `pageDown`, `home`, `end`, and the full `meta` / `ctrl` arrow-direction set
- low-level input parsing now also recognizes broader upstream raw escape variants, including SS3 cursor keys, putty-style `ESC [[5~` / `ESC [[6~` page keys, and modifier cursor sequences such as `ESC [1;5A`
- upstream `Text` parity now also covers non-TTY style props such as `color`, `backgroundColor`, `dimColor`, `bold`, `italic`, `underline`, `inverse`, `strikethrough`, and nested style inheritance/override cases
- upstream `Text` parity now also covers the baseline width-constrained `wrap`, `truncate`, `truncate-middle`, and `truncate-start` behaviors inside fixed-width boxes
- TTY-backed runtime rendering now uses the ANSI layout renderer for dynamic and static sections, so mounted sessions preserve `color`, `backgroundColor`, `borderColor`, and the basic text modifiers (`bold`, `dimColor`, `italic`, `underline`, `inverse`, `strikethrough`) instead of dropping back to plain text
- ANSI border rendering now also supports per-side border color and dim props (`borderTopColor`, `borderBottomColor`, `borderLeftColor`, `borderRightColor`, `borderDimColor`, and the per-side dim variants)
- ANSI text rendering now preserves nested inline style/background overrides for plain `Text` trees, including parent color/background resumption after nested child overrides or clears, the same resumption behavior for `truncate` / `truncate-start` / `truncate-middle` output, and transform-aware style transitions through an ANSI roundtrip path for the current `Transform` parity cases
- hook runtime now reuses `UseInput` and `UseFocus` slots across rerenders instead of accumulating duplicate registrations, and focus navigation now handles missing-current-focus, inactive targets, and missing-id no-op behavior more like upstream Ink
- `UseFocus` rerenders now also replay `autoFocus` registration changes more like upstream Ink’s effect-driven hook behavior, while the wrapper raw-mode effect no longer depends on the generated focus id
- element refs now receive the rendered node handle through `props["ref"]`, and `MeasureElement` can read the latest computed width and height from that handle
- effect-driven measurement updates now trigger a mounted follow-up rerender after the current commit, which is enough to support the basic `measureElement` flow
- `DOMElement` now exposes a broader DOM-like surface through parent/child traversal, attribute access, text-content accessors, and computed-layout accessors, and `MeasureElement` now has a narrow intrinsic fallback for bare text-node refs
- input payloads are now normalized for hooks, including special-key delivery and basic `tab` / `shift-tab` focus navigation parity
- input payloads now also expose an Ink-like boolean key object for `upArrow`, `downArrow`, `leftArrow`, `rightArrow`, `return`, `escape`, `ctrl`, `shift`, `tab`, `backspace`, `delete`, and related flags
- renderer-owned layout parsing now accepts native Go numeric props for spacing and flex values, and screen-reader state narration now also accepts `map[string]bool` state inputs in addition to `vdom.Props`
- renderer-owned child ordering now skips `nil` children entirely, which matches upstream null-child behavior instead of reserving zero-size layout slots that could distort spacing
- upstream golden parity now also covers more `borders.tsx` and `background.tsx` cases, including full-width round borders, centered/bottom-aligned bordered content, long wrapped text inside bordered boxes, nested bordered boxes, nested row borders with wide/emoji content, extra ANSI background inheritance variants, RGB box backgrounds, and the remaining screen-reader null-child case from `screen-reader.tsx`
- upstream golden parity now also covers direct upstream cases for colored leading whitespace, single-node wide/emoji fit-content round borders, column-stacked bordered wide/emoji content, single-child `flexGrow`, styled `justifyContent="flex-end"` alignment, and plain `overflowY` multi-box/top-intersection clipping
- screen-reader role suppression now matches upstream direct-parent semantics, so neutral wrappers no longer hide nested same-role narration from accessibility output
- upstream golden parity now also covers additional direct upstream accessibility/layout cases, including `screen-reader` nested same-role wrappers, `Static` plus screen-reader parent-role behavior, the previously added `Transform` accessibility labels, empty-content role/state narration spacing, ordered multi-state narration, newline/wrapped-text padding and margin cases, the remaining direct `minWidth="50%"` case from `width-height.tsx`, the remaining fit-content colorful multi-node border case from `borders.tsx`, a direct alias for the `borders.tsx` fit-content variation-selector emoji round-border fixture, direct `undefined` / `null` / single-empty-text child cases from `text.tsx` and `components.tsx`, the direct `text with component` / `text with fragment` / `fragment` cases from `components.tsx`, the direct `text with variable` / `number` cases from `components.tsx`, direct aliases for `Transform` children/squashing fixtures plus the direct `Newline` and `Spacer` component fixtures from `components.tsx`, direct aliases for the upstream `text.tsx` color/background/inversion fixtures, direct aliases for the upstream `screen-reader.tsx` baseline plus aria-state/multiline/listbox/select-input cases, direct aliases for the upstream `components.tsx` basic text plus `width-height.tsx` width/min-height cases, a broader block of direct `borders.tsx` aliases for single-node, multi-node, and nested round-border fixtures, the remaining direct `components.tsx` text aliases for variable/component/fragment children, wrap behavior, truncation, and empty-text handling, plus exact-name aliases for the upstream `display.tsx`, `gap.tsx`, `flex-direction.tsx`, `flex-wrap.tsx`, `padding.tsx`, `margin.tsx`, `text-width.tsx`, `screen-reader.tsx`, and `width-height.tsx` fixture titles where the rendered output is already covered, along with the remaining parity-safe concurrent exact-title aliases from `components.tsx`, `text.tsx`, `width-height.tsx`, and `borders.tsx`, direct aliases for upstream flex shrink/basis and column justify-content fixtures, the exact-title alias for `gap - concurrent`, and the final failure/runtime fixture titles now handled directly in the parity harness through `mode: "error"`, `mode: "managed-frames"`, runtime measurement modes, and runtime render modes, bringing the suite to 784 cases total with 467 `Box` cases, 157 `Text` cases, 49 `Transform` cases, 38 `Static` cases, 33 `Newline` cases, 34 `Spacer` cases, 2 `Measure` cases, and 4 `Render` cases
- mounted runtime now matches upstream `render all frames if CI environment variable equals false` behavior more closely by disabling synchronized redraw semantics in CI, skipping managed resize subscriptions there, streaming static output immediately, and emitting only the final dynamic frame on unmount
- public `components.Box` / `components.Text` trees now enforce the upstream invalid-text fixture semantics at render time, and the upstream parity harness now understands those failure cases directly through `mode: "error"` plus the remaining CI=false multi-frame runtime case through `mode: "managed-frames"`
- render-phase panic recovery now covers both the initial mount and subsequent managed rerenders, so thrown render errors surface through `WaitUntilExit()` instead of escaping the process
- mounted session `WriteStdout` / `WriteStderr` restore cycles now also use synchronized update wrapping on TTYs, matching upstream clear-write-restore behavior more closely
- mounted `UseStdout().Write(...)` / `UseStderr().Write(...)` calls now also flow through the same session-managed clear-and-restore path as explicit `Instance.WriteStdout` / `WriteStderr`
- `WaitUntilExit()` now flushes buffered stdout writers during unmount and propagates flush failures, while CI detection now matches upstream `is-in-ci` semantics more closely
- disabling focus management now preserves visible focus state instead of hiding the currently focused item, while manual `focus()` / `blur()` updates continue to trigger rerenders
- missing `focus(id)` targets now fall back to the first registered focus target, which matches the upstream `useFocusManager().focus()` contract more closely
- `MeasureText` is now exposed publicly, ignores ANSI escape sequences when computing visible width, and `DOMElement` now also supports DOM-style navigation helpers such as sibling traversal and ancestry checks
- `MeasureText` now also treats ASCII and C1 control characters as zero-width, and DOM-style node mutations invalidate stale computed layout so later measurements do not reuse obsolete dimensions
- low-level input parsing now also collapses more upstream-style meta-prefixed raw input variants, including `ESC m`, `ESC space`, `ESC return`, `ESC delete`, and doubled-escape meta-arrow sequences
- modern `UseInput` key objects now treat bare `Escape` as `escape + meta`, while legacy key-name slices keep the upstream-compatible bare `["escape"]` shape
- ANSI styled-text rendering now also treats variation selectors and other zero-width code points as display width `0`, keeping emoji sequences stable during styled wrapping and truncation
- `MeasureText` and text-node `MeasureElement` fallback now also treat emoji modifier presentation clusters such as `✌🏽` as width `2`, and ignore OSC hyperlink `ST` escape sequences in width calculations
- screen-reader rendering now treats a default `Box` as row-oriented output, joining sibling content with spaces unless `flexDirection` explicitly switches it to a column
- renderer text layout and terminal writes now operate on grapheme clusters for width-sensitive paths, preserving combining marks, variation selectors, ZWJ emoji clusters, and OSC hyperlink sequences through plain and ANSI rendering
- screen-reader runtime `Static` deltas now append the same newline boundary as upstream, so newly appended static lines do not concatenate with the live region
- `UseMouse` dispatch is now scoped per mounted app instance while preserving the package-global hooks compatibility API for low-level tests
- legacy X10 mouse frames now drive mounted `UseMouse` subscribers through the same app-scoped runtime path as SGR 1006 reports
- reconciler keyed diff now falls back to positional diff when duplicate sibling keys are present, avoiding map overwrite behavior for invalid-but-possible input trees
- upstream golden parity now includes Node-generated `measureElement` runtime fixtures for the direct debug flow and the throttled `render(null) -> rerender(<Test />)` flow; throttled pending renders now settle effect-driven state changes before committing stale intermediate frames
- upstream golden parity now includes a Node-generated non-TTY `render()` maxFps throttle write sequence, and terminal output now matches upstream by skipping initial cursor hide on non-TTY streams and emitting `ESC[G` for column-zero cursor movement
- upstream golden parity now also covers TTY throttled render no-op behavior for unchanged output/cursor state and synchronized `bsu`/`esu` wrapping for trailing throttled writes

Public export audit against `../ink/src/index.ts`:

- core component exports are present
- core runtime hook exports are present
- `CursorPosition` is present
- `measureElement` now has a minimal Go equivalent through `MeasureElement`
- `DOMElement` now exists as a minimal Go-level ref handle backed by the rendered `vdom.Node`
- full ref parity and broader state-driven rerender behavior still differ from upstream Ink

Recently closed gaps (parallel-agent sweep on 2026-04-25):

- accessibility: `aria-state` arbitrary truthy keys with `accessibilityStateOrder` precedence and alphabetical fallback, `aria-state.checked="mixed"`, top-level shorthand props (`aria-busy`/`aria-checked`/`aria-disabled`/`aria-expanded`/`aria-readonly`/`aria-required`/`aria-selected`), `aria-level` heading narration (int/float forms), and `aria-live="off"` subtree suppression
- DOM ref / `measureElement`: new methods on `*vdom.Node` (`GetAttribute`, `Style`, `InternalStatic`, `ElementChildren`, `OwnerRoot`, `Position`) plus `pkg/ink.MeasureElementPosition` returning computed `(left, top)`
- `useInput`: bracketed-paste detection and split (`\x1b[200~ ... \x1b[201~`) including coalesced leading/trailing keypress chunks and embedded CR/LF/NUL/tab/UTF-8/emoji payloads, full F1-F12 dispatch matrix coverage (SS3/CSI/Cygwin variants with ctrl/shift/meta modifiers), full ctrl+letter (a-z) matrix coverage, and `isActive: true → false → true` lifecycle re-registration with raw-mode refcount
- `useFocus`: parity-correct `UseFocusOpts(...FocusOptions) FocusState` returning `{IsFocused, Focus, Blur}` while keeping the legacy 3-tuple form for back-compat; `disable()` now preserves focus state and re-enable rejoins, `focus(missingID)` is silent no-op
- `render()`: debug-mode short-circuit precedence over screen-reader, debug rerenders emit append even on identical output, `staticOutput == "\n"` no longer triggers a rewrite of the dynamic block, incremental `Clear()` is idempotent
- Box/Text styles: ten new parity cases across Transform / overflow-with-border / default-flexShrink, plus a renderer fix that clips `textWidth` to actual ancestor-allotted width when a flex container shrinks (`childWidthConstraintWithLayout` + `horizontalPaddingInsets`)

Closed in second parallel-agent sweep:

- Yoga layout: 17 new parity cases covering proportional `flex-grow` (3:1, 2:1, 0.5:1), `min-width`/`min-height` interaction with `width`/`flex-grow`, negative-gap clamping, `display:none` siblings, `position:"absolute"` + `flex-grow`/`margin`, `align-self="stretch"` overriding `align-items`. `pkg/layout` now clamps negative `gap`/`rowGap`/`columnGap` to zero (Yoga spec)
- Aria ID resolution: `aria-labelledby` / `aria-describedby` resolved against an indexed id-map built from the full tree (including hidden subtrees), with cycle-safe `visited` set. `aria-labelledby` substitutes the host's narration with the joined narration of each referenced node (precedence: labelledby > aria-label > children); `aria-describedby` appends after the host narration before role/state decoration. Self-references resolve to "" without recursion
- Aria live announcer: `aria-live="polite"`/`"assertive"` regions are duplicated into a dedicated announcer block at the end of `Output`, prefixed `[polite]`/`[assertive]`, with assertive ordered before polite
- Cell-level dirty-rect repaints: opt-in `RenderOptions.CellLevelDiff` adds a `[][]renderedCell` parser plus per-cell diff write path that emits `cursorTo(x, y)` + SGR transition + cell text only for changed cells. Falls back safely to line/column diff on parse error, frame-size change, row width drift, identical output, or first paint
- Form validation: `FormFieldConfirm` + `FormFieldMultiSelect` field kinds with kind-aware "is-present" semantics for `Required`. New `FormState.OrderedErrors() []error` / `FieldError(name)` / `HasErrors()` accessors, plus `ErrorOverviewFromForm(state, runtime...)` helper that pipes Form errors directly into `ErrorOverviewGroup`
- Cross-chunk bracketed paste: stateful `Instance.pendingPaste` buffer holds payloads when `\x1b[200~` arrives without a matching `\x1b[201~`; subsequent chunks append until the end marker (handled even when split across chunks) then dispatch as a single paste event. Buffer cleared on unmount
- Newline edge cases: `count=0` emits empty, negative counts clamp to 0 (documented divergence — upstream RangeError-throws), large counts pass through. Parity harness `Count` field migrated to `*int` to round-trip explicit zeros and negatives
- BigText: `FontSlim` (4×3 hand-drawn box-drawing glyphs) + `FontOutline` (5×5 algorithmically derived from Block via 4-neighbour rewrite producing `▀`/`▄`/`▒`)
- Syntax: `SyntaxPython` (keywords + decorators + triple-quoted strings + f/r/b prefixes), `SyntaxRust` (lifetimes + attributes + char literals + underscore-numeric literals), `SyntaxSQL` (case-insensitive 60+ keyword set with original casing preserved + `''`-escaped strings + `--`/`/* */` comments)
- proportional `flex-shrink` with non-1 weights + text wrapping: `layout.Node` now exposes a `SetMeasureSizeFunc` callback that returns both the floored wrap-budget width and the resulting wrapped height. The shrink-time remeasure pass records the floored width as `measuredWidth` on the node, and the renderer honors it via `ShouldHonorMeasuredWidth` (gated on `parent.sizeAdjusted`) so an inner text inside a fractionally-shrunk box wraps at the same budget upstream Yoga's `getMaxWidth(yogaNode)` flow uses. Parity cases `box/flex-shrink-2-1` and `box/flex-shrink-clamped-by-min-width-percent` now match upstream. The directly-shrunk text-pair case (e.g., "Hello "+"World") deliberately keeps the existing ceil-based textLike rounding via the `parent.sizeAdjusted` gate, since goink does not iterate the Yoga measure-and-redistribute pass that upstream uses to reclaim trailing-space slack.

Closed in third parallel-agent sweep:

- runtime aria-live announcer + `UseAnnounce()`: session-scoped `Announcer` with `BeginRender`/`Active`/`Pending` rotation; polite/assertive announcements duplicated into screen-reader output as `[polite]`/`[assertive]` blocks; cache-compatible
- minimal SGR diff in `CellLevelDiff`: `parseSGR` + `sgrTransition` emit only flipped attributes; adjacent same-style cells share one SGR setup (~50-62% byte reduction)
- reconciler caches: `vdom.Node` carries `transformCache` + `cachedStaticOutput`/`staticDirty`; `applyNodeTransform`/`applyLineTransform` and `RenderWithLayoutSectionsMode` consult them; `SetAttribute`/`SetNodeValue` walk static ancestors and invalidate
- wide-rune correctness: cluster-aware `wrapLine`/`truncateEnd`/`truncateStart`/`truncateMiddle` keep ZWJ families intact at wrap points and clip wide runes correctly at fixed-box boundaries; CJK + ANSI styling preserves style across both columns
- `FormFieldTab` + `FormFieldQuickSearch` field kinds plus `FormWizard` controller (Next/Prev/Submit with validation gates, `ErrorOverviewFromWizard` aggregation, `examples/wizard-demo`)

Closed in current sweep:

- renderer OSC preservation: ANSI text collection now parses raw ANSI/OSC sequences in text nodes, so OSC 8 hyperlinks survive the ANSI layout renderer instead of being counted as visible columns
- upstream golden coverage now includes explicit plain and ANSI OSC 8 hyperlink fixtures generated by Node Ink
- border / negative-margin draw order: box borders are painted before descendants, while `overflow: "hidden"` clips descendants to the inner border box, matching the upstream negative-margin border fixture
- wrap-ansi leading/trailing space parity: plain wrapping now preserves upstream blank-line and separator behavior for `text/wrap-preserves-leading-and-trailing-spaces`
- fixed-width ZWJ emoji parity: plain buffer writes now operate on grapheme clusters, so `box/emoji-zwj-width-fixed-box` preserves the full family emoji before adjacent siblings
- upstream parity skip list is now empty; all 784 generated upstream golden cases are active

Still incomplete compared with upstream Ink:

- iterative Yoga shrink-and-redistribute pass for directly-shrunk text-only sibling rows (ceil-based textLike rounding currently masks this for 1:1 siblings; `parent.sizeAdjusted` gate dodges fractional cases)
- no generated upstream golden cases are currently skipped via `knownDeferredParityCases` in `tests/upstream_parity_test.go`
- remaining risk is mostly coverage breadth: more real-project runtime patterns under `tests/project_upstream`, more terminal edge cases, and more external `ink-*` component compatibility targets

## Suggested Next Steps

1. Add more Node-generated runtime goldens for real upstream project usages under `tests/project_upstream`.
2. Iterative shrink-and-redistribute pass for directly-shrunk text leaves (Yoga's measure-and-redistribute behavior).
3. Continue expanding component coverage to upstream `ink-*` parity targets (more `BigText` fonts and `Syntax` languages).
4. Promote the `examples/` demos into a curated documentation site.
