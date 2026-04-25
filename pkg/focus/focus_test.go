package focus_test

import (
	"fmt"
	"testing"

	"github.com/dh-kam/goink.go/pkg/focus"
)

// TestNewFocusManager tests focus manager creation
func TestNewFocusManager(t *testing.T) {
	fm := focus.NewFocusManager()

	if fm == nil {
		t.Fatal("Expected non-nil focus manager")
	}

	if fm.FocusableCount() != 0 {
		t.Errorf("Expected 0 focusable components, got %d", fm.FocusableCount())
	}
}

// TestRegister tests registering focusable components
func TestRegister(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")

	fm.Register(id1, true)
	fm.Register(id2, false)

	if fm.FocusableCount() != 2 {
		t.Errorf("Expected 2 focusable components, got %d", fm.FocusableCount())
	}
	if !fm.IsActive(id1) || !fm.IsActive(id2) {
		t.Fatal("expected registered components to be active by default")
	}
}

// TestRegisterAutoFocus tests auto-focus on first registration
func TestRegisterAutoFocus(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	fm.Register(id1, true)

	if !fm.IsFocused(id1) {
		t.Error("Expected first component with autoFocus to be focused")
	}
}

// TestUnregister tests unregistering components
func TestUnregister(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")

	fm.Register(id1, true)
	fm.Register(id2, false)

	fm.Unregister(id1)

	if fm.FocusableCount() != 1 {
		t.Errorf("Expected 1 focusable component after unregister, got %d", fm.FocusableCount())
	}

	// Focus should be cleared when focused component is unregistered
	if fm.FocusedID() != "" {
		t.Error("Expected focus to be cleared after unregistering focused component")
	}
}

// TestFocus tests manual focus setting
func TestFocus(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")

	fm.Register(id1, false)
	fm.Register(id2, false)

	// Focus on component-2
	if !fm.Focus(id2) {
		t.Error("Focus() should return true for focusable component")
	}

	if !fm.IsFocused(id2) {
		t.Error("Expected component-2 to be focused")
	}

	// Try to focus non-existent component
	id3 := focus.FocusID("component-3")
	if fm.Focus(id3) {
		t.Error("Focus() should return false when the target is missing")
	}
	if !fm.IsFocused(id2) {
		t.Error("Expected focus() to leave the current focus unchanged when the target is missing")
	}
}

func TestFocusMissingTargetKeepsCurrentFocus(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, false)
	fm.Register(id2, false)
	fm.Register(id3, false)
	fm.Focus(id3)

	if fm.Focus("missing") {
		t.Fatal("expected missing focus target to return false")
	}
	if !fm.IsFocused(id3) {
		t.Fatal("expected missing focus target to leave the current focus unchanged")
	}
}

func TestFocusMissingTargetWithoutCurrentFocusDoesNothing(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")

	fm.Register(id1, false)
	fm.Register(id2, false)

	if fm.Focus("missing") {
		t.Fatal("expected missing focus target to return false when nothing is focused")
	}
	if fm.HasFocus() {
		t.Fatalf("expected missing focus target to leave focus empty, got %q", fm.FocusedID())
	}
}

func TestFocusCanStillTargetInactiveComponentExplicitly(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")

	fm.Register(id1, false)
	fm.Register(id2, false)
	fm.Deactivate(id1)

	if !fm.Focus(id1) {
		t.Fatal("expected explicit focus target to allow inactive components")
	}
	if !fm.IsFocused(id1) {
		t.Fatal("expected explicit inactive target to become focused")
	}
}

// TestBlur tests blur functionality
func TestBlur(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	fm.Register(id1, true)

	if !fm.IsFocused(id1) {
		t.Error("Expected component to be focused initially")
	}

	fm.Blur()

	if fm.IsFocused(id1) {
		t.Error("Expected component to not be focused after blur")
	}

	if fm.FocusedID() != "" {
		t.Error("Expected no focused component after blur")
	}
}

func TestBlurClearsExplicitEmptyStringFocus(t *testing.T) {
	fm := focus.NewFocusManager()

	fm.Register("", true)

	if !fm.HasFocus() {
		t.Fatal("expected explicit empty string id to count as focused")
	}

	fm.Blur()

	if fm.HasFocus() {
		t.Fatal("expected blur to clear explicit empty string focus")
	}
	if fm.FocusedID() != "" {
		t.Fatalf("expected no focused id after blur, got %q", fm.FocusedID())
	}
}

