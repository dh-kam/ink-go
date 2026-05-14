package vdom_test

import (
	"testing"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// TestNodeType tests basic node type definitions
func TestNodeType(t *testing.T) {
	tests := []struct {
		name     string
		nodeType vdom.NodeType
		want     string
	}{
		{"Text node", vdom.TextNode, "text"},
		{"Element node", vdom.ElementNode, "element"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.nodeType.String(); got != tt.want {
				t.Errorf("NodeType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCreateTextNode tests creating a text node
func TestCreateTextNode(t *testing.T) {
	text := "Hello, World!"
	node := vdom.CreateTextNode(text)

	if node.Type != vdom.TextNode {
		t.Errorf("Expected type TextNode, got %v", node.Type)
	}

	if node.Text != text {
		t.Errorf("Expected text %q, got %q", text, node.Text)
	}

	if node.Children != nil {
		t.Error("Text node should not have children")
	}
}

// TestCreateElement tests creating an element node
func TestCreateElement(t *testing.T) {
	elementType := "box"
	props := vdom.Props{
		"width": 10,
		"color": "red",
	}

	node := vdom.CreateElement(elementType, props)

	if node.Type != vdom.ElementNode {
		t.Errorf("Expected type ElementNode, got %v", node.Type)
	}

	if node.ElementType != elementType {
		t.Errorf("Expected elementType %q, got %q", elementType, node.ElementType)
	}

	if node.Props["width"] != 10 {
		t.Errorf("Expected width 10, got %v", node.Props["width"])
	}

	if node.Props["color"] != "red" {
		t.Errorf("Expected color red, got %v", node.Props["color"])
	}

	if node.Children == nil {
		t.Error("Element node should have children array (even if empty)")
	}
}

// TestCreateElementWithChildren tests creating element with children
func TestCreateElementWithChildren(t *testing.T) {
	child1 := vdom.CreateTextNode("Child 1")
	child2 := vdom.CreateTextNode("Child 2")

	parent := vdom.CreateElement("box", vdom.Props{}, child1, child2)

	if len(parent.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(parent.Children))
	}

	if parent.Children[0].Text != "Child 1" {
		t.Errorf("Expected first child text 'Child 1', got %q", parent.Children[0].Text)
	}

	if parent.Children[1].Text != "Child 2" {
		t.Errorf("Expected second child text 'Child 2', got %q", parent.Children[1].Text)
	}
}

// TestNodeClone tests cloning a node
func TestNodeClone(t *testing.T) {
	original := vdom.CreateElement("box", vdom.Props{"width": 10})
	child := vdom.CreateTextNode("test")
	original.Children = append(original.Children, child)

	cloned := original.Clone()

	// Check that values are equal
	if cloned.ElementType != original.ElementType {
		t.Error("Cloned node should have same element type")
	}

	if cloned.Props["width"] != original.Props["width"] {
		t.Error("Cloned node should have same props")
	}

	// Check that it's a deep copy
	cloned.Props["width"] = 20
	if original.Props["width"] == 20 {
		t.Error("Modifying clone should not affect original")
	}
}

// TestCloneTextNode tests cloning a text node
func TestCloneTextNode(t *testing.T) {
	original := vdom.CreateTextNode("hello world")
	cloned := original.Clone()

	if cloned.Type != original.Type {
		t.Error("Cloned text node should have same type")
	}
	if cloned.Text != original.Text {
		t.Error("Cloned text node should have same text")
	}
	if cloned.Text != "hello world" {
		t.Errorf("Expected text 'hello world', got %q", cloned.Text)
	}
}

// TestCloneWithKey tests cloning a node with a key
func TestCloneWithKey(t *testing.T) {
	original := vdom.CreateElement("item", vdom.Props{})
	original.Key = "unique-key"

	cloned := original.Clone()

	if cloned.Key != original.Key {
		t.Errorf("Expected key %q, got %q", original.Key, cloned.Key)
	}
}

// TestCloneDeepCopyChildren tests that cloning creates deep copy of children
func TestCloneDeepCopyChildren(t *testing.T) {
	child1 := vdom.CreateTextNode("child1")
	child2 := vdom.CreateTextNode("child2")

	original := vdom.CreateElement("container", vdom.Props{})
	original.Children = append(original.Children, child1, child2)

	cloned := original.Clone()

	// Modify cloned child's text
	cloned.Children[0].Text = "modified"

	// Original should be unchanged
	if original.Children[0].Text == "modified" {
		t.Error("Modifying cloned child should not affect original child")
	}
	if original.Children[0].Text != "child1" {
		t.Errorf("Expected original child text 'child1', got %q", original.Children[0].Text)
	}
}

// TestNodeString tests String representation
func TestNodeString(t *testing.T) {
	textNode := vdom.CreateTextNode("test text")
	if textNode.String() != `Text("test text")` {
		t.Errorf("Expected Text(\"test text\"), got %q", textNode.String())
	}

	elementNode := vdom.CreateElement("box", vdom.Props{})
	str := elementNode.String()
	// Should contain element type and children count
	if str != `Element(box, 0 children)` {
		t.Errorf("Expected Element(box, 0 children), got %q", str)
	}

	elementWithChildren := vdom.CreateElement("container", vdom.Props{},
		vdom.CreateTextNode("a"),
		vdom.CreateTextNode("b"),
	)
	if elementWithChildren.String() != `Element(container, 2 children)` {
		t.Errorf("Expected Element(container, 2 children), got %q", elementWithChildren.String())
	}
}

// TestNodeTypeInvalid tests String() with invalid node type
func TestNodeTypeInvalid(t *testing.T) {
	invalidType := vdom.NodeType(99)
	if invalidType.String() != "unknown" {
		t.Errorf("Expected 'unknown' for invalid type, got %q", invalidType.String())
	}
}

// TestCreateElementWithNilProps tests creating element with nil props
func TestCreateElementWithNilProps(t *testing.T) {
	node := vdom.CreateElement("box", nil)

	if node.Props == nil {
		t.Error("Props should be initialized to empty map, not nil")
	}
	if len(node.Props) != 0 {
		t.Errorf("Expected empty props, got %d items", len(node.Props))
	}
}

// TestCreateElementWithEmptyProps tests creating element with empty props
func TestCreateElementWithEmptyProps(t *testing.T) {
	node := vdom.CreateElement("box", vdom.Props{})

	if node.Props == nil {
		t.Error("Props should not be nil")
	}
	if len(node.Props) != 0 {
		t.Errorf("Expected empty props, got %d items", len(node.Props))
	}
}

// TestNodeKeyField tests node Key field
func TestNodeKeyField(t *testing.T) {
	node := vdom.CreateElement("item", vdom.Props{})
	node.Key = "test-key"

	if node.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got %q", node.Key)
	}
}

func TestCreateElementPropagatesKeyProp(t *testing.T) {
	node := vdom.CreateElement("item", vdom.Props{"key": "test-key"})
	if node.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got %q", node.Key)
	}
	if node.Props["key"] != "test-key" {
		t.Errorf("Expected key prop to remain available, got %v", node.Props["key"])
	}
}

func TestCreateElementPropagatesNumericKeyProp(t *testing.T) {
	node := vdom.CreateElement("item", vdom.Props{"key": 42})
	if node.Key != "42" {
		t.Errorf("Expected numeric key '42', got %q", node.Key)
	}
}

func TestSetAttributeUpdatesKey(t *testing.T) {
	node := vdom.CreateElement("item", nil)
	node.SetAttribute("key", "updated-key")
	if node.Key != "updated-key" {
		t.Errorf("Expected key 'updated-key', got %q", node.Key)
	}
}

// TestTextNoChildren tests that text nodes don't have children
func TestTextNoChildren(t *testing.T) {
	node := vdom.CreateTextNode("test")

	// Text nodes should have nil children
	if node.Children != nil {
		t.Error("Text node should have nil children")
	}
}

// TestElementHasChildrenArray tests elements always have children array
func TestElementHasChildrenArray(t *testing.T) {
	node := vdom.CreateElement("box", vdom.Props{})

	if node.Children == nil {
		t.Error("Element node should have children array (even if empty)")
	}
	if len(node.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(node.Children))
	}
}

// TestCloneEmptyProps tests cloning node with empty props
func TestCloneEmptyProps(t *testing.T) {
	original := vdom.CreateElement("box", vdom.Props{})
	cloned := original.Clone()

	if cloned.Props == nil {
		t.Error("Cloned node should have props map")
	}
	if len(cloned.Props) != 0 {
		t.Errorf("Expected empty props, got %d items", len(cloned.Props))
	}
}

// TestCloneNilProps tests cloning node with nil props
func TestCloneNilProps(t *testing.T) {
	original := vdom.CreateElement("box", nil)
	cloned := original.Clone()

	if cloned.Props == nil {
		t.Error("Cloned node should have props map")
	}
}

// TestCloneEmptyChildren tests cloning node with no children
func TestCloneEmptyChildren(t *testing.T) {
	original := vdom.CreateElement("box", vdom.Props{})
	cloned := original.Clone()

	if cloned.Children == nil {
		t.Error("Cloned node should have children array")
	}
	if len(cloned.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(cloned.Children))
	}
}

// TestNestedElementChildren tests deeply nested element structure
func TestNestedElementChildren(t *testing.T) {
	inner := vdom.CreateTextNode("inner text")
	middle := vdom.CreateElement("middle", vdom.Props{"id": 1}, inner)
	outer := vdom.CreateElement("outer", vdom.Props{}, middle)

	if len(outer.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(outer.Children))
	}

	if outer.Children[0].ElementType != "middle" {
		t.Errorf("Expected 'middle' element, got %q", outer.Children[0].ElementType)
	}

	if len(outer.Children[0].Children) != 1 {
		t.Fatalf("Expected 1 grandchild, got %d", len(outer.Children[0].Children))
	}

	if outer.Children[0].Children[0].Text != "inner text" {
		t.Errorf("Expected 'inner text', got %q", outer.Children[0].Children[0].Text)
	}
}

