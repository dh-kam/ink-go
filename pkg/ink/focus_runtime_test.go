package ink

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

type focusRuntimeProps struct {
	autoFocus       bool
	disabled        bool
	focusNext       bool
	focusPrevious   bool
	unmountChildren bool
}

func newFocusRuntimeApp(props *focusRuntimeProps) *App {
	return NewApp(func() *vdom.Node {
		manager := UseFocusManager()

		UseEffect(func() func() {
			if props.disabled {
				manager.DisableFocus()
			} else {
				manager.EnableFocus()
			}

			return nil
		}, []interface{}{"focus-disabled", props.disabled})

		UseEffect(func() func() {
			if props.focusNext {
				manager.FocusNext()
			}

			return nil
		}, []interface{}{"focus-next", props.focusNext})

		UseEffect(func() func() {
			if props.focusPrevious {
				manager.FocusPrevious()
			}

			return nil
		}, []interface{}{"focus-previous", props.focusPrevious})

		if props.unmountChildren {
			return nil
		}

		children := []*vdom.Node{
			focusRuntimeItem("First", props.autoFocus),
			focusRuntimeItem("Second", props.autoFocus),
			focusRuntimeItem("Third", props.autoFocus),
		}

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"}, children...)
	})
}

func focusRuntimeItem(label string, autoFocus bool) *vdom.Node {
	isFocused, _, _ := UseFocus(FocusOptions{AutoFocus: autoFocus})
	if isFocused() {
		label += " *"
	}

	return components.Text(label)
}

func TestFocusRuntimeFocusPreviousWrapsFromFirstToLast(t *testing.T) {
	props := &focusRuntimeProps{autoFocus: true}
	app := newFocusRuntimeApp(props)

	initial := app.RenderOnce()
	if initial != "First *\nSecond\nThird" {
		t.Fatalf("expected initial autoFocus output, got %q", initial)
	}

	props.focusPrevious = true
	beforeEffect := app.RenderOnce()
	afterEffect := app.RenderOnce()

	if beforeEffect != "First *\nSecond\nThird" {
		t.Fatalf("expected focusPrevious to apply after the render pass, got %q", beforeEffect)
	}
	if afterEffect != "First\nSecond\nThird *" {
		t.Fatalf("expected focusPrevious to wrap to the last child, got %q", afterEffect)
	}
}

func TestFocusRuntimeFocusNextWrapsFromLastToFirst(t *testing.T) {
	props := &focusRuntimeProps{autoFocus: true}
	app := newFocusRuntimeApp(props)

	if output := app.RenderOnce(); output != "First *\nSecond\nThird" {
		t.Fatalf("expected initial autoFocus output, got %q", output)
	}

	props.focusPrevious = true
	_ = app.RenderOnce()
	if output := app.RenderOnce(); output != "First\nSecond\nThird *" {
		t.Fatalf("expected focusPrevious setup to move focus to the last child, got %q", output)
	}

	props.focusPrevious = false
	props.focusNext = true
	beforeEffect := app.RenderOnce()
	afterEffect := app.RenderOnce()

	if beforeEffect != "First\nSecond\nThird *" {
		t.Fatalf("expected focusNext to apply after the render pass, got %q", beforeEffect)
	}
	if afterEffect != "First *\nSecond\nThird" {
		t.Fatalf("expected focusNext to wrap back to the first child, got %q", afterEffect)
	}
}

func TestFocusRuntimeFocusNextWithUnmountedChildrenIsSafe(t *testing.T) {
	props := &focusRuntimeProps{autoFocus: true}
	app := newFocusRuntimeApp(props)

	if output := app.RenderOnce(); output != "First *\nSecond\nThird" {
		t.Fatalf("expected initial autoFocus output, got %q", output)
	}

	props.focusNext = true
	props.unmountChildren = true

	if output := app.RenderOnce(); output != "" {
		t.Fatalf("expected focusNext on unmounted children to render empty output, got %q", output)
	}
	if output := app.RenderOnce(); output != "" {
		t.Fatalf("expected follow-up render after unmounted-child focusNext to stay empty, got %q", output)
	}
}

