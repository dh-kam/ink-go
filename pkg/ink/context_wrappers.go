package ink

import (
	gocontext "github.com/dh-kam/ink-go/pkg/context"
	"github.com/dh-kam/ink-go/pkg/hooks"
)

// Context is the Ink-level alias for context.Context, kept as a generic type
// alias so callers do not need to import the underlying package.
type Context[T any] = gocontext.Context[T]

// NewContext constructs a Context with the supplied default value.
func NewContext[T any](defaultValue T) *Context[T] {
	return gocontext.New(defaultValue)
}

// UseContext returns the current value of c for the rendering component.
// Must be called during a render — panics outside one, mirroring the other
// public hooks.
func UseContext[T any](c *Context[T]) T {
	return hooks.UseContext(requireHooksContext("UseContext"), c)
}
