package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func FocusableBox(label string, id string) *vdom.Node {
	focus := ink.UseFocusOpts(ink.FocusOptions{ID: id})

	color := "white"
	suffix := ""
	if focus.IsFocused {
		color = "green"
		suffix = "[FOCUSED]"
	}
	text := label + " " + suffix

	return ink.Box(vdom.Props{
		"borderStyle": "single",
		"borderColor": color,
		"paddingX":    1.0,
	},
		ink.Text(vdom.Props{"color": color}, text),
	)
}

func TestFocus() *vdom.Node {
	return ink.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		ink.Text(vdom.Props{"color": "yellow"}, "--- Focus Management Test ---"),
		ink.Text("Press Tab to cycle focus, Shift+Tab to reverse."),
		ink.Box(vdom.Props{"gap": 2.0},
			FocusableBox("Box 1", "box1"),
			FocusableBox("Box 2", "box2"),
			FocusableBox("Box 3", "box3"),
		),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(TestFocus)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(TestFocus, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		fmt.Println(err)
	}
}