func TestFocusRuntimeFocusPreviousWithUnmountedChildrenIsSafe(t *testing.T) {
	props := &focusRuntimeProps{autoFocus: true}
	app := newFocusRuntimeApp(props)

	if output := app.RenderOnce(); output != "First *\nSecond\nThird" {
		t.Fatalf("expected initial autoFocus output, got %q", output)
	}

	props.focusPrevious = true
	props.unmountChildren = true

	if output := app.RenderOnce(); output != "" {
		t.Fatalf("expected focusPrevious on unmounted children to render empty output, got %q", output)
	}
	if output := app.RenderOnce(); output != "" {
		t.Fatalf("expected follow-up render after unmounted-child focusPrevious to stay empty, got %q", output)
	}
}

func TestFocusRuntimeTabFocusesFirstComponentWhenNothingIsFocusedLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("focus-runtime-tab-first", false)
		secondFocused, _, _ := UseFocus("focus-runtime-tab-second", false)
		thirdFocused, _, _ := UseFocus("focus-runtime-tab-third", false)

		firstLabel := "First"
		secondLabel := "Second"
		thirdLabel := "Third"
		if firstFocused() {
			firstLabel += " *"
		}
		if secondFocused() {
			secondLabel += " *"
		}
		if thirdFocused() {
			thirdLabel += " *"
		}

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			components.Text(firstLabel),
			components.Text(secondLabel),
			components.Text(thirdLabel),
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if got := stdout.last(); got != "First\nSecond\nThird" {
		t.Fatalf("expected initial unfocused output, got %q", got)
	}

	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("tab navigation failed: %v", err)
	}

	if got := stdout.last(); got != "First *\nSecond\nThird" {
		t.Fatalf("expected tab with no active focus to move focus to the first component, got %q", got)
	}
}

func TestFocusRuntimeFocusNextAdvancesPastExplicitEmptyStringID(t *testing.T) {
	phase := 0

	app := NewApp(func() *vdom.Node {
		manager := UseFocusManager()
		_, _, _ = UseFocus("", true)
		secondFocused, _, _ := UseFocus("focus-runtime-after-empty", false)

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				manager.FocusNext()
			}

			return nil
		}, []interface{}{phase})

		firstLabel := "First"
		secondLabel := "Second"
		if secondFocused() {
			secondLabel += " *"
		}

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			components.Text(firstLabel),
			components.Text(secondLabel),
		)
	})

	first := app.RenderOnce()
	second := app.RenderOnce()

	if first != "First\nSecond" {
		t.Fatalf("expected explicit empty string id to stay visually unfocused on first render, got %q", first)
	}
	if second != "First\nSecond *" {
		t.Fatalf("expected focusNext to advance to the next visible focus target after an explicit empty string id, got %q", second)
	}
}

func TestFocusRuntimeFocusEmptyStringTargetHidesVisibleFocusOnNextRender(t *testing.T) {
	phase := 0

	app := NewApp(func() *vdom.Node {
		firstFocused, focusFirst, _ := UseFocus("focus-runtime-visible-first", true)
		_, _, _ = UseFocus("", false)

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				focusFirst("")
			}

			return nil
		}, []interface{}{phase})

		label := "First"
		if firstFocused() {
			label += " *"
		}

		return components.Text(label)
	})

	first := app.RenderOnce()
	second := app.RenderOnce()

	if first != "First *" {
		t.Fatalf("expected first render to show the visible focused target, got %q", first)
	}
	if second != "First" {
		t.Fatalf("expected focus(\"\") to move focus to the hidden explicit empty-string target on the next render, got %q", second)
	}
	if !app.hooksCtx.FocusManager().HasFocus() {
		t.Fatal("expected hidden explicit empty-string target to remain focused internally")
	}
}

func TestFocusRuntimeProgrammaticNavigationStillWorksWhileDisabled(t *testing.T) {
	props := &focusRuntimeProps{autoFocus: true}
	app := newFocusRuntimeApp(props)

	if output := app.RenderOnce(); output != "First *\nSecond\nThird" {
		t.Fatalf("expected initial autoFocus output, got %q", output)
	}

	props.disabled = true
	props.focusNext = true
	beforeEffect := app.RenderOnce()
	afterNext := app.RenderOnce()

	if beforeEffect != "First *\nSecond\nThird" {
		t.Fatalf("expected disabled focusNext to apply after the render pass, got %q", beforeEffect)
	}
	if afterNext != "First\nSecond *\nThird" {
		t.Fatalf("expected programmatic focusNext to keep working while disabled, got %q", afterNext)
	}

	props.focusNext = false
	props.focusPrevious = true
	beforePrevious := app.RenderOnce()
	afterPrevious := app.RenderOnce()

	if beforePrevious != "First\nSecond *\nThird" {
		t.Fatalf("expected disabled focusPrevious to apply after the render pass, got %q", beforePrevious)
	}
	if afterPrevious != "First *\nSecond\nThird" {
		t.Fatalf("expected programmatic focusPrevious to keep working while disabled, got %q", afterPrevious)
	}
}

