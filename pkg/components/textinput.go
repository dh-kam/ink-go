package components

import (
	"strings"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TextInputProps defines the properties for a TextInput component
type TextInputProps struct {
	// Value is the current text value
	Value string
	// Placeholder is shown when value is empty
	Placeholder string
	// Width is the width of the input field
	Width int
	// Mask hides the input (for passwords)
	Mask bool
	// MaskChar is the character to use for masking
	MaskChar string
	// Focus indicates if the input has focus
	Focus bool
	// CursorPos is the cursor position
	CursorPos int
}

// DefaultTextInputWidth is the default width for text inputs
const DefaultTextInputWidth = 30

// DefaultMaskChar is the default character for masking
const DefaultMaskChar = "•"

// TextInput creates a text input component
func TextInput(props TextInputProps) *vdom.Node {
	// Set defaults
	if props.Width <= 0 {
		props.Width = DefaultTextInputWidth
	}
	if props.MaskChar == "" {
		props.MaskChar = DefaultMaskChar
	}

	// Determine display value
	displayValue := props.Value
	if props.Mask {
		displayValue = strings.Repeat(props.MaskChar, len(props.Value))
	}

	// Use placeholder if empty
	if displayValue == "" && props.Placeholder != "" {
		displayValue = props.Placeholder
	}

	// Truncate to width
	if len(displayValue) > props.Width {
		displayValue = displayValue[:props.Width]
	}

	// Build the input display
	var builder strings.Builder

	// Opening bracket
	builder.WriteString("[")

	// Add the text
	builder.WriteString(displayValue)

	// Pad with spaces
	remaining := props.Width - len(displayValue)
	if remaining > 0 {
		builder.WriteString(strings.Repeat(" ", remaining))
	}

	// Closing bracket
	builder.WriteString("]")

	// Add cursor indicator if focused
	if props.Focus {
		cursorPos := props.CursorPos
		if cursorPos < 0 {
			cursorPos = 0
		}
		if cursorPos > props.Width {
			cursorPos = props.Width
		}
		if cursorPos > len(displayValue) {
			cursorPos = len(displayValue)
		}
	}

	return vdom.CreateElement("textinput", nil, vdom.CreateTextNode(builder.String()))
}

// PasswordInput creates a password input with masked text
func PasswordInput(props TextInputProps) *vdom.Node {
	props.Mask = true
	return TextInput(props)
}

// TextInputSimple creates a simple text input
func TextInputSimple(value string) *vdom.Node {
	return TextInput(TextInputProps{
		Value: value,
	})
}

// TextInputWithPlaceholder creates a text input with placeholder
func TextInputWithPlaceholder(value, placeholder string) *vdom.Node {
	return TextInput(TextInputProps{
		Value:      value,
		Placeholder: placeholder,
	})
}

// TextInputState represents the state of a text input
type TextInputState struct {
	Value     string
	CursorPos int
	Focus     bool
}

// NewTextInputState creates a new text input state
func NewTextInputState(initialValue string) *TextInputState {
	return &TextInputState{
		Value:     initialValue,
		CursorPos: len(initialValue),
		Focus:     false,
	}
}

// Insert inserts a character at the cursor position
func (s *TextInputState) Insert(ch string) {
	before := s.Value[:s.CursorPos]
	after := s.Value[s.CursorPos:]
	s.Value = before + ch + after
	s.CursorPos += len(ch)
}

// Delete deletes a character at the cursor position
func (s *TextInputState) Delete() {
	if s.CursorPos > 0 {
		before := s.Value[:s.CursorPos-1]
		after := s.Value[s.CursorPos:]
		s.Value = before + after
		s.CursorPos--
	}
}

// MoveLeft moves the cursor left
func (s *TextInputState) MoveLeft() {
	if s.CursorPos > 0 {
		s.CursorPos--
	}
}

// MoveRight moves the cursor right
func (s *TextInputState) MoveRight() {
	if s.CursorPos < len(s.Value) {
		s.CursorPos++
	}
}

// MoveToStart moves cursor to the start
func (s *TextInputState) MoveToStart() {
	s.CursorPos = 0
}

// MoveToEnd moves cursor to the end
func (s *TextInputState) MoveToEnd() {
	s.CursorPos = len(s.Value)
}

// Clear clears the input value
func (s *TextInputState) Clear() {
	s.Value = ""
	s.CursorPos = 0
}

// SetValue sets the value and updates cursor position
func (s *TextInputState) SetValue(value string) {
	s.Value = value
	s.CursorPos = len(value)
}