func TestActiveIDPreservesExplicitEmptyStringAndNilState(t *testing.T) {
	fm := focus.NewFocusManager()

	if activeID := fm.ActiveID(); activeID != nil {
		t.Fatalf("expected nil active id without focus, got %q", *activeID)
	}

	fm.Register("", true)

	activeID := fm.ActiveID()
	if activeID == nil {
		t.Fatal("expected active id for explicit empty-string focus target")
	}
	if *activeID != "" {
		t.Fatalf("expected explicit empty-string active id, got %q", *activeID)
	}

	fm.Blur()

	if activeID := fm.ActiveID(); activeID != nil {
		t.Fatalf("expected nil active id after blur, got %q", *activeID)
	}
}

func TestRegisterDoesNotReapplyAutoFocusToExistingComponent(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	fm.Register(id1, true)

	if !fm.IsFocused(id1) {
		t.Fatal("expected first register with autoFocus to focus the component")
	}

	fm.Blur()
	fm.Register(id1, true)

	if fm.IsFocused(id1) {
		t.Fatal("expected re-registering an existing component not to reapply autoFocus")
	}
	if fm.FocusedID() != "" {
		t.Fatalf("expected no focused component after re-register blur path, got %q", fm.FocusedID())
	}
}

// TestFocusNext tests forward focus navigation
func TestFocusNext(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, false)
	fm.Register(id2, false)
	fm.Register(id3, false)

	fm.Focus(id1)

	// Move to next
	fm.FocusNext()
	if !fm.IsFocused(id2) {
		t.Error("Expected focus to move to component-2")
	}

	fm.FocusNext()
	if !fm.IsFocused(id3) {
		t.Error("Expected focus to move to component-3")
	}

	// Should wrap around
	fm.FocusNext()
	if !fm.IsFocused(id1) {
		t.Error("Expected focus to wrap around to component-1")
	}
}

func TestFocusNextWithoutCurrentFocusStartsAtFirst(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")

	fm.Register(id1, false)
	fm.Register(id2, false)

	if !fm.FocusNext() {
		t.Fatal("expected FocusNext to succeed with registered focusables")
	}
	if !fm.IsFocused(id1) {
		t.Error("Expected FocusNext to focus the first component when none is active")
	}
}

func TestFocusNextSkipsInactiveComponents(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, true)
	fm.Register(id2, false)
	fm.Register(id3, false)
	fm.Deactivate(id2)

	if !fm.FocusNext() {
		t.Fatal("expected FocusNext to succeed")
	}
	if !fm.IsFocused(id3) {
		t.Fatal("expected FocusNext to skip inactive components")
	}
}

func TestFocusNextWrapSkipsInactiveComponents(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, false)
	fm.Register(id2, false)
	fm.Register(id3, false)
	fm.Deactivate(id1)
	fm.Focus(id3)

	if !fm.FocusNext() {
		t.Fatal("expected FocusNext to succeed")
	}
	if !fm.IsFocused(id2) {
		t.Fatal("expected FocusNext to wrap to the next active component")
	}
}

func TestFocusNextAdvancesPastExplicitEmptyStringID(t *testing.T) {
	fm := focus.NewFocusManager()

	emptyID := focus.FocusID("")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(emptyID, true)
	fm.Register(id2, false)
	fm.Register(id3, false)

	if !fm.IsFocused(emptyID) {
		t.Fatal("expected explicit empty string id to be tracked as the focused target internally")
	}

	if !fm.FocusNext() {
		t.Fatal("expected FocusNext to succeed from an explicit empty string id")
	}
	if !fm.IsFocused(id2) {
		t.Fatal("expected FocusNext to move to the next component after an explicit empty string id")
	}
}

// TestFocusPrevious tests backward focus navigation
func TestFocusPrevious(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, false)
	fm.Register(id2, false)
	fm.Register(id3, false)

	fm.Focus(id3)

	// Move to previous
	fm.FocusPrevious()
	if !fm.IsFocused(id2) {
		t.Error("Expected focus to move to component-2")
	}

	fm.FocusPrevious()
	if !fm.IsFocused(id1) {
		t.Error("Expected focus to move to component-1")
	}

	// Should wrap around
	fm.FocusPrevious()
	if !fm.IsFocused(id3) {
		t.Error("Expected focus to wrap around to component-3")
	}
}

func TestFocusPreviousWithoutCurrentFocusStartsAtLastActive(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")

	fm.Register(id1, false)
	fm.Register(id2, false)

	if !fm.FocusPrevious() {
		t.Fatal("expected FocusPrevious to succeed with registered focusables")
	}
	if !fm.IsFocused(id2) {
		t.Error("Expected FocusPrevious to focus the last active component when none is active")
	}
}

func TestFocusPreviousWrapSkipsInactiveComponents(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, false)
	fm.Register(id2, false)
	fm.Register(id3, false)
	fm.Deactivate(id3)
	fm.Focus(id1)

	if !fm.FocusPrevious() {
		t.Fatal("expected FocusPrevious to succeed")
	}
	if !fm.IsFocused(id2) {
		t.Fatal("expected FocusPrevious to wrap to the previous active component")
	}
}

