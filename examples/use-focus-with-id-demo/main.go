package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dh-kam/ink-go/internal/ttyinput"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func UseFocusWithIDDemo() *vdom.Node {
	manager := ink.UseFocusManager()

	ink.UseInput(func(input string, _ ink.InputKey) {
		switch input {
		case "1", "2", "3":
			manager.Focus(input)
		}
	})

	return ink.Box(vdom.Props{"flexDirection": "column", "padding": 1},
		ink.Box(vdom.Props{"marginBottom": 1},
			ink.Text("Press Tab to focus next element, Shift+Tab to focus previous element, Esc to reset focus."),
		),
		focusWithIDItem("1", "Press 1 to focus"),
		focusWithIDItem("2", "Press 2 to focus"),
		focusWithIDItem("3", "Press 3 to focus"),
	)
}

func focusWithIDItem(id string, label string) *vdom.Node {
	state := ink.UseFocusOpts(ink.FocusOptions{ID: id})
	if state.IsFocused {
		return ink.Text(label+" ", ink.Text(vdom.Props{"color": "green"}, "(focused)"))
	}
	return ink.Text(label)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(UseFocusWithIDDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(UseFocusWithIDDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := runInputLoop(instance); err != nil {
		fmt.Println(err)
	}
}

func runInputLoop(instance *ink.Instance) error {
	return ttyinput.Run(os.Stdin, instance.HandleInput, func(input string) bool {
		return strings.Contains(input, "\x03")
	})
}