// TestNodeTypeValues tests NodeType constant values
func TestNodeTypeValues(t *testing.T) {
	if vdom.TextNode != 0 {
		t.Errorf("Expected TextNode=0, got %d", vdom.TextNode)
	}
	if vdom.ElementNode != 1 {
		t.Errorf("Expected ElementNode=1, got %d", vdom.ElementNode)
	}
}

// TestPropsType tests Props type
func TestPropsType(t *testing.T) {
	props := vdom.Props{
		"string": "value",
		"number": 42,
		"bool":   true,
		"nil":    nil,
	}

	if props["string"] != "value" {
		t.Error("String prop mismatch")
	}
	if props["number"] != 42 {
		t.Error("Number prop mismatch")
	}
	if props["bool"] != true {
		t.Error("Bool prop mismatch")
	}
	if props["nil"] != nil {
		t.Error("Nil prop mismatch")
	}
}

func TestNodeDOMLikeAccessorsExposeTreeStructure(t *testing.T) {
	textChild := vdom.CreateTextNode("prefix")
	nestedText := vdom.CreateTextNode("suffix")
	nested := vdom.CreateElement("text", vdom.Props{"color": "green"}, nestedText)
	root := vdom.CreateElement("box", vdom.Props{"data-id": "root", "width": 10}, textChild, nested)

	if got := root.NodeName(); got != "ink-box" {
		t.Fatalf("expected root node name %q, got %q", "ink-box", got)
	}
	if got := textChild.NodeName(); got != "#text" {
		t.Fatalf("expected text node name %q, got %q", "#text", got)
	}
	if got := nested.NodeName(); got != "ink-text" {
		t.Fatalf("expected nested node name %q, got %q", "ink-text", got)
	}
	if got := textChild.NodeValue(); got != "prefix" {
		t.Fatalf("expected text node value %q, got %q", "prefix", got)
	}
	if got := root.Attributes()["data-id"]; got != "root" {
		t.Fatalf("expected data-id attribute %q, got %v", "root", got)
	}
	if _, ok := root.Attributes()["width"]; ok {
		t.Fatal("expected style width prop to be hidden from DOM attributes")
	}
	if got := len(root.ChildNodes()); got != 2 {
		t.Fatalf("expected 2 child nodes, got %d", got)
	}
	if textChild.ParentNode() != root {
		t.Fatal("expected text child parent to be the root element")
	}
	if nested.ParentNode() != root {
		t.Fatal("expected nested element parent to be the root element")
	}
	if nestedText.ParentNode() != nested {
		t.Fatal("expected nested text parent to be the nested element")
	}
	if got := root.TextContent(); got != "prefixsuffix" {
		t.Fatalf("expected text content %q, got %q", "prefixsuffix", got)
	}
	if got := nested.TextContent(); got != "suffix" {
		t.Fatalf("expected nested text content %q, got %q", "suffix", got)
	}
}

