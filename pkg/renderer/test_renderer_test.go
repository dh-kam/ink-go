package renderer_test

import (
	"testing"

	"github.com/dh-kam/goink.go/pkg/renderer"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func textNode(s string) *vdom.Node { return vdom.CreateTextNode(s) }

func TestRenderInitialFrame(t *testing.T) {
	inst := renderer.Render(textNode("hello"))
	defer inst.Cleanup()
	if got := inst.LastFrame(); got == "" {
		t.Fatal("LastFrame should be non-empty after Render")
	}
	if c := inst.FrameCount(); c != 1 {
		t.Fatalf("FrameCount = %d, want 1", c)
	}
}

func TestRendererOverride(t *testing.T) {
	stub := func(n *vdom.Node) string { return "STUB:" + n.Text }
	inst := renderer.Render(textNode("payload"), renderer.WithRenderer(stub))
	defer inst.Cleanup()
	if got := inst.LastFrame(); got != "STUB:payload" {
		t.Fatalf("LastFrame = %q, want STUB:payload", got)
	}
}

func TestWithRendererNilIgnored(t *testing.T) {
	// Passing a nil RenderFunc must keep the default — otherwise we'd nil-deref.
	inst := renderer.Render(textNode("x"), renderer.WithRenderer(nil))
	defer inst.Cleanup()
	if inst.LastFrame() == "" {
		t.Fatal("expected default renderer when WithRenderer(nil)")
	}
}

func TestRerenderAccumulatesFrames(t *testing.T) {
	stub := func(n *vdom.Node) string { return n.Text }
	inst := renderer.Render(textNode("one"), renderer.WithRenderer(stub))
	defer inst.Cleanup()
	inst.Rerender(textNode("two"))
	inst.Rerender(textNode("three"))
	if c := inst.FrameCount(); c != 3 {
		t.Fatalf("FrameCount = %d, want 3", c)
	}
	frames := inst.Frames()
	want := []string{"one", "two", "three"}
	for i, w := range want {
		if frames[i] != w {
			t.Errorf("frame %d = %q, want %q", i, frames[i], w)
		}
	}
}

func TestFramesReturnsDefensiveCopy(t *testing.T) {
	stub := func(n *vdom.Node) string { return n.Text }
	inst := renderer.Render(textNode("a"), renderer.WithRenderer(stub))
	defer inst.Cleanup()
	frames := inst.Frames()
	frames[0] = "MUTATED"
	if got := inst.LastFrame(); got != "a" {
		t.Fatalf("internal frame leaked through Frames(): got %q", got)
	}
}

func TestCleanupBlocksRerender(t *testing.T) {
	stub := func(n *vdom.Node) string { return n.Text }
	inst := renderer.Render(textNode("one"), renderer.WithRenderer(stub))
	inst.Cleanup()
	inst.Rerender(textNode("two"))
	if c := inst.FrameCount(); c != 0 {
		t.Fatalf("FrameCount after cleanup = %d, want 0 (cleared)", c)
	}
	if got := inst.LastFrame(); got != "" {
		t.Fatalf("LastFrame after cleanup = %q, want empty", got)
	}
}

func TestCleanupIdempotent(t *testing.T) {
	inst := renderer.Render(textNode("x"))
	inst.Cleanup()
	inst.Cleanup() // must not panic
}

func TestEmptyInstance(t *testing.T) {
	// We can't construct an Instance directly (no exported zero ctor) but we
	// can trigger the empty-frames path by Rendering then Cleaning before
	// asserting.
	inst := renderer.Render(textNode("x"))
	inst.Cleanup()
	if got := inst.LastFrame(); got != "" {
		t.Fatalf("expected empty LastFrame after cleanup, got %q", got)
	}
}
