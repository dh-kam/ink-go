package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

// TestSpinnerDefault tests creating a default spinner
func TestSpinnerDefault(t *testing.T) {
	node := components.Spinner(components.SpinnerProps{})

	if node == nil {
		t.Fatal("Spinner should return a non-nil node")
	}

	if node.ElementType != "spinner" {
		t.Errorf("Expected element type 'spinner', got %q", node.ElementType)
	}

	if len(node.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(node.Children))
	}
}

// TestSpinnerWithText tests spinner with text
func TestSpinnerWithText(t *testing.T) {
	node := components.SpinnerWithText("Loading...")

	if node == nil {
		t.Fatal("SpinnerWithText should return a non-nil node")
	}

	if node.ElementType != "spinner" {
		t.Errorf("Expected element type 'spinner', got %q", node.ElementType)
	}

	// Check that the text is present in the output
	if len(node.Children) == 0 {
		t.Fatal("Expected at least 1 child")
	}

	text := node.Children[0].Text
	if !strings.Contains(text, "Loading...") {
		t.Errorf("Expected text to contain 'Loading...', got %q", text)
	}
}

// TestSpinnerWithType tests spinner with specific type
func TestSpinnerWithType(t *testing.T) {
	node := components.SpinnerWithType(
		components.LineSpinner,
		"Processing",
	)

	if node == nil {
		t.Fatal("SpinnerWithType should return a non-nil node")
	}

	if len(node.Children) == 0 {
		t.Fatal("Expected at least 1 child")
	}

	text := node.Children[0].Text
	if !strings.Contains(text, "Processing") {
		t.Errorf("Expected text to contain 'Processing', got %q", text)
	}
}

// TestSpinnerTypes tests all spinner types
func TestSpinnerTypes(t *testing.T) {
	types := []*components.SpinnerFrames{
		components.DotsSpinner,
		components.LineSpinner,
		components.ArrowSpinner,
		components.PlusMinusSpinner,
	}

	for _, spinnerType := range types {
		t.Run("", func(t *testing.T) {
			node := components.Spinner(components.SpinnerProps{
				Type: spinnerType,
			})

			if node == nil {
				t.Fatal("Spinner should return a non-nil node")
			}

			if len(node.Children) == 0 {
				t.Fatal("Expected at least 1 child")
			}

			// The spinner should have some text (the frame character)
			text := node.Children[0].Text
			if text == "" {
				t.Error("Expected spinner to have frame text")
			}
		})
	}
}

// TestSpinnerFramesStructure tests spinner frame structure
func TestSpinnerFramesStructure(t *testing.T) {
	// Test DotsSpinner
	if components.DotsSpinner == nil {
		t.Error("DotsSpinner should not be nil")
	}
	if len(components.DotsSpinner.Frames()) == 0 {
		t.Error("DotsSpinner should have frames")
	}

	// Test LineSpinner
	if components.LineSpinner == nil {
		t.Error("LineSpinner should not be nil")
	}
	if len(components.LineSpinner.Frames()) == 0 {
		t.Error("LineSpinner should have frames")
	}

	// Test ArrowSpinner
	if components.ArrowSpinner == nil {
		t.Error("ArrowSpinner should not be nil")
	}

	// Test PlusMinusSpinner
	if components.PlusMinusSpinner == nil {
		t.Error("PlusMinusSpinner should not be nil")
	}

	// Test DefaultSpinner
	if components.DefaultSpinner == nil {
		t.Error("DefaultSpinner should not be nil")
	}

	// DefaultSpinner should equal DotsSpinner
	if components.DefaultSpinner != components.DotsSpinner {
		t.Error("DefaultSpinner should equal DotsSpinner")
	}
}

// TestSpinnerFramesContent tests that spinner frames have expected content
func TestSpinnerFramesContent(t *testing.T) {
	// DotsSpinner should have 10 frames
	if len(components.DotsSpinner.Frames()) != 10 {
		t.Errorf("Expected DotsSpinner to have 10 frames, got %d", len(components.DotsSpinner.Frames()))
	}

	// LineSpinner should have 4 frames
	if len(components.LineSpinner.Frames()) != 4 {
		t.Errorf("Expected LineSpinner to have 4 frames, got %d", len(components.LineSpinner.Frames()))
	}

	// All frames should be non-empty
	for _, spinner := range []*components.SpinnerFrames{
		components.DotsSpinner,
		components.LineSpinner,
		components.ArrowSpinner,
		components.PlusMinusSpinner,
	} {
		for i, frame := range spinner.Frames() {
			if frame == "" {
				t.Errorf("Frame %d should not be empty", i)
			}
		}
	}
}
