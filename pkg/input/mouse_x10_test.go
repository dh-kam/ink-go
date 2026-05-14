package input_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/input"
)

// x10 builds a legacy X10 mouse frame. cb/cx/cy are the *unbiased* values;
// the +32 wire bias is applied here so individual tests stay readable.
func x10(cb, cx, cy int) string {
	return "\x1b[M" + string([]byte{byte(cb + 32), byte(cx + 32), byte(cy + 32)})
}

func TestIsX10MouseSequence(t *testing.T) {
	cases := map[string]bool{
		x10(0, 1, 1):         true,
		x10(2, 80, 24):       true,
		x10(64, 1, 1):        true,
		"":                   false,
		"\x1b[M":             false, // prefix only
		"\x1b[M\x20":         false, // 1 data byte
		"\x1b[M\x20\x20":     false, // 2 data bytes
		"\x1b[M\x20\x20\x20\x20": false, // too long
		"\x1b[<0;1;1M":       false, // SGR, not X10
		"random":             false,
	}
	for s, want := range cases {
		if got := input.IsX10MouseSequence(s); got != want {
			t.Errorf("IsX10MouseSequence(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestParseX10MouseLeftPress(t *testing.T) {
	ev, err := input.ParseX10Mouse(x10(0, 10, 20))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Button != input.MouseLeft || ev.Action != input.MouseActionPress {
		t.Errorf("button/action = %v/%v, want left/press", ev.Button, ev.Action)
	}
	if ev.X != 10 || ev.Y != 20 {
		t.Errorf("X,Y = %d,%d, want 10,20", ev.X, ev.Y)
	}
	if ev.Mods != (input.Modifiers{}) {
		t.Errorf("Mods = %+v, want zero", ev.Mods)
	}
}

func TestParseX10MouseMiddlePress(t *testing.T) {
	ev, err := input.ParseX10Mouse(x10(1, 5, 5))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Button != input.MouseMiddle || ev.Action != input.MouseActionPress {
		t.Errorf("got %v/%v, want middle/press", ev.Button, ev.Action)
	}
}

func TestParseX10MouseRightPress(t *testing.T) {
	ev, err := input.ParseX10Mouse(x10(2, 5, 5))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Button != input.MouseRight || ev.Action != input.MouseActionPress {
		t.Errorf("got %v/%v, want right/press", ev.Button, ev.Action)
	}
}

func TestParseX10MouseRelease(t *testing.T) {
	// X10 cannot tell us which button was released — bits == 3.
	ev, err := input.ParseX10Mouse(x10(3, 5, 5))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Action != input.MouseActionRelease {
		t.Errorf("got action %v, want release", ev.Action)
	}
	if ev.Button != input.MouseNone {
		t.Errorf("got button %v, want none (X10 cannot recover release button)", ev.Button)
	}
}

func TestParseX10MouseWheelUp(t *testing.T) {
	// Wheel bit (64) + low bit 0 = wheel up; cb = 64.
	ev, err := input.ParseX10Mouse(x10(64, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Action != input.MouseActionWheel || ev.Button != input.MouseWheelUp {
		t.Errorf("got %v/%v, want wheel/wheelUp", ev.Action, ev.Button)
	}
}

func TestParseX10MouseWheelDown(t *testing.T) {
	// Wheel bit (64) + low bit 1 = wheel down; cb = 65.
	ev, err := input.ParseX10Mouse(x10(65, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Button != input.MouseWheelDown {
		t.Errorf("got button %v, want wheelDown", ev.Button)
	}
}

func TestParseX10MouseDrag(t *testing.T) {
	// Motion bit (32) + button 0 (left) = left-drag; cb = 32.
	ev, err := input.ParseX10Mouse(x10(32, 5, 5))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Action != input.MouseActionDrag {
		t.Errorf("got action %v, want drag", ev.Action)
	}
	if ev.Button != input.MouseLeft {
		t.Errorf("got button %v, want left", ev.Button)
	}
}

func TestParseX10MouseMoveNoButton(t *testing.T) {
	// Motion bit + button-bits 3 (none) = pure move; cb = 35.
	ev, err := input.ParseX10Mouse(x10(35, 5, 5))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Action != input.MouseActionMove {
		t.Errorf("got action %v, want move", ev.Action)
	}
	if ev.Button != input.MouseNone {
		t.Errorf("got button %v, want none", ev.Button)
	}
}

func TestParseX10MouseAllModifiers(t *testing.T) {
	// cb = 0 (left) + 4 (shift) + 8 (alt) + 16 (ctrl) = 28.
	ev, err := input.ParseX10Mouse(x10(28, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Mods.Shift || !ev.Mods.Alt || !ev.Mods.Ctrl {
		t.Errorf("mods = %+v, want all true", ev.Mods)
	}
	if ev.Button != input.MouseLeft || ev.Action != input.MouseActionPress {
		t.Errorf("got %v/%v, want left/press", ev.Button, ev.Action)
	}
}

func TestParseX10MouseShiftOnly(t *testing.T) {
	ev, err := input.ParseX10Mouse(x10(4, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Mods.Shift || ev.Mods.Alt || ev.Mods.Ctrl {
		t.Errorf("mods = %+v, want shift-only", ev.Mods)
	}
}

func TestParseX10MouseCtrlWheel(t *testing.T) {
	// Ctrl + wheel-up: cb = 64 | 16 = 80.
	ev, err := input.ParseX10Mouse(x10(80, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Button != input.MouseWheelUp || !ev.Mods.Ctrl {
		t.Errorf("got button=%v ctrl=%v, want wheelUp + ctrl", ev.Button, ev.Mods.Ctrl)
	}
}

func TestParseX10MouseMaxCoord(t *testing.T) {
	// Highest coordinate that still fits in an unsigned byte after the +32
	// bias: 255 - 32 = 223. Anything larger is unrepresentable in X10.
	ev, err := input.ParseX10Mouse(x10(0, 223, 223))
	if err != nil {
		t.Fatal(err)
	}
	if ev.X != 223 || ev.Y != 223 {
		t.Errorf("X,Y = %d,%d, want 223,223", ev.X, ev.Y)
	}
}

func TestParseX10MouseErrors(t *testing.T) {
	cases := map[string]string{
		"empty":       "",
		"prefix-only": "\x1b[M",
		"too-short":   "\x1b[M\x20\x20",
		"too-long":    "\x1b[M\x20\x20\x20\x20",
		"bad-prefix":  "\x1b[<\x20\x20\x20",
		"sgr-shape":   "\x1b[<0;1;1M",
		// Bytes below the +32 bias should be rejected. We hand-craft the
		// frame here so the helper's int->byte conversion doesn't mask the
		// underflow.
		"button-underflow": "\x1b[M\x00\x21\x21",
		"x-underflow":      "\x1b[M\x21\x00\x21",
		"y-underflow":      "\x1b[M\x21\x21\x00",
	}
	for name, s := range cases {
		if _, err := input.ParseX10Mouse(s); err == nil {
			t.Errorf("ParseX10Mouse(%s=%q) expected error", name, s)
		}
	}
}

// Sanity check: the X10 helper output really is the shape IsX10MouseSequence
// recognises. Catches accidental drift in the test helper.
func TestX10HelperRoundTrip(t *testing.T) {
	frame := x10(0, 1, 1)
	if !strings.HasPrefix(frame, "\x1b[M") {
		t.Fatalf("helper produced bad prefix: %q", frame)
	}
	if !input.IsX10MouseSequence(frame) {
		t.Fatalf("helper output not recognised as X10: %q", frame)
	}
}
