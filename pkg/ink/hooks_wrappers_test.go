package ink

import (
	"io"
	"testing"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

type rawModeTestStdin struct{}

func boolPtr(value bool) *bool {
	return &value
}

func (rawModeTestStdin) Read(data []byte) (int, error) {
	return 0, io.EOF
}

func (rawModeTestStdin) Fd() int {
	return 0
}

func TestUseInputWrapperRegistersHook(t *testing.T) {
	app := NewApp(func() *vdom.Node {
		cleanup := UseInput(func(input interface{}, keys []string) bool {
			return false
		})
		if cleanup == nil {
			t.Fatal("expected cleanup function")
		}

		return vdom.CreateTextNode("input")
	})

	app.RenderOnce()
	if len(app.hooksCtx.GetInputHooks()) != 1 {
		t.Fatalf("expected one input hook, got %d", len(app.hooksCtx.GetInputHooks()))
	}
}

func TestUseInputWrapperInactiveOptionSkipsHookAndRawMode(t *testing.T) {
	app := NewAppWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {}, InputOptions{IsActive: false})
		return vdom.CreateTextNode("input")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if len(app.hooksCtx.GetInputHooks()) != 0 {
		t.Fatalf("expected no active input hooks, got %d", len(app.hooksCtx.GetInputHooks()))
	}
	if app.rawModeUsers != 0 {
		t.Fatalf("expected raw mode user count to stay at zero, got %d", app.rawModeUsers)
	}
	if app.rawState != nil {
		t.Fatal("expected raw mode to remain disabled for inactive useInput hook")
	}
}

func TestUseInputWrapperRefCountsRawModeAcrossHooks(t *testing.T) {
	app := NewAppWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {})
		UseInput(func(input string, key InputKey) {})
		return vdom.CreateTextNode("input")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if app.rawModeUsers != 2 {
		t.Fatalf("expected two raw mode users after render, got %d", app.rawModeUsers)
	}
	if app.rawState == nil {
		t.Fatal("expected raw mode state to be enabled")
	}

	app.hooksCtx.RunCleanup()

	if app.rawModeUsers != 0 {
		t.Fatalf("expected raw mode users to reach zero after cleanup, got %d", app.rawModeUsers)
	}
	if app.rawState != nil {
		t.Fatal("expected raw mode state to be restored after cleanup")
	}
}

