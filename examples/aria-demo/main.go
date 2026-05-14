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

func AriaExample() *vdom.Node {
	checkedValue, setChecked := ink.UseState(false)
	checked := checkedValue.(bool)

	ink.UseInput(func(input string, _ ink.InputKey) {
		if input == " " {
			setChecked(!checked)
		}
	})

	mark := "[ ]"
	if checked {
		mark = "[x]"
	}

	return renderAriaDemo(mark)
}

func renderAriaDemo(mark string) *vdom.Node {
	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Text("Press spacebar to toggle the checkbox. This example is best experienced with a screen reader."),
		ink.Box(vdom.Props{"marginTop": 1.0},
			ink.Box(vdom.Props{"aria-role": "checkbox", "aria-state": vdom.Props{"checked": mark == "[x]"}},
				ink.Text(mark),
			),
		),
		ink.Box(vdom.Props{"marginTop": 1.0},
			ink.Text(vdom.Props{"aria-hidden": true}, "This text is hidden from screen readers."),
		),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(AriaExample)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(AriaExample, ink.RenderOptions{})
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
