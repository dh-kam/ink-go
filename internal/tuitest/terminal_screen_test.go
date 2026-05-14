package tuitest

import "testing"

func TestTerminalScreenPlainTextAndCRLF(t *testing.T) {
	screen := NewTerminalScreen(20, 5)
	screen.Apply("hello\r\nworld")

	if got := screen.PlainString(); got != "hello\nworld" {
		t.Fatalf("screen mismatch:\n%s", got)
	}
}

func TestTerminalScreenCursorMovementAndEraseLine(t *testing.T) {
	screen := NewTerminalScreen(20, 5)
	screen.Apply("first\r\nsecond\r\nthird\x1b[1A\r\x1b[2Kupdated")

	if got := screen.PlainString(); got != "first\nupdated\nthird" {
		t.Fatalf("screen mismatch:\n%s", got)
	}
}

func TestTerminalScreenClearDisplay(t *testing.T) {
	screen := NewTerminalScreen(20, 5)
	screen.Apply("stale\r\ntext\x1b[2J\x1b[Hfresh")

	if got := screen.PlainString(); got != "fresh" {
		t.Fatalf("screen mismatch:\n%s", got)
	}
}

func TestTerminalScreenWideRune(t *testing.T) {
	screen := NewTerminalScreen(20, 5)
	screen.Apply("A한B")

	if got := screen.PlainString(); got != "A한B" {
		t.Fatalf("screen mismatch:\n%s", got)
	}
}
