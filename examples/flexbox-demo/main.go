package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/internal/renderer"
	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/layout"
	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func main() {
	// Create a flexbox layout with colored boxes
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection":  layout.FlexDirectionRow,
		"justifyContent": layout.JustifySpaceBetween,
		"width":          80.0,
		"height":         20.0,
		"padding":        2.0,
	})

	// Box 1 - Red
	box1 := vdom.CreateElement("box", vdom.Props{
		"width":  20.0,
		"height": 10.0,
	}, components.Text(styles.Colorize("Box 1", styles.Red, styles.Foreground)))

	// Box 2 - Green
	box2 := vdom.CreateElement("box", vdom.Props{
		"width":  20.0,
		"height": 10.0,
	}, components.Text(styles.Colorize("Box 2", styles.Green, styles.Foreground)))

	// Box 3 - Blue
	box3 := vdom.CreateElement("box", vdom.Props{
		"width":  20.0,
		"height": 10.0,
	}, components.Text(styles.Colorize("Box 3", styles.Blue, styles.Foreground)))

	container.Children = append(container.Children, box1, box2, box3)

	// Render with layout
	output := renderer.RenderWithLayout(container, 80, 25)

	fmt.Println("Flexbox Layout Demo - SpaceBetween")
	fmt.Println("===================================")
	fmt.Println(output)
	fmt.Println("\n✨ Flexbox + Colors working together!")
}
