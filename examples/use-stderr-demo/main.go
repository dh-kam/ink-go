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

func UseStderrDemo() *vdom.Node {
	stderr := ink.UseStderr()

	ink.UseEffect(func() func() {
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					_, _ = stderr.Write("Hello from Ink to stderr\n")
				case <-done:
					return
				}
			}
		}()

		return func() {
			close(done)
		}
	}, []interface{}{})

	return ink.Text("Hello World")
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(UseStderrDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(UseStderrDemo, ink.RenderOptions{})
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
