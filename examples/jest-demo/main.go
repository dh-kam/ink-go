package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

var jestPaths = []string{
	"tests/login.js",
	"tests/signup.js",
	"tests/forgot-password.js",
	"tests/reset-password.js",
	"tests/view-profile.js",
	"tests/edit-profile.js",
	"tests/delete-profile.js",
	"tests/posts.js",
	"tests/post.js",
	"tests/comments.js",
}

type jestTest struct {
	Path   string
	Status string
}

type jestState struct {
	StartTime      time.Time
	CompletedTests []jestTest
	RunningTests   []jestTest
	Started        bool
	Random         *rand.Rand
}

var jestStore = struct {
	sync.Mutex
	State jestState
}{}

var jestVersion int64

func JestDemo() *vdom.Node {
	app := ink.UseApp()
	versionRaw, setVersion := ink.UseState(int64(0))
	_ = versionRaw

	ink.UseEffect(func() func() {
		startJestQueue(app, func() {
			setVersion(atomic.AddInt64(&jestVersion, 1))
		})
		return nil
	}, []interface{}{})

	state := snapshotJestState()

	children := []*vdom.Node{
		ink.Static(state.CompletedTests, func(test jestTest) *vdom.Node {
			return renderJestTest(test)
		}),
	}

	if len(state.RunningTests) > 0 {
		running := make([]*vdom.Node, 0, len(state.RunningTests))
		for _, test := range state.RunningTests {
			running = append(running, renderJestTest(test))
		}
		children = append(children, ink.Box(vdom.Props{"flexDirection": "column", "marginTop": 1}, running...))
	}

	passed := 0
	failed := 0
	for _, test := range state.CompletedTests {
		switch test.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		}
	}

	children = append(children, renderJestSummary(jestSummary{
		IsFinished: len(state.RunningTests) == 0 && len(state.CompletedTests) == len(jestPaths),
		Passed:     passed,
		Failed:     failed,
		Time:       formatJestDuration(time.Since(state.StartTime)),
	}))

	return ink.Box(vdom.Props{"flexDirection": "column"}, children...)
}

type jestSummary struct {
	IsFinished bool
	Passed     int
	Failed     int
	Time       string
}

func renderJestTest(test jestTest) *vdom.Node {
	parts := strings.SplitN(test.Path, "/", 2)
	dir := parts[0]
	file := ""
	if len(parts) > 1 {
		file = parts[1]
	}

	return ink.Box(nil,
		ink.Text(vdom.Props{
			"color":           "black",
			"backgroundColor": backgroundForJestStatus(test.Status),
		}, fmt.Sprintf(" %s ", strings.ToUpper(test.Status))),
		ink.Box(vdom.Props{"marginLeft": 1},
			ink.Text(vdom.Props{"dimColor": true}, dir+"/"),
			ink.Text(vdom.Props{"bold": true, "color": "white"}, file),
		),
	)
}

func renderJestSummary(summary jestSummary) *vdom.Node {
	testSuiteChildren := []*vdom.Node{
		ink.Box(vdom.Props{"width": 14}, ink.Text(vdom.Props{"bold": true}, "Test Suites:")),
	}
	if summary.Failed > 0 {
		testSuiteChildren = append(testSuiteChildren, ink.Text(vdom.Props{"bold": true, "color": "red"}, fmt.Sprintf("%d failed, ", summary.Failed)))
	}
	if summary.Passed > 0 {
		testSuiteChildren = append(testSuiteChildren, ink.Text(vdom.Props{"bold": true, "color": "green"}, fmt.Sprintf("%d passed, ", summary.Passed)))
	}
	testSuiteChildren = append(testSuiteChildren, ink.Text(fmt.Sprintf("%d total", summary.Passed+summary.Failed)))

	children := []*vdom.Node{
		ink.Box(nil, testSuiteChildren...),
		ink.Box(nil,
			ink.Box(vdom.Props{"width": 14}, ink.Text(vdom.Props{"bold": true}, "Time:")),
			ink.Text(summary.Time),
		),
	}

	if summary.IsFinished {
		children = append(children, ink.Box(nil, ink.Text(vdom.Props{"dimColor": true}, "Ran all test suites.")))
	}

	return ink.Box(vdom.Props{"flexDirection": "column", "marginTop": 1}, children...)
}

