package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func main() {
	fmt.Println(renderBordersDemo())
}

func renderBordersDemo() string {
	root := components.Box(
		vdom.Props{
			"flexDirection": "column",
			"padding":       2.0,
		},
		components.Box(nil,
			borderBox("single", "single", vdom.Props{"marginRight": 2.0}),
			borderBox("double", "double", vdom.Props{"marginRight": 2.0}),
			borderBox("round", "round", vdom.Props{"marginRight": 2.0}),
			borderBox("bold", "bold", nil),
		),
		components.Box(vdom.Props{"marginTop": 1.0},
			borderBox("singleDouble", "singleDouble", vdom.Props{"marginRight": 2.0}),
			borderBox("doubleSingle", "doubleSingle", vdom.Props{"marginRight": 2.0}),
			borderBox("classic", "classic", nil),
		),
	)

	return renderer.RenderWithLayout(root, 100, 20)
}

func borderBox(style string, label string, props vdom.Props) *vdom.Node {
	if props == nil {
		props = vdom.Props{}
	}

	props["borderStyle"] = style
	return components.Box(props, components.Text(label))
}
