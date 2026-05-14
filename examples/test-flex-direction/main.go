package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text("--- Flex Direction Test ---"),

		components.Text(vdom.Props{"color": "green"}, "Row (default):"),
		components.Box(vdom.Props{"flexDirection": "row"},
			components.Text("A"),
			components.Text("B"),
			components.Text("C"),
		),

		components.Text(vdom.Props{"color": "green"}, "Column:"),
		components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("B"),
			components.Text("C"),
		),

		components.Text(vdom.Props{"color": "green"}, "Row-Reverse:"),
		components.Box(vdom.Props{"flexDirection": "row-reverse", "width": 10.0},
			components.Text("A"),
			components.Text("B"),
			components.Text("C"),
		),

		components.Text(vdom.Props{"color": "green"}, "Column-Reverse:"),
		components.Box(vdom.Props{"flexDirection": "column-reverse", "height": 3.0},
			components.Text("A"),
			components.Text("B"),
			components.Text("C"),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	out, _ := app.RenderSplitOnce()
	fmt.Println(out)
}
