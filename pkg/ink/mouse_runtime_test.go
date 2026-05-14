package ink

import (
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/hooks"
	"github.com/dh-kam/ink-go/pkg/input"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func x10MouseFrame(cb, cx, cy int) string {
	return "\x1b[M" + string([]byte{byte(cb + 32), byte(cx + 32), byte(cy + 32)})
}

func TestRouteMouseInputDispatches(t *testing.T) {
	hooks.ResetMouseHooks()
	var got input.MouseEvent
	defer hooks.UseMouse(func(ev input.MouseEvent) { got = ev })()

	if !routeMouseInput("\x1b[<0;5;7M") {
		t.Fatal("expected SGR sequence to be consumed")
	}
	if got.X != 5 || got.Y != 7 || got.Button != input.MouseLeft || got.Action != input.MouseActionPress {
		t.Fatalf("got = %+v, want left/press at (5,7)", got)
	}
}

func TestRouteMouseInputDispatchesX10(t *testing.T) {
	hooks.ResetMouseHooks()
	var got input.MouseEvent
	defer hooks.UseMouse(func(ev input.MouseEvent) { got = ev })()

	if !routeMouseInput(x10MouseFrame(2, 9, 11)) {
		t.Fatal("expected X10 sequence to be consumed")
	}
	if got.X != 9 || got.Y != 11 || got.Button != input.MouseRight || got.Action != input.MouseActionPress {
		t.Fatalf("got = %+v, want right/press at (9,11)", got)
	}
}

func TestRouteMouseInputDoesNotConsumeWithoutSubscriber(t *testing.T) {
	hooks.ResetMouseHooks()

	if routeMouseInput("\x1b[<0;5;7M") {
		t.Fatal("expected SGR sequence without subscribers not to be consumed")
	}
	if routeMouseInput(x10MouseFrame(0, 5, 7)) {
		t.Fatal("expected X10 sequence without subscribers not to be consumed")
	}
}

func TestRouteMouseInputIgnoresNonMouse(t *testing.T) {
	hooks.ResetMouseHooks()
	called := false
	defer hooks.UseMouse(func(input.MouseEvent) { called = true })()

	if routeMouseInput("a") {
		t.Fatal("plain key should not be routed as mouse")
	}
	if routeMouseInput("\x1b[A") {
		t.Fatal("arrow key should not be routed as mouse")
	}
	if called {
		t.Fatal("mouse handler should not fire for key input")
	}
}

func TestRouteMouseInputMalformedReturnsFalse(t *testing.T) {
	hooks.ResetMouseHooks()
	called := false
	defer hooks.UseMouse(func(input.MouseEvent) { called = true })()

	// Looks like SGR (prefix + terminator) but body is malformed.
	if routeMouseInput("\x1b[<abc;1;1M") {
		t.Fatal("malformed SGR body should not consume the input")
	}
	if called {
		t.Fatal("malformed SGR must not dispatch")
	}
}

func TestMouseRuntimePassesUnhandledSGRBytesToUseInput(t *testing.T) {
	hooks.ResetMouseHooks()
	stdout := &recordingWriter{}
	var received string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			received = input
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b[<0;5;7M"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if received != "[<0;5;7M" {
		t.Fatalf("expected unhandled SGR bytes to reach useInput, got %q", received)
	}
}

func TestMouseRuntimeRoutesX10ToMountedApp(t *testing.T) {
	hooks.ResetMouseHooks()
	stdout := &recordingWriter{}
	var received input.MouseEvent

	instance, err := MountWithOptions(func() *vdom.Node {
		UseMouse(func(ev MouseEvent) {
			received = ev
		})

		return components.Text("Mouse")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput(x10MouseFrame(64, 3, 4)); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if received.Button != input.MouseWheelUp || received.Action != input.MouseActionWheel || received.X != 3 || received.Y != 4 {
		t.Fatalf("expected X10 wheel-up at (3,4), got %+v", received)
	}
}

// TestMouseRuntimeRoutesConcatenatedX10Frames verifies the input loop
// peels multiple X10 frames out of a single TTY chunk. Bursty mouse-move
// reports routinely arrive coalesced; the legacy parser previously only
// matched chunks whose total length was exactly 6 bytes, dropping every
// frame after the first when the kernel batched them.
func TestMouseRuntimeRoutesConcatenatedX10Frames(t *testing.T) {
	hooks.ResetMouseHooks()
	stdout := &recordingWriter{}
	var received []input.MouseEvent

	instance, err := MountWithOptions(func() *vdom.Node {
		UseMouse(func(ev MouseEvent) {
			received = append(received, ev)
		})

		return components.Text("Mouse")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	chunk := x10MouseFrame(0, 1, 1) + x10MouseFrame(2, 4, 5)
	if err := instance.HandleInput(chunk); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 dispatched X10 frames, got %d (events=%+v)", len(received), received)
	}
	if received[0].Button != input.MouseLeft || received[0].X != 1 || received[0].Y != 1 {
		t.Fatalf("first frame mismatch: %+v", received[0])
	}
	if received[1].Button != input.MouseRight || received[1].X != 4 || received[1].Y != 5 {
		t.Fatalf("second frame mismatch: %+v", received[1])
	}
}

// TestMouseRuntimeRoutesMixedSGRAndX10Frames covers the case where a
// terminal mid-stream toggle leaves an SGR frame and an X10 frame stitched
// together in the same chunk. Both should peel cleanly off the front and
// dispatch in order.
func TestMouseRuntimeRoutesMixedSGRAndX10Frames(t *testing.T) {
	hooks.ResetMouseHooks()
	stdout := &recordingWriter{}
	var received []input.MouseEvent

	instance, err := MountWithOptions(func() *vdom.Node {
		UseMouse(func(ev MouseEvent) {
			received = append(received, ev)
		})

		return components.Text("Mouse")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	chunk := "\x1b[<0;5;7M" + x10MouseFrame(2, 9, 11)
	if err := instance.HandleInput(chunk); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 dispatched frames, got %d (events=%+v)", len(received), received)
	}
	if received[0].Button != input.MouseLeft || received[0].X != 5 || received[0].Y != 7 {
		t.Fatalf("first (SGR) frame mismatch: %+v", received[0])
	}
	if received[1].Button != input.MouseRight || received[1].X != 9 || received[1].Y != 11 {
		t.Fatalf("second (X10) frame mismatch: %+v", received[1])
	}
}

// TestMouseRuntimeFlowsKeyTailAfterX10Frame verifies that a chunk
// containing an X10 mouse frame followed by plain key bytes dispatches
// the mouse event AND surfaces the trailing key bytes through useInput.
// Without leftover-routing, a stitched chunk would either drop the key
// (false-positive mouse handling) or drop the mouse (false-negative).
func TestMouseRuntimeFlowsKeyTailAfterX10Frame(t *testing.T) {
	hooks.ResetMouseHooks()
	stdout := &recordingWriter{}
	var mouseEvents []input.MouseEvent
	var keyInput string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseMouse(func(ev MouseEvent) {
			mouseEvents = append(mouseEvents, ev)
		})
		UseInput(func(input string, key InputKey) {
			keyInput = input
		})

		return components.Text("Mouse")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	chunk := x10MouseFrame(0, 2, 3) + "x"
	if err := instance.HandleInput(chunk); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if len(mouseEvents) != 1 {
		t.Fatalf("expected 1 mouse event, got %d", len(mouseEvents))
	}
	if mouseEvents[0].X != 2 || mouseEvents[0].Y != 3 {
		t.Fatalf("expected mouse event at (2,3), got %+v", mouseEvents[0])
	}
	if keyInput != "x" {
		t.Fatalf("expected trailing 'x' to reach useInput, got %q", keyInput)
	}
}

func TestMouseRuntimeScopesDispatchToMountedApp(t *testing.T) {
	hooks.ResetMouseHooks()
	stdoutA := &recordingWriter{}
	stdoutB := &recordingWriter{}
	receivedA := 0
	receivedB := 0

	instanceA, err := MountWithOptions(func() *vdom.Node {
		UseMouse(func(MouseEvent) {
			receivedA++
		})

		return components.Text("A")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdoutA},
	})
	if err != nil {
		t.Fatalf("mount A failed: %v", err)
	}
	defer instanceA.Unmount()

	instanceB, err := MountWithOptions(func() *vdom.Node {
		UseMouse(func(MouseEvent) {
			receivedB++
		})

		return components.Text("B")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdoutB},
	})
	if err != nil {
		t.Fatalf("mount B failed: %v", err)
	}
	defer instanceB.Unmount()

	if err := instanceA.HandleInput("\x1b[<0;5;7M"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if receivedA != 1 || receivedB != 0 {
		t.Fatalf("expected mouse input to dispatch only to instance A, got A=%d B=%d", receivedA, receivedB)
	}
}
