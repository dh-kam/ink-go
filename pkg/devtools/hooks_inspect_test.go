package devtools

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/hooks"
	inkinput "github.com/dh-kam/goink.go/pkg/input"
)

// TestSnapshotContext_Nil verifies that passing a nil context returns a
// zero-valued snapshot rather than panicking. This is important because
// devtools panels may attempt to introspect a context before any render has
// produced one.
func TestSnapshotContext_Nil(t *testing.T) {
	snap := SnapshotContext(nil)

	if len(snap.States) != 0 {
		t.Errorf("expected 0 states, got %d", len(snap.States))
	}
	if len(snap.Refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(snap.Refs))
	}
	if len(snap.Effects) != 0 {
		t.Errorf("expected 0 effects, got %d", len(snap.Effects))
	}
	if len(snap.Memos) != 0 {
		t.Errorf("expected 0 memos, got %d", len(snap.Memos))
	}
	if len(snap.Callbacks) != 0 {
		t.Errorf("expected 0 callbacks, got %d", len(snap.Callbacks))
	}
	if snap.Inputs != 0 || snap.Foci != 0 {
		t.Errorf("expected 0 inputs/foci, got inputs=%d foci=%d", snap.Inputs, snap.Foci)
	}
}

// TestSnapshotContext_Empty checks that a freshly-created Context with no
// hooks produces zero counts across the board.
func TestSnapshotContext_Empty(t *testing.T) {
	ctx := hooks.NewContext()
	ctx.Reset()

	snap := SnapshotContext(ctx)

	if len(snap.States) != 0 {
		t.Errorf("expected 0 states, got %d", len(snap.States))
	}
	if len(snap.Refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(snap.Refs))
	}
	if len(snap.Effects) != 0 {
		t.Errorf("expected 0 effects, got %d", len(snap.Effects))
	}
	if len(snap.Memos) != 0 {
		t.Errorf("expected 0 memos, got %d", len(snap.Memos))
	}
	if len(snap.Callbacks) != 0 {
		t.Errorf("expected 0 callbacks, got %d", len(snap.Callbacks))
	}
	if snap.Inputs != 0 || snap.Foci != 0 {
		t.Errorf("expected 0 inputs/foci, got inputs=%d foci=%d", snap.Inputs, snap.Foci)
	}
}

// TestSnapshotContext_AfterUseState confirms that values stored via UseState
// flow through SnapshotContext into the States slice intact.
func TestSnapshotContext_AfterUseState(t *testing.T) {
	ctx := hooks.NewContext()
	ctx.Reset()

	val, _ := hooks.UseState(ctx, 42)
	if val.(int) != 42 {
		t.Fatalf("UseState returned wrong initial value: %v", val)
	}

	snap := SnapshotContext(ctx)
	if len(snap.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(snap.States))
	}
	got := snap.States[0]
	if got.Kind != "state" {
		t.Errorf("expected Kind=state, got %q", got.Kind)
	}
	if got.Index != 0 {
		t.Errorf("expected Index=0, got %d", got.Index)
	}
	if got.Value.(int) != 42 {
		t.Errorf("expected Value=42, got %v", got.Value)
	}
}

// TestSnapshotContext_MultipleHooks exercises mixed slot ordering: two
// UseState calls plus one UseRef. Each slot index inside its own kind must be
// monotonic and the stored values must round-trip.
func TestSnapshotContext_MultipleHooks(t *testing.T) {
	ctx := hooks.NewContext()
	ctx.Reset()

	_, _ = hooks.UseState(ctx, 1)
	_, _ = hooks.UseState(ctx, "hello")
	ref := hooks.UseRef(ctx, "ref-init")
	if ref == nil {
		t.Fatal("UseRef returned nil")
	}

	snap := SnapshotContext(ctx)

	if len(snap.States) != 2 {
		t.Fatalf("expected 2 states, got %d", len(snap.States))
	}
	if snap.States[0].Index != 0 || snap.States[0].Value.(int) != 1 {
		t.Errorf("state[0] mismatch: %+v", snap.States[0])
	}
	if snap.States[1].Index != 1 || snap.States[1].Value.(string) != "hello" {
		t.Errorf("state[1] mismatch: %+v", snap.States[1])
	}

	if len(snap.Refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(snap.Refs))
	}
	if snap.Refs[0].Index != 0 || snap.Refs[0].Value.(string) != "ref-init" {
		t.Errorf("ref[0] mismatch: %+v", snap.Refs[0])
	}

	// Mutating the ref via the public API must be visible on a fresh
	// snapshot, proving the snapshot reflects current state and not a stale
	// copy of the original initial value.
	ref.SetCurrent("ref-updated")
	snap2 := SnapshotContext(ctx)
	if snap2.Refs[0].Value.(string) != "ref-updated" {
		t.Errorf("expected updated ref value, got %v", snap2.Refs[0].Value)
	}
}

