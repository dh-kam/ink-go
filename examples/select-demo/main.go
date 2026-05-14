package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
)

// Demo: a controlled Select rendered against a SelectState. The driver
// loop simulates arrow-key presses by calling state.MoveUp / MoveDown
// directly so the demo runs in any environment (no real TTY required).
func main() {
	items := []components.SelectItem{
		{Label: "Apple", Value: "apple"},
		{Label: "Banana", Value: "banana"},
		{Label: "Cherry", Value: "cherry"},
		{Label: "Date", Value: "date"},
		{Label: "Elderberry", Value: "elderberry"},
	}
	state := components.NewSelectState(items, 3)

	render := func(label string) {
		fmt.Printf("--- %s (selected=%d) ---\n", label, state.Selected)
		node := components.Select(components.SelectProps{
			Items:    state.Items,
			Selected: state.Selected,
			Limit:    state.Limit,
			Focused:  true,
		})
		fmt.Println(ink.RenderToString(node))
	}

	render("initial")
	state.MoveDown()
	state.MoveDown()
	render("after 2x down")
	state.MoveDown()
	state.MoveDown()
	state.MoveDown()
	render("after 3x more down (wrap test inside Limit window)")
	state.MoveUp()
	render("after up")

	if cur, ok := state.Confirm(); ok {
		fmt.Printf("\nConfirm: %s (value=%s)\n", cur.Label, cur.Value)
	}
}
