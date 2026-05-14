package main

import (
        "fmt"
        "github.com/dh-kam/ink-go/pkg/components"
        "github.com/dh-kam/ink-go/pkg/ink"
        "github.com/dh-kam/ink-go/pkg/vdom"
)

func Row(justify string) *vdom.Node {
        return components.Box(vdom.Props{"flexDirection": "row"},
                components.Text("["),
                components.Box(vdom.Props{
                        "justifyContent": justify,
                        "width":          20.0,
                        "height":         1.0,
                },
                        components.Text("X"),
                        components.Text("Y"),
                ),
                components.Text("] "),
                components.Text(justify),
        )
}

func App() *vdom.Node {
        return components.Box(vdom.Props{"flexDirection": "column"},
                Row("flex-start"),
                Row("flex-end"),
                Row("center"),
                Row("space-around"),
                Row("space-between"),
                Row("space-evenly"),
        )
}

func main() {
        app := ink.NewApp(App)
        fmt.Println(app.RenderOnce())
}
