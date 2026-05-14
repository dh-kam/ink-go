package terminal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// State holds the terminal state
type State struct {
	oldState *term.State
	fd       int
}

// IsTerminal checks if the given file descriptor is a terminal
func IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

// MakeRaw puts the terminal into raw mode
func MakeRaw(fd int) (*State, error) {
	restoreOutputMode, restoreOutputModeErr := captureOutputMode(fd)
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("make raw: %w", err)
	}
	if restoreOutputModeErr == nil {
		if err := restoreOutputMode(); err != nil {
			_ = term.Restore(fd, oldState)
			return nil, fmt.Errorf("restore output mode: %w", err)
		}
	}

	state := &State{
		oldState: oldState,
		fd:       fd,
	}

	return state, nil
}

// Restore restores the terminal to its original state
func (s *State) Restore() error {
	if s == nil || s.oldState == nil {
		return nil
	}

	return term.Restore(s.fd, s.oldState)
}

// GetSize returns the terminal dimensions
func GetSize(fd int) (width, height int, err error) {
	width, height, err = term.GetSize(fd)
	if err != nil {
		return 80, 24, nil
	}

	return width, height, nil
}

// SetupSignalHandler sets up signal handlers for graceful shutdown
func SetupSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,   // Ctrl+C
		syscall.SIGTERM,  // kill
		syscall.SIGQUIT,  // Ctrl+\
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
