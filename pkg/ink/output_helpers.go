package ink

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dh-kam/goink.go/pkg/terminal"
)

const (
	showCursorEscape    = "\x1b[?25h"
	hideCursorEscape    = "\x1b[?25l"
	bsu                 = "\x1b[?2026h"
	esu                 = "\x1b[?2026l"
	clearTerminalEscape = "\x1b[2J\x1b[3J\x1b[H"
)

type ttyAwareWriter interface {
	IsTTY() bool
}

type columnsAwareWriter interface {
	Columns() int
}

type rowsAwareWriter interface {
	Rows() int
}

type resizeSubscriber interface {
	SubscribeResize(func()) func()
}

type inputSubscriber interface {
	SubscribeInput(func(string)) func()
}

func shouldSynchronize(writer io.Writer) bool {
	if writer == nil {
		return false
	}

	if isCIEnvironment() {
		return false
	}

	if ttyWriter, ok := writer.(ttyAwareWriter); ok {
		if ttyWriter.IsTTY() {
			return true
		}
	}

	fd, ok := streamFD(writer)
	if !ok {
		return false
	}

	return terminal.IsTerminal(fd)
}

func isCIEnvironment() bool {
	return isTruthyEnvironmentFlag("CI") || isTruthyEnvironmentFlag("CONTINUOUS_INTEGRATION")
}

func isTruthyEnvironmentFlag(key string) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return false
	}

	switch value {
	case "0", "false":
		return false
	default:
		return true
	}
}

type flushableWriter interface {
	Flush() error
}

type waitableWriter interface {
	Wait() error
}

func flushWriter(writer io.Writer) error {
	if writer == nil {
		return nil
	}

	if flushable, ok := writer.(flushableWriter); ok {
		return flushable.Flush()
	}

	return nil
}

func waitWriter(writer io.Writer) error {
	if writer == nil {
		return nil
	}

	if waitable, ok := writer.(waitableWriter); ok {
		return waitable.Wait()
	}

	return nil
}

func terminalViewportHeight(writer io.Writer, fallback int) int {
	if writer == nil {
		return fallback
	}

	if rowsWriter, ok := writer.(rowsAwareWriter); ok {
		if rows := rowsWriter.Rows(); rows > 0 {
			return rows
		}
	}

	if fd, ok := streamFD(writer); ok {
		_, height, err := terminal.GetSize(fd)
		if err == nil && height > 0 {
			return height
		}
	}

	return fallback
}

func terminalViewportWidth(writer io.Writer, fallback int) int {
	if writer == nil {
		return fallback
	}

	if columnsWriter, ok := writer.(columnsAwareWriter); ok {
		if columns := columnsWriter.Columns(); columns > 0 {
			return columns
		}
	}

	if fd, ok := streamFD(writer); ok {
		width, _, err := terminal.GetSize(fd)
		if err == nil && width > 0 {
			return width
		}
	}

	return fallback
}

func terminalViewportSize(writer io.Writer, fallbackWidth, fallbackHeight int) (int, int) {
	return terminalViewportWidth(writer, fallbackWidth), terminalViewportHeight(writer, fallbackHeight)
}

func subscribeResize(writer io.Writer, handler func()) func() {
	if writer == nil || handler == nil {
		return nil
	}

	if subscriber, ok := writer.(resizeSubscriber); ok {
		return subscriber.SubscribeResize(handler)
	}

	return nil
}

func subscribeInput(reader io.Reader, handler func(string)) func() {
	if reader == nil || handler == nil {
		return nil
	}

	if subscriber, ok := reader.(inputSubscriber); ok {
		return subscriber.SubscribeInput(handler)
	}

	return nil
}

func cursorPositionChanged(a, b *CursorPosition) bool {
	if a == nil && b == nil {
		return false
	}

	if a == nil || b == nil {
		return true
	}

	return a.X != b.X || a.Y != b.Y
}

func visibleLineCount(output string) int {
	if output == "" {
		return 0
	}

	lines := strings.Split(output, "\n")
	if strings.HasSuffix(output, "\n") {
		return len(lines) - 1
	}

	return len(lines)
}

func outputLineCount(output string) int {
	if output == "" {
		return 0
	}

	return len(strings.Split(output, "\n"))
}

func splitOutputLines(output string) []string {
	if output == "" {
		return []string{}
	}

	return strings.Split(output, "\n")
}

func ensureTrailingNewline(output string) string {
	if output == "" {
		return "\n"
	}

	if strings.HasSuffix(output, "\n") {
		return output
	}

	return output + "\n"
}

func buildCursorSuffix(visibleLines int, position *CursorPosition) string {
	if position == nil {
		return ""
	}

	moveUp := visibleLines - position.Y
	return ansiCursorUp(moveUp) + ansiCursorTo(position.X) + showCursorEscape
}

func buildReturnToBottom(lineCount int, previousPosition *CursorPosition) string {
	if previousPosition == nil {
		return ""
	}

	down := lineCount - 1 - previousPosition.Y
	return ansiCursorDown(down) + ansiCursorTo(0)
}

func buildCursorOnlySequence(
	cursorWasShown bool,
	previousLineCount int,
	previousCursorPosition *CursorPosition,
	visibleLines int,
	cursorPosition *CursorPosition,
) string {
	prefix := ""
	if cursorWasShown {
		prefix = hideCursorEscape
	}

	return prefix +
		buildReturnToBottom(previousLineCount, previousCursorPosition) +
		buildCursorSuffix(visibleLines, cursorPosition)
}

func buildReturnToBottomPrefix(
	cursorWasShown bool,
	previousLineCount int,
	previousCursorPosition *CursorPosition,
) string {
	if !cursorWasShown {
		return ""
	}

	return hideCursorEscape + buildReturnToBottom(previousLineCount, previousCursorPosition)
}

func ansiCursorUp(lines int) string {
	if lines <= 0 {
		return ""
	}

	return fmt.Sprintf("\x1b[%dA", lines)
}

func ansiCursorDown(lines int) string {
	if lines <= 0 {
		return ""
	}

	return fmt.Sprintf("\x1b[%dB", lines)
}

func ansiCursorTo(column int) string {
	if column < 0 {
		column = 0
	}

	return fmt.Sprintf("\x1b[%dG", column+1)
}

func ansiCursorNextLine() string {
	return "\x1b[E"
}

func ansiEraseEndLine() string {
	return "\x1b[K"
}

func ansiEraseLines(lineCount int) string {
	if lineCount <= 0 {
		return ""
	}

	var builder strings.Builder
	for index := 0; index < lineCount; index++ {
		builder.WriteString("\x1b[2K")
		if index < lineCount-1 {
			builder.WriteString("\x1b[1A")
		}
	}

	builder.WriteString(ansiCursorTo(0))
	return builder.String()
}
