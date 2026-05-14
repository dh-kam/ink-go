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

func TestDisplay() *vdom.Node {
	app := ink.UseApp()
	showValue, setShow := ink.UseState(true)
	show := showValue.(bool)

	ink.UseEffect(func() func() {
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					app.Schedule(func() {
						setShow(func(previous bool) bool {
							return !previous
						})
					})
				case <-done:
					return
				}
			}
		}()

		return func() {
			close(done)
		}
	}, []interface{}{})

	display := "flex"
	if !show {
		display = "none"
	}

	return ink.Box(vdom.Props{"flexDirection": "column", "gap": 1.0},
		ink.Text(vdom.Props{"color": "yellow"}, "--- Display None Test ---"),
		ink.Text("The box below should toggle every 1s:"),
		ink.Box(vdom.Props{"display": display, "borderStyle": "single", "backgroundColor": "blue"},
			ink.Text(" I AM VISIBLE "),
		),
		ink.Box(vdom.Props{"borderStyle": "single"},
			ink.Text(" This box is always visible "),
		),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(TestDisplay)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(TestDisplay, ink.RenderOptions{})
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
