package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

// TestTextInputDefault tests default text input
func TestTextInputDefault(t *testing.T) {
	node := components.TextInput(components.TextInputProps{
		Value: "hello",
	})

	if node == nil {
		t.Fatal("TextInput should return a non-nil node")
	}

	if node.ElementType != "textinput" {
		t.Errorf("Expected element type 'textinput', got %q", node.ElementType)
	}

	text := node.Children[0].Text
	if !strings.Contains(text, "hello") {
		t.Errorf("Expected text to contain 'hello', got %q", text)
	}

	// Should be wrapped in brackets
	if !strings.HasPrefix(text, "[") {
		t.Error("Expected text to start with '['")
	}
	if !strings.HasSuffix(text, "]") {
		t.Error("Expected text to end with ']'")
	}
}

// TestTextInputEmpty tests empty text input
func TestTextInputEmpty(t *testing.T) {
	node := components.TextInput(components.TextInputProps{
		Value: "",
	})

	text := node.Children[0].Text
	// Should have brackets and spaces
	if !strings.HasPrefix(text, "[") {
		t.Error("Expected text to start with '['")
	}
	if !strings.HasSuffix(text, "]") {
		t.Error("Expected text to end with ']'")
	}
}

// TestTextInputWithPlaceholder tests placeholder text
func TestTextInputWithPlaceholder(t *testing.T) {
	node := components.TextInput(components.TextInputProps{
		Value:       "",
		Placeholder: "Enter text...",
	})

	text := node.Children[0].Text
	if !strings.Contains(text, "Enter text...") {
		t.Errorf("Expected text to contain placeholder, got %q", text)
	}
}

// TestPlaceholderShownWithValue tests that placeholder is not shown when value exists
func TestPlaceholderShownWithValue(t *testing.T) {
	node := components.TextInput(components.TextInputProps{
		Value:       "actual",
		Placeholder: "placeholder",
	})

	text := node.Children[0].Text
	if strings.Contains(text, "placeholder") {
		t.Error("Expected placeholder to be hidden when value exists")
	}
	if !strings.Contains(text, "actual") {
		t.Error("Expected value to be shown")
	}
}

// TestPasswordInput tests password masking
func TestPasswordInput(t *testing.T) {
	node := components.PasswordInput(components.TextInputProps{
		Value: "secret",
	})

	text := node.Children[0].Text

	// Should not contain the actual value
	if strings.Contains(text, "secret") {
		t.Error("Expected password to be masked")
	}

	// Should contain mask characters
	if !strings.Contains(text, components.DefaultMaskChar) {
		t.Error("Expected mask character in output")
	}
}

// TestTextInputMaskCustom tests custom mask character
func TestTextInputMaskCustom(t *testing.T) {
	node := components.TextInput(components.TextInputProps{
		Value:    "secret",
		Mask:     true,
		MaskChar: "*",
	})

	text := node.Children[0].Text

	if strings.Contains(text, "secret") {
		t.Error("Expected password to be masked")
	}

	if !strings.Contains(text, "*") {
		t.Error("Expected custom mask character '*'")
	}
}

// TestTextInputWidth tests custom width
func TestTextInputWidth(t *testing.T) {
	node := components.TextInput(components.TextInputProps{
		Value: "test",
		Width: 10,
	})

	text := node.Children[0].Text
	// Length should be width + 2 (for brackets) = 12
	expectedLen := 12
	actualLen := 0
	for range text {
		actualLen++
	}

	if actualLen != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, actualLen)
	}
}

// TestTextInputSimple tests simple text input helper
func TestTextInputSimple(t *testing.T) {
	node := components.TextInputSimple("hello world")

	if node == nil {
		t.Fatal("TextInputSimple should return a non-nil node")
	}

	text := node.Children[0].Text
	if !strings.Contains(text, "hello world") {
		t.Errorf("Expected text to contain 'hello world', got %q", text)
	}
}

// TestTextInputWithPlaceholderHelper tests placeholder helper
func TestTextInputWithPlaceholderHelper(t *testing.T) {
	node := components.TextInputWithPlaceholder("", "type here")

	if node == nil {
		t.Fatal("TextInputWithPlaceholder should return a non-nil node")
	}

	text := node.Children[0].Text
	if !strings.Contains(text, "type here") {
		t.Errorf("Expected text to contain 'type here', got %q", text)
	}
}

// TestTextInputLongValue tests truncation of long values
func TestTextInputLongValue(t *testing.T) {
	longValue := strings.Repeat("x", 100)
	node := components.TextInput(components.TextInputProps{
		Value: longValue,
		Width: 10,
	})

	text := node.Children[0].Text
	// Should be truncated to width
	// Count characters between brackets
	content := strings.TrimPrefix(text, "[")
	content = strings.TrimSuffix(content, "]")

	if len(content) > 10 {
		t.Errorf("Expected content to be truncated to 10 chars, got %d", len(content))
	}
}

// TestNewTextInputState tests creating new input state
func TestNewTextInputState(t *testing.T) {
	state := components.NewTextInputState("hello")

	if state.Value != "hello" {
		t.Errorf("Expected value 'hello', got %q", state.Value)
	}

	if state.CursorPos != 5 {
		t.Errorf("Expected cursor position 5, got %d", state.CursorPos)
	}

	if state.Focus {
		t.Error("Expected focus to be false initially")
	}
}

// TestTextInputStateInsert tests inserting characters
func TestTextInputStateInsert(t *testing.T) {
	state := components.NewTextInputState("")
	state.Insert("h")
	state.Insert("e")
	state.Insert("l")
	state.Insert("l")
	state.Insert("o")

	if state.Value != "hello" {
		t.Errorf("Expected value 'hello', got %q", state.Value)
	}

	if state.CursorPos != 5 {
		t.Errorf("Expected cursor position 5, got %d", state.CursorPos)
	}
}

