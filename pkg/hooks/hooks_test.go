package hooks_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/focus"
	"github.com/dh-kam/goink.go/pkg/hooks"
	inkinput "github.com/dh-kam/goink.go/pkg/input"
)

// TestUseState tests basic state management
func TestUseState(t *testing.T) {
	// Initialize hooks context
	ctx := hooks.NewContext()

	// First render
	value, setValue := hooks.UseState(ctx, 0)

	if value != 0 {
		t.Errorf("Expected initial value 0, got %d", value)
	}

	// Update state
	setValue(42)

	// Second render (hooks should return updated value)
	ctx.Reset() // Reset for next render cycle
	value2, _ := hooks.UseState(ctx, 0)

	if value2 != 42 {
		t.Errorf("Expected updated value 42, got %d", value2)
	}
}

func TestUseStateFunctionalUpdate(t *testing.T) {
	ctx := hooks.NewContext()

	_, setCount := hooks.UseState(ctx, 0)
	setCount(func(previous int) int {
		return previous + 1
	})
	setCount(func(previous int) int {
		return previous + 1
	})

	ctx.Reset()
	value, _ := hooks.UseState(ctx, 0)
	if value != 2 {
		t.Fatalf("expected functional updates to compose to 2, got %v", value)
	}
}

// TestUseStateMultiple tests multiple state hooks
func TestUseStateMultiple(t *testing.T) {
	ctx := hooks.NewContext()

	// First render - two state hooks
	count, setCount := hooks.UseState(ctx, 0)
	name, setName := hooks.UseState(ctx, "default")

	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
	if name != "default" {
		t.Errorf("Expected name 'default', got %q", name)
	}

	// Update states
	setCount(5)
	setName("updated")

	// Second render
	ctx.Reset()
	count2, _ := hooks.UseState(ctx, 0)
	name2, _ := hooks.UseState(ctx, "default")

	if count2 != 5 {
		t.Errorf("Expected count 5, got %d", count2)
	}
	if name2 != "updated" {
		t.Errorf("Expected name 'updated', got %q", name2)
	}
}

// TestUseStateOrder tests that hook call order matters
func TestUseStateOrder(t *testing.T) {
	ctx := hooks.NewContext()

	// First render
	_, set1 := hooks.UseState(ctx, "first")
	_, set2 := hooks.UseState(ctx, "second")

	set1("updated-first")
	set2("updated-second")

	// Second render - MUST call hooks in same order
	ctx.Reset()
	val1_r2, _ := hooks.UseState(ctx, "first")
	val2_r2, _ := hooks.UseState(ctx, "second")

	if val1_r2 != "updated-first" {
		t.Errorf("Expected 'updated-first', got %q", val1_r2)
	}
	if val2_r2 != "updated-second" {
		t.Errorf("Expected 'updated-second', got %q", val2_r2)
	}
}

// TestUseInput tests input hook registration
func TestUseInput(t *testing.T) {
	ctx := hooks.NewContext()

	called := false
	callback := func(input interface{}, keys []string) bool {
		called = true
		return false
	}

	cleanup := hooks.UseInput(ctx, callback, strings.NewReader(""), true)

	if called {
		t.Error("Callback should not be called immediately")
	}

	// Cleanup should work
	cleanup()
}

// TestUseInputMultiple tests multiple input hooks
func TestUseInputMultiple(t *testing.T) {
	ctx := hooks.NewContext()

	callCount := 0
	callback1 := func(input interface{}, keys []string) bool {
		callCount++
		return false
	}
	callback2 := func(input interface{}, keys []string) bool {
		callCount++
		return false
	}

	hooks.UseInput(ctx, callback1, nil, true)
	hooks.UseInput(ctx, callback2, nil, true)

	inputHooks := ctx.GetInputHooks()
	if len(inputHooks) != 2 {
		t.Errorf("Expected 2 input hooks, got %d", len(inputHooks))
	}
}