func TestChildNodesSkipsNilPlaceholders(t *testing.T) {
	left := vdom.CreateTextNode("left")
	right := vdom.CreateTextNode("right")
	root := vdom.CreateElement("box", nil, left, nil, right)

	children := root.ChildNodes()
	if len(children) != 2 {
		t.Fatalf("expected 2 non-nil child nodes, got %d", len(children))
	}
	if children[0] != left || children[1] != right {
		t.Fatalf("expected child nodes [left right], got %#v", children)
	}
}

func TestNodeNameUsesInkHostNamesForKnownElements(t *testing.T) {
	tests := []struct {
		elementType string
		want        string
	}{
		{elementType: "root", want: "ink-root"},
		{elementType: "box", want: "ink-box"},
		{elementType: "text", want: "ink-text"},
		{elementType: "virtual-text", want: "ink-virtual-text"},
		{elementType: "table", want: "table"},
	}

	for _, tt := range tests {
		t.Run(tt.elementType, func(t *testing.T) {
			node := vdom.CreateElement(tt.elementType, nil)
			if got := node.NodeName(); got != tt.want {
				t.Fatalf("expected node name %q, got %q", tt.want, got)
			}
		})
	}
}

func TestAttributesHideInternalProps(t *testing.T) {
	node := vdom.CreateElement("box", vdom.Props{
		"data-id":                "visible",
		"width":                  12,
		"color":                  "green",
		"ref":                    "hidden",
		"internal_accessibility": "hidden",
	})

	attributes := node.Attributes()
	if got := attributes["data-id"]; got != "visible" {
		t.Fatalf("expected data-id attribute %q, got %v", "visible", got)
	}
	if _, ok := attributes["width"]; ok {
		t.Fatal("expected width style prop to be hidden from DOM attributes")
	}
	if _, ok := attributes["color"]; ok {
		t.Fatal("expected color style prop to be hidden from DOM attributes")
	}
	if _, ok := attributes["ref"]; ok {
		t.Fatal("expected ref prop to be hidden from DOM attributes")
	}
	if _, ok := attributes["internal_accessibility"]; ok {
		t.Fatal("expected internal accessibility prop to be hidden from DOM attributes")
	}
}

