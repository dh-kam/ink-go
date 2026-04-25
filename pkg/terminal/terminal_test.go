package terminal_test

import (
	"os"
	"testing"

	"github.com/dh-kam/goink.go/pkg/terminal"
)

// TestIsTerminal tests terminal detection
func TestIsTerminal(t *testing.T) {
	// Stdin should be detected (in test environment it may not be, but we test the function)
	fd := int(os.Stdin.Fd())
	isTerm := terminal.IsTerminal(fd)
	// We just verify the function doesn't panic
	_ = isTerm
}

// TestMakeRaw tests raw mode setup
func TestMakeRaw(t *testing.T) {
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		// May fail in non-terminal environment, that's ok
		t.Skip("Not a terminal environment")
	}
	defer state.Restore()

	if state == nil {
		t.Error("Expected state to be non-nil")
	}
}

// TestGetSize tests getting terminal size
func TestGetSize(t *testing.T) {
	width, height, err := terminal.GetSize(0)
	if err != nil {
		t.Errorf("GetSize failed: %v", err)
	}
	if width <= 0 {
		t.Errorf("Expected positive width, got %d", width)
	}
	if height <= 0 {
		t.Errorf("Expected positive height, got %d", height)
	}
}

// TestSetupSignalHandler tests signal handler setup
func TestSetupSignalHandler(t *testing.T) {
	sigChan := terminal.SetupSignalHandler()
	if sigChan == nil {
		t.Error("Expected signal channel to be non-nil")
	}
	// Don't wait for signals, just verify setup worked
}

// TestStdinIsTerminal tests stdin terminal check
func TestStdinIsTerminal(t *testing.T) {
	isTerm := terminal.StdinIsTerminal()
	_ = isTerm // Just verify it doesn't panic
}

// TestStdoutIsTerminal tests stdout terminal check
func TestStdoutIsTerminal(t *testing.T) {
	isTerm := terminal.StdoutIsTerminal()
	_ = isTerm // Just verify it doesn't panic
}

// TestStateRestore tests state restoration
func TestStateRestore(t *testing.T) {
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		t.Skip("Not a terminal environment")
	}

	// Restore should not error
	err = state.Restore()
	if err != nil {
		t.Errorf("Restore failed: %v", err)
	}
}

// TestClearScreen doesn't actually test output (hard to capture)
// but verifies the function exists and can be called
func TestClearScreen(t *testing.T) {
	terminal.ClearScreen()
}

// TestHideCursor tests cursor hiding
func TestHideCursor(t *testing.T) {
	terminal.HideCursor()
	terminal.ShowCursor() // Restore
}

// TestShowCursor tests cursor showing
func TestShowCursor(t *testing.T) {
	terminal.ShowCursor()
}

// TestMoveCursor tests cursor movement
func TestMoveCursor(t *testing.T) {
	terminal.MoveCursor(1, 1)
	terminal.MoveCursor(10, 20)
}

// TestEnableAlternateScreenBuffer tests alternate screen
func TestEnableAlternateScreenBuffer(t *testing.T) {
	terminal.EnableAlternateScreenBuffer()
	terminal.DisableAlternateScreenBuffer() // Restore
}

// TestDisableAlternateScreenBuffer tests disabling alternate screen
func TestDisableAlternateScreenBuffer(t *testing.T) {
	terminal.DisableAlternateScreenBuffer()
}