func TestFocusRuntimeMissingTargetKeepsCurrentFocusAndDoesNotRequestRerender(t *testing.T) {
	phase := 0

	app := NewApp(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("focus-runtime-missing-first", true)
		secondFocused, _, _ := UseFocus("focus-runtime-missing-second", false)
		manager := UseFocusManager()

		UseEffect(func() func() {
			switch phase {
			case 0:
				phase = 1
				manager.Focus("focus-runtime-missing-second")
			case 1:
				phase = 2
				manager.Focus("missing-focus-id")
			}

			return nil
		}, []interface{}{phase})

		firstLabel := "First"
		secondLabel := "Second"
		if firstFocused() {
			firstLabel += " *"
		}
		if secondFocused() {
			secondLabel += " *"
		}

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			components.Text(firstLabel),
			components.Text(secondLabel),
		)
	})

	first := app.RenderOnce()
	if first != "First *\nSecond" {
		t.Fatalf("expected first render to show the initial focus, got %q", first)
	}
	if !app.consumeStateChange() {
		t.Fatal("expected focusing a valid target from an effect to request a rerender")
	}

	second := app.RenderOnce()
	if second != "First\nSecond *" {
		t.Fatalf("expected second render to show the newly focused target, got %q", second)
	}
	if app.consumeStateChange() {
		t.Fatal("expected missing focus target not to request a rerender when focus is unchanged")
	}
	if got := app.hooksCtx.FocusManager().FocusedID(); got != "focus-runtime-missing-second" {
		t.Fatalf("expected missing focus target to leave the current focus unchanged, got %q", got)
	}

	third := app.RenderOnce()
	if third != "First\nSecond *" {
		t.Fatalf("expected follow-up render after missing target to remain unchanged, got %q", third)
	}
}

func TestFocusRuntimeEscapeDoesNotBlurWhileDisabled(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("focus-runtime-escape", true)
		manager := UseFocusManager()
		manager.DisableFocus()

		label := "none"
		if firstFocused() {
			label = "first"
		}

		return components.Text(label)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "first") {
		t.Fatalf("expected initial focused output before disabled escape input, got %q", joined)
	}
	if strings.Contains(joined, "none") {
		t.Fatalf("expected disabled focus management to ignore escape blur, got %q", joined)
	}
}

func TestFocusRuntimeEscapeBlursExplicitEmptyStringTargetBeforeTabNavigation(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("focus-runtime-visible-first", false)
		_, _, _ = UseFocus("", true)
		thirdFocused, _, _ := UseFocus("focus-runtime-visible-third", false)

		firstLabel := "First"
		thirdLabel := "Third"
		if firstFocused() {
			firstLabel += " *"
		}
		if thirdFocused() {
			thirdLabel += " *"
		}

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			components.Text(firstLabel),
			components.Text(thirdLabel),
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if got := stdout.last(); got != "First\nThird" {
		t.Fatalf("expected explicit empty string autoFocus to stay visually hidden, got %q", got)
	}

	if err := instance.HandleInput("\x1b"); err != nil {
		t.Fatalf("escape blur failed: %v", err)
	}
	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("tab navigation failed: %v", err)
	}

	if got := stdout.last(); got != "First *\nThird" {
		t.Fatalf("expected tab after escape blur to focus the first visible target, got %q", got)
	}
}

func TestFocusRuntimeResetsFocusWhenFocusedComponentUnregisters(t *testing.T) {
	stdout := &recordingWriter{}
	showFirst := true

	render := func() *vdom.Node {
		children := make([]*vdom.Node, 0, 3)
		if showFirst {
			firstFocused, _, _ := UseFocus("focus-runtime-reset-first", true)
			label := "First"
			if firstFocused() {
				label += " *"
			}
			children = append(children, components.Text(label))
		}

		secondFocused, _, _ := UseFocus("focus-runtime-reset-second", false)
		thirdFocused, _, _ := UseFocus("focus-runtime-reset-third", false)

		secondLabel := "Second"
		thirdLabel := "Third"
		if secondFocused() {
			secondLabel += " *"
		}
		if thirdFocused() {
			thirdLabel += " *"
		}

		children = append(children, components.Text(secondLabel), components.Text(thirdLabel))
		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"}, children...)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if got := stdout.last(); got != "First *\nSecond\nThird" {
		t.Fatalf("expected initial autoFocus output, got %q", got)
	}

	showFirst = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if got := stdout.last(); got != "Second\nThird" {
		t.Fatalf("expected focus reset after unregistering the focused component, got %q", got)
	}
}

