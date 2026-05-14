package devtools

// Tracer collects a chronological log of vdom diff events for inspection
// during development.  Entries can come from Diff (which delegates to
// reconciler.Diff and records the result) or from Record (for callers that
// already have patches in hand from another source — replays, network sync,
// etc.).
//
// The type is safe for concurrent use; the public API never returns the
// internal slice to callers, all reads return defensive copies.

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dh-kam/goink.go/pkg/reconciler"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TraceEntry is one diff event in chronological order.
type TraceEntry struct {
	At      time.Time
	Patches []reconciler.Patch
	Note    string // optional caller-provided context (e.g., "useState update", "resize")
}

// Tracer accumulates TraceEntry items thread-safely.
//
// When limit is greater than zero the tracer behaves as a bounded ring:
// recording an entry while at capacity drops the oldest one (FIFO).  A limit
// of zero means unbounded growth — useful in short-lived tests, but callers
// hosting a long-running app should always pass a positive bound.
type Tracer struct {
	mu      sync.Mutex
	entries []TraceEntry
	limit   int // max entries; 0 = unlimited
	now     func() time.Time
}

// NewTracer constructs a Tracer.  Pass limit = 0 for an unbounded log.
func NewTracer(limit int) *Tracer {
	if limit < 0 {
		limit = 0
	}
	return &Tracer{
		limit: limit,
		now:   time.Now,
	}
}

// Diff records a Diff(old, new) call and returns the resulting patches.
//
// The patches returned are exactly what reconciler.Diff produced — callers
// that need to apply them can do so without re-running the diff.  When the
// diff produces no patches the entry is still recorded so the timeline shows
// when the no-op happened (useful for debugging unexpected re-renders).
func (t *Tracer) Diff(old, new *vdom.Node, note string) []reconciler.Patch {
	patches := reconciler.Diff(old, new)
	t.append(TraceEntry{
		At:      t.now(),
		Patches: patches,
		Note:    note,
	})
	return patches
}

// Record manually appends an entry (for non-Diff sources).
//
// The patches slice is shallow-copied so later mutations by the caller do
// not retroactively change the recorded history.
func (t *Tracer) Record(patches []reconciler.Patch, note string) {
	var copied []reconciler.Patch
	if len(patches) > 0 {
		copied = make([]reconciler.Patch, len(patches))
		copy(copied, patches)
	}
	t.append(TraceEntry{
		At:      t.now(),
		Patches: copied,
		Note:    note,
	})
}

// append is the single mutation point — keeps the FIFO trim logic in one
// place so Diff and Record stay symmetrical.
func (t *Tracer) append(e TraceEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = append(t.entries, e)
	if t.limit > 0 && len(t.entries) > t.limit {
		// Drop the oldest entries until we're back within the bound. We use
		// copy + reslice (rather than entries[1:]) so the backing array
		// doesn't grow without bound when running near capacity.
		drop := len(t.entries) - t.limit
		copy(t.entries, t.entries[drop:])
		t.entries = t.entries[:t.limit]
	}
}

// Entries returns a defensive copy of the recorded timeline.
//
// The outer slice and each entry's Patches slice are independent of the
// tracer's internal state; callers may freely sort, filter, or mutate the
// result.
func (t *Tracer) Entries() []TraceEntry {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.entries) == 0 {
		return []TraceEntry{}
	}
	out := make([]TraceEntry, len(t.entries))
	for i, e := range t.entries {
		copied := TraceEntry{At: e.At, Note: e.Note}
		if len(e.Patches) > 0 {
			copied.Patches = make([]reconciler.Patch, len(e.Patches))
			copy(copied.Patches, e.Patches)
		}
		out[i] = copied
	}
	return out
}

// Reset clears all entries.
func (t *Tracer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = nil
}

// Len returns the number of currently retained entries — handy for tests
// and for monitoring how close a bounded tracer is to its capacity.
func (t *Tracer) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.entries)
}

// FormatEntry returns a one-line summary of a TraceEntry, e.g.
//
//	[10:23:01.234] (3 patches: 1 Insert, 2 UpdateText) note
//
// When the entry has zero patches the breakdown collapses to "(0 patches)".
// When Note is empty the trailing space is omitted.
func FormatEntry(e TraceEntry) string {
	stamp := e.At.Format("15:04:05.000")
	count := len(e.Patches)

	var summary string
	if count == 0 {
		summary = "(0 patches)"
	} else {
		summary = fmt.Sprintf("(%d %s: %s)", count, pluralize("patch", count, "patches"), summarizePatches(e.Patches))
	}

	if e.Note == "" {
		return fmt.Sprintf("[%s] %s", stamp, summary)
	}
	return fmt.Sprintf("[%s] %s %s", stamp, summary, e.Note)
}

// FormatPatch renders a single patch as a debug-friendly one-liner.  The
// shape is intentionally compact so a TraceEntry's Patches can be dumped
// line-by-line without overwhelming the terminal.
func FormatPatch(p reconciler.Patch) string {
	switch p.Type {
	case reconciler.Insert:
		return fmt.Sprintf("Insert at %s index=%d", formatPath(p.Path), p.Index)
	case reconciler.Remove:
		return fmt.Sprintf("Remove at %s index=%d", formatPath(p.Path), p.Index)
	case reconciler.Replace:
		return fmt.Sprintf("Replace at %s", formatPath(p.Path))
	case reconciler.UpdateText:
		return fmt.Sprintf("UpdateText at %s text=%q", formatPath(p.Path), p.NewText)
	case reconciler.UpdateProps:
		return fmt.Sprintf("UpdateProps at %s set=%d remove=%d",
			formatPath(p.Path), len(p.PropsSet), len(p.PropsRemove))
	case reconciler.Move:
		return fmt.Sprintf("Move at %s from=%d to=%d", formatPath(p.Path), p.FromIndex, p.ToIndex)
	default:
		return fmt.Sprintf("%s at %s", p.Type, formatPath(p.Path))
	}
}

// summarizePatches groups patches by type and renders "1 Insert, 2 UpdateText"
// in deterministic (alphabetical) order so log output is diff-friendly.
func summarizePatches(patches []reconciler.Patch) string {
	counts := make(map[string]int, 6)
	for _, p := range patches {
		counts[p.Type.String()]++
	}
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%d %s", counts[name], name))
	}
	return strings.Join(parts, ", ")
}

// formatPath renders a Path slice as the JSON-array-ish "[0,1,2]" form so
// it reads the same as the Patch struct field in source — empty paths show
// as "[]" rather than "<nil>".
func formatPath(path []int) string {
	if len(path) == 0 {
		return "[]"
	}
	parts := make([]string, len(path))
	for i, v := range path {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func pluralize(singular string, n int, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
