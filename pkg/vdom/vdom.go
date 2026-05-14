package vdom

import (
	"fmt"
	"strings"
)

// NodeType represents the type of a virtual DOM node
type NodeType int

const (
	// TextNode is a text node
	TextNode NodeType = iota
	// ElementNode is an element node
	ElementNode
)

// String returns string representation of NodeType
func (nt NodeType) String() string {
	switch nt {
	case TextNode:
		return "text"
	case ElementNode:
		return "element"
	default:
		return "unknown"
	}
}

// Props represents properties/attributes of an element
type Props map[string]interface{}

// Node represents a virtual DOM node
type Node struct {
	Type        NodeType
	ElementType string  // e.g., "box", "text"
	Text        string  // for text nodes
	Props       Props   // for element nodes
	Children    []*Node // child nodes
	Key         string  // optional key for reconciliation
	Layout      Layout  // latest computed layout metadata
	parent      *Node

	// Reconciler-side caches mirroring upstream Ink's `internal_transform` /
	// `staticNode` / `isStaticDirty` fields. transformCache memoizes the
	// last (input, index) -> output of a transform fn so the measure +
	// render passes do not re-invoke the user fn within a single frame.
	// The static* fields cache a previously rendered <Static> block so an
	// unchanged subtree avoids a layout + render round-trip.
	transformCache     *transformCacheEntry
	staticDirty        bool
	cachedStaticOutput string
	cachedStaticANSI   bool
	cachedStaticWidth  int
	cachedStaticHeight int
}

// transformCacheEntry is the per-node memoized output of an `internal_transform`
// invocation. The entry is keyed by the input string and index. Swapping the
// transform fn flows through SetAttribute -> InvalidateTransformCache, so the
// fn itself need not be part of the key.
type transformCacheEntry struct {
	input  string
	index  int
	output string
}

// Layout stores the latest computed layout information for a node.
type Layout struct {
	Left   int
	Top    int
	Width  int
	Height int
}

func clearLayoutSubtree(node *Node) {
	if node == nil {
		return
	}

	node.Layout = Layout{}
	for _, child := range node.Children {
		clearLayoutSubtree(child)
	}
}

func clearLayoutLineage(node *Node) {
	for current := node; current != nil; current = current.parent {
		current.Layout = Layout{}
	}
}

// markStaticAncestorsDirty walks parent links and flips staticDirty on any
// ancestor that is a `<Static>` element. Mirrors upstream's commit-update
// hook flipping rootNode.isStaticDirty whenever a node inside a Static
// subtree changes.
func (n *Node) markStaticAncestorsDirty() {
	for current := n; current != nil; current = current.parent {
		if current.Type == ElementNode && current.ElementType == "static" {
			current.staticDirty = true
			current.cachedStaticOutput = ""
		}
	}
}

// LookupTransformCache returns the cached transform output for (input, index)
// if one is recorded. The bool is false on cache miss. Mirrors upstream's
// "don't re-run the transform on unchanged subtrees" optimisation. Callers
// should pair this with StoreTransformCache to populate misses.
func (n *Node) LookupTransformCache(input string, index int) (string, bool) {
	if n == nil || n.transformCache == nil {
		return "", false
	}

	entry := n.transformCache
	if entry.index != index || entry.input != input {
		return "", false
	}

	return entry.output, true
}

// StoreTransformCache memoizes the (input, index) -> output result of a
// transform fn invocation. Subsequent calls with the same input/index return
// the cached output.
func (n *Node) StoreTransformCache(input string, index int, output string) {
	if n == nil {
		return
	}

	n.transformCache = &transformCacheEntry{
		input:  input,
		index:  index,
		output: output,
	}
}

// InvalidateTransformCache clears the memoized transform output. Callers
// invoke this when the transform function itself changes; prop / child
// mutations already invalidate via the layout-clearing path.
func (n *Node) InvalidateTransformCache() {
	if n == nil {
		return
	}
	n.transformCache = nil
}

