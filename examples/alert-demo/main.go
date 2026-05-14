package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func main() {
	variants := []components.AlertProps{
		{Variant: components.AlertInfo, Title: "Heads up", Message: "Background sync started."},
		{Variant: components.AlertSuccess, Title: "Done", Message: "All 12 files written."},
		{Variant: components.AlertWarning, Title: "Slow", Message: "Disk I/O above threshold."},
		{Variant: components.AlertError, Title: "Failed", Message: "Cannot reach upstream registry."},
	}

	stack := make([]*vdom.Node, 0, len(variants))
	for _, p := range variants {
		stack = append(stack, components.Alert(p))
	}

	root := components.Box(vdom.Props{"flexDirection": "column"}, stack...)
	fmt.Println(ink.RenderToString(root))

	// Inline form for quick log lines
	fmt.Println()
	for _, p := range variants {
		fmt.Println(components.AlertText(p))
	}
}
