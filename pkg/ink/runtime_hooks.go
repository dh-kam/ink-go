package ink

import "io"

// AppContext exposes app-level lifecycle controls.
type AppContext struct {
	app *App
}

// Exit requests that the current app exit, optionally storing an error.
func (ctx AppContext) Exit(err ...error) {
	if ctx.app != nil {
		ctx.app.Exit(err...)
	}
}

// UseApp exposes app lifecycle controls for the currently rendering app.
func UseApp() AppContext {
	return AppContext{app: requireCurrentApp("UseApp")}
}

// StdinContext exposes stdin-related runtime state.
type StdinContext struct {
	app                *App
	Stdin              io.Reader
	IsRawModeSupported bool
}

// SetRawMode toggles raw mode on the current app stdin.
func (ctx StdinContext) SetRawMode(enabled bool) error {
	if ctx.app == nil {
		return nil
	}

	return ctx.app.SetRawMode(enabled)
}

// UseStdin exposes the stdin stream and raw-mode helpers.
func UseStdin() StdinContext {
	app := requireCurrentApp("UseStdin")
	return StdinContext{
		app:                app,
		Stdin:              app.Stdin(),
		IsRawModeSupported: app.IsRawModeSupported(),
	}
}

// StdoutContext exposes stdout-related runtime state.
type StdoutContext struct {
	app    *App
	Stdout io.Writer
}

// Write writes directly to the current app stdout stream.
func (ctx StdoutContext) Write(data string) (int, error) {
	if ctx.app == nil {
		return 0, nil
	}

	return ctx.app.WriteStdout(data)
}

// UseStdout exposes the stdout stream and write helper.
func UseStdout() StdoutContext {
	app := requireCurrentApp("UseStdout")
	return StdoutContext{
		app:    app,
		Stdout: app.Stdout(),
	}
}

// StderrContext exposes stderr-related runtime state.
type StderrContext struct {
	app    *App
	Stderr io.Writer
}

// Write writes directly to the current app stderr stream.
func (ctx StderrContext) Write(data string) (int, error) {
	if ctx.app == nil {
		return 0, nil
	}

	return ctx.app.WriteStderr(data)
}

// UseStderr exposes the stderr stream and write helper.
func UseStderr() StderrContext {
	app := requireCurrentApp("UseStderr")
	return StderrContext{
		app:    app,
		Stderr: app.Stderr(),
	}
}

// CursorContext exposes cursor-position controls.
type CursorContext struct {
	app *App
}

// SetCursorPosition updates the current app cursor position. Pass nil to hide the cursor.
func (ctx CursorContext) SetCursorPosition(position *CursorPosition) {
	if ctx.app != nil {
		ctx.app.SetCursorPosition(position)
	}
}

// UseCursor exposes cursor-position controls for the currently rendering app.
func UseCursor() CursorContext {
	return CursorContext{app: requireCurrentApp("UseCursor")}
}

// FocusManagerContext exposes focus-management controls.
type FocusManagerContext struct {
	app *App

	// ActiveID stores the currently focused component ID for this render pass,
	// or nil when nothing is focused.
	ActiveID *string
}

// EnableFocus re-enables keyboard focus management.
func (ctx FocusManagerContext) EnableFocus() {
	if ctx.app != nil {
		ctx.app.setFocusEnabled(true)
	}
}

// DisableFocus disables keyboard focus management until re-enabled.
func (ctx FocusManagerContext) DisableFocus() {
	if ctx.app != nil {
		ctx.app.setFocusEnabled(false)
	}
}

// FocusNext switches focus to the next focusable component.
func (ctx FocusManagerContext) FocusNext() bool {
	if ctx.app == nil {
		return false
	}

	manager := ctx.app.focusManager()
	if !manager.FocusNext() {
		return false
	}

	ctx.app.hooksCtx.RequestRerender()
	return true
}

// FocusPrevious switches focus to the previous focusable component.
func (ctx FocusManagerContext) FocusPrevious() bool {
	if ctx.app == nil {
		return false
	}

	manager := ctx.app.focusManager()
	if !manager.FocusPrevious() {
		return false
	}

	ctx.app.hooksCtx.RequestRerender()
	return true
}

// Focus switches focus to the specified id if it is registered.
func (ctx FocusManagerContext) Focus(id string) bool {
	if ctx.app == nil {
		return false
	}

	return ctx.app.focus(id)
}

// UseFocusManager exposes focus-navigation controls for the currently rendering app.
func UseFocusManager() FocusManagerContext {
	app := requireCurrentApp("UseFocusManager")
	return FocusManagerContext{
		app:      app,
		ActiveID: app.activeFocusID(),
	}
}

// UseIsScreenReaderEnabled reports the screen-reader mode flag for the current app.
func UseIsScreenReaderEnabled() bool {
	return requireCurrentApp("UseIsScreenReaderEnabled").IsScreenReaderEnabled()
}
