package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type dimensionCase struct {
	name   string
	root   *vdom.Node
	width  int
	height int
}

func render(testCase dimensionCase) string {
	width := testCase.width
	if width == 0 {
		width = 100
	}

	height := testCase.height
	if height == 0 {
		height = 24
	}

	return renderer.RenderWithLayout(testCase.root, width, height)
}

func visible(output string) string {
	output = strings.ReplaceAll(output, " ", ".")
	output = strings.ReplaceAll(output, "\n", `\n`)
	return output
}

func printCase(testCase dimensionCase) {
	fmt.Printf("%s: %s\n", testCase.name, visible(render(testCase)))
}

func cases() []dimensionCase {
	setWidth := components.Box(nil,
		components.Box(vdom.Props{"width": 5.0}, components.Text("A")),
		components.Text("B"),
	)
	setHeight := components.Box(vdom.Props{"height": 4.0},
		components.Text("A"),
		components.Text("B"),
	)

	return []dimensionCase{
		{
			name: "set width",
			root: setWidth,
		},
		{
			name: "set width in percent",
			root: components.Box(vdom.Props{"width": 10.0},
				components.Box(vdom.Props{"width": "50%"}, components.Text("A")),
				components.Text("B"),
			),
		},
		{
			name: "set min width smaller",
			root: components.Box(nil,
				components.Box(vdom.Props{"minWidth": 5.0}, components.Text("A")),
				components.Text("B"),
			),
		},
		{
			name: "set min width larger",
			root: components.Box(nil,
				components.Box(vdom.Props{"minWidth": 2.0}, components.Text("AAAAA")),
				components.Text("B"),
			),
		},
		{
			name: "set min width in percent",
			root: components.Box(vdom.Props{"width": 10.0},
				components.Box(vdom.Props{"minWidth": "50%"}, components.Text("A")),
				components.Text("B"),
			),
		},
		{
			name: "set height",
			root: setHeight,
		},
		{
			name: "set height in percent",
			root: components.Box(vdom.Props{"height": 6.0, "flexDirection": "column"},
				components.Box(vdom.Props{"height": "50%"}, components.Text("A")),
				components.Text("B"),
			),
		},
		{
			name:  "cut text over the set height",
			width: 4,
			root: components.Box(vdom.Props{"height": 2.0},
				components.Text("AAAABBBBCCCC"),
			),
		},
		{
			name: "set min height smaller",
			root: components.Box(vdom.Props{"minHeight": 4.0},
				components.Text("A"),
			),
		},
		{
			name: "set min height larger",
			root: components.Box(vdom.Props{"minHeight": 2.0},
				components.Box(vdom.Props{"height": 4.0},
					components.Text("A"),
				),
			),
		},
		{
			name: "set width - concurrent",
			root: setWidth,
		},
		{
			name: "set height - concurrent",
			root: setHeight,
		},
	}
}

func main() {
	for _, testCase := range cases() {
		printCase(testCase)
	}
}
