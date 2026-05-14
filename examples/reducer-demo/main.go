package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type action int

const (
	increment action = iota
	decrement
	reset
)

func reducer(state int, a action) int {
	switch a {
	case increment:
		return state + 1
	case decrement:
		return state - 1
	case reset:
		return 0
	}
	return state
}

// Capture dispatch on first render so the driver loop can fire actions.
var capturedDispatch func(action)

func Counter() *vdom.Node {
	count, dispatch := ink.UseReducer(reducer, 0)
	capturedDispatch = dispatch
	return components.Box(vdom.Props{},
		components.Text(fmt.Sprintf("Count: %d", count)),
	)
}

func main() {
	app := ink.NewApp(Counter)
	fmt.Println("--- initial ---")
	fmt.Println(app.RenderOnce())

	capturedDispatch(increment)
	capturedDispatch(increment)
	capturedDispatch(increment)
	fmt.Println("--- after +++ ---")
	fmt.Println(app.RenderOnce())

	capturedDispatch(decrement)
	fmt.Println("--- after - ---")
	fmt.Println(app.RenderOnce())

	capturedDispatch(reset)
	fmt.Println("--- after reset ---")
	fmt.Println(app.RenderOnce())
}
