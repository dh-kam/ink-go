package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	value, setValue := ink.UseState("")
	app := ink.UseApp()

	ink.UseInput(func(input string, key ink.InputKey) {
		if key.Escape {
			app.Exit()
		}
		if key.Return {
			return
		} else if key.Backspace || key.Delete {
			setValue(func(prev interface{}) interface{} {
				p := prev.(string)
				if len(p) > 0 {
					return p[:len(p)-1]
				}
				return p
			})
		} else if input != "" {
			setValue(func(prev interface{}) interface{} {
				return prev.(string) + input
			})
		}
	})

	val := value.(string)

	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "yellow"}, "--- Text Input Test ---"),
		components.Text("Type something. Press ESC to quit."),
		components.Box(vdom.Props{"borderStyle": "single", "paddingX": 1.0, "width": 40.0},
			components.Text(fmt.Sprintf("Value: %s", val)),
			components.Text(vdom.Props{"color": "blue"}, "|"),
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