func TestAttributesHideUnsupportedCustomValues(t *testing.T) {
	node := vdom.CreateElement("box", vdom.Props{
		"data-id":       "visible",
		"data-count":    2,
		"data-enabled":  true,
		"data-list":     []string{"hidden"},
		"data-function": func() {},
		"data-nil":      nil,
	})

	attributes := node.Attributes()
	if got := attributes["data-id"]; got != "visible" {
		t.Fatalf("expected data-id attribute %q, got %v", "visible", got)
	}
	if got := attributes["data-count"]; got != 2 {
		t.Fatalf("expected data-count attribute %d, got %v", 2, got)
	}
	if got := attributes["data-enabled"]; got != true {
		t.Fatalf("expected data-enabled attribute %t, got %v", true, got)
	}
	if _, ok := attributes["data-list"]; ok {
		t.Fatal("expected slice-valued prop to be hidden from DOM attributes")
	}
	if _, ok := attributes["data-function"]; ok {
		t.Fatal("expected function-valued prop to be hidden from DOM attributes")
	}
	if _, ok := attributes["data-nil"]; ok {
		t.Fatal("expected nil-valued prop to be hidden from DOM attributes")
	}
}

func TestClonePreservesLayoutAndReparentsChildren(t *testing.T) {
	child := vdom.CreateTextNode("child")
	child.Layout = vdom.Layout{Width: 5, Height: 1}

	original := vdom.CreateElement("box", vdom.Props{"width": 10}, child)
	original.Layout = vdom.Layout{Left: 1, Top: 2, Width: 10, Height: 3}

	cloned := original.Clone()

	if cloned.ParentNode() != nil {
		t.Fatal("expected cloned root to have no parent")
	}
	if got := cloned.ComputedLayout(); got != original.Layout {
		t.Fatalf("expected cloned layout %+v, got %+v", original.Layout, got)
	}
	if len(cloned.ChildNodes()) != 1 {
		t.Fatalf("expected one cloned child, got %d", len(cloned.ChildNodes()))
	}
	if cloned.ChildNodes()[0] == original.ChildNodes()[0] {
		t.Fatal("expected cloned child to be a deep copy")
	}
	if cloned.ChildNodes()[0].ParentNode() != cloned {
		t.Fatal("expected cloned child parent to be the cloned root")
	}
	if original.ChildNodes()[0].ParentNode() != original {
		t.Fatal("expected original child parent to remain the original root")
	}
	if got := cloned.ChildNodes()[0].ComputedLayout(); got != child.Layout {
		t.Fatalf("expected cloned child layout %+v, got %+v", child.Layout, got)
	}
}

