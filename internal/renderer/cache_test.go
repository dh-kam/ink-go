package renderer_test

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// countingTransform returns a transform fn that increments a shared counter
// every time it runs. Tests use the counter to assert that the renderer
// actually consults vdom.Node's transform cache instead of re-invoking the
// fn for each measure / render pass.
func countingTransform(counter *int64, transform func(string) string) func(string, int) string {
	return func(text string, _ int) string {
		atomic.AddInt64(counter, 1)
		return transform(text)
	}
}

// TestTransformCacheHitWithinFrame_NodeTransform verifies that a single render
// frame on a nested <transform> child invokes the user transform exactly once
// for the (input, index) pair, even though both the measurement pass and the
// rendering pass route through applyNodeTransform. Mirrors upstream Ink's
// internal_transform memoization on the DOMElement.
func TestTransformCacheHitWithinFrame_NodeTransform(t *testing.T) {
	var calls int64
	root := vdom.CreateElement("box", nil,
		components.Text("",
			components.Transform(countingTransform(&calls, strings.ToUpper),
				vdom.CreateTextNode("hello"),
			),
		),
	)

	output := renderer.Render(root, 40, 4)
	if !strings.Contains(output, "HELLO") {
		t.Fatalf("expected transformed output to contain HELLO, got %q", output)
	}

	// Measure pass + render pass would call the transform 2x without the
	// cache; with the cache a single render frame must hit the user fn at
	// most once.
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("expected transform to run exactly once per frame, got %d", got)
	}
}

// TestTransformCacheHitWithinFrame_LineTransform exercises the
// applyLineTransform path used when a <transform> sits directly inside a
// <box>. Both measureTextLikeNode (during yoga measurement) and
// renderTextLikeNode (during the render pass) invoke the line transform with
// the same multi-line input, so the second invocation must be served from the
// cache.
func TestTransformCacheHitWithinFrame_LineTransform(t *testing.T) {
	var calls int64
	root := vdom.CreateElement("box", nil,
		components.Transform(countingTransform(&calls, strings.ToUpper),
			vdom.CreateTextNode("hello"),
		),
	)

	output := renderer.Render(root, 40, 4)
	if !strings.Contains(output, "HELLO") {
		t.Fatalf("expected transformed output to contain HELLO, got %q", output)
	}

	// The single line "hello" yields a single user-fn call; the second
	// pass is served from the cache and never reaches the user fn.
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("expected line transform to run once per frame, got %d", got)
	}
}

// TestTransformCacheRepeatedRendersStableInput verifies that re-rendering an
// unchanged tree invokes the transform fn exactly once across both renders
// — the per-node cache survives between frames as long as the input is the
// same and no mutation has flipped the cache.
func TestTransformCacheRepeatedRendersStableInput(t *testing.T) {
	var calls int64
	root := vdom.CreateElement("box", nil,
		components.Transform(countingTransform(&calls, strings.ToUpper),
			vdom.CreateTextNode("hello"),
		),
	)

	_ = renderer.Render(root, 40, 4)
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("baseline: expected 1 transform call, got %d", got)
	}

	// Render the same tree again. Without invalidation, the second frame
	// must reuse the cached result and never hit the user fn.
	_ = renderer.Render(root, 40, 4)
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("identical re-render must not run the transform again, got %d calls", got)
	}
}

// TestTransformCacheInvalidatedOnTextChange verifies that mutating the inner
// text content via SetNodeValue (the standard mutation hook) flushes the
// cache without manual invalidation. Upstream's commitTextUpdate path
// bubbles a markNodeAsDirty / setTextNodeValue chain that effectively
// reruns the transform; goink's SetNodeValue must do the same.
func TestTransformCacheInvalidatedOnTextChange(t *testing.T) {
	var calls int64
	innerText := vdom.CreateTextNode("hello")
	transformNode := components.Transform(countingTransform(&calls, strings.ToUpper),
		innerText,
	)
	root := vdom.CreateElement("box", nil, transformNode)

	out1 := renderer.Render(root, 40, 4)
	if !strings.Contains(out1, "HELLO") {
		t.Fatalf("baseline: missing HELLO, got %q", out1)
	}

	// SetNodeValue flows through the mutation hook that walks ancestors
	// flipping their transform caches. The next render must miss the
	// cache and run the transform anew.
	innerText.SetNodeValue("world")

	out2 := renderer.Render(root, 40, 4)
	if !strings.Contains(out2, "WORLD") {
		t.Fatalf("expected WORLD after SetNodeValue, got %q", out2)
	}

	// One call for "hello", one for "world". A stale cache would leave
	// "HELLO" in the output and only count one call.
	if got := atomic.LoadInt64(&calls); got != 2 {
		t.Fatalf("expected 2 transform calls (hello + world), got %d", got)
	}
}