// StaticDirty reports whether a `<Static>` node's cached output has been
// invalidated and must be recomputed. Returns true on a fresh node so the
// first render always computes the static output.
func (n *Node) StaticDirty() bool {
	if n == nil || n.Type != ElementNode || n.ElementType != "static" {
		return true
	}
	if n.cachedStaticOutput == "" {
		return true
	}
	return n.staticDirty
}

// LookupStaticOutput returns the cached static-block output if it matches
// (width, height, ansi). On any mismatch the entry is treated as a miss so a
// terminal resize or mode flip refreshes the cache.
func (n *Node) LookupStaticOutput(width, height int, ansi bool) (string, bool) {
	if n == nil || n.Type != ElementNode || n.ElementType != "static" {
		return "", false
	}
	if n.staticDirty || n.cachedStaticOutput == "" {
		return "", false
	}
	if n.cachedStaticWidth != width || n.cachedStaticHeight != height || n.cachedStaticANSI != ansi {
		return "", false
	}
	return n.cachedStaticOutput, true
}

// StoreStaticOutput stamps the rendered static-block string onto the node so
// subsequent renders with the same (width, height, ansi) skip the renderer.
func (n *Node) StoreStaticOutput(output string, width, height int, ansi bool) {
	if n == nil || n.Type != ElementNode || n.ElementType != "static" {
		return
	}

	n.cachedStaticOutput = output
	n.cachedStaticWidth = width
	n.cachedStaticHeight = height
	n.cachedStaticANSI = ansi
	n.staticDirty = false
}

// MarkStaticDirty forces the next render to recompute the static-block output.
// Mostly useful in tests; prop/child mutations already mark Static ancestors
// dirty via the standard Ink mutation hooks.
func (n *Node) MarkStaticDirty() {
	if n == nil || n.Type != ElementNode || n.ElementType != "static" {
		return
	}
	n.staticDirty = true
	n.cachedStaticOutput = ""
}

// CreateTextNode creates a new text node
func CreateTextNode(text string) *Node {
	return &Node{
		Type: TextNode,
		Text: text,
	}
}

// CreateElement creates a new element node with optional children
func CreateElement(elementType string, props Props, children ...*Node) *Node {
	if props == nil {
		props = Props{}
	}

	// Always initialize children array, even if empty
	childArray := children
	if childArray == nil {
		childArray = []*Node{}
	}

	node := &Node{
		Type:        ElementNode,
		ElementType: elementType,
		Props:       props,
		Children:    childArray,
	}

	if rawKey, ok := props["key"]; ok {
		node.Key = stringifyKey(rawKey)
	}

	for _, child := range childArray {
		if child != nil {
			child.parent = node
		}
	}

	return node
}

// stringifyKey normalizes the `key` prop into the canonical Node.Key
// string. Numeric keys (int / int64 / float64) are formatted with
// fmt.Sprintf("%v", ...) so React-style numeric child keys propagate
// correctly into the reconciler's keyed-children diff.
func stringifyKey(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

func detachChild(parent *Node, child *Node) bool {
	if parent == nil || child == nil || len(parent.Children) == 0 {
		return false
	}

	for index, existing := range parent.Children {
		if existing != child {
			continue
		}

		copy(parent.Children[index:], parent.Children[index+1:])
		parent.Children[len(parent.Children)-1] = nil
		parent.Children = parent.Children[:len(parent.Children)-1]
		child.parent = nil
		return true
	}

	return false
}

// Clone creates a deep copy of the node
func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}

	cloned := &Node{
		Type:        n.Type,
		ElementType: n.ElementType,
		Text:        n.Text,
		Key:         n.Key,
		Layout:      n.Layout,
	}

	// Deep copy props
	if n.Props != nil {
		cloned.Props = make(Props)
		for k, v := range n.Props {
			cloned.Props[k] = v
		}
	}

	// Deep copy children
	if n.Children != nil {
		cloned.Children = make([]*Node, len(n.Children))
		for i, child := range n.Children {
			if child == nil {
				continue
			}

			clonedChild := child.Clone()
			clonedChild.parent = cloned
			cloned.Children[i] = clonedChild
		}
	}

	return cloned
}

