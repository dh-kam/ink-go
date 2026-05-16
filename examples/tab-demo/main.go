package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// Demo: a controlled Tabs rendered against a TabsState. The driver loop
// simulates Tab / Shift-Tab key presses by calling state.Next / state.Prev
// directly so the demo runs without a real TTY. Each step prints the active
// tab's panel to show that switching tabs swaps the rendered content.
func main() {
	items := []components.TabItem{
		{
			Label:   "Logs",
			Content: logsPanel(),
		},
		{
			Label:   "Settings",
			Content: settingsPanel(),
		},
		{
			Label:   "About",
			Content: aboutPanel(),
		},
	}

	state := components.NewTabsState(items)

	render := func(label string) {
		cur, _ := state.Current()
		fmt.Printf("--- %s (active=%d %q) ---\n", label, state.Active, cur.Label)
		node := components.Tabs(components.TabsProps{
			Items:   state.Items,
			Active:  state.Active,
			Focused: true,
		})
		fmt.Println(ink.RenderToString(node))
		fmt.Println()
	}

	render("initial")

	// Simulate Tab -> Settings
	state.Next()
	render("after Tab")

	// Simulate Tab -> About
	state.Next()
	render("after Tab")

	// Simulate Tab -> wraps to Logs
	state.Next()
	render("after Tab (wrap)")

	// Simulate Shift-Tab -> wraps back to About
	state.Prev()
	render("after Shift-Tab (wrap)")

	// Jump directly via SetActive
	state.SetActive(1)
	render("after SetActive(1)")
}

func logsPanel() *vdom.Node {
	return components.Box(
		vdom.Props{"flexDirection": "column", "paddingTop": 1},
		components.Text("[12:00:01] server started"),
		components.Text("[12:00:02] listening on :8080"),
		components.Text("[12:00:05] handled GET / in 3ms"),
	)
}

func settingsPanel() *vdom.Node {
	return components.Box(
		vdom.Props{"flexDirection": "column", "paddingTop": 1},
		components.Text("theme:    dark"),
		components.Text("editor:   vim"),
		components.Text("autosave: on"),
	)
}

func aboutPanel() *vdom.Node {
	return components.Box(
		vdom.Props{"flexDirection": "column", "paddingTop": 1},
		components.Text("ink-go tab demo"),
		components.Text("press Tab / Shift-Tab to cycle (simulated)"),
	)
}
