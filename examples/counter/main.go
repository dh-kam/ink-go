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

var counterValue int64

func Counter() *vdom.Node {
	count := atomic.LoadInt64(&counterValue)
	return ink.Text(vdom.Props{"color": "green"}, fmt.Sprintf("%d tests passed", count))
}

func main() {
	atomic.StoreInt64(&counterValue, 0)

	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(Counter)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(Counter, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			atomic.AddInt64(&counterValue, 1)
			if _, err := ink.RenderWithOptions(Counter, ink.RenderOptions{}); err != nil {
				fmt.Println(err)
				return
			}
		case <-signals:
			if err := instance.HandleInput("\x03"); err != nil {
				fmt.Println(err)
			}
			return
		}
	}
}
