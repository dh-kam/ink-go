package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type captureStdout struct {
	mu     sync.Mutex
	writes []string
}

func (stdout *captureStdout) Write(data []byte) (int, error) {
	stdout.mu.Lock()
	defer stdout.mu.Unlock()

	stdout.writes = append(stdout.writes, string(data))
	return len(data), nil
}

func (stdout *captureStdout) Columns() int {
	return 100
}

func (stdout *captureStdout) Rows() int {
	return 24
}

func (stdout *captureStdout) Snapshot() []string {
	stdout.mu.Lock()
	defer stdout.mu.Unlock()

	out := make([]string, len(stdout.writes))
	copy(out, stdout.writes)
	return out
}

func App() *vdom.Node {
	widthValue, setWidth := ink.UseState(0)
	ref := ink.UseRef((*ink.DOMElement)(nil))

	ink.UseEffect(func() func() {
		current, _ := ref.Current().(*ink.DOMElement)
		if current != nil {
			setWidth(ink.MeasureElement(current).Width)
		}

		return nil
	}, []interface{}{"measure"})

	return ink.Box(vdom.Props{"ref": ref},
		ink.Text(fmt.Sprintf("Width: %d", widthValue.(int))),
	)
}

func main() {
	stdout := &captureStdout{}
	instance, err := ink.RenderWithOptions(App, ink.RenderOptions{
		AppOptions: ink.AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		panic(err)
	}

	time.Sleep(150 * time.Millisecond)
	_ = instance.Cleanup()

	for _, write := range stdout.Snapshot() {
		fmt.Fprintln(os.Stdout, strings.TrimSuffix(write, "\n"))
	}
}
