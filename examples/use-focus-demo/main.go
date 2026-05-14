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

func UseFocusDemo() *vdom.Node {
	return ink.Box(vdom.Props{"flexDirection": "column", "padding": 1},
		ink.Box(vdom.Props{"marginBottom": 1},
			ink.Text("Press Tab to focus next element, Shift+Tab to focus previous element, Esc to reset focus."),
		),
		focusItem("First"),
		focusItem("Second"),
		focusItem("Third"),
	)
}

func focusItem(label string) *vdom.Node {
	isFocused, _, _ := ink.UseFocus()
	if isFocused() {
		return ink.Text(label+" ", ink.Text(vdom.Props{"color": "green"}, "(focused)"))
	}
	return ink.Text(label)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(UseFocusDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(UseFocusDemo, ink.RenderOptions{})
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
