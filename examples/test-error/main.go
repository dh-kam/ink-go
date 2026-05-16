package main

import (
	"errors"
	"fmt"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
	return components.ErrorOverview(components.ErrorOverviewProps{
		Err: errors.New("Oh no"),
	})
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
