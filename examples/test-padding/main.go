package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text("--- Padding & Margin Test ---"),
		components.Box(vdom.Props{"backgroundColor": "green", "padding": 1.0},
			components.Text("Padding 1"),
		),
		components.Box(vdom.Props{"backgroundColor": "blue", "paddingX": 2.0, "paddingY": 1.0},
			components.Text("Padding X:2 Y:1"),
		),
		components.Box(vdom.Props{"backgroundColor": "red", "marginLeft": 2.0, "marginTop": 1.0},
			components.Text("Margin L:2 T:1"),
		),
		components.Box(vdom.Props{"borderStyle": "single", "padding": 1.0, "margin": 1.0},
			components.Text("Border + Padding 1 + Margin 1"),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	out, _ := app.RenderSplitOnce()
	fmt.Println(out)
}
