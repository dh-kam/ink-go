package input

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// MouseButton enumerates the buttons reported by xterm SGR (1006) mouse mode.
type MouseButton int

const (
	// MouseNone covers motion-only events (no button held).
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
)

func (b MouseButton) String() string {
	switch b {
	case MouseNone:
		return "none"
	case MouseLeft:
		return "left"
	case MouseMiddle:
		return "middle"
	case MouseRight:
		return "right"
	case MouseWheelUp:
		return "wheelUp"
	case MouseWheelDown:
		return "wheelDown"
	default:
		return fmt.Sprintf("button(%d)", int(b))
	}
}

// MouseAction distinguishes a press from a release / drag / move / wheel.
type MouseAction int

const (
	MouseActionPress MouseAction = iota
	MouseActionRelease
	MouseActionMove
	MouseActionDrag
	MouseActionWheel
)

func (a MouseAction) String() string {
	switch a {
	case MouseActionPress:
		return "press"
	case MouseActionRelease:
		return "release"
	case MouseActionMove:
		return "move"
	case MouseActionDrag:
		return "drag"
	case MouseActionWheel:
		return "wheel"
	default:
		return fmt.Sprintf("action(%d)", int(a))
	}
}

// Modifiers reports which modifier keys were active when the event fired.
type Modifiers struct {
	Shift bool
	Alt   bool
	Ctrl  bool
}

// MouseEvent describes a single decoded mouse event. Coordinates are
// 1-based, matching the SGR wire format.
type MouseEvent struct {
	X, Y   int
	Button MouseButton
	Action MouseAction
	Mods   Modifiers
}

// SGR mouse mode bit layout (DECSET 1006).
const (
	mouseBitButtonMask = 0b0000_0011 // bits 0-1
	mouseBitShift      = 1 << 2
	mouseBitAlt        = 1 << 3
	mouseBitCtrl       = 1 << 4
	mouseBitMotion     = 1 << 5
	mouseBitWheel      = 1 << 6
)

// SGR press value 0b11 (3) is reserved for "no button" / release-only when
// reported via the legacy X10 stream, but in SGR mode releases use the
// terminator 'm' instead.
const sgrButtonNone = 3

// SGR sequence prefix and terminators.
const (
	sgrPrefix      = "\x1b[<"
	sgrTermPress   = byte('M')
	sgrTermRelease = byte('m')
)

// IsSGRMouseSequence reports whether s looks like an SGR mouse report
// (prefix \x1b[< and terminator M or m).
func IsSGRMouseSequence(s string) bool {
	if !strings.HasPrefix(s, sgrPrefix) {
		return false
	}
	if len(s) < len(sgrPrefix)+1 {
		return false
	}
	last := s[len(s)-1]
	return last == sgrTermPress || last == sgrTermRelease
}

// ParseSGRMouse decodes a single SGR (1006) mouse sequence into a
// MouseEvent. Format: ESC [ < b ; x ; y M|m
func ParseSGRMouse(s string) (MouseEvent, error) {
	if !strings.HasPrefix(s, sgrPrefix) {
		return MouseEvent{}, errors.New("ParseSGRMouse: missing ESC[< prefix")
	}
	if len(s) < len(sgrPrefix)+1 {
		return MouseEvent{}, errors.New("ParseSGRMouse: input too short")
	}

	terminator := s[len(s)-1]
	if terminator != sgrTermPress && terminator != sgrTermRelease {
		return MouseEvent{}, fmt.Errorf("ParseSGRMouse: bad terminator %q", terminator)
	}
	released := terminator == sgrTermRelease

	body := s[len(sgrPrefix) : len(s)-1]
	parts := strings.Split(body, ";")
	if len(parts) != 3 {
		return MouseEvent{}, fmt.Errorf("ParseSGRMouse: expected 3 fields, got %d (%q)", len(parts), body)
	}

	b, err := strconv.Atoi(parts[0])
	if err != nil {
		return MouseEvent{}, fmt.Errorf("ParseSGRMouse: button %q: %w", parts[0], err)
	}
	if b < 0 {
		return MouseEvent{}, fmt.Errorf("ParseSGRMouse: negative button %d", b)
	}
	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return MouseEvent{}, fmt.Errorf("ParseSGRMouse: x %q: %w", parts[1], err)
	}
	if x < 0 {
		return MouseEvent{}, fmt.Errorf("ParseSGRMouse: negative x %d", x)
	}
	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return MouseEvent{}, fmt.Errorf("ParseSGRMouse: y %q: %w", parts[2], err)
	}
	if y < 0 {
		return MouseEvent{}, fmt.Errorf("ParseSGRMouse: negative y %d", y)
	}

	ev := MouseEvent{X: x, Y: y}
	ev.Mods = Modifiers{
		Shift: b&mouseBitShift != 0,
		Alt:   b&mouseBitAlt != 0,
		Ctrl:  b&mouseBitCtrl != 0,
	}

	switch {
	case b&mouseBitWheel != 0:
		ev.Action = MouseActionWheel
		// Lower bit of the masked button code distinguishes up vs. down.
		if b&1 == 0 {
			ev.Button = MouseWheelUp
		} else {
			ev.Button = MouseWheelDown
		}
	case b&mouseBitMotion != 0:
		buttonBits := b & mouseBitButtonMask
		ev.Button = decodeButtonBits(buttonBits)
		if ev.Button == MouseNone {
			ev.Action = MouseActionMove
		} else {
			ev.Action = MouseActionDrag
		}
	default:
		buttonBits := b & mouseBitButtonMask
		ev.Button = decodeButtonBits(buttonBits)
		if released {
			ev.Action = MouseActionRelease
		} else {
			ev.Action = MouseActionPress
		}
	}

	return ev, nil
}

func decodeButtonBits(bits int) MouseButton {
	switch bits {
	case 0:
		return MouseLeft
	case 1:
		return MouseMiddle
	case 2:
		return MouseRight
	case sgrButtonNone:
		return MouseNone
	}
	return MouseNone
}
