package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func fixedWidthRow(text string) *vdom.Node {
	return components.Box(nil,
		components.Box(vdom.Props{"width": 2.0},
			components.Text(text),
		),
		components.Text("|"),
	)
}

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column"},
		fixedWidthRow("🍔"),
		fixedWidthRow("⏳"),
		fixedWidthRow("👨‍👩‍👧‍👦"),
		fixedWidthRow("日"),
		fixedWidthRow("e\u0301"),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
