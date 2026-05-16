package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type paddingCase struct {
	name string
	root *vdom.Node
}

func visible(output string) string {
	output = strings.ReplaceAll(output, " ", ".")
	output = strings.ReplaceAll(output, "\n", `\n`)
	return output
}

func box(props vdom.Props, children ...*vdom.Node) *vdom.Node {
	return components.Box(props, children...)
}

func text(value string) *vdom.Node {
	return components.Text(value)
}

func printCase(testCase paddingCase) {
	output := renderer.RenderWithLayout(testCase.root, 100, 24)
	fmt.Printf("%s: %s\n", testCase.name, visible(output))
}

func main() {
	padding := box(vdom.Props{"padding": 2.0}, text("X"))
	nestedPadding := box(vdom.Props{"padding": 2.0},
		box(vdom.Props{"padding": 2.0}, text("X")),
	)

	cases := []paddingCase{
		{name: "padding", root: padding},
		{
			name: "padding X",
			root: box(nil,
				box(vdom.Props{"paddingX": 2.0}, text("X")),
				text("Y"),
			),
		},
		{name: "padding Y", root: box(vdom.Props{"paddingY": 2.0}, text("X"))},
		{name: "padding top", root: box(vdom.Props{"paddingTop": 2.0}, text("X"))},
		{name: "padding bottom", root: box(vdom.Props{"paddingBottom": 2.0}, text("X"))},
		{name: "padding left", root: box(vdom.Props{"paddingLeft": 2.0}, text("X"))},
		{
			name: "padding right",
			root: box(nil,
				box(vdom.Props{"paddingRight": 2.0}, text("X")),
				text("Y"),
			),
		},
		{name: "nested padding", root: nestedPadding},
		{name: "padding with multiline string", root: box(vdom.Props{"padding": 2.0}, text("A\nB"))},
		{name: "apply padding to text with newlines", root: box(vdom.Props{"padding": 1.0}, text("Hello\nWorld"))},
		{name: "apply padding to wrapped text", root: box(vdom.Props{"padding": 1.0, "width": 5.0}, text("Hello World"))},
		{name: "padding - concurrent", root: padding},
		{name: "nested padding - concurrent", root: nestedPadding},
	}

	for _, testCase := range cases {
		printCase(testCase)
	}
}
