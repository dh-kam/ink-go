package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

var renderThrottleCount int64

func RenderThrottleDemo() *vdom.Node {
	count := atomic.LoadInt64(&renderThrottleCount)

	return ink.Box(vdom.Props{"flexDirection": "column", "padding": 1},
		ink.Text(fmt.Sprintf("Counter: %d", count)),
		ink.Text("This updates every 10ms but renders are throttled"),
		ink.Text("Press Ctrl+C to exit"),
	)
}

func main() {
	atomic.StoreInt64(&renderThrottleCount, 0)

	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(RenderThrottleDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	maxFPS := 10
	instance, err := ink.RenderWithOptions(RenderThrottleDemo, ink.RenderOptions{
		MaxFPSLimit: &maxFPS,
	})
	if err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			atomic.AddInt64(&renderThrottleCount, 1)
			if err := instance.Rerender(RenderThrottleDemo); err != nil {
				fmt.Println(err)
				return
			}
		case <-signals:
			if err := instance.HandleInput("\x03"); err != nil {
				fmt.Println(err)
			}
			fmt.Println()
			return
		}
	}
}
