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
		components.Text(vdom.Props{"color": "yellow"}, "--- Box Gap Test ---"),
		section("gap + wrap",
			components.Box(vdom.Props{"gap": 1.0, "width": 3.0, "flexWrap": "wrap"},
				components.Text("A"),
				components.Text("B"),
				components.Text("C"),
			),
		),
		section("column gap",
			components.Box(vdom.Props{"gap": 1.0},
				components.Text("A"),
				components.Text("B"),
			),
		),
		section("row gap",
			components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
				components.Text("A"),
				components.Text("B"),
			),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