func TestUseStdinWrapperCanDisableAndReenableRawMode(t *testing.T) {
	app := NewAppWithOptions(func() *vdom.Node {
		stdin := UseStdin()

		UseEffect(func() func() {
			if err := stdin.SetRawMode(true); err != nil {
				t.Fatalf("expected first raw mode enable to succeed: %v", err)
			}
			if err := stdin.SetRawMode(false); err != nil {
				t.Fatalf("expected raw mode disable to succeed: %v", err)
			}
			if err := stdin.SetRawMode(true); err != nil {
				t.Fatalf("expected raw mode re-enable to succeed: %v", err)
			}

			return func() {
				if err := stdin.SetRawMode(false); err != nil {
					t.Fatalf("expected cleanup raw mode disable to succeed: %v", err)
				}
			}
		}, []interface{}{"use-stdin-raw-mode-reenable"})

		return vdom.CreateTextNode("stdin")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	output := app.RenderOnce()

	if output != "stdin" {
		t.Fatalf("expected render output %q, got %q", "stdin", output)
	}
	if app.rawModeUsers != 1 {
		t.Fatalf("expected raw mode to remain enabled after disable/re-enable cycle, got %d users", app.rawModeUsers)
	}
	if app.rawState == nil {
		t.Fatal("expected raw mode state to be enabled after re-enable cycle")
	}

	app.hooksCtx.RunCleanup()

	if app.rawModeUsers != 0 {
		t.Fatalf("expected cleanup to fully release raw mode, got %d users", app.rawModeUsers)
	}
	if app.rawState != nil {
		t.Fatal("expected cleanup to restore raw mode state after re-enable cycle")
	}
}

func TestUseFocusWrapper(t *testing.T) {
	var focusedInitially bool
	var focusedAfterBlur bool
	var focusedAfterFocus bool

	app := NewApp(func() *vdom.Node {
		isFocused, focusFn, blurFn := UseFocus("wrapper-focus", true)
		focusedInitially = isFocused()
		blurFn()
		focusedAfterBlur = isFocused()
		focusFn()
		focusedAfterFocus = isFocused()
		return vdom.CreateTextNode("focus")
	})

	app.RenderOnce()

	if !focusedInitially {
		t.Fatal("expected focus hook to autofocus")
	}
	if focusedAfterBlur {
		t.Fatal("expected blur to clear focus")
	}
	if !focusedAfterFocus {
		t.Fatal("expected focus function to restore focus")
	}
}

func TestUseFocusWrapperNilOptionsUsesDefaultBehavior(t *testing.T) {
	var focusedInitially bool
	var blurLeavesFocusOff bool

	app := NewAppWithOptions(func() *vdom.Node {
		isFocused, _, blurFn := UseFocus(nil)
		focusedInitially = isFocused()
		blurFn()
		blurLeavesFocusOff = !isFocused()
		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	output := app.RenderOnce()

	if output != "focus" {
		t.Fatalf("expected default focus render, got %q", output)
	}
	if focusedInitially {
		t.Fatal("expected nil options to behave like default useFocus without autoFocus")
	}
	if !blurLeavesFocusOff {
		t.Fatal("expected blur on an unfocused nil-options hook to keep focus cleared")
	}
	if len(app.hooksCtx.GetFocusHooks()) != 1 {
		t.Fatalf("expected one focus hook from nil options, got %d", len(app.hooksCtx.GetFocusHooks()))
	}
}

func TestUseFocusWrapperPreservesExplicitEmptyStringID(t *testing.T) {
	var focusedInitially bool

	app := NewAppWithOptions(func() *vdom.Node {
		isFocused, _, _ := UseFocus("", true)
		focusedInitially = isFocused()
		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if focusedInitially {
		t.Fatal("expected explicit empty string id to remain unfocused in public hook state")
	}
	order := app.hooksCtx.FocusManager().FocusOrder()
	if len(order) != 1 {
		t.Fatalf("expected one focusable for explicit empty string id, got %d", len(order))
	}
	if order[0] != "" {
		t.Fatalf("expected explicit empty string id to be preserved, got %q", order[0])
	}
}

func TestUseFocusWrapperPreservesExplicitEmptyStringIDAcrossRerenders(t *testing.T) {
	app := NewAppWithOptions(func() *vdom.Node {
		UseFocus("", true)
		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()
	app.RenderOnce()

	order := app.hooksCtx.FocusManager().FocusOrder()
	if len(order) != 1 {
		t.Fatalf("expected one focusable after rerender with explicit empty string id, got %d", len(order))
	}
	if order[0] != "" {
		t.Fatalf("expected explicit empty string id to stay stable across rerenders, got %q", order[0])
	}
}

func TestUseFocusWrapperFocusNextAdvancesPastExplicitEmptyStringID(t *testing.T) {
	var secondFocusedAfterNext bool

	app := NewAppWithOptions(func() *vdom.Node {
		_, _, _ = UseFocus("", true)
		secondFocused, _, _ := UseFocus("wrapper-focus-after-empty", false)
		manager := UseFocusManager()

		manager.FocusNext()
		secondFocusedAfterNext = secondFocused()

		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if !secondFocusedAfterNext {
		t.Fatal("expected focusNext to advance past an explicitly empty string id")
	}
}

func TestUseFocusWrapperCanTargetExplicitEmptyStringID(t *testing.T) {
	var firstFocusedAfterEmptyTarget bool

	app := NewAppWithOptions(func() *vdom.Node {
		firstFocused, focusFirst, _ := UseFocus("wrapper-focus-visible", true)
		_, _, _ = UseFocus("", false)

		focusFirst("")
		firstFocusedAfterEmptyTarget = firstFocused()

		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if firstFocusedAfterEmptyTarget {
		t.Fatal("expected wrapper focus(\"\") to target the explicit empty-string id instead of self")
	}
	if !app.hooksCtx.FocusManager().HasFocus() {
		t.Fatal("expected explicit empty-string wrapper target to still count as focused internally")
	}
	if !app.hooksCtx.FocusManager().IsFocused("") {
		t.Fatal("expected explicit empty-string wrapper target to become focused internally")
	}
}

func TestUseFocusWrapperExplicitEmptyStringFocusFromEffectRequestsRerender(t *testing.T) {
	phase := 0

	app := NewAppWithOptions(func() *vdom.Node {
		_, focusEmpty, _ := UseFocus("", false)

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				focusEmpty()
			}

			return nil
		}, []interface{}{phase})

		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if !app.consumeStateChange() {
		t.Fatal("expected focusing an explicit empty-string id from an effect to request a rerender")
	}
	if !app.hooksCtx.FocusManager().HasFocus() {
		t.Fatal("expected focusing an explicit empty-string id to set internal focus state")
	}
}

func TestUseFocusWrapperExplicitEmptyStringBlurFromEffectRequestsRerender(t *testing.T) {
	phase := 0

	app := NewAppWithOptions(func() *vdom.Node {
		_, _, blurEmpty := UseFocus("", true)

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				blurEmpty()
			}

			return nil
		}, []interface{}{phase})

		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if !app.consumeStateChange() {
		t.Fatal("expected blurring an explicit empty-string id from an effect to request a rerender")
	}
	if app.hooksCtx.FocusManager().HasFocus() {
		t.Fatal("expected blurring an explicit empty-string id to clear internal focus state")
	}
}

func TestUseFocusWrapperMissingTargetKeepsCurrentFocus(t *testing.T) {
	var firstFocusedAfterMissingID bool
	var secondFocusedAfterMissingID bool

	app := NewAppWithOptions(func() *vdom.Node {
		firstFocused, focusFirst, _ := UseFocus("wrapper-focus-missing-1", true)
		secondFocused, focusSecond, _ := UseFocus("wrapper-focus-missing-2", false)

		focusSecond()
		focusFirst("missing-focus-id")
		firstFocusedAfterMissingID = firstFocused()
		secondFocusedAfterMissingID = secondFocused()

		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if firstFocusedAfterMissingID {
		t.Fatal("expected missing focus target not to move focus back to the first component")
	}
	if !secondFocusedAfterMissingID {
		t.Fatal("expected missing focus target to leave the current focus unchanged")
	}
}

func TestUseFocusWrapperSupportsOptionsAndInactiveState(t *testing.T) {
	var firstFocused bool
	var secondFocused bool

	app := NewAppWithOptions(func() *vdom.Node {
		firstState, focusFn, _ := UseFocus(FocusOptions{
			ID:        "wrapper-focus-1",
			AutoFocus: true,
		})
		secondState, _, _ := UseFocus(FocusOptions{
			ID:       "wrapper-focus-2",
			IsActive: boolPtr(false),
		})

		firstFocused = firstState()
		focusFn("wrapper-focus-2")
		secondFocused = secondState()

		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if !firstFocused {
		t.Fatal("expected options-based focus hook to honor autoFocus")
	}
	if !secondFocused {
		t.Fatal("expected explicit focus target to work with options-based focus hook")
	}
	if app.rawModeUsers != 1 {
		t.Fatalf("expected only active focus hook to contribute raw mode, got %d", app.rawModeUsers)
	}
}

func TestUseFocusWrapperReappliesAutoFocusAcrossRenders(t *testing.T) {
	autoFocus := false
	var focusedBefore bool
	var focusedAfter bool

	app := NewAppWithOptions(func() *vdom.Node {
		isFocused, _, _ := UseFocus(FocusOptions{
			ID:        "wrapper-focus-rerender",
			AutoFocus: autoFocus,
		})

		if autoFocus {
			focusedAfter = isFocused()
		} else {
			focusedBefore = isFocused()
		}

		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()
	autoFocus = true
	app.RenderOnce()

	if focusedBefore {
		t.Fatal("expected initial render without autoFocus to stay unfocused")
	}
	if !focusedAfter {
		t.Fatal("expected rerender with autoFocus to focus the component")
	}
}

func TestUseFocusWrapperFocusFromEffectTriggersNextRender(t *testing.T) {
	phase := 0

	app := NewApp(func() *vdom.Node {
		firstFocused, focusFirst, _ := UseFocus("wrapper-focus-effect-1", true)
		secondFocused, _, _ := UseFocus("wrapper-focus-effect-2", false)

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				focusFirst("wrapper-focus-effect-2")
			}

			return nil
		}, []interface{}{phase})

		label := "none"
		switch {
		case firstFocused():
			label = "first"
		case secondFocused():
			label = "second"
		}

		return vdom.CreateTextNode(label)
	})

	first := app.RenderOnce()
	second := app.RenderOnce()

	if first != "first" {
		t.Fatalf("expected first render to show initial focus, got %q", first)
	}
	if second != "second" {
		t.Fatalf("expected second render to reflect effect-driven focus change, got %q", second)
	}
}

func TestUseFocusWrapperBlurFromEffectTriggersNextRender(t *testing.T) {
	phase := 0

	app := NewApp(func() *vdom.Node {
		firstFocused, _, blurFirst := UseFocus("wrapper-focus-blur", true)

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				blurFirst()
			}

			return nil
		}, []interface{}{phase})

		label := "none"
		if firstFocused() {
			label = "first"
		}

		return vdom.CreateTextNode(label)
	})

	first := app.RenderOnce()
	second := app.RenderOnce()

	if first != "first" {
		t.Fatalf("expected first render to show focused state, got %q", first)
	}
	if second != "none" {
		t.Fatalf("expected second render to reflect effect-driven blur, got %q", second)
	}
}

func TestUseFocusWrapperInactiveOptionSkipsRawMode(t *testing.T) {
	app := NewAppWithOptions(func() *vdom.Node {
		UseFocus(FocusOptions{IsActive: boolPtr(false)})
		return vdom.CreateTextNode("focus")
	}, AppOptions{Stdin: rawModeTestStdin{}})

	app.RenderOnce()

	if app.rawModeUsers != 0 {
		t.Fatalf("expected inactive focus hook not to enable raw mode, got %d", app.rawModeUsers)
	}
	if app.rawState != nil {
		t.Fatal("expected inactive focus hook to keep raw mode disabled")
	}
}

func TestUseEffectWrapper(t *testing.T) {
	runCount := 0
	dependency := "a"

	app := NewApp(func() *vdom.Node {
		UseEffect(func() func() {
			runCount++
			return nil
		}, []interface{}{dependency})
		return vdom.CreateTextNode("effect")
	})

	app.RenderOnce()
	app.RenderOnce()
	dependency = "b"
	app.RenderOnce()

	if runCount != 2 {
		t.Fatalf("expected effect to run twice, got %d", runCount)
	}
}

func TestUseMemoWrapper(t *testing.T) {
	computeCount := 0
	dependency := 1

	app := NewApp(func() *vdom.Node {
		value := UseMemo(func() interface{} {
			computeCount++
			return dependency * 2
		}, []interface{}{dependency})

		return vdom.CreateTextNode(string(rune(value.(int) + '0')))
	})

	app.RenderOnce()
	app.RenderOnce()
	dependency = 2
	app.RenderOnce()

	if computeCount != 2 {
		t.Fatalf("expected memo to compute twice, got %d", computeCount)
	}
}

func TestUseCallbackWrapper(t *testing.T) {
	captured := 0
	dependency := 1

	app := NewApp(func() *vdom.Node {
		callback := UseCallback(func() {
			captured = dependency
		}, []interface{}{dependency})

		callback()
		return vdom.CreateTextNode("callback")
	})

	app.RenderOnce()
	if captured != 1 {
		t.Fatalf("expected callback to capture first dependency, got %d", captured)
	}

	dependency = 2
	app.RenderOnce()
	if captured != 2 {
		t.Fatalf("expected callback to update when dependencies change, got %d", captured)
	}
}

func TestUseRefWrapper(t *testing.T) {
	var firstRef *Ref
	var secondRef *Ref

	app := NewApp(func() *vdom.Node {
		ref := UseRef("initial")
		if firstRef == nil {
			firstRef = ref
			ref.SetCurrent("updated")
		} else {
			secondRef = ref
		}

		return vdom.CreateTextNode(ref.Current().(string))
	})

	app.RenderOnce()
	app.RenderOnce()

	if firstRef == nil || secondRef == nil {
		t.Fatal("expected refs to be captured across renders")
	}
	if firstRef != secondRef {
		t.Fatal("expected ref instance to persist across renders")
	}
	if secondRef.Current() != "updated" {
		t.Fatalf("expected persisted ref value, got %v", secondRef.Current())
	}
}

func TestUseInputWrapperReusesHookSlotAcrossRenders(t *testing.T) {
	app := NewApp(func() *vdom.Node {
		UseInput(func(input interface{}, keys []string) bool {
			return false
		})

		return vdom.CreateTextNode("input")
	})

	app.RenderOnce()
	app.RenderOnce()

	if len(app.hooksCtx.GetInputHooks()) != 1 {
		t.Fatalf("expected one active input hook after rerender, got %d", len(app.hooksCtx.GetInputHooks()))
	}
}

func TestUseFocusManagerWrapper(t *testing.T) {
	var focusedInitially bool
	var focusedWhileDisabled bool
	var secondFocusedWhileDisabled bool
	var secondFocusedAfterDisabledNext bool
	var focusedAfterDisabledPrevious bool
	var focusedAfterEnable bool
	var secondFocusedAfterNext bool
	var firstFocusedAfterMissingID bool
	var secondFocusedAfterMissingID bool

	app := NewApp(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("wrapper-focus-1", true)
		secondFocused, _, _ := UseFocus("wrapper-focus-2", false)
		manager := UseFocusManager()

		focusedInitially = firstFocused()
		manager.DisableFocus()
		focusedWhileDisabled = firstFocused()
		manager.FocusNext()
		secondFocusedAfterDisabledNext = secondFocused()
		manager.FocusPrevious()
		focusedAfterDisabledPrevious = firstFocused()
		manager.Focus("wrapper-focus-2")
		secondFocusedWhileDisabled = secondFocused()
		manager.EnableFocus()
		focusedAfterEnable = secondFocused()
		manager.Focus("missing-focus-id")
		firstFocusedAfterMissingID = firstFocused()
		secondFocusedAfterMissingID = secondFocused()
		manager.FocusNext()
		secondFocusedAfterNext = firstFocused()

		return vdom.CreateTextNode("focus-manager")
	})

	app.RenderOnce()

	if !focusedInitially {
		t.Fatal("expected initial focus to be visible")
	}
	if !focusedWhileDisabled {
		t.Fatal("expected disabling focus management not to hide the current focus")
	}
	if !secondFocusedWhileDisabled {
		t.Fatal("expected programmatic focus to remain visible while focus management is disabled")
	}
	if !secondFocusedAfterDisabledNext {
		t.Fatal("expected focusNext to keep working while focus management is disabled")
	}
	if !focusedAfterDisabledPrevious {
		t.Fatal("expected focusPrevious to keep working while focus management is disabled")
	}
	if !focusedAfterEnable {
		t.Fatal("expected focused state to remain visible after re-enabling focus management")
	}
	if !secondFocusedAfterNext {
		t.Fatal("expected focusNext to wrap to the next component after re-enabling focus management")
	}
	if firstFocusedAfterMissingID {
		t.Fatal("expected missing focus target not to change the current focus")
	}
	if !secondFocusedAfterMissingID {
		t.Fatal("expected missing focus target to leave the current target focused")
	}
}

func TestUseFocusManagerWrapperMissingTargetSkipsInactiveFallbacks(t *testing.T) {
	var secondFocusedAfterMissingID bool
	var thirdFocusedAfterMissingID bool
	var inactiveFocusedAfterMissingID bool

	app := NewApp(func() *vdom.Node {
		inactiveFocused, _, _ := UseFocus(FocusOptions{
			ID:       "wrapper-focus-inactive",
			IsActive: boolPtr(false),
		})
		secondFocused, _, _ := UseFocus("wrapper-focus-active-1", false)
		thirdFocused, _, _ := UseFocus("wrapper-focus-active-2", true)
		manager := UseFocusManager()

		manager.Focus("missing-focus-id")
		inactiveFocusedAfterMissingID = inactiveFocused()
		secondFocusedAfterMissingID = secondFocused()
		thirdFocusedAfterMissingID = thirdFocused()

		return vdom.CreateTextNode("focus-manager")
	})

	app.RenderOnce()

	if inactiveFocusedAfterMissingID {
		t.Fatal("expected missing focus target not to focus inactive components")
	}
	if secondFocusedAfterMissingID {
		t.Fatal("expected missing focus target not to move focus to a different active component")
	}
	if !thirdFocusedAfterMissingID {
		t.Fatal("expected missing focus target to leave the current target focused")
	}
}

func TestUseFocusManagerWrapperFocusPreviousWrapsToLast(t *testing.T) {
	var thirdFocusedAfterPrevious bool

	app := NewApp(func() *vdom.Node {
		_, _, _ = UseFocus("wrapper-focus-1", true)
		_, _, _ = UseFocus("wrapper-focus-2", false)
		thirdFocused, _, _ := UseFocus("wrapper-focus-3", false)
		manager := UseFocusManager()

		manager.FocusPrevious()
		thirdFocusedAfterPrevious = thirdFocused()

		return vdom.CreateTextNode("focus-manager")
	})

	app.RenderOnce()

	if !thirdFocusedAfterPrevious {
		t.Fatal("expected focusPrevious to wrap to the last component")
	}
}

func TestUseFocusManagerWrapperActiveIDReflectsAutoFocus(t *testing.T) {
	var activeID *string

	app := NewApp(func() *vdom.Node {
		_, _, _ = UseFocus("wrapper-focus-activeid", true)
		manager := UseFocusManager()
		activeID = manager.ActiveID

		return vdom.CreateTextNode("focus-manager")
	})

	app.RenderOnce()

	if activeID == nil {
		t.Fatal("expected active id snapshot for an auto-focused component")
	}
	if *activeID != "wrapper-focus-activeid" {
		t.Fatalf("expected auto-focused active id, got %q", *activeID)
	}
}

func TestUseFocusManagerWrapperActiveIDPreservesExplicitEmptyStringFocus(t *testing.T) {
	var activeID *string

	app := NewApp(func() *vdom.Node {
		_, _, _ = UseFocus("", true)
		manager := UseFocusManager()
		activeID = manager.ActiveID

		return vdom.CreateTextNode("focus-manager")
	})

	app.RenderOnce()

	if activeID == nil {
		t.Fatal("expected active id snapshot for an explicit empty-string focus target")
	}
	if *activeID != "" {
		t.Fatalf("expected explicit empty-string active id, got %q", *activeID)
	}
}
