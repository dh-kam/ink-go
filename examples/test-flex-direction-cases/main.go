package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type flexDirectionCase struct {
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

func printCase(testCase flexDirectionCase) {
	output := renderer.RenderWithLayout(testCase.root, 100, 24)
	fmt.Printf("%s: %s\n", testCase.name, visible(output))
}

func main() {
	row := box(vdom.Props{"flexDirection": "row"}, text("A"), text("B"))
	column := box(vdom.Props{"flexDirection": "column"}, text("A"), text("B"))

	cases := []flexDirectionCase{
		{name: "direction row", root: row},
		{name: "direction row reverse", root: box(vdom.Props{"flexDirection": "row-reverse", "width": 4.0}, text("A"), text("B"))},
		{name: "direction column", root: column},
		{name: "direction column reverse", root: box(vdom.Props{"flexDirection": "column-reverse", "height": 4.0}, text("A"), text("B"))},
		{name: "don't squash text nodes when column direction is applied", root: column},
		{name: "direction row - concurrent", root: row},
		{name: "direction column - concurrent", root: column},
	}

	for _, testCase := range cases {
		printCase(testCase)
	}
}
