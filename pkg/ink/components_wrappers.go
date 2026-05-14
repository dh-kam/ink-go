package ink

import (
	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TransformFunc transforms rendered text output.
type TransformFunc = components.TransformFunc

// Box creates a container element.
func Box(props vdom.Props, children ...*vdom.Node) *vdom.Node {
	return components.Box(props, children...)
}

// Text creates a text element.
func Text(args ...any) *vdom.Node {
	return components.Text(args...)
}

// Static creates a static output element.
func Static(args ...any) *vdom.Node {
	return components.Static(args...)
}

// Transform creates a text-like node that transforms its rendered output.
func Transform(transform TransformFunc, args ...any) *vdom.Node {
	return components.Transform(transform, args...)
}

// Newline creates a newline text node.
func Newline(count ...int) *vdom.Node {
	return components.Newline(count...)
}

// Spacer creates a flexible box that expands along the main axis.
func Spacer() *vdom.Node {
	return components.Spacer()
}
