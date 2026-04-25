package terminal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// State holds the terminal state
type State struct {
	oldState *State
	fd       int
}

// IsTerminal checks if the given file descriptor is a terminal
func IsTerminal(fd int) bool {
	// In a real implementation, this would use isatty
	// For now, we'll assume stdin is a terminal
	return fd == 0
}

// MakeRaw puts the terminal into raw mode
// This is a simplified version - in production you'd use golang.org/x/term
func MakeRaw(fd int) (*State, error) {
	if !IsTerminal(fd) {
		return nil, fmt.Errorf("not a terminal")
	}

	// Store old state for restoration
	state := &State{
		fd: fd,
	}

	// In a real implementation with golang.org/x/term:
	// oldState, err := term.MakeRaw(fd)
	// state.oldState = oldState

	return state, nil
}

// Restore restores the terminal to its original state
func (s *State) Restore() error {
	// In a real implementation with golang.org/x/term:
	// return term.Restore(s.fd, s.oldState)
	return nil
}

// GetSize returns the terminal dimensions
func GetSize(fd int) (width, height int, err error) {
	// In a real implementation with golang.org/x/term:
	// width, height, err = term.GetSize(fd)
	// For now, return defaults
	return 80, 24, nil
}

// SetupSignalHandler sets up signal handlers for graceful shutdown
func SetupSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // kill
		syscall.SIGQUIT, // Ctrl+\
		syscall.SIGWINCH, // Window resize
	)
	return sigChan
}

// ClearScreen clears the terminal screen
func ClearScreen() {
	fmt.Print("\x1b[2J\x1b[H")
}

// HideCursor hides the cursor
func HideCursor() {
	fmt.Print("\x1b[?25l")
}

// ShowCursor shows the cursor
func ShowCursor() {
	fmt.Print("\x1b[?25h")
}

// MoveCursor moves the cursor to the specified position
func MoveCursor(row, col int) {
	fmt.Printf("\x1b[%d;%dH", row, col)
}

// EnableAlternateScreenBuffer switches to the alternate screen buffer
func EnableAlternateScreenBuffer() {
	fmt.Print("\x1b[?1049h")
}

// DisableAlternateScreenBuffer returns from the alternate screen buffer
func DisableAlternateScreenBuffer() {
	fmt.Print("\x1b[?1049l")
}

// StdinIsTerminal checks if stdin is a terminal
func StdinIsTerminal() bool {
	return IsTerminal(int(os.Stdin.Fd()))
}

// StdoutIsTerminal checks if stdout is a terminal
func StdoutIsTerminal() bool {
	return IsTerminal(int(os.Stdout.Fd()))
}
