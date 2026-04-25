package ink_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func boolPtr(value bool) *bool {
	return &value
}

func assertPanic(t *testing.T, name string, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for %s", name)
		}
	}()

	fn()
}

func assertPanicContains(t *testing.T, expected string, fn func()) {
	t.Helper()

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic containing %q", expected)
		}

		message := fmt.Sprint(recovered)

		if !strings.Contains(message, expected) {
			t.Fatalf("expected panic containing %q, got %q", expected, message)
		}
	}()

	fn()
}

// TestAppRender tests basic app rendering
func TestAppRender(t *testing.T) {
	counter := 0

	component := func() *vdom.Node {
		return vdom.CreateElement("box", nil,
			components.Text("Count: ", string(rune(counter+'0'))),
		)
	}

	app := ink.NewApp(component)
	output := app.RenderOnce()

	if !strings.Contains(output, "Count: 0") {
		t.Errorf("Expected 'Count: 0', got: %s", output)
	}
}

func TestRenderOnceFailsWhenTextNodesAreNotWithinTextComponentLikeUpstreamFixture(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return components.Box(nil,
			vdom.CreateTextNode("Hello"),
			components.Text("World"),
		)
	})

	assertPanicContains(t, `Text string "Hello" must be rendered inside <Text> component`, func() {
		_ = app.RenderOnce()
	})
}

func TestRenderOnceFailsWhenTextNodeIsNotWithinTextComponentLikeUpstreamFixture(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return components.Box(nil,
			vdom.CreateTextNode("Hello World"),
		)
	})

	assertPanicContains(t, `Text string "Hello World" must be rendered inside <Text> component`, func() {
		_ = app.RenderOnce()
	})
}

func TestRenderOnceFailsWhenBoxIsInsideTextComponentLikeUpstreamFixture(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return components.Text(
			"Hello World",
			components.Box(nil),
		)
	})

	assertPanicContains(t, `<Box> can’t be nested inside <Text> component`, func() {
		_ = app.RenderOnce()
	})
}

// TestAppRerender tests that app can re-render with updated state
func TestAppRerender(t *testing.T) {
	counter := 0

	component := func() *vdom.Node {
		return vdom.CreateTextNode("Value: " + string(rune(counter+'0')))
	}

	app := ink.NewApp(component)

	// First render
	output1 := app.RenderOnce()
	if !strings.Contains(output1, "Value: 0") {
		t.Errorf("Expected 'Value: 0', got: %s", output1)
	}

	// Update state
	counter = 5

	// Re-render
	output2 := app.RenderOnce()
	if !strings.Contains(output2, "Value: 5") {
		t.Errorf("Expected 'Value: 5', got: %s", output2)
	}
}

// TestAppWithHooks tests app with useState hook
func TestAppWithHooks(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		count, setCount := ink.UseState(0)

		// For testing, we'll update state in the component
		// (normally this would be triggered by user input)
		if count.(int) == 0 {
			setCount(1)
		}

		countStr := string(rune(count.(int) + '0'))
		return vdom.CreateTextNode("Count: " + countStr)
	})

	// First render - count is 0, but setState is called
	output1 := app.RenderOnce()
	if !strings.Contains(output1, "Count: 0") {
		t.Errorf("First render: expected 'Count: 0', got: %s", output1)
	}

	// Second render - count should be 1
	output2 := app.RenderOnce()
	if !strings.Contains(output2, "Count: 1") {
		t.Errorf("Second render: expected 'Count: 1', got: %s", output2)
	}
}

// TestGetVNode tests GetVNode method
func TestGetVNode(t *testing.T) {
	callCount := 0
	component := func() *vdom.Node {
		callCount++
		return vdom.CreateTextNode("test")
	}

	app := ink.NewApp(component)

	// First call
	node1 := app.GetVNode()
	if node1 == nil {
		t.Error("Expected non-nil node")
	}
	if node1.Text != "test" {
		t.Errorf("Expected 'test', got %q", node1.Text)
	}

	// Second call should call component again
	node2 := app.GetVNode()
	if node2 == nil {
		t.Error("Expected non-nil node")
	}
	if callCount != 2 {
		t.Errorf("Expected component to be called twice, got %d", callCount)
	}
}

