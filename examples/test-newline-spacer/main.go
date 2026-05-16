package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func section(label string, child *vdom.Node) *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Text(vdom.Props{"color": "cyan"}, label),
		child,
	)
}

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "yellow"}, "--- Newline and Spacer Test ---"),
		section("newline",
			components.Text("Hello", components.Newline(), "World"),
		),
		section("multiple newlines",
			components.Text("Hello", components.Newline(2), "World"),
		),
		section("horizontal spacer",
			components.Box(vdom.Props{"width": 20.0},
				components.Text("Left"),
				components.Spacer(),
				components.Text("Right"),
			),
		),
		section("vertical spacer",
			components.Box(vdom.Props{"flexDirection": "column", "height": 6.0},
				components.Text("Top"),
				components.Spacer(),
				components.Text("Bottom"),
			),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
