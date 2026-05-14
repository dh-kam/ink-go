package reconciler_test

import (
	"reflect"
	"testing"

	"github.com/dh-kam/ink-go/pkg/reconciler"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func box(props vdom.Props, children ...*vdom.Node) *vdom.Node {
	return vdom.CreateElement("box", props, children...)
}

func text(s string) *vdom.Node { return vdom.CreateTextNode(s) }

func keyed(key, label string) *vdom.Node {
	n := vdom.CreateElement("box", nil, text(label))
	n.Key = key
	return n
}

// structuralEqual asserts two trees describe the same shape/content.
// Layout fields are ignored.
func structuralEqual(a, b *vdom.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type != b.Type || a.ElementType != b.ElementType || a.Text != b.Text || a.Key != b.Key {
		return false
	}
	if !reflect.DeepEqual(a.Props, b.Props) {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if !structuralEqual(a.Children[i], b.Children[i]) {
			return false
		}
	}
	return true
}

func TestDiffIdenticalTreesEmpty(t *testing.T) {
	a := box(vdom.Props{"x": 1}, text("hi"))
	b := box(vdom.Props{"x": 1}, text("hi"))
	if got := reconciler.Diff(a, b); len(got) != 0 {
		t.Fatalf("expected no patches, got %d: %+v", len(got), got)
	}
}

func TestDiffTextChange(t *testing.T) {
	a := text("old")
	b := text("new")
	patches := reconciler.Diff(a, b)
	if len(patches) != 1 || patches[0].Type != reconciler.UpdateText || patches[0].NewText != "new" {
		t.Fatalf("expected one UpdateText patch, got %+v", patches)
	}
	final, err := reconciler.ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !structuralEqual(final, b) {
		t.Fatalf("round-trip mismatch")
	}
}

func TestDiffPropsAddRemoveChange(t *testing.T) {
	a := box(vdom.Props{"a": 1, "b": 2})
	b := box(vdom.Props{"a": 1, "b": 3, "c": 4})
	patches := reconciler.Diff(a, b)
	if len(patches) != 1 || patches[0].Type != reconciler.UpdateProps {
		t.Fatalf("expected one UpdateProps, got %+v", patches)
	}
	final, err := reconciler.ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !structuralEqual(final, b) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", final.Props, b.Props)
	}
}

func TestDiffReplaceOnDifferentType(t *testing.T) {
	a := box(nil, text("a"))
	b := vdom.CreateElement("text", nil, text("b"))
	patches := reconciler.Diff(a, b)
	if len(patches) != 1 || patches[0].Type != reconciler.Replace {
		t.Fatalf("expected one Replace, got %+v", patches)
	}
	final, err := reconciler.ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if final.ElementType != "text" {
		t.Fatalf("expected replacement element type 'text', got %q", final.ElementType)
	}
}

func TestDiffChildInsert(t *testing.T) {
	a := box(nil, text("a"))
	b := box(nil, text("a"), text("b"))
	patches := reconciler.Diff(a, b)
	final, err := reconciler.ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !structuralEqual(final, b) {
		t.Fatalf("round-trip mismatch")
	}
}

func TestDiffChildRemove(t *testing.T) {
	a := box(nil, text("a"), text("b"), text("c"))
	b := box(nil, text("a"))
	patches := reconciler.Diff(a, b)
	final, err := reconciler.ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !structuralEqual(final, b) {
		t.Fatalf("round-trip mismatch: got %d kids, want 1", len(final.Children))
	}
}

func TestDiffNestedTextChange(t *testing.T) {
	a := box(nil, box(nil, text("hi")))
	b := box(nil, box(nil, text("hello")))
	patches := reconciler.Diff(a, b)
	final, err := reconciler.ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !structuralEqual(final, b) {
		t.Fatalf("round-trip mismatch")
	}
}

func TestDiffKeyedReorderUpdatesMatchedSubtrees(t *testing.T) {
	a := box(nil,
		vdom.CreateElement("text", vdom.Props{"key": "a", "color": "red"}, text("A")),
		vdom.CreateElement("text", vdom.Props{"key": "b", "color": "blue"}, text("B")),
	)
	b := box(nil,
		vdom.CreateElement("text", vdom.Props{"key": "b", "color": "cyan"}, text("B2")),
		vdom.CreateElement("text", vdom.Props{"key": "a", "color": "red"}, text("A")),
	)

	patches := reconciler.Diff(a, b)
	hasMove := false
	for _, patch := range patches {
		if patch.Type == reconciler.Move {
			hasMove = true
			break
		}
	}
	if !hasMove {
		t.Fatalf("expected keyed reorder to produce a Move patch, got %+v", patches)
	}

	final, err := reconciler.ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !structuralEqual(final, b) {
		t.Fatalf("round-trip mismatch after keyed reorder/update")
	}
	if got := final.Children[0].Key; got != "b" {
		t.Fatalf("first child key = %q, want b", got)
	}
}

func TestDiffNilOldInserts(t *testing.T) {
	b := box(nil, text("hi"))
	patches := reconciler.Diff(nil, b)
	if len(patches) != 1 || patches[0].Type != reconciler.Replace {
		t.Fatalf("expected Replace patch, got %+v", patches)
	}

	final, err := reconciler.ApplyAll(nil, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !structuralEqual(final, b) {
		t.Fatalf("round-trip mismatch")
	}
}

func TestDiffNilNewDeletesRoot(t *testing.T) {
	a := box(nil, text("hi"))
	patches := reconciler.Diff(a, nil)
	if len(patches) != 1 || patches[0].Type != reconciler.Replace {
		t.Fatalf("expected Replace patch, got %+v", patches)
	}

	final, err := reconciler.ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if final != nil {
		t.Fatalf("expected nil root after delete, got %+v", final)
	}
}

func TestDiffBothNil(t *testing.T) {
	if patches := reconciler.Diff(nil, nil); len(patches) != 0 {
		t.Fatalf("expected no patches, got %+v", patches)
	}
}

func TestApplyAllErrorPropagation(t *testing.T) {
	root := box(nil)
	bad := reconciler.Patch{Type: reconciler.Remove, Index: 5}
	if _, err := reconciler.ApplyAll(root, []reconciler.Patch{bad}); err == nil {
		t.Fatal("expected error from out-of-range Remove")
	}
}

func TestPatchTypeString(t *testing.T) {
	cases := map[reconciler.PatchType]string{
		reconciler.Insert:      "Insert",
		reconciler.Remove:      "Remove",
		reconciler.Replace:     "Replace",
		reconciler.UpdateText:  "UpdateText",
		reconciler.UpdateProps: "UpdateProps",
		reconciler.Move:        "Move",
	}
	for pt, want := range cases {
		if got := pt.String(); got != want {
			t.Errorf("%v.String() = %q, want %q", pt, got, want)
		}
	}
	unknown := reconciler.PatchType(99)
	if got := unknown.String(); got == "" {
		t.Errorf("unknown PatchType returned empty string")
	}
}