// TestUseInputCleanup tests cleanup of input hooks
func TestUseInputCleanup(t *testing.T) {
	ctx := hooks.NewContext()

	callback := func(input interface{}, keys []string) bool {
		return false
	}

	cleanup := hooks.UseInput(ctx, callback, nil, true)

	// Should have 1 active hook
	inputHooks := ctx.GetInputHooks()
	if len(inputHooks) != 1 {
		t.Errorf("Expected 1 input hook before cleanup, got %d", len(inputHooks))
	}

	// Cleanup
	cleanup()

	// Should have 0 active hooks
	inputHooks = ctx.GetInputHooks()
	if len(inputHooks) != 0 {
		t.Errorf("Expected 0 input hooks after cleanup, got %d", len(inputHooks))
	}
}

func TestUseInputInactiveHookDoesNotDispatch(t *testing.T) {
	ctx := hooks.NewContext()
	called := false

	hooks.UseInput(ctx, func(input string, key inkinput.HookKey) {
		called = true
	}, nil, false)

	if len(ctx.GetInputHooks()) != 0 {
		t.Fatalf("expected inactive input hooks to be filtered out, got %d", len(ctx.GetInputHooks()))
	}

	if shouldExit := ctx.DispatchInput("a", inkinput.HookKey{}, nil); shouldExit {
		t.Fatal("expected inactive hook not to request exit")
	}
	if called {
		t.Fatal("expected inactive hook not to receive input")
	}
}

func TestUseInputSupportsKeyObjectSignature(t *testing.T) {
	ctx := hooks.NewContext()

	var receivedInput string
	var receivedKey inkinput.HookKey

	hooks.UseInput(ctx, func(input string, key inkinput.HookKey) {
		receivedInput = input
		receivedKey = key
	}, nil, true)

	if shouldExit := ctx.DispatchInput("", inkinput.HookKey{LeftArrow: true}, []string{inkinput.KeyLeft}); shouldExit {
		t.Fatal("expected key-object handler not to request exit")
	}

	if receivedInput != "" {
		t.Fatalf("expected empty input text for arrow key, got %q", receivedInput)
	}
	if !receivedKey.LeftArrow {
		t.Fatal("expected leftArrow flag to be delivered")
	}
}

func TestUseInputSupportsLegacySignature(t *testing.T) {
	ctx := hooks.NewContext()

	var receivedInput interface{}
	var receivedKeys []string

	hooks.UseInput(ctx, func(input interface{}, keys []string) bool {
		receivedInput = input
		receivedKeys = append([]string(nil), keys...)
		return input == "q"
	}, nil, true)

	if shouldExit := ctx.DispatchInput("q", inkinput.HookKey{}, []string{"q"}); !shouldExit {
		t.Fatal("expected legacy handler to request exit")
	}

	if receivedInput != "q" {
		t.Fatalf("expected legacy handler input %q, got %#v", "q", receivedInput)
	}
	if strings.Join(receivedKeys, ",") != "q" {
		t.Fatalf("expected legacy handler keys to be preserved, got %#v", receivedKeys)
	}
}

// TestUseFocus tests basic focus hook functionality
func TestUseFocus(t *testing.T) {
	ctx := hooks.NewContext()

	id := "test-focus-basic"
	isFocused, focusFn, blurFn := hooks.UseFocus(ctx, id, true, true)

	// First component with autoFocus should be focused
	if !isFocused() {
		t.Error("Expected component to be focused with autoFocus=true")
	}

	// Blur should work
	blurFn()
	if isFocused() {
		t.Error("Expected component to not be focused after blur")
	}

	// Focus should work
	focusFn()
	if !isFocused() {
		t.Error("Expected component to be focused after focus")
	}

	// Cleanup
	focusMgr := ctx.FocusManager()
	focusMgr.Unregister(focus.FocusID(id))
}

