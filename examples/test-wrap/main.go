package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func App() *vdom.Node {
	items := make([]*vdom.Node, 0, 15)
	for i := 1; i <= 15; i++ {
		items = append(items, components.Box(vdom.Props{
			"margin":          1.0,
			"backgroundColor": "blue",
		}, components.Text(fmt.Sprintf(" #%d ", i))))
	}

	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "magenta"}, "--- Flex Wrap Test ---"),

		components.Text("flexWrap=\"wrap\" with many items (width=40):"),
		components.Box(vdom.Props{
			"flexWrap":    "wrap",
			"width":       40.0,
			"borderStyle": "single",
		}, items...),
	)
}

func main() {
	app := ink.NewApp(App)
	out, _ := app.RenderSplitOnce()
	fmt.Println(out)
}
