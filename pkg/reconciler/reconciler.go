package reconciler

import "github.com/dh-kam/ink-go/pkg/vdom"

// Diff returns the patches required to transform old into new. The trees are
// not mutated; pass the patches to ApplyAll (or a renderer) to materialize
// the changes.
//
// When old is nil the result is a root Replace that creates the new tree.
// When new is nil the result is a root Replace that deletes the old tree.
func Diff(old, new *vdom.Node) []Patch {
	switch {
	case old == nil && new == nil:
		return nil
	case old == nil:
		return []Patch{{Type: Replace, NewNode: new}}
	case new == nil:
		return []Patch{{Type: Replace, NewNode: nil}}
	}
	return diffNode(old, new, nil)
}
