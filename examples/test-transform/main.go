package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func bracketIndex(children string, index int) string {
	return fmt.Sprintf("[%d: %s]", index, children)
}

func braceIndex(children string, index int) string {
	return fmt.Sprintf("{%d: %s}", index, children)
}

func upper(children string, _ int) string {
	output := make([]rune, 0, len(children))
	for _, r := range children {
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		output = append(output, r)
	}
	return string(output)
}

func section(label string, child *vdom.Node) *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Text(vdom.Props{"color": "cyan"}, label),
		child,
	)
}

func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		components.Text(vdom.Props{"color": "yellow"}, "--- Transform Test ---"),
		section("nested transform",
			components.Transform(bracketIndex,
				components.Text(
					components.Transform(braceIndex,
						components.Text("test"),
					),
				),
			),
		),
		section("squash multiple text nodes",
			components.Transform(bracketIndex,
				components.Text(
					components.Transform(braceIndex,
						components.Text("hello", " ", "world"),
					),
				),
			),
		),
		section("multi-line transform",
			components.Transform(bracketIndex,
				components.Text("hello world\ngoodbye world"),
			),
		),
		section("transform inside text",
			components.Text("pre ",
				components.Transform(upper,
					components.Text("mid"),
				),
				" post",
			),
		),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
