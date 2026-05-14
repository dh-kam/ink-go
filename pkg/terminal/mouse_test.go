package terminal_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/dh-kam/goink.go/pkg/terminal"
)

func TestEnableMouseToBytes(t *testing.T) {
	var buf bytes.Buffer
	if err := terminal.EnableMouseTo(&buf); err != nil {
		t.Fatalf("EnableMouseTo: %v", err)
	}
	want := "\x1b[?1000h\x1b[?1006h"
	if buf.String() != want {
		t.Fatalf("EnableMouseTo wrote %q, want %q", buf.String(), want)
	}
}

func TestDisableMouseToBytes(t *testing.T) {
	var buf bytes.Buffer
	if err := terminal.DisableMouseTo(&buf); err != nil {
		t.Fatalf("DisableMouseTo: %v", err)
	}
	want := "\x1b[?1006l\x1b[?1000l"
	if buf.String() != want {
		t.Fatalf("DisableMouseTo wrote %q, want %q", buf.String(), want)
	}
}

func TestEnableMouseToNil(t *testing.T) {
	if err := terminal.EnableMouseTo(nil); !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("EnableMouseTo(nil) err = %v, want os.ErrInvalid", err)
	}
}

func TestDisableMouseToNil(t *testing.T) {
	if err := terminal.DisableMouseTo(nil); !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("DisableMouseTo(nil) err = %v, want os.ErrInvalid", err)
	}
}

func TestEnableMouseDoesNotPanic(t *testing.T) {
	// Calls real stdout; just assert no panic / error.
	if err := terminal.EnableMouse(); err != nil {
		t.Fatalf("EnableMouse: %v", err)
	}
	if err := terminal.DisableMouse(); err != nil {
		t.Fatalf("DisableMouse: %v", err)
	}
}
