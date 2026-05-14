package ink

import (
	"fmt"
	"testing"

	"github.com/dh-kam/goink.go/internal/renderer"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// buildTextTree returns a simple <Text>text</Text> element used as a stable
// fixture across cache-equality scenarios.
func buildTextTree(text string) *vdom.Node {
	return vdom.CreateElement("Text", nil, vdom.CreateTextNode(text))
}

// countingRenderer returns a render func that records the inputs it sees so
// tests can assert exactly when the cache delegated to the underlying
// renderer versus serving cached output.
func countingRenderer(label string) (func(*vdom.Node) string, *int) {
	calls := 0
	render := func(node *vdom.Node) string {
		calls++
		if node == nil {
			return ""
		}

		// Encode the call index into the output so we can prove a stale
		// cached frame was returned (vs. a fresh re-render that would
		// produce a different number).
		return fmt.Sprintf("%s#%d", label, calls)
	}

	return render, &calls
}

func TestRenderTracker_FirstRenderIsFresh(t *testing.T) {
	render, calls := countingRenderer("first")
	tracker := newRenderTracker(render)

	out, fresh := tracker.Render(buildTextTree("hello"))

	if !fresh {
		t.Fatalf("first render should report fresh=true, got false")
	}

	if out != "first#1" {
		t.Fatalf("first render should produce renderer output, got %q", out)
	}

	if *calls != 1 {
		t.Fatalf("first render should invoke renderer exactly once, got %d", *calls)
	}

	hits, misses := tracker.Stats()
	if hits != 0 || misses != 1 {
		t.Fatalf("expected (hits, misses) = (0, 1), got (%d, %d)", hits, misses)
	}
}

func TestRenderTracker_IdenticalTreeReturnsCachedOutput(t *testing.T) {
	render, calls := countingRenderer("same")
	tracker := newRenderTracker(render)

	first, _ := tracker.Render(buildTextTree("hello"))

	// Construct an independent tree with identical structure/content. The
	// reconciler's Diff should report zero patches and the tracker should
	// hand back the cached output without re-invoking the renderer.
	second, fresh := tracker.Render(buildTextTree("hello"))

	if fresh {
		t.Fatalf("second identical render should report fresh=false, got true")
	}

	if second != first {
		t.Fatalf("cached output must equal previous output: first=%q second=%q", first, second)
	}

	if *calls != 1 {
		t.Fatalf("identical tree must not re-invoke renderer (calls=%d, want 1)", *calls)
	}

	hits, misses := tracker.Stats()
	if hits != 1 || misses != 1 {
		t.Fatalf("expected (hits, misses) = (1, 1), got (%d, %d)", hits, misses)
	}
}

func TestRenderTracker_DifferentTreeIsFresh(t *testing.T) {
	render, calls := countingRenderer("diff")
	tracker := newRenderTracker(render)

	if _, _ = tracker.Render(buildTextTree("alpha")); *calls != 1 {
		t.Fatalf("setup: expected renderer to run once, got %d", *calls)
	}

	out, fresh := tracker.Render(buildTextTree("beta"))

	if !fresh {
		t.Fatalf("changed tree must report fresh=true, got false")
	}

	if out != "diff#2" {
		t.Fatalf("changed tree should re-render, got %q", out)
	}

	if *calls != 2 {
		t.Fatalf("changed tree should invoke renderer again (calls=%d, want 2)", *calls)
	}

	// A third call with a third distinct tree should also re-render.
	out3, fresh3 := tracker.Render(buildTextTree("gamma"))
	if !fresh3 || out3 != "diff#3" || *calls != 3 {
		t.Fatalf("third distinct tree: fresh=%v out=%q calls=%d (want true, \"diff#3\", 3)", fresh3, out3, *calls)
	}
}

func TestRenderTracker_ResetForcesFreshRender(t *testing.T) {
	render, calls := countingRenderer("reset")
	tracker := newRenderTracker(render)

	tracker.Render(buildTextTree("same"))

	// Without Reset, an identical tree would return cached output. Reset
	// must invalidate the cache so the next call re-renders even when the
	// tree is unchanged.
	tracker.Reset()

	out, fresh := tracker.Render(buildTextTree("same"))
	if !fresh {
		t.Fatalf("Reset followed by identical tree should be fresh=true, got false")
	}

	if out != "reset#2" {
		t.Fatalf("Reset should force a renderer invocation, got %q", out)
	}

	if *calls != 2 {
		t.Fatalf("expected renderer call count 2 after Reset, got %d", *calls)
	}

	// Reset on a freshly-reset tracker is a no-op and must not panic or
	// disturb subsequent caching.
	tracker.Reset()
	tracker.Reset()
	if _, fresh := tracker.Render(buildTextTree("same")); !fresh {
		t.Fatalf("double-Reset should still produce a fresh render on next call")
	}
}

func TestRenderTracker_NilReceiverDegradesGracefully(t *testing.T) {
	// A nil tracker is the documented "no cache attached" path used by the
	// session when it opts out. None of the methods should panic.
	var tracker *renderTracker

	out, fresh := tracker.Render(buildTextTree("nil"))
	if !fresh || out != "" {
		t.Fatalf("nil tracker should return ('' , true), got (%q, %v)", out, fresh)
	}

	tracker.Reset()

	hits, misses := tracker.Stats()
	if hits != 0 || misses != 0 {
		t.Fatalf("nil tracker stats should be zero, got (%d, %d)", hits, misses)
	}
}

func TestRenderTracker_RenderSectionsCachesIdenticalTree(t *testing.T) {
	tracker := newRenderTracker(RenderToString)

	calls := 0
	doRender := func(node *vdom.Node) renderer.RenderSections {
		calls++
		return renderer.RenderSections{Output: fmt.Sprintf("frame#%d", calls)}
	}

	ctx := sectionsCacheContext{width: 80, height: 24}
	first, fresh := tracker.RenderSections(buildTextTree("hi"), ctx, doRender)
	if !fresh || first.Output != "frame#1" {
		t.Fatalf("first call: fresh=%v out=%q (want true, frame#1)", fresh, first.Output)
	}

	second, fresh := tracker.RenderSections(buildTextTree("hi"), ctx, doRender)
	if fresh {
		t.Fatalf("identical tree should be cached, got fresh=true")
	}
	if second.Output != first.Output {
		t.Fatalf("cached sections must match: first=%q second=%q", first.Output, second.Output)
	}
	if calls != 1 {
		t.Fatalf("expected single doRender invocation, got %d", calls)
	}

	hits, misses := tracker.SectionsStats()
	if hits != 1 || misses != 1 {
		t.Fatalf("expected sections stats (1, 1), got (%d, %d)", hits, misses)
	}
}

func TestRenderTracker_RenderSectionsRefreshesOnContextChange(t *testing.T) {
	tracker := newRenderTracker(RenderToString)

	calls := 0
	doRender := func(node *vdom.Node) renderer.RenderSections {
		calls++
		return renderer.RenderSections{Output: fmt.Sprintf("frame#%d", calls)}
	}

	ctxA := sectionsCacheContext{width: 80, height: 24}
	ctxB := sectionsCacheContext{width: 80, height: 25}

	tracker.RenderSections(buildTextTree("hi"), ctxA, doRender)
	_, fresh := tracker.RenderSections(buildTextTree("hi"), ctxB, doRender)
	if !fresh {
		t.Fatalf("changed context should re-render, got cached")
	}
	if calls != 2 {
		t.Fatalf("expected 2 doRender invocations after context change, got %d", calls)
	}
}

func TestRenderTracker_RenderSectionsRefreshesOnTreeChange(t *testing.T) {
	tracker := newRenderTracker(RenderToString)

	calls := 0
	doRender := func(node *vdom.Node) renderer.RenderSections {
		calls++
		return renderer.RenderSections{Output: fmt.Sprintf("frame#%d", calls)}
	}

	ctx := sectionsCacheContext{width: 80, height: 24}
	tracker.RenderSections(buildTextTree("alpha"), ctx, doRender)
	_, fresh := tracker.RenderSections(buildTextTree("beta"), ctx, doRender)
	if !fresh {
		t.Fatalf("changed tree should re-render, got cached")
	}
	if calls != 2 {
		t.Fatalf("expected 2 doRender invocations after tree change, got %d", calls)
	}
}

func TestRenderTracker_RenderSectionsResetClearsCache(t *testing.T) {
	tracker := newRenderTracker(RenderToString)

	calls := 0
	doRender := func(node *vdom.Node) renderer.RenderSections {
		calls++
		return renderer.RenderSections{Output: fmt.Sprintf("frame#%d", calls)}
	}

	ctx := sectionsCacheContext{width: 80, height: 24}
	tracker.RenderSections(buildTextTree("hi"), ctx, doRender)
	tracker.Reset()
	_, fresh := tracker.RenderSections(buildTextTree("hi"), ctx, doRender)
	if !fresh {
		t.Fatalf("Reset should force a fresh render, got cached")
	}
	if calls != 2 {
		t.Fatalf("expected 2 doRender invocations after Reset, got %d", calls)
	}
}

func TestRenderTracker_RendererInvokedOnPropChange(t *testing.T) {
	render, calls := countingRenderer("prop")
	tracker := newRenderTracker(render)

	first := vdom.CreateElement("Box", vdom.Props{"flexDirection": "row"})
	tracker.Render(first)

	// A prop change must be visible to Diff and re-render.
	second := vdom.CreateElement("Box", vdom.Props{"flexDirection": "column"})
	_, fresh := tracker.Render(second)

	if !fresh {
		t.Fatalf("prop change must trigger fresh render, got cached")
	}

	if *calls != 2 {
		t.Fatalf("expected 2 renderer invocations after prop change, got %d", *calls)
	}
}