func TestFocusRuntimeTabsToFirstRemainingComponentAfterFocusedUnregisters(t *testing.T) {
	stdout := &recordingWriter{}
	showFirst := true

	render := func() *vdom.Node {
		children := make([]*vdom.Node, 0, 3)
		if showFirst {
			firstFocused, _, _ := UseFocus("focus-runtime-unregister-first", true)
			label := "First"
			if firstFocused() {
				label += " *"
			}
			children = append(children, components.Text(label))
		}

		secondFocused, _, _ := UseFocus("focus-runtime-unregister-second", false)
		thirdFocused, _, _ := UseFocus("focus-runtime-unregister-third", false)

		secondLabel := "Second"
		thirdLabel := "Third"
		if secondFocused() {
			secondLabel += " *"
		}
		if thirdFocused() {
			thirdLabel += " *"
		}

		children = append(children, components.Text(secondLabel), components.Text(thirdLabel))
		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"}, children...)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	showFirst = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if got := stdout.last(); got != "Second\nThird" {
		t.Fatalf("expected focus reset after unregistering the focused component, got %q", got)
	}

	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("tab navigation failed: %v", err)
	}

	if got := stdout.last(); got != "Second *\nThird" {
		t.Fatalf("expected tab after unregister to focus the first remaining component, got %q", got)
	}
}

func TestFocusRuntimeToggleFocusManagementBlocksTabUntilReenabled(t *testing.T) {
	stdout := &recordingWriter{}
	disabled := false

	render := func() *vdom.Node {
		manager := UseFocusManager()
		UseEffect(func() func() {
			if disabled {
				manager.DisableFocus()
			} else {
				manager.EnableFocus()
			}

			return nil
		}, []interface{}{"focus-runtime-disabled", disabled})

		firstFocused, _, _ := UseFocus(FocusOptions{AutoFocus: true})
		secondFocused, _, _ := UseFocus(FocusOptions{AutoFocus: true})
		thirdFocused, _, _ := UseFocus(FocusOptions{AutoFocus: true})

		firstLabel := "First"
		secondLabel := "Second"
		thirdLabel := "Third"
		if firstFocused() {
			firstLabel += " *"
		}
		if secondFocused() {
			secondLabel += " *"
		}
		if thirdFocused() {
			thirdLabel += " *"
		}

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			components.Text(firstLabel),
			components.Text(secondLabel),
			components.Text(thirdLabel),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if got := stdout.last(); got != "First *\nSecond\nThird" {
		t.Fatalf("expected initial autoFocus output, got %q", got)
	}

	disabled = true
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("disable rerender failed: %v", err)
	}

	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("tab while disabled failed: %v", err)
	}

	if got := stdout.last(); got != "First *\nSecond\nThird" {
		t.Fatalf("expected tab while disabled to leave focus unchanged, got %q", got)
	}

	disabled = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("reenable rerender failed: %v", err)
	}

	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("tab after reenable failed: %v", err)
	}

	if got := stdout.last(); got != "First\nSecond *\nThird" {
		t.Fatalf("expected tab after reenable to advance focus, got %q", got)
	}
}

func TestFocusRuntimeActiveIDIsSetImmediatelyWhenComponentUsesAutoFocus(t *testing.T) {
	var activeID *string

	app := NewApp(func() *vdom.Node {
		_, _, _ = UseFocus("focus-runtime-activeid-auto", true)
		manager := UseFocusManager()
		activeID = manager.ActiveID

		return components.Text("focus")
	})

	app.RenderOnce()

	if activeID == nil {
		t.Fatal("expected active id snapshot for auto-focused component")
	}
	if *activeID != "focus-runtime-activeid-auto" {
		t.Fatalf("expected auto-focused active id, got %q", *activeID)
	}
}

