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

	for _, child := range childArray {
		if child != nil {
			child.parent = node
		}
	}

	return node
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
	clearLayoutSubtree(n)
	clearLayoutLineage(n.parent)
}

// SetNodeValue updates the text value for a text node.
func (n *Node) SetNodeValue(text string) {
	if n == nil || n.Type != TextNode {
		return
	}

	n.Text = text
	clearLayoutLineage(n)
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
