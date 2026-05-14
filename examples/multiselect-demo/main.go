package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
)

// Demo: a controlled MultiSelect rendered against a MultiSelectState.
// We simulate cursor movement and toggles directly so the demo runs in
// any environment (no real TTY required).
func main() {
	items := []components.MultiSelectItem{
		{Label: "Apple", Value: "apple"},
		{Label: "Banana", Value: "banana"},
		{Label: "Cherry", Value: "cherry"},
		{Label: "Date", Value: "date", Disabled: true},
		{Label: "Elderberry", Value: "elderberry"},
	}
	state := components.NewMultiSelectState(items, 4)

	render := func(label string) {
		fmt.Printf("--- %s (cursor=%d, selected=[%s]) ---\n",
			label, state.Cursor, strings.Join(state.Values(), ","))
		node := components.MultiSelect(components.MultiSelectProps{
			Items:    state.Items,
			Selected: state.Values(),
			Cursor:   state.Cursor,
			Limit:    state.Limit,
			Focused:  true,
		})
		fmt.Println(ink.RenderToString(node))
	}

	render("initial")

	// Toggle the first item.
	state.Toggle()
	render("toggle Apple")

	// Move down twice and toggle Cherry.
	state.MoveDown()
	state.MoveDown()
	state.Toggle()
	render("toggle Cherry")

	// Try toggling the disabled Date — should be a no-op.
	state.MoveDown()
	state.Toggle()
	render("toggle Date (disabled, no-op)")

	// SelectAll then ClearAll.
	state.SelectAll()
	render("SelectAll")
	state.ClearAll()
	render("ClearAll")

	// Final values output.
	state.Toggle()
	state.MoveDown()
	state.Toggle()
	fmt.Printf("\nFinal Values(): %v\n", state.Values())
}
