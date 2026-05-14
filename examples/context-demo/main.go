package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

type Theme struct {
	Name      string
	Accent    string
	Secondary string
}

type User struct {
	Name string
	Role string
}

var (
	ThemeContext = ink.NewContext(Theme{Name: "light", Accent: "blue", Secondary: "gray"})
	UserContext  = ink.NewContext(User{Name: "anonymous", Role: "guest"})
)

// Toolbar reads both contexts. Uses two hooks during render.
func Toolbar(label string) *vdom.Node {
	theme := ink.UseContext(ThemeContext)
	user := ink.UseContext(UserContext)
	return components.Box(vdom.Props{},
		components.Text(fmt.Sprintf("%s: [%s] %s (%s)", label, theme.Name, user.Name, user.Role)),
	)
}

// App renders an outer toolbar with defaults and an inner toolbar inside
// nested Providers so the rendered tree shows both values side-by-side.
func App() *vdom.Node {
	outer := Toolbar("outer")

	var inner *vdom.Node
	ThemeContext.Provider(Theme{Name: "dark", Accent: "magenta", Secondary: "white"}, func() {
		UserContext.Provider(User{Name: "alice", Role: "admin"}, func() {
			inner = Toolbar("inner")
		})
	})

	return components.Box(vdom.Props{"flexDirection": "column"},
		outer,
		inner,
	)
}

func main() {
	app := ink.NewApp(App)
	output := app.RenderOnce()
	fmt.Println(output)
}
