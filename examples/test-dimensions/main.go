package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "cyan"}, "--- Dimensions Test ---"),

		components.Text("Fixed Width (20) & Height (3):"),
		components.Box(vdom.Props{"width": 20.0, "height": 3.0, "borderStyle": "single", "backgroundColor": "green"},
			components.Text("Fixed"),
		),

		components.Text("Percentage Width (50%):"),
		components.Box(vdom.Props{"width": "50%", "borderStyle": "single", "backgroundColor": "blue"},
			components.Text("50% Width"),
		),

		components.Text("MinWidth (30):"),
		components.Box(vdom.Props{"minWidth": 30.0, "borderStyle": "single", "backgroundColor": "red"},
			components.Text("Short"),
		),

		components.Text("MinHeight (5) with short content:"),
		components.Box(vdom.Props{"minHeight": 5.0, "borderStyle": "single"},
			components.Text("MinHeight 5"),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	out, _ := app.RenderSplitOnce()
	fmt.Println(out)
}
