package ink

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/focus"
	"github.com/dh-kam/goink.go/pkg/hooks"
	inkinput "github.com/dh-kam/goink.go/pkg/input"
)

// Ref is the Ink-level alias for hooks.Ref.
type Ref = hooks.Ref

// SetStateFunc is the Ink-level alias for hooks.SetStateFunc.
type SetStateFunc = hooks.SetStateFunc

// InputKey is the Ink-level alias for the boolean key object used by useInput.
type InputKey = inkinput.HookKey

// InputCallback is the Ink-level alias for hooks.InputCallback.
type InputCallback = hooks.InputCallback

// InputOptions configures UseInput behavior.
type InputOptions struct {
	IsActive bool
}

// FocusOptions configures UseFocus behavior.
type FocusOptions struct {
	ID        string
	AutoFocus bool
	IsActive  *bool
}

// UseInput registers an input handler against the current app stdin.
func UseInput(callback InputCallback, options ...InputOptions) func() {
	app := requireCurrentApp("UseInput")
	ctx := requireHooksContext("UseInput")
	active := true
	if len(options) > 0 {
		active = options[0].IsActive
	}

	UseEffect(func() func() {
		if !active {
			return nil
		}

		_ = app.SetRawMode(true)
		return func() {
			_ = app.SetRawMode(false)
		}
	}, []interface{}{"use-input-raw-mode", active})

	return hooks.UseInput(ctx, callback, app.Stdin(), active)
}

// UseFocus registers a focusable component and returns its focus controls.
func UseFocus(args ...interface{}) (func() bool, func(...string), func()) {
	app := requireCurrentApp("UseFocus")
	ctx := requireHooksContext("UseFocus")

	options := FocusOptions{}
	active := true
	customIDProvided := false
	switch len(args) {
	case 0:
	case 1:
		switch typed := args[0].(type) {
		case nil:
		case FocusOptions:
			options = typed
			customIDProvided = typed.ID != ""
		case *FocusOptions:
			if typed != nil {
				options = *typed
				customIDProvided = typed.ID != ""
			}
		case string:
			options.ID = typed
			customIDProvided = true
		default:
			panic(fmt.Sprintf("UseFocus does not support argument type %T", args[0]))
		}
	case 2:
		id, ok := args[0].(string)
		if !ok {
			panic(fmt.Sprintf("UseFocus expected id string, got %T", args[0]))
		}
		autoFocus, ok := args[1].(bool)
		if !ok {
			panic(fmt.Sprintf("UseFocus expected autoFocus bool, got %T", args[1]))
		}
		options.ID = id
		options.AutoFocus = autoFocus
		customIDProvided = true
	default:
		panic("UseFocus accepts either no args, FocusOptions, id, or id+autoFocus")
	}

	if options.IsActive != nil {
		active = *options.IsActive
	}

	id := UseMemo(func() interface{} {
		if customIDProvided || options.ID != "" {
			return options.ID
		}

		return string(focus.GenerateID("focus"))
	}, []interface{}{options.ID, customIDProvided}).(string)

	UseEffect(func() func() {
		if !active {
			return nil
		}

		_ = app.SetRawMode(true)
		return func() {
			_ = app.SetRawMode(false)
		}
	}, []interface{}{"use-focus-raw-mode", active})

	isFocused, focusFn, blurFn := hooks.UseFocus(ctx, id, options.AutoFocus, active)
	if customIDProvided && id == "" {
		return func() bool {
			return false
		}, focusFn, blurFn
	}

	return isFocused, focusFn, blurFn
}

// UseEffect runs an effect for the currently rendering component.
func UseEffect(effect func() func(), deps []interface{}) {
	hooks.UseEffect(requireHooksContext("UseEffect"), effect, deps)
}

// UseMemo memoizes a computed value for the current component.
func UseMemo(compute func() interface{}, deps []interface{}) interface{} {
	return hooks.UseMemo(requireHooksContext("UseMemo"), compute, deps)
}

// UseCallback memoizes a callback function for the current component.
func UseCallback(callback func(), deps []interface{}) func() {
	return hooks.UseCallback(requireHooksContext("UseCallback"), callback, deps)
}

// UseRef returns a mutable value that persists across renders.
func UseRef(initialValue interface{}) *Ref {
	return hooks.UseRef(requireHooksContext("UseRef"), initialValue)
}
