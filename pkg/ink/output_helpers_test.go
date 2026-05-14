package ink

import (
	"strings"
	"testing"
)

func TestCursorPositionChanged(t *testing.T) {
	if cursorPositionChanged(nil, nil) {
		t.Fatal("expected nil positions to be equal")
	}

	if !cursorPositionChanged(&CursorPosition{X: 1, Y: 2}, &CursorPosition{X: 2, Y: 2}) {
		t.Fatal("expected differing positions to be detected")
	}

	if !cursorPositionChanged(nil, &CursorPosition{X: 0, Y: 0}) {
		t.Fatal("expected nil vs non-nil positions to be different")
	}
}

func TestCursorPositionChangedSamePosition(t *testing.T) {
	if cursorPositionChanged(&CursorPosition{X: 1, Y: 2}, &CursorPosition{X: 1, Y: 2}) {
		t.Fatal("expected identical positions to be equal")
	}
}

func TestVisibleLineCount(t *testing.T) {
	if visibleLineCount("") != 0 {
		t.Fatalf("expected empty output to have zero visible lines")
	}

	if visibleLineCount("one") != 1 {
		t.Fatalf("expected single line output to have one visible line")
	}

	if visibleLineCount("one\ntwo\n") != 2 {
		t.Fatalf("expected trailing newline to be ignored in visible line count")
	}
}

func TestBuildCursorSuffix(t *testing.T) {
	result := buildCursorSuffix(3, &CursorPosition{X: 5, Y: 1})
	expected := "\x1b[2A\x1b[6G" + showCursorEscape

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestBuildCursorSuffixAtLastVisibleLine(t *testing.T) {
	result := buildCursorSuffix(3, &CursorPosition{X: 0, Y: 3})
	expected := "\x1b[G" + showCursorEscape

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestBuildCursorSuffixAtFirstLineOfSingleLineOutput(t *testing.T) {
	result := buildCursorSuffix(1, &CursorPosition{X: 4, Y: 0})
	expected := "\x1b[1A\x1b[5G" + showCursorEscape

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestBuildCursorSuffixWithoutCursor(t *testing.T) {
	if buildCursorSuffix(3, nil) != "" {
		t.Fatal("expected empty cursor suffix without a cursor position")
	}
}

func TestBuildReturnToBottomWithoutPreviousCursor(t *testing.T) {
	if buildReturnToBottom(4, nil) != "" {
		t.Fatal("expected empty return-to-bottom sequence without previous cursor")
	}
}

func TestBuildReturnToBottom(t *testing.T) {
	result := buildReturnToBottom(4, &CursorPosition{X: 5, Y: 0})
	expected := "\x1b[3B\x1b[G"

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestBuildReturnToBottomAlreadyAtBottom(t *testing.T) {
	result := buildReturnToBottom(4, &CursorPosition{X: 0, Y: 3})
	expected := "\x1b[G"

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestBuildCursorOnlySequence(t *testing.T) {
	result := buildCursorOnlySequence(
		true,
		2,
		&CursorPosition{X: 0, Y: 0},
		1,
		&CursorPosition{X: 3, Y: 0},
	)

	expected := hideCursorEscape + "\x1b[1B\x1b[G" + "\x1b[1A\x1b[4G" + showCursorEscape
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestBuildCursorOnlySequenceWithoutPreviouslyShownCursor(t *testing.T) {
	result := buildCursorOnlySequence(
		false,
		0,
		nil,
		1,
		&CursorPosition{X: 3, Y: 0},
	)

	if strings.HasPrefix(result, hideCursorEscape) {
		t.Fatalf("expected no hide prefix, got %q", result)
	}
	if !strings.Contains(result, showCursorEscape) {
		t.Fatalf("expected cursor show suffix, got %q", result)
	}
}

func TestBuildReturnToBottomPrefixWithoutShownCursor(t *testing.T) {
	if buildReturnToBottomPrefix(false, 4, &CursorPosition{X: 0, Y: 0}) != "" {
		t.Fatal("expected empty return-to-bottom prefix when cursor was not shown")
	}
}

func TestBuildReturnToBottomPrefixWithoutPreviousCursorPosition(t *testing.T) {
	result := buildReturnToBottomPrefix(true, 4, nil)
	expected := hideCursorEscape

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestAnsiEraseLines(t *testing.T) {
	result := ansiEraseLines(2)
	expected := "\x1b[2K\x1b[1A\x1b[2K\x1b[G"

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestShouldSynchronizeDisablesTTYUpdatesInCI(t *testing.T) {
	t.Setenv("CI", "true")

	if shouldSynchronize(&ttyRecordingWriter{}) {
		t.Fatal("expected CI mode to disable synchronized updates")
	}
}

func TestShouldSynchronizeTreatsFalseCIFlagAsDisabled(t *testing.T) {
	t.Setenv("CI", "false")

	if !shouldSynchronize(&ttyRecordingWriter{}) {
		t.Fatal("expected explicit CI=false to preserve synchronized updates")
	}
}

func TestShouldSynchronizeMatchesCaseSensitiveCIFlags(t *testing.T) {
	t.Setenv("CI", "FALSE")

	if shouldSynchronize(&ttyRecordingWriter{}) {
		t.Fatal("expected CI=FALSE to still count as CI")
	}
}

type helperFlushWriter struct {
	flushed bool
	err     error
}

func (writer *helperFlushWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (writer *helperFlushWriter) Flush() error {
	writer.flushed = true
	return writer.err
}

func TestFlushWriterUsesFlushMethod(t *testing.T) {
	writer := &helperFlushWriter{}

	if err := flushWriter(writer); err != nil {
		t.Fatalf("expected flush to succeed, got %v", err)
	}

	if !writer.flushed {
		t.Fatal("expected flushWriter to invoke Flush")
	}
}
