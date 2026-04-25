package main

import (
"fmt"

"github.com/dh-kam/goink.go/pkg/components"
"github.com/dh-kam/goink.go/pkg/ink"
"github.com/dh-kam/goink.go/pkg/vdom"
)

// App component using helper functions
func App() *vdom.Node {
return components.Box(nil,
components.Text("Welcome to Goink!"),
components.Newline(),
components.Newline(),
components.Text("This is a React-like framework for CLI apps in Go."),
components.Newline(),
components.Text("Built with TDD and clean architecture."),
)
}

func main() {
app := ink.NewApp(App)
output := app.RenderOnce()

fmt.Println(output)
}
