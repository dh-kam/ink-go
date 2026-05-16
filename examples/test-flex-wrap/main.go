package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type flexWrapCase struct {
	name string
	root *vdom.Node
}

func visible(output string) string {
	output = strings.ReplaceAll(output, " ", ".")
	output = strings.ReplaceAll(output, "\n", `\n`)
	return output
}

func printCase(testCase flexWrapCase) {
	output := renderer.RenderWithLayout(testCase.root, 100, 24)
	fmt.Printf("%s: %s\n", testCase.name, visible(output))
}

func main() {
	cases := []flexWrapCase{
		{
			name: "row - no wrap",
			root: components.Box(vdom.Props{"width": 2.0},
				components.Text("A"),
				components.Text("BC"),
			),
		},
		{
			name: "column - no wrap",
			root: components.Box(vdom.Props{"flexDirection": "column", "height": 2.0},
				components.Text("A"),
				components.Text("B"),
				components.Text("C"),
			),
		},
		{
			name: "row - wrap content",
			root: components.Box(vdom.Props{"width": 2.0, "flexWrap": "wrap"},
				components.Text("A"),
				components.Text("BC"),
			),
		},
		{
			name: "column - wrap content",
			root: components.Box(vdom.Props{"flexDirection": "column", "height": 2.0, "flexWrap": "wrap"},
				components.Text("A"),
				components.Text("B"),
				components.Text("C"),
			),
		},
		{
			name: "column - wrap content reverse",
			root: components.Box(vdom.Props{"flexDirection": "column", "height": 2.0, "width": 3.0, "flexWrap": "wrap-reverse"},
				components.Text("A"),
				components.Text("B"),
				components.Text("C"),
			),
		},
		{
			name: "row - wrap content reverse",
			root: components.Box(vdom.Props{"height": 3.0, "width": 2.0, "flexWrap": "wrap-reverse"},
				components.Text("A"),
				components.Text("B"),
				components.Text("C"),
			),
		},
	}

	for _, testCase := range cases {
		printCase(testCase)
	}
}
