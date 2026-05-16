package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type textCase struct {
	name string
	root *vdom.Node
}

var sgrPattern = regexp.MustCompile("\x1b\\[([0-9;]*)m")

func visible(output string) string {
	output = sgrPattern.ReplaceAllString(output, "<sgr:$1>")
	output = strings.ReplaceAll(output, "\n", `\n`)
	return output
}

func render(root *vdom.Node) string {
	if root == nil {
		return ""
	}

	return renderer.RenderWithLayoutANSI256(root, 100, 24)
}

func printCase(testCase textCase) {
	fmt.Printf("%s: %s\n", testCase.name, visible(render(testCase.root)))
}

func main() {
	syncCases := []textCase{
		{name: "undefined children", root: nil},
		{name: "null children", root: components.Text(nil)},
		{name: "standard color", root: components.Text(vdom.Props{"color": "green"}, "Test")},
		{name: "dim+bold", root: components.Text(vdom.Props{"dimColor": true, "bold": true}, "Test")},
		{name: "dimmed color", root: components.Text(vdom.Props{"dimColor": true, "color": "green"}, "Test")},
		{name: "hex color", root: components.Text(vdom.Props{"color": "#FF8800"}, "Test")},
		{name: "rgb color", root: components.Text(vdom.Props{"color": "rgb(255, 136, 0)"}, "Test")},
		{name: "ansi256 color", root: components.Text(vdom.Props{"color": "ansi256(194)"}, "Test")},
		{name: "standard background color", root: components.Text(vdom.Props{"backgroundColor": "green"}, "Test")},
		{name: "hex background color", root: components.Text(vdom.Props{"backgroundColor": "#FF8800"}, "Test")},
		{name: "rgb background color", root: components.Text(vdom.Props{"backgroundColor": "rgb(255, 136, 0)"}, "Test")},
		{name: "ansi256 background color", root: components.Text(vdom.Props{"backgroundColor": "ansi256(194)"}, "Test")},
		{name: "inversion", root: components.Text(vdom.Props{"inverse": true}, "Test")},
		{name: "constructor wraps correctly", root: components.Text("constructor")},
		{name: "remeasure text initial", root: components.Box(nil, components.Text("abc"))},
		{name: "remeasure text updated", root: components.Box(nil, components.Text("abcx"))},
		{name: "remeasure text nodes initial", root: components.Box(nil, components.Text("abc"))},
		{name: "remeasure text nodes updated", root: components.Box(nil, components.Text("abc", components.Text("x")))},
	}

	for _, testCase := range syncCases {
		printCase(testCase)
	}

	concurrentCases := []textCase{
		{name: "undefined children - concurrent", root: nil},
		{name: "null children - concurrent", root: components.Text(nil)},
		{name: "standard color - concurrent", root: components.Text(vdom.Props{"color": "green"}, "Test")},
		{name: "dim+bold - concurrent", root: components.Text(vdom.Props{"dimColor": true, "bold": true}, "Test")},
		{name: "hex color - concurrent", root: components.Text(vdom.Props{"color": "#FF8800"}, "Test")},
		{name: "inversion - concurrent", root: components.Text(vdom.Props{"inverse": true}, "Test")},
		{name: "remeasure text initial - concurrent", root: components.Box(nil, components.Text("abc"))},
		{name: "remeasure text updated - concurrent", root: components.Box(nil, components.Text("abcx"))},
		{name: "remeasure text nodes initial - concurrent", root: components.Box(nil, components.Text("abc"))},
		{name: "remeasure text nodes updated - concurrent", root: components.Box(nil, components.Text("abc", components.Text("x")))},
	}

	for _, testCase := range concurrentCases {
		printCase(testCase)
	}
}
