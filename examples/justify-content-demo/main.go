package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func main() {
	fmt.Println(renderJustifyContentDemo())
}

func renderJustifyContentDemo() string {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		justifyRow("flex-start", "flex-start"),
		justifyRow("flex-end", "flex-end"),
		justifyRow("center", "center"),
		justifyRow("space-around", "space-around"),
		justifyRow("space-between", "space-between"),
		justifyRow("space-evenly", "space-evenly"),
	)

	return renderer.RenderWithLayout(root, 100, 20)
}

func justifyRow(justifyContent string, label string) *vdom.Node {
	return components.Box(nil,
		components.Text("["),
		components.Box(vdom.Props{
			"justifyContent": justifyContent,
			"width":          20.0,
			"height":         1.0,
		},
			components.Text("X"),
			components.Text("Y"),
		),
		components.Text("] "+label),
	)
}
