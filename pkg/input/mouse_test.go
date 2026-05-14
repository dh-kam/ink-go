package input_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/input"
)

func TestIsSGRMouseSequence(t *testing.T) {
	cases := map[string]bool{
		"\x1b[<0;1;1M":  true,
		"\x1b[<2;5;3m":  true,
		"\x1b[<64;1;1M": true,
		"\x1b[0;1;1M":   false, // missing <
		"random":        false,
		"":              false,
		"\x1b[<":        false,
	}
	for s, want := range cases {
		if got := input.IsSGRMouseSequence(s); got != want {
			t.Errorf("IsSGRMouseSequence(%q) = %v, want %v", s, got, want)
		}
	}
}

func TestParseSGRMouseLeftPress(t *testing.T) {
	ev, err := input.ParseSGRMouse("\x1b[<0;10;20M")
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

func TestParseSGRMouseMiddlePress(t *testing.T) {
	ev, _ := input.ParseSGRMouse("\x1b[<1;5;5M")
	if ev.Button != input.MouseMiddle || ev.Action != input.MouseActionPress {
		t.Errorf("got %v/%v, want middle/press", ev.Button, ev.Action)
	}
}

func TestParseSGRMouseRightPress(t *testing.T) {
	ev, _ := input.ParseSGRMouse("\x1b[<2;5;5M")
	if ev.Button != input.MouseRight || ev.Action != input.MouseActionPress {
		t.Errorf("got %v/%v, want right/press", ev.Button, ev.Action)
	}
}

func TestParseSGRMouseRelease(t *testing.T) {
	ev, _ := input.ParseSGRMouse("\x1b[<0;5;5m")
	if ev.Action != input.MouseActionRelease {
		t.Errorf("got action %v, want release", ev.Action)
	}
	if ev.Button != input.MouseLeft {
		t.Errorf("got button %v, want left", ev.Button)
	}
}

func TestParseSGRMouseDrag(t *testing.T) {
	// Motion bit (32) + button 0 = left-drag (b=32)
	ev, _ := input.ParseSGRMouse("\x1b[<32;5;5M")
	if ev.Action != input.MouseActionDrag {
		t.Errorf("got action %v, want drag", ev.Action)
	}
	if ev.Button != input.MouseLeft {
		t.Errorf("got button %v, want left", ev.Button)
	}
}

func TestParseSGRMouseMoveNoButton(t *testing.T) {
	// Motion bit + button bits = 11 (sgrButtonNone), b = 32 + 3 = 35
	ev, _ := input.ParseSGRMouse("\x1b[<35;5;5M")
	if ev.Action != input.MouseActionMove {
		t.Errorf("got action %v, want move", ev.Action)
	}
	if ev.Button != input.MouseNone {
		t.Errorf("got button %v, want none", ev.Button)
	}
}

func TestParseSGRMouseWheelUp(t *testing.T) {
	ev, _ := input.ParseSGRMouse("\x1b[<64;1;1M")
	if ev.Action != input.MouseActionWheel || ev.Button != input.MouseWheelUp {
		t.Errorf("got %v/%v, want wheel/wheelUp", ev.Action, ev.Button)
	}
}

func TestParseSGRMouseWheelDown(t *testing.T) {
	ev, _ := input.ParseSGRMouse("\x1b[<65;1;1M")
	if ev.Button != input.MouseWheelDown {
		t.Errorf("got button %v, want wheelDown", ev.Button)
	}
}

func TestParseSGRMouseModifiers(t *testing.T) {
	// b = 0 (left) + 4 (shift) + 8 (alt) + 16 (ctrl) = 28
	ev, _ := input.ParseSGRMouse("\x1b[<28;1;1M")
	if !ev.Mods.Shift || !ev.Mods.Alt || !ev.Mods.Ctrl {
		t.Errorf("mods = %+v, want all true", ev.Mods)
	}
}

func TestParseSGRMouseShiftOnly(t *testing.T) {
	ev, _ := input.ParseSGRMouse("\x1b[<4;1;1M")
	if !ev.Mods.Shift || ev.Mods.Alt || ev.Mods.Ctrl {
		t.Errorf("mods = %+v, want shift-only", ev.Mods)
	}
}

func TestParseSGRMouseErrors(t *testing.T) {
	cases := []string{
		"",
		"random",
		"\x1b[<",                  // too short
		"\x1b[<0;1;1X",            // bad terminator
		"\x1b[<0;1M",              // 2 fields
		"\x1b[<0;1;2;3M",          // 4 fields
		"\x1b[<abc;1;1M",          // non-numeric button
		"\x1b[<0;abc;1M",          // non-numeric x
		"\x1b[<0;1;abcM",          // non-numeric y
		"\x1b[<-1;1;1M",           // negative button
		"\x1b[<0;-1;1M",           // negative x
		"\x1b[<0;1;-1M",           // negative y
	}
	for _, s := range cases {
		if _, err := input.ParseSGRMouse(s); err == nil {
			t.Errorf("ParseSGRMouse(%q) expected error", s)
		}
	}
}

func TestMouseButtonString(t *testing.T) {
	got := input.MouseLeft.String()
	if got != "left" {
		t.Errorf("MouseLeft.String() = %q, want left", got)
	}
	if !strings.Contains(input.MouseButton(99).String(), "button") {
		t.Errorf("unknown button missing fallback")
	}
}

func TestMouseActionString(t *testing.T) {
	if input.MouseActionPress.String() != "press" {
		t.Errorf("press action mislabeled")
	}
	if !strings.Contains(input.MouseAction(99).String(), "action") {
		t.Errorf("unknown action missing fallback")
	}
}
