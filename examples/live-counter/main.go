package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func LiveCounter() *vdom.Node {
	app := ink.UseApp()
	countValue, setCount := ink.UseState(0)
	count := countValue.(int)

	ink.UseEffect(func() func() {
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					app.Schedule(func() {
						setCount(func(previous int) int {
							return previous + 1
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

	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Text("=== Live Counter Demo ==="),
		ink.Text("\n"),
		ink.Text(fmt.Sprintf("Count: %d", count)),
		ink.Text("\n"),
		ink.Text("Counter is incrementing automatically..."),
		ink.Text("Press Ctrl+C to exit"),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(LiveCounter)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(LiveCounter, ink.RenderOptions{})
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
