package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func HelloWorldDemo() *vdom.Node {
	return ink.Box(vdom.Props{"padding": 1},
		ink.Text(vdom.Props{"color": "cyan"}, "Hello, Ink (Node.js)!"),
	)
}

func main() {
	app := ink.NewApp(HelloWorldDemo)
	fmt.Println(app.RenderOnce())
}