// AppendChild appends a child node, reparenting it if needed.
func (n *Node) AppendChild(child *Node) {
	if n == nil || child == nil {
		return
	}

	if child.parent != nil {
		oldParent := child.parent
		detachChild(oldParent, child)
		clearLayoutLineage(oldParent)
	}

	child.parent = n
	n.Children = append(n.Children, child)
	clearLayoutSubtree(child)
	clearLayoutLineage(n)
}

// InsertBefore inserts a child before another child or appends if the anchor is absent.
func (n *Node) InsertBefore(newChild *Node, beforeChild *Node) {
	if n == nil || newChild == nil {
		return
	}

	if newChild.parent != nil {
		oldParent := newChild.parent
		detachChild(oldParent, newChild)
		clearLayoutLineage(oldParent)
	}

	newChild.parent = n

	for index, child := range n.Children {
		if child != beforeChild {
			continue
		}

		n.Children = append(n.Children, nil)
		copy(n.Children[index+1:], n.Children[index:])
		n.Children[index] = newChild
		clearLayoutSubtree(newChild)
		clearLayoutLineage(n)
		return
	}

	n.Children = append(n.Children, newChild)
	clearLayoutSubtree(newChild)
	clearLayoutLineage(n)
}

// RemoveChild removes a child node and detaches its parent reference.
func (n *Node) RemoveChild(child *Node) bool {
	removed := detachChild(n, child)
	if !removed {
		return false
	}

	clearLayoutSubtree(child)
	clearLayoutLineage(n)
	return true
}

// SetAttribute sets an element attribute/prop.
func (n *Node) SetAttribute(key string, value interface{}) {
	if n == nil || n.Type != ElementNode {
		return
	}

	if n.Props == nil {
		n.Props = Props{}
	}

	n.Props[key] = value
	if key == "key" {
		n.Key = stringifyKey(value)
	}
	clearLayoutSubtree(n)
	clearLayoutLineage(n.parent)
	// Mirror upstream's commit-update hook flipping caches when a prop
	// changes: our transform memoization on this node is now stale, and
	// any Static ancestor needs to re-render.
	n.transformCache = nil
	n.markStaticAncestorsDirty()
}

// SetNodeValue updates the text value for a text node.
func (n *Node) SetNodeValue(text string) {
	if n == nil || n.Type != TextNode {
		return
	}

	n.Text = text
	clearLayoutLineage(n)
	// Walk up flipping caches: a text-content change invalidates the
	// transform memoization on the nearest text-like ancestor and any
	// Static ancestor.
	for current := n.parent; current != nil; current = current.parent {
		current.transformCache = nil
	}
	n.markStaticAncestorsDirty()
}

// ParentNode returns the node's parent in the current tree, if any.
func (n *Node) ParentNode() *Node {
	if n == nil {
		return nil
	}

	return n.parent
}

// ChildNodes returns the node's children.
func (n *Node) ChildNodes() []*Node {
	if n == nil {
		return nil
	}

	if n.Type != ElementNode {
		return nil
	}

	children := make([]*Node, 0, len(n.Children))
	for _, child := range n.Children {
		if child != nil {
			children = append(children, child)
		}
	}

	return children
}

// HasChildNodes reports whether the node has any non-nil children.
func (n *Node) HasChildNodes() bool {
	if n == nil {
		return false
	}

	for _, child := range n.Children {
		if child != nil {
			return true
		}
	}

	return false
}

// FirstChild returns the first non-nil child node.
func (n *Node) FirstChild() *Node {
	if n == nil {
		return nil
	}

	for _, child := range n.Children {
		if child != nil {
			return child
		}
	}

	return nil
}

// LastChild returns the last non-nil child node.
func (n *Node) LastChild() *Node {
	if n == nil {
		return nil
	}

	for index := len(n.Children) - 1; index >= 0; index-- {
		if n.Children[index] != nil {
			return n.Children[index]
		}
	}

	return nil
}

