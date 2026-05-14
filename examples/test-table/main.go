package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type tableRow struct {
	ID    int
	Name  string
	Email string
}

var data = []tableRow{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
	{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
}

func TestTable() *vdom.Node {
	rows := make([]*vdom.Node, 0, len(data)+1)
	rows = append(rows, ink.Box(vdom.Props{
		"borderStyle": "single",
		"borderTop":   false,
		"borderLeft":  false,
		"borderRight": false,
	},
		ink.Box(vdom.Props{"width": 10.0}, ink.Text(vdom.Props{"bold": true}, "ID")),
		ink.Box(vdom.Props{"width": 20.0}, ink.Text(vdom.Props{"bold": true}, "Name")),
		ink.Box(vdom.Props{"width": 30.0}, ink.Text(vdom.Props{"bold": true}, "Email")),
	))

	for _, item := range data {
		rows = append(rows, ink.Box(nil,
			ink.Box(vdom.Props{"width": 10.0}, ink.Text(fmt.Sprintf("%d", item.ID))),
			ink.Box(vdom.Props{"width": 20.0}, ink.Text(item.Name)),
			ink.Box(vdom.Props{"width": 30.0}, ink.Text(item.Email)),
		))
	}

	return ink.Box(vdom.Props{"flexDirection": "column", "padding": 1.0},
		ink.Text(vdom.Props{"color": "yellow"}, "--- Table Test ---"),
		ink.Box(vdom.Props{"borderStyle": "single", "flexDirection": "column", "marginTop": 1.0}, rows...),
	)
}

func main() {
	app := ink.NewApp(TestTable)
	fmt.Println(app.RenderOnce())
}
