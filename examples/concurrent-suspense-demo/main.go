package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

type suspenseDataRecord struct {
	done chan struct{}
	data string
}

var concurrentSuspenseCache = struct {
	sync.Mutex
	records map[string]*suspenseDataRecord
}{
	records: make(map[string]*suspenseDataRecord),
}

func ConcurrentSuspenseDemo() *vdom.Node {
	app := ink.UseApp()
	showMoreRaw, setShowMore := ink.UseState(false)
	showMore := showMoreRaw.(bool)

	ink.UseEffect(func() func() {
		timer := time.AfterFunc(2*time.Second, func() {
			app.Schedule(func() {
				setShowMore(true)
			})
		})

		return func() {
			timer.Stop()
		}
	}, []interface{}{})

	children := []*vdom.Node{
		ink.Text(vdom.Props{"bold": true, "underline": true}, "Concurrent Suspense Demo"),
		ink.Text(vdom.Props{"dimColor": true}, "(With concurrent: true, Suspense re-renders automatically)"),
		ink.Box(vdom.Props{"marginTop": 1}),
		ink.Text("Fast data (200ms):"),
		ink.Suspense(loadingLine("Loading fast data..."), func() *vdom.Node {
			return dataLine(fetchData("fast", 200*time.Millisecond))
		}),
		ink.Box(vdom.Props{"marginTop": 1}),
		ink.Text("Medium data (800ms):"),
		ink.Suspense(loadingLine("Loading medium data..."), func() *vdom.Node {
			return dataLine(fetchData("medium", 800*time.Millisecond))
		}),
		ink.Box(vdom.Props{"marginTop": 1}),
		ink.Text("Slow data (1500ms):"),
		ink.Suspense(loadingLine("Loading slow data..."), func() *vdom.Node {
			return dataLine(fetchData("slow", 1500*time.Millisecond))
		}),
	}

	if showMore {
		children = append(children,
			ink.Box(vdom.Props{"marginTop": 1}),
			ink.Text("Dynamically added (500ms):"),
			ink.Suspense(loadingLine("Loading dynamic data..."), func() *vdom.Node {
				return dataLine(fetchData("dynamic", 500*time.Millisecond))
			}),
		)
	}

	return ink.Box(vdom.Props{"flexDirection": "column"}, children...)
}

func fetchData(key string, delay time.Duration) string {
	concurrentSuspenseCache.Lock()
	record, exists := concurrentSuspenseCache.records[key]
	if exists && record.data != "" {
		data := record.data
		concurrentSuspenseCache.Unlock()
		return data
	}

	if !exists {
		record = &suspenseDataRecord{done: make(chan struct{})}
		concurrentSuspenseCache.records[key] = record
		go func(record *suspenseDataRecord) {
			time.Sleep(delay)
			concurrentSuspenseCache.Lock()
			record.data = fmt.Sprintf("Data for %q (fetched in %dms)", key, delay.Milliseconds())
			close(record.done)
			concurrentSuspenseCache.Unlock()
		}(record)
	}
	done := record.done
	concurrentSuspenseCache.Unlock()

	ink.SuspendUntil(done)
	return ""
}

func loadingLine(message string) *vdom.Node {
	return ink.Box(vdom.Props{"marginLeft": 2},
		ink.Text(vdom.Props{"color": "yellow"}, message),
	)
}

func dataLine(data string) *vdom.Node {
	return ink.Box(vdom.Props{"marginLeft": 2},
		ink.Text(vdom.Props{"color": "green"}, data),
	)
}

func main() {
	resetConcurrentSuspenseCache()

	if !terminal.StdoutIsTerminal() {
		app := ink.NewApp(ConcurrentSuspenseDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(ConcurrentSuspenseDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	time.Sleep(2600 * time.Millisecond)
	if err := instance.Rerender(ConcurrentSuspenseDemo); err != nil {
		fmt.Println(err)
	}
}

func resetConcurrentSuspenseCache() {
	concurrentSuspenseCache.Lock()
	defer concurrentSuspenseCache.Unlock()

	concurrentSuspenseCache.records = make(map[string]*suspenseDataRecord)
}
