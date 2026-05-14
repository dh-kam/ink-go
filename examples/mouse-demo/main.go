package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/input"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// This demo parses a few SGR mouse sequences offline so it runs in any
// environment (the live runtime version would call ink.UseMouse from
// inside a Mount loop, which requires a real TTY).
func main() {
	samples := []string{
		"\x1b[<0;5;5M",   // left press at (5,5)
		"\x1b[<0;5;5m",   // left release at (5,5)
		"\x1b[<2;10;3M",  // right press at (10,3)
		"\x1b[<32;7;7M",  // left drag at (7,7)
		"\x1b[<35;8;8M",  // move (no button) at (8,8)
		"\x1b[<64;1;1M",  // wheel up at (1,1)
		"\x1b[<65;1;1M",  // wheel down at (1,1)
		"\x1b[<28;3;3M",  // shift+alt+ctrl + left press
	}

	for _, s := range samples {
		ev, err := input.ParseSGRMouse(s)
		if err != nil {
			fmt.Printf("parse error: %v\n", err)
			continue
		}
		fmt.Printf("%-25s -> %-7s %-9s @ (%d,%d) shift=%v alt=%v ctrl=%v\n",
			fmt.Sprintf("%q", s),
			ev.Button.String(),
			ev.Action.String(),
			ev.X, ev.Y,
			ev.Mods.Shift, ev.Mods.Alt, ev.Mods.Ctrl,
		)
	}

	// Confirm vdom imports compile in the demo binary.
	_ = components.Box(vdom.Props{}, components.Text("mouse-demo"))
}