// TestUseFocusMultiple tests multiple focus hooks
func TestUseFocusMultiple(t *testing.T) {
	ctx := hooks.NewContext()

	id1 := "multiple-comp-1"
	id2 := "multiple-comp-2"

	isFocused1, _, _ := hooks.UseFocus(ctx, id1, false, true)
	isFocused2, _, _ := hooks.UseFocus(ctx, id2, false, true)

	// Neither should be focused initially (autoFocus=false)
	if isFocused1() {
		t.Error("Component 1 should not be focused initially")
	}
	if isFocused2() {
		t.Error("Component 2 should not be focused initially")
	}

	// Cleanup
	focusMgr := ctx.FocusManager()
	focusMgr.Unregister(focus.FocusID(id1))
	focusMgr.Unregister(focus.FocusID(id2))
}

// TestUseFocusAutoFocus tests autoFocus behavior
func TestUseFocusAutoFocus(t *testing.T) {
	ctx := hooks.NewContext()

	// Use unique IDs to avoid conflicts with other tests
	id1 := "autofocus-comp-1"
	id2 := "autofocus-comp-2"

	isFocused1, _, _ := hooks.UseFocus(ctx, id1, true, true)
	isFocused2, _, _ := hooks.UseFocus(ctx, id2, false, true)

	// First component with autoFocus should be focused
	if !isFocused1() {
		t.Error("First component with autoFocus should be focused")
	}

	if isFocused2() {
		t.Error("Second component should not be focused")
	}

	// Cleanup
	ctx.FocusManager().Unregister(focus.FocusID(id1))
	ctx.FocusManager().Unregister(focus.FocusID(id2))
}

func TestUseFocusAutoFocusChangeReRegistersComponent(t *testing.T) {
	ctx := hooks.NewContext()

	isFocused, _, _ := hooks.UseFocus(ctx, "component-1", false, true)
	ctx.FinalizeRender()

	if isFocused() {
		t.Fatal("expected component to remain unfocused when autoFocus starts disabled")
	}

	ctx.Reset()
	isFocused, _, _ = hooks.UseFocus(ctx, "component-1", true, true)
	ctx.FinalizeRender()

	if !isFocused() {
		t.Fatal("expected enabling autoFocus on rerender to focus the component")
	}
	if ctx.FocusManager().FocusableCount() != 1 {
		t.Fatalf("expected exactly one focusable after autoFocus rerender, got %d", ctx.FocusManager().FocusableCount())
	}
}

func TestUseFocusAutoFocusSkipsInactiveComponentsOnInitialRender(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, _, _ := hooks.UseFocus(ctx, "component-1", true, false)
	secondFocused, _, _ := hooks.UseFocus(ctx, "component-2", true, false)
	thirdFocused, _, _ := hooks.UseFocus(ctx, "component-3", true, true)

	if firstFocused() {
		t.Fatal("expected inactive autoFocus component not to remain focused")
	}
	if secondFocused() {
		t.Fatal("expected subsequent inactive autoFocus component not to remain focused")
	}
	if !thirdFocused() {
		t.Fatal("expected first active autoFocus component to receive focus")
	}
}

// TestUseFocusHooks tests getting focus hooks from context
func TestUseFocusHooks(t *testing.T) {
	ctx := hooks.NewContext()

	hooks.UseFocus(ctx, "component-1", false, true)
	hooks.UseFocus(ctx, "component-2", false, true)

	focusHooks := ctx.GetFocusHooks()
	if len(focusHooks) != 2 {
		t.Errorf("Expected 2 focus hooks, got %d", len(focusHooks))
	}
}

// TestUseFocusNavigation tests focus navigation between components
func TestUseFocusNavigation(t *testing.T) {
	ctx := hooks.NewContext()

	_, focus1, _ := hooks.UseFocus(ctx, "component-1", true, true)
	_, _, _ = hooks.UseFocus(ctx, "component-2", false, true)

	// component-1 should be focused initially
	focus1() // Ensure component-1 is focused

	fm := ctx.FocusManager()

	// Move to next component
	fm.FocusNext()

	// The second component should now be focused
	if !fm.IsFocused("component-2") {
		t.Error("Expected component-2 to be focused after FocusNext")
	}
}

