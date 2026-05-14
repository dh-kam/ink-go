package hooks

import (
	"fmt"
	"io"
	"reflect"

	"github.com/dh-kam/goink.go/pkg/focus"
	inkinput "github.com/dh-kam/goink.go/pkg/input"
)

type hookPhase int

const (
	hookPhaseIdle hookPhase = iota
	hookPhaseRender
	hookPhaseEffects
)

// Context holds the state for hooks during a component render
type Context struct {
	states        []interface{}  // Stored state values
	stateIdx      int            // Current state index for state hooks
	inputs        []inputHook    // Input hooks
	inputIdx      int            // Current index for input hooks
	foci          []focusHook    // Focus hooks
	focusIdx      int            // Current index for focus hooks
	effects       []effectHook   // Effect hooks
	memos         []memoHook     // Memo hooks
	callbacks     []callbackHook // Callback hooks
	refs          []refHook      // Ref hooks
	transitions   []transitionHook
	deferred      []deferredHook
	effectIdx     int // Current index for effect hooks
	memoIdx       int // Current index for memo hooks
	callbackIdx   int // Current index for callback hooks
	refIdx        int // Current index for ref hooks
	transitionIdx int
	deferredIdx   int
	focusManager  *focus.FocusManager
	focusEnabled  bool
	phase         hookPhase
	stateChanged  bool
	scheduleWork  func(func())
}

// NewContext creates a new hooks context
func NewContext() *Context {
	return &Context{
		states:       make([]interface{}, 0),
		inputs:       make([]inputHook, 0),
		foci:         make([]focusHook, 0),
		effects:      make([]effectHook, 0),
		memos:        make([]memoHook, 0),
		callbacks:    make([]callbackHook, 0),
		refs:         make([]refHook, 0),
		transitions:  make([]transitionHook, 0),
		deferred:     make([]deferredHook, 0),
		focusManager: focus.NewFocusManager(),
		focusEnabled: true,
		phase:        hookPhaseIdle,
	}
}

// Reset resets the state index for a new render cycle
func (c *Context) Reset() {
	c.stateIdx = 0
	c.inputIdx = 0
	c.focusIdx = 0
	c.effectIdx = 0
	c.memoIdx = 0
	c.callbackIdx = 0
	c.refIdx = 0
	c.transitionIdx = 0
	c.deferredIdx = 0
	c.phase = hookPhaseRender

	for index := range c.inputs {
		c.inputs[index].active = false
	}

	for index := range c.foci {
		c.foci[index].active = false
	}

	for index := range c.effects {
		c.effects[index].active = false
		c.effects[index].pending = false
	}

	for index := range c.transitions {
		c.transitions[index].active = false
	}

	for index := range c.deferred {
		c.deferred[index].active = false
	}
}

// FinalizeRender cleans up hooks that were not used during the current render pass.
func (c *Context) FinalizeRender() {
	activeFocusIDs := make(map[focus.FocusID]struct{}, len(c.foci))
	for _, hook := range c.foci {
		if hook.active {
			activeFocusIDs[hook.state.id] = struct{}{}
		}
	}

	for _, hook := range c.foci {
		if !hook.active && hook.manager != nil {
			if _, stillRepresented := activeFocusIDs[hook.state.id]; stillRepresented {
				continue
			}
			hook.manager.Unregister(hook.state.id)
		}
	}

	for index := range c.effects {
		hook := &c.effects[index]
		if hook.active {
			continue
		}

		if hook.cleanup != nil {
			hook.cleanup()
			hook.cleanup = nil
		}

		hook.hasRun = false
		hook.lastDeps = nil
		hook.effect = nil
		hook.deps = nil
		hook.pending = false
	}

	for index := range c.transitions {
		hook := &c.transitions[index]
		if hook.active {
			continue
		}

		hook.pending = false
		hook.start = nil
	}

	for index := range c.deferred {
		hook := &c.deferred[index]
		if hook.active {
			continue
		}

		hook.latest = nil
		hook.current = nil
		hook.hasValue = false
		hook.pending = false
	}
}

// FocusManager returns the focus manager associated with this hook context.
func (c *Context) FocusManager() *focus.FocusManager {
	return c.focusManager
}

// SetFocusEnabled toggles whether runtime focus-management shortcuts are active.
func (c *Context) SetFocusEnabled(enabled bool) {
	c.focusEnabled = enabled
}

// FocusEnabled reports whether runtime focus management is currently enabled.
func (c *Context) FocusEnabled() bool {
	return c.focusEnabled
}

