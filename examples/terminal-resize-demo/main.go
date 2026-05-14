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

func TerminalResizeDemo() *vdom.Node {
	valueRaw, setValue := ink.UseState("")
	value := valueRaw.(string)

	ink.UseInput(func(input string, _ ink.InputKey) {
		switch input {
		case "\r":
			setValue("")
		case "\u007f", "\b":
			if value == "" {
				return
			}
			setValue(value[:len(value)-1])
		default:
			setValue(value + input)
		}
	})

	return ink.Box(vdom.Props{"flexDirection": "column", "padding": 1.0},
		ink.Text(vdom.Props{"bold": true, "color": "cyan"}, "=== Terminal Resize Test ==="),
		ink.Text("Type something and then resize your terminal (drag the edge or press Cmd/Ctrl -/+)"),
		ink.Text(fmt.Sprintf("Input: %q", value)),
		ink.Box(vdom.Props{"marginTop": 1.0},
			ink.Text(vdom.Props{"dimColor": true}, "Press Ctrl+C to exit"),
		),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(TerminalResizeDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(TerminalResizeDemo, ink.RenderOptions{})
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