// GetAttribute returns the value stored at key in the node's attribute
// map (Props), filtered through the same exposure rules as Attributes().
// Returns (nil, false) when the key is missing, hidden by exposure
// rules, or the receiver is nil / not an element node.
func (n *Node) GetAttribute(key string) (interface{}, bool) {
	if n == nil || n.Type != ElementNode || n.Props == nil {
		return nil, false
	}
	if !isExposedAttributeKey(n.ElementType, key) {
		return nil, false
	}
	value, ok := n.Props[key]
	if !ok || !isExposedAttributeValue(value) {
		return nil, false
	}
	return value, true
}

// Style returns the layout/style props that Attributes() filters out:
// flex, padding, margin, width/height, border, color, etc. Mirrors
// upstream DOMElement.style. Returns nil for non-element nodes.
func (n *Node) Style() Props {
	if n == nil || n.Type != ElementNode {
		return nil
	}
	style := make(Props)
	for key, value := range n.Props {
		if isStylePropKey(key) {
			style[key] = value
		}
	}
	return style
}

// isStylePropKey reports whether key is a layout/style prop (the inverse
// of the user-attribute exposure rule). Keep in sync with the host-element
// suppress list in isExposedAttributeKey.
func isStylePropKey(key string) bool {
	switch key {
	case "textWrap", "wrap",
		"position", "columnGap", "rowGap", "gap",
		"margin", "marginX", "marginY", "marginTop", "marginBottom", "marginLeft", "marginRight",
		"padding", "paddingX", "paddingY", "paddingTop", "paddingBottom", "paddingLeft", "paddingRight",
		"flexGrow", "flexShrink", "flexDirection", "flexBasis", "flexWrap", "alignItems", "alignSelf", "justifyContent",
		"width", "height", "minWidth", "minHeight", "display",
		"borderStyle", "borderTop", "borderBottom", "borderLeft", "borderRight",
		"borderColor", "borderTopColor", "borderBottomColor", "borderLeftColor", "borderRightColor",
		"borderDimColor", "borderTopDimColor", "borderBottomDimColor", "borderLeftDimColor", "borderRightDimColor",
		"overflow", "overflowX", "overflowY",
		"backgroundColor", "color", "dimColor", "bold", "italic", "underline", "strikethrough", "inverse":
		return true
	}
	return false
}

// InternalStatic reports whether this node represents a `<static>`
// element or carries the explicit `internal_static` marker. Used by
// callers walking a tree to detect frozen subtrees.
func (n *Node) InternalStatic() bool {
	if n == nil || n.Type != ElementNode {
		return false
	}
	if n.ElementType == "static" {
		return true
	}
	if n.Props != nil {
		if value, ok := n.Props["internal_static"]; ok {
			if enabled, _ := value.(bool); enabled {
				return true
			}
		}
	}
	return false
}

// ElementChildren returns only the element-typed children, mirroring
// the DOM `Element.children` collection (skips text nodes). Named
// distinctly from the existing Children slice field to avoid collision.
func (n *Node) ElementChildren() []*Node {
	if n == nil {
		return nil
	}
	out := make([]*Node, 0, len(n.Children))
	for _, child := range n.Children {
		if child == nil || child.Type != ElementNode {
			continue
		}
		out = append(out, child)
	}
	return out
}

// OwnerRoot walks up parent pointers to the topmost ancestor and
// returns it. Returns the receiver itself when it has no parent.
func (n *Node) OwnerRoot() *Node {
	if n == nil {
		return nil
	}
	current := n
	for current.parent != nil {
		current = current.parent
	}
	return current
}

// Position returns the computed (left, top) layout coordinates if a
// computed layout is attached, otherwise (0, 0).
func (n *Node) Position() (left, top int) {
	if n == nil {
		return 0, 0
	}
	layout := n.ComputedLayout()
	return int(layout.Left), int(layout.Top)
}

// Attributes returns the node's attributes/props.
func (n *Node) Attributes() Props {
	if n == nil {
		return nil
	}

	if n.Type != ElementNode {
		return nil
	}

	attributes := make(Props)
	for key, value := range n.Props {
		if !isExposedAttributeKey(n.ElementType, key) || !isExposedAttributeValue(value) {
			continue
		}

		attributes[key] = value
	}

	return attributes
}