func TestAppendChildReparentsExistingChild(t *testing.T) {
	child := vdom.CreateTextNode("child")
	firstParent := vdom.CreateElement("box", nil, child)
	secondParent := vdom.CreateElement("box", nil)
	firstParent.Layout = vdom.Layout{Width: 4, Height: 1}
	secondParent.Layout = vdom.Layout{Width: 2, Height: 1}
	child.Layout = vdom.Layout{Width: 5, Height: 1}

	secondParent.AppendChild(child)

	if got := len(firstParent.ChildNodes()); got != 0 {
		t.Fatalf("expected first parent to lose the child, got %d children", got)
	}
	if got := len(secondParent.ChildNodes()); got != 1 {
		t.Fatalf("expected second parent to gain the child, got %d children", got)
	}
	if child.ParentNode() != secondParent {
		t.Fatal("expected child parent to be updated to the second parent")
	}
	if got := firstParent.ComputedLayout(); got != (vdom.Layout{}) {
		t.Fatalf("expected old parent layout to be invalidated, got %+v", got)
	}
	if got := secondParent.ComputedLayout(); got != (vdom.Layout{}) {
		t.Fatalf("expected new parent layout to be invalidated, got %+v", got)
	}
	if got := child.ComputedLayout(); got != (vdom.Layout{}) {
		t.Fatalf("expected moved child layout to be invalidated, got %+v", got)
	}
}

func TestInsertBeforePlacesChildAheadOfAnchor(t *testing.T) {
	left := vdom.CreateTextNode("left")
	right := vdom.CreateTextNode("right")
	inserted := vdom.CreateTextNode("inserted")
	parent := vdom.CreateElement("box", nil, left, right)

	parent.InsertBefore(inserted, right)

	if got := len(parent.ChildNodes()); got != 3 {
		t.Fatalf("expected 3 children, got %d", got)
	}
	if parent.ChildNodes()[0] != left || parent.ChildNodes()[1] != inserted || parent.ChildNodes()[2] != right {
		t.Fatalf("expected insert ordering [left inserted right], got %#v", parent.ChildNodes())
	}
	if inserted.ParentNode() != parent {
		t.Fatal("expected inserted child parent to be updated")
	}
}

func TestInsertBeforeAppendsWhenAnchorMissing(t *testing.T) {
	left := vdom.CreateTextNode("left")
	inserted := vdom.CreateTextNode("inserted")
	missing := vdom.CreateTextNode("missing")
	parent := vdom.CreateElement("box", nil, left)

	parent.InsertBefore(inserted, missing)

	if got := len(parent.ChildNodes()); got != 2 {
		t.Fatalf("expected 2 children, got %d", got)
	}
	if parent.ChildNodes()[1] != inserted {
		t.Fatal("expected insert-before with missing anchor to append the child")
	}
}

func TestRemoveChildDetachesParentAndReturnsStatus(t *testing.T) {
	child := vdom.CreateTextNode("child")
	parent := vdom.CreateElement("box", nil, child)

	if removed := parent.RemoveChild(child); !removed {
		t.Fatal("expected child removal to report success")
	}
	if removed := parent.RemoveChild(child); removed {
		t.Fatal("expected removing an absent child to report failure")
	}
	if got := len(parent.ChildNodes()); got != 0 {
		t.Fatalf("expected no children after removal, got %d", got)
	}
	if child.ParentNode() != nil {
		t.Fatal("expected removed child parent to be cleared")
	}
}