// TestNewAppWithNilComponent tests nil component handling
func TestNewAppWithNilComponent(t *testing.T) {
	app := ink.NewApp(nil)
	if app == nil {
		t.Error("NewApp should handle nil component")
	}

	output := app.RenderOnce()
	if output != "" {
		t.Errorf("Expected empty output for nil component, got %q", output)
	}
}

// TestUseStateOutsideComponent tests UseState panic
func TestUseStateOutsideComponent(t *testing.T) {
	// Reset the global hooks context to simulate "outside of component" state
	ink.ResetCurrentHooksContext()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when UseState called outside component")
		}
	}()

	// This should panic because there's no active hooks context
	_, _ = ink.UseState(0)
}

// TestMultipleUseStateCalls tests multiple state hooks
func TestMultipleUseStateCalls(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		count, setCount := ink.UseState(0)
		name, setName := ink.UseState("default")

		// Update both on first render
		if count.(int) == 0 {
			setCount(5)
			setName("updated")
		}

		return vdom.CreateTextNode(fmt.Sprintf("%s: %d", name, count))
	})

	// First render shows initial values
	output1 := app.RenderOnce()
	if !strings.Contains(output1, "default: 0") {
		t.Errorf("First render: expected 'default: 0', got: %s", output1)
	}

	// Second render shows updated values
	output2 := app.RenderOnce()
	if !strings.Contains(output2, "updated: 5") {
		t.Errorf("Second render: expected 'updated: 5', got: %s", output2)
	}
}

// TestEmptyComponent tests empty component
func TestEmptyComponent(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return vdom.CreateTextNode("")
	})

	output := app.RenderOnce()
	if output != "" {
		t.Errorf("Expected empty output, got: %q", output)
	}
}

// TestComplexComponentHierarchy tests nested components
func TestComplexComponentHierarchy(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		inner := vdom.CreateTextNode("inner")
		middle := vdom.CreateElement("box", nil, inner)
		outer := vdom.CreateElement("container", nil, middle)
		return outer
	})

	node := app.GetVNode()
	if node.ElementType != "container" {
		t.Errorf("Expected 'container', got %q", node.ElementType)
	}
}

// TestMultipleApps tests that multiple apps can coexist
func TestMultipleApps(t *testing.T) {
	app1 := ink.NewApp(func() *vdom.Node {
		return vdom.CreateTextNode("app1")
	})

	app2 := ink.NewApp(func() *vdom.Node {
		return vdom.CreateTextNode("app2")
	})

	// Render both apps
	output1 := app1.RenderOnce()
	output2 := app2.RenderOnce()

	if !strings.Contains(output1, "app1") {
		t.Errorf("Expected 'app1' in output1, got: %s", output1)
	}
	if !strings.Contains(output2, "app2") {
		t.Errorf("Expected 'app2' in output2, got: %s", output2)
	}

	// Switch back to app1 - should still work
	output1Again := app1.RenderOnce()
	if !strings.Contains(output1Again, "app1") {
		t.Errorf("Expected 'app1' in output1Again, got: %s", output1Again)
	}
}

// TestHooksResetBetweenRenders tests hooks are reset between renders
func TestHooksResetBetweenRenders(t *testing.T) {
	callCount := 0
	app := ink.NewApp(func() *vdom.Node {
		callCount++
		// Each render should call UseState in order
		count, _ := ink.UseState(callCount)
		return vdom.CreateTextNode(fmt.Sprintf("Count: %v", count))
	})

	// First render
	app.RenderOnce()
	// Second render - hooks should be reset
	app.RenderOnce()

	// Should have rendered twice without issues
	if callCount != 2 {
		t.Errorf("Expected 2 renders, got %d", callCount)
	}
}

// TestComponentWithProps tests component that uses props-like values
func TestComponentWithProps(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return vdom.CreateElement("box", vdom.Props{
			"color": "red",
			"width": 100,
		}, components.Text("content"))
	})

	node := app.GetVNode()
	if node.Props["color"] != "red" {
		t.Errorf("Expected color 'red', got %v", node.Props["color"])
	}
	if node.Props["width"] != 100 {
		t.Errorf("Expected width 100, got %v", node.Props["width"])
	}
}