// TestSnapshotContext_EffectMemoCallback ensures that effect/callback slots
// are counted but their callbacks are not surfaced through Value, while
// memo's cached value is exposed.
func TestSnapshotContext_EffectMemoCallback(t *testing.T) {
	ctx := hooks.NewContext()
	ctx.Reset()

	hooks.UseEffect(ctx, func() func() { return nil }, []interface{}{})
	memoVal := hooks.UseMemo(ctx, func() interface{} { return 99 }, []interface{}{1})
	if memoVal.(int) != 99 {
		t.Fatalf("UseMemo returned wrong value: %v", memoVal)
	}
	cb := hooks.UseCallback(ctx, func() {}, []interface{}{1})
	if cb == nil {
		t.Fatal("UseCallback returned nil")
	}

	snap := SnapshotContext(ctx)

	if len(snap.Effects) != 1 {
		t.Fatalf("expected 1 effect, got %d", len(snap.Effects))
	}
	if snap.Effects[0].Value != nil {
		t.Errorf("effect Value must be nil for safety, got %v", snap.Effects[0].Value)
	}
	if snap.Effects[0].Kind != "effect" {
		t.Errorf("expected Kind=effect, got %q", snap.Effects[0].Kind)
	}

	if len(snap.Memos) != 1 {
		t.Fatalf("expected 1 memo, got %d", len(snap.Memos))
	}
	if snap.Memos[0].Value.(int) != 99 {
		t.Errorf("expected memo Value=99, got %v", snap.Memos[0].Value)
	}

	if len(snap.Callbacks) != 1 {
		t.Fatalf("expected 1 callback, got %d", len(snap.Callbacks))
	}
	if snap.Callbacks[0].Value != nil {
		t.Errorf("callback Value must be nil for safety, got %v", snap.Callbacks[0].Value)
	}
	if snap.Callbacks[0].Kind != "callback" {
		t.Errorf("expected Kind=callback, got %q", snap.Callbacks[0].Kind)
	}
}

// TestFormat_EmptySnapshot confirms Format produces a sensible report on a
// zero-valued snapshot.
func TestFormat_EmptySnapshot(t *testing.T) {
	out := ContextSnapshot{}.Format()

	for _, want := range []string{
		"States    (0)",
		"Refs      (0)",
		"Effects   (0)",
		"Memos     (0)",
		"Callbacks (0)",
		"Inputs    : 0",
		"Foci      : 0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Format() missing %q\nfull output:\n%s", want, out)
		}
	}
}

// TestFormat_PopulatedSnapshot verifies that counts and value text appear in
// the rendered report.
func TestFormat_PopulatedSnapshot(t *testing.T) {
	snap := ContextSnapshot{
		States: []HookInfo{
			{Kind: "state", Index: 0, Value: 42},
			{Kind: "state", Index: 1, Value: "hello"},
		},
		Refs: []HookInfo{
			{Kind: "ref", Index: 0, Value: nil},
		},
		Effects: []HookInfo{
			{Kind: "effect", Index: 0},
			{Kind: "effect", Index: 1},
			{Kind: "effect", Index: 2},
		},
		Memos: []HookInfo{
			{Kind: "memo", Index: 0, Value: 7},
		},
		Callbacks: []HookInfo{
			{Kind: "callback", Index: 0},
		},
		Inputs: 2,
		Foci:   1,
	}

	out := snap.Format()

	for _, want := range []string{
		"States    (2)",
		"int=42",
		`string="hello"`,
		"Refs      (1)",
		"Effects   (3)",
		"(callbacks not displayed)",
		"Memos     (1)",
		"cached=int=7",
		"Callbacks (1)",
		"Inputs    : 2",
		"Foci      : 1",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Format() missing %q\nfull output:\n%s", want, out)
		}
	}
}

// TestFormatValue_Variants exercises the small switch inside formatValue so
// each rendering branch is covered.
func TestFormatValue_Variants(t *testing.T) {
	if got := formatValue(nil); got != "nil" {
		t.Errorf("nil formatting wrong: %q", got)
	}
	if got := formatValue("x"); got != `string="x"` {
		t.Errorf("string formatting wrong: %q", got)
	}
	if got := formatValue(7); got != "int=7" {
		t.Errorf("int formatting wrong: %q", got)
	}
}