func TestUseFocusInactiveHooksAreSkippedByNavigation(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, _, _ := hooks.UseFocus(ctx, "component-1", true, true)
	secondFocused, _, _ := hooks.UseFocus(ctx, "component-2", false, false)
	thirdFocused, _, _ := hooks.UseFocus(ctx, "component-3", false, true)

	if !firstFocused() {
		t.Fatal("expected first active component to receive autoFocus")
	}
	if secondFocused() {
		t.Fatal("expected inactive component not to be focused")
	}

	ctx.FocusManager().FocusNext()
	if !thirdFocused() {
		t.Fatal("expected FocusNext to skip inactive components")
	}
}

func TestUseFocusPreviousWithoutCurrentFocusStartsAtLastActive(t *testing.T) {
	ctx := hooks.NewContext()

	_, _, _ = hooks.UseFocus(ctx, "component-1", false, true)
	_, _, _ = hooks.UseFocus(ctx, "component-2", false, false)
	thirdFocused, _, _ := hooks.UseFocus(ctx, "component-3", false, true)

	if !ctx.FocusManager().FocusPrevious() {
		t.Fatal("expected FocusPrevious to succeed with active focusables")
	}
	if !thirdFocused() {
		t.Fatal("expected FocusPrevious without focus to select the last active component")
	}
}

func TestUseFocusCanTargetExplicitID(t *testing.T) {
	ctx := hooks.NewContext()

	_, focusFirst, _ := hooks.UseFocus(ctx, "component-1", true, true)
	secondFocused, _, _ := hooks.UseFocus(ctx, "component-2", false, true)

	focusFirst("component-2")
	if !secondFocused() {
		t.Fatal("expected explicit focus target to move focus to the provided id")
	}
}

func TestUseFocusCanTargetExplicitEmptyStringID(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, focusFirst, _ := hooks.UseFocus(ctx, "component-1", true, true)
	_, _, _ = hooks.UseFocus(ctx, "", false, true)

	focusFirst("")

	if firstFocused() {
		t.Fatal("expected explicit empty-string target to move focus away from the visible component")
	}
	if !ctx.FocusManager().HasFocus() {
		t.Fatal("expected explicit empty-string target to still count as focused internally")
	}
	if !ctx.FocusManager().IsFocused(focus.FocusID("")) {
		t.Fatal("expected explicit empty-string target to become focused internally")
	}
}

func TestUseFocusFocusFuncRequestsRerenderOutsideRender(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, focusFirst, _ := hooks.UseFocus(ctx, "component-1", true, true)
	secondFocused, _, _ := hooks.UseFocus(ctx, "component-2", false, true)
	hooks.UseEffect(ctx, func() func() {
		focusFirst("component-2")
		return nil
	}, []interface{}{"focus-second"})

	ctx.FinalizeRender()
	ctx.RunEffects()

	if !ctx.ConsumeStateChange() {
		t.Fatal("expected focus() outside render to request a rerender")
	}
	if firstFocused() {
		t.Fatal("expected focus() outside render to move focus away from the first component")
	}
	if !secondFocused() {
		t.Fatal("expected focus() outside render to focus the target component")
	}
}

func TestUseFocusBlurFuncRequestsRerenderOutsideRender(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, _, blurFirst := hooks.UseFocus(ctx, "component-1", true, true)
	hooks.UseEffect(ctx, func() func() {
		blurFirst()
		return nil
	}, []interface{}{"blur-first"})

	ctx.FinalizeRender()
	ctx.RunEffects()

	if !ctx.ConsumeStateChange() {
		t.Fatal("expected blur() outside render to request a rerender")
	}
	if firstFocused() {
		t.Fatal("expected blur() outside render to clear focus")
	}
	if ctx.FocusManager().FocusedID() != "" {
		t.Fatal("expected blur() outside render to clear the focused id")
	}
}