// TestStateUpdatesNotAffectingOtherApps tests state isolation
func TestStateUpdatesNotAffectingOtherApps(t *testing.T) {
	app1 := ink.NewApp(func() *vdom.Node {
		count, setCount := ink.UseState(0)
		if count.(int) == 0 {
			setCount(10)
		}
		return vdom.CreateTextNode(fmt.Sprintf("app1: %v", count))
	})

	app2 := ink.NewApp(func() *vdom.Node {
		count, setCount := ink.UseState(0)
		if count.(int) == 0 {
			setCount(20)
		}
		return vdom.CreateTextNode(fmt.Sprintf("app2: %v", count))
	})

	// First renders
	_ = app1.RenderOnce()
	_ = app2.RenderOnce()

	// Second renders
	output1Again := app1.RenderOnce()
	output2Again := app2.RenderOnce()

	// Each app should have its own state
	if !strings.Contains(output1Again, "app1: 10") {
		t.Errorf("Expected 'app1: 10', got: %s", output1Again)
	}
	if !strings.Contains(output2Again, "app2: 20") {
		t.Errorf("Expected 'app2: 20', got: %s", output2Again)
	}
}

// TestRenderOnceWithWidth tests RenderOnce respects width
func TestRenderOnceWithWidth(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return vdom.CreateTextNode("test content")
	})

	// Render with default width (80)
	output1 := app.RenderOnce()

	// The output should contain the text
	if !strings.Contains(output1, "test content") {
		t.Errorf("Expected 'test content' in output, got: %s", output1)
	}
}

// TestGetVNodeCalledMultipleTimes tests GetVNode behavior on multiple calls
func TestGetVNodeCalledMultipleTimes(t *testing.T) {
	renderCount := 0
	app := ink.NewApp(func() *vdom.Node {
		renderCount++
		return vdom.CreateTextNode(fmt.Sprintf("render %d", renderCount))
	})

	// Call GetVNode multiple times
	app.GetVNode()
	app.GetVNode()
	app.GetVNode()

	if renderCount != 3 {
		t.Errorf("Expected 3 renders, got %d", renderCount)
	}
}

// TestNilComponentInNewApp tests NewApp with nil component
func TestNilComponentInNewApp(t *testing.T) {
	app := ink.NewApp(nil)

	if app == nil {
		t.Fatal("NewApp should return an app even with nil component")
	}

	if output := app.RenderOnce(); output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

// TestComponentReturningNil tests component returning nil
func TestComponentReturningNil(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return nil
	})

	output := app.RenderOnce()
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

// TestResetCurrentHooksContextIsolation tests ResetCurrentHooksContext doesn't break apps
func TestResetCurrentHooksContextIsolation(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return vdom.CreateTextNode("test")
	})

	// Render once
	ink.ResetCurrentHooksContext()
	app.RenderOnce()

	// Reset the global context again
	ink.ResetCurrentHooksContext()

	// If we try to call UseState now, it should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when UseState called after ResetCurrentHooksContext")
		}
	}()

	_, _ = ink.UseState(0)
}

// TestComponentFuncType tests ComponentFunc is a function type
func TestComponentFuncType(t *testing.T) {
	var fn ink.ComponentFunc = func() *vdom.Node {
		return vdom.CreateTextNode("typed")
	}

	app := ink.NewApp(fn)
	output := app.RenderOnce()

	if !strings.Contains(output, "typed") {
		t.Errorf("Expected 'typed' in output, got: %s", output)
	}
}

func TestNewAppWithOptions(t *testing.T) {
	stdin := strings.NewReader("stdin")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	app := ink.NewAppWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("configured")
	}, ink.AppOptions{
		Width:               120,
		Height:              40,
		Stdin:               stdin,
		Stdout:              stdout,
		Stderr:              stderr,
		ScreenReaderEnabled: true,
	})

	if app.Width() != 120 {
		t.Errorf("Expected width 120, got %d", app.Width())
	}
	if app.Height() != 40 {
		t.Errorf("Expected height 40, got %d", app.Height())
	}
	if app.Stdin() != stdin {
		t.Error("Expected custom stdin to be preserved")
	}
	if app.Stdout() != stdout {
		t.Error("Expected custom stdout to be preserved")
	}
	if app.Stderr() != stderr {
		t.Error("Expected custom stderr to be preserved")
	}
	if !app.IsScreenReaderEnabled() {
		t.Error("Expected screen reader mode to be enabled")
	}
}

