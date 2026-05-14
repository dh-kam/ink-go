package main

import (
        "fmt"
        "github.com/dh-kam/ink-go/pkg/components"
        "github.com/dh-kam/ink-go/pkg/ink"
        "github.com/dh-kam/ink-go/pkg/vdom"
)

func App() *vdom.Node {
        return components.Box(vdom.Props{
                "width": 10.0,
                "height": 3.0,
                "borderStyle": "single",
        }, components.Text("Hi"))
}

func main() {
        app := ink.NewApp(App)
        fmt.Println(app.RenderOnce())
}
