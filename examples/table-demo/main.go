package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type tableUser struct {
	ID    int
	Name  string
	Email string
}

var tableUsers = []tableUser{
	{ID: 0, Name: "Darrin_Kassulke", Email: "Ignatius.Ritchie@hotmail.com"},
	{ID: 1, Name: "Rosalyn.Reichert36", Email: "Homer42@gmail.com"},
	{ID: 2, Name: "Luis_Johns", Email: "Bette.Little@yahoo.com"},
	{ID: 3, Name: "Joesph_Muller84", Email: "Lindsay_Murazik@yahoo.com"},
	{ID: 4, Name: "Charles38", Email: "Ethel.Ziemann@yahoo.com"},
	{ID: 5, Name: "Glenn_Langworth10", Email: "Kathryne17@gmail.com"},
	{ID: 6, Name: "Lois_Yundt20", Email: "Derick_Huels@hotmail.com"},
	{ID: 7, Name: "Ardella_Klein9", Email: "Philip16@yahoo.com"},
	{ID: 8, Name: "Mary_Torphy", Email: "Lance57@yahoo.com"},
	{ID: 9, Name: "Raven.Bode", Email: "Rhonda_Adams@hotmail.com"},
}

func TableDemo() *vdom.Node {
	children := []*vdom.Node{
		ink.Box(nil,
			tableCell("10%", "ID"),
			tableCell("50%", "Name"),
			tableCell("40%", "Email"),
		),
	}

	for _, user := range tableUsers {
		children = append(children, ink.Box(vdom.Props{"key": user.ID},
			tableCell("10%", fmt.Sprintf("%d", user.ID)),
			tableCell("50%", user.Name),
			tableCell("40%", user.Email),
		))
	}

	return ink.Box(vdom.Props{"flexDirection": "column", "width": 80}, children...)
}

func tableCell(width string, text string) *vdom.Node {
	return ink.Box(vdom.Props{"width": width}, ink.Text(text))
}

func main() {
	app := ink.NewApp(TableDemo)
	fmt.Println(app.RenderOnce())
}
