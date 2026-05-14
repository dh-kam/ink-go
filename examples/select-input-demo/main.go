package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dh-kam/goink.go/internal/ttyinput"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

var items = []string{"Red", "Green", "Blue", "Yellow", "Magenta", "Cyan"}

func SelectInput() *vdom.Node {
	selectedValue, setSelectedIndex := ink.UseState(0)
	selectedIndex := selectedValue.(int)
	isScreenReaderEnabled := ink.UseIsScreenReaderEnabled()

	ink.UseInput(func(input string, key ink.InputKey) {
		if key.UpArrow {
			if selectedIndex == 0 {
				setSelectedIndex(len(items) - 1)
			} else {
				setSelectedIndex(selectedIndex - 1)
			}
		}

		if key.DownArrow {
			if selectedIndex == len(items)-1 {
				setSelectedIndex(0)
			} else {
				setSelectedIndex(selectedIndex + 1)
			}
		}

		if isScreenReaderEnabled {
			var number int
			if _, err := fmt.Sscanf(input, "%d", &number); err == nil && number > 0 && number <= len(items) {
				setSelectedIndex(number - 1)
			}
		}
	})

	children := []*vdom.Node{
		ink.Text("Select a color:"),
	}
	for index, item := range items {
		isSelected := index == selectedIndex
		label := "  " + item
		if isSelected {
			label = "> " + item
		}

		props := vdom.Props{"aria-role": "listitem", "aria-state": vdom.Props{"selected": isSelected}}
		if isScreenReaderEnabled {
			props["aria-label"] = fmt.Sprintf("%d. %s", index+1, item)
		}

		textProps := vdom.Props{}
		if isSelected {
			textProps["color"] = "blue"
		}

		children = append(children, ink.Box(props, ink.Text(textProps, label)))
	}

	return ink.Box(vdom.Props{"flexDirection": "column", "aria-role": "list"}, children...)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(SelectInput)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(SelectInput, ink.RenderOptions{})
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
