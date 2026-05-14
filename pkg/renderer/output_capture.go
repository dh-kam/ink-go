package renderer

import (
	"io"
	"sync"
)

// WithStdoutCapture attaches a stdout capture writer to the Instance. Tests
// can later call inst.Stdout() to obtain an io.Writer for components that
// write directly to stdout (bypassing the vdom render pipeline), and
// inst.StdoutFrames() to retrieve the accumulated buffer of writes.
//
// This is the renderer-side analogue of ink-testing-library's stdout
// capture — every Write call is recorded as a single frame, mirroring how
// ink-testing-library exposes incremental output for assertions.
func WithStdoutCapture() Option {
	return func(i *Instance) {
		i.stdout = newCaptureWriter()
	}
}

// WithStderrCapture attaches a stderr capture writer to the Instance.
// Mirrors WithStdoutCapture for the standard-error stream.
func WithStderrCapture() Option {
	return func(i *Instance) {
		i.stderr = newCaptureWriter()
	}
}

// captureWriter is a thread-safe io.Writer that records every Write call
// verbatim into an internal []string buffer. Each Write is appended as one
// frame (write-unit semantics) — callers do not need to flush. After
// close(), further writes return io.ErrClosedPipe.
type captureWriter struct {
	mu     sync.Mutex
	frames []string
	closed bool
}

func newCaptureWriter() *captureWriter {
	return &captureWriter{}
}

// Write appends p as a new frame and returns len(p). After close, returns
// io.ErrClosedPipe so tests can assert the Cleanup contract.
func (c *captureWriter) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	// Defensive copy of the byte slice so the caller can recycle their
	// buffer without corrupting the recorded frame.
	clone := make([]byte, len(p))
	copy(clone, p)
	c.frames = append(c.frames, string(clone))
	return len(p), nil
}

// snapshot returns a defensive copy of the current frame buffer.
func (c *captureWriter) snapshot() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.frames))
	copy(out, c.frames)
	return out
}

func (c *captureWriter) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}