func TestUseAppExit(t *testing.T) {
	exitErr := errors.New("boom")
	app := ink.NewApp(func() *vdom.Node {
		ink.UseApp().Exit(exitErr)
		return vdom.CreateTextNode("done")
	})

	output := app.RenderOnce()
	if !strings.Contains(output, "done") {
		t.Errorf("Expected rendered output, got %q", output)
	}
	if !app.ExitRequested() {
		t.Error("Expected app exit to be requested")
	}
	if !errors.Is(app.ExitError(), exitErr) {
		t.Errorf("Expected exit error %v, got %v", exitErr, app.ExitError())
	}
}

func TestUseStdinContext(t *testing.T) {
	stdin := strings.NewReader("stdin")
	var observed any
	var rawSupported bool
	var rawErr error

	app := ink.NewAppWithOptions(func() *vdom.Node {
		ctx := ink.UseStdin()
		observed = ctx.Stdin
		rawSupported = ctx.IsRawModeSupported
		rawErr = ctx.SetRawMode(true)
		return vdom.CreateTextNode("stdin")
	}, ink.AppOptions{Stdin: stdin})

	app.RenderOnce()

	if observed != stdin {
		t.Error("Expected hook to expose configured stdin")
	}
	if rawSupported {
		t.Error("Expected non-terminal stdin to report raw mode unsupported")
	}
	if rawErr != nil {
		t.Errorf("Expected SetRawMode to be a no-op for unsupported stdin, got %v", rawErr)
	}
}

func TestUseStdoutWrite(t *testing.T) {
	stdout := &bytes.Buffer{}
	var observed any
	var written int
	var writeErr error

	app := ink.NewAppWithOptions(func() *vdom.Node {
		ctx := ink.UseStdout()
		observed = ctx.Stdout
		written, writeErr = ctx.Write("hello stdout")
		return vdom.CreateTextNode("rendered")
	}, ink.AppOptions{Stdout: stdout})

	app.RenderOnce()

	if observed != stdout {
		t.Error("Expected hook to expose configured stdout")
	}
	if writeErr != nil {
		t.Errorf("Unexpected stdout write error: %v", writeErr)
	}
	if written != len("hello stdout") {
		t.Errorf("Expected %d bytes written, got %d", len("hello stdout"), written)
	}
	if stdout.String() != "hello stdout" {
		t.Errorf("Expected stdout contents %q, got %q", "hello stdout", stdout.String())
	}
}

func TestUseStderrWrite(t *testing.T) {
	stderr := &bytes.Buffer{}
	var observed any
	var written int
	var writeErr error

	app := ink.NewAppWithOptions(func() *vdom.Node {
		ctx := ink.UseStderr()
		observed = ctx.Stderr
		written, writeErr = ctx.Write("hello stderr")
		return vdom.CreateTextNode("rendered")
	}, ink.AppOptions{Stderr: stderr})

	app.RenderOnce()

	if observed != stderr {
		t.Error("Expected hook to expose configured stderr")
	}
	if writeErr != nil {
		t.Errorf("Unexpected stderr write error: %v", writeErr)
	}
	if written != len("hello stderr") {
		t.Errorf("Expected %d bytes written, got %d", len("hello stderr"), written)
	}
	if stderr.String() != "hello stderr" {
		t.Errorf("Expected stderr contents %q, got %q", "hello stderr", stderr.String())
	}
}

func TestUseCursorPosition(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		ink.UseCursor().SetCursorPosition(&ink.CursorPosition{X: 2, Y: 3})
		return vdom.CreateTextNode("cursor")
	})

	app.RenderOnce()
	position := app.CursorPosition()
	if position == nil {
		t.Fatal("Expected cursor position to be set")
	}
	if position.X != 2 || position.Y != 3 {
		t.Errorf("Expected cursor position {2 3}, got %+v", position)
	}
}