// ConsumeStateChange reports whether external state changes requested a follow-up render.
func (c *Context) ConsumeStateChange() bool {
	changed := c.stateChanged
	c.stateChanged = false
	return changed
}

// RequestRerender marks that runtime work outside the render phase requires another render pass.
func (c *Context) RequestRerender() {
	if c.phase != hookPhaseRender {
		c.stateChanged = true
	}
}

// SetWorkScheduler configures how deferred hook work is executed. Mounted Ink
// sessions install a scheduler that runs work under the instance lock and then
// flushes any requested render. Bare hook tests leave it unset, which makes
// transition work run synchronously.
func (c *Context) SetWorkScheduler(schedule func(func())) {
	c.scheduleWork = schedule
}

// ScheduleWork schedules runtime work with the mounted renderer, when one is
// installed. Without a mounted renderer the work runs synchronously, matching
// bare hook tests and one-shot render calls.
func (c *Context) ScheduleWork(work func()) {
	c.scheduleDeferredWork(work)
}

func (c *Context) scheduleDeferredWork(work func()) {
	if c.scheduleWork == nil {
		work()
		return
	}

	c.scheduleWork(work)
}

// SetStateFunc is a function to update state
type SetStateFunc func(interface{})

// UseState is a hook for managing component state
func UseState(ctx *Context, initialValue interface{}) (interface{}, SetStateFunc) {
	currentIdx := ctx.stateIdx
	ctx.stateIdx++

	// If this is the first time this hook is called, initialize state
	if currentIdx >= len(ctx.states) {
		ctx.states = append(ctx.states, initialValue)
	}

	// Get current state value
	value := ctx.states[currentIdx]

	// Create setter function
	setValue := func(newValue interface{}) {
		if currentIdx >= len(ctx.states) {
			return
		}

		nextValue := resolveStateUpdate(ctx.states[currentIdx], newValue)
		if reflect.DeepEqual(ctx.states[currentIdx], nextValue) {
			return
		}

		ctx.states[currentIdx] = nextValue
		ctx.RequestRerender()
	}

	return value, setValue
}

func resolveStateUpdate(previous interface{}, update interface{}) interface{} {
	if update == nil {
		return nil
	}

	updateValue := reflect.ValueOf(update)
	if updateValue.Kind() != reflect.Func {
		return update
	}

	updateType := updateValue.Type()
	if updateType.NumIn() != 1 || updateType.NumOut() != 1 {
		return update
	}

	inputType := updateType.In(0)
	var argument reflect.Value
	if previous == nil {
		argument = reflect.Zero(inputType)
	} else {
		previousValue := reflect.ValueOf(previous)
		switch {
		case previousValue.Type().AssignableTo(inputType):
			argument = previousValue
		case previousValue.Type().ConvertibleTo(inputType):
			argument = previousValue.Convert(inputType)
		default:
			return update
		}
	}

	return updateValue.Call([]reflect.Value{argument})[0].Interface()
}

// transitionHook stores the state for one useTransition call.
type transitionHook struct {
	active     bool
	pending    bool
	generation uint64
	start      func(func())
}

// UseTransition mirrors React's useTransition shape. The returned start
// function runs its callback through the context scheduler, so urgent state
// changes made before startTransition can render before lower-priority work.
func UseTransition(ctx *Context) (bool, func(func())) {
	currentIdx := ctx.transitionIdx
	ctx.transitionIdx++

	hook := transitionHook{active: true}
	if currentIdx >= len(ctx.transitions) {
		ctx.transitions = append(ctx.transitions, hook)
	} else {
		hook = ctx.transitions[currentIdx]
		hook.active = true
	}

	start := func(work func()) {
		if work == nil {
			return
		}

		hook := &ctx.transitions[currentIdx]
		hook.generation++
		generation := hook.generation
		hook.pending = true
		ctx.RequestRerender()

		ctx.scheduleDeferredWork(func() {
			if currentIdx >= len(ctx.transitions) {
				return
			}

			hook := &ctx.transitions[currentIdx]
			if hook.generation != generation {
				return
			}

			work()

			if hook.generation == generation {
				hook.pending = false
				ctx.RequestRerender()
			}
		})
	}

	hook.start = start
	ctx.transitions[currentIdx] = hook

	return hook.pending, start
}

// deferredHook stores the state for one useDeferredValue call.
type deferredHook struct {
	active     bool
	current    interface{}
	latest     interface{}
	hasValue   bool
	pending    bool
	generation uint64
}

