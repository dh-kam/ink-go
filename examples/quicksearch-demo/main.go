package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
)

// Demo: QuickSearch driven by a QuickSearchState. The driver simulates
// query input and arrow-key presses by calling state methods directly so
// the demo runs in any environment (no real TTY required).
func main() {
	items := []components.SelectItem{
		{Label: "Apple", Value: "apple"},
		{Label: "Apricot", Value: "apricot"},
		{Label: "Banana", Value: "banana"},
		{Label: "Blackberry", Value: "blackberry"},
		{Label: "Blueberry", Value: "blueberry"},
		{Label: "Cherry", Value: "cherry"},
		{Label: "Date", Value: "date"},
		{Label: "Elderberry", Value: "elderberry"},
		{Label: "Fig", Value: "fig"},
		{Label: "Grape", Value: "grape"},
	}
	state := components.NewQuickSearchState(items, 5)

	render := func(label string) {
		filtered := state.Filtered()
		fmt.Printf("--- %s (query=%q, selected=%d, matches=%d) ---\n",
			label, state.Query, state.Selected, len(filtered))
		node := components.QuickSearch(components.QuickSearchProps{
			Items:    state.Items,
			Query:    state.Query,
			Selected: state.Selected,
			Limit:    state.Limit,
			Focused:  true,
		})
		fmt.Println(ink.RenderToString(node))
	}

	render("initial (empty query)")

	// Type "b" — filter to B-words.
	state.AppendQuery('b')
	render("typed 'b'")

	// Type "l" — filter to "Black..."/"Blue...".
	state.AppendQuery('l')
	render("typed 'l' -> 'bl'")

	// Move down within filtered list.
	state.MoveDown()
	render("MoveDown within filtered")

	// Backspace one char — back to all B-words.
	state.BackspaceQuery()
	render("backspace -> 'b'")

	// Reset with SetQuery to "ap" (case-insensitive, matches Apple/Apricot).
	state.SetQuery("AP")
	render("SetQuery 'AP' (case-insensitive)")

	// Wrap-around movement on filtered list.
	state.MoveUp()
	render("MoveUp wraps to last filtered")

	if cur, ok := state.Confirm(); ok {
		fmt.Printf("\nConfirm: %s (value=%s)\n", cur.Label, cur.Value)
	} else {
		fmt.Println("\nConfirm: no selection")
	}

	// Empty-filter confirm path.
	state.SetQuery("zzzz")
	render("query with no matches")
	if _, ok := state.Confirm(); !ok {
		fmt.Println("Confirm: (empty filtered list — no match)")
	}
}
