package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

var suspenseReady atomic.Bool

func SuspenseDemo() *vdom.Node {
	if !suspenseReady.Load() {
		return ink.Text("Loading...")
	}
	return ink.Text("Hello World")
}

func main() {
	suspenseReady.Store(false)

	if !terminal.StdoutIsTerminal() {
		app := ink.NewApp(SuspenseDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(SuspenseDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	time.Sleep(500 * time.Millisecond)
	suspenseReady.Store(true)
	if err := instance.Rerender(SuspenseDemo); err != nil {
		fmt.Println(err)
	}
}
