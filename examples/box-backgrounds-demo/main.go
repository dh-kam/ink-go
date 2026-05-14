package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/internal/renderer"
	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

const (
	synchronizedUpdateStart = "\x1b[?2026h"
	synchronizedUpdateEnd   = "\x1b[?2026l"
	clearTerminal           = "\x1b[2J\x1b[3J\x1b[H"
)

func main() {
	if terminal.StdoutIsTerminal() {
		fmt.Print(synchronizedUpdateStart + clearTerminal + renderBoxBackgroundsDemo() + synchronizedUpdateEnd)
		return
	}

	fmt.Print(renderBoxBackgroundsDemo())
}

func BoxBackgrounds() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"bold": true}, "Box Background Examples:"),

		components.Box(nil, components.Text("1. Standard red background (10x3):")),
		components.Box(vdom.Props{"backgroundColor": "red", "width": 10.0, "height": 3.0, "alignSelf": "flex-start"},
			components.Text("Hello"),
		),

		components.Box(nil, components.Text("2. Blue background with border (12x4):")),
		components.Box(vdom.Props{"backgroundColor": "blue", "borderStyle": "round", "width": 12.0, "height": 4.0, "alignSelf": "flex-start"},
			components.Text("Border"),
		),

		components.Box(nil, components.Text("3. Green background with padding (14x4):")),
		components.Box(vdom.Props{"backgroundColor": "green", "padding": 1.0, "width": 14.0, "height": 4.0, "alignSelf": "flex-start"},
			components.Text("Padding"),
		),

		components.Box(nil, components.Text("4. Yellow background with center alignment (16x3):")),
		components.Box(vdom.Props{"backgroundColor": "yellow", "width": 16.0, "height": 3.0, "justifyContent": "center", "alignSelf": "flex-start"},
			components.Text("Centered"),
		),

		components.Box(nil, components.Text("5. Magenta background, column layout (12x5):")),
		components.Box(vdom.Props{"backgroundColor": "magenta", "flexDirection": "column", "width": 12.0, "height": 5.0, "alignSelf": "flex-start"},
			components.Text("Line 1"),
			components.Text("Line 2"),
		),

		components.Box(nil, components.Text("6. Hex color background #FF8800 (10x3):")),
		components.Box(vdom.Props{"backgroundColor": "#FF8800", "width": 10.0, "height": 3.0, "alignSelf": "flex-start"},
			components.Text("Hex"),
		),

		components.Box(nil, components.Text("7. RGB background rgb(0,255,0) (10x3):")),
		components.Box(vdom.Props{"backgroundColor": "rgb(0,255,0)", "width": 10.0, "height": 3.0, "alignSelf": "flex-start"},
			components.Text("RGB"),
		),

		components.Box(nil, components.Text("8. Text inheritance test:")),
		components.Box(vdom.Props{"backgroundColor": "cyan", "alignSelf": "flex-start"},
			components.Text("Inherited "),
			components.Text(vdom.Props{"backgroundColor": "red"}, "Override "),
			components.Text("Back to inherited"),
		),

		components.Box(nil, components.Text("9. Nested background inheritance:")),
		components.Box(vdom.Props{"backgroundColor": "blue", "alignSelf": "flex-start"},
			components.Text("Outer: "),
			components.Box(vdom.Props{"backgroundColor": "yellow"},
				components.Text("Inner: "),
				components.Text(vdom.Props{"backgroundColor": "red"}, "Deep"),
			),
		),

		components.Box(vdom.Props{"marginTop": 1.0}, components.Text("Press Ctrl+C to exit")),
	)
}

func renderBoxBackgroundsDemo() string {
	return renderer.RenderWithLayoutANSI256(BoxBackgrounds(), 100, 80)
}
