package main

import (
	"fmt"
	"strings"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/utils"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func CursorIMEDemo() *vdom.Node {
	textRaw, setText := ink.UseState("")
	text := textRaw.(string)
	cursor := ink.UseCursor()

	ink.UseInput(func(input string, key ink.InputKey) {
		if key.Backspace || key.Delete {
			setText(func(previous string) string {
				return dropLastGrapheme(previous)
			})
			return
		}
		if !key.Ctrl && !key.Meta && input != "" {
			setText(func(previous string) string {
				return previous + input
			})
		}
	})

	prompt := "> "
	cursor.SetCursorPosition(&ink.CursorPosition{
		X: utils.StringWidth(prompt + text),
		Y: 1,
	})

	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Text("Type Korean (Ctrl+C to exit):"),
		ink.Text(prompt+text),
	)
}

func dropLastGrapheme(text string) string {
	clusters := utils.GraphemeClusters(text)
	if len(clusters) == 0 {
		return ""
	}
	return strings.Join(clusters[:len(clusters)-1], "")
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(CursorIMEDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(CursorIMEDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		panic(err)
	}
}
