package renderer

import (
	"io"
	"sync"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

// Instance is the handle returned by Render. It accumulates every frame
// produced (initial + each Rerender) so tests can assert on the full
// history. Methods are safe to call from multiple goroutines.
type Instance struct {
	mu      sync.Mutex
	frames  []string
	cleaned bool
	render  RenderFunc
	stdin   *fakeStdin     // nil unless WithStdin was provided
	stdout  *captureWriter // nil unless WithStdoutCapture was provided
	stderr  *captureWriter // nil unless WithStderrCapture was provided
}

// Stdout returns the writable stdout capture handle, or nil when
// WithStdoutCapture was not supplied to Render. Each Write to the returned
// writer becomes a single frame retrievable via StdoutFrames.
func (i *Instance) Stdout() io.Writer {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stdout == nil {
		return nil
	}
	return i.stdout
}

// Stderr returns the writable stderr capture handle, or nil when
// WithStderrCapture was not supplied to Render.
func (i *Instance) Stderr() io.Writer {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stderr == nil {
		return nil
	}
	return i.stderr
}

// StdoutFrames returns a defensive copy of every chunk written to the
// stdout capture so far. Returns nil when WithStdoutCapture was not
// configured. Each frame corresponds to one Write call (write-unit, not
// flush-unit).
func (i *Instance) StdoutFrames() []string {
	i.mu.Lock()
	cw := i.stdout
	i.mu.Unlock()
	if cw == nil {
		return nil
	}
	return cw.snapshot()
}

// StderrFrames returns a defensive copy of every chunk written to the
// stderr capture so far. Returns nil when WithStderrCapture was not
// configured.
func (i *Instance) StderrFrames() []string {
	i.mu.Lock()
	cw := i.stderr
	i.mu.Unlock()
	if cw == nil {
		return nil
	}
	return cw.snapshot()
}

// Stdin returns the writable stdin handle, or nil when WithStdin was not
// supplied to Render. Tests can call Write/WriteString to push bytes to
// any subscriber registered via SubscribeInput.
func (i *Instance) Stdin() io.Writer {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stdin == nil {
		return nil
	}
	return i.stdin
}

// LastFrame returns the most recent frame, or empty string if none.
func (i *Instance) LastFrame() string {
	i.mu.Lock()
	defer i.mu.Unlock()
	if len(i.frames) == 0 {
		return ""
	}
	return i.frames[len(i.frames)-1]
}

// Frames returns a defensive copy of every frame captured so far.
func (i *Instance) Frames() []string {
	i.mu.Lock()
	defer i.mu.Unlock()
	out := make([]string, len(i.frames))
	copy(out, i.frames)
	return out
}

// FrameCount reports how many frames have been recorded.
func (i *Instance) FrameCount() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return len(i.frames)
}

// Rerender renders node again and appends the new frame. No-op after Cleanup.
func (i *Instance) Rerender(node *vdom.Node) {
	i.mu.Lock()
	if i.cleaned {
		i.mu.Unlock()
		return
	}
	render := i.render
	i.mu.Unlock()

	// Render outside the lock — user RenderFuncs may be expensive and we
	// don't want to block other readers.
	frame := render(node)
	i.appendFrame(frame)
}

// Cleanup marks the instance as torn down and discards captured frames.
// Idempotent.
func (i *Instance) Cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.cleaned = true
	i.frames = nil
	if i.stdin != nil {
		i.stdin.close()
	}
	if i.stdout != nil {
		i.stdout.close()
	}
	if i.stderr != nil {
		i.stderr.close()
	}
}

func (i *Instance) appendFrame(frame string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.cleaned {
		return
	}
	i.frames = append(i.frames, frame)
}
