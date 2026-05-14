package ink

import (
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// TestUseFocusOptsReturnsParityShape locks in the upstream `useFocus` return
// shape: an object with `IsFocused` (boolean snapshot) and `Focus` (single-arg
// function), matching `{isFocused, focus}` from `ink/src/hooks/use-focus.ts`.
func TestUseFocusOptsReturnsParityShape(t *testing.T) {
	var captured FocusState

	app := NewApp(func() *vdom.Node {
		captured = UseFocusOpts(FocusOptions{ID: "use-focus-opts-parity", AutoFocus: true})
		return components.Text("focus")
	})

	app.RenderOnce()

	if !captured.IsFocused {
		t.Fatal("expected UseFocusOpts to expose a snapshot IsFocused matching upstream {isFocused: boolean}")
	}
	if captured.Focus == nil {
		t.Fatal("expected UseFocusOpts to expose a non-nil Focus(id string) function matching upstream {focus}")
	}
}

// TestUseFocusOptsFocusFunctionTargetsByID exercises the `focus(id)` callback
// returned by upstream's useFocus — calling it with a registered id moves focus
// to that component.
func TestUseFocusOptsFocusFunctionTargetsByID(t *testing.T) {
	phase := 0
	var firstState FocusState
	var secondState FocusState

	app := NewApp(func() *vdom.Node {
		firstState = UseFocusOpts(FocusOptions{ID: "use-focus-opts-target-1", AutoFocus: true})
		secondState = UseFocusOpts(FocusOptions{ID: "use-focus-opts-target-2"})

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				firstState.Focus("use-focus-opts-target-2")
			}

			return nil
		}, []interface{}{phase})

		return components.Text("focus")
	})

	app.RenderOnce()

	if !firstState.IsFocused || secondState.IsFocused {
		t.Fatalf("expected first to be focused initially, got first=%v second=%v", firstState.IsFocused, secondState.IsFocused)
	}

	app.RenderOnce()

	if firstState.IsFocused || !secondState.IsFocused {
		t.Fatalf("expected Focus(id) to move focus to second, got first=%v second=%v", firstState.IsFocused, secondState.IsFocused)
	}
}

// TestUseFocusOptsMissingTargetIsNoOp matches upstream's `focus` (in App.tsx),
// which silently does nothing if the id is not registered. Focus state stays.
func TestUseFocusOptsMissingTargetIsNoOp(t *testing.T) {
	phase := 0
	var first FocusState

	app := NewApp(func() *vdom.Node {
		first = UseFocusOpts(FocusOptions{ID: "use-focus-opts-missing", AutoFocus: true})

		UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				first.Focus("definitely-not-registered")
			}

			return nil
		}, []interface{}{phase})

		return components.Text("focus")
	})

	app.RenderOnce()
	if !first.IsFocused {
		t.Fatal("expected initial autoFocus")
	}

	app.RenderOnce()

	if !first.IsFocused {
		t.Fatal("expected missing-id Focus call to leave existing focus unchanged (upstream parity)")
	}
}

// TestUseFocusOptsInactiveStaysUnfocusable mirrors upstream:
// `useFocus({isActive: false})` is unfocusable while inactive. Tab navigation
// must skip it, and isFocused stays false even with autoFocus.
func TestUseFocusOptsInactiveStaysUnfocusable(t *testing.T) {
	var inactive FocusState
	var second FocusState
	var third FocusState

	app := NewApp(func() *vdom.Node {
		inactiveFlag := false
		inactive = UseFocusOpts(FocusOptions{
			ID:        "use-focus-opts-inactive",
			AutoFocus: true,
			IsActive:  &inactiveFlag,
		})
		second = UseFocusOpts(FocusOptions{ID: "use-focus-opts-second"})
		third = UseFocusOpts(FocusOptions{ID: "use-focus-opts-third", AutoFocus: true})

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			components.Text("first"),
			components.Text("second"),
			components.Text("third"),
		)
	})

	app.RenderOnce()

	if inactive.IsFocused {
		t.Fatal("expected isActive:false to remain unfocusable even with autoFocus (upstream parity)")
	}
	if second.IsFocused {
		t.Fatal("expected second to not be focused initially")
	}
	if !third.IsFocused {
		t.Fatal("expected third (next active autoFocus candidate) to take focus")
	}
}

