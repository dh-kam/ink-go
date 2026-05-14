package main

import (
	"fmt"
	"time"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

type TestItem struct {
	ID    int
	Label string
}

func TestStatic() *vdom.Node {
	app := ink.UseApp()
	testsRaw, setTests := ink.UseState([]TestItem{})
	counterRaw, setCounter := ink.UseState(0)
	tests := testsRaw.([]TestItem)
	counter := counterRaw.(int)

	ink.UseEffect(func() func() {
		if counter >= 5 {
			app.Schedule(func() {
				app.Exit()
			})
			return nil
		}

		done := make(chan struct{})
		timer := time.NewTimer(500 * time.Millisecond)

		go func() {
			select {
			case <-timer.C:
				app.Schedule(func() {
					setTests(func(previous []TestItem) []TestItem {
						next := append([]TestItem{}, previous...)
						next = append(next, TestItem{
							ID:    counter,
							Label: fmt.Sprintf("Test #%d passed", counter+1),
						})
						return next
					})
					setCounter(func(previous int) int {
						return previous + 1
					})
				})
			case <-done:
			}
		}()

		return func() {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			close(done)
		}
	}, []interface{}{counter})

	status := fmt.Sprintf("Running tests... (%d/5)", counter)
	if counter >= 5 {
		status = "All tests complete!"
	}

	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Static(tests, func(test TestItem) *vdom.Node {
			return ink.Box(nil,
				ink.Text(vdom.Props{"color": "green"}, "✔ "+test.Label),
			)
		}),
		ink.Box(vdom.Props{"marginTop": 1.0},
			ink.Text(status),
		),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(TestStatic)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(TestStatic, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		fmt.Println(err)
	}
}
