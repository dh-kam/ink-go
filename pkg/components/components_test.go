package components_test

import (
	"fmt"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func TestBox(t *testing.T) {
	child := vdom.CreateTextNode("test")
	box := components.Box(nil, child)

	if box.Type != vdom.ElementNode {
		t.Error("Box should create an element node")
	}

	if box.ElementType != "box" {
		t.Errorf("Expected element type 'box', got %q", box.ElementType)
	}

	if len(box.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(box.Children))
	}
}

func TestText(t *testing.T) {
	text := components.Text("Hello")

	if text.Type != vdom.ElementNode {
		t.Error("Text should create an element node")
	}

	if text.ElementType != "text" {
		t.Errorf("Expected element type 'text', got %q", text.ElementType)
	}

	if len(text.Children) != 1 || text.Children[0].Text != "Hello" {
		t.Error("Text should wrap a text node")
	}
}

func TestTextWithPropsAndChildren(t *testing.T) {
	text := components.Text(
		vdom.Props{"color": "green"},
		vdom.CreateTextNode("Hello"),
		" ",
		vdom.CreateElement("text", nil, vdom.CreateTextNode("World")),
	)

	if text.Type != vdom.ElementNode {
		t.Error("Text should create an element node")
	}

	if text.Props["color"] != "green" {
		t.Errorf("Expected text color prop to be preserved, got %v", text.Props["color"])
	}

	if len(text.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(text.Children))
	}

	if text.Children[0].Text != "Hello" {
		t.Errorf("Expected first child text 'Hello', got %q", text.Children[0].Text)
	}

	if text.Children[1].Text != " " {
		t.Errorf("Expected second child text ' ', got %q", text.Children[1].Text)
	}

	if text.Children[2].ElementType != "text" {
		t.Errorf("Expected third child to be nested text, got %q", text.Children[2].ElementType)
	}
}

func TestNewline(t *testing.T) {
	nl := components.Newline()

	if nl.Type != vdom.TextNode {
		t.Error("Newline should create a text node")
	}

	if nl.Text != "\n" {
		t.Errorf("Expected newline character, got %q", nl.Text)
	}
}

func TestNewlineCount(t *testing.T) {
	nl := components.Newline(3)

	if nl.Type != vdom.TextNode {
		t.Error("Newline should create a text node")
	}

	if nl.Text != "\n\n\n" {
		t.Errorf("Expected three newline characters, got %q", nl.Text)
	}
}

func TestSpace(t *testing.T) {
	space := components.Space()

	if space.Type != vdom.TextNode {
		t.Error("Space should create a text node")
	}

	if space.Text != " " {
		t.Errorf("Expected space character, got %q", space.Text)
	}
}

func TestSpacer(t *testing.T) {
	spacer := components.Spacer()

	if spacer.Type != vdom.ElementNode {
		t.Error("Spacer should create an element node")
	}

	if spacer.ElementType != "box" {
		t.Errorf("Expected spacer to be a box, got %q", spacer.ElementType)
	}

	if spacer.Props["flexGrow"] != 1.0 {
		t.Errorf("Expected spacer flexGrow=1.0, got %v", spacer.Props["flexGrow"])
	}
}

func TestTransform(t *testing.T) {
	transform := components.Transform(func(children string, index int) string {
		return children
	}, vdom.CreateTextNode("hello"))

	if transform.Type != vdom.ElementNode {
		t.Error("Transform should create an element node")
	}

	if transform.ElementType != "transform" {
		t.Errorf("Expected element type 'transform', got %q", transform.ElementType)
	}

	if _, ok := transform.Props["transform"].(func(string, int) string); !ok {
		t.Error("Transform function should be stored in props")
	}
}

// TestBorder tests Border component
func TestBorder(t *testing.T) {
	borderProps := components.BorderProps{
		Style:  components.BorderSingle,
		Top:    true,
		Bottom: true,
		Left:   true,
		Right:  true,
	}

	child := vdom.CreateTextNode("content")
	border := components.Border(borderProps, nil, child)

	if border.Type != vdom.ElementNode {
		t.Error("Border should create an element node")
	}

	if border.ElementType != "border" {
		t.Errorf("Expected element type 'border', got %q", border.ElementType)
	}

	// Check props
	if border.Props["borderStyle"] != string(components.BorderSingle) {
		t.Error("Border style prop not set correctly")
	}

	if border.Props["borderTop"] != true {
		t.Error("Border top prop not set correctly")
	}
}