func TestSetAttributeInitializesPropsMap(t *testing.T) {
	child := vdom.CreateTextNode("child")
	node := vdom.CreateElement("box", nil, child)
	node.Layout = vdom.Layout{Width: 4, Height: 1}
	child.Layout = vdom.Layout{Width: 5, Height: 1}

	node.SetAttribute("data-id", "box-1")
	node.SetAttribute("title", "green")

	if got := node.Attributes()["data-id"]; got != "box-1" {
		t.Fatalf("expected data-id attribute %q, got %v", "box-1", got)
	}
	if got := node.Attributes()["title"]; got != "green" {
		t.Fatalf("expected title attribute %q, got %v", "green", got)
	}
	if got := node.ComputedLayout(); got != (vdom.Layout{}) {
		t.Fatalf("expected node layout to be invalidated, got %+v", got)
	}
	if got := child.ComputedLayout(); got != (vdom.Layout{}) {
		t.Fatalf("expected descendant layout to be invalidated, got %+v", got)
	}
}

func TestSetNodeValueUpdatesTextContent(t *testing.T) {
	node := vdom.CreateTextNode("before")
	parent := vdom.CreateElement("text", nil, node)
	parent.Layout = vdom.Layout{Width: 6, Height: 1}
	node.Layout = vdom.Layout{Width: 6, Height: 1}

	node.SetNodeValue("after")

	if got := node.NodeValue(); got != "after" {
		t.Fatalf("expected node value %q, got %q", "after", got)
	}
	if got := node.TextContent(); got != "after" {
		t.Fatalf("expected text content %q, got %q", "after", got)
	}
	if got := node.ComputedLayout(); got != (vdom.Layout{}) {
		t.Fatalf("expected text node layout to be invalidated, got %+v", got)
	}
	if got := parent.ComputedLayout(); got != (vdom.Layout{}) {
		t.Fatalf("expected ancestor layout to be invalidated, got %+v", got)
	}
}

func TestNodeDOMNavigationHelpers(t *testing.T) {
	left := vdom.CreateTextNode("left")
	middle := vdom.CreateTextNode("middle")
	right := vdom.CreateTextNode("right")
	root := vdom.CreateElement("box", nil, left, nil, middle, nil, right)

	if !root.HasChildNodes() {
		t.Fatal("expected root to report child nodes")
	}
	if got := root.FirstChild(); got != left {
		t.Fatalf("expected first child %v, got %v", left, got)
	}
	if got := root.LastChild(); got != right {
		t.Fatalf("expected last child %v, got %v", right, got)
	}
	if got := middle.PreviousSibling(); got != left {
		t.Fatalf("expected previous sibling %v, got %v", left, got)
	}
	if got := middle.NextSibling(); got != right {
		t.Fatalf("expected next sibling %v, got %v", right, got)
	}
	if got := left.PreviousSibling(); got != nil {
		t.Fatalf("expected first child to have no previous sibling, got %v", got)
	}
	if got := right.NextSibling(); got != nil {
		t.Fatalf("expected last child to have no next sibling, got %v", got)
	}
	if !root.Contains(root) {
		t.Fatal("expected node to contain itself")
	}
	if !root.Contains(right) {
		t.Fatal("expected root to contain a descendant")
	}
	if middle.Contains(root) {
		t.Fatal("expected descendant not to contain its ancestor")
	}

	var nilNode *vdom.Node
	if nilNode.HasChildNodes() {
		t.Fatal("expected nil node not to report children")
	}
	if nilNode.FirstChild() != nil || nilNode.LastChild() != nil {
		t.Fatal("expected nil node child helpers to return nil")
	}
	if nilNode.PreviousSibling() != nil || nilNode.NextSibling() != nil {
		t.Fatal("expected nil node sibling helpers to return nil")
	}
	if nilNode.Contains(root) {
		t.Fatal("expected nil node not to contain any other node")
	}
}
