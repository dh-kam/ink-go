package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "yellow"}, "--- Screen Reader (ARIA) Test ---"),
		components.Text("Enable screen reader mode (INK_SCREEN_READER=true) to see alternate text."),

		components.Box(vdom.Props{"aria-label": "Aria Container"},
			components.Text(vdom.Props{"aria-label": "Accessible Hello"}, "Hi"),
		),

		components.Box(nil,
			components.Text("Standard Text"),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Print(app.RenderOnce())
}