// TestBorderWithLabel tests Border component with label
func TestBorderWithLabel(t *testing.T) {
	borderProps := components.BorderProps{
		Style: components.BorderDouble,
		Label: "My Border",
	}

	border := components.Border(borderProps, nil)

	if border.Props["borderLabel"] != "My Border" {
		t.Errorf("Expected label 'My Border', got %v", border.Props["borderLabel"])
	}
}

// TestBorderWithTitle tests Border component with title (alias for label)
func TestBorderWithTitle(t *testing.T) {
	borderProps := components.BorderProps{
		Style: components.BorderRounded,
		Title: "Title Text",
	}

	border := components.Border(borderProps, nil)

	if border.Props["borderLabel"] != "Title Text" {
		t.Errorf("Expected title 'Title Text' as label, got %v", border.Props["borderLabel"])
	}
}

// TestBorderStyleConstants tests border style constants
func TestBorderStyleConstants(t *testing.T) {
	tests := []struct {
		style    components.BorderStyle
		expected string
	}{
		{components.BorderSingle, "single"},
		{components.BorderDouble, "double"},
		{components.BorderRounded, "rounded"},
		{components.BorderBold, "bold"},
	}

	for _, tt := range tests {
		if string(tt.style) != tt.expected {
			t.Errorf("Expected %q, got %q", tt.expected, tt.style)
		}
	}
}

// TestStatic tests Static component
func TestStatic(t *testing.T) {
	child := vdom.CreateTextNode("static content")
	static := components.Static(nil, child)

	if static.Type != vdom.ElementNode {
		t.Error("Static should create an element node")
	}

	if static.ElementType != "static" {
		t.Errorf("Expected element type 'static', got %q", static.ElementType)
	}

	if static.Props["static"] != true {
		t.Error("Static prop should be set to true")
	}
}

func TestStaticItems(t *testing.T) {
	static := components.Static(
		[]string{"A", "B"},
		func(item string, index int) *vdom.Node {
			return components.Text("[", item, "]", vdom.Props{"index": index})
		},
		vdom.Props{"paddingBottom": 1.0},
	)

	if static.Type != vdom.ElementNode {
		t.Error("Static should create an element node")
	}

	if static.ElementType != "static" {
		t.Errorf("Expected element type 'static', got %q", static.ElementType)
	}

	if static.Props["paddingBottom"] != 1.0 {
		t.Errorf("Expected custom props to be preserved, got %v", static.Props["paddingBottom"])
	}

	if len(static.Children) != 2 {
		t.Fatalf("Expected 2 static children, got %d", len(static.Children))
	}

	if static.Children[0].ElementType != "text" {
		t.Fatalf("Expected rendered static child to be text, got %q", static.Children[0].ElementType)
	}
}

func TestStaticItemsTypedHelper(t *testing.T) {
	static := components.StaticItems([]int{1, 2, 3}, func(item int, index int) *vdom.Node {
		return components.Text(vdom.Props{"index": index}, fmt.Sprintf("%d", item))
	})

	if static.ElementType != "static" {
		t.Fatalf("expected static element, got %q", static.ElementType)
	}

	if len(static.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(static.Children))
	}

	if static.Children[2].Children[0].Text != "3" {
		t.Fatalf("expected last child text '3', got %q", static.Children[2].Children[0].Text)
	}
}

// TestStaticWithProps tests Static component with custom props
func TestStaticWithProps(t *testing.T) {
	props := vdom.Props{"id": "my-static"}
	child := vdom.CreateTextNode("content")
	static := components.Static(props, child)

	if static.Props["id"] != "my-static" {
		t.Error("Custom props should be preserved")
	}

	if static.Props["static"] != true {
		t.Error("Static prop should be set to true")
	}
}

// TestStaticText tests StaticText component
func TestStaticText(t *testing.T) {
	staticText := components.StaticText("Hello, World!")

	if staticText.Type != vdom.ElementNode {
		t.Error("StaticText should create an element node")
	}

	if staticText.ElementType != "text" {
		t.Errorf("Expected element type 'text', got %q", staticText.ElementType)
	}

	if staticText.Props["static"] != true {
		t.Error("Static prop should be set to true")
	}

	if len(staticText.Children) != 1 {
		t.Error("StaticText should have one child")
	}

	if staticText.Children[0].Text != "Hello, World!" {
		t.Errorf("Expected text 'Hello, World!', got %q", staticText.Children[0].Text)
	}
}