// TestTextInputStateDelete tests deleting characters
func TestTextInputStateDelete(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.CursorPos = 5
	state.Delete() // delete 'o'
	state.Delete() // delete 'l'

	if state.Value != "hel" {
		t.Errorf("Expected value 'hel', got %q", state.Value)
	}

	if state.CursorPos != 3 {
		t.Errorf("Expected cursor position 3, got %d", state.CursorPos)
	}
}

// TestTextInputStateDeleteAtStart tests delete at position 0
func TestTextInputStateDeleteAtStart(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.CursorPos = 0
	state.Delete()

	// Should not delete anything at position 0
	if state.Value != "hello" {
		t.Errorf("Expected value 'hello', got %q", state.Value)
	}
	if state.CursorPos != 0 {
		t.Errorf("Expected cursor position 0, got %d", state.CursorPos)
	}
}

// TestTextInputStateMoveLeft tests moving cursor left
func TestTextInputStateMoveLeft(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.MoveLeft()
	state.MoveLeft()

	if state.CursorPos != 3 {
		t.Errorf("Expected cursor position 3, got %d", state.CursorPos)
	}
}

// TestTextInputStateMoveLeftAtStart tests moving left at position 0
func TestTextInputStateMoveLeftAtStart(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.CursorPos = 0
	state.MoveLeft()

	if state.CursorPos != 0 {
		t.Errorf("Expected cursor position 0, got %d", state.CursorPos)
	}
}

// TestTextInputStateMoveRight tests moving cursor right
func TestTextInputStateMoveRight(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.CursorPos = 0
	state.MoveRight()
	state.MoveRight()

	if state.CursorPos != 2 {
		t.Errorf("Expected cursor position 2, got %d", state.CursorPos)
	}
}

// TestTextInputStateMoveRightAtEnd tests moving right at end
func TestTextInputStateMoveRightAtEnd(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.MoveRight()

	// Should not go past end
	if state.CursorPos != 5 {
		t.Errorf("Expected cursor position 5, got %d", state.CursorPos)
	}
}

// TestTextInputStateMoveToStart tests moving to start
func TestTextInputStateMoveToStart(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.CursorPos = 5
	state.MoveToStart()

	if state.CursorPos != 0 {
		t.Errorf("Expected cursor position 0, got %d", state.CursorPos)
	}
}

// TestTextInputStateMoveToEnd tests moving to end
func TestTextInputStateMoveToEnd(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.CursorPos = 0
	state.MoveToEnd()

	if state.CursorPos != 5 {
		t.Errorf("Expected cursor position 5, got %d", state.CursorPos)
	}
}

// TestTextInputStateClear tests clearing the input
func TestTextInputStateClear(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.Clear()

	if state.Value != "" {
		t.Errorf("Expected empty value, got %q", state.Value)
	}

	if state.CursorPos != 0 {
		t.Errorf("Expected cursor position 0, got %d", state.CursorPos)
	}
}

// TestTextInputStateSetValue tests setting value
func TestTextInputStateSetValue(t *testing.T) {
	state := components.NewTextInputState("hello")
	state.SetValue("world")

	if state.Value != "world" {
		t.Errorf("Expected value 'world', got %q", state.Value)
	}

	if state.CursorPos != 5 {
		t.Errorf("Expected cursor position 5, got %d", state.CursorPos)
	}
}

// TestTextInputStateComplexOperations tests complex state operations
func TestTextInputStateComplexOperations(t *testing.T) {
	state := components.NewTextInputState("")

	// Type "hello"
	state.Insert("h")
	state.Insert("e")
	state.Insert("l")
	state.Insert("l")
	state.Insert("o")

	// Move left twice
	state.MoveLeft()
	state.MoveLeft()

	// Delete removes character BEFORE cursor
	// After "hello" with cursor at pos 3, delete removes char at pos 2 ('l')
	// Result: "helo", cursor at pos 2
	state.Delete()

	if state.Value != "helo" {
		t.Errorf("Expected value 'helo', got %q", state.Value)
	}

	if state.CursorPos != 2 {
		t.Errorf("Expected cursor position 2, got %d", state.CursorPos)
	}
}

// TestDefaultTextInputConstants tests default constants
func TestDefaultTextInputConstants(t *testing.T) {
	if components.DefaultTextInputWidth != 30 {
		t.Errorf("Expected DefaultTextInputWidth 30, got %d", components.DefaultTextInputWidth)
	}

	if components.DefaultMaskChar != "•" {
		t.Errorf("Expected DefaultMaskChar '•', got %q", components.DefaultMaskChar)
	}
}

// TestTextInputWithFocus tests input with focus indicator
func TestTextInputWithFocus(t *testing.T) {
	node := components.TextInput(components.TextInputProps{
		Value: "test",
		Focus: true,
	})

	if node == nil {
		t.Fatal("TextInput should return a non-nil node")
	}

	// Focus is handled internally, just verify it doesn't crash
	_ = node.Children[0].Text
}

// TestTextInputCursorPos tests cursor position handling
func TestTextInputCursorPos(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		cursorPos   int
		expectValid bool
	}{
		{"valid position", "hello", 2, true},
		{"at start", "hello", 0, true},
		{"at end", "hello", 5, true},
		{"past end", "hello", 10, true},
		{"negative", "hello", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := components.TextInput(components.TextInputProps{
				Value:     tt.value,
				CursorPos: tt.cursorPos,
			})

			if node == nil {
				t.Fatal("TextInput should return a non-nil node")
			}

			// Should handle all cursor positions gracefully
			_ = node.Children[0].Text
		})
	}
}
