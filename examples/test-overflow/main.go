package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "yellow"}, "--- Overflow Test ---"),

		components.Text("overflow=\"hidden\" (Box 10x5, Child 10x10 at 5,2):"),
		components.Box(vdom.Props{"width": 10.0, "height": 5.0, "overflow": "hidden", "borderStyle": "single"},
			components.Box(vdom.Props{"marginTop": 2.0, "marginLeft": 5.0, "width": 10.0, "height": 10.0, "backgroundColor": "red"},
				components.Text("OVR"),
			),
		),

		components.Text("overflow=\"visible\" (default, Box 10x5):"),
		components.Box(vdom.Props{"width": 10.0, "height": 5.0, "borderStyle": "single"},
			components.Box(vdom.Props{"marginTop": 2.0, "marginLeft": 5.0, "width": 10.0, "height": 10.0, "backgroundColor": "blue"},
				components.Text("VIS"),
			),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	out, _ := app.RenderSplitOnce()
	fmt.Println(out)
}
