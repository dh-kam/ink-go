package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func App() *vdom.Node {
	selectedIndex, setSelectedIndex := ink.UseState(0)
	app := ink.UseApp()

	items := []string{"Option 1", "Option 2", "Option 3", "Option 4"}

	ink.UseInput(func(input string, key ink.InputKey) {
		if key.UpArrow {
			setSelectedIndex(func(prev interface{}) interface{} {
				p := prev.(int)
				if p == 0 {
					return len(items) - 1
				}
				return p - 1
			})
		}
		if key.DownArrow {
			setSelectedIndex(func(prev interface{}) interface{} {
				p := prev.(int)
				if p == len(items)-1 {
					return 0
				}
				return p + 1
			})
		}
		if input == "q" {
			app.Exit()
		}
	})

	idx := selectedIndex.(int)

	children := make([]*vdom.Node, 0, len(items))
	for i, item := range items {
		color := ""
		prefix := "  "
		if i == idx {
			color = "green"
			prefix = "> "
		}
		children = append(children, components.Text(vdom.Props{"color": color}, prefix+item))
	}

	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Text(vdom.Props{"color": "yellow"}, "--- Select Input Test ---"),
		components.Text("Use arrow keys to move, 'q' to quit."),
		components.Box(vdom.Props{"flexDirection": "column", "marginTop": 1.0},
			children...,
		),
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
