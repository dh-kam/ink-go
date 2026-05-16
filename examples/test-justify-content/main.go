package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type justifyCase struct {
	name string
	root *vdom.Node
}

func render(root *vdom.Node) string {
	return renderer.RenderWithLayoutANSI(root, 80, 24)
}

func visible(output string) string {
	replacer := strings.NewReplacer(
		"\x1b[32m", "<green>",
		"\x1b[39m", "<fgReset>",
	)

	output = replacer.Replace(output)
	output = strings.ReplaceAll(output, "\x1b", "<esc>")
	output = strings.ReplaceAll(output, " ", ".")
	output = strings.ReplaceAll(output, "\n", `\n`)
	return output
}

func printCase(testCase justifyCase) {
	fmt.Printf("%s: %s\n", testCase.name, visible(render(testCase.root)))
}

func cases() []justifyCase {
	return []justifyCase{
		{
			name: "row - align text to center",
			root: components.Box(vdom.Props{"justifyContent": "center", "width": 10.0},
				components.Text("Test"),
			),
		},
		{
			name: "row - align multiple text nodes to center",
			root: components.Box(vdom.Props{"justifyContent": "center", "width": 10.0},
				components.Text("A"),
				components.Text("B"),
			),
		},
		{
			name: "row - align text to right",
			root: components.Box(vdom.Props{"justifyContent": "flex-end", "width": 10.0},
				components.Text("Test"),
			),
		},
		{
			name: "row - align multiple text nodes to right",
			root: components.Box(vdom.Props{"justifyContent": "flex-end", "width": 10.0},
				components.Text("A"),
				components.Text("B"),
			),
		},
		{
			name: "row - align two text nodes on the edges",
			root: components.Box(vdom.Props{"justifyContent": "space-between", "width": 4.0},
				components.Text("A"),
				components.Text("B"),
			),
		},
		{
			name: "row - space evenly two text nodes",
			root: components.Box(vdom.Props{"justifyContent": "space-evenly", "width": 10.0},
				components.Text("A"),
				components.Text("B"),
			),
		},
		{
			name: "row - align two text nodes with equal space around them",
			root: components.Box(vdom.Props{"justifyContent": "space-around", "width": 5.0},
				components.Text("A"),
				components.Text("B"),
			),
		},
		{
			name: "row - align colored text node when text is squashed",
			root: components.Box(vdom.Props{"justifyContent": "flex-end", "width": 5.0},
				components.Text(vdom.Props{"color": "green"}, "X"),
			),
		},
		{
			name: "column - align text to center",
			root: components.Box(vdom.Props{"flexDirection": "column", "justifyContent": "center", "height": 3.0},
				components.Text("Test"),
			),
		},
		{
			name: "column - align text to bottom",
			root: components.Box(vdom.Props{"flexDirection": "column", "justifyContent": "flex-end", "height": 3.0},
				components.Text("Test"),
			),
		},
		{
			name: "column - align two text nodes on the edges",
			root: components.Box(vdom.Props{"flexDirection": "column", "justifyContent": "space-between", "height": 4.0},
				components.Text("A"),
				components.Text("B"),
			),
		},
		{
			name: "column - align two text nodes with equal space around them",
			root: components.Box(vdom.Props{"flexDirection": "column", "justifyContent": "space-around", "height": 5.0},
				components.Text("A"),
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
