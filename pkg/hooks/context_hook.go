package hooks

import (
	gocontext "github.com/dh-kam/goink.go/pkg/context"
)

// UseContext returns the current value of the supplied Context.
//
// Unlike UseState/UseEffect this hook consumes no slot in the hook Context —
// the value lives on the gocontext stack, so call ordering between renders
// does not matter and it is safe to call inside conditionals (still
// discouraged for clarity, but it will not corrupt other hook indices).
//
// The ctx *Context argument is accepted for API symmetry with the rest of
// the hook surface; it is intentionally unused.
func UseContext[T any](ctx *Context, c *gocontext.Context[T]) T {
	_ = ctx
	if c == nil {
		panic("UseContext: context must not be nil")
	}
	return c.Current()
}
