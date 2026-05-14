package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "cyan"}, "--- Text Wrap & Truncation Test ---"),

		components.Text("wrap=\"wrap\" (width 20):"),
		components.Box(vdom.Props{"width": 20.0, "borderStyle": "single"},
			components.Text(vdom.Props{"wrap": "wrap"}, "This is a very long text that should wrap inside the box."),
		),

		components.Text("wrap=\"truncate-end\" (width 20):"),
		components.Box(vdom.Props{"width": 20.0, "borderStyle": "single"},
			components.Text(vdom.Props{"wrap": "truncate-end"}, "This is a very long text that should be truncated at the end."),
		),

		components.Text("wrap=\"truncate-middle\" (width 20):"),
		components.Box(vdom.Props{"width": 20.0, "borderStyle": "single"},
			components.Text(vdom.Props{"wrap": "truncate-middle"}, "This is a very long text that should be truncated in the middle."),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	out, _ := app.RenderSplitOnce()
	fmt.Println(out)
}
