package main

import (
	"fmt"
	"time"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func App() *vdom.Node {
	app := ink.UseApp()

	ink.UseEffect(func() func() {
		timer := time.AfterFunc(2*time.Second, func() {
			app.Exit()
		})
		return func() {
			timer.Stop()
		}
	}, []any{})

	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "red"}, "--- Manual Exit Test ---"),
		components.Text("This program will exit automatically in 2 seconds..."),
	)
}

func main() {
	instance, err := ink.Mount(App)
	if err != nil {
		fmt.Printf("Error mounting app: %v\n", err)
		return
	}

	instance.WaitUntilExit()
}
