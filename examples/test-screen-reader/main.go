package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"aria-role": "list", "flexDirection": "column"},
			components.Text("Select a color:"),
			components.Box(vdom.Props{"aria-label": "1. Red", "aria-role": "listitem"},
				components.Text("Red"),
			),
			components.Box(vdom.Props{"aria-label": "2. Green", "aria-role": "listitem", "aria-state": vdom.Props{"selected": true}},
				components.Text("Green"),
			),
			components.Box(vdom.Props{"aria-label": "3. Blue", "aria-role": "listitem"},
				components.Text("Blue"),
			),
		),
		components.Box(vdom.Props{"aria-role": "checkbox", "aria-state": vdom.Props{"checked": true}},
			components.Text("Accept terms"),
		),
		components.Box(vdom.Props{"aria-role": "button", "aria-state": vdom.Props{"disabled": true}},
			components.Text("Submit"),
		),
		components.Box(vdom.Props{"aria-role": "combobox", "aria-state": vdom.Props{"expanded": true}},
			components.Text("Select"),
		),
		components.Box(vdom.Props{"aria-role": "textbox", "aria-state": vdom.Props{"readonly": true}},
			components.Text("Email"),
		),
		components.Box(vdom.Props{"aria-role": "listbox", "aria-state": vdom.Props{"multiselectable": true}, "flexDirection": "column"},
			components.Box(vdom.Props{"aria-role": "option", "aria-state": vdom.Props{"selected": true}},
				components.Text("Option 1"),
			),
			components.Box(vdom.Props{"aria-role": "option", "aria-state": vdom.Props{"selected": false}},
				components.Text("Option 2"),
			),
			components.Box(vdom.Props{"aria-role": "option", "aria-state": vdom.Props{"selected": true}},
				components.Text("Option 3"),
			),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Print(app.RenderOnce())
}
