package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/dh-kam/goink.go/internal/ttyinput"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func Robot() *vdom.Node {
	app := ink.UseApp()
	xValue, setX := ink.UseState(1)
	yValue, setY := ink.UseState(1)
	x := xValue.(int)
	y := yValue.(int)

	ink.UseInput(func(input string, key ink.InputKey) {
		if input == "q" {
			app.Exit()
			return
		}

		if key.LeftArrow {
			setX(max(1, x-1))
		}
		if key.RightArrow {
			setX(min(20, x+1))
		}
		if key.UpArrow {
			setY(max(1, y-1))
		}
		if key.DownArrow {
			setY(min(10, y+1))
		}
	})

	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Text("Use arrow keys to move the face. Press “q” to exit."),
		ink.Box(vdom.Props{
			"height":      12.0,
			"paddingLeft": float64(x),
			"paddingTop":  float64(y),
		}, ink.Text("^_^")),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(Robot)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(Robot, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := runInputLoop(instance); err != nil {
		fmt.Println(err)
	}
}

func runInputLoop(instance *ink.Instance) error {
	return ttyinput.Run(os.Stdin, instance.HandleInput, func(input string) bool {
		return strings.Contains(input, "q") || strings.Contains(input, "\x03")
	})
}

func min(a int, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func max(a int, b int) int {
	return int(math.Max(float64(a), float64(b)))
}
