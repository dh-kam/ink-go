package reconciler_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dh-kam/goink.go/pkg/reconciler"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func TestTrackerFirstRenderAlwaysExecutes(t *testing.T) {
	var calls int32
	tk := reconciler.NewTracker(func(n *vdom.Node) string {
		atomic.AddInt32(&calls, 1)
		return n.Text
	})
	out, fresh := tk.Render(vdom.CreateTextNode("hi"))
	if !fresh {
		t.Error("first render must report fresh=true")
	}
	if out != "hi" {
		t.Errorf("output = %q, want hi", out)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestTrackerSkipsRenderForIdenticalTrees(t *testing.T) {
	var calls int32
	tk := reconciler.NewTracker(func(n *vdom.Node) string {
		atomic.AddInt32(&calls, 1)
		return n.Text
	})
	tk.Render(vdom.CreateTextNode("hi"))
	out, fresh := tk.Render(vdom.CreateTextNode("hi"))
	if fresh {
		t.Error("identical tree should skip render")
	}
	if out != "hi" {
		t.Errorf("cached output = %q, want hi", out)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1 (cached second pass)", got)
	}
	hits, misses := tk.Stats()
	if hits != 1 || misses != 1 {
		t.Errorf("stats = (hits=%d, misses=%d), want (1,1)", hits, misses)
	}
}

func TestTrackerRendersOnChange(t *testing.T) {
	var calls int32
	tk := reconciler.NewTracker(func(n *vdom.Node) string {
		atomic.AddInt32(&calls, 1)
		return n.Text
	})
	tk.Render(vdom.CreateTextNode("a"))
	out, fresh := tk.Render(vdom.CreateTextNode("b"))
	if !fresh {
		t.Error("changed text should re-render")
	}
	if out != "b" {
		t.Errorf("output = %q, want b", out)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestTrackerResetClearsCache(t *testing.T) {
	var calls int32
	tk := reconciler.NewTracker(func(n *vdom.Node) string {
		atomic.AddInt32(&calls, 1)
		return n.Text
	})
	tk.Render(vdom.CreateTextNode("x"))
	tk.Reset()
	tk.Render(vdom.CreateTextNode("x"))
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls after Reset = %d, want 2", got)
	}
}

func TestNewTrackerNilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil RenderFunc")
		}
	}()
	reconciler.NewTracker(nil)
}

func TestTrackerConcurrent(t *testing.T) {
	tk := reconciler.NewTracker(func(n *vdom.Node) string { return n.Text })
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tk.Render(vdom.CreateTextNode("payload"))
		}(i)
	}
	wg.Wait()
	hits, misses := tk.Stats()
	if hits+misses != 16 {
		t.Fatalf("hits+misses = %d, want 16", hits+misses)
	}
	if misses < 1 {
		t.Fatalf("expected at least one miss (first-render), got %d", misses)
	}
}
