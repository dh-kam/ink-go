package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Text("Welcome to Ink with Colors!"),
		ink.Text("\n"),
		ink.Text("\n"),
		ink.Text(vdom.Props{"color": "red"}, "Red Text"),
		ink.Text("\n"),
		ink.Text(vdom.Props{"color": "green"}, "Green Text"),
		ink.Text("\n"),
		ink.Text(vdom.Props{"color": "blue"}, "Blue Text"),
		ink.Text("\n"),
		ink.Text("\n"),
		ink.Text(vdom.Props{"bold": true}, "Bold Text"),
		ink.Text("\n"),
		ink.Text(vdom.Props{"color": "red", "bold": true}, "Bold Red Text"),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
