// Package renderer provides ink-testing-library-equivalent helpers: render
// a vdom tree in isolation, capture every frame produced (including
// re-renders), strip ANSI for textual assertions, and snapshot-test against
// golden files under testdata/__snapshots__/.
//
// Default renderer is ink.RenderToString; inject a different one via
// WithRenderer for tests that want to substitute a stub.
package renderer

import (
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// RenderFunc is the signature this package uses to materialize a vdom node
// into a printable frame.
type RenderFunc func(*vdom.Node) string

// Option configures a test render. Apply via Render(node, opts...).
type Option func(*Instance)

// WithRenderer overrides the default ink.RenderToString with a custom
// RenderFunc. Useful for tests that want to assert on the raw vdom tree
// or inject deterministic output.
func WithRenderer(fn RenderFunc) Option {
	return func(i *Instance) {
		if fn != nil {
			i.render = fn
		}
	}
}

// Render renders node and returns an Instance you can query for frames or
// re-render via Rerender. Always call Cleanup when done (defer is fine) so
// further Rerender calls are rejected — protects test goroutines from
// racing on a discarded Instance.
func Render(node *vdom.Node, opts ...Option) *Instance {
	inst := &Instance{render: ink.RenderToString}
	for _, opt := range opts {
		opt(inst)
	}
	inst.appendFrame(inst.render(node))
	return inst
}