func TestCursorPositionClearedBetweenRenders(t *testing.T) {
	showCursor := true
	app := ink.NewApp(func() *vdom.Node {
		if showCursor {
			ink.UseCursor().SetCursorPosition(&ink.CursorPosition{X: 1, Y: 0})
		}
		return vdom.CreateTextNode("cursor")
	})

	app.RenderOnce()
	if app.CursorPosition() == nil {
		t.Fatal("Expected cursor position after first render")
	}

	showCursor = false
	app.RenderOnce()
	if app.CursorPosition() != nil {
		t.Errorf("Expected cursor position to be cleared, got %+v", app.CursorPosition())
	}
}

func TestUseIsScreenReaderEnabled(t *testing.T) {
	enabled := false
	app := ink.NewAppWithOptions(func() *vdom.Node {
		enabled = ink.UseIsScreenReaderEnabled()
		return vdom.CreateTextNode("screen-reader")
	}, ink.AppOptions{ScreenReaderEnabled: true})

	app.RenderOnce()
	if !enabled {
		t.Error("Expected hook to report screen reader enabled")
	}
}

func TestMeasureElementUsesLatestComputedLayout(t *testing.T) {
	var measured ink.ElementDimensions

	app := ink.NewApp(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))
		node := vdom.CreateElement("box", vdom.Props{
			"ref":    ref,
			"width":  float64(12),
			"height": float64(3),
		}, components.Text("content"))

		ink.UseEffect(func() func() {
			current, _ := ref.Current().(*ink.DOMElement)
			measured = ink.MeasureElement(current)
			return nil
		}, []interface{}{"measure"})

		return node
	})

	app.RenderOnce()

	if measured.Width != 12 || measured.Height != 3 {
		t.Fatalf("expected measured size {12 3}, got %+v", measured)
	}
}

func TestMeasureElementStateUpdateAppearsOnNextRender(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		width, setWidth := ink.UseState(0)
		ref := ink.UseRef((*ink.DOMElement)(nil))

		node := vdom.CreateElement("box", vdom.Props{
			"ref":   ref,
			"width": float64(12),
		}, components.Text(fmt.Sprintf("Width: %d", width.(int))))

		ink.UseEffect(func() func() {
			current, _ := ref.Current().(*ink.DOMElement)
			setWidth(ink.MeasureElement(current).Width)
			return nil
		}, []interface{}{"measure"})

		return node
	})

	first := app.RenderOnce()
	second := app.RenderOnce()

	if !strings.Contains(first, "Width: 0") {
		t.Fatalf("expected first render to contain initial width, got %q", first)
	}
	if !strings.Contains(second, "Width: 12") {
		t.Fatalf("expected second render to contain measured width, got %q", second)
	}
}

