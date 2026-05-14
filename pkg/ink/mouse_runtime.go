package ink

import (
	"strings"

	"github.com/dh-kam/ink-go/pkg/hooks"
	"github.com/dh-kam/ink-go/pkg/input"
)

// routeMouseInput inspects raw stdin data, parses SGR 1006 or legacy X10
// mouse frames, and fans the resulting MouseEvent out through the global
// compatibility hooks.DispatchMouse pipeline. Returns true only when a mouse
// subscriber received the event so unhandled mouse-looking bytes can still
// flow through useInput.
//
// Lives outside session.go so the session keeps a one-line hook into the
// mouse subsystem instead of carrying SGR parsing logic itself.
func routeMouseInput(data string) bool {
	return routeMouseInputWithManager(data, nil)
}

func routeMouseInputWithManager(data string, manager *hooks.MouseManager) bool {
	consumed, _ := consumeMouseFramesWithManager(data, manager)
	return consumed
}

// consumeMouseFramesWithManager peels mouse frames (SGR 1006 or legacy X10)
// from the front of data and dispatches each through the supplied manager
// (or, when manager is nil, the global hooks.DispatchMouse pipeline).
// Returns (anyDispatched, leftover): leftover holds the suffix of data
// after the last consumed frame so callers can pass it on to keyboard
// handling. anyDispatched mirrors the legacy single-frame return contract.
//
// The peel-from-front design handles the realistic case where a TTY raw
// read coalesces a burst of mouse-move frames, or stitches a mouse frame
// onto a subsequent keypress, into a single chunk. Without it, a 12-byte
// chunk containing two X10 frames would silently fail length checks and
// neither event would dispatch.
func consumeMouseFramesWithManager(data string, manager *hooks.MouseManager) (bool, string) {
	dispatched := false
	cursor := 0
	for cursor < len(data) {
		frameLen, ev, ok := parseLeadingMouseFrame(data[cursor:])
		if !ok {
			break
		}
		if dispatchMouseInput(ev, manager) {
			dispatched = true
		}
		cursor += frameLen
	}
	return dispatched, data[cursor:]
}

// parseLeadingMouseFrame inspects the start of data for a single SGR 1006
// or legacy X10 mouse frame. On success returns the byte length consumed
// and the decoded MouseEvent. On failure returns ok=false; callers should
// not advance the cursor.
func parseLeadingMouseFrame(data string) (int, input.MouseEvent, bool) {
	if len(data) == 0 {
		return 0, input.MouseEvent{}, false
	}

	// SGR 1006: ESC [ < ... M|m — variable length, terminated by 'M' or 'm'.
	if strings.HasPrefix(data, "\x1b[<") {
		// Find the terminator without scanning past a second escape that
		// would belong to the next sequence. SGR fields are ASCII digits
		// and ';', so any ESC ahead of the terminator means the frame is
		// malformed and we must not consume it.
		for i := 3; i < len(data); i++ {
			ch := data[i]
			if ch == '\x1b' {
				return 0, input.MouseEvent{}, false
			}
			if ch == 'M' || ch == 'm' {
				frame := data[:i+1]
				ev, err := input.ParseSGRMouse(frame)
				if err != nil {
					return 0, input.MouseEvent{}, false
				}
				return i + 1, ev, true
			}
		}
		return 0, input.MouseEvent{}, false
	}

	// Legacy X10: ESC [ M Cb Cx Cy — fixed 6 bytes total.
	if strings.HasPrefix(data, "\x1b[M") && len(data) >= 6 {
		ev, err := input.ParseX10Mouse(data[:6])
		if err != nil {
			return 0, input.MouseEvent{}, false
		}
		return 6, ev, true
	}

	return 0, input.MouseEvent{}, false
}

func dispatchMouseInput(ev input.MouseEvent, manager *hooks.MouseManager) bool {
	if manager != nil {
		return manager.Dispatch(ev)
	}
	return hooks.DispatchMouse(ev)
}
