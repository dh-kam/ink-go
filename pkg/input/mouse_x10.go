package input

import (
	"errors"
	"fmt"
	"strings"
)

// Legacy X10 mouse reporting (DECSET 9 / 1000 without 1006).
//
// Wire format: ESC '[' 'M' Cb Cx Cy
//
//	Cb = button-state byte + 32
//	Cx = x coordinate     + 32   (1-based column)
//	Cy = y coordinate     + 32   (1-based row)
//
// The +32 bias keeps every byte inside the printable ASCII range so the
// stream never collides with C0 controls. Each field is therefore limited
// to values 0..223 in practice (255-32), which is why xterm later added the
// SGR 1006 extension for larger terminals.
//
// Cb bit layout matches the SGR scheme except that the terminator is always
// 'M' (no separate release byte): bits 0-1 carry the button, with the value
// 0b11 (3) used to indicate a release where the originating button is
// unknown.

const (
	x10Prefix    = "\x1b[M"
	x10TotalLen  = 6 // ESC + '[' + 'M' + Cb + Cx + Cy
	x10ByteBias  = 32
)

// IsX10MouseSequence reports whether s is a well-formed legacy X10 mouse
// report: exactly 6 bytes beginning with ESC '[' 'M'.
//
// We deliberately require the exact length here rather than accepting any
// string with the prefix; an X10 frame is fixed-size and a longer or
// shorter buffer is by definition something else (often a partial read or
// an SGR sequence that happens to share two of the prefix bytes).
func IsX10MouseSequence(s string) bool {
	if len(s) != x10TotalLen {
		return false
	}
	return strings.HasPrefix(s, x10Prefix)
}

// ParseX10Mouse decodes a single legacy X10 (DECSET 1000) mouse report.
//
// All three data bytes are biased by 32 on the wire; we subtract that
// before decoding. Coordinates are 1-based and use the same MouseEvent
// shape as ParseSGRMouse so downstream consumers can stay agnostic to the
// underlying transport.
//
// Limitations of the X10 frame (not bugs in this parser):
//   - Releases always report button-bits == 3, so the *originating* button
//     for a release cannot be recovered. We surface this as
//     {Button: MouseNone, Action: MouseActionRelease}.
//   - Coordinates above 223 cannot be encoded; values that decode to
//     negatives are rejected.
func ParseX10Mouse(s string) (MouseEvent, error) {
	if len(s) != x10TotalLen {
		return MouseEvent{}, fmt.Errorf("ParseX10Mouse: expected %d bytes, got %d", x10TotalLen, len(s))
	}
	if !strings.HasPrefix(s, x10Prefix) {
		return MouseEvent{}, errors.New("ParseX10Mouse: missing ESC[M prefix")
	}

	cb := int(s[3]) - x10ByteBias
	cx := int(s[4]) - x10ByteBias
	cy := int(s[5]) - x10ByteBias

	if cb < 0 {
		return MouseEvent{}, fmt.Errorf("ParseX10Mouse: button byte 0x%02x below bias", s[3])
	}
	if cx < 0 {
		return MouseEvent{}, fmt.Errorf("ParseX10Mouse: x byte 0x%02x below bias", s[4])
	}
	if cy < 0 {
		return MouseEvent{}, fmt.Errorf("ParseX10Mouse: y byte 0x%02x below bias", s[5])
	}

	ev := MouseEvent{X: cx, Y: cy}
	ev.Mods = Modifiers{
		Shift: cb&mouseBitShift != 0,
		Alt:   cb&mouseBitAlt != 0,
		Ctrl:  cb&mouseBitCtrl != 0,
	}

	switch {
	case cb&mouseBitWheel != 0:
		ev.Action = MouseActionWheel
		// Same convention as SGR: low bit picks up vs. down.
		if cb&1 == 0 {
			ev.Button = MouseWheelUp
		} else {
			ev.Button = MouseWheelDown
		}
	case cb&mouseBitMotion != 0:
		buttonBits := cb & mouseBitButtonMask
		ev.Button = decodeButtonBits(buttonBits)
		if ev.Button == MouseNone {
			ev.Action = MouseActionMove
		} else {
			ev.Action = MouseActionDrag
		}
	default:
		buttonBits := cb & mouseBitButtonMask
		// In X10, button bits == 3 always means "some button was released"
		// without telling us which one. SGR uses a separate 'm' terminator
		// to convey the same thing, so it can preserve the button identity.
		if buttonBits == sgrButtonNone {
			ev.Button = MouseNone
			ev.Action = MouseActionRelease
		} else {
			ev.Button = decodeButtonBits(buttonBits)
			ev.Action = MouseActionPress
		}
	}

	return ev, nil
}
