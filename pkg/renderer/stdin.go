package renderer

import (
	"errors"
	"io"
	"sync"
)

// WithStdin attaches a fake stdin to the Instance. Tests can later call
// inst.Stdin().Write([]byte("...")) to push bytes; subscribers registered
// via inst.SubscribeInput receive them synchronously.
//
// This is the renderer-side analogue of ink-testing-library's `stdin`
// helper — it lets tests drive UseInput / UseMouse handlers without a
// real TTY.
func WithStdin() Option {
	return func(i *Instance) {
		i.stdin = newFakeStdin()
	}
}

// SubscribeInput registers fn to receive every byte slice written to the
// fake stdin. Returns an unsubscribe func. Returns nil unsubscribe when
// the Instance has no stdin attached.
func (i *Instance) SubscribeInput(fn func([]byte)) func() {
	i.mu.Lock()
	stdin := i.stdin
	i.mu.Unlock()
	if stdin == nil || fn == nil {
		return func() {}
	}
	return stdin.subscribe(fn)
}

// fakeStdin is a tiny io.Writer that fans every Write out to its
// subscribers. Closed instances reject further writes with io.ErrClosedPipe.
type fakeStdin struct {
	mu          sync.Mutex
	subscribers map[uint64]func([]byte)
	nextID      uint64
	closed      bool
}

func newFakeStdin() *fakeStdin {
	return &fakeStdin{subscribers: make(map[uint64]func([]byte))}
}

func (f *fakeStdin) Write(p []byte) (int, error) {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	subs := make([]func([]byte), 0, len(f.subscribers))
	for _, sub := range f.subscribers {
		subs = append(subs, sub)
	}
	f.mu.Unlock()

	// Defensive copy so subscribers can hold the slice without aliasing the
	// caller's buffer.
	clone := make([]byte, len(p))
	copy(clone, p)
	for _, sub := range subs {
		sub(clone)
	}
	return len(p), nil
}

func (f *fakeStdin) subscribe(fn func([]byte)) func() {
	f.mu.Lock()
	f.nextID++
	id := f.nextID
	f.subscribers[id] = fn
	f.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			f.mu.Lock()
			delete(f.subscribers, id)
			f.mu.Unlock()
		})
	}
}

func (f *fakeStdin) close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	f.subscribers = nil
}

// ErrStdinClosed is returned when writing to a Cleanup'd stdin.
var ErrStdinClosed = errors.New("renderer: stdin is closed")
