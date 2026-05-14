package terminal

import (
	"io"
	"os"
)

// Mouse-mode DECSET sequences (xterm). 1000 enables button events, 1006
// switches the report format to SGR (extended coordinates, separate
// release events).
const (
	enableMouseSeq  = "\x1b[?1000h\x1b[?1006h"
	disableMouseSeq = "\x1b[?1006l\x1b[?1000l"
)

// EnableMouse enables xterm SGR mouse reporting on the process stdout.
// Call DisableMouse during shutdown to restore the terminal state.
func EnableMouse() error {
	return EnableMouseTo(os.Stdout)
}

// DisableMouse turns off xterm mouse reporting on the process stdout.
func DisableMouse() error {
	return DisableMouseTo(os.Stdout)
}

// EnableMouseTo writes the enable sequence to w. Useful for tests with a
// bytes.Buffer instead of stdout.
func EnableMouseTo(w io.Writer) error {
	if w == nil {
		return os.ErrInvalid
	}
	_, err := io.WriteString(w, enableMouseSeq)
	return err
}

// DisableMouseTo writes the disable sequence to w.
func DisableMouseTo(w io.Writer) error {
	if w == nil {
		return os.ErrInvalid
	}
	_, err := io.WriteString(w, disableMouseSeq)
	return err
}