func TestUseFocusExplicitEmptyStringFocusRequestsRerenderOutsideRender(t *testing.T) {
	ctx := hooks.NewContext()

	_, focusEmpty, _ := hooks.UseFocus(ctx, "", false, true)
	hooks.UseEffect(ctx, func() func() {
		focusEmpty()
		return nil
	}, []interface{}{"focus-empty"})

	ctx.FinalizeRender()
	ctx.RunEffects()

	if !ctx.ConsumeStateChange() {
		t.Fatal("expected focus() on an explicit empty-string id outside render to request a rerender")
	}
	if !ctx.FocusManager().HasFocus() {
		t.Fatal("expected focus() on an explicit empty-string id to set focus")
	}
	if !ctx.FocusManager().IsFocused(focus.FocusID("")) {
		t.Fatal("expected explicit empty-string id to become focused internally")
	}
}

func TestUseFocusExplicitEmptyStringTargetRequestsRerenderOutsideRender(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, focusFirst, _ := hooks.UseFocus(ctx, "component-1", true, true)
	_, _, _ = hooks.UseFocus(ctx, "", false, true)
	hooks.UseEffect(ctx, func() func() {
		focusFirst("")
		return nil
	}, []interface{}{"focus-empty-target"})

	ctx.FinalizeRender()
	ctx.RunEffects()

	if !ctx.ConsumeStateChange() {
		t.Fatal("expected focusing an explicit empty-string target outside render to request a rerender")
	}
	if firstFocused() {
		t.Fatal("expected explicit empty-string target to move focus away from the visible component")
	}
	if !ctx.FocusManager().IsFocused(focus.FocusID("")) {
		t.Fatal("expected explicit empty-string target to become focused internally")
	}
}

func TestUseFocusExplicitEmptyStringBlurRequestsRerenderOutsideRender(t *testing.T) {
	ctx := hooks.NewContext()

	_, _, blurEmpty := hooks.UseFocus(ctx, "", true, true)
	hooks.UseEffect(ctx, func() func() {
		blurEmpty()
		return nil
	}, []interface{}{"blur-empty"})

	ctx.FinalizeRender()
	ctx.RunEffects()

	if !ctx.ConsumeStateChange() {
		t.Fatal("expected blur() on an explicit empty-string id outside render to request a rerender")
	}
	if ctx.FocusManager().HasFocus() {
		t.Fatal("expected blur() on an explicit empty-string id to clear focus")
	}
}

func TestUseFocusMissingTargetKeepsCurrentFocus(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, _, _ := hooks.UseFocus(ctx, "component-1", false, true)
	secondFocused, focusSecond, _ := hooks.UseFocus(ctx, "component-2", true, true)
	hooks.UseEffect(ctx, func() func() {
		focusSecond("missing-focus-id")
		return nil
	}, []interface{}{"missing-target"})

	ctx.FinalizeRender()
	ctx.RunEffects()

	if ctx.ConsumeStateChange() {
		t.Fatal("expected missing focus target to avoid requesting a rerender")
	}
	if firstFocused() {
		t.Fatal("expected missing focus target not to move focus to a different component")
	}
	if !secondFocused() {
		t.Fatal("expected missing focus target to leave the current component focused")
	}
}

