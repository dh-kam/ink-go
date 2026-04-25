package input_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/input"
)

// TestKeyString tests Key string representation
func TestKeyString(t *testing.T) {
	tests := []struct {
		key      input.Key
		expected string
	}{
		{input.Key{Char: 'a', Name: ""}, "a"},
		{input.Key{Char: 0, Name: "escape"}, "escape"},
		{input.Key{Char: 0, Name: "up"}, "up"},
	}

	for _, tt := range tests {
		if tt.key.String() != tt.expected {
			t.Errorf("Key.String() = %q, want %q", tt.key.String(), tt.expected)
		}
	}
}

// TestEscapeSequenceParsing tests ANSI escape sequence parsing
func TestEscapeSequenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected input.Key
	}{
		{"\x1bOP", input.Key{Name: input.KeyF1, Char: 0}},
		{"\x1bOQ", input.Key{Name: input.KeyF2, Char: 0}},
		{"\x1b[11~", input.Key{Name: input.KeyF1, Char: 0}},
		{"\x1b[[E", input.Key{Name: input.KeyF5, Char: 0}},
		{"\x1b[A", input.Key{Name: "up", Char: 0}},
		{"\x1b[B", input.Key{Name: "down", Char: 0}},
		{"\x1b[C", input.Key{Name: "right", Char: 0}},
		{"\x1b[D", input.Key{Name: "left", Char: 0}},
		{"\x1b[E", input.Key{Name: input.KeyClear, Char: 0}},
		{"\x1b", input.Key{Name: "escape", Char: 0}},
		{"\x1b[2~", input.Key{Name: input.KeyInsert, Char: 0}},
		{"\x1b[3~", input.Key{Name: "delete", Char: 0}},
		{"\x1b[1~", input.Key{Name: "home", Char: 0}},
		{"\x1b[4~", input.Key{Name: "end", Char: 0}},
		{"\x1b[5~", input.Key{Name: "pageup", Char: 0}},
		{"\x1b[6~", input.Key{Name: "pagedown", Char: 0}},
		{"\x1bOA", input.Key{Name: "up", Char: 0}},
		{"\x1bOF", input.Key{Name: "end", Char: 0}},
		{"\x1b[7~", input.Key{Name: "home", Char: 0}},
		{"\x1b[8~", input.Key{Name: "end", Char: 0}},
		{"\x1b[[5~", input.Key{Name: "pageup", Char: 0}},
		{"\x1b[[6~", input.Key{Name: "pagedown", Char: 0}},
		{"\x1bOH", input.Key{Name: "home", Char: 0}},
		{"\x1b[1;5A", input.Key{Name: "up", Char: 0}},
		{"\x1b[1;5C", input.Key{Name: "right", Char: 0}},
		{"\x1b[e", input.Key{Name: input.KeyClear, Char: 0}},
		{"\x1bOe", input.Key{Name: input.KeyClear, Char: 0}},
		{"\x1b[a", input.Key{Name: "up", Char: 0}},
		{"\x1bOa", input.Key{Name: "up", Char: 0}},
		{"\x1b[Z", input.Key{Name: "shift-tab", Char: 0}},
		{"\x1b\r", input.Key{Name: "return", Char: 0}},
		{"\x1b\b", input.Key{Name: "backspace", Char: 0}},
		{"\x1b\x7f", input.Key{Name: "delete", Char: 0}},
		{"\x1b ", input.Key{Char: ' ', Name: ""}},
		{"\x1bm", input.Key{Char: 'm', Name: ""}},
		{"\x1bM", input.Key{Char: 'M', Name: ""}},
		{"\x1b\x1b[A", input.Key{Name: "up", Char: 0}},
	}

	for _, tt := range tests {
		result := input.ParseEscapeSequence(tt.input)
		if result.Name != tt.expected.Name || result.Char != tt.expected.Char {
			t.Errorf("ParseEscapeSequence(%q) = %+v, want %+v", tt.input, result, tt.expected)
		}
	}
}