// UseDeferredValue returns a value that is allowed to lag one scheduler tick
// behind the latest input value. If a newer value arrives before the pending
// work runs, the older deferred update is ignored.
func UseDeferredValue[T any](ctx *Context, value T) T {
	currentIdx := ctx.deferredIdx
	ctx.deferredIdx++

	hook := deferredHook{
		active:   true,
		current:  value,
		latest:   value,
		hasValue: true,
	}
	if currentIdx >= len(ctx.deferred) {
		ctx.deferred = append(ctx.deferred, hook)
		return value
	}

	hook = ctx.deferred[currentIdx]
	hook.active = true

	if !hook.hasValue {
		hook.current = value
		hook.latest = value
		hook.hasValue = true
		ctx.deferred[currentIdx] = hook
		return value
	}

	if reflect.DeepEqual(hook.latest, value) {
		ctx.deferred[currentIdx] = hook
		current, ok := hook.current.(T)
		if !ok {
			return value
		}
		return current
	}

	hook.latest = value
	hook.pending = true
	hook.generation++
	generation := hook.generation
	ctx.deferred[currentIdx] = hook

	ctx.scheduleDeferredWork(func() {
		if currentIdx >= len(ctx.deferred) {
			return
		}

		hook := &ctx.deferred[currentIdx]
		if hook.generation != generation {
			return
		}

		if reflect.DeepEqual(hook.current, hook.latest) {
			hook.pending = false
			return
		}

		hook.current = hook.latest
		hook.pending = false
		ctx.RequestRerender()
	})

	current, ok := hook.current.(T)
	if !ok {
		return value
	}

	return current
}

// InputCallback is any supported useInput handler signature.
type InputCallback interface{}

type normalizedInputCallback func(string, inkinput.HookKey, []string) bool

// inputHook represents a useInput hook
type inputHook struct {
	callback normalizedInputCallback
	active   bool
}

func normalizeInputCallback(callback InputCallback) normalizedInputCallback {
	switch typed := callback.(type) {
	case func(string, inkinput.HookKey):
		return func(input string, key inkinput.HookKey, keys []string) bool {
			typed(input, key)
			return false
		}
	case func(string, inkinput.HookKey) bool:
		return func(input string, key inkinput.HookKey, keys []string) bool {
			return typed(input, key)
		}
	case func(string, *inkinput.HookKey):
		return func(input string, key inkinput.HookKey, keys []string) bool {
			typed(input, &key)
			return false
		}
	case func(string, *inkinput.HookKey) bool:
		return func(input string, key inkinput.HookKey, keys []string) bool {
			return typed(input, &key)
		}
	case func(interface{}, []string):
		return func(input string, key inkinput.HookKey, keys []string) bool {
			typed(input, keys)
			return false
		}
	case func(interface{}, []string) bool:
		return func(input string, key inkinput.HookKey, keys []string) bool {
			return typed(input, keys)
		}
	case func(string, []string):
		return func(input string, key inkinput.HookKey, keys []string) bool {
			typed(input, keys)
			return false
		}
	case func(string, []string) bool:
		return func(input string, key inkinput.HookKey, keys []string) bool {
			return typed(input, keys)
		}
	default:
		panic(fmt.Sprintf("UseInput does not support callback type %T", callback))
	}
}

// UseInput registers an input handler callback
// The callback receives the input and a list of keys, returns true to exit
func UseInput(ctx *Context, callback InputCallback, input io.Reader, active bool) func() {
	currentIdx := ctx.inputIdx
	ctx.inputIdx++

	hook := inputHook{
		callback: normalizeInputCallback(callback),
		active:   active,
	}

	// If this is the first time, initialize
	if currentIdx >= len(ctx.inputs) {
		ctx.inputs = append(ctx.inputs, hook)
	} else {
		ctx.inputs[currentIdx] = hook
	}

	// Return cleanup function
	return func() {
		if currentIdx < len(ctx.inputs) {
			ctx.inputs[currentIdx].active = false
		}
	}
}

// GetInputHooks returns all active input hooks
func (c *Context) GetInputHooks() []inputHook {
	active := make([]inputHook, 0)
	for _, hook := range c.inputs {
		if hook.active {
			active = append(active, hook)
		}
	}
	return active
}

// DispatchInput invokes all active input hooks and returns true if any requested exit.
func (c *Context) DispatchInput(input string, key inkinput.HookKey, keys []string) bool {
	shouldExit := false

	for _, hook := range c.inputs {
		if !hook.active || hook.callback == nil {
			continue
		}

		if hook.callback(input, key, keys) {
			shouldExit = true
		}
	}

	return shouldExit
}