func TestUseFocusMissingTargetSkipsInactiveFallbacks(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, _, _ := hooks.UseFocus(ctx, "component-1", false, false)
	secondFocused, _, _ := hooks.UseFocus(ctx, "component-2", false, true)
	thirdFocused, focusThird, _ := hooks.UseFocus(ctx, "component-3", true, true)
	hooks.UseEffect(ctx, func() func() {
		focusThird("missing-focus-id")
		return nil
	}, []interface{}{"missing-target-active"})

	ctx.FinalizeRender()
	ctx.RunEffects()

	if ctx.ConsumeStateChange() {
		t.Fatal("expected missing focus target to avoid requesting a rerender")
	}
	if firstFocused() {
		t.Fatal("expected missing focus target not to focus an inactive component")
	}
	if secondFocused() {
		t.Fatal("expected missing focus target not to move focus to a different active component")
	}
	if !thirdFocused() {
		t.Fatal("expected missing focus target to leave the current component focused")
	}
}

func TestUseInputReusesHookSlotsAcrossRenders(t *testing.T) {
	ctx := hooks.NewContext()

	callback1 := func(input interface{}, keys []string) bool {
		return false
	}
	callback2 := func(input interface{}, keys []string) bool {
		return false
	}

	hooks.UseInput(ctx, callback1, nil, true)
	if len(ctx.GetInputHooks()) != 1 {
		t.Fatalf("expected one input hook on first render, got %d", len(ctx.GetInputHooks()))
	}

	ctx.Reset()
	hooks.UseInput(ctx, callback2, nil, true)

	inputHooks := ctx.GetInputHooks()
	if len(inputHooks) != 1 {
		t.Fatalf("expected one input hook after rerender, got %d", len(inputHooks))
	}
}

func TestUseFocusReusesHookSlotsAcrossRenders(t *testing.T) {
	ctx := hooks.NewContext()

	firstID := "rerender-focus-1"
	secondID := "rerender-focus-2"

	hooks.UseFocus(ctx, firstID, true, true)
	ctx.FinalizeRender()

	if !ctx.FocusManager().IsFocused(focus.FocusID(firstID)) {
		t.Fatal("expected first focus id to be focused")
	}

	ctx.Reset()
	hooks.UseFocus(ctx, secondID, false, true)
	ctx.FinalizeRender()

	if ctx.FocusManager().FocusableCount() != 1 {
		t.Fatalf("expected exactly one focusable after rerender, got %d", ctx.FocusManager().FocusableCount())
	}
	if ctx.FocusManager().IsFocused(focus.FocusID(firstID)) {
		t.Fatal("expected first focus id to be removed after rerender")
	}
}

func TestUseFocusDisablingFocusManagementPreservesVisibleFocus(t *testing.T) {
	ctx := hooks.NewContext()

	firstFocused, focusFirst, _ := hooks.UseFocus(ctx, "focus-enabled-1", true, true)
	secondFocused, _, _ := hooks.UseFocus(ctx, "focus-enabled-2", false, true)

	if !firstFocused() {
		t.Fatal("expected focus to be visible while enabled")
	}

	ctx.SetFocusEnabled(false)
	if !firstFocused() {
		t.Fatal("expected disabling focus management not to hide the current focus")
	}

	focusFirst("focus-enabled-2")
	if !secondFocused() {
		t.Fatal("expected programmatic focus to remain visible while focus management is disabled")
	}

	ctx.SetFocusEnabled(true)
	if !secondFocused() {
		t.Fatal("expected focused state to remain visible after re-enabling focus management")
	}
}

// TestUseEffect tests useEffect hook
func TestUseEffect(t *testing.T) {
	ctx := hooks.NewContext()

	runCount := 0
	effect := func() func() {
		runCount++
		return nil
	}

	// First run - should execute
	hooks.UseEffect(ctx, effect, nil)
	ctx.RunEffects()
	if runCount != 1 {
		t.Errorf("Expected effect to run once, got %d", runCount)
	}

	// Reset for next render
	ctx.Reset()

	// Second run with nil deps - should run again
	hooks.UseEffect(ctx, effect, nil)
	ctx.RunEffects()
	if runCount != 2 {
		t.Errorf("Expected effect to run again, got %d", runCount)
	}
}

