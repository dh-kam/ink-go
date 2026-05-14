package ink

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TestHandleInputDispatchesFunctionKeyMatrix verifies that the entire F1-F12
// matrix and its xterm modifier variants reach the legacy keys[] slice on
// useInput hooks. Upstream Ink does not expose F-key fields on its Key struct,
// so the modern callback only sees `key.Ctrl`/`key.Shift`/`key.Meta` and an
// empty `input` payload — the keys[] slice is goink's escape hatch for letting
// callers actually distinguish F1 from F5.
func TestHandleInputDispatchesFunctionKeyMatrix(t *testing.T) {
	type expected struct {
		keys  []string
		ctrl  bool
		shift bool
		meta  bool
	}

	cases := []struct {
		name string
		raw  string
		want expected
	}{
		// Plain function keys via xterm/SS3.
		{name: "f1-ss3", raw: "\x1bOP", want: expected{keys: []string{"f1"}}},
		{name: "f2-ss3", raw: "\x1bOQ", want: expected{keys: []string{"f2"}}},
		{name: "f3-ss3", raw: "\x1bOR", want: expected{keys: []string{"f3"}}},
		{name: "f4-ss3", raw: "\x1bOS", want: expected{keys: []string{"f4"}}},
		{name: "f5-csi", raw: "\x1b[15~", want: expected{keys: []string{"f5"}}},
		{name: "f6-csi", raw: "\x1b[17~", want: expected{keys: []string{"f6"}}},
		{name: "f7-csi", raw: "\x1b[18~", want: expected{keys: []string{"f7"}}},
		{name: "f8-csi", raw: "\x1b[19~", want: expected{keys: []string{"f8"}}},
		{name: "f9-csi", raw: "\x1b[20~", want: expected{keys: []string{"f9"}}},
		{name: "f10-csi", raw: "\x1b[21~", want: expected{keys: []string{"f10"}}},
		{name: "f11-csi", raw: "\x1b[23~", want: expected{keys: []string{"f11"}}},
		{name: "f12-csi", raw: "\x1b[24~", want: expected{keys: []string{"f12"}}},

		// Ctrl + function keys via the xterm modifyOtherKeys 1;5 protocol.
		{name: "ctrl-f1", raw: "\x1b[1;5P", want: expected{keys: []string{"ctrl", "f1"}, ctrl: true}},
		{name: "ctrl-f5", raw: "\x1b[15;5~", want: expected{keys: []string{"ctrl", "f5"}, ctrl: true}},
		{name: "ctrl-f12", raw: "\x1b[24;5~", want: expected{keys: []string{"ctrl", "f12"}, ctrl: true}},

		// Shift + function keys via the xterm 1;2 modifier.
		{name: "shift-f1", raw: "\x1b[1;2P", want: expected{keys: []string{"f1"}, shift: true}},
		{name: "shift-f5", raw: "\x1b[15;2~", want: expected{keys: []string{"f5"}, shift: true}},

		// Alt/meta + function keys via the xterm 1;3 modifier.
		{name: "meta-f1", raw: "\x1b[1;3P", want: expected{keys: []string{"meta", "f1"}, meta: true}},

		// Cygwin/libuv variants for F1-F4.
		{name: "f1-cygwin", raw: "\x1b[[A", want: expected{keys: []string{"f1"}}},
		{name: "f5-cygwin", raw: "\x1b[[E", want: expected{keys: []string{"f5"}}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			var modernInput string
			var modernKey InputKey
			var legacyKeys []string
			var modernCalls int

			instance, err := MountWithOptions(func() *vdom.Node {
				UseInput(func(input string, key InputKey) {
					modernInput = input
					modernKey = key
					modernCalls++
				})
				UseInput(func(input interface{}, keys []string) bool {
					legacyKeys = append([]string(nil), keys...)
					return false
				})

				return components.Text("Input")
			}, RenderOptions{
				AppOptions: AppOptions{Stdout: stdout},
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.HandleInput(testCase.raw); err != nil {
				t.Fatalf("handle input failed: %v", err)
			}

			if modernCalls != 1 {
				t.Fatalf("expected exactly one modern callback dispatch, got %d", modernCalls)
			}
			if modernInput != "" {
				t.Fatalf("expected empty input string for function key, got %q", modernInput)
			}
			if modernKey.Ctrl != testCase.want.ctrl {
				t.Fatalf("ctrl mismatch: want %v, got %+v", testCase.want.ctrl, modernKey)
			}
			if modernKey.Shift != testCase.want.shift {
				t.Fatalf("shift mismatch: want %v, got %+v", testCase.want.shift, modernKey)
			}
			if modernKey.Meta != testCase.want.meta {
				t.Fatalf("meta mismatch: want %v, got %+v", testCase.want.meta, modernKey)
			}
			if strings.Join(legacyKeys, ",") != strings.Join(testCase.want.keys, ",") {
				t.Fatalf("legacy keys mismatch: want %q, got %q",
					strings.Join(testCase.want.keys, ","),
					strings.Join(legacyKeys, ","))
			}
		})
	}
}

// TestHandleInputDispatchesEntireCtrlLetterMatrix verifies that every ctrl+letter
// combination (a-z) lands as `ctrl=true`, `input="<letter>"`, `keys=["ctrl","ctrl-<letter>"]`.
// Several control bytes overlap with named keys (ctrl+H = backspace, ctrl+I = tab,
// ctrl+J = enter, ctrl+M = return) and must keep their named identities, so this
// test pins the expected boundary explicitly.
func TestHandleInputDispatchesEntireCtrlLetterMatrix(t *testing.T) {
	// Bytes that the parser intentionally remaps to a named key rather than
	// surfacing as ctrl+letter. This mirrors upstream Ink's parseKeypress.
	namedOverrides := map[byte]string{
		0x08: "backspace", // ctrl+H
		0x09: "tab",       // ctrl+I
		0x0a: "enter",     // ctrl+J (line feed) — upstream names this "enter" but goink delivers it as "return".
		0x0d: "return",    // ctrl+M (carriage return)
	}

	for letter := byte('a'); letter <= byte('z'); letter++ {
		ctrlByte := letter - 'a' + 1
		t.Run(fmt.Sprintf("ctrl-%c", letter), func(t *testing.T) {
			if name, ok := namedOverrides[ctrlByte]; ok {
				// Skip the named overrides — they are tested elsewhere (see
				// TestHandleInputMatchesRemainingUseInputFixtureMatrix). The
				// matrix is exhaustive only over the ctrl-letter slots that
				// actually map to a ctrl-letter key event.
				t.Skipf("ctrl+%c maps to %q rather than a ctrl-letter event", letter, name)
			}

			// ctrl+C is the global exit shortcut by default; turn it off so
			// the dispatch reaches the user hook.
			stdout := &recordingWriter{}
			var modernInput string
			var modernKey InputKey
			var legacyKeys []string

			instance, err := MountWithOptions(func() *vdom.Node {
				UseInput(func(input string, key InputKey) {
					modernInput = input
					modernKey = key
				})
				UseInput(func(input interface{}, keys []string) bool {
					legacyKeys = append([]string(nil), keys...)
					return false
				})

				return components.Text("Input")
			}, RenderOptions{
				AppOptions:  AppOptions{Stdout: stdout},
				ExitOnCtrlC: boolPtr(false),
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.HandleInput(string([]byte{ctrlByte})); err != nil {
				t.Fatalf("handle input failed: %v", err)
			}

			if modernInput != string(letter) {
				t.Fatalf("ctrl+%c modern input = %q, want %q", letter, modernInput, string(letter))
			}
			if !modernKey.Ctrl {
				t.Fatalf("ctrl+%c expected ctrl=true, got %+v", letter, modernKey)
			}
			wantKeys := []string{"ctrl", fmt.Sprintf("ctrl-%c", letter)}
			if strings.Join(legacyKeys, ",") != strings.Join(wantKeys, ",") {
				t.Fatalf("ctrl+%c legacy keys = %q, want %q", letter,
					strings.Join(legacyKeys, ","),
					strings.Join(wantKeys, ","))
			}
		})
	}
}

// TestHandleInputBracketedPastePreservesEmbeddedSpecialBytes verifies that the
// bracketed-paste detection in NormalizeHookInput surfaces the inner payload
// verbatim — including embedded carriage returns, line feeds, NUL bytes, and
// multi-byte UTF-8 emoji — without coercing them into named-key events.
//
// Goink intentionally diverges from upstream here: upstream Ink has no
// bracketed-paste detection, so a paste of "hello" arrives as the literal
// "\x1b[200~hello\x1b[201~" string. Goink strips the markers, since paste is
// the whole reason terminals send the markers in the first place.
func TestHandleInputBracketedPastePreservesEmbeddedSpecialBytes(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "embedded-cr", body: "line1\rline2"},
		{name: "embedded-lf", body: "line1\nline2"},
		{name: "embedded-crlf", body: "win\r\nstyle"},
		{name: "embedded-nul", body: "before\x00after"},
		{name: "embedded-tab", body: "col1\tcol2"},
		{name: "multibyte-utf8", body: "résumé漢字"},
		{name: "emoji", body: "👋🌍"},
		{name: "empty", body: ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			var modernInput string
			var modernKey InputKey
			var legacyInput interface{}
			var legacyKeys []string

			instance, err := MountWithOptions(func() *vdom.Node {
				UseInput(func(input string, key InputKey) {
					modernInput = input
					modernKey = key
				})
				UseInput(func(input interface{}, keys []string) bool {
					legacyInput = input
					legacyKeys = append([]string(nil), keys...)
					return false
				})

				return components.Text("Input")
			}, RenderOptions{
				AppOptions: AppOptions{Stdout: stdout},
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			raw := "\x1b[200~" + testCase.body + "\x1b[201~"
			if err := instance.HandleInput(raw); err != nil {
				t.Fatalf("handle input failed: %v", err)
			}

			if modernInput != testCase.body {
				t.Fatalf("expected modern input %q, got %q", testCase.body, modernInput)
			}
			// Bracketed paste must not synthesize special-key flags. The user
			// is pasting opaque text — we should not pretend Enter/Tab/etc.
			// were pressed simply because the bytes happen to be in there.
			if modernKey.Return || modernKey.Tab || modernKey.Backspace ||
				modernKey.Delete || modernKey.Escape || modernKey.Ctrl ||
				modernKey.Shift || modernKey.Meta {
				t.Fatalf("expected no synthetic key flags for paste, got %+v", modernKey)
			}
			if legacyInput != testCase.body {
				t.Fatalf("expected legacy input %q, got %#v", testCase.body, legacyInput)
			}
			if legacyKeys != nil {
				t.Fatalf("expected no legacy keys for paste, got %#v", legacyKeys)
			}
		})
	}
}

// TestHandleInputBracketedPasteSplitsLeadingKeypress verifies that a chunk
// where a regular keypress arrived in the same TTY read as a bracketed paste
// — for example "x\x1b[200~paste\x1b[201~" — surfaces the leading 'x' as its
// own dispatch and the inner body as the paste event.
func TestHandleInputBracketedPasteSplitsLeadingKeypress(t *testing.T) {
	stdout := &recordingWriter{}
	type event struct {
		input string
	}
	var events []event

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			events = append(events, event{input: input})
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("x\x1b[200~paste\x1b[201~"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected leading keypress and paste body to dispatch as two events, got %d", len(events))
	}
	if events[0].input != "x" {
		t.Fatalf("expected first event input to be leading 'x', got %q", events[0].input)
	}
	if events[1].input != "paste" {
		t.Fatalf("expected second event input to be paste body, got %q", events[1].input)
	}
}

// TestHandleInputBracketedPasteSplitsTrailingKeypress verifies that a chunk
// containing a complete bracketed paste followed by additional bytes (which
// terminals do emit when the user pastes and then immediately presses a key
// before the kernel scheduling separates the read) is split cleanly: the
// paste body lands on the hook first, then the trailing key.
//
// Upstream Ink does not implement this — a chunk like "\x1b[200~hi\x1b[201~q"
// reaches the useInput callback as a single mangled string. Goink detects
// the markers and emits two events to better match user intent.
func TestHandleInputBracketedPasteSplitsTrailingKeypress(t *testing.T) {
	stdout := &recordingWriter{}
	type event struct {
		input string
		key   InputKey
	}
	var events []event

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			events = append(events, event{input: input, key: key})
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b[200~paste\x1b[201~q"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected paste body and trailing keypress to dispatch as two events, got %d", len(events))
	}
	if events[0].input != "paste" {
		t.Fatalf("expected first event input to be paste body, got %q", events[0].input)
	}
	if events[1].input != "q" {
		t.Fatalf("expected second event to be trailing 'q', got %q", events[1].input)
	}
	if events[1].key.Ctrl || events[1].key.Meta || events[1].key.Return {
		t.Fatalf("expected trailing 'q' to carry no modifier flags, got %+v", events[1].key)
	}
}

// TestHandleInputBracketedPasteCrossChunk verifies that when the kernel splits
// a bracketed-paste sequence across two TTY reads — start marker + leading
// payload in one chunk, trailing payload + end marker in the next — goink
// stitches the two halves together and dispatches a single paste event. This
// avoids leaking the start marker bytes back into the useInput callback's
// `input` string when the kernel happens to split the read at an inopportune
// boundary, which is something the stateless splitBracketedPaste fallback
// could not cover on its own.
func TestHandleInputBracketedPasteCrossChunk(t *testing.T) {
	stdout := &recordingWriter{}
	type event struct {
		input string
		key   InputKey
	}
	var events []event

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			events = append(events, event{input: input, key: key})
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b[200~ABC"); err != nil {
		t.Fatalf("handle first chunk failed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events while paste body is incomplete, got %d", len(events))
	}

	if err := instance.HandleInput("DEF\x1b[201~"); err != nil {
		t.Fatalf("handle second chunk failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one paste event after end marker arrives, got %d", len(events))
	}
	if events[0].input != "ABCDEF" {
		t.Fatalf("expected reassembled paste body \"ABCDEF\", got %q", events[0].input)
	}
	if events[0].key.Ctrl || events[0].key.Meta || events[0].key.Return {
		t.Fatalf("expected paste event to carry no modifier flags, got %+v", events[0].key)
	}
}

// TestHandleInputBracketedPasteCrossChunkWithTrailingKey verifies that when a
// chunk delivers the end marker followed by additional bytes, the buffered
// paste body and the trailing keystroke are surfaced as two distinct events.
func TestHandleInputBracketedPasteCrossChunkWithTrailingKey(t *testing.T) {
	stdout := &recordingWriter{}
	var inputs []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			inputs = append(inputs, input)
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b[200~hello"); err != nil {
		t.Fatalf("handle first chunk failed: %v", err)
	}
	if err := instance.HandleInput(" world\x1b[201~q"); err != nil {
		t.Fatalf("handle second chunk failed: %v", err)
	}

	if len(inputs) != 2 {
		t.Fatalf("expected paste + trailing key dispatched as two events, got %d (%v)", len(inputs), inputs)
	}
	if inputs[0] != "hello world" {
		t.Fatalf("expected paste body \"hello world\", got %q", inputs[0])
	}
	if inputs[1] != "q" {
		t.Fatalf("expected trailing key 'q', got %q", inputs[1])
	}
}

// TestHandleInputBracketedPasteCrossChunkSplitEndMarker verifies that when the
// end marker itself straddles two TTY reads (e.g. "\x1b[20" in one read,
// "1~" in the next), the buffered payload is still dispatched correctly once
// the second chunk completes the marker.
func TestHandleInputBracketedPasteCrossChunkSplitEndMarker(t *testing.T) {
	stdout := &recordingWriter{}
	var inputs []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			inputs = append(inputs, input)
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b[200~payload\x1b[20"); err != nil {
		t.Fatalf("handle first chunk failed: %v", err)
	}
	if len(inputs) != 0 {
		t.Fatalf("expected no events while end marker is incomplete, got %d", len(inputs))
	}

	if err := instance.HandleInput("1~"); err != nil {
		t.Fatalf("handle second chunk failed: %v", err)
	}
	if len(inputs) != 1 {
		t.Fatalf("expected one paste event once the end marker completes, got %d", len(inputs))
	}
	if inputs[0] != "payload" {
		t.Fatalf("expected paste body \"payload\", got %q", inputs[0])
	}
}

// TestHandleInputBracketedPasteCrossChunkUnmountDiscardsBuffer verifies that
// when a paste is in flight (start marker received but no end marker yet) and
// the session unmounts, the partial paste payload is discarded rather than
// delivered as an event.
func TestHandleInputBracketedPasteCrossChunkUnmountDiscardsBuffer(t *testing.T) {
	stdout := &recordingWriter{}
	var inputs []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			inputs = append(inputs, input)
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if err := instance.HandleInput("\x1b[200~partial"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}
	if !instance.pastePending {
		t.Fatalf("expected pastePending=true after start marker without end marker")
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if len(inputs) != 0 {
		t.Fatalf("expected no input events delivered, got %d (%v)", len(inputs), inputs)
	}
	if instance.pastePending || instance.pendingPaste != nil {
		t.Fatalf("expected paste buffer cleared after unmount, got pending=%v buf=%q", instance.pastePending, string(instance.pendingPaste))
	}
}

// TestHandleInputBracketedPasteCrossChunkLeadingKey verifies that when a chunk
// contains a leading keypress before an open-but-unterminated paste marker,
// the leading key is dispatched immediately and the paste body buffers across
// the chunk boundary.
func TestHandleInputBracketedPasteCrossChunkLeadingKey(t *testing.T) {
	stdout := &recordingWriter{}
	var inputs []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			inputs = append(inputs, input)
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("x\x1b[200~ab"); err != nil {
		t.Fatalf("handle first chunk failed: %v", err)
	}
	if len(inputs) != 1 || inputs[0] != "x" {
		t.Fatalf("expected leading 'x' dispatched immediately, got %v", inputs)
	}

	if err := instance.HandleInput("c\x1b[201~"); err != nil {
		t.Fatalf("handle second chunk failed: %v", err)
	}
	if len(inputs) != 2 {
		t.Fatalf("expected paste body event after end marker, got %d events", len(inputs))
	}
	if inputs[1] != "abc" {
		t.Fatalf("expected paste body \"abc\", got %q", inputs[1])
	}
}

// TestUseInputIsActiveToggleReRegistersDispatchAndRawMode verifies that when a
// useInput hook flips its IsActive option from true to false to true across
// renders, raw mode is released on disable and re-acquired on enable, and the
// hook stops/resumes receiving dispatches accordingly.
//
// This mirrors upstream Ink's pattern where `useEffect(() => setRawMode(true),
// [options.isActive, setRawMode])` cleans up and re-runs on isActive
// transitions.
func TestUseInputIsActiveToggleReRegistersDispatchAndRawMode(t *testing.T) {
	stdout := &recordingWriter{}
	active := true
	callCount := 0

	render := func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			callCount++
		}, InputOptions{IsActive: active})

		return components.Text("Input")
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  rawModeTestStdin{},
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if instance.app.rawModeUsers != 1 {
		t.Fatalf("expected raw mode acquired while active, got %d users", instance.app.rawModeUsers)
	}
	if err := instance.HandleInput("a"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected dispatch while active, got %d callback invocations", callCount)
	}

	// Toggle off — the cleanup from the prior effect must release raw mode,
	// and the hook must stop receiving dispatches.
	active = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender to inactive failed: %v", err)
	}
	if instance.app.rawModeUsers != 0 {
		t.Fatalf("expected raw mode released on isActive=false, got %d users", instance.app.rawModeUsers)
	}
	if instance.app.rawState != nil {
		t.Fatal("expected raw state to be cleared while hook is inactive")
	}
	if err := instance.HandleInput("b"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected inactive hook to stop receiving input, got %d callback invocations", callCount)
	}

	// Toggle back on — raw mode must re-acquire and dispatch must resume.
	active = true
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender to active failed: %v", err)
	}
	if instance.app.rawModeUsers != 1 {
		t.Fatalf("expected raw mode re-acquired on isActive=true, got %d users", instance.app.rawModeUsers)
	}
	if instance.app.rawState == nil {
		t.Fatal("expected raw state to be re-enabled when hook becomes active again")
	}
	if err := instance.HandleInput("c"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("expected reactivated hook to resume receiving input, got %d callback invocations", callCount)
	}
}