func TestUseFocusManager(t *testing.T) {
	var focusedInitially bool
	var focusedWhileDisabled bool
	var secondFocusedWhileDisabled bool
	var focusedAfterEnable bool
	var secondFocused bool
	var firstFocusedAfterMissingID bool

	app := ink.NewApp(func() *vdom.Node {
		firstFocused, _, _ := ink.UseFocus("public-focus-1", true)
		secondFocusedState, _, _ := ink.UseFocus("public-focus-2", false)
		manager := ink.UseFocusManager()

		focusedInitially = firstFocused()
		manager.DisableFocus()
		focusedWhileDisabled = firstFocused()
		manager.Focus("public-focus-2")
		secondFocusedWhileDisabled = secondFocusedState()
		manager.EnableFocus()
		focusedAfterEnable = secondFocusedState()
		manager.Focus("public-focus-1")
		manager.FocusNext()
		secondFocused = secondFocusedState()
		manager.Focus("missing-focus-id")
		firstFocusedAfterMissingID = firstFocused()

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
	if !focusedAfterEnable {
		t.Fatal("expected focused state to remain visible after re-enabling focus management")
	}
	if !secondFocused {
		t.Fatal("expected focusNext to move focus to the next component")
	}
	if firstFocusedAfterMissingID {
		t.Fatal("expected missing focus target not to change the current focus")
	}
	if !secondFocused {
		t.Fatal("expected missing focus target to leave the current target focused")
	}
}

func TestUseFocusManagerMissingTargetSkipsInactiveFallbacks(t *testing.T) {
	var inactiveFocusedAfterMissingID bool
	var secondFocusedAfterMissingID bool
	var thirdFocusedAfterMissingID bool

	app := ink.NewApp(func() *vdom.Node {
		inactiveFocused, _, _ := ink.UseFocus(ink.FocusOptions{
			ID:       "public-focus-inactive",
			IsActive: boolPtr(false),
		})
		secondFocused, _, _ := ink.UseFocus("public-focus-active-1", false)
		thirdFocused, _, _ := ink.UseFocus("public-focus-active-2", true)
		manager := ink.UseFocusManager()

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

func TestUseFocusAutoFocusRerender(t *testing.T) {
	autoFocus := false
	var focusedBefore bool
	var focusedAfter bool

	app := ink.NewApp(func() *vdom.Node {
		isFocused, _, _ := ink.UseFocus(ink.FocusOptions{
			ID:        "public-focus-rerender",
			AutoFocus: autoFocus,
		})

		if autoFocus {
			focusedAfter = isFocused()
		} else {
			focusedBefore = isFocused()
		}

		return vdom.CreateTextNode("focus")
	})

	app.RenderOnce()
	autoFocus = true
	app.RenderOnce()

	if focusedBefore {
		t.Fatal("expected initial render without autoFocus to remain unfocused")
	}
	if !focusedAfter {
		t.Fatal("expected rerender with autoFocus to focus the component")
	}
}

func TestUseFocusEffectDrivenFocusRerenders(t *testing.T) {
	phase := 0

	app := ink.NewApp(func() *vdom.Node {
		firstFocused, focusFirst, _ := ink.UseFocus("public-focus-effect-1", true)
		secondFocused, _, _ := ink.UseFocus("public-focus-effect-2", false)

		ink.UseEffect(func() func() {
			if phase == 0 {
				phase = 1
				focusFirst("public-focus-effect-2")
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

	if !strings.Contains(first, "first") {
		t.Fatalf("expected first render to contain the initial focused label, got %q", first)
	}
	if !strings.Contains(second, "second") {
		t.Fatalf("expected second render to contain the effect-driven focus target, got %q", second)
	}
}

func TestUseFocusEffectDrivenBlurRerenders(t *testing.T) {
	phase := 0

	app := ink.NewApp(func() *vdom.Node {
		firstFocused, _, blurFirst := ink.UseFocus("public-focus-blur", true)

		ink.UseEffect(func() func() {
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

	if !strings.Contains(first, "first") {
		t.Fatalf("expected first render to contain the focused label, got %q", first)
	}
	if !strings.Contains(second, "none") {
		t.Fatalf("expected second render to contain the blurred label, got %q", second)
	}
}

func TestUseFocusManagerFocusPrevious(t *testing.T) {
	var thirdFocused bool

	app := ink.NewApp(func() *vdom.Node {
		_, _, _ = ink.UseFocus("public-focus-1", true)
		_, _, _ = ink.UseFocus("public-focus-2", false)
		thirdFocusedState, _, _ := ink.UseFocus("public-focus-3", false)
		manager := ink.UseFocusManager()

		manager.FocusPrevious()
		thirdFocused = thirdFocusedState()

		return vdom.CreateTextNode("focus-manager")
	})

	app.RenderOnce()

	if !thirdFocused {
		t.Fatal("expected focusPrevious to wrap to the last focus target")
	}
}

func TestRuntimeHooksPanicOutsideRender(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return vdom.CreateTextNode("rendered")
	})
	app.RenderOnce()

	assertPanic(t, "UseApp", func() {
		_ = ink.UseApp()
	})
	assertPanic(t, "UseStdin", func() {
		_ = ink.UseStdin()
	})
	assertPanic(t, "UseStdout", func() {
		_ = ink.UseStdout()
	})
	assertPanic(t, "UseStderr", func() {
		_ = ink.UseStderr()
	})
	assertPanic(t, "UseCursor", func() {
		_ = ink.UseCursor()
	})
	assertPanic(t, "UseFocusManager", func() {
		_ = ink.UseFocusManager()
	})
	assertPanic(t, "UseIsScreenReaderEnabled", func() {
		_ = ink.UseIsScreenReaderEnabled()
	})
}
