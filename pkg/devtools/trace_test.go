package devtools

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dh-kam/goink.go/pkg/reconciler"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// --- helpers --------------------------------------------------------------

// fixedTracer returns a tracer whose now() always yields the supplied time.
// Format tests benefit from a stable timestamp; FIFO and concurrency tests
// don't care about the value but reuse the helper to stay consistent.
func fixedTracer(limit int, when time.Time) *Tracer {
	t := NewTracer(limit)
	t.now = func() time.Time { return when }
	return t
}

func makeBox(children ...*vdom.Node) *vdom.Node {
	return vdom.CreateElement("box", nil, children...)
}

func makeText(s string) *vdom.Node {
	return vdom.CreateTextNode(s)
}

// --- tests ----------------------------------------------------------------

func TestNewTracer_NegativeLimitTreatedAsUnbounded(t *testing.T) {
	tr := NewTracer(-5)
	if tr.limit != 0 {
		t.Fatalf("expected negative limit to clamp to 0, got %d", tr.limit)
	}
}

func TestTracer_EmptyEntries(t *testing.T) {
	tr := NewTracer(0)
	got := tr.Entries()
	if got == nil {
		t.Fatalf("Entries() must never return nil; got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(got))
	}
	if tr.Len() != 0 {
		t.Fatalf("Len() should be 0, got %d", tr.Len())
	}
}

func TestTracer_DiffRecordsAndReturnsPatches(t *testing.T) {
	old := makeBox(makeText("a"))
	next := makeBox(makeText("b"))

	tr := NewTracer(0)
	got := tr.Diff(old, next, "text update")
	want := reconciler.Diff(old, next)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Diff returned %#v, reconciler.Diff returned %#v", got, want)
	}

	entries := tr.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Note != "text update" {
		t.Fatalf("note mismatch: %q", entries[0].Note)
	}
	if !reflect.DeepEqual(entries[0].Patches, want) {
		t.Fatalf("entry patches mismatch: %#v vs %#v", entries[0].Patches, want)
	}
}

func TestTracer_DiffRecordsNoOpDiffs(t *testing.T) {
	// Even when reconciler.Diff returns nil/empty patches the entry must
	// land on the timeline so reviewers can spot redundant re-renders.
	tree := makeBox(makeText("same"))
	tr := NewTracer(0)
	patches := tr.Diff(tree, tree, "noop")
	if len(patches) != 0 {
		t.Fatalf("expected 0 patches for identical trees, got %d", len(patches))
	}
	if tr.Len() != 1 {
		t.Fatalf("noop diff must still record an entry; got Len=%d", tr.Len())
	}
}

func TestTracer_RecordAppends(t *testing.T) {
	tr := NewTracer(0)
	patches := []reconciler.Patch{
		{Type: reconciler.Insert, Path: []int{0}, Index: 1, NewNode: makeText("x")},
	}
	tr.Record(patches, "manual")
	tr.Record(nil, "empty")

	entries := tr.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Note != "manual" || len(entries[0].Patches) != 1 {
		t.Fatalf("manual entry malformed: %#v", entries[0])
	}
	if entries[1].Note != "empty" || len(entries[1].Patches) != 0 {
		t.Fatalf("empty entry malformed: %#v", entries[1])
	}
}

func TestTracer_RecordCopiesPatches(t *testing.T) {
	tr := NewTracer(0)
	patches := []reconciler.Patch{{Type: reconciler.UpdateText, NewText: "hello"}}
	tr.Record(patches, "snap")

	// Mutating the caller-side slice must not mutate the recorded entry.
	patches[0].NewText = "MUTATED"

	entries := tr.Entries()
	if entries[0].Patches[0].NewText != "hello" {
		t.Fatalf("Record did not snapshot patches; saw %q", entries[0].Patches[0].NewText)
	}
}

