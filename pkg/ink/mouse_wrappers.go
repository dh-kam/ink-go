package ink

import (
	"github.com/dh-kam/ink-go/pkg/hooks"
	"github.com/dh-kam/ink-go/pkg/input"
	"github.com/dh-kam/ink-go/pkg/terminal"
)

// MouseEvent re-exports input.MouseEvent so consumers don't have to import
// pkg/input alongside pkg/ink.
type MouseEvent = input.MouseEvent

// MouseButton, MouseAction, Modifiers re-exports.
type (
	MouseButton = input.MouseButton
	MouseAction = input.MouseAction
	Modifiers   = input.Modifiers
)

// MouseHandler is the callback shape registered via UseMouse.
type MouseHandler = hooks.MouseCallback

// UseMouse subscribes to mouse events for the current app. The terminal
// is switched into SGR mouse reporting on first mount and switched back on
// the final unmount.
//
// Multiple components calling UseMouse independently each get their own
// callback registration. The DECSET enable/disable sequences are written
// inside a UseEffect so they ride the normal mount/unmount lifecycle.
func UseMouse(handler MouseHandler) func() {
	app := requireCurrentApp("UseMouse")
	_ = requireHooksContext("UseMouse")

	UseEffect(func() func() {
		_ = terminal.EnableMouseTo(app.Stdout())
		dereg := app.mouseManager.UseMouse(handler)
		return func() {
			dereg()
			_ = terminal.DisableMouseTo(app.Stdout())
		}
	}, []interface{}{"use-mouse"})

	// Return a no-op cleanup — the UseEffect cleanup handles deregistration
	// at unmount. Returned for API symmetry with hooks.UseMouse.
	return func() {}
}
