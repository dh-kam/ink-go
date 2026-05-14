package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
)

// Demo: a controlled Confirm rendered against a ConfirmState. We
// simulate a few key presses ('y', 'n', Enter) by feeding runes through
// state.HandleKey directly, so the demo runs in any environment without
// a real TTY.
func main() {
	render := func(label string, state *components.ConfirmState) {
		fmt.Printf("--- %s ---\n", label)
		node := components.Confirm(components.ConfirmProps{
			Question: state.Question,
			Default:  state.Default,
			Value:    state.Answer,
		})
		fmt.Println(ink.RenderToString(node))
	}

	// Scenario 1: explicit 'y'.
	s1 := components.NewConfirmState("Delete file?", false)
	render("initial (default=N)", s1)
	if resolved, v := s1.HandleKey('y'); resolved {
		fmt.Printf("key 'y' resolved → %v\n", v)
	}
	render("after 'y'", s1)

	// Scenario 2: explicit 'n'.
	s2 := components.NewConfirmState("Continue?", true)
	render("\ninitial (default=Y)", s2)
	if resolved, v := s2.HandleKey('n'); resolved {
		fmt.Printf("key 'n' resolved → %v\n", v)
	}
	render("after 'n'", s2)

	// Scenario 3: Enter takes the default.
	s3 := components.NewConfirmState("Save changes?", true)
	render("\ninitial (default=Y)", s3)
	if resolved, v := s3.HandleKey('\r'); resolved {
		fmt.Printf("Enter resolved → %v (default)\n", v)
	}
	render("after Enter", s3)

	// Reset demo.
	s3.Reset()
	render("\nafter Reset", s3)
}
