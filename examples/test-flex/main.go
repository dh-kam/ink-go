package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type flexCase struct {
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

func printCase(testCase flexCase) {
	fmt.Printf("%s: %s\n", testCase.name, visible(render(testCase.root)))
}

func cases() []flexCase {
	return []flexCase{
		{
			name: "grow equally",
			root: components.Box(vdom.Props{"width": 6.0},
				components.Box(vdom.Props{"flexGrow": 1.0}, components.Text("A")),
				components.Box(vdom.Props{"flexGrow": 1.0}, components.Text("B")),
			),
		},
		{
			name: "grow one element",
			root: components.Box(vdom.Props{"width": 6.0},
				components.Box(vdom.Props{"flexGrow": 1.0}, components.Text("A")),
				components.Text("B"),
			),
		},
		{
			name: "dont shrink",
			root: components.Box(vdom.Props{"width": 16.0},
				components.Box(vdom.Props{"flexShrink": 0.0, "width": 6.0}, components.Text("A")),
				components.Box(vdom.Props{"flexShrink": 0.0, "width": 6.0}, components.Text("B")),
				components.Box(vdom.Props{"width": 6.0}, components.Text("C")),
			),
		},
		{
			name: "shrink equally",
			root: components.Box(vdom.Props{"width": 10.0},
				components.Box(vdom.Props{"flexShrink": 1.0, "width": 6.0}, components.Text("A")),
				components.Box(vdom.Props{"flexShrink": 1.0, "width": 6.0}, components.Text("B")),
				components.Text("C"),
			),
		},
		{
			name: "row flex basis",
			root: components.Box(vdom.Props{"width": 6.0},
				components.Box(vdom.Props{"flexBasis": 3.0}, components.Text("A")),
				components.Text("B"),
			),
		},
		{
			name: "row percent flex basis",
			root: components.Box(vdom.Props{"width": 6.0},
				components.Box(vdom.Props{"flexBasis": "50%"}, components.Text("A")),
				components.Text("B"),
			),
		},
		{
			name: "column flex basis",
			root: components.Box(vdom.Props{"height": 6.0, "flexDirection": "column"},
				components.Box(vdom.Props{"flexBasis": 3.0}, components.Text("A")),
				components.Text("B"),
			),
		},
		{
			name: "column percent flex basis",
			root: components.Box(vdom.Props{"height": 6.0, "flexDirection": "column"},
				components.Box(vdom.Props{"flexBasis": "50%"}, components.Text("A")),
				components.Text("B"),
			),
		},
	}
}

func main() {
	for _, testCase := range cases() {
		printCase(testCase)
	}
}
