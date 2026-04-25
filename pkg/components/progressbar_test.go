package components_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/dh-kam/goink.go/pkg/components"
)

// TestProgressBarDefault tests default progress bar
func TestProgressBarDefault(t *testing.T) {
	node := components.ProgressBar(components.ProgressBarProps{
		Percent: 50,
	})

	if node == nil {
		t.Fatal("ProgressBar should return a non-nil node")
	}

	if node.ElementType != "progress" {
		t.Errorf("Expected element type 'progress', got %q", node.ElementType)
	}

	if len(node.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(node.Children))
	}

	text := node.Children[0].Text
	// Default width is 40, 50% filled = 20 chars
	expectedLen := 40
	if utf8.RuneCountInString(text) != expectedLen {
		t.Errorf("Expected text length %d, got %d", expectedLen, utf8.RuneCountInString(text))
	}
}

// TestProgressBarZero tests zero percent
func TestProgressBarZero(t *testing.T) {
	node := components.ProgressBar(components.ProgressBarProps{
		Percent: 0,
		Width:   10,
	})

	text := node.Children[0].Text
	// Should be all spaces
	if text != strings.Repeat(" ", 10) {
		t.Errorf("Expected 10 spaces, got %q", text)
	}
}

// TestProgressBarFull tests 100 percent
func TestProgressBarFull(t *testing.T) {
	node := components.ProgressBar(components.ProgressBarProps{
		Percent: 100,
		Width:   10,
	})

	text := node.Children[0].Text
	// Should be all filled characters
	expected := strings.Repeat(components.DefaultProgressBarChar, 10)
	if text != expected {
		t.Errorf("Expected %q, got %q", expected, text)
	}
}

// TestProgressBarNegative tests negative percent clamps to 0
func TestProgressBarNegative(t *testing.T) {
	node := components.ProgressBar(components.ProgressBarProps{
		Percent: -10,
		Width:   10,
	})

	text := node.Children[0].Text
	// Should be all spaces (clamped to 0%)
	if text != strings.Repeat(" ", 10) {
		t.Errorf("Expected 10 spaces, got %q", text)
	}
}

// TestProgressBarOver100 tests percent over 100 clamps to 100
func TestProgressBarOver100(t *testing.T) {
	node := components.ProgressBar(components.ProgressBarProps{
		Percent: 150,
		Width:   10,
	})

	text := node.Children[0].Text
	// Should be all filled (clamped to 100%)
	expected := strings.Repeat(components.DefaultProgressBarChar, 10)
	if text != expected {
		t.Errorf("Expected %q, got %q", expected, text)
	}
}

// TestProgressBarWithPercent tests showing percentage
func TestProgressBarWithPercent(t *testing.T) {
	node := components.ProgressBar(components.ProgressBarProps{
		Percent:     75,
		Width:       10,
		ShowPercent: true,
	})

	text := node.Children[0].Text
	if !strings.Contains(text, "75%") {
		t.Errorf("Expected text to contain '75%%', got %q", text)
	}
}

// TestProgressBarCustomCharacter tests custom character
func TestProgressBarCustomCharacter(t *testing.T) {
	node := components.ProgressBar(components.ProgressBarProps{
		Percent:   50,
		Width:     10,
		Character: "=",
	})

	text := node.Children[0].Text
	// Should have 5 '=' and 5 spaces
	expected := "=====     "
	if text != expected {
		t.Errorf("Expected %q, got %q", expected, text)
	}
}

// TestProgressBarSimple tests simple progress bar
func TestProgressBarSimple(t *testing.T) {
	node := components.ProgressBarSimple(25)

	if node == nil {
		t.Fatal("ProgressBarSimple should return a non-nil node")
	}

	// Default width is 40, 25% = 10 chars filled
	text := node.Children[0].Text
	expectedFilled := 10
	actualFilled := 0
	for _, ch := range text {
		if ch == ' ' {
			break
		}
		actualFilled++
	}

	if actualFilled != expectedFilled {
		t.Errorf("Expected %d filled characters, got %d", expectedFilled, actualFilled)
	}
}