func TestFocusRuntimeActiveIDUpdatesWhenFocusChangesProgrammatically(t *testing.T) {
	phase := 0
	var capturedActiveID *string

	app := NewApp(func() *vdom.Node {
		_, _, _ = UseFocus("focus-runtime-activeid-first", false)
		_, _, _ = UseFocus("focus-runtime-activeid-second", false)
		manager := UseFocusManager()
		capturedActiveID = manager.ActiveID

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				manager.Focus("focus-runtime-activeid-second")
			}

			return nil
		}, []interface{}{phase})

		return components.Text("focus")
	})

	app.RenderOnce()
	if capturedActiveID != nil {
		t.Fatalf("expected nil active id before programmatic focus, got %q", *capturedActiveID)
	}

	app.RenderOnce()

	if capturedActiveID == nil {
		t.Fatal("expected active id snapshot after programmatic focus")
	}
	if *capturedActiveID != "focus-runtime-activeid-second" {
		t.Fatalf("expected programmatic focus to update active id, got %q", *capturedActiveID)
	}
}

func TestFocusRuntimeMissingTargetLeavesActiveIDUnchangedAndDoesNotRequestRerender(t *testing.T) {
	phase := 0
	var capturedActiveID *string

	app := NewApp(func() *vdom.Node {
		_, _, _ = UseFocus("focus-runtime-activeid-missing-first", true)
		_, _, _ = UseFocus("focus-runtime-activeid-missing-second", false)
		manager := UseFocusManager()
		capturedActiveID = manager.ActiveID

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				manager.Focus("missing-focus-id")
			}

			return nil
		}, []interface{}{phase})

		return components.Text("focus")
	})

	app.RenderOnce()

	if capturedActiveID == nil {
		t.Fatal("expected active id before focusing a missing target")
	}
	if *capturedActiveID != "focus-runtime-activeid-missing-first" {
		t.Fatalf("expected initial active id to stay on the focused component, got %q", *capturedActiveID)
	}
	if app.consumeStateChange() {
		t.Fatal("expected missing focus target not to request a rerender")
	}

	app.RenderOnce()

	if capturedActiveID == nil {
		t.Fatal("expected active id to remain set after focusing a missing target")
	}
	if *capturedActiveID != "focus-runtime-activeid-missing-first" {
		t.Fatalf("expected missing focus target to leave active id unchanged, got %q", *capturedActiveID)
	}
}

func TestFocusRuntimeActiveIDResetsToNilOnEscape(t *testing.T) {
	stdout := &recordingWriter{}
	var capturedActiveID *string

	instance, err := MountWithOptions(func() *vdom.Node {
		_, _, _ = UseFocus("focus-runtime-activeid-escape", true)
		manager := UseFocusManager()
		capturedActiveID = manager.ActiveID

		return components.Text("focus")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if capturedActiveID == nil || *capturedActiveID != "focus-runtime-activeid-escape" {
		if capturedActiveID == nil {
			t.Fatal("expected active id before escape blur")
		}
		t.Fatalf("expected focused active id before escape blur, got %q", *capturedActiveID)
	}

	if err := instance.HandleInput("\x1b"); err != nil {
		t.Fatalf("escape blur failed: %v", err)
	}

	if capturedActiveID != nil {
		t.Fatalf("expected active id to reset after escape blur, got %q", *capturedActiveID)
	}
}

func TestFocusRuntimeActiveIDResetsToNilWhenFocusedComponentUnmounts(t *testing.T) {
	showFirst := true
	var capturedActiveID *string

	render := func() *vdom.Node {
		children := make([]*vdom.Node, 0, 2)
		if showFirst {
			_, _, _ = UseFocus("focus-runtime-activeid-unmount", true)
			children = append(children, components.Text("First"))
		}

		_, _, _ = UseFocus("focus-runtime-activeid-second", false)
		manager := UseFocusManager()
		capturedActiveID = manager.ActiveID

		children = append(children, components.Text("Second"))
		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"}, children...)
	}

	instance, err := MountWithOptions(render, RenderOptions{Debug: true})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if capturedActiveID == nil || *capturedActiveID != "focus-runtime-activeid-unmount" {
		if capturedActiveID == nil {
			t.Fatal("expected active id before unmounting the focused component")
		}
		t.Fatalf("expected focused active id before unmounting, got %q", *capturedActiveID)
	}

	showFirst = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if capturedActiveID != nil {
		t.Fatalf("expected active id to reset after unmounting the focused component, got %q", *capturedActiveID)
	}
}
