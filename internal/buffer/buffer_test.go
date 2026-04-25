package buffer_test

import (
	"testing"

	"github.com/dh-kam/goink.go/internal/buffer"
)

// TestNewBuffer tests creating a new buffer
func TestNewBuffer(t *testing.T) {
	width, height := 10, 5
	buf := buffer.New(width, height)

	if buf.Width() != width {
		t.Errorf("Expected width %d, got %d", width, buf.Width())
	}

	if buf.Height() != height {
		t.Errorf("Expected height %d, got %d", height, buf.Height())
	}
}

// TestBufferSet tests setting a character at a position
func TestBufferSet(t *testing.T) {
	buf := buffer.New(10, 5)

	buf.Set(2, 3, 'A')

	if got := buf.Get(2, 3); got != 'A' {
		t.Errorf("Expected 'A' at (2,3), got %q", got)
	}
}

// TestBufferClear tests clearing the buffer
func TestBufferClear(t *testing.T) {
	buf := buffer.New(10, 5)

	buf.Set(2, 3, 'A')
	buf.Clear()

	if got := buf.Get(2, 3); got != ' ' {
		t.Errorf("Expected space after clear, got %q", got)
	}
}

// TestBufferWriteString tests writing a string
func TestBufferWriteString(t *testing.T) {
	buf := buffer.New(20, 5)

	buf.WriteString(0, 0, "Hello")

	expected := "Hello"
	for i, ch := range expected {
		if got := buf.Get(i, 0); got != ch {
			t.Errorf("Expected %q at (%d,0), got %q", ch, i, got)
		}
	}
}

// TestBufferWriteStringWrapping tests string doesn't overflow
func TestBufferWriteStringWrapping(t *testing.T) {
	buf := buffer.New(5, 3)

	// This should only write "Hello" and truncate " World"
	buf.WriteString(0, 0, "Hello World")

	// Check that only "Hello" was written
	if buf.Get(0, 0) != 'H' || buf.Get(4, 0) != 'o' {
		t.Error("String should be truncated to buffer width")
	}
}

func TestBufferWriteStringMultiline(t *testing.T) {
	buf := buffer.New(10, 3)
	buf.WriteString(2, 0, "A\nB")

	if got := buf.Get(2, 0); got != 'A' {
		t.Errorf("Expected 'A' at (2,0), got %q", got)
	}

	if got := buf.Get(2, 1); got != 'B' {
		t.Errorf("Expected 'B' at (2,1), got %q", got)
	}
}

// TestBufferRender tests rendering buffer to string
func TestBufferRender(t *testing.T) {
	buf := buffer.New(5, 2)

	buf.WriteString(0, 0, "Hello")
	buf.WriteString(0, 1, "World")

	output := buf.Render()
	expected := "Hello\nWorld"

	if output != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, output)
	}
}

// TestBufferRenderTrimsTrailingSpaces tests that trailing spaces are removed
func TestBufferRenderTrimsTrailingSpaces(t *testing.T) {
	buf := buffer.New(10, 2)

	buf.WriteString(0, 0, "Hi")
	// Second line is empty (all spaces)

	output := buf.Render()
	expected := "Hi" // Empty lines at the end are removed

	if output != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, output)
	}
}

func TestBufferRenderRowsPreservesTrailingEmptyLines(t *testing.T) {
	buf := buffer.New(5, 3)
	buf.WriteString(0, 0, "X")

	output := buf.RenderRows(3)
	expected := "X\n\n"

	if output != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, output)
	}
}

func TestBufferRenderRowsSkipsUndefinedHoles(t *testing.T) {
	buf := buffer.New(10, 1)
	buf.Set(0, 0, '│')
	buf.Set(11, 0, '│')

	output := buf.RenderRows(1)
	expected := "│         │"

	if output != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, output)
	}
}

func TestBufferWriteStringWideRuneSkipsContinuationCell(t *testing.T) {
	buf := buffer.New(4, 1)
	buf.WriteString(0, 0, "🍔|")

	output := buf.RenderRows(1)
	expected := "🍔|"

	if output != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, output)
	}
}
