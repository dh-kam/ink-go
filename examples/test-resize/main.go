package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func TestResize() *vdom.Node {
	stdout := ink.UseStdout()

	return ink.Box(vdom.Props{"flexDirection": "column", "borderStyle": "single", "padding": 1.0},
		ink.Text(vdom.Props{"color": "cyan"}, "--- Terminal Resize Test ---"),
		ink.Text(fmt.Sprintf("Current Size: %d x %d", stdout.Columns, stdout.Rows)),
		ink.Text("Resize your terminal window to see updates."),
	)
}

func main() {
	app := ink.NewApp(TestResize)
	fmt.Println(app.RenderOnce())
}