// TestReadKey tests basic key reading from buffer
func TestReadKey(t *testing.T) {
	tests := []struct {
		input    string
		expected input.Key
	}{
		{"\x1bOP", input.Key{Name: input.KeyF1, Char: 0}},
		{"\x1b[[E", input.Key{Name: input.KeyF5, Char: 0}},
		{"a", input.Key{Char: 'a', Name: ""}},
		{"\x1b[A", input.Key{Name: "up", Char: 0}},
		{"\x1bOA", input.Key{Name: "up", Char: 0}},
		{"\x1b[E", input.Key{Name: input.KeyClear, Char: 0}},
		{"\x1b[2~", input.Key{Name: input.KeyInsert, Char: 0}},
		{"\x1bOF", input.Key{Name: "end", Char: 0}},
		{"\x1bOH", input.Key{Name: "home", Char: 0}},
		{"\x1b[1;5C", input.Key{Name: "right", Char: 0}},
		{"\x1b[e", input.Key{Name: input.KeyClear, Char: 0}},
		{"\x1bOe", input.Key{Name: input.KeyClear, Char: 0}},
		{"\x1b[a", input.Key{Name: "up", Char: 0}},
		{"\x1bOa", input.Key{Name: "up", Char: 0}},
		{"\x1b[[5~", input.Key{Name: "pageup", Char: 0}},
		{"\x1b[Z", input.Key{Name: "shift-tab", Char: 0}},
		{"\x1b\r", input.Key{Name: "return", Char: 0}},
		{"\x1b\b", input.Key{Name: "backspace", Char: 0}},
		{"\x1b\x7f", input.Key{Name: "delete", Char: 0}},
		{"\x1b ", input.Key{Char: ' ', Name: ""}},
		{"\x1bm", input.Key{Char: 'm', Name: ""}},
		{"\x1b\x1b[A", input.Key{Name: "up", Char: 0}},
		{string([]byte{0xe9}), input.Key{Char: 'i', Name: ""}},
		{"\x1b\x1b", input.Key{Name: "escape", Char: 0}},
		{"\r", input.Key{Name: "return", Char: 0}},
		{"\n", input.Key{Name: "return", Char: 0}},
		{"\x7f", input.Key{Name: "delete", Char: 0}},
		{"\x08", input.Key{Name: "backspace", Char: 0}}, // Windows backspace
		{"\t", input.Key{Name: "tab", Char: 0}},
		{" ", input.Key{Char: ' ', Name: ""}},
		{"q", input.Key{Char: 'q', Name: ""}},
		{"é", input.Key{Char: 'é', Name: ""}},
		{"🙂", input.Key{Char: '🙂', Name: ""}},
		{"漢", input.Key{Char: '漢', Name: ""}},
	}

	for _, tt := range tests {
		reader := strings.NewReader(tt.input)
		result, err := input.ReadKey(reader)
		if err != nil {
			t.Errorf("ReadKey(%q) error = %v", tt.input, err)
			continue
		}
		if result.Name != tt.expected.Name || result.Char != tt.expected.Char {
			t.Errorf("ReadKey(%q) = %+v, want %+v", tt.input, result, tt.expected)
		}
	}
}

func TestInputHandlerConsumesDoubleEscapeAsSingleEvent(t *testing.T) {
	reader := strings.NewReader("\x1b\x1bx")
	handler := input.NewInputHandler(reader)

	key, err := handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Name != input.KeyEscape {
		t.Fatalf("expected first key to be a single escape event, got %+v", key)
	}

	key, err = handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Char != 'x' {
		t.Fatalf("expected trailing character after meta+escape sequence, got %+v", key)
	}
}

// TestReadKeyMultiple tests reading multiple keys
func TestReadKeyMultiple(t *testing.T) {
	inputStr := "abc"
	baseReader := strings.NewReader(inputStr)

	expected := []input.Key{
		{Char: 'a', Name: ""},
		{Char: 'b', Name: ""},
		{Char: 'c', Name: ""},
	}

	// Create InputHandler to maintain the buffered reader
	handler := input.NewInputHandler(baseReader)

	for i, exp := range expected {
		result, err := handler.ReadKey()
		if err != nil {
			t.Errorf("ReadKey %d error = %v", i, err)
			continue
		}
		if result.Char != exp.Char {
			t.Errorf("ReadKey %d = %+v, want %+v", i, result, exp)
		}
	}
}

func TestInputHandlerConsumesExtendedEscapeSequences(t *testing.T) {
	reader := strings.NewReader("\x1bOAx")
	handler := input.NewInputHandler(reader)

	key, err := handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Name != input.KeyUp {
		t.Fatalf("expected first key to be up, got %+v", key)
	}

	key, err = handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Char != 'x' {
		t.Fatalf("expected trailing character after escape sequence, got %+v", key)
	}
}

