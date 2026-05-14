package components_test

import (
	"fmt"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
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

func TestTextConvertsNumericChildren(t *testing.T) {
	text := components.Text("Value: ", 42)

	if text == nil {
		t.Fatal("Text should render numeric children")
	}

	if len(text.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(text.Children))
	}

	if text.Children[1].Text != "42" {
		t.Errorf("Expected numeric child to render as %q, got %q", "42", text.Children[1].Text)
	}
}

func TestTextWithoutChildrenReturnsNil(t *testing.T) {
	if text := components.Text(nil); text != nil {
		t.Fatalf("Expected nil text node for nil children, got %#v", text)
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

	if nl.Type != vdom.ElementNode {
		t.Error("Newline should create a text element")
	}

	if nl.ElementType != "text" {
		t.Errorf("Expected newline element type %q, got %q", "text", nl.ElementType)
	}

	if len(nl.Children) != 1 || nl.Children[0].Text != "\n" {
		t.Errorf("Expected newline character, got %#v", nl.Children)
	}
}

func TestNewlineCount(t *testing.T) {
	nl := components.Newline(3)

	if nl.Type != vdom.ElementNode {
		t.Error("Newline should create a text element")
	}

	if len(nl.Children) != 1 || nl.Children[0].Text != "\n\n\n" {
		t.Errorf("Expected three newline characters, got %#v", nl.Children)
	}
}

func TestNewlineZeroCount(t *testing.T) {
	nl := components.Newline(0)

	if nl.Type != vdom.ElementNode {
		t.Error("Newline should create a text element")
	}

	if len(nl.Children) != 1 || nl.Children[0].Text != "" {
		t.Errorf("Expected no newline characters, got %#v", nl.Children)
	}
}

// TestNewlineNegativeCount verifies that a negative count clamps to zero —
// upstream Ink calls '\n'.repeat(count) which would throw RangeError for a
// negative value, but goink's variadic int signature can't surface that
// cleanly in Go callers, so we treat negative counts as a no-op (empty
// payload), matching the count=0 case.
func TestNewlineNegativeCount(t *testing.T) {
	cases := []int{-1, -3, -1000}
	for _, count := range cases {
		nl := components.Newline(count)
		if nl.Type != vdom.ElementNode {
			t.Errorf("count=%d: Newline should create a text element, got %v", count, nl.Type)
		}
		if len(nl.Children) != 1 || nl.Children[0].Text != "" {
			t.Errorf("count=%d: expected empty text for negative count, got %#v", count, nl.Children)
		}
	}
}

// TestNewlineLargeCount verifies that very large counts emit the requested
// number of newline characters with no cap — matching upstream's
// '\n'.repeat(count) which has no upper bound short of v8's string-length
// limit. This pins down behavior so a future "safety cap" cannot be added
// without breaking parity.
func TestNewlineLargeCount(t *testing.T) {
	const count = 1000
	nl := components.Newline(count)
	if nl.Type != vdom.ElementNode {
		t.Fatalf("expected text element, got %v", nl.Type)
	}
	if len(nl.Children) != 1 {
		t.Fatalf("expected one text child, got %#v", nl.Children)
	}
	text := nl.Children[0].Text
	if len(text) != count {
		t.Fatalf("expected %d newline characters, got %d", count, len(text))
	}
	for i, r := range text {
		if r != '\n' {
			t.Fatalf("expected '\\n' at position %d, got %q", i, r)
		}
	}
}

func TestSpace(t *testing.T) {
	space := components.Space()

	if space.Type != vdom.ElementNode {
		t.Error("Space should create a text element")
	}

	if space.ElementType != "text" {
		t.Errorf("Expected space element type %q, got %q", "text", space.ElementType)
	}

	if len(space.Children) != 1 || space.Children[0].Text != " " {
		t.Errorf("Expected space character, got %#v", space.Children)
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

func TestTransformWithProps(t *testing.T) {
	transform := components.Transform(
		func(children string, index int) string {
			return children
		},
		vdom.Props{"accessibilityLabel": "spoken"},
		vdom.CreateTextNode("visual"),
	)

	if transform.Props["accessibilityLabel"] != "spoken" {
		t.Errorf("Expected accessibilityLabel prop, got %v", transform.Props["accessibilityLabel"])
	}
}

func TestTransformWithoutChildrenReturnsNil(t *testing.T) {
	if transform := components.Transform(func(children string, index int) string {
		return children
	}); transform != nil {
		t.Fatalf("Expected nil transform for nil children, got %#v", transform)
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

	if border.ElementType != "box" {
		t.Errorf("Expected element type 'box', got %q", border.ElementType)
	}

	// Check props
	if border.Props["borderStyle"] != string(components.BorderSingle) {
		t.Error("Border style prop not set correctly")
	}

	if border.Props["borderTop"] != true {
		t.Error("Border top prop not set correctly")
	}
}

func TestBorderDefaultsToAllSides(t *testing.T) {
	border := components.Border(components.BorderProps{Style: components.BorderSingle}, nil)

	if _, ok := border.Props["borderTop"]; ok {
		t.Fatal("Expected default border to omit side overrides")
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
