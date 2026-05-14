package renderer_test

import (
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/dh-kam/ink-go/pkg/renderer"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func TestStdinNilWhenNotConfigured(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"))
	defer inst.Cleanup()
	if inst.Stdin() != nil {
		t.Fatal("Stdin() should be nil without WithStdin option")
	}
	if dereg := inst.SubscribeInput(func([]byte) {}); dereg == nil {
		t.Fatal("SubscribeInput should return a no-op (non-nil) function")
	}
}

func TestStdinWriteFansOutToSubscribers(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdin())
	defer inst.Cleanup()

	var mu sync.Mutex
	var got []string
	dereg := inst.SubscribeInput(func(b []byte) {
		mu.Lock()
		got = append(got, string(b))
		mu.Unlock()
	})
	defer dereg()

	if _, err := inst.Stdin().Write([]byte("hello")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := inst.Stdin().Write([]byte("world")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 2 || got[0] != "hello" || got[1] != "world" {
		t.Fatalf("got = %v, want [hello world]", got)
	}
}

func TestStdinUnsubscribeStopsDelivery(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdin())
	defer inst.Cleanup()

	var count int
	dereg := inst.SubscribeInput(func([]byte) { count++ })
	dereg()
	if _, err := inst.Stdin().Write([]byte("ignored")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d after unsubscribe, want 0", count)
	}
}

func TestStdinDoubleUnsubscribeNoPanic(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdin())
	defer inst.Cleanup()
	dereg := inst.SubscribeInput(func([]byte) {})
	dereg()
	dereg() // must not panic
}

func TestStdinMultipleSubscribersIndependent(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdin())
	defer inst.Cleanup()

	var a, b int
	defer inst.SubscribeInput(func([]byte) { a++ })()
	defer inst.SubscribeInput(func([]byte) { b++ })()

	for i := 0; i < 3; i++ {
		_, _ = inst.Stdin().Write([]byte{'a'})
	}
	if a != 3 || b != 3 {
		t.Fatalf("a=%d b=%d, want 3,3", a, b)
	}
}

func TestStdinDefensiveCopyIsolatesSubscribers(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdin())
	defer inst.Cleanup()

	var seen []byte
	defer inst.SubscribeInput(func(b []byte) { seen = b })()

	buf := []byte("hello")
	_, _ = inst.Stdin().Write(buf)
	buf[0] = 'X' // mutate caller's slice — must not affect subscriber's copy
	if string(seen) != "hello" {
		t.Fatalf("subscriber saw mutation: %q", string(seen))
	}
}

func TestStdinClosedAfterCleanup(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdin())
	w := inst.Stdin()
	inst.Cleanup()
	_, err := w.Write([]byte("after"))
	if err == nil {
		t.Fatal("expected error writing to closed stdin")
	}
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("err = %v, want io.ErrClosedPipe", err)
	}
}

func TestStdinSubscribeNilFnIsNoop(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdin())
	defer inst.Cleanup()
	if dereg := inst.SubscribeInput(nil); dereg == nil {
		t.Fatal("nil fn should still return a callable dereg")
	}
}
