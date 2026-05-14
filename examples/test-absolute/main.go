package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func section(label string, child *vdom.Node) *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Text(vdom.Props{"color": "cyan"}, label),
		child,
	)
}

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "yellow"}, "--- Absolute Position Test ---"),
		section("absolute overlays first column",
			components.Box(vdom.Props{"flexDirection": "column"},
				components.Text("Line 1"),
				components.Box(vdom.Props{"position": "absolute"},
					components.Text("ABS"),
				),
				components.Text("Line 2"),
			),
		),
		section("absolute does not consume row width",
			components.Box(vdom.Props{"width": 8.0},
				components.Box(vdom.Props{"flexGrow": 1.0},
					components.Text("A"),
				),
				components.Box(vdom.Props{"position": "absolute"},
					components.Text("Z"),
				),
				components.Box(vdom.Props{"flexGrow": 1.0},
					components.Text("B"),
				),
			),
		),
		section("absolute with margin",
			components.Box(vdom.Props{"width": 10.0, "height": 3.0},
				components.Text("Base"),
				components.Box(vdom.Props{"position": "absolute", "marginLeft": 4.0, "marginTop": 1.0},
					components.Text("Z"),
				),
			),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
