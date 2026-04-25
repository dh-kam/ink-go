package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// App is our main component
func App() *vdom.Node {
	return vdom.CreateElement("box", nil,
		components.Text("Hello, Goink!"),
	)
}

func main() {
	// For now, just render once
	output := ink.RenderToString(App())
	fmt.Println(output)
}
