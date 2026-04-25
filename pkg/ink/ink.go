package ink

import (
	"github.com/dh-kam/goink.go/internal/renderer"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// ComponentFunc is a function that returns a virtual DOM node
type ComponentFunc func() *vdom.Node

// RenderToString renders a component to a string
func RenderToString(node *vdom.Node) string {
	// Default terminal size (can be made configurable later)
	const defaultWidth = 80
	const defaultHeight = 24

	return renderer.Render(node, defaultWidth, defaultHeight)
}

// Render renders a component function to a string
func Render(component ComponentFunc) string {
	node := component()
	return RenderToString(node)
}