// TestUseEffectWithDeps tests useEffect with dependency array
func TestUseEffectWithDeps(t *testing.T) {
	ctx := hooks.NewContext()

	runCount := 0
	effect := func() func() {
		runCount++
		return nil
	}

	// First run with deps
	hooks.UseEffect(ctx, effect, []interface{}{"a"})
	ctx.RunEffects()
	if runCount != 1 {
		t.Errorf("Expected effect to run once, got %d", runCount)
	}

	ctx.Reset()

	// Second run with same deps - should not run
	hooks.UseEffect(ctx, effect, []interface{}{"a"})
	ctx.RunEffects()
	if runCount != 1 {
		t.Errorf("Expected effect to not run, got %d", runCount)
	}

	ctx.Reset()

	// Third run with different deps - should run
	hooks.UseEffect(ctx, effect, []interface{}{"b"})
	ctx.RunEffects()
	if runCount != 2 {
		t.Errorf("Expected effect to run again, got %d", runCount)
	}
}

// TestUseEffectEmptyDeps tests useEffect with empty dependency array
func TestUseEffectEmptyDeps(t *testing.T) {
	ctx := hooks.NewContext()

	runCount := 0
	effect := func() func() {
		runCount++
		return nil
	}

	// First run with empty deps
	hooks.UseEffect(ctx, effect, []interface{}{})
	ctx.RunEffects()
	if runCount != 1 {
		t.Errorf("Expected effect to run once, got %d", runCount)
	}

	ctx.Reset()

	// Second run with empty deps - should not run
	hooks.UseEffect(ctx, effect, []interface{}{})
	ctx.RunEffects()
	if runCount != 1 {
		t.Errorf("Expected effect to not run, got %d", runCount)
	}
}

// TestUseEffectCleanup tests useEffect cleanup function
func TestUseEffectCleanup(t *testing.T) {
	ctx := hooks.NewContext()

	cleanupRun := false
	effect := func() func() {
		return func() {
			cleanupRun = true
		}
	}

	// First run
	hooks.UseEffect(ctx, effect, []interface{}{"a"})
	ctx.RunEffects()

	ctx.Reset()

	// Second run with different deps - should run cleanup
	hooks.UseEffect(ctx, effect, []interface{}{"b"})
	ctx.RunEffects()
	if !cleanupRun {
		t.Error("Expected cleanup to run")
	}
}

func TestUseEffectRunsAfterRenderPhase(t *testing.T) {
	ctx := hooks.NewContext()
	runCount := 0

	hooks.UseEffect(ctx, func() func() {
		runCount++
		return nil
	}, []interface{}{"effect"})

	if runCount != 0 {
		t.Fatalf("expected effect not to run until RunEffects, got %d", runCount)
	}

	ctx.RunEffects()
	if runCount != 1 {
		t.Fatalf("expected effect to run once after RunEffects, got %d", runCount)
	}
}

// TestUseMemo tests useMemo hook
func TestUseMemo(t *testing.T) {
	ctx := hooks.NewContext()

	computeCount := 0
	compute := func() interface{} {
		computeCount++
		return 42
	}

	// First call - should compute
	result := hooks.UseMemo(ctx, compute, []interface{}{"a"})
	if result != 42 {
		t.Errorf("Expected result 42, got %v", result)
	}
	if computeCount != 1 {
		t.Errorf("Expected compute to run once, got %d", computeCount)
	}

	ctx.Reset()

	// Second call with same deps - should use memoized value
	result = hooks.UseMemo(ctx, compute, []interface{}{"a"})
	if result != 42 {
		t.Errorf("Expected result 42, got %v", result)
	}
	if computeCount != 1 {
		t.Errorf("Expected compute to not run again, got %d", computeCount)
	}
}

