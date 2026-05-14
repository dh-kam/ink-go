package reconciler

import (
	"sync"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// RenderFunc materializes a vdom tree into a frame string. Matches the
// shape exported by pkg/ink.RenderToString and pkg/renderer.RenderFunc, so
// a Tracker can wrap either renderer transparently.
type RenderFunc func(*vdom.Node) string

// Tracker is a thin diff-aware cache over a RenderFunc. It remembers the
// last (tree, output) pair and short-circuits Render calls when the
// incoming tree produces no patches relative to the previous one.
//
// This is the safest first-stage integration with a renderer that always
// emits a full frame — full repaints still happen on any change, but
// idle ticks (zero patches) skip the renderer entirely. Per-patch dirty
// rect optimization is left to a future stage.
//
// Tracker is safe for concurrent use; renders run outside the lock.
type Tracker struct {
	render RenderFunc

	mu     sync.Mutex
	prev   *vdom.Node
	output string
	hits   int // diagnostic: skipped renders since construction
	misses int // diagnostic: full renders performed
}

// NewTracker builds a Tracker that delegates to render. Panics if render
// is nil — there is no sensible default.
func NewTracker(render RenderFunc) *Tracker {
	if render == nil {
		panic("reconciler: NewTracker requires a non-nil RenderFunc")
	}
	return &Tracker{render: render}
}

// Render returns the rendered frame for next, re-running render only when
// Diff(prev, next) is non-empty. The bool reports whether a fresh render
// happened (true) or the cached output was reused (false).
func (t *Tracker) Render(next *vdom.Node) (string, bool) {
	t.mu.Lock()
	prev := t.prev
	cached := t.output
	t.mu.Unlock()

	// First render — no prev to diff against.
	if prev == nil {
		out := t.render(next)
		t.mu.Lock()
		t.prev = next
		t.output = out
		t.misses++
		t.mu.Unlock()
		return out, true
	}

	patches := Diff(prev, next)
	if len(patches) == 0 {
		t.mu.Lock()
		t.hits++
		t.mu.Unlock()
		return cached, false
	}

	out := t.render(next)
	t.mu.Lock()
	t.prev = next
	t.output = out
	t.misses++
	t.mu.Unlock()
	return out, true
}

// Reset drops the cached tree and output so the next Render is a full
// repaint. Useful after terminal resizes or when the renderer
// configuration changed underneath the Tracker.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.prev = nil
	t.output = ""
}

// Stats returns diagnostic counters: (cacheHits, cacheMisses).
func (t *Tracker) Stats() (hits, misses int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.hits, t.misses
}
