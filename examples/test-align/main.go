package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "yellow"}, "--- Align Items Test ---"),

		components.Text("alignItems=\"flex-start\":"),
		components.Box(vdom.Props{"alignItems": "flex-start", "height": 5.0, "borderStyle": "single"},
			components.Text("A"),
		),

		components.Text("alignItems=\"center\":"),
		components.Box(vdom.Props{"alignItems": "center", "height": 5.0, "borderStyle": "single"},
			components.Text("B"),
		),

		components.Text("alignItems=\"flex-end\":"),
		components.Box(vdom.Props{"alignItems": "flex-end", "height": 5.0, "borderStyle": "single"},
			components.Text("C"),
		),

		components.Text("alignSelf=\"flex-end\" inside alignItems=\"flex-start\":"),
		components.Box(vdom.Props{"alignItems": "flex-start", "height": 5.0, "borderStyle": "single"},
			components.Text("A"),
			components.Box(vdom.Props{"alignSelf": "flex-end"}, components.Text("B")),
			components.Text("C"),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	out, _ := app.RenderSplitOnce()
	fmt.Println(out)
}