func TestInputHandlerConsumesMetaEscapeSequences(t *testing.T) {
	reader := strings.NewReader("\x1bmx\x1b\x1b[Ay")
	handler := input.NewInputHandler(reader)

	key, err := handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Char != 'm' {
		t.Fatalf("expected meta+character to collapse to character, got %+v", key)
	}

	key, err = handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Char != 'x' {
		t.Fatalf("expected trailing character after meta+character, got %+v", key)
	}

	key, err = handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Name != input.KeyUp {
		t.Fatalf("expected meta+up-arrow to collapse to up, got %+v", key)
	}

	key, err = handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Char != 'y' {
		t.Fatalf("expected trailing character after meta+arrow, got %+v", key)
	}
}

func TestInputHandlerConsumesMetaUTF8EscapeSequences(t *testing.T) {
	reader := strings.NewReader("\x1béx")
	handler := input.NewInputHandler(reader)

	key, err := handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Char != 'é' {
		t.Fatalf("expected meta+UTF-8 character to collapse to character, got %+v", key)
	}

	key, err = handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key.Char != 'x' {
		t.Fatalf("expected trailing character after meta+UTF-8 sequence, got %+v", key)
	}
}

func TestInputHandlerConsumesPlainUTF8Runes(t *testing.T) {
	reader := strings.NewReader("é🙂漢x")
	handler := input.NewInputHandler(reader)

	expected := []input.Key{
		{Char: 'é'},
		{Char: '🙂'},
		{Char: '漢'},
		{Char: 'x'},
	}

	for index, want := range expected {
		key, err := handler.ReadKey()
		if err != nil {
			t.Fatalf("ReadKey %d failed: %v", index, err)
		}
		if key.Char != want.Char || key.Name != want.Name {
			t.Fatalf("ReadKey %d = %+v, want %+v", index, key, want)
		}
	}
}