// TestSnapshotContext_InputsAndFoci confirms that UseInput / UseFocus slot
// counts surface through SnapshotContext.Inputs / .Foci.
func TestSnapshotContext_InputsAndFoci(t *testing.T) {
	ctx := hooks.NewContext()
	ctx.Reset()

	// A no-op InputCallback is enough; we never dispatch.
	cb := func(input string, key inkinput.HookKey) {}
	_ = hooks.UseInput(ctx, cb, nil, true)
	_ = hooks.UseInput(ctx, cb, nil, false)

	_, _, _ = hooks.UseFocus(ctx, "id-a", false, true)

	snap := SnapshotContext(ctx)
	if snap.Inputs != 2 {
		t.Errorf("expected Inputs=2, got %d", snap.Inputs)
	}
	if snap.Foci != 1 {
		t.Errorf("expected Foci=1, got %d", snap.Foci)
	}
}

// TestReadSliceField_Defensive directly exercises the defensive branches in
// readSliceField that the production callers never hit because they only
// pass real slot field names.
func TestReadSliceField_Defensive(t *testing.T) {
	type sample struct {
		Slice []int
		Num   int
	}
	s := &sample{Slice: []int{1, 2, 3}, Num: 7}
	v := reflect.ValueOf(s).Elem()

	// Missing field -> not ok.
	if _, ok := readSliceField(v, "DoesNotExist"); ok {
		t.Error("expected ok=false for missing field")
	}
	// Non-slice field -> not ok.
	if _, ok := readSliceField(v, "Num"); ok {
		t.Error("expected ok=false for non-slice field")
	}
	// Valid slice field -> usable Value.
	got, ok := readSliceField(v, "Slice")
	if !ok {
		t.Fatal("expected ok=true for slice field")
	}
	if got.Len() != 3 {
		t.Errorf("expected len 3, got %d", got.Len())
	}
}

// TestRefValueAt_Defensive walks refValueAt's branches that cannot be
// triggered through the public hooks API. Specifically: a struct whose only
// pointer field is nil, and a struct that contains no pointer-to-struct
// field at all, must both yield a nil result.
func TestRefValueAt_Defensive(t *testing.T) {
	// No pointer field at all.
	type noPtr struct {
		X int
	}
	if v := refValueAt(reflect.ValueOf(noPtr{X: 1})); v != nil {
		t.Errorf("expected nil for struct with no ptr field, got %v", v)
	}

	// Pointer field present but nil.
	type withNilPtr struct {
		P *struct{ value int }
	}
	if v := refValueAt(reflect.ValueOf(withNilPtr{})); v != nil {
		t.Errorf("expected nil for nil ptr field, got %v", v)
	}

	// Pointer to non-struct.
	x := 5
	type withPrimPtr struct {
		P *int
	}
	if v := refValueAt(reflect.ValueOf(withPrimPtr{P: &x})); v != nil {
		t.Errorf("expected nil for ptr-to-non-struct, got %v", v)
	}

	// Pointer to struct without a "value" field.
	type other struct{ Other int }
	type withOther struct{ P *other }
	if v := refValueAt(reflect.ValueOf(withOther{P: &other{Other: 9}})); v != nil {
		t.Errorf("expected nil for struct without value field, got %v", v)
	}
}

// TestMemoValueAt_Defensive triggers memoValueAt's invalid-field branch.
func TestMemoValueAt_Defensive(t *testing.T) {
	type noValue struct {
		Other int
	}
	if v := memoValueAt(reflect.ValueOf(noValue{Other: 1})); v != nil {
		t.Errorf("expected nil for struct without value field, got %v", v)
	}
}

// TestSnapshotContext_RefHookWithNilRef defensively constructs the case
// where a refHook's pointer field is nil. The helper must skip it without
// panicking. We do this by snapshotting an empty context (refs len == 0),
// then by invoking UseRef with an explicit nil initial value to make sure
// the value path also handles nil interfaces.
func TestSnapshotContext_RefHookWithNilValue(t *testing.T) {
	ctx := hooks.NewContext()
	ctx.Reset()

	r := hooks.UseRef(ctx, nil)
	if r == nil {
		t.Fatal("UseRef returned nil even though it should always return a Ref")
	}

	snap := SnapshotContext(ctx)
	if len(snap.Refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(snap.Refs))
	}
	if snap.Refs[0].Value != nil {
		t.Errorf("expected nil ref value, got %v", snap.Refs[0].Value)
	}
}