func TestDeactivateClearsCurrentFocusAndPreservesOrder(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")

	fm.Register(id1, true)
	fm.Register(id2, false)
	if !fm.IsFocused(id1) {
		t.Fatal("expected first component to be focused")
	}

	if !fm.Deactivate(id1) {
		t.Fatal("expected deactivate to succeed")
	}
	if fm.FocusedID() != "" {
		t.Fatal("expected deactivating the focused component to clear focus")
	}
	if len(fm.FocusOrder()) != 2 {
		t.Fatal("expected deactivated component to remain in focus order")
	}
}

func TestFocusNextAfterFocusedComponentUnregistersStartsAtFirstRemaining(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, true)
	fm.Register(id2, false)
	fm.Register(id3, false)

	fm.Unregister(id1)
	if fm.FocusedID() != "" {
		t.Fatal("expected unregistering the focused component to clear focus first")
	}

	if !fm.FocusNext() {
		t.Fatal("expected FocusNext to succeed after unregistering the focused component")
	}
	if !fm.IsFocused(id2) {
		t.Fatal("expected FocusNext after unregister to start at the first remaining component")
	}
}

func TestFocusPreviousAfterFocusedComponentUnregistersStartsAtLastRemaining(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, true)
	fm.Register(id2, false)
	fm.Register(id3, false)

	fm.Unregister(id1)
	if fm.FocusedID() != "" {
		t.Fatal("expected unregistering the focused component to clear focus first")
	}

	if !fm.FocusPrevious() {
		t.Fatal("expected FocusPrevious to succeed after unregistering the focused component")
	}
	if !fm.IsFocused(id3) {
		t.Fatal("expected FocusPrevious after unregister to start at the last remaining component")
	}
}

// TestFocusOrder tests focus order tracking
func TestFocusOrder(t *testing.T) {
	fm := focus.NewFocusManager()

	id1 := focus.FocusID("component-1")
	id2 := focus.FocusID("component-2")
	id3 := focus.FocusID("component-3")

	fm.Register(id1, false)
	fm.Register(id2, false)
	fm.Register(id3, false)

	order := fm.FocusOrder()

	if len(order) != 3 {
		t.Errorf("Expected 3 components in order, got %d", len(order))
	}

	if order[0] != id1 || order[1] != id2 || order[2] != id3 {
		t.Error("Focus order not preserved")
	}
}

// TestEmptyFocusManager tests operations on empty manager
func TestEmptyFocusManager(t *testing.T) {
	fm := focus.NewFocusManager()

	if fm.Focus("missing") {
		t.Error("Focus() should return false for empty manager")
	}

	if fm.FocusNext() {
		t.Error("FocusNext() should return false for empty manager")
	}

	if fm.FocusPrevious() {
		t.Error("FocusPrevious() should return return false for empty manager")
	}

	if fm.FocusedID() != "" {
		t.Error("Expected empty focused ID for empty manager")
	}
}

// TestComponent tests Component implementation
func TestComponent(t *testing.T) {
	c := focus.NewComponent("test-component")

	if c.ID() != "test-component" {
		t.Errorf("Expected ID 'test-component', got %q", c.ID())
	}

	if c.IsFocused() {
		t.Error("New component should not be focused")
	}

	c.SetFocus(true)

	if !c.IsFocused() {
		t.Error("Component should be focused after SetFocus(true)")
	}
}

// TestGenerateID tests ID generation
func TestGenerateID(t *testing.T) {
	id1 := focus.GenerateID("button")
	id2 := focus.GenerateID("button")

	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
}

// TestGlobal tests global focus manager
func TestGlobal(t *testing.T) {
	fm := focus.Global()

	if fm == nil {
		t.Fatal("Global focus manager should not be nil")
	}

	// Register a component
	id := focus.FocusID("global-test")
	fm.Register(id, true)

	if !fm.IsFocused(id) {
		t.Error("Component should be focused in global manager")
	}

	// Cleanup
	fm.Unregister(id)
}

// TestConcurrentAccess tests concurrent access to focus manager
func TestConcurrentAccess(t *testing.T) {
	fm := focus.NewFocusManager()

	done := make(chan bool)

	// Concurrent registration
	for i := 0; i < 10; i++ {
		go func(n int) {
			id := focus.FocusID(fmt.Sprintf("component-%d", n))
			fm.Register(id, false)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if fm.FocusableCount() != 10 {
		t.Errorf("Expected 10 components, got %d", fm.FocusableCount())
	}
}
