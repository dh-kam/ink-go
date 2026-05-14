package tuitest

import (
	"strings"
	"sync"
)

type Recorder struct {
	mu     sync.Mutex
	writes []string
}

func (recorder *Recorder) Write(data []byte) (int, error) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	recorder.writes = append(recorder.writes, string(data))
	return len(data), nil
}

func (recorder *Recorder) WriteCount() int {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	return len(recorder.writes)
}

func (recorder *Recorder) WritesFrom(index int) []string {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	if index < 0 {
		index = 0
	}
	if index > len(recorder.writes) {
		index = len(recorder.writes)
	}

	writes := make([]string, len(recorder.writes[index:]))
	copy(writes, recorder.writes[index:])
	return writes
}

func (recorder *Recorder) Capture(mode string, trimTrailingNewline bool) string {
	raw := recorder.lastWrite()
	var captured string
	switch mode {
	case CaptureANSI:
		captured = raw
	case CaptureANSIEscape:
		captured = escapeForFixture(raw)
	default:
		captured = stripANSI(raw)
	}

	if trimTrailingNewline {
		captured = strings.TrimSuffix(captured, "\n")
	}
	return captured
}

func (recorder *Recorder) Preview(mode string, trimTrailingNewline bool) string {
	switch mode {
	case CaptureANSI, CaptureANSIEscape:
		return recorder.lastWrite()
	default:
		return recorder.Capture(CapturePlain, trimTrailingNewline)
	}
}

func (recorder *Recorder) lastWrite() string {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	if len(recorder.writes) == 0 {
		return ""
	}
	return recorder.writes[len(recorder.writes)-1]
}