// FocusHookState holds the state for a useFocus hook
type FocusHookState struct {
	id        focus.FocusID
	isFocused bool
	autoFocus bool
	isActive  bool
}

// focusHook represents a useFocus hook
type focusHook struct {
	state   FocusHookState
	active  bool
	manager *focus.FocusManager
}

// UseFocus registers a focusable component and returns its focus state
// The id parameter uniquely identifies this focusable component
// autoFocus determines if this component should be focused initially
func UseFocus(ctx *Context, id string, autoFocus bool, active bool) (isFocused func() bool, focusFunc func(...string), blurFunc func()) {
	currentIdx := ctx.focusIdx
	ctx.focusIdx++

	manager := ctx.FocusManager()

	hook := focusHook{
		state: FocusHookState{
			id:        focus.FocusID(id),
			isFocused: false,
			autoFocus: autoFocus,
			isActive:  active,
		},
		active:  true,
		manager: manager,
	}

	// If this is the first time, initialize
	if currentIdx >= len(ctx.foci) {
		ctx.foci = append(ctx.foci, hook)
	} else {
		hook = ctx.foci[currentIdx]
		if hook.manager != nil && (hook.state.id != focus.FocusID(id) || hook.state.autoFocus != autoFocus) {
			// Ink re-runs the registration effect when either the generated id or
			// autoFocus changes, so mirror that by removing the previous focusable
			// before re-registering the current one.
			alreadyRepresented := false
			for index := 0; index < currentIdx; index++ {
				previous := ctx.foci[index]
				if previous.active && previous.state.id == hook.state.id {
					alreadyRepresented = true
					break
				}
			}
			if !alreadyRepresented {
				hook.manager.Unregister(hook.state.id)
			}
		}
		hook.state.id = focus.FocusID(id)
		hook.state.autoFocus = autoFocus
		hook.state.isActive = active
		hook.active = true
		hook.manager = manager
	}

	// Register with focus manager
	manager.Register(hook.state.id, autoFocus)
	if active {
		manager.Activate(hook.state.id)
	} else {
		manager.Deactivate(hook.state.id)
	}
	hook.state.isFocused = manager.IsFocused(hook.state.id)

	// Update the hook in context
	ctx.foci[currentIdx] = hook

	// Return functions to manage focus
	isFocusedFunc := func() bool {
		return manager.IsFocused(hook.state.id)
	}

	focusFunc = func(targets ...string) {
		targetID := hook.state.id
		if len(targets) > 0 {
			targetID = focus.FocusID(targets[0])
		}

		before := manager.FocusedID()
		beforeHasFocus := manager.HasFocus()
		manager.Focus(targetID)
		if manager.FocusedID() != before || manager.HasFocus() != beforeHasFocus {
			ctx.RequestRerender()
		}
	}

	blurFunc = func() {
		if manager.IsFocused(hook.state.id) {
			before := manager.FocusedID()
			beforeHasFocus := manager.HasFocus()
			manager.Blur()
			if manager.FocusedID() != before || manager.HasFocus() != beforeHasFocus {
				ctx.RequestRerender()
			}
		}
	}

	return isFocusedFunc, focusFunc, blurFunc
}

// GetFocusHooks returns all active focus hooks
func (c *Context) GetFocusHooks() []focusHook {
	active := make([]focusHook, 0)
	for _, hook := range c.foci {
		if hook.active {
			active = append(active, hook)
		}
	}
	return active
}

// EffectHook types and functions

// effectHook represents a useEffect hook
type effectHook struct {
	effect   func() func()
	deps     []interface{}
	hasRun   bool
	cleanup  func()
	lastDeps []interface{}
	active   bool
	pending  bool
}

// UseEffect runs an effect function after render
// The effect can return a cleanup function that runs before unmount or before next effect
// deps controls when the effect runs: empty [] means run once, nil means run every render
func UseEffect(ctx *Context, effect func() func(), deps []interface{}) {
	currentIdx := ctx.effectIdx
	ctx.effectIdx++

	hook := effectHook{
		effect: effect,
		deps:   deps,
		active: true,
	}

	if currentIdx >= len(ctx.effects) {
		ctx.effects = append(ctx.effects, hook)
	} else {
		hook = ctx.effects[currentIdx]
		hook.effect = effect
		hook.deps = deps
		hook.active = true
	}

	// Check if dependencies changed
	// If deps is nil, run on every render
	shouldRun := !hook.hasRun || deps == nil
	if !shouldRun && deps != nil {
		if len(hook.lastDeps) != len(deps) {
			shouldRun = true
		} else {
			for i := range deps {
				if hook.lastDeps[i] != deps[i] {
					shouldRun = true
					break
				}
			}
		}
	}

	if shouldRun {
		hook.pending = true
	}

	ctx.effects[currentIdx] = hook
}

