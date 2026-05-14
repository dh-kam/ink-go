package ink

import (
	"sync"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/reconciler"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// renderTracker is a thin pkg/ink-local wrapper around reconciler.Tracker.
// It exists so the mounted Instance can talk to a single, package-private
// type instead of plumbing reconciler.* into every call site, and so we can
// later attach session-specific bookkeeping (resize-driven Reset, metrics)
// without changing the underlying reconciler API.
//
// The wrapper deliberately mirrors Tracker's surface area — Render returns
// (output, fresh) and Reset drops the cached frame — so callers can treat
// it as a drop-in cache layer.
//
// In addition to the string-based Render API, the tracker exposes
// RenderSections, which caches a renderer.RenderSections payload keyed on
// (vdom tree, render context). Mounted sessions go through the sections
// path so that an idle tick — same component output, same width/height,
// same screen-reader/ANSI mode — short-circuits the renderer entirely
// instead of repainting a full frame.
type renderTracker struct {
	tracker *reconciler.Tracker

	mu                   sync.Mutex
	prevTree             *vdom.Node
	prevSections         renderer.RenderSections
	prevSectionsCtx      sectionsCacheContext
	prevSectionsValid    bool
	sectionsHits         int
	sectionsMisses       int
}

// sectionsCacheContext captures the render-time inputs that, alongside the
// tree itself, determine the rendered output. A change in any of these
// fields invalidates the cached sections even if the tree is byte-identical.
type sectionsCacheContext struct {
	width                int
	height               int
	screenReader         bool
	ansi                 bool
	previousStaticCounts []int
}

func (c sectionsCacheContext) equals(other sectionsCacheContext) bool {
	if c.width != other.width || c.height != other.height {
		return false
	}
	if c.screenReader != other.screenReader || c.ansi != other.ansi {
		return false
	}
	if len(c.previousStaticCounts) != len(other.previousStaticCounts) {
		return false
	}
	for i, n := range c.previousStaticCounts {
		if other.previousStaticCounts[i] != n {
			return false
		}
	}
	return true
}

func (c sectionsCacheContext) clone() sectionsCacheContext {
	out := c
	if c.previousStaticCounts != nil {
		out.previousStaticCounts = append([]int(nil), c.previousStaticCounts...)
	}
	return out
}

// newRenderTracker builds a renderTracker that delegates to render. render
// must be non-nil; reconciler.NewTracker panics otherwise and we surface
// that contract directly rather than papering over a programming bug.
func newRenderTracker(render func(*vdom.Node) string) *renderTracker {
	return &renderTracker{tracker: reconciler.NewTracker(reconciler.RenderFunc(render))}
}

// Render returns the rendered output for node. fresh is true when render
// was actually invoked, false when the cached output from a structurally
// identical previous tree was reused. A nil receiver is treated as "always
// fresh" so callers can opt out by leaving the cache unset.
func (t *renderTracker) Render(node *vdom.Node) (out string, fresh bool) {
	if t == nil || t.tracker == nil {
		// No cache attached — degrade gracefully. Callers that need the
		// rendered string should not be using this path; returning the
		// empty string here matches what the renderer produces for a nil
		// tree and keeps the (out, fresh) contract intact.
		return "", true
	}

	return t.tracker.Render(node)
}

// RenderSections is the multi-piece counterpart to Render: it caches the
// last (tree, ctx, sections) triple and short-circuits doRender when the
// next tree diffs to zero patches against the cached one and ctx is
// equal. fresh reports whether doRender was actually invoked.
//
// A nil receiver always calls doRender — sessions that opt out by leaving
// renderCache unset retain their previous behavior of rendering every
// commit.
func (t *renderTracker) RenderSections(
	node *vdom.Node,
	ctx sectionsCacheContext,
	doRender func(*vdom.Node) renderer.RenderSections,
) (renderer.RenderSections, bool) {
	if t == nil || doRender == nil {
		if doRender != nil {
			return doRender(node), true
		}
		return renderer.RenderSections{}, true
	}

	t.mu.Lock()
	prevTree := t.prevTree
	prevSections := t.prevSections
	prevCtx := t.prevSectionsCtx
	prevValid := t.prevSectionsValid
	t.mu.Unlock()

	if prevValid && prevTree != nil && prevCtx.equals(ctx) {
		if patches := reconciler.Diff(prevTree, node); len(patches) == 0 {
			t.mu.Lock()
			t.sectionsHits++
			// Refresh the stored tree to the live node so subsequent diffs
			// keep working against the freshly-built component output.
			t.prevTree = node
			t.mu.Unlock()
			return prevSections, false
		}
	}

	sections := doRender(node)

	t.mu.Lock()
	t.prevTree = node
	t.prevSections = sections
	t.prevSectionsCtx = ctx.clone()
	t.prevSectionsValid = true
	t.sectionsMisses++
	t.mu.Unlock()
	return sections, true
}

// Reset clears the cached tree/output so the next Render is treated as a
// first render. Used after terminal resizes or option changes that
// invalidate the previous frame.
func (t *renderTracker) Reset() {
	if t == nil {
		return
	}

	if t.tracker != nil {
		t.tracker.Reset()
	}

	t.mu.Lock()
	t.prevTree = nil
	t.prevSections = renderer.RenderSections{}
	t.prevSectionsCtx = sectionsCacheContext{}
	t.prevSectionsValid = false
	t.mu.Unlock()
}

// Stats returns (hits, misses) from the underlying tracker for diagnostics
// and tests. A nil receiver reports zeros.
func (t *renderTracker) Stats() (hits, misses int) {
	if t == nil || t.tracker == nil {
		return 0, 0
	}

	return t.tracker.Stats()
}

// SectionsStats returns hit/miss counters for the multi-piece sections
// cache used by mounted sessions. A nil receiver reports zeros so tests
// and diagnostics can probe the cache without nil-checking the tracker.
func (t *renderTracker) SectionsStats() (hits, misses int) {
	if t == nil {
		return 0, 0
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sectionsHits, t.sectionsMisses
}
