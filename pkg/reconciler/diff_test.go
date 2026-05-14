package reconciler

import (
	"reflect"
	"testing"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

func TestLISIndices(t *testing.T) {
	cases := []struct {
		name string
		seq  []int
		want []int // indices into seq
	}{
		{"empty", nil, nil},
		{"single", []int{5}, []int{0}},
		{"sorted", []int{1, 2, 3}, []int{0, 1, 2}},
		{"reverse", []int{3, 2, 1}, []int{2}},
		{"mixed", []int{2, 1, 5, 3, 4}, []int{1, 3, 4}},
	}
	for _, tc := range cases {
		got := lisIndices(tc.seq)
		if len(got) != len(tc.want) {
			t.Errorf("%s: lis len = %d, want %d (got=%v)", tc.name, len(got), len(tc.want), got)
			continue
		}
		// LIS length is what matters; element identity may vary across equal-length subsequences
		// — verify the values form a non-decreasing sequence of the right length.
		prev := -1 << 30
		for _, idx := range got {
			if tc.seq[idx] < prev {
				t.Errorf("%s: not non-decreasing: %v on seq %v", tc.name, got, tc.seq)
				break
			}
			prev = tc.seq[idx]
		}
	}
}

func TestKeyedReorderProducesMoves(t *testing.T) {
	a := vdom.CreateElement("box", nil, keyedNode("k1"), keyedNode("k2"), keyedNode("k3"))
	b := vdom.CreateElement("box", nil, keyedNode("k3"), keyedNode("k1"), keyedNode("k2"))

	patches := diffNode(a, b, nil)
	moves := 0
	for _, p := range patches {
		if p.Type == Move {
			moves++
		}
	}
	if moves == 0 {
		t.Fatalf("expected at least one Move, got patches=%+v", patches)
	}
	final, err := ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	keys := make([]string, len(final.Children))
	for i, c := range final.Children {
		keys[i] = c.Key
	}
	want := []string{"k3", "k1", "k2"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("post-apply key order = %v, want %v", keys, want)
	}
}

func TestKeyedReorderPermutationProducesTargetOrder(t *testing.T) {
	a := vdom.CreateElement("box", nil, keyedNode("a"), keyedNode("b"), keyedNode("c"), keyedNode("d"))
	b := vdom.CreateElement("box", nil, keyedNode("b"), keyedNode("d"), keyedNode("c"), keyedNode("a"))

	patches := diffNode(a, b, nil)
	final, err := ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	keys := make([]string, len(final.Children))
	for i, c := range final.Children {
		keys[i] = c.Key
	}
	want := []string{"b", "d", "c", "a"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("post-apply key order = %v, want %v", keys, want)
	}
}

func TestKeyedInsertRemoveMix(t *testing.T) {
	a := vdom.CreateElement("box", nil, keyedNode("a"), keyedNode("b"), keyedNode("c"))
	b := vdom.CreateElement("box", nil, keyedNode("b"), keyedNode("d"), keyedNode("a"))

	patches := diffNode(a, b, nil)
	final, err := ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	keys := make([]string, len(final.Children))
	for i, c := range final.Children {
		keys[i] = c.Key
	}
	want := []string{"b", "d", "a"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("post-apply key order = %v, want %v", keys, want)
	}
}

func TestMixedKeysFallsBackToPositional(t *testing.T) {
	a := vdom.CreateElement("box", nil, keyedNode("a"), vdom.CreateTextNode("x"))
	b := vdom.CreateElement("box", nil, keyedNode("a"), vdom.CreateTextNode("y"))

	// Should treat as positional since one side has a child without a Key.
	patches := diffNode(a, b, nil)
	if len(patches) == 0 {
		t.Fatal("expected at least one patch")
	}
	final, err := ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if final.Children[1].Text != "y" {
		t.Fatalf("expected text 'y', got %q", final.Children[1].Text)
	}
}

func TestDuplicateKeysFallBackToPositional(t *testing.T) {
	a := vdom.CreateElement("box", nil,
		keyedTextNode("dup", "A"),
		keyedTextNode("dup", "B"),
	)
	b := vdom.CreateElement("box", nil,
		keyedTextNode("dup", "A2"),
		keyedTextNode("dup", "B2"),
	)

	patches := diffNode(a, b, nil)
	for _, patch := range patches {
		if patch.Type == Move {
			t.Fatalf("duplicate keys must not use keyed Move patches, got %+v", patches)
		}
	}

	final, err := ApplyAll(a, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if final.Children[0].Children[0].Text != "A2" || final.Children[1].Children[0].Text != "B2" {
		t.Fatalf("expected positional text updates, got %+v", final.Children)
	}
}

func TestAllKeyedHelper(t *testing.T) {
	if !allKeyed(nil) {
		t.Error("nil children should be treated as all-keyed")
	}
	if !allKeyed([]*vdom.Node{keyedNode("a"), keyedNode("b")}) {
		t.Error("all-keyed slice misclassified")
	}
	mixed := []*vdom.Node{keyedNode("a"), vdom.CreateTextNode("x")}
	if allKeyed(mixed) {
		t.Error("mixed slice misclassified")
	}
	duplicates := []*vdom.Node{keyedNode("a"), keyedNode("a")}
	if allUniquelyKeyed(duplicates) {
		t.Error("duplicate keyed slice misclassified as uniquely keyed")
	}
}

func TestSameKindRules(t *testing.T) {
	a := vdom.CreateElement("box", nil)
	b := vdom.CreateElement("box", nil)
	if !sameKind(a, b) {
		t.Error("identical boxes should be sameKind")
	}
	c := vdom.CreateElement("text", nil)
	if sameKind(a, c) {
		t.Error("different ElementType should not be sameKind")
	}
	d := vdom.CreateElement("box", nil)
	d.Key = "k"
	if sameKind(a, d) {
		t.Error("different Key should not be sameKind")
	}
}

func keyedNode(key string) *vdom.Node {
	n := vdom.CreateElement("box", nil)
	n.Key = key
	return n
}

func keyedTextNode(key string, text string) *vdom.Node {
	n := vdom.CreateElement("box", nil, vdom.CreateTextNode(text))
	n.Key = key
	return n
}
