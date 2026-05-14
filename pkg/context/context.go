// Package context implements React-style Context.Provider / useContext value
// propagation for Goink. A Context[T] holds a per-instance stack of values;
// the deepest active Provider supplies the value returned by Current() and
// UseContext().
//
// The stack is goroutine-safe via a Mutex. Push returns an idempotent closer
// that pops the slot it created (and any slots above it that survived a
// premature panic), so paired Push/closer calls stay balanced even under
// out-of-order recovery.
package context

import (
	"sync"
)

// Context propagates a value of type T down a render tree without prop
// drilling. Use New to construct one with a default value, then wrap
// rendering work in Provider (or Push/closer) to override it.
type Context[T any] struct {
	defaultValue T
	mu           sync.Mutex
	stack        []T
}

// New constructs a Context whose Current() value is defaultValue when no
// Provider is active.
func New[T any](defaultValue T) *Context[T] {
	return &Context[T]{defaultValue: defaultValue}
}

// Default returns the value Current() yields when the provider stack is empty.
func (c *Context[T]) Default() T {
	return c.defaultValue
}

// Current returns the deepest provided value, or Default() when no Provider
// is active.
func (c *Context[T]) Current() T {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n := len(c.stack); n > 0 {
		return c.stack[n-1]
	}
	return c.defaultValue
}

// Depth reports the current provider stack depth. Useful for tests and
// debugging; not part of the React API.
func (c *Context[T]) Depth() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.stack)
}

// Push installs value as the top of the provider stack and returns a closer
// that pops it. The closer is safe to call multiple times (subsequent calls
// are no-ops via sync.Once) and tolerates out-of-order pops by truncating to
// the slot it created — useful when a panic skipped intermediate closers.
func (c *Context[T]) Push(value T) func() {
	c.mu.Lock()
	c.stack = append(c.stack, value)
	myIdx := len(c.stack) - 1
	c.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			if myIdx < len(c.stack) {
				// Truncate to my slot — drops anything still above me too.
				var zero T
				for i := myIdx; i < len(c.stack); i++ {
					c.stack[i] = zero
				}
				c.stack = c.stack[:myIdx]
			}
		})
	}
}

// Provider runs fn with value installed as the current Context value. The
// value is automatically popped when fn returns, even if fn panics.
func (c *Context[T]) Provider(value T, fn func()) {
	if fn == nil {
		panic("context.Provider: fn must not be nil")
	}
	closer := c.Push(value)
	defer closer()
	fn()
}

// Reset clears the provider stack. Intended for tests that share a Context
// across cases.
func (c *Context[T]) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.stack {
		var zero T
		c.stack[i] = zero
	}
	c.stack = c.stack[:0]
}
