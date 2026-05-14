package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/layout"
	"github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func Dashboard() *vdom.Node {
	// Get state
	visitors, _ := ink.UseState(1337)
	sales, _ := ink.UseState(42)

	// Header
	title := components.Text(
		styles.Bold(
			styles.Colorize("Dashboard", styles.Cyan, styles.Foreground),
		),
	)

	subtitle := components.Text(
		styles.Dim("Real-time statistics"),
	)

	header := vdom.CreateElement("box", vdom.Props{
		"width":         80.0,
		"height":        5.0,
		"flexDirection": layout.FlexDirectionColumn,
		"padding":       1.0,
	})
	header.Children = append(header.Children, title, subtitle)

	// Stats boxes
	statsContainer := vdom.CreateElement("box", vdom.Props{
		"width":          80.0,
		"height":         10.0,
		"flexDirection":  layout.FlexDirectionRow,
		"justifyContent": layout.JustifySpaceAround,
		"padding":        2.0,
	})

	// Visitors box
	visitorsBox := vdom.CreateElement("box", vdom.Props{
		"width":  25.0,
		"height": 8.0,
	})
	visitorsLabel := components.Text(
		styles.Colorize("VISITORS", styles.Green, styles.Foreground),
	)
	visitorsValue := components.Text(
		styles.Bold(fmt.Sprintf("%d", visitors)),
	)
	visitorsBox.Children = append(visitorsBox.Children, visitorsLabel, visitorsValue)

	// Sales box
	salesBox := vdom.CreateElement("box", vdom.Props{
		"width":  25.0,
		"height": 8.0,
	})
	salesLabel := components.Text(
		styles.Colorize("SALES", styles.Yellow, styles.Foreground),
	)
	salesValue := components.Text(
		styles.Bold(fmt.Sprintf("$%d", sales)),
	)
	salesBox.Children = append(salesBox.Children, salesLabel, salesValue)

	// Status box
	statusBox := vdom.CreateElement("box", vdom.Props{
		"width":  25.0,
		"height": 8.0,
	})
	statusLabel := components.Text(
		styles.Colorize("STATUS", styles.Blue, styles.Foreground),
	)
	statusValue := components.Text(
		styles.Bold(styles.Colorize("ONLINE", styles.Green, styles.Foreground)),
	)
	statusBox.Children = append(statusBox.Children, statusLabel, statusValue)

	statsContainer.Children = append(statsContainer.Children, visitorsBox, salesBox, statusBox)

	// Main container
	root := vdom.CreateElement("box", vdom.Props{
		"width":         80.0,
		"height":        20.0,
		"flexDirection": layout.FlexDirectionColumn,
	})
	root.Children = append(root.Children, header, statsContainer)

	return root
}

func main() {
	app := ink.NewApp(Dashboard)
	node := app.GetVNode()
	output := renderer.RenderWithLayout(node, 80, 25)

	fmt.Println(output)
	fmt.Println("\n🎉 Complete dashboard with layout, colors, and state!")
}
