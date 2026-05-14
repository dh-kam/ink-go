// Package reconciler implements a Virtual DOM diff/patch algorithm against
// pkg/vdom.Node trees. It is intentionally renderer-independent: callers
// generate a list of Patch values via Diff() and apply them with ApplyAll()
// (or hand them to a renderer that knows how to repaint specific subtrees).
package reconciler

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

// PatchType enumerates the discrete edits produced by Diff.
type PatchType int

const (
	// Insert adds NewNode at Index inside the parent located at Path.
	Insert PatchType = iota
	// Remove removes the child at Index from the parent located at Path.
	Remove
	// Replace swaps the node located at Path with NewNode.
	Replace
	// UpdateText sets the text content of the TextNode located at Path.
	UpdateText
	// UpdateProps merges PropsSet and removes PropsRemove on the element at Path.
	UpdateProps
	// Move shifts the child at FromIndex to ToIndex inside the parent at Path.
	Move
)

// String returns a short human-readable name for the patch type. Used for
// test failure messages and debug logging.
func (p PatchType) String() string {
	switch p {
	case Insert:
		return "Insert"
	case Remove:
		return "Remove"
	case Replace:
		return "Replace"
	case UpdateText:
		return "UpdateText"
	case UpdateProps:
		return "UpdateProps"
	case Move:
		return "Move"
	default:
		return fmt.Sprintf("PatchType(%d)", int(p))
	}
}

// Patch describes one edit to apply to a vdom tree. Path is the index path
// from the root: parent-mutating patches (Insert/Remove/Move) interpret Path
// as the parent node, node-mutating patches (Replace/UpdateText/UpdateProps)
// interpret it as the node itself.
type Patch struct {
	Type        PatchType
	Path        []int
	Index       int        // Insert / Remove
	FromIndex   int        // Move
	ToIndex     int        // Move
	NewNode     *vdom.Node // Insert / Replace (deep cloned)
	NewText     string     // UpdateText
	PropsSet    vdom.Props // UpdateProps: keys to set/overwrite
	PropsRemove []string   // UpdateProps: keys to delete
}

// Apply applies a single Patch to the tree rooted at root, returning the
// (possibly new) root. A root Replace can create the root from nil or delete
// it by using a nil NewNode; callers must use the returned value as the new
// root.
func (p Patch) Apply(root *vdom.Node) (*vdom.Node, error) {
	if p.Type == Replace && len(p.Path) == 0 {
		if p.NewNode == nil {
			return nil, nil
		}
		return p.NewNode.Clone(), nil
	}

	if root == nil {
		return nil, fmt.Errorf("reconciler: cannot apply %s to nil root", p.Type)
	}

	switch p.Type {
	case Replace:
		parent, idx, err := walkToParent(root, p.Path)
		if err != nil {
			return nil, err
		}
		if p.NewNode == nil {
			return nil, fmt.Errorf("reconciler: Replace requires NewNode")
		}
		clone := p.NewNode.Clone()
		old := parent.Children[idx]
		parent.InsertBefore(clone, old)
		parent.RemoveChild(old)
		return root, nil

	case UpdateText:
		node, err := walkTo(root, p.Path)
		if err != nil {
			return nil, err
		}
		node.SetNodeValue(p.NewText)
		return root, nil

	case UpdateProps:
		node, err := walkTo(root, p.Path)
		if err != nil {
			return nil, err
		}
		for _, key := range p.PropsRemove {
			delete(node.Props, key)
			if key == "key" {
				node.Key = ""
			}
		}
		for key, value := range p.PropsSet {
			node.SetAttribute(key, value)
		}
		return root, nil

	case Insert:
		parent, err := walkTo(root, p.Path)
		if err != nil {
			return nil, err
		}
		if p.NewNode == nil {
			return nil, fmt.Errorf("reconciler: Insert requires NewNode")
		}
		if p.Index < 0 {
			return nil, fmt.Errorf("reconciler: Insert index %d out of range [0,%d]", p.Index, len(parent.Children))
		}
		clone := p.NewNode.Clone()
		if p.Index >= len(parent.Children) {
			parent.AppendChild(clone)
		} else {
			parent.InsertBefore(clone, parent.Children[p.Index])
		}
		return root, nil

	case Remove:
		parent, err := walkTo(root, p.Path)
		if err != nil {
			return nil, err
		}
		if p.Index < 0 || p.Index >= len(parent.Children) {
			return nil, fmt.Errorf("reconciler: Remove index %d out of range [0,%d)", p.Index, len(parent.Children))
		}
		parent.RemoveChild(parent.Children[p.Index])
		return root, nil

	case Move:
		parent, err := walkTo(root, p.Path)
		if err != nil {
			return nil, err
		}
		n := len(parent.Children)
		if p.FromIndex < 0 || p.FromIndex >= n {
			return nil, fmt.Errorf("reconciler: Move from %d out of range [0,%d)", p.FromIndex, n)
		}
		if p.ToIndex < 0 || p.ToIndex >= n {
			return nil, fmt.Errorf("reconciler: Move to %d out of range [0,%d)", p.ToIndex, n)
		}
		if p.FromIndex == p.ToIndex {
			return root, nil
		}
		moving := parent.Children[p.FromIndex]
		parent.RemoveChild(moving)
		// After removal indices >= FromIndex shift by 1; recompute target.
		insertAt := p.ToIndex
		if p.ToIndex > p.FromIndex {
			insertAt = p.ToIndex // ToIndex was computed against the post-remove list
		}
		if insertAt >= len(parent.Children) {
			parent.AppendChild(moving)
		} else {
			parent.InsertBefore(moving, parent.Children[insertAt])
		}
		return root, nil
	}
	return nil, fmt.Errorf("reconciler: unknown PatchType %d", int(p.Type))
}

// ApplyAll applies patches in order, threading the (possibly replaced) root
// through each step. Returns the final root.
func ApplyAll(root *vdom.Node, patches []Patch) (*vdom.Node, error) {
	current := root
	for i, p := range patches {
		next, err := p.Apply(current)
		if err != nil {
			return nil, fmt.Errorf("patch %d (%s): %w", i, p.Type, err)
		}
		current = next
	}
	return current, nil
}

// walkTo descends the tree following child indices in path. An empty path
// returns root.
func walkTo(root *vdom.Node, path []int) (*vdom.Node, error) {
	current := root
	for i, idx := range path {
		if current == nil {
			return nil, fmt.Errorf("reconciler: walk hit nil at depth %d", i)
		}
		if idx < 0 || idx >= len(current.Children) {
			return nil, fmt.Errorf("reconciler: index %d out of range at depth %d", idx, i)
		}
		current = current.Children[idx]
	}
	return current, nil
}

// walkToParent descends to the parent of the leaf identified by path,
// returning (parent, leafIndex). Path must be non-empty.
func walkToParent(root *vdom.Node, path []int) (*vdom.Node, int, error) {
	if len(path) == 0 {
		return nil, 0, fmt.Errorf("reconciler: walkToParent requires non-empty path")
	}
	parent, err := walkTo(root, path[:len(path)-1])
	if err != nil {
		return nil, 0, err
	}
	leaf := path[len(path)-1]
	if leaf < 0 || leaf >= len(parent.Children) {
		return nil, 0, fmt.Errorf("reconciler: leaf index %d out of range", leaf)
	}
	return parent, leaf, nil
}
