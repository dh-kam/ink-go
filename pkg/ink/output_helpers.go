package ink

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/dh-kam/ink-go/internal/ttyinput"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/utils"
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

func shouldManageCursor(writer io.Writer) bool {
	if writer == nil {
		return false
	}

	if ttyWriter, ok := writer.(ttyAwareWriter); ok {
		return ttyWriter.IsTTY()
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

	fd, ok := streamFD(writer)
	if !ok || !terminal.IsTerminal(fd) {
		return nil
	}

	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signals, syscall.SIGWINCH)

	go func() {
		for {
			select {
			case <-signals:
				handler()
			case <-done:
				return
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			signal.Stop(signals)
			close(done)
		})
	}

}

func subscribeInput(reader io.Reader, handler func(string)) func() {
	if reader == nil || handler == nil {
		return nil
	}

	if subscriber, ok := reader.(inputSubscriber); ok {
		return subscriber.SubscribeInput(handler)
	}

	fd, ok := streamFD(reader)
	if ok && terminal.IsTerminal(fd) {
		done := make(chan struct{})
		go func() {
			buf := make([]byte, 1024)
			decoder := ttyinput.UTF8Decoder{}
			for {
				select {
				case <-done:
					return
				default:
					n, err := reader.Read(buf)
					if n > 0 {
						if input := decoder.Write(buf[:n]); input != "" {
							handler(input)
						}
					}
					if err != nil {
						if input := decoder.Flush(); input != "" {
							handler(input)
						}
						return
					}
				}
			}
		}()
		var once sync.Once
		return func() {
			once.Do(func() {
				close(done)
			})
		}
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
	if column <= 0 {
		return "\x1b[G"
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

// containsANSIEscape reports whether s contains any ANSI control sequence
// introducer. Used to gate the column-level dirty-rect optimization on
// plain text only — any embedded escape (color/style) would invalidate
// naïve column counting against the previous line.
func containsANSIEscape(s string) bool {
	return strings.IndexByte(s, '\x1b') >= 0
}

// commonPlainPrefixWidth returns (visibleColumns, ok) — the number of
// terminal columns occupied by the longest grapheme-cluster prefix shared
// by previous and next, plus a bool reporting whether both lines are plain
// (ANSI-free). When ok is false the caller must fall back to a full-line
// rewrite. A leading-cluster mismatch returns (0, true): the caller still
// emits eraseEndLine + tail starting at column 0.
//
// The optimization is intentionally narrow: when an entire line changes
// (e.g. line index is brand new) the LCP is empty and the result is
// equivalent to the prior cursorTo(0) + line path. The wins come from
// localized changes (counters incrementing, spinners advancing) where
// most of the line is unchanged.
func commonPlainPrefixWidth(previous, next string) (int, bool) {
	if containsANSIEscape(previous) || containsANSIEscape(next) {
		return 0, false
	}

	prevClusters := utils.GraphemeClusters(previous)
	nextClusters := utils.GraphemeClusters(next)
	limit := len(prevClusters)
	if len(nextClusters) < limit {
		limit = len(nextClusters)
	}

	cols := 0
	for i := 0; i < limit; i++ {
		if prevClusters[i] != nextClusters[i] {
			break
		}
		cols += utils.StringWidth(prevClusters[i])
	}
	return cols, true
}

// commonByteOffsetForWidth walks s's grapheme clusters and returns the
// byte offset at which the accumulated visible width first reaches
// targetWidth. Used to slice off the divergent tail from next without
// re-tokenizing.
func commonByteOffsetForWidth(s string, targetWidth int) int {
	if targetWidth <= 0 {
		return 0
	}

	cols := 0
	offset := 0
	for _, cluster := range utils.GraphemeClusters(s) {
		if cols >= targetWidth {
			break
		}
		cols += utils.StringWidth(cluster)
		offset += len(cluster)
	}
	return offset
}