// TestIsEscapeSequence tests escape sequence detection
func TestIsEscapeSequence(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"\x1b[A", true},
		{"\x1bOA", true},
		{"\x1b[a", true},
		{"\x1bOa", true},
		{"\x1b[B", true},
		{"\x1b[C", true},
		{"\x1b[D", true},
		{"\x1b", true}, // Single escape
		{"\x1b\r", true},
		{"\x1bm", true},
		{"\x1bX", true},
		{"\x1b\x1b[A", true},
		{"a", false},
		{"", false},
		{"\x1b[", false},
		{"\x1b[unknown", false},
	}

	for _, tt := range tests {
		result := input.IsEscapeSequence(tt.input)
		if result != tt.expected {
			t.Errorf("IsEscapeSequence(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// TestCtrlKeys tests control key combinations
func TestCtrlKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected input.Key
	}{
		{"\x01", input.Key{Name: "ctrl-a", Char: 0}}, // Ctrl+A
		{"\x02", input.Key{Name: "ctrl-b", Char: 0}}, // Ctrl+B
		{"\x03", input.Key{Name: "ctrl-c", Char: 0}}, // Ctrl+C
		{"\x04", input.Key{Name: "ctrl-d", Char: 0}}, // Ctrl+D
		{"\x05", input.Key{Name: "ctrl-e", Char: 0}}, // Ctrl+E
		{"\x1a", input.Key{Name: "ctrl-z", Char: 0}}, // Ctrl+Z
	}

	for _, tt := range tests {
		reader := strings.NewReader(tt.input)
		result, err := input.ReadKey(reader)
		if err != nil {
			t.Errorf("ReadKey(%q) error = %v", tt.input, err)
			continue
		}
		if result.Name != tt.expected.Name {
			t.Errorf("ReadKey(%q) = %+v, want %+v", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeHookInput(t *testing.T) {
	tests := []struct {
		raw          string
		expectedText string
		assertKey    func(t *testing.T, key input.HookKey)
		expectedKeys []string
	}{
		{
			raw:          "\r",
			expectedText: "\r",
			expectedKeys: []string{input.KeyReturn},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Return {
					t.Fatal("expected carriage return to set return=true")
				}
				if key.Shift {
					t.Fatal("expected carriage return not to set shift=true")
				}
			},
		},
		{
			raw:          "\rtest",
			expectedText: "\rtest",
			expectedKeys: nil,
		},
		{
			raw:          "\ttest",
			expectedText: "\ttest",
			expectedKeys: nil,
		},
		{
			raw:          "\x1b[200~hello\x1b[201~",
			expectedText: "hello",
			expectedKeys: nil,
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if key.Return || key.Tab || key.Delete || key.Backspace || key.Meta || key.Ctrl || key.Shift {
					t.Fatalf("expected bracketed paste payload not to synthesize key flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b",
			expectedText: "",
			expectedKeys: []string{input.KeyEscape},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Escape || !key.Meta || key.Ctrl || key.Return || key.Tab {
					t.Fatalf("expected bare escape to set escape=true and meta=true, got %+v", key)
				}
			},
		},
		{
			raw:          "A",
			expectedText: "A",
			expectedKeys: nil,
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Shift {
					t.Fatal("expected uppercase input to mark shift=true")
				}
			},
		},
		{
			raw:          "\x1b[A",
			expectedText: "",
			expectedKeys: []string{input.KeyUp},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.UpArrow {
					t.Fatal("expected up-arrow key flag")
				}
			},
		},
		{
			raw:          "\x1b[B",
			expectedText: "",
			expectedKeys: []string{input.KeyDown},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.DownArrow {
					t.Fatal("expected down-arrow key flag")
				}
			},
		},
		{
			raw:          "\x1b[C",
			expectedText: "",
			expectedKeys: []string{input.KeyRight},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.RightArrow {
					t.Fatal("expected right-arrow key flag")
				}
			},
		},
		{
			raw:          "\x1b[D",
			expectedText: "",
			expectedKeys: []string{input.KeyLeft},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.LeftArrow {
					t.Fatal("expected left-arrow key flag")
				}
			},
		},
		{
			raw:          "\x1b[6~",
			expectedText: "",
			expectedKeys: []string{input.KeyPageDown},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.PageDown {
					t.Fatal("expected pageDown flag")
				}
			},
		},
		{
			raw:          "\x1b[5~",
			expectedText: "",
			expectedKeys: []string{input.KeyPageUp},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.PageUp {
					t.Fatal("expected pageUp flag")
				}
			},
		},
		{
			raw:          "\x1b[[6~",
			expectedText: "",
			expectedKeys: []string{input.KeyPageDown},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.PageDown {
					t.Fatal("expected putty pageDown flag")
				}
			},
		},
		{
			raw:          "\x1b[[5~",
			expectedText: "",
			expectedKeys: []string{input.KeyPageUp},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.PageUp {
					t.Fatal("expected putty pageUp flag")
				}
			},
		},
		{
			raw:          "\x1b[H",
			expectedText: "",
			expectedKeys: []string{input.KeyHome},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Home {
					t.Fatal("expected home flag")
				}
			},
		},
		{
			raw:          "\x1b[F",
			expectedText: "",
			expectedKeys: []string{input.KeyEnd},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.End {
					t.Fatal("expected end flag")
				}
			},
		},
		{
			raw:          "\x1bOH",
			expectedText: "",
			expectedKeys: []string{input.KeyHome},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Home {
					t.Fatal("expected ss3 home flag")
				}
			},
		},
		{
			raw:          "\x1bOF",
			expectedText: "",
			expectedKeys: []string{input.KeyEnd},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.End {
					t.Fatal("expected ss3 end flag")
				}
			},
		},
		{
			raw:          "\x1bOA",
			expectedText: "",
			expectedKeys: []string{input.KeyUp},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.UpArrow {
					t.Fatal("expected ss3 up-arrow key flag")
				}
			},
		},
		{
			raw:          "\x1bOB",
			expectedText: "",
			expectedKeys: []string{input.KeyDown},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.DownArrow {
					t.Fatal("expected ss3 down-arrow key flag")
				}
			},
		},
		{
			raw:          "\x1b[c",
			expectedText: "",
			expectedKeys: []string{input.KeyRight},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.RightArrow || !key.Shift {
					t.Fatalf("expected shift+right-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1bOd",
			expectedText: "",
			expectedKeys: []string{"ctrl", input.KeyLeft},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.LeftArrow || !key.Ctrl {
					t.Fatalf("expected ctrl+left-arrow rxvt flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x03",
			expectedText: "c",
			expectedKeys: []string{"ctrl", "ctrl-c"},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Ctrl {
					t.Fatal("expected ctrl flag")
				}
			},
		},
		{
			raw:          "\x1b[Z",
			expectedText: "",
			expectedKeys: []string{input.KeyTab, "shift", input.KeyShiftTab},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Tab || !key.Shift {
					t.Fatal("expected shift-tab to set tab=true and shift=true")
				}
			},
		},
		{
			raw:          "\x1bm",
			expectedText: "m",
			expectedKeys: []string{"meta", "m"},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta {
					t.Fatal("expected meta+character to set meta=true")
				}
			},
		},
		{
			raw:          "\x1b[",
			expectedText: "[",
			expectedKeys: nil,
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if key.Meta || key.Ctrl || key.Return || key.Tab {
					t.Fatalf("expected ESC[ prefix flush to stay literal, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1bO",
			expectedText: "O",
			expectedKeys: []string{"meta", "O"},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta || !key.Shift || key.Ctrl || key.Return {
					t.Fatalf("expected meta+uppercase literal flags, got %+v", key)
				}
			},
		},
		{
			raw:          string([]byte{0xe9}),
			expectedText: "i",
			expectedKeys: []string{"meta", "i"},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta || key.Ctrl || key.Return || key.Delete {
					t.Fatalf("expected 8-bit meta byte to normalize to meta+i, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b\x1b[A",
			expectedText: "",
			expectedKeys: []string{"meta", input.KeyUp},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta || !key.UpArrow {
					t.Fatalf("expected meta+up-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b\x1b[B",
			expectedText: "",
			expectedKeys: []string{"meta", input.KeyDown},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta || !key.DownArrow {
					t.Fatalf("expected meta+down-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b\x1b[D",
			expectedText: "",
			expectedKeys: []string{"meta", input.KeyLeft},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta || !key.LeftArrow {
					t.Fatalf("expected meta+left-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b\x1b[C",
			expectedText: "",
			expectedKeys: []string{"meta", input.KeyRight},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta || !key.RightArrow {
					t.Fatalf("expected meta+right-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b\x1bOA",
			expectedText: "",
			expectedKeys: []string{"meta", input.KeyUp},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta || !key.UpArrow {
					t.Fatalf("expected meta+ss3 up-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b[1;5A",
			expectedText: "",
			expectedKeys: []string{"ctrl", input.KeyUp},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Ctrl || !key.UpArrow {
					t.Fatalf("expected ctrl+up-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b[1;5B",
			expectedText: "",
			expectedKeys: []string{"ctrl", input.KeyDown},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Ctrl || !key.DownArrow {
					t.Fatalf("expected ctrl+down-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b[1;5D",
			expectedText: "",
			expectedKeys: []string{"ctrl", input.KeyLeft},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Ctrl || !key.LeftArrow {
					t.Fatalf("expected ctrl+left-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b[1;5C",
			expectedText: "",
			expectedKeys: []string{"ctrl", input.KeyRight},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Ctrl || !key.RightArrow {
					t.Fatalf("expected ctrl+right-arrow flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b\r",
			expectedText: "\r",
			expectedKeys: []string{"meta", input.KeyReturn},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Meta || !key.Return {
					t.Fatalf("expected option+return flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b[1;5P",
			expectedText: "",
			expectedKeys: []string{"ctrl", input.KeyF1},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Ctrl || key.Meta || key.Return || key.Tab {
					t.Fatalf("expected ctrl+f1 flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1b[1;5I",
			expectedText: "",
			expectedKeys: nil,
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Ctrl || key.Meta || key.Return || key.Tab || key.UpArrow || key.DownArrow || key.LeftArrow || key.RightArrow {
					t.Fatalf("expected unmapped ctrl escape sequence to keep only ctrl=true, got %+v", key)
				}
			},
		},
		{
			raw:          "\b",
			expectedText: "",
			expectedKeys: []string{input.KeyBackspace},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Backspace {
					t.Fatal("expected backspace flag")
				}
			},
		},
		{
			raw:          "\x7f",
			expectedText: "",
			expectedKeys: []string{input.KeyDelete},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if key.Backspace || !key.Delete || key.Meta {
					t.Fatal("expected raw DEL byte to map to delete only")
				}
			},
		},
		{
			raw:          "\x1b\x7f",
			expectedText: "",
			expectedKeys: []string{"meta", input.KeyDelete},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if key.Backspace || !key.Delete || !key.Meta {
					t.Fatal("expected escaped DEL byte to map to meta+delete")
				}
			},
		},
		{
			raw:          "\x1b[3~",
			expectedText: "",
			expectedKeys: []string{input.KeyDelete},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Delete {
					t.Fatal("expected remove/delete escape sequence to set delete flag")
				}
			},
		},
		{
			raw:          "\x1b[2~",
			expectedText: "",
			expectedKeys: []string{input.KeyInsert},
		},
		{
			raw:          "\x1b[E",
			expectedText: "",
			expectedKeys: []string{input.KeyClear},
		},
		{
			raw:          "\x1b[e",
			expectedText: "",
			expectedKeys: []string{input.KeyClear},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Shift || key.Ctrl || key.Meta {
					t.Fatalf("expected shift+clear flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1bOe",
			expectedText: "",
			expectedKeys: []string{"ctrl", input.KeyClear},
			assertKey: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if !key.Ctrl || key.Shift || key.Meta {
					t.Fatalf("expected ctrl+clear flags, got %+v", key)
				}
			},
		},
		{
			raw:          "\x1bOP",
			expectedText: "",
			expectedKeys: []string{input.KeyF1},
		},
		{
			raw:          "\x1b[[E",
			expectedText: "",
			expectedKeys: []string{input.KeyF5},
		},
	}

	for _, tt := range tests {
		text, key, keys, err := input.NormalizeHookInput(tt.raw)
		if err != nil {
			t.Fatalf("NormalizeHookInput(%q) failed: %v", tt.raw, err)
		}

		if text != tt.expectedText {
			t.Fatalf("NormalizeHookInput(%q) text = %q, want %q", tt.raw, text, tt.expectedText)
		}

		if strings.Join(keys, ",") != strings.Join(tt.expectedKeys, ",") {
			t.Fatalf("NormalizeHookInput(%q) keys = %#v, want %#v", tt.raw, keys, tt.expectedKeys)
		}
		if tt.assertKey != nil {
			tt.assertKey(t, key)
		}
	}
}

// TestKeyEquals tests Key comparison
func TestKeyEquals(t *testing.T) {
	key1 := input.Key{Char: 'a', Name: ""}
	key2 := input.Key{Char: 'a', Name: ""}
	key3 := input.Key{Char: 'b', Name: ""}
	key4 := input.Key{Name: "escape", Char: 0}

	if !key1.Equals(key2) {
		t.Error("Key{a}.Equals(Key{a}) should be true")
	}
	if key1.Equals(key3) {
		t.Error("Key{a}.Equals(Key{b}) should be false")
	}
	if key1.Equals(key4) {
		t.Error("Key{a}.Equals(Key{escape}) should be false")
	}
}

// TestBufferedReader tests buffered reading with peek
func TestBufferedReader(t *testing.T) {
	input := "\x1b[A"
	reader := bytes.NewBufferString(input)

	// Peek should show escape byte
	buf := make([]byte, 1)
	n, err := reader.Read(buf)
	if err != nil || n != 1 {
		t.Fatalf("Read failed: %v", err)
	}

	if buf[0] != '\x1b' {
		t.Errorf("Expected escape byte, got %c", buf[0])
	}
}

// TestNewInputHandler tests InputHandler creation
func TestNewInputHandler(t *testing.T) {
	reader := strings.NewReader("test")
	handler := input.NewInputHandler(reader)

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
}

// TestInputHandlerReadKey tests reading keys from InputHandler
func TestInputHandlerReadKey(t *testing.T) {
	inputStr := "abc"
	reader := strings.NewReader(inputStr)
	handler := input.NewInputHandler(reader)

	key1, err := handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key1.Char != 'a' {
		t.Errorf("Expected 'a', got %c", key1.Char)
	}

	key2, err := handler.ReadKey()
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}
	if key2.Char != 'b' {
		t.Errorf("Expected 'b', got %c", key2.Char)
	}
}

// TestStdinHandler tests StdinHandler
func TestStdinHandler(t *testing.T) {
	callCount := 0
	var receivedKey input.Key

	handler := input.NewStdinHandler(func(key input.Key, keys []string) bool {
		callCount++
		receivedKey = key
		return false
	})

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}

	// Test HandleInput
	err := handler.HandleInput("a")
	if err != nil {
		t.Fatalf("HandleInput failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 callback call, got %d", callCount)
	}

	if receivedKey.Char != 'a' {
		t.Errorf("Expected key 'a', got %c", receivedKey.Char)
	}
}

// TestStdinHandlerExit tests StdinHandler with exit
func TestStdinHandlerExit(t *testing.T) {
	handler := input.NewStdinHandler(func(key input.Key, keys []string) bool {
		return true // Signal exit
	})

	err := handler.HandleInput("q")
	if err == nil {
		t.Error("Expected error when handler returns true")
	}
	if err.Error() != "exit requested" {
		t.Errorf("Expected 'exit requested' error, got %v", err)
	}
}

// TestStdinHandlerWithKeyTracking tests key tracking in StdinHandler
func TestStdinHandlerWithKeyTracking(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		expectedKeys []string
	}{
		{name: "plain", raw: "a", expectedKeys: nil},
		{name: "arrow", raw: "\x1b[A", expectedKeys: []string{input.KeyUp}},
		{name: "ctrl-c", raw: "\x03", expectedKeys: []string{"ctrl", "ctrl-c"}},
		{name: "shift-tab", raw: "\x1b[Z", expectedKeys: []string{input.KeyTab, "shift", input.KeyShiftTab}},
		{name: "meta-escape", raw: "\x1b\x1b", expectedKeys: []string{"meta", input.KeyEscape}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var receivedKeys []string

			handler := input.NewStdinHandler(func(key input.Key, keys []string) bool {
				receivedKeys = keys
				return false
			})

			if err := handler.HandleInput(testCase.raw); err != nil {
				t.Fatalf("HandleInput failed: %v", err)
			}
			if len(testCase.expectedKeys) == 0 {
				if receivedKeys != nil {
					t.Fatalf("expected nil keys for plain input, got %#v", receivedKeys)
				}
				return
			}
			if strings.Join(receivedKeys, ",") != strings.Join(testCase.expectedKeys, ",") {
				t.Fatalf("expected keys %#v, got %#v", testCase.expectedKeys, receivedKeys)
			}
		})
	}
}

func TestStdinHandlerCopiesKeySlicesAcrossInvocations(t *testing.T) {
	received := make([][]string, 0, 2)

	handler := input.NewStdinHandler(func(key input.Key, keys []string) bool {
		received = append(received, keys)
		return false
	})

	if err := handler.HandleInput("\x1b[Z"); err != nil {
		t.Fatalf("first HandleInput failed: %v", err)
	}
	if err := handler.HandleInput("\x03"); err != nil {
		t.Fatalf("second HandleInput failed: %v", err)
	}

	if len(received) != 2 {
		t.Fatalf("expected two handler calls, got %d", len(received))
	}
	if strings.Join(received[0], ",") != strings.Join([]string{input.KeyTab, "shift", input.KeyShiftTab}, ",") {
		t.Fatalf("expected first key snapshot to remain shift-tab, got %#v", received[0])
	}
	if strings.Join(received[1], ",") != strings.Join([]string{"ctrl", "ctrl-c"}, ",") {
		t.Fatalf("expected second key snapshot to be ctrl-c, got %#v", received[1])
	}
}

func TestStdinHandlerParsesMetaUTF8Characters(t *testing.T) {
	var receivedKey input.Key

	handler := input.NewStdinHandler(func(key input.Key, keys []string) bool {
		receivedKey = key
		return false
	})

	if err := handler.HandleInput("\x1bé"); err != nil {
		t.Fatalf("HandleInput failed: %v", err)
	}
	if receivedKey.Char != 'é' {
		t.Fatalf("expected meta+UTF-8 input to collapse to the character, got %+v", receivedKey)
	}
}

func TestStdinHandlerParsesPlainUTF8Characters(t *testing.T) {
	var receivedKey input.Key
	var receivedKeys []string

	handler := input.NewStdinHandler(func(key input.Key, keys []string) bool {
		receivedKey = key
		receivedKeys = keys
		return false
	})

	if err := handler.HandleInput("漢"); err != nil {
		t.Fatalf("HandleInput failed: %v", err)
	}
	if receivedKey.Char != '漢' {
		t.Fatalf("expected plain UTF-8 input to remain intact, got %+v", receivedKey)
	}
	if receivedKeys != nil {
		t.Fatalf("expected plain UTF-8 input not to synthesize keys, got %#v", receivedKeys)
	}
}

func TestStdinHandlerParses8BitMetaCharacters(t *testing.T) {
	var receivedKey input.Key
	var receivedKeys []string

	handler := input.NewStdinHandler(func(key input.Key, keys []string) bool {
		receivedKey = key
		receivedKeys = append([]string(nil), keys...)
		return false
	})

	if err := handler.HandleInput(string([]byte{0xe9})); err != nil {
		t.Fatalf("HandleInput failed: %v", err)
	}

	if receivedKey.Char != 'i' {
		t.Fatalf("expected 8-bit meta byte to collapse to 'i', got %+v", receivedKey)
	}
	if strings.Join(receivedKeys, ",") != "meta,i" {
		t.Fatalf("expected 8-bit meta byte keys to match meta+i, got %#v", receivedKeys)
	}
}

// TestDecodeUTF8Rune tests UTF-8 rune decoding
func TestDecodeUTF8Rune(t *testing.T) {
	tests := []struct {
		input        []byte
		expected     rune
		expectedSize int
	}{
		{[]byte{'a'}, 'a', 1},
		{[]byte{0xE2, 0x82, 0xAC}, '€', 3},       // Euro sign
		{[]byte{0xF0, 0x9F, 0x98, 0x80}, '😀', 4}, // Smiling face
	}

	for _, tt := range tests {
		r, size := input.DecodeUTF8Rune(tt.input)
		if r != tt.expected {
			t.Errorf("DecodeUTF8Rune(%v) = %c, want %c", tt.input, r, tt.expected)
		}
		if size != tt.expectedSize {
			t.Errorf("DecodeUTF8Rune(%v) size = %d, want %d", tt.input, size, tt.expectedSize)
		}
	}
}

// TestParseEscapeSequenceUnknown tests unknown escape sequences
func TestParseEscapeSequenceUnknown(t *testing.T) {
	// Unknown escape sequence should return escape key
	result := input.ParseEscapeSequence("\x1b[unknown")
	if result.Name != "escape" {
		t.Errorf("Expected 'escape' for unknown sequence, got %q", result.Name)
	}
}

// TestParseEscapeSequenceNonEscape tests non-escape input
func TestParseEscapeSequenceNonEscape(t *testing.T) {
	result := input.ParseEscapeSequence("abc")
	if result.Name != "" || result.Char != 0 {
		t.Errorf("Expected empty Key for non-escape input, got %+v", result)
	}
}

// TestKeyStringSpecialChar tests Key.String with special characters
func TestKeyStringSpecialChar(t *testing.T) {
	key := input.Key{Char: '\n', Name: ""}
	// Should return the character as string
	result := key.String()
	if result != "\n" {
		t.Errorf("Expected newline, got %q", result)
	}
}

// TestReadKeyEOF tests reading from empty input
func TestReadKeyEOF(t *testing.T) {
	reader := strings.NewReader("")
	_, err := input.ReadKey(reader)
	if err == nil {
		t.Error("Expected error when reading from empty input")
	}
}

// TestInputHandlerEOF tests InputHandler with EOF
func TestInputHandlerEOF(t *testing.T) {
	reader := strings.NewReader("")
	handler := input.NewInputHandler(reader)

	_, err := handler.ReadKey()
	if err == nil {
		t.Error("Expected error when reading from empty input")
	}
}

// TestHandleInputError tests HandleInput with invalid input
func TestHandleInputError(t *testing.T) {
	handler := input.NewStdinHandler(func(key input.Key, keys []string) bool {
		return false
	})

	// Empty input should cause an error
	err := handler.HandleInput("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestNormalizeHookInputKeepsBackspaceAndDeleteDistinct(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantText string
		wantKeys string
		assert   func(t *testing.T, key input.HookKey)
	}{
		{
			name:     "del-byte-is-delete",
			raw:      "\x7f",
			wantText: "",
			wantKeys: "delete",
			assert: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if key.Backspace || !key.Delete || key.Meta {
					t.Fatalf("expected raw DEL byte to map to delete only, got %+v", key)
				}
			},
		},
		{
			name:     "meta-del-byte-is-meta-delete",
			raw:      "\x1b\x7f",
			wantText: "",
			wantKeys: "meta,delete",
			assert: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if key.Backspace || !key.Delete || !key.Meta {
					t.Fatalf("expected escaped DEL byte to map to meta+delete, got %+v", key)
				}
			},
		},
		{
			name:     "remove-sequence-stays-delete",
			raw:      "\x1b[3~",
			wantText: "",
			wantKeys: "delete",
			assert: func(t *testing.T, key input.HookKey) {
				t.Helper()
				if key.Backspace || !key.Delete || key.Meta {
					t.Fatalf("expected remove sequence to stay delete-only, got %+v", key)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			text, key, keys, err := input.NormalizeHookInput(testCase.raw)
			if err != nil {
				t.Fatalf("NormalizeHookInput(%q) failed: %v", testCase.raw, err)
			}
			if text != testCase.wantText {
				t.Fatalf("expected input text %q, got %q", testCase.wantText, text)
			}
			if strings.Join(keys, ",") != testCase.wantKeys {
				t.Fatalf("expected keys %q, got %#v", testCase.wantKeys, keys)
			}
			testCase.assert(t, key)
		})
	}
}
