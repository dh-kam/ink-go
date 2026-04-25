package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func App() *vdom.Node {
	// Create styled text nodes
	redText := components.Text(styles.Colorize("Red Text", styles.Red, styles.Foreground))
	greenText := components.Text(styles.Colorize("Green Text", styles.Green, styles.Foreground))
	blueText := components.Text(styles.Colorize("Blue Text", styles.Blue, styles.Foreground))

	// Bold text
	boldText := components.Text(styles.Bold("Bold Text"))

	// Combined styles
	fancyText := components.Text(
		styles.Bold(
			styles.Colorize("Bold Red Text", styles.Red, styles.Foreground),
		),
	)

	return components.Box(nil,
		components.Text("Welcome to Goink with Colors!"),
		components.Newline(),
		components.Newline(),
		redText,
		components.Newline(),
		greenText,
		components.Newline(),
		blueText,
		components.Newline(),
		components.Newline(),
		boldText,
		components.Newline(),
		fancyText,
	)
}

func main() {
	app := ink.NewApp(App)
	output := app.RenderOnce()

	fmt.Println(output)
	fmt.Println("\n✨ Colors and styles working!")
}
