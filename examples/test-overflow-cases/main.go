package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type overflowCase struct {
	name   string
	root   *vdom.Node
	width  int
	height int
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

func rounded(props vdom.Props, children ...*vdom.Node) *vdom.Node {
	if props == nil {
		props = vdom.Props{}
	}
	props["borderStyle"] = "round"
	return box(props, children...)
}

func render(testCase overflowCase) string {
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

func printCase(testCase overflowCase) {
	fmt.Printf("%s: %s\n", testCase.name, visible(render(testCase)))
}

func multilineABCD() *vdom.Node {
	return text("AAAA\nBBBB\nCCCC\nDDDD")
}

func nestedOverflowTree() *vdom.Node {
	return box(vdom.Props{"paddingBottom": 1.0},
		box(vdom.Props{"width": 4.0, "height": 4.0, "overflow": "hidden", "flexDirection": "column"},
			box(vdom.Props{"width": 2.0, "height": 2.0, "overflow": "hidden"},
				box(vdom.Props{"width": 4.0, "height": 4.0, "flexShrink": 0.0}, multilineABCD()),
			),
			box(vdom.Props{"width": 4.0, "height": 3.0},
				text("XXXX\nYYYY\nZZZZ"),
			),
		),
	)
}

func main() {
	lineBoxes := func() []*vdom.Node {
		return []*vdom.Node{
			box(vdom.Props{"flexShrink": 0.0}, text("Line #1")),
			box(vdom.Props{"flexShrink": 0.0}, text("Line #2")),
			box(vdom.Props{"flexShrink": 0.0}, text("Line #3")),
			box(vdom.Props{"flexShrink": 0.0}, text("Line #4")),
		}
	}

	cases := []overflowCase{
		{
			name: "overflowX - single text node in a box inside overflow container",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"width": 16.0, "flexShrink": 0.0}, text("Hello World")),
			),
		},
		{
			name: "overflowX - single text node inside overflow container with border",
			root: rounded(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"width": 16.0, "flexShrink": 0.0}, text("Hello World")),
			),
		},
		{
			name: "overflowX - single text node in a box with border inside overflow container",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				rounded(vdom.Props{"width": 16.0, "flexShrink": 0.0}, text("Hello World")),
			),
		},
		{
			name: "overflowX - multiple text nodes in a box inside overflow container",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"width": 12.0, "flexShrink": 0.0}, text("Hello "), text("World")),
			),
		},
		{
			name: "overflowX - multiple text nodes in a box inside overflow container with border",
			root: rounded(vdom.Props{"width": 8.0, "overflowX": "hidden"},
				box(vdom.Props{"width": 12.0, "flexShrink": 0.0}, text("Hello "), text("World")),
			),
		},
		{
			name: "overflowX - multiple text nodes in a box with border inside overflow container",
			root: box(vdom.Props{"width": 8.0, "overflowX": "hidden"},
				rounded(vdom.Props{"width": 12.0, "flexShrink": 0.0}, text("Hello "), text("World")),
			),
		},
		{
			name: "overflowX - multiple boxes inside overflow container",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"width": 6.0, "flexShrink": 0.0}, text("Hello ")),
				box(vdom.Props{"width": 6.0, "flexShrink": 0.0}, text("World")),
			),
		},
		{
			name: "overflowX - multiple boxes inside overflow container with border",
			root: rounded(vdom.Props{"width": 8.0, "overflowX": "hidden"},
				box(vdom.Props{"width": 6.0, "flexShrink": 0.0}, text("Hello ")),
				box(vdom.Props{"width": 6.0, "flexShrink": 0.0}, text("World")),
			),
		},
		{
			name: "overflowX - box before left edge of overflow container",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"marginLeft": -12.0, "width": 6.0, "flexShrink": 0.0}, text("Hello")),
			),
		},
		{
			name: "overflowX - box before left edge of overflow container with border",
			root: rounded(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"marginLeft": -12.0, "width": 6.0, "flexShrink": 0.0}, text("Hello")),
			),
		},
		{
			name: "overflowX - box intersecting with left edge of overflow container",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"marginLeft": -3.0, "width": 12.0, "flexShrink": 0.0}, text("Hello World")),
			),
		},
		{
			name: "overflowX - box intersecting with left edge of overflow container with border",
			root: rounded(vdom.Props{"width": 8.0, "overflowX": "hidden"},
				box(vdom.Props{"marginLeft": -3.0, "width": 12.0, "flexShrink": 0.0}, text("Hello World")),
			),
		},
		{
			name: "overflowX - box after right edge of overflow container",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"marginLeft": 6.0, "width": 6.0, "flexShrink": 0.0}, text("Hello")),
			),
		},
		{
			name: "overflowX - box intersecting with right edge of overflow container",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"marginLeft": 3.0, "width": 6.0, "flexShrink": 0.0}, text("Hello")),
			),
		},
		{
			name: "overflowY - single text node inside overflow container",
			root: box(vdom.Props{"height": 1.0, "overflowY": "hidden"}, text("Hello\nWorld")),
		},
		{
			name: "overflowY - single text node inside overflow container with border",
			root: rounded(vdom.Props{"width": 20.0, "height": 3.0, "overflowY": "hidden"}, text("Hello\nWorld")),
		},
		{
			name: "overflowY - multiple boxes inside overflow container",
			root: box(vdom.Props{"height": 2.0, "overflowY": "hidden", "flexDirection": "column"}, lineBoxes()...),
		},
		{
			name: "overflowY - multiple boxes inside overflow container with border",
			root: rounded(vdom.Props{"width": 9.0, "height": 4.0, "overflowY": "hidden", "flexDirection": "column"}, lineBoxes()...),
		},
		{
			name: "overflowY - box above top edge of overflow container",
			root: box(vdom.Props{"height": 1.0, "overflowY": "hidden"},
				box(vdom.Props{"marginTop": -2.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
			),
		},
		{
			name: "overflowY - box above top edge of overflow container with border",
			root: rounded(vdom.Props{"width": 7.0, "height": 3.0, "overflowY": "hidden"},
				box(vdom.Props{"marginTop": -3.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
			),
		},
		{
			name: "overflowY - box intersecting with top edge of overflow container",
			root: box(vdom.Props{"height": 1.0, "overflowY": "hidden"},
				box(vdom.Props{"marginTop": -1.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
			),
		},
		{
			name: "overflowY - box intersecting with top edge of overflow container with border",
			root: rounded(vdom.Props{"width": 7.0, "height": 3.0, "overflowY": "hidden"},
				box(vdom.Props{"marginTop": -1.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
			),
		},
		{
			name: "overflowY - box below bottom edge of overflow container",
			root: box(vdom.Props{"height": 1.0, "overflowY": "hidden"},
				box(vdom.Props{"marginTop": 1.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
			),
		},
		{
			name: "overflowY - box below bottom edge of overflow container with border",
			root: rounded(vdom.Props{"width": 7.0, "height": 3.0, "overflowY": "hidden"},
				box(vdom.Props{"marginTop": 2.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
			),
		},
		{
			name: "overflowY - box intersecting with bottom edge of overflow container",
			root: box(vdom.Props{"height": 1.0, "overflowY": "hidden"},
				box(vdom.Props{"height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
			),
		},
		{
			name: "overflowY - box intersecting with bottom edge of overflow container with border",
			root: rounded(vdom.Props{"width": 7.0, "height": 3.0, "overflowY": "hidden"},
				box(vdom.Props{"height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
			),
		},
		{
			name: "overflow - single text node inside overflow container",
			root: box(vdom.Props{"paddingBottom": 1.0},
				box(vdom.Props{"width": 6.0, "height": 1.0, "overflow": "hidden"},
					box(vdom.Props{"width": 12.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
				),
			),
		},
		{
			name: "overflow - single text node inside overflow container with border",
			root: box(vdom.Props{"paddingBottom": 1.0},
				rounded(vdom.Props{"width": 8.0, "height": 3.0, "overflow": "hidden"},
					box(vdom.Props{"width": 12.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
				),
			),
		},
		{
			name: "overflow - multiple boxes inside overflow container",
			root: box(vdom.Props{"paddingBottom": 1.0},
				box(vdom.Props{"width": 4.0, "height": 1.0, "overflow": "hidden"},
					box(vdom.Props{"width": 2.0, "height": 2.0, "flexShrink": 0.0}, text("TL\nBL")),
					box(vdom.Props{"width": 2.0, "height": 2.0, "flexShrink": 0.0}, text("TR\nBR")),
				),
			),
		},
		{
			name: "overflow - multiple boxes inside overflow container with border",
			root: box(vdom.Props{"paddingBottom": 1.0},
				rounded(vdom.Props{"width": 6.0, "height": 3.0, "overflow": "hidden"},
					box(vdom.Props{"width": 2.0, "height": 2.0, "flexShrink": 0.0}, text("TL\nBL")),
					box(vdom.Props{"width": 2.0, "height": 2.0, "flexShrink": 0.0}, text("TR\nBR")),
				),
			),
		},
		{
			name: "overflow - box intersecting with top left edge of overflow container",
			root: box(vdom.Props{"width": 4.0, "height": 4.0, "overflow": "hidden"},
				box(vdom.Props{"marginTop": -2.0, "marginLeft": -2.0, "width": 4.0, "height": 4.0, "flexShrink": 0.0}, multilineABCD()),
			),
		},
		{
			name: "overflow - box intersecting with top right edge of overflow container",
			root: box(vdom.Props{"width": 4.0, "height": 4.0, "overflow": "hidden"},
				box(vdom.Props{"marginTop": -2.0, "marginLeft": 2.0, "width": 4.0, "height": 4.0, "flexShrink": 0.0}, multilineABCD()),
			),
		},
		{
			name: "overflow - box intersecting with bottom left edge of overflow container",
			root: box(vdom.Props{"width": 4.0, "height": 4.0, "overflow": "hidden"},
				box(vdom.Props{"marginTop": 2.0, "marginLeft": -2.0, "width": 4.0, "height": 4.0, "flexShrink": 0.0}, multilineABCD()),
			),
		},
		{
			name: "overflow - box intersecting with bottom right edge of overflow container",
			root: box(vdom.Props{"width": 4.0, "height": 4.0, "overflow": "hidden"},
				box(vdom.Props{"marginTop": 2.0, "marginLeft": 2.0, "width": 4.0, "height": 4.0, "flexShrink": 0.0}, multilineABCD()),
			),
		},
		{name: "nested overflow", root: nestedOverflowTree()},
		{
			name:  "out of bounds writes do not crash",
			root:  rounded(vdom.Props{"width": 12.0, "height": 10.0}),
			width: 10,
		},
		{
			name: "overflowX - single text node in a box inside overflow container - concurrent",
			root: box(vdom.Props{"width": 6.0, "overflowX": "hidden"},
				box(vdom.Props{"width": 16.0, "flexShrink": 0.0}, text("Hello World")),
			),
		},
		{
			name: "overflowY - single text node inside overflow container - concurrent",
			root: box(vdom.Props{"height": 1.0, "overflowY": "hidden"}, text("Hello\nWorld")),
		},
		{
			name: "overflow - single text node inside overflow container - concurrent",
			root: box(vdom.Props{"paddingBottom": 1.0},
				box(vdom.Props{"width": 6.0, "height": 1.0, "overflow": "hidden"},
					box(vdom.Props{"width": 12.0, "height": 2.0, "flexShrink": 0.0}, text("Hello\nWorld")),
				),
			),
		},
		{name: "nested overflow - concurrent", root: nestedOverflowTree()},
	}

	for _, testCase := range cases {
		printCase(testCase)
	}
}
