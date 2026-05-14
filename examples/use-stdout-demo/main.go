package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func UseStdoutDemo() *vdom.Node {
	stdout := ink.UseStdout()

	ink.UseEffect(func() func() {
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					_, _ = stdout.Write("Hello from Ink to stdout\n")
				case <-done:
					return
				}
			}
		}()

		return func() {
			close(done)
		}
	}, []interface{}{})

	return ink.Box(vdom.Props{"flexDirection": "column", "paddingX": 2, "paddingY": 1},
		ink.Text(vdom.Props{"bold": true, "underline": true}, "Terminal dimensions:"),
		ink.Box(vdom.Props{"marginTop": 1},
			ink.Text("Width: ", ink.Text(vdom.Props{"bold": true}, fmt.Sprintf("%d", stdout.Columns))),
		),
		ink.Box(nil,
			ink.Text("Height: ", ink.Text(vdom.Props{"bold": true}, fmt.Sprintf("%d", stdout.Rows))),
		),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(UseStdoutDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(UseStdoutDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	<-signals
	if err := instance.HandleInput("\x03"); err != nil {
		fmt.Println(err)
	}
	fmt.Println()
}