// TestStaticOutputCacheHitOnUnchangedSubtree exercises the staticDirty cache:
// a Static block rendered twice with no mutations should be served from the
// cached string the second time. We assert this by stashing a sentinel into
// the cache by hand and verifying it shows up in the next render.
func TestStaticOutputCacheHitOnUnchangedSubtree(t *testing.T) {
	staticRoot := components.Static(nil,
		components.Text("static-line"),
	)
	root := vdom.CreateElement("box", nil, staticRoot)

	first := renderer.RenderWithLayoutSections(root, 40, 8)
	if !strings.Contains(first.StaticOutput, "static-line") {
		t.Fatalf("baseline static output must contain static-line, got %q", first.StaticOutput)
	}

	// Overwrite the cached output with a sentinel. If the renderer truly
	// consults the cache on the second pass, the sentinel will appear
	// verbatim. If it ignores the cache and re-renders, the sentinel is
	// discarded.
	const sentinel = "CACHED-STATIC-SENTINEL\n"
	staticRoot.StoreStaticOutput(sentinel, 40, 8, false)

	second := renderer.RenderWithLayoutSections(root, 40, 8)
	if second.StaticOutput != sentinel {
		t.Fatalf("expected cache sentinel (%q), got %q", sentinel, second.StaticOutput)
	}
}

// TestStaticOutputCacheMissOnDimensionChange verifies that resizing the
// terminal between renders (different width or height) bypasses the cache,
// since layout output depends on those dimensions. Without the dimension
// guard the cached output would be wrong after a resize.
func TestStaticOutputCacheMissOnDimensionChange(t *testing.T) {
	staticRoot := components.Static(nil, components.Text("hello"))
	root := vdom.CreateElement("box", nil, staticRoot)

	_ = renderer.RenderWithLayoutSections(root, 40, 8)
	cached, ok := staticRoot.LookupStaticOutput(40, 8, false)
	if !ok {
		t.Fatalf("expected cache populated at 40x8")
	}

	// Same dimensions hit the cache.
	if _, hit := staticRoot.LookupStaticOutput(40, 8, false); !hit {
		t.Fatalf("re-lookup at same dims must hit cache")
	}

	// Different width misses.
	if _, hit := staticRoot.LookupStaticOutput(80, 8, false); hit {
		t.Fatalf("lookup at 80x8 must miss the 40x8 cache")
	}

	// Different ANSI flag misses.
	if _, hit := staticRoot.LookupStaticOutput(40, 8, true); hit {
		t.Fatalf("lookup with ansi=true must miss the ansi=false cache")
	}

	// Original key still hits.
	if again, _ := staticRoot.LookupStaticOutput(40, 8, false); again != cached {
		t.Fatalf("expected stable cached output for unchanged key, got %q vs %q", again, cached)
	}
}

// TestStaticOutputCacheInvalidatedBySetAttribute verifies that mutating a
// descendant of a Static block via the standard SetAttribute mutation hook
// propagates up to the Static ancestor and clears its cache. This mirrors
// upstream's `commitUpdate` flipping rootNode.isStaticDirty whenever
// internal_static is touched (or any descendant is mutated).
func TestStaticOutputCacheInvalidatedBySetAttribute(t *testing.T) {
	textNode := components.Text("alpha")
	staticRoot := components.Static(nil, textNode)
	root := vdom.CreateElement("box", nil, staticRoot)

	_ = renderer.RenderWithLayoutSections(root, 40, 8)
	if _, ok := staticRoot.LookupStaticOutput(40, 8, false); !ok {
		t.Fatalf("expected cache populated after first render")
	}

	// SetAttribute on a descendant must walk up to the Static ancestor
	// and flip the dirty flag, bypassing the cache on the next render.
	textNode.SetAttribute("color", "red")
	if _, ok := staticRoot.LookupStaticOutput(40, 8, false); ok {
		t.Fatalf("expected cache cleared after descendant SetAttribute")
	}
}

// TestStaticOutputCacheManualInvalidation locks down the API surface used by
// the renderer wiring: MarkStaticDirty must clear the cache so the next
// render re-renders the static block.
func TestStaticOutputCacheManualInvalidation(t *testing.T) {
	staticRoot := components.Static(nil, components.Text("alpha"))
	root := vdom.CreateElement("box", nil, staticRoot)

	first := renderer.RenderWithLayoutSections(root, 40, 8)
	if !strings.Contains(first.StaticOutput, "alpha") {
		t.Fatalf("baseline: missing alpha, got %q", first.StaticOutput)
	}

	if _, ok := staticRoot.LookupStaticOutput(40, 8, false); !ok {
		t.Fatalf("expected cache populated after first render")
	}

	staticRoot.MarkStaticDirty()
	if _, ok := staticRoot.LookupStaticOutput(40, 8, false); ok {
		t.Fatalf("expected cache cleared after MarkStaticDirty")
	}

	// Re-rendering must repopulate the cache.
	_ = renderer.RenderWithLayoutSections(root, 40, 8)
	if _, ok := staticRoot.LookupStaticOutput(40, 8, false); !ok {
		t.Fatalf("expected cache repopulated after second render")
	}
}