func TestTracer_FIFODropsOldest(t *testing.T) {
	tr := NewTracer(3)

	tr.Record([]reconciler.Patch{{Type: reconciler.UpdateText, NewText: "1"}}, "one")
	tr.Record([]reconciler.Patch{{Type: reconciler.UpdateText, NewText: "2"}}, "two")
	tr.Record([]reconciler.Patch{{Type: reconciler.UpdateText, NewText: "3"}}, "three")

	if tr.Len() != 3 {
		t.Fatalf("expected len 3 after 3 inserts, got %d", tr.Len())
	}

	// 4th push — oldest ("one") drops.
	tr.Record([]reconciler.Patch{{Type: reconciler.UpdateText, NewText: "4"}}, "four")

	if tr.Len() != 3 {
		t.Fatalf("expected len 3 after FIFO trim, got %d", tr.Len())
	}

	notes := make([]string, 0, 3)
	for _, e := range tr.Entries() {
		notes = append(notes, e.Note)
	}
	want := []string{"two", "three", "four"}
	if !reflect.DeepEqual(notes, want) {
		t.Fatalf("FIFO order mismatch: got %v want %v", notes, want)
	}

	// Push many more in a burst — the bound holds and only the most recent
	// `limit` survive.
	for i := 0; i < 50; i++ {
		tr.Record(nil, fmt.Sprintf("burst-%d", i))
	}
	if tr.Len() != 3 {
		t.Fatalf("FIFO bound violated under burst: Len=%d", tr.Len())
	}
	last := tr.Entries()
	wantBurst := []string{"burst-47", "burst-48", "burst-49"}
	gotBurst := []string{last[0].Note, last[1].Note, last[2].Note}
	if !reflect.DeepEqual(gotBurst, wantBurst) {
		t.Fatalf("burst FIFO mismatch: got %v want %v", gotBurst, wantBurst)
	}
}

func TestTracer_UnlimitedGrowth(t *testing.T) {
	tr := NewTracer(0)
	for i := 0; i < 200; i++ {
		tr.Record(nil, fmt.Sprintf("n-%d", i))
	}
	if tr.Len() != 200 {
		t.Fatalf("expected unlimited tracer to retain all 200 entries, got %d", tr.Len())
	}
}

func TestTracer_Reset(t *testing.T) {
	tr := NewTracer(0)
	tr.Record(nil, "a")
	tr.Record(nil, "b")
	tr.Reset()
	if tr.Len() != 0 {
		t.Fatalf("Reset did not clear entries; Len=%d", tr.Len())
	}
	// Should still be usable after reset.
	tr.Record(nil, "after-reset")
	if tr.Len() != 1 {
		t.Fatalf("tracer broke after Reset; Len=%d", tr.Len())
	}
}

func TestTracer_EntriesIsDefensiveCopy(t *testing.T) {
	tr := NewTracer(0)
	tr.Record([]reconciler.Patch{{Type: reconciler.UpdateText, NewText: "orig"}}, "first")

	snap := tr.Entries()
	// Mutate the snapshot in every observable way.
	snap[0].Note = "MUTATED"
	snap[0].Patches[0].NewText = "MUTATED"
	snap = append(snap, TraceEntry{Note: "extra"})
	_ = snap

	live := tr.Entries()
	if len(live) != 1 {
		t.Fatalf("appending to snapshot leaked into tracer; Len=%d", len(live))
	}
	if live[0].Note != "first" {
		t.Fatalf("note mutation leaked: %q", live[0].Note)
	}
	if live[0].Patches[0].NewText != "orig" {
		t.Fatalf("patch mutation leaked: %q", live[0].Patches[0].NewText)
	}
}

func TestTracer_ConcurrentDiffAndRecord(t *testing.T) {
	// 8 producers, half via Diff(), half via Record(), 200 ops each.
	// Bounded tracer so we also exercise the FIFO trim path under load.
	const (
		workers       = 8
		opsPerWorker  = 200
		totalExpected = workers * opsPerWorker
		limit         = 64
	)

	tr := NewTracer(limit)
	old := makeBox(makeText("old"))
	next := makeBox(makeText("new"))

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerWorker; i++ {
				if id%2 == 0 {
					tr.Diff(old, next, fmt.Sprintf("w%d-%d", id, i))
				} else {
					tr.Record([]reconciler.Patch{{
						Type:    reconciler.UpdateText,
						NewText: fmt.Sprintf("w%d-%d", id, i),
					}}, fmt.Sprintf("rec-w%d-%d", id, i))
				}
			}
		}(w)
	}
	wg.Wait()

	// With a hard limit, post-burst Len must equal the limit (we definitely
	// pushed past it). No assertion on order is meaningful under concurrency
	// — we only care that the structure is intact and within bounds.
	if tr.Len() != limit {
		t.Fatalf("expected Len=%d after %d concurrent ops, got %d",
			limit, totalExpected, tr.Len())
	}

	entries := tr.Entries()
	if len(entries) != limit {
		t.Fatalf("Entries() returned %d, want %d", len(entries), limit)
	}
	for i, e := range entries {
		if e.At.IsZero() {
			t.Fatalf("entry %d has zero timestamp", i)
		}
	}
}

