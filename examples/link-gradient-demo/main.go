package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// App composes a single bordered column showing:
//   - a clickable hyperlink rendered with OSC 8
//   - the word "RAINBOW" smeared from red to violet via Gradient
//
// On supporting terminals (iTerm2 / WezTerm / Alacritty / kitty / VS Code)
// the first line will be a click target; everywhere else it gracefully
// falls back to plain text.
func App() *vdom.Node {
	link := components.Link(components.LinkProps{
		URL:  "https://github.com/dh-kam/ink-go",
		Text: "goink.go on GitHub",
	})

	rainbow := components.Gradient(components.GradientProps{
		Text: "RAINBOW",
		From: [3]uint8{255, 0, 0},   // red
		To:   [3]uint8{148, 0, 211}, // violet
	})

	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("Open in your terminal:"),
		components.Text(link),
		components.Text(""),
		components.Text(rainbow),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
