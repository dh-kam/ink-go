package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type alignCase struct {
	name string
	root *vdom.Node
}

func render(root *vdom.Node) string {
	app := ink.NewAppWithOptions(func() *vdom.Node {
		return root
	}, ink.AppOptions{Width: 80, Height: 24})

	return app.RenderOnce()
}

func visible(output string) string {
	output = strings.ReplaceAll(output, " ", ".")
	output = strings.ReplaceAll(output, "\n", `\n`)
	return output
}

func printCase(testCase alignCase) {
	fmt.Printf("%s: %s\n", testCase.name, visible(render(testCase.root)))
}

func cases() []alignCase {
	return []alignCase{
		{
			name: "items row - align text to center",
			root: components.Box(vdom.Props{"alignItems": "center", "height": 3.0},
				components.Text("Test"),
			),
		},
		{
			name: "items row - align multiple text nodes to center",
			root: components.Box(vdom.Props{"alignItems": "center", "height": 3.0},
				components.Text("A"),
				components.Text("B"),
			),
		},
		{
			name: "items row - align text to bottom",
			root: components.Box(vdom.Props{"alignItems": "flex-end", "height": 3.0},
				components.Text("Test"),
			),
		},
		{
			name: "items row - align multiple text nodes to bottom",
			root: components.Box(vdom.Props{"alignItems": "flex-end", "height": 3.0},
				components.Text("A"),
				components.Text("B"),
			),
		},
		{
			name: "items column - align text to center",
			root: components.Box(vdom.Props{"flexDirection": "column", "alignItems": "center", "width": 10.0},
				components.Text("Test"),
			),
		},
		{
			name: "items column - align text to right",
			root: components.Box(vdom.Props{"flexDirection": "column", "alignItems": "flex-end", "width": 10.0},
				components.Text("Test"),
			),
		},
		{
			name: "self row - align text to center",
			root: components.Box(vdom.Props{"height": 3.0},
				components.Box(vdom.Props{"alignSelf": "center"},
					components.Text("Test"),
				),
			),
		},
		{
			name: "self row - align multiple text nodes to center",
			root: components.Box(vdom.Props{"height": 3.0},
				components.Box(vdom.Props{"alignSelf": "center"},
					components.Text("A"),
					components.Text("B"),
				),
			),
		},
		{
			name: "self row - align text to bottom",
			root: components.Box(vdom.Props{"height": 3.0},
				components.Box(vdom.Props{"alignSelf": "flex-end"},
					components.Text("Test"),
				),
			),
		},
		{
			name: "self row - align multiple text nodes to bottom",
			root: components.Box(vdom.Props{"height": 3.0},
				components.Box(vdom.Props{"alignSelf": "flex-end"},
					components.Text("A"),
					components.Text("B"),
				),
			),
		},
		{
			name: "self column - align text to center",
			root: components.Box(vdom.Props{"flexDirection": "column", "width": 10.0},
				components.Box(vdom.Props{"alignSelf": "center"},
					components.Text("Test"),
				),
			),
		},
		{
			name: "self column - align text to right",
			root: components.Box(vdom.Props{"flexDirection": "column", "width": 10.0},
				components.Box(vdom.Props{"alignSelf": "flex-end"},
					components.Text("Test"),
				),
			),
		},
	}
}

func main() {
	for _, testCase := range cases() {
		printCase(testCase)
	}
}
