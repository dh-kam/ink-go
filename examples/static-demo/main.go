package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

type testResult struct {
	ID    int
	Title string
}

var (
	testsMu sync.Mutex
	tests   []testResult
)

func StaticDemo() *vdom.Node {
	currentTests := snapshotTests()

	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Static(currentTests, func(test testResult) *vdom.Node {
			return ink.Box(nil,
				ink.Text(vdom.Props{"color": "green"}, "✔ "+test.Title),
			)
		}),
		ink.Box(vdom.Props{"marginTop": 1},
			ink.Text(vdom.Props{"dimColor": true}, fmt.Sprintf("Completed tests: %d", len(currentTests))),
		),
	)
}

func main() {
	resetTests()

	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(StaticDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(StaticDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	completed := 0
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		<-timer.C
		if completed >= 10 {
			return
		}

		completed++
		appendTest(completed)
		if _, err := ink.RenderWithOptions(StaticDemo, ink.RenderOptions{}); err != nil {
			fmt.Println(err)
			return
		}
		if completed == 10 {
			// Leave the final frame on screen while restoring terminal state.
			if err := instance.HandleInput("\x03"); err != nil {
				fmt.Println(err)
			}
			return
		}
		timer.Reset(100 * time.Millisecond)
	}
}

func resetTests() {
	testsMu.Lock()
	defer testsMu.Unlock()
	tests = nil
}

func appendTest(number int) {
	testsMu.Lock()
	defer testsMu.Unlock()
	tests = append(tests, testResult{
		ID:    number - 1,
		Title: fmt.Sprintf("Test #%d", number),
	})
}

func snapshotTests() []testResult {
	testsMu.Lock()
	defer testsMu.Unlock()

	copied := make([]testResult, len(tests))
	copy(copied, tests)
	return copied
}
