package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type backgroundCase struct {
	name string
	root *vdom.Node
}

func render(root *vdom.Node) string {
	return renderer.RenderWithLayoutANSI(root, 80, 24)
}

func visible(output string) string {
	replacer := strings.NewReplacer(
		"\x1b[41m", "<bgRed>",
		"\x1b[42m", "<bgGreen>",
		"\x1b[43m", "<bgYellow>",
		"\x1b[44m", "<bgBlue>",
		"\x1b[45m", "<bgMagenta>",
		"\x1b[46m", "<bgCyan>",
		"\x1b[48;2;255;0;0m", "<bgRGB255000>",
		"\x1b[48;5;9m", "<bgAnsi256_9>",
		"\x1b[49m", "<bgReset>",
	)

	output = replacer.Replace(output)
	output = strings.ReplaceAll(output, "\x1b", "<esc>")
	output = strings.ReplaceAll(output, " ", ".")
	output = strings.ReplaceAll(output, "\n", `\n`)
	return output
}

func printCase(testCase backgroundCase) {
	fmt.Printf("%s: %s\n", testCase.name, visible(render(testCase.root)))
}

func cases() []backgroundCase {
	return []backgroundCase{
		{
			name: "text inherits parent box background",
			root: components.Box(vdom.Props{"backgroundColor": "green", "alignSelf": "flex-start"},
				components.Text("Hello World"),
			),
		},
		{
			name: "text explicit background overrides inherited",
			root: components.Box(vdom.Props{"backgroundColor": "red", "alignSelf": "flex-start"},
				components.Text(vdom.Props{"backgroundColor": "blue"}, "Hello World"),
			),
		},
		{
			name: "nested box background inheritance",
			root: components.Box(vdom.Props{"backgroundColor": "red", "alignSelf": "flex-start"},
				components.Box(vdom.Props{"backgroundColor": "blue"},
					components.Text("Hello World"),
				),
			),
		},
		{
			name: "text without parent box background",
			root: components.Box(vdom.Props{"alignSelf": "flex-start"},
				components.Text("Hello World"),
			),
		},
		{
			name: "multiple text elements inherit same background",
			root: components.Box(vdom.Props{"backgroundColor": "yellow", "alignSelf": "flex-start"},
				components.Text("Hello "),
				components.Text("World"),
			),
		},
		{
			name: "mixed inheritance clear and explicit",
			root: components.Box(vdom.Props{"backgroundColor": "green", "alignSelf": "flex-start"},
				components.Text("Inherited "),
				components.Text(vdom.Props{"backgroundColor": ""}, "No BG "),
				components.Text(vdom.Props{"backgroundColor": "red"}, "Red BG"),
			),
		},
		{
			name: "complex nested inheritance",
			root: components.Box(vdom.Props{"backgroundColor": "yellow", "alignSelf": "flex-start"},
				components.Box(nil,
					components.Text("Outer: "),
					components.Box(vdom.Props{"backgroundColor": "blue"},
						components.Text("Inner: "),
						components.Text(vdom.Props{"backgroundColor": "red"}, "Explicit"),
					),
				),
			),
		},
		{
			name: "box background standard color",
			root: components.Box(vdom.Props{"backgroundColor": "red", "alignSelf": "flex-start"},
				components.Text("Hello"),
			),
		},
		{
			name: "box background hex color",
			root: components.Box(vdom.Props{"backgroundColor": "#FF0000", "alignSelf": "flex-start"},
				components.Text("Hello"),
			),
		},
		{
			name: "box background rgb color",
			root: components.Box(vdom.Props{"backgroundColor": "rgb(255, 0, 0)", "alignSelf": "flex-start"},
				components.Text("Hello"),
			),
		},
		{
			name: "box background ansi256 color",
			root: components.Box(vdom.Props{"backgroundColor": "ansi256(9)", "alignSelf": "flex-start"},
				components.Text("Hello"),
			),
		},
		{
			name: "box background wide characters",
			root: components.Box(vdom.Props{"backgroundColor": "yellow", "alignSelf": "flex-start"},
				components.Text("こんにちは"),
			),
		},
		{
			name: "box background emojis",
			root: components.Box(vdom.Props{"backgroundColor": "red", "alignSelf": "flex-start"},
				components.Text("🎉🎊"),
			),
		},
		{
			name: "box background fills standard color",
			root: components.Box(vdom.Props{"backgroundColor": "red", "width": 10.0, "height": 3.0, "alignSelf": "flex-start"},
				components.Text("Hello"),
			),
		},
		{
			name: "box background with border",
			root: components.Box(vdom.Props{"backgroundColor": "cyan", "borderStyle": "round", "width": 10.0, "height": 5.0, "alignSelf": "flex-start"},
				components.Text("Hi"),
			),
		},
		{
			name: "box background with padding",
			root: components.Box(vdom.Props{"backgroundColor": "magenta", "padding": 1.0, "width": 10.0, "height": 5.0, "alignSelf": "flex-start"},
				components.Text("Hi"),
			),
		},
		{
			name: "box background center alignment",
			root: components.Box(vdom.Props{"backgroundColor": "blue", "width": 10.0, "height": 3.0, "justifyContent": "center", "alignSelf": "flex-start"},
				components.Text("Hi"),
			),
		},
		{
			name: "box background column layout",
			root: components.Box(vdom.Props{"backgroundColor": "green", "flexDirection": "column", "width": 10.0, "height": 5.0, "alignSelf": "flex-start"},
				components.Text("Line 1"),
				components.Text("Line 2"),
			),
		},
	}
}

func main() {
	for _, testCase := range cases() {
		printCase(testCase)
	}
}