func isExposedAttributeKey(elementType string, key string) bool {
	switch key {
	case "ref", "internal_accessibility", "internal_transform", "internal_static", "transform", "static":
		return false
	}

	if strings.HasPrefix(key, "__") {
		return false
	}

	if !isInkHostElementType(elementType) {
		return true
	}

	switch key {
	case "aria-label", "aria-hidden", "aria-role", "aria-state",
		"textWrap", "wrap",
		"position", "columnGap", "rowGap", "gap",
		"margin", "marginX", "marginY", "marginTop", "marginBottom", "marginLeft", "marginRight",
		"padding", "paddingX", "paddingY", "paddingTop", "paddingBottom", "paddingLeft", "paddingRight",
		"flexGrow", "flexShrink", "flexDirection", "flexBasis", "flexWrap", "alignItems", "alignSelf", "justifyContent",
		"width", "height", "minWidth", "minHeight", "display",
		"borderStyle", "borderTop", "borderBottom", "borderLeft", "borderRight",
		"borderColor", "borderTopColor", "borderBottomColor", "borderLeftColor", "borderRightColor",
		"borderDimColor", "borderTopDimColor", "borderBottomDimColor", "borderLeftDimColor", "borderRightDimColor",
		"overflow", "overflowX", "overflowY",
		"backgroundColor", "color", "dimColor", "bold", "italic", "underline", "strikethrough", "inverse":
		return false
	default:
		return true
	}
}

func isExposedAttributeValue(value interface{}) bool {
	switch value.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64:
		return true
	default:
		return false
	}
}

func isInkHostElementType(elementType string) bool {
	switch elementType {
	case "root", "box", "text", "virtual-text":
		return true
	default:
		return false
	}
}

// NodeName returns the DOM-like node name for the current node.
func (n *Node) NodeName() string {
	if n == nil {
		return ""
	}

	if n.Type == TextNode {
		return "#text"
	}

	switch n.ElementType {
	case "root":
		return "ink-root"
	case "box":
		return "ink-box"
	case "text":
		return "ink-text"
	case "virtual-text":
		return "ink-virtual-text"
	default:
		return n.ElementType
	}
}

// NodeValue returns the text value for text nodes.
func (n *Node) NodeValue() string {
	if n == nil || n.Type != TextNode {
		return ""
	}

	return n.Text
}

// PreviousSibling returns the closest non-nil sibling before the current node.
func (n *Node) PreviousSibling() *Node {
	if n == nil || n.parent == nil {
		return nil
	}

	var previous *Node
	for _, sibling := range n.parent.Children {
		if sibling == n {
			return previous
		}
		if sibling != nil {
			previous = sibling
		}
	}

	return nil
}

// NextSibling returns the closest non-nil sibling after the current node.
func (n *Node) NextSibling() *Node {
	if n == nil || n.parent == nil {
		return nil
	}

	foundCurrent := false
	for _, sibling := range n.parent.Children {
		if !foundCurrent {
			if sibling == n {
				foundCurrent = true
			}
			continue
		}

		if sibling != nil {
			return sibling
		}
	}

	return nil
}

// TextContent returns the concatenated text content for a node and its descendants.
func (n *Node) TextContent() string {
	if n == nil {
		return ""
	}

	if n.Type == TextNode {
		return n.Text
	}

	var builder strings.Builder
	for _, child := range n.Children {
		if child == nil {
			continue
		}

		builder.WriteString(child.TextContent())
	}

	return builder.String()
}

// Contains reports whether the receiver is the same node as target or an ancestor of it.
func (n *Node) Contains(target *Node) bool {
	if n == nil || target == nil {
		return false
	}

	for current := target; current != nil; current = current.parent {
		if current == n {
			return true
		}
	}

	return false
}

// ComputedLayout returns the latest synced layout metadata for the node.
func (n *Node) ComputedLayout() Layout {
	if n == nil {
		return Layout{}
	}

	return n.Layout
}

// String returns string representation for debugging
func (n *Node) String() string {
	if n == nil {
		return "Node(nil)"
	}

	if n.Type == TextNode {
		return fmt.Sprintf("Text(%q)", n.Text)
	}
	return fmt.Sprintf("Element(%s, %d children)", n.ElementType, len(n.Children))
}