// TestUseFocusOptsToggleIsActiveRejoinsFocusOrder verifies upstream behavior:
// flipping isActive false -> true restores the component in its original
// registration position so tab navigation reaches it again.
func TestUseFocusOptsToggleIsActiveRejoinsFocusOrder(t *testing.T) {
	active := true

	render := func() *vdom.Node {
		_ = UseFocusOpts(FocusOptions{ID: "rejoin-1", AutoFocus: true})
		_ = UseFocusOpts(FocusOptions{ID: "rejoin-2", IsActive: &active})
		_ = UseFocusOpts(FocusOptions{ID: "rejoin-3"})
		return components.Text("focus")
	}

	stdout := &recordingWriter{}
	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	manager := instance.app.focusManager()

	// Initially focus is on rejoin-1 due to autoFocus.
	if got := manager.FocusedID(); string(got) != "rejoin-1" {
		t.Fatalf("expected rejoin-1 to autofocus, got %q", got)
	}

	// Deactivate rejoin-2.
	active = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("deactivate rerender failed: %v", err)
	}

	// Tab forward: should skip rejoin-2 -> land on rejoin-3.
	if !manager.FocusNext() {
		t.Fatal("expected FocusNext to advance past inactive rejoin-2")
	}
	if got := manager.FocusedID(); string(got) != "rejoin-3" {
		t.Fatalf("expected FocusNext past inactive to land on rejoin-3, got %q", got)
	}

	// Re-activate rejoin-2 and confirm it rejoins focus order in original
	// position (between rejoin-1 and rejoin-3).
	active = true
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("reactivate rerender failed: %v", err)
	}

	// Move focus back to rejoin-1, then tab forward — should land on rejoin-2.
	if !manager.Focus("rejoin-1") {
		t.Fatal("expected Focus(rejoin-1) to succeed")
	}
	if !manager.FocusNext() {
		t.Fatal("expected FocusNext after reactivate to advance")
	}
	if got := manager.FocusedID(); string(got) != "rejoin-2" {
		t.Fatalf("expected reactivated rejoin-2 to be reachable in original position, got %q", got)
	}
}

// TestUseFocusManagerDisablePreservesFocusedID locks upstream parity: calling
// disableFocus must NOT clear the active focus id; it only blocks tab handling.
// Re-enabling restores tab navigation seamlessly.
func TestUseFocusManagerDisablePreservesFocusedID(t *testing.T) {
	disabled := false

	render := func() *vdom.Node {
		_ = UseFocusOpts(FocusOptions{ID: "disable-preserve-1", AutoFocus: true})
		_ = UseFocusOpts(FocusOptions{ID: "disable-preserve-2"})

		manager := UseFocusManager()
		UseEffect(func() func() {
			if disabled {
				manager.DisableFocus()
			} else {
				manager.EnableFocus()
			}
			return nil
		}, []interface{}{disabled})

		return components.Text("focus")
	}

	instance, err := MountWithOptions(render, RenderOptions{Debug: true})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	manager := instance.app.focusManager()
	if string(manager.FocusedID()) != "disable-preserve-1" {
		t.Fatalf("expected initial autoFocus on disable-preserve-1, got %q", manager.FocusedID())
	}

	// Disable focus management and confirm focused id is preserved.
	disabled = true
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("disable rerender failed: %v", err)
	}

	if string(manager.FocusedID()) != "disable-preserve-1" {
		t.Fatalf("expected disableFocus to preserve focused id (upstream parity), got %q", manager.FocusedID())
	}

	// Re-enable; tab should resume.
	disabled = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("enable rerender failed: %v", err)
	}

	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("tab failed: %v", err)
	}

	if string(manager.FocusedID()) != "disable-preserve-2" {
		t.Fatalf("expected tab after reenable to advance to disable-preserve-2, got %q", manager.FocusedID())
	}
}