func backgroundForJestStatus(status string) string {
	switch status {
	case "runs":
		return "yellow"
	case "pass":
		return "green"
	case "fail":
		return "red"
	default:
		return ""
	}
}

func snapshotJestState() jestState {
	jestStore.Lock()
	defer jestStore.Unlock()

	state := jestStore.State
	state.CompletedTests = append([]jestTest(nil), state.CompletedTests...)
	state.RunningTests = append([]jestTest(nil), state.RunningTests...)
	return state
}

func startJestQueue(app ink.AppContext, rerender func()) {
	jestStore.Lock()
	if jestStore.State.Started {
		jestStore.Unlock()
		return
	}

	jestStore.State.Started = true
	if jestStore.State.StartTime.IsZero() {
		jestStore.State.StartTime = time.Now()
	}
	if jestStore.State.Random == nil {
		jestStore.State.Random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	jestStore.Unlock()

	jobs := make(chan string)
	var workers sync.WaitGroup
	for i := 0; i < 4; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for path := range jobs {
				runJestTest(app, rerender, path)
			}
		}()
	}

	go func() {
		for _, path := range jestPaths {
			jobs <- path
		}
		close(jobs)
		workers.Wait()
		app.Schedule(func() {
			if rerender != nil {
				rerender()
			}
			app.Exit()
		})
	}()
}

func runJestTest(app ink.AppContext, rerender func(), path string) {
	scheduleJestWork(app, func() {
		jestStore.Lock()
		jestStore.State.RunningTests = append(jestStore.State.RunningTests, jestTest{Status: "runs", Path: path})
		jestStore.Unlock()
		if rerender != nil {
			rerender()
		}
	})

	time.Sleep(randomJestDelay())

	status := randomJestStatus()
	scheduleJestWork(app, func() {
		jestStore.Lock()
		defer jestStore.Unlock()

		nextRunning := make([]jestTest, 0, len(jestStore.State.RunningTests))
		for _, test := range jestStore.State.RunningTests {
			if test.Path != path {
				nextRunning = append(nextRunning, test)
			}
		}
		jestStore.State.RunningTests = nextRunning
		jestStore.State.CompletedTests = append(jestStore.State.CompletedTests, jestTest{Status: status, Path: path})

		if rerender != nil {
			rerender()
		}
	})
}

func scheduleJestWork(app ink.AppContext, work func()) {
	done := make(chan struct{})
	app.Schedule(func() {
		defer close(done)
		if work != nil {
			work()
		}
	})
	<-done
}

func randomJestDelay() time.Duration {
	jestStore.Lock()
	defer jestStore.Unlock()

	return time.Duration(jestStore.State.Random.Intn(1000)) * time.Millisecond
}

func randomJestStatus() string {
	jestStore.Lock()
	defer jestStore.Unlock()

	if jestStore.State.Random.Float64() < 0.5 {
		return "pass"
	}
	return "fail"
}

func formatJestDuration(duration time.Duration) string {
	if duration < time.Second {
		return fmt.Sprintf("%dms", duration.Milliseconds())
	}

	seconds := int(duration.Round(time.Second).Seconds())
	return fmt.Sprintf("%ds", seconds)
}

func resetJestState() {
	jestStore.Lock()
	defer jestStore.Unlock()

	jestStore.State = jestState{
		StartTime: time.Now(),
		Random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	atomic.StoreInt64(&jestVersion, 0)
}

func main() {
	resetJestState()

	if !terminal.StdoutIsTerminal() && os.Getenv("GOINK_JEST_STREAM") == "" {
		app := ink.NewApp(JestDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(JestDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		fmt.Println(err)
	}
}
