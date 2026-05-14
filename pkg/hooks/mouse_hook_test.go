package hooks_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dh-kam/ink-go/pkg/hooks"
	"github.com/dh-kam/ink-go/pkg/input"
)

func TestMouseHookRegisterDispatchUnregister(t *testing.T) {
	hooks.ResetMouseHooks()
	var got input.MouseEvent
	dereg := hooks.UseMouse(func(ev input.MouseEvent) {
		got = ev
	})
	if hooks.MouseHookCount() != 1 {
		t.Fatalf("MouseHookCount = %d, want 1", hooks.MouseHookCount())
	}
	want := input.MouseEvent{X: 1, Y: 2, Button: input.MouseLeft, Action: input.MouseActionPress}
	if !hooks.DispatchMouse(want) {
		t.Fatal("DispatchMouse returned false with one subscriber")
	}
	if got != want {
		t.Fatalf("got = %+v, want %+v", got, want)
	}
	dereg()
	if hooks.MouseHookCount() != 0 {
		t.Fatalf("MouseHookCount after dereg = %d, want 0", hooks.MouseHookCount())
	}
	if hooks.DispatchMouse(want) {
		t.Fatal("DispatchMouse returned true with no subscribers")
	}
}

func TestMouseHookNilCallbackNoOp(t *testing.T) {
	hooks.ResetMouseHooks()
	dereg := hooks.UseMouse(nil)
	if hooks.MouseHookCount() != 0 {
		t.Fatalf("nil callback should not register; count = %d", hooks.MouseHookCount())
	}
	dereg() // must not panic
}

func TestMouseHookDoubleDeregister(t *testing.T) {
	hooks.ResetMouseHooks()
	dereg := hooks.UseMouse(func(input.MouseEvent) {})
	dereg()
	dereg() // idempotent
	if hooks.MouseHookCount() != 0 {
		t.Fatalf("count = %d, want 0", hooks.MouseHookCount())
	}
}

func TestMouseHookMultipleSubscribers(t *testing.T) {
	hooks.ResetMouseHooks()
	var a, b int32
	dA := hooks.UseMouse(func(input.MouseEvent) { atomic.AddInt32(&a, 1) })
	dB := hooks.UseMouse(func(input.MouseEvent) { atomic.AddInt32(&b, 1) })
	defer dA()
	defer dB()

	hooks.DispatchMouse(input.MouseEvent{})
	hooks.DispatchMouse(input.MouseEvent{})
	if atomic.LoadInt32(&a) != 2 || atomic.LoadInt32(&b) != 2 {
		t.Fatalf("counts a=%d b=%d, want 2,2", a, b)
	}
}

func TestMouseHookConcurrent(t *testing.T) {
	hooks.ResetMouseHooks()
	var wg sync.WaitGroup
	var fires int64

	const subscribers = 20
	const dispatchPerG = 50

	deregs := make([]func(), subscribers)
	for i := 0; i < subscribers; i++ {
		deregs[i] = hooks.UseMouse(func(input.MouseEvent) {
			atomic.AddInt64(&fires, 1)
		})
	}

	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < dispatchPerG; i++ {
				hooks.DispatchMouse(input.MouseEvent{})
			}
		}()
	}
	wg.Wait()

	for _, d := range deregs {
		d()
	}

	want := int64(subscribers * 8 * dispatchPerG)
	if got := atomic.LoadInt64(&fires); got != want {
		t.Fatalf("fires = %d, want %d", got, want)
	}
}

func TestResetMouseHooksClears(t *testing.T) {
	hooks.UseMouse(func(input.MouseEvent) {})
	hooks.UseMouse(func(input.MouseEvent) {})
	hooks.ResetMouseHooks()
	if hooks.MouseHookCount() != 0 {
		t.Fatalf("Reset did not clear; count = %d", hooks.MouseHookCount())
	}
}