// TestProgressBarWithPercentFunc tests the helper function
func TestProgressBarWithPercentFunc(t *testing.T) {
	node := components.ProgressBarWithPercent(33, 20)

	text := node.Children[0].Text
	if !strings.Contains(text, "33%") {
		t.Errorf("Expected text to contain '33%%', got %q", text)
	}
}

// TestProgressBarIndeterminate tests indeterminate progress bar
func TestProgressBarIndeterminate(t *testing.T) {
	node := components.ProgressBarIndeterminate(20)

	if node == nil {
		t.Fatal("ProgressBarIndeterminate should return a non-nil node")
	}

	if node.ElementType != "progress" {
		t.Errorf("Expected element type 'progress', got %q", node.ElementType)
	}

	text := node.Children[0].Text
	// Should have a spinner character in the middle
	hasSpinnerChar := false
	for _, ch := range text {
		// Check for any of the spinner frame characters
		if ch == '⠋' || ch == '⠙' || ch == '⠹' || ch == '⠸' ||
			ch == '⠼' || ch == '⠴' || ch == '⠦' || ch == '⠧' ||
			ch == '⠇' || ch == '⠏' {
			hasSpinnerChar = true
			break
		}
	}

	if !hasSpinnerChar {
		t.Error("Expected indeterminate bar to contain spinner character")
	}
}

// TestProgressBarVariousPercentages tests various percentage values
func TestProgressBarVariousPercentages(t *testing.T) {
	tests := []struct {
		percent    int
		width      int
		wantFilled int
	}{
		{0, 10, 0},
		{10, 10, 1},
		{25, 10, 2},   // 25% of 10 = 2.5 -> 2 (integer division)
		{33, 10, 3},   // 33% of 10 = 3.3 -> 3
		{50, 10, 5},
		{75, 10, 7},   // 75% of 10 = 7.5 -> 7
		{90, 10, 9},
		{100, 10, 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			node := components.ProgressBar(components.ProgressBarProps{
				Percent: tt.percent,
				Width:   tt.width,
			})

			text := node.Children[0].Text
			actualFilled := 0
			for _, ch := range text {
				if ch != ' ' {
					actualFilled++
				}
			}

			if actualFilled != tt.wantFilled {
				t.Errorf("Percent %d, Width %d: expected %d filled, got %d",
					tt.percent, tt.width, tt.wantFilled, actualFilled)
			}
		})
	}
}

// TestProgressBarEdgeCases tests edge cases
func TestProgressBarEdgeCases(t *testing.T) {
	// Very small width
	node := components.ProgressBar(components.ProgressBarProps{
		Percent: 50,
		Width:   1,
	})
	text := node.Children[0].Text
	if utf8.RuneCountInString(text) != 1 {
		t.Errorf("Expected length 1, got %d", utf8.RuneCountInString(text))
	}

	// Zero width should use default (40 chars filled portion only)
	node = components.ProgressBar(components.ProgressBarProps{
		Percent: 50,
		Width:   0,
	})
	text = node.Children[0].Text
	// Should be 40 chars (bar only, no percent)
	if utf8.RuneCountInString(text) != components.DefaultProgressBarWidth {
		t.Errorf("Expected length %d, got %d", components.DefaultProgressBarWidth, utf8.RuneCountInString(text))
	}

	// Negative width should use default
	node = components.ProgressBar(components.ProgressBarProps{
		Percent: 50,
		Width:   -5,
	})
	text = node.Children[0].Text
	if utf8.RuneCountInString(text) != components.DefaultProgressBarWidth {
		t.Errorf("Expected length %d, got %d", components.DefaultProgressBarWidth, utf8.RuneCountInString(text))
	}
}

// TestDefaultProgressBarConstants tests default constants
func TestDefaultProgressBarConstants(t *testing.T) {
	if components.DefaultProgressBarWidth != 40 {
		t.Errorf("Expected DefaultProgressBarWidth 40, got %d", components.DefaultProgressBarWidth)
	}

	if components.DefaultProgressBarChar != "█" {
		t.Errorf("Expected DefaultProgressBarChar '█', got %q", components.DefaultProgressBarChar)
	}
}
