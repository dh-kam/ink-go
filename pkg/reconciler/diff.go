package reconciler

import (
	"reflect"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// diffNode produces the patches needed to transform old into new at the
// given path. Both old and new are non-nil — callers handle the nil cases
// (Insert/Remove) at the children level.
func diffNode(old, new *vdom.Node, path []int) []Patch {
	if !sameKind(old, new) {
		return []Patch{{Type: Replace, Path: clonePath(path), NewNode: new}}
	}
	if old.Type == vdom.TextNode {
		if old.Text == new.Text {
			return nil
		}
		return []Patch{{Type: UpdateText, Path: clonePath(path), NewText: new.Text}}
	}

	var patches []Patch
	if propPatch, ok := diffProps(old.Props, new.Props, path); ok {
		patches = append(patches, propPatch)
	}
	patches = append(patches, diffChildren(old, new, path)...)
	return patches
}

// sameKind returns true when two nodes can be updated in-place (no Replace).
// Element nodes must share ElementType and Key; text nodes always match.
func sameKind(a, b *vdom.Node) bool {
	if a.Type != b.Type {
		return false
	}
	if a.Type == vdom.ElementNode {
		if a.ElementType != b.ElementType {
			return false
		}
		if a.Key != b.Key {
			return false
		}
	}
	return true
}

// diffProps emits a single UpdateProps patch describing the delta between
// old and new. Returns ok=false when no change is needed.
func diffProps(old, new vdom.Props, path []int) (Patch, bool) {
	set := vdom.Props{}
	var remove []string

	for k, v := range new {
		if oldV, ok := old[k]; !ok || !reflect.DeepEqual(oldV, v) {
			set[k] = v
		}
	}
	for k := range old {
		if _, ok := new[k]; !ok {
			remove = append(remove, k)
		}
	}

	if len(set) == 0 && len(remove) == 0 {
		return Patch{}, false
	}
	return Patch{
		Type:        UpdateProps,
		Path:        clonePath(path),
		PropsSet:    set,
		PropsRemove: remove,
	}, true
}

// diffChildren chooses between keyed and positional diff strategies.
// If every child on both sides has a non-empty Key the keyed path runs;
// otherwise positional fallback (mirrors React's mixed-keys rule).
func diffChildren(oldParent, newParent *vdom.Node, path []int) []Patch {
	if allUniquelyKeyed(oldParent.Children) && allUniquelyKeyed(newParent.Children) {
		return diffKeyedChildren(oldParent.Children, newParent.Children, path)
	}
	return diffPositionalChildren(oldParent.Children, newParent.Children, path)
}

func allUniquelyKeyed(children []*vdom.Node) bool {
	if !allKeyed(children) {
		return false
	}

	seen := make(map[string]struct{}, len(children))
	for _, child := range children {
		if child == nil {
			return false
		}

		if _, exists := seen[child.Key]; exists {
			return false
		}
		seen[child.Key] = struct{}{}
	}

	return true
}

func allKeyed(children []*vdom.Node) bool {
	if len(children) == 0 {
		return true
	}
	for _, c := range children {
		if c == nil || c.Key == "" {
			return false
		}
	}
	return true
}

// diffPositionalChildren pairs children by index, emitting Insert/Remove
// for length differences.
func diffPositionalChildren(oldKids, newKids []*vdom.Node, path []int) []Patch {
	var patches []Patch

	common := min(len(oldKids), len(newKids))
	for i := 0; i < common; i++ {
		patches = append(patches, diffNode(oldKids[i], newKids[i], append(path, i))...)
	}
	// Removals — iterate from the end so indices stay stable as patches apply
	// in order (each Remove drops the last current child).
	for i := len(oldKids) - 1; i >= len(newKids); i-- {
		patches = append(patches, Patch{
			Type:  Remove,
			Path:  clonePath(path),
			Index: i,
		})
	}
	// Insertions
	for i := len(oldKids); i < len(newKids); i++ {
		patches = append(patches, Patch{
			Type:    Insert,
			Path:    clonePath(path),
			Index:   i,
			NewNode: newKids[i],
		})
	}
	return patches
}

// diffKeyedChildren matches children by Key, updates surviving children while
// they are still at old indices, then rewrites direct-child order.
func diffKeyedChildren(oldKids, newKids []*vdom.Node, path []int) []Patch {
	oldIndex := make(map[string]int, len(oldKids))
	for i, c := range oldKids {
		oldIndex[c.Key] = i
	}

	var patches []Patch

	// Step 1: in-place updates for keys that survive. These must be anchored
	// at the old child index because parent-level removals/moves/inserts have
	// not been applied yet.
	matchedOldIdx := make([]int, len(newKids))
	for i := range matchedOldIdx {
		matchedOldIdx[i] = -1
	}
	for i, nc := range newKids {
		if oi, ok := oldIndex[nc.Key]; ok {
			matchedOldIdx[i] = oi
			patches = append(patches, diffNode(oldKids[oi], nc, append(path, oi))...)
		}
	}

	// Step 2: removals for old keys not present in new. Apply in reverse so
	// each Remove pops from the current end of the list.
	newKeys := make(map[string]struct{}, len(newKids))
	for _, nc := range newKids {
		newKeys[nc.Key] = struct{}{}
	}
	for i := len(oldKids) - 1; i >= 0; i-- {
		if _, kept := newKeys[oldKids[i].Key]; !kept {
			patches = append(patches, Patch{Type: Remove, Path: clonePath(path), Index: i})
		}
	}

	// Step 3: now the parent's children are the kept set in old-order.
	// Walk new children left to right and rewrite currentOrder to mirror every
	// emitted parent-level patch.
	currentOrder := make([]string, 0, len(oldKids))
	for _, c := range oldKids {
		if _, kept := newKeys[c.Key]; kept {
			currentOrder = append(currentOrder, c.Key)
		}
	}

	for i, nc := range newKids {
		if matchedOldIdx[i] < 0 {
			// new key: insert at i
			patches = append(patches, Patch{
				Type:    Insert,
				Path:    clonePath(path),
				Index:   i,
				NewNode: nc,
			})
			currentOrder = insertAt(currentOrder, i, nc.Key)
			continue
		}
		// kept child — find its current position
		pos := indexOf(currentOrder, nc.Key)
		if pos == i {
			continue
		}
		patches = append(patches, Patch{
			Type:      Move,
			Path:      clonePath(path),
			FromIndex: pos,
			ToIndex:   i,
		})
		currentOrder = moveItem(currentOrder, pos, i)
	}

	return patches
}

func indexOf(xs []string, s string) int {
	for i, v := range xs {
		if v == s {
			return i
		}
	}
	return -1
}

func insertAt(xs []string, i int, s string) []string {
	if i >= len(xs) {
		return append(xs, s)
	}
	xs = append(xs, "")
	copy(xs[i+1:], xs[i:])
	xs[i] = s
	return xs
}

func moveItem(xs []string, from, to int) []string {
	if from == to {
		return xs
	}
	v := xs[from]
	xs = append(xs[:from], xs[from+1:]...)
	if to >= len(xs) {
		return append(xs, v)
	}
	xs = append(xs, "")
	copy(xs[to+1:], xs[to:])
	xs[to] = v
	return xs
}

// lisIndices returns indices of a longest non-decreasing subsequence of seq.
// O(n log n) patience-sort style. Empty / single-element inputs return
// trivially.
func lisIndices(seq []int) []int {
	if len(seq) == 0 {
		return nil
	}
	tails := []int{} // tails[k] = index in seq of smallest tail of length k+1 LIS
	prev := make([]int, len(seq))
	for i := range prev {
		prev[i] = -1
	}
	for i, v := range seq {
		// binary search for the leftmost tails position with seq[tails[j]] > v
		lo, hi := 0, len(tails)
		for lo < hi {
			mid := (lo + hi) / 2
			if seq[tails[mid]] <= v {
				lo = mid + 1
			} else {
				hi = mid
			}
		}
		if lo > 0 {
			prev[i] = tails[lo-1]
		}
		if lo == len(tails) {
			tails = append(tails, i)
		} else {
			tails[lo] = i
		}
	}
	// Reconstruct
	out := make([]int, len(tails))
	k := tails[len(tails)-1]
	for i := len(tails) - 1; i >= 0; i-- {
		out[i] = k
		k = prev[k]
	}
	return out
}

func clonePath(path []int) []int {
	out := make([]int, len(path))
	copy(out, path)
	return out
}