func TestFormatEntry_WithPatches(t *testing.T) {
	when := time.Date(2026, 4, 25, 10, 23, 1, 234_000_000, time.UTC)
	e := TraceEntry{
		At: when,
		Patches: []reconciler.Patch{
			{Type: reconciler.Insert},
			{Type: reconciler.UpdateText},
			{Type: reconciler.UpdateText},
		},
		Note: "useState update",
	}
	got := FormatEntry(e)
	want := "[10:23:01.234] (3 patches: 1 Insert, 2 UpdateText) useState update"
	if got != want {
		t.Fatalf("FormatEntry mismatch:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestFormatEntry_NoPatchesNoNote(t *testing.T) {
	when := time.Date(2026, 4, 25, 10, 23, 1, 0, time.UTC)
	got := FormatEntry(TraceEntry{At: when})
	want := "[10:23:01.000] (0 patches)"
	if got != want {
		t.Fatalf("FormatEntry empty mismatch:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestFormatEntry_SinglePatchSingularLabel(t *testing.T) {
	when := time.Date(2026, 4, 25, 10, 23, 1, 5_000_000, time.UTC)
	e := TraceEntry{
		At:      when,
		Patches: []reconciler.Patch{{Type: reconciler.Move}},
		Note:    "resize",
	}
	got := FormatEntry(e)
	want := "[10:23:01.005] (1 patch: 1 Move) resize"
	if got != want {
		t.Fatalf("FormatEntry singular mismatch:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestFormatPatch_AllVariants(t *testing.T) {
	cases := []struct {
		name  string
		patch reconciler.Patch
		want  string
	}{
		{
			name:  "insert",
			patch: reconciler.Patch{Type: reconciler.Insert, Path: []int{0, 1}, Index: 2},
			want:  "Insert at [0,1] index=2",
		},
		{
			name:  "remove",
			patch: reconciler.Patch{Type: reconciler.Remove, Path: []int{3}, Index: 0},
			want:  "Remove at [3] index=0",
		},
		{
			name:  "replace",
			patch: reconciler.Patch{Type: reconciler.Replace, Path: []int{1, 2, 3}},
			want:  "Replace at [1,2,3]",
		},
		{
			name:  "replace_root",
			patch: reconciler.Patch{Type: reconciler.Replace},
			want:  "Replace at []",
		},
		{
			name:  "update_text",
			patch: reconciler.Patch{Type: reconciler.UpdateText, Path: []int{0}, NewText: "hi"},
			want:  `UpdateText at [0] text="hi"`,
		},
		{
			name: "update_props",
			patch: reconciler.Patch{
				Type:        reconciler.UpdateProps,
				Path:        []int{2},
				PropsSet:    vdom.Props{"a": 1, "b": 2},
				PropsRemove: []string{"c"},
			},
			want: "UpdateProps at [2] set=2 remove=1",
		},
		{
			name:  "move",
			patch: reconciler.Patch{Type: reconciler.Move, Path: []int{0}, FromIndex: 3, ToIndex: 1},
			want:  "Move at [0] from=3 to=1",
		},
		{
			name:  "unknown",
			patch: reconciler.Patch{Type: reconciler.PatchType(99), Path: []int{}},
			want:  "PatchType(99) at []",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatPatch(tc.patch)
			if got != tc.want {
				t.Fatalf("FormatPatch mismatch:\n  got:  %q\n  want: %q", got, tc.want)
			}
		})
	}
}

func TestFormatEntry_UsesTracerTimestamp(t *testing.T) {
	// End-to-end: a tracer-recorded entry, when fed through FormatEntry,
	// reproduces the timestamp the tracer captured at record time.
	when := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	tr := fixedTracer(0, when)
	tr.Record([]reconciler.Patch{{Type: reconciler.Insert}}, "x")

	line := FormatEntry(tr.Entries()[0])
	if !strings.HasPrefix(line, "[12:00:00.000]") {
		t.Fatalf("expected formatted line to use captured time; got %q", line)
	}
	if !strings.HasSuffix(line, " x") {
		t.Fatalf("expected formatted line to end with note; got %q", line)
	}
}
