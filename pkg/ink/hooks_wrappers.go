package ink

import (
	"fmt"
	"reflect"

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
	// IsActive accepts bool or *bool. Nil means the upstream default: active.
	IsActive interface{}
}

// FocusOptions configures UseFocus behavior.
type FocusOptions struct {
	// ID accepts string or *string. Nil means generate an ID; "" is preserved.
	ID        interface{}
	AutoFocus bool
	IsActive  *bool
}

// FocusState mirrors the object shape returned by upstream Ink's useFocus hook
// (`{isFocused, focus}`). It is returned by UseFocusOpts and is the recommended
// shape for new code that wants to match the upstream API exactly.
type FocusState struct {
	// IsFocused snapshots whether this component is currently focused, taken at
	// the time UseFocusOpts is invoked during a render pass.
	IsFocused bool

	// Focus targets a focusable by id. Calling Focus("") with an unknown id is a
	// no-op (matching upstream, where a missing id leaves focus unchanged).
	Focus func(id string)
}

// UseInput registers an input handler against the current app stdin.
func UseInput(callback InputCallback, options ...InputOptions) func() {
	app := requireCurrentApp("UseInput")
	ctx := requireHooksContext("UseInput")
	active := true
	if len(options) > 0 {
		active = normalizeOptionalBool(options[0].IsActive, true, "InputOptions.IsActive")
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
	optionID := ""
	switch len(args) {
	case 0:
	case 1:
		switch typed := args[0].(type) {
		case nil:
		case FocusOptions:
			options = typed
			optionID, customIDProvided = normalizeOptionalString(typed.ID, "FocusOptions.ID")
		case *FocusOptions:
			if typed != nil {
				options = *typed
				optionID, customIDProvided = normalizeOptionalString(typed.ID, "FocusOptions.ID")
			}
		case string:
			optionID = typed
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
		optionID = id
		options.AutoFocus = autoFocus
		customIDProvided = true
	default:
		panic("UseFocus accepts either no args, FocusOptions, id, or id+autoFocus")
	}

	if options.IsActive != nil {
		active = *options.IsActive
	}

	id := UseMemo(func() interface{} {
		if customIDProvided {
			return optionID
		}

		return string(focus.GenerateID("focus"))
	}, []interface{}{optionID, customIDProvided}).(string)

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

// UseFocusOpts is the parity-shaped variant of UseFocus. Its options and return
// shape match upstream Ink's `useFocus({isActive, autoFocus, id}) -> {isFocused, focus}`.
//
// The returned IsFocused is a plain boolean snapshot (not a function), and
// Focus accepts a single id string. The legacy three-tuple UseFocus is kept for
// backwards compatibility; new code should prefer UseFocusOpts.
func UseFocusOpts(options ...FocusOptions) FocusState {
	var opts FocusOptions
	if len(options) > 0 {
		opts = options[0]
	}

	isFocusedFn, focusFn, _ := UseFocus(opts)
	return FocusState{
		IsFocused: isFocusedFn(),
		Focus: func(id string) {
			focusFn(id)
		},
	}
}

func normalizeOptionalBool(value interface{}, defaultValue bool, optionName string) bool {
	if value == nil {
		return defaultValue
	}

	current := reflect.ValueOf(value)
	for current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return defaultValue
		}
		current = current.Elem()
	}

	if current.Kind() != reflect.Bool {
		panic(fmt.Sprintf("%s must be a bool or *bool, got %T", optionName, value))
	}

	return current.Bool()
}

func normalizeOptionalString(value interface{}, optionName string) (string, bool) {
	if value == nil {
		return "", false
	}

	current := reflect.ValueOf(value)
	for current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return "", false
		}
		current = current.Elem()
	}

	if current.Kind() != reflect.String {
		panic(fmt.Sprintf("%s must be a string or *string, got %T", optionName, value))
	}

	return current.String(), true
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

// UseTransition returns whether transition work is pending plus a scheduler
// for lower-priority updates. Urgent updates made before startTransition render
// first; updates inside the callback are flushed on the next scheduler tick.
func UseTransition() (bool, func(func())) {
	return hooks.UseTransition(requireHooksContext("UseTransition"))
}

// UseDeferredValue returns a version of value that may lag one scheduler tick
// behind urgent state. This mirrors React's useDeferredValue for expensive
// derived rendering where responsiveness matters more than immediate parity.
func UseDeferredValue[T any](value T) T {
	return hooks.UseDeferredValue[T](requireHooksContext("UseDeferredValue"), value)
}