// TestUseMemoWithDepsChange tests useMemo when deps change
func TestUseMemoWithDepsChange(t *testing.T) {
	ctx := hooks.NewContext()

	computeCount := 0
	compute := func() interface{} {
		computeCount++
		return computeCount * 10
	}

	// First call
	hooks.UseMemo(ctx, compute, []interface{}{1})

	ctx.Reset()

	// Second call with different deps - should recompute
	result := hooks.UseMemo(ctx, compute, []interface{}{2})
	if result != 20 {
		t.Errorf("Expected result 20, got %v", result)
	}
}

// TestUseCallback tests useCallback hook
func TestUseCallback(t *testing.T) {
	ctx := hooks.NewContext()

	callCount := 0
	callback := func() {
		callCount++
	}
	deps := []interface{}{"a"}

	// First call
	cb1 := hooks.UseCallback(ctx, callback, deps)
	if cb1 == nil {
		t.Error("Expected non-nil callback")
	}

	ctx.Reset()

	// Second call with same deps - should return memoized callback
	cb2 := hooks.UseCallback(ctx, callback, deps)
	if cb2 == nil {
		t.Error("Expected non-nil callback")
	}

	// Both should be callable functions
	cb1()
	if callCount != 1 {
		t.Errorf("Expected callback to be called, got %d", callCount)
	}
}

// TestUseCallbackWithDepsChange tests useCallback when deps change
func TestUseCallbackWithDepsChange(t *testing.T) {
	ctx := hooks.NewContext()

	callback := func() {}

	// First call
	_ = hooks.UseCallback(ctx, callback, []interface{}{1})

	ctx.Reset()

	// Second call with different deps - should work without panic
	cb := hooks.UseCallback(ctx, callback, []interface{}{2})
	if cb == nil {
		t.Error("Expected non-nil callback")
	}

	// Verify it's callable
	cb()
}

// TestUseRef tests useRef hook
func TestUseRef(t *testing.T) {
	ctx := hooks.NewContext()

	ref := hooks.UseRef(ctx, "initial")

	if ref == nil {
		t.Fatal("Expected non-nil ref")
	}

	if ref.Current() != "initial" {
		t.Errorf("Expected initial value 'initial', got %v", ref.Current())
	}

	// Update ref value
	ref.SetCurrent("updated")

	if ref.Current() != "updated" {
		t.Errorf("Expected updated value 'updated', got %v", ref.Current())
	}

	ctx.Reset()

	// Ref should persist across renders
	ref2 := hooks.UseRef(ctx, "ignored")
	if ref2.Current() != "updated" {
		t.Errorf("Expected ref to persist, got %v", ref2.Current())
	}
}

// TestUseRefWithNil tests useRef with nil initial value
func TestUseRefWithNil(t *testing.T) {
	ctx := hooks.NewContext()

	ref := hooks.UseRef(ctx, nil)

	if ref.Current() != nil {
		t.Errorf("Expected nil, got %v", ref.Current())
	}
}

// TestRunCleanup tests RunCleanup function
func TestRunCleanup(t *testing.T) {
	ctx := hooks.NewContext()

	cleanupCalled := false
	effect := func() func() {
		return func() {
			cleanupCalled = true
		}
	}

	hooks.UseEffect(ctx, effect, []interface{}{"a"})
	ctx.Reset()
	hooks.UseEffect(ctx, effect, []interface{}{"b"})
	ctx.RunEffects()

	// Run cleanup
	ctx.RunCleanup()

	if !cleanupCalled {
		t.Error("Expected cleanup to be called")
	}
}

func TestFinalizeRenderCleansUnusedEffects(t *testing.T) {
	ctx := hooks.NewContext()
	cleanupCalled := false

	hooks.UseEffect(ctx, func() func() {
		return func() {
			cleanupCalled = true
		}
	}, []interface{}{"effect"})
	ctx.RunEffects()

	ctx.Reset()
	ctx.FinalizeRender()

	if !cleanupCalled {
		t.Fatal("expected finalize render to clean up dropped effects")
	}
}
