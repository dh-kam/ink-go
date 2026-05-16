package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type marginCase struct {
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

func printCase(testCase marginCase) {
	fmt.Printf("%s: %s\n", testCase.name, visible(render(testCase.root)))
}

func cases() []marginCase {
	margin := components.Box(vdom.Props{"margin": 2.0}, components.Text("X"))
	nestedMargin := components.Box(vdom.Props{"margin": 2.0},
		components.Box(vdom.Props{"margin": 2.0}, components.Text("X")),
	)

	return []marginCase{
		{
			name: "margin",
			root: margin,
		},
		{
			name: "margin X",
			root: components.Box(nil,
				components.Box(vdom.Props{"marginX": 2.0}, components.Text("X")),
				components.Text("Y"),
			),
		},
		{
			name: "margin Y",
			root: components.Box(vdom.Props{"marginY": 2.0}, components.Text("X")),
		},
		{
			name: "margin top",
			root: components.Box(vdom.Props{"marginTop": 2.0}, components.Text("X")),
		},
		{
			name: "margin bottom",
			root: components.Box(vdom.Props{"marginBottom": 2.0}, components.Text("X")),
		},
		{
			name: "margin left",
			root: components.Box(vdom.Props{"marginLeft": 2.0}, components.Text("X")),
		},
		{
			name: "margin right",
			root: components.Box(nil,
				components.Box(vdom.Props{"marginRight": 2.0}, components.Text("X")),
				components.Text("Y"),
			),
		},
		{
			name: "nested margin",
			root: nestedMargin,
		},
		{
			name: "margin with multiline string",
			root: components.Box(vdom.Props{"margin": 2.0},
				components.Text("A\nB"),
			),
		},
		{
			name: "apply margin to text with newlines",
			root: components.Box(vdom.Props{"margin": 1.0},
				components.Text("Hello\nWorld"),
			),
		},
		{
			name: "apply margin to wrapped text",
			root: components.Box(vdom.Props{"margin": 1.0, "width": 6.0},
				components.Text("Hello World"),
			),
		},
		{
			name: "margin - concurrent",
			root: margin,
		},
		{
			name: "nested margin - concurrent",
			root: nestedMargin,
		},
	}
}

func main() {
	for _, testCase := range cases() {
		printCase(testCase)
	}
}