// RunEffects executes any effects scheduled during the latest render pass.
func (c *Context) RunEffects() {
	c.phase = hookPhaseEffects
	defer func() {
		c.phase = hookPhaseIdle
	}()

	for index := range c.effects {
		hook := &c.effects[index]
		if !hook.active || !hook.pending {
			continue
		}

		if hook.cleanup != nil {
			hook.cleanup()
			hook.cleanup = nil
		}

		if hook.effect != nil {
			hook.cleanup = hook.effect()
		}

		hook.hasRun = true
		hook.pending = false
		if hook.deps != nil {
			hook.lastDeps = make([]interface{}, len(hook.deps))
			copy(hook.lastDeps, hook.deps)
		} else {
			hook.lastDeps = nil
		}
	}
}

// Memo and Callback types

// memoHook represents a useMemo hook
type memoHook struct {
	value    interface{}
	deps     []interface{}
	hasValue bool
}

// UseMemo memoizes a computed value
// The value is only recomputed when dependencies change
func UseMemo(ctx *Context, compute func() interface{}, deps []interface{}) interface{} {
	currentIdx := ctx.memoIdx
	ctx.memoIdx++

	hook := memoHook{}

	if currentIdx >= len(ctx.memos) {
		ctx.memos = append(ctx.memos, hook)
	} else {
		hook = ctx.memos[currentIdx]
	}

	// Check if dependencies changed
	shouldCompute := !hook.hasValue
	if !shouldCompute && deps != nil {
		if len(hook.deps) != len(deps) {
			shouldCompute = true
		} else {
			for i := range deps {
				if hook.deps[i] != deps[i] {
					shouldCompute = true
					break
				}
			}
		}
	}

	if shouldCompute {
		hook.value = compute()
		hook.hasValue = true
		if deps != nil {
			hook.deps = make([]interface{}, len(deps))
			copy(hook.deps, deps)
		}
	}

	ctx.memos[currentIdx] = hook
	return hook.value
}

// callbackHook represents a useCallback hook
type callbackHook struct {
	callback func()
	deps     []interface{}
}

// UseCallback memoizes a callback function
// The callback is only recreated when dependencies change
func UseCallback(ctx *Context, callback func(), deps []interface{}) func() {
	currentIdx := ctx.callbackIdx
	ctx.callbackIdx++

	hook := callbackHook{}

	if currentIdx >= len(ctx.callbacks) {
		ctx.callbacks = append(ctx.callbacks, hook)
	} else {
		hook = ctx.callbacks[currentIdx]
	}

	// Check if dependencies changed
	shouldUpdate := hook.callback == nil
	if !shouldUpdate && deps != nil {
		if len(hook.deps) != len(deps) {
			shouldUpdate = true
		} else {
			for i := range deps {
				if hook.deps[i] != deps[i] {
					shouldUpdate = true
					break
				}
			}
		}
	}

	if shouldUpdate {
		hook.callback = callback
		if deps != nil {
			hook.deps = make([]interface{}, len(deps))
			copy(hook.deps, deps)
		}
	}

	ctx.callbacks[currentIdx] = hook
	return hook.callback
}

// Ref types

// refHook represents a useRef hook
type refHook struct {
	ref *Ref
}

// Ref holds a mutable value that persists across renders
type Ref struct {
	value interface{}
}

// Current returns the current value of the ref
func (r *Ref) Current() interface{} {
	return r.value
}

// SetCurrent sets the value of the ref
func (r *Ref) SetCurrent(value interface{}) {
	r.value = value
}

// UseRef returns a ref object that persists across renders
func UseRef(ctx *Context, initialValue interface{}) *Ref {
	currentIdx := ctx.refIdx
	ctx.refIdx++

	hook := refHook{}

	if currentIdx >= len(ctx.refs) {
		hook = refHook{ref: &Ref{value: initialValue}}
		ctx.refs = append(ctx.refs, hook)
	} else {
		hook = ctx.refs[currentIdx]
	}

	ctx.refs[currentIdx] = hook
	return hook.ref
}

// RunCleanup runs all pending cleanup functions for effects
func (c *Context) RunCleanup() {
	for _, hook := range c.effects {
		if hook.cleanup != nil {
			hook.cleanup()
			hook.cleanup = nil
		}
	}
}
