package ink

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

type resizeRecordingWriter struct {
	recordingWriter
	columns   int
	rows      int
	nextID    int
	listeners map[int]func()
}

func (writer *resizeRecordingWriter) Columns() int {
	return writer.columns
}

func (writer *resizeRecordingWriter) Rows() int {
	if writer.rows > 0 {
		return writer.rows
	}

	return 24
}

func (writer *resizeRecordingWriter) SubscribeResize(listener func()) func() {
	if writer.listeners == nil {
		writer.listeners = make(map[int]func())
	}

	id := writer.nextID
	writer.nextID++
	writer.listeners[id] = listener

	return func() {
		delete(writer.listeners, id)
	}
}

func (writer *resizeRecordingWriter) EmitResize() {
	callbacks := make([]func(), 0, len(writer.listeners))
	for _, listener := range writer.listeners {
		callbacks = append(callbacks, listener)
	}

	for _, listener := range callbacks {
		listener()
	}
}

func (writer *resizeRecordingWriter) ResizeListenerCount() int {
	return len(writer.listeners)
}

func TestRenderWithOptionsReusesInstanceForSameStdout(t *testing.T) {
	stdout := &recordingWriter{}

	first, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("first")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("first render failed: %v", err)
	}
	defer first.Unmount()

	second, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("second")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	if first != second {
		t.Fatal("expected managed render to reuse the same instance")
	}

	if len(stdout.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(stdout.writes))
	}

	if !strings.Contains(stdout.writes[1], ansiEraseLines(1)) || !strings.Contains(stdout.writes[1], "second") {
		t.Fatalf("expected rerender payload to erase and rewrite output, got %q", stdout.writes[1])
	}
}

func TestRenderWithOptionsRemeasuresTextChangeLikeUpstreamConcurrentFixture(t *testing.T) {
	stdout := &recordingWriter{}
	add := false

	render := func() *vdom.Node {
		value := "abc"
		if add {
			value = "abcx"
		}

		return components.Box(nil,
			components.Text(value),
		)
	}

	instance, err := RenderWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("initial render failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.last() != "abc" {
		t.Fatalf("expected initial concurrent-shaped output %q, got %#v", "abc", stdout.writes)
	}

	stdout.writes = nil
	add = true

	next, err := RenderWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("text-change rerender failed: %v", err)
	}

	if next != instance {
		t.Fatal("expected concurrent-shaped rerender to reuse the managed instance")
	}

	if stdout.last() != "abcx" {
		t.Fatalf("expected text-change rerender output %q, got %#v", "abcx", stdout.writes)
	}
}

func TestRenderWithOptionsRemeasuresNestedTextNodeChangeLikeUpstreamConcurrentFixture(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := RenderWithOptions(func() *vdom.Node {
		return components.Box(nil,
			components.Text("abc"),
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("initial render failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.last() != "abc" {
		t.Fatalf("expected initial concurrent-shaped output %q, got %#v", "abc", stdout.writes)
	}

	stdout.writes = nil

	next, err := RenderWithOptions(func() *vdom.Node {
		return components.Box(nil,
			components.Text(
				"abc",
				components.Text("x"),
			),
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("nested text-node rerender failed: %v", err)
	}

	if next != instance {
		t.Fatal("expected concurrent-shaped rerender to reuse the managed instance")
	}

	if stdout.last() != "abcx" {
		t.Fatalf("expected nested text-node rerender output %q, got %#v", "abcx", stdout.writes)
	}
}

func TestCleanupDetachesManagedInstance(t *testing.T) {
	stdout := &recordingWriter{}

	first, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("first")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("first render failed: %v", err)
	}
	defer first.Unmount()

	if err := first.Cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	second, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("second")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("second render failed: %v", err)
	}
	defer second.Unmount()

	if first == second {
		t.Fatal("expected cleanup to force a new managed instance")
	}
}

func TestRenderWithOptionsRerendersOnResizeImmediately(t *testing.T) {
	stdout := &resizeRecordingWriter{columns: 10, rows: 24}

	instance, err := RenderWithOptions(func() *vdom.Node {
		return components.Text(strings.Repeat("X", stdout.columns))
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.ResizeListenerCount() != 1 {
		t.Fatalf("expected one resize listener, got %d", stdout.ResizeListenerCount())
	}

	stdout.writes = nil
	stdout.columns = 5
	stdout.EmitResize()

	if len(stdout.writes) != 2 {
		t.Fatalf("expected resize to clear and rerender immediately, got %#v", stdout.writes)
	}

	if !strings.Contains(stdout.writes[0], ansiEraseLines(1)) {
		t.Fatalf("expected resize shrink to clear previous output, got %q", stdout.writes[0])
	}

	if stdout.writes[1] != "XXXXX\n" {
		t.Fatalf("expected resize rerender to use latest width immediately, got %q", stdout.writes[1])
	}
}

func TestRenderWithOptionsMaxFPSLimitUpdatesReusedInstance(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("first")
	}, RenderOptions{
		AppOptions:  AppOptions{Stdout: stdout},
		MaxFPSLimit: fpsLimit(1),
	})
	if err != nil {
		t.Fatalf("initial render failed: %v", err)
	}
	defer instance.Unmount()

	if instance.maxFPS != 1 || instance.renderThrottle != time.Second {
		t.Fatalf("expected initial max FPS override, got fps=%d throttle=%s", instance.maxFPS, instance.renderThrottle)
	}

	next, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("second")
	}, RenderOptions{
		AppOptions:  AppOptions{Stdout: stdout},
		MaxFPSLimit: fpsLimit(DefaultMaxFPS),
	})
	if err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	if next != instance {
		t.Fatal("expected managed render to reuse instance while updating max FPS")
	}

	if instance.maxFPS != DefaultMaxFPS || instance.renderThrottle != 34*time.Millisecond {
		t.Fatalf("expected reused instance to adopt upstream default max FPS, got fps=%d throttle=%s", instance.maxFPS, instance.renderThrottle)
	}

	next, err = RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("third")
	}, RenderOptions{
		AppOptions:  AppOptions{Stdout: stdout},
		MaxFPS:      1,
		MaxFPSLimit: fpsLimit(0),
	})
	if err != nil {
		t.Fatalf("third render failed: %v", err)
	}

	if next != instance {
		t.Fatal("expected managed render to keep reusing instance")
	}

	if instance.maxFPS != 0 || instance.renderThrottle != 0 {
		t.Fatalf("expected MaxFPSLimit: 0 to disable throttling on reused instance, got fps=%d throttle=%s", instance.maxFPS, instance.renderThrottle)
	}
}

func TestRenderWithOptionsThrownRenderErrorSurfacesViaWaitUntilExitLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}
	renderErr := errors.New("errored")

	instance, err := RenderWithOptions(func() *vdom.Node {
		panic(renderErr)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("expected render panic to exit through the session, got mount error %v", err)
	}

	if instance == nil {
		t.Fatal("expected render panic to still return an instance")
	}

	if err := instance.WaitUntilExit(); !errors.Is(err, renderErr) {
		t.Fatalf("expected waitUntilExit to return %v, got %v", renderErr, err)
	}

	if stdout.joined() != "" {
		t.Fatalf("expected render panic to avoid visible output writes, got %#v", stdout.writes)
	}
}

func TestRenderWithOptionsThrownRerenderErrorSurfacesViaWaitUntilExitLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}
	renderErr := errors.New("rerender errored")

	instance, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("safe")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("initial render failed: %v", err)
	}

	stdout.writes = nil

	next, err := RenderWithOptions(func() *vdom.Node {
		panic(renderErr)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("expected rerender panic to exit through the session, got rerender error %v", err)
	}

	if next != instance {
		t.Fatal("expected rerender panic to reuse the same managed instance")
	}

	if err := instance.WaitUntilExit(); !errors.Is(err, renderErr) {
		t.Fatalf("expected waitUntilExit to return %v, got %v", renderErr, err)
	}

	if len(stdout.writes) == 0 {
		t.Fatal("expected rerender panic to clear the previous frame during unmount")
	}

	if !strings.Contains(stdout.joined(), ansiEraseLines(1)) {
		t.Fatalf("expected rerender panic to erase the previous frame, got %#v", stdout.writes)
	}

	if strings.Contains(stdout.joined(), showCursorEscape) {
		t.Fatalf("expected non-TTY rerender panic unmount not to restore cursor, got %#v", stdout.writes)
	}
}

func TestRenderWithOptionsThrownRerenderErrorDisablesRawModeLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}
	stdin := &rawModeRecordingStdin{}
	renderErr := errors.New("raw mode rerender errored")

	instance, err := RenderWithOptions(func() *vdom.Node {
		stdinCtx := UseStdin()

		UseEffect(func() func() {
			if err := stdinCtx.SetRawMode(true); err != nil {
				t.Fatalf("enable raw mode failed: %v", err)
			}

			return nil
		}, []interface{}{"managed-raw-mode"})

		return vdom.CreateTextNode("safe")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
	})
	if err != nil {
		t.Fatalf("initial render failed: %v", err)
	}

	if instance.app.rawModeUsers != 1 {
		t.Fatalf("expected managed render to hold one raw mode user before rerender panic, got %d", instance.app.rawModeUsers)
	}

	next, err := RenderWithOptions(func() *vdom.Node {
		panic(renderErr)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
	})
	if err != nil {
		t.Fatalf("expected rerender panic to exit through the managed session, got %v", err)
	}

	if next != instance {
		t.Fatal("expected rerender panic to reuse the same managed instance")
	}

	if err := instance.WaitUntilExit(); !errors.Is(err, renderErr) {
		t.Fatalf("expected waitUntilExit to return %v, got %v", renderErr, err)
	}

	if instance.app.rawModeUsers != 0 {
		t.Fatalf("expected rerender panic exit to fully clear raw mode users, got %d", instance.app.rawModeUsers)
	}

	if instance.app.rawState != nil {
		t.Fatalf("expected rerender panic exit to restore raw terminal state, got %#v", instance.app.rawState)
	}
}

func TestRenderWithOptionsRemovesResizeListenerOnUnmount(t *testing.T) {
	stdout := &resizeRecordingWriter{columns: 10, rows: 24}

	instance, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if stdout.ResizeListenerCount() != 1 {
		t.Fatalf("expected one resize listener, got %d", stdout.ResizeListenerCount())
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if stdout.ResizeListenerCount() != 0 {
		t.Fatalf("expected resize listener cleanup on unmount, got %d", stdout.ResizeListenerCount())
	}
}

func TestRenderWithOptionsSkipsResizeListenerInCI(t *testing.T) {
	t.Setenv("CI", "true")

	stdout := &resizeRecordingWriter{columns: 10, rows: 24}

	instance, err := RenderWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.ResizeListenerCount() != 0 {
		t.Fatalf("expected CI mode to skip resize subscriptions, got %d", stdout.ResizeListenerCount())
	}
}

func TestRenderWithOptionsNilToMeasuredComponent(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := RenderWithOptions(nil, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("initial nil render failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil

	_, err = RenderWithOptions(func() *vdom.Node {
		width, setWidth := UseState(0)
		ref := UseRef((*DOMElement)(nil))

		node := vdom.CreateElement("box", vdom.Props{
			"ref":   ref,
			"width": float64(12),
		}, vdom.CreateTextNode(fmt.Sprintf("Width: %d", width.(int))))

		UseEffect(func() func() {
			current, _ := ref.Current().(*DOMElement)
			setWidth(MeasureElement(current).Width)
			return nil
		}, []interface{}{"measure"})

		return node
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("measured rerender failed: %v", err)
	}

	if len(stdout.writes) < 2 {
		t.Fatalf("expected initial and follow-up measured writes, got %#v", stdout.writes)
	}
	if stdout.writes[0] != "Width: 0" {
		t.Fatalf("expected initial measured output, got %q", stdout.writes[0])
	}
	if stdout.writes[len(stdout.writes)-1] != "Width: 12" {
		t.Fatalf("expected follow-up measured output, got %#v", stdout.writes)
	}
}

func TestRenderWithOptionsMeasuredComponentWhileThrottled(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := RenderWithOptions(nil, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("initial nil render failed: %v", err)
	}
	defer instance.Unmount()

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	_, err = RenderWithOptions(func() *vdom.Node {
		width, setWidth := UseState(0)
		ref := UseRef((*DOMElement)(nil))

		node := vdom.CreateElement("box", vdom.Props{
			"ref":   ref,
			"width": float64(12),
		}, vdom.CreateTextNode(fmt.Sprintf("Width: %d", width.(int))))

		UseEffect(func() func() {
			current, _ := ref.Current().(*DOMElement)
			setWidth(MeasureElement(current).Width)
			return nil
		}, []interface{}{"measure"})

		return node
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("throttled measured rerender failed: %v", err)
	}

	clock.Advance(1 * time.Second)

	if stdout.writes[len(stdout.writes)-1] != "Width: 12" {
		t.Fatalf("expected throttled measured output to converge to measured width, got %#v", stdout.writes)
	}
}

func TestRenderWithOptionsSwitchingToDebugClearsAndReplaysStaticOutput(t *testing.T) {
	stdout := &recordingWriter{}

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"A"}, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("Live"),
		)
	}

	instance, err := RenderWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("initial render failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil

	next, err := RenderWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("debug rerender failed: %v", err)
	}

	if next != instance {
		t.Fatal("expected managed render to reuse the same instance across mode changes")
	}

	if len(stdout.writes) < 2 {
		t.Fatalf("expected clear plus replayed debug output, got %#v", stdout.writes)
	}

	if !strings.Contains(stdout.writes[0], ansiEraseLines(3)) {
		t.Fatalf("expected mode switch to clear previous output, got %#v", stdout.writes)
	}

	if stdout.writes[len(stdout.writes)-1] != "A\nLive" {
		t.Fatalf("expected debug mode switch to replay full static output, got %#v", stdout.writes)
	}
}

func TestRenderWithOptionsSwitchingToIncrementalReplaysStaticOutput(t *testing.T) {
	stdout := &recordingWriter{}

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"A"}, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("Live"),
		)
	}

	instance, err := RenderWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("initial render failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil

	next, err := RenderWithOptions(render, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("incremental rerender failed: %v", err)
	}

	if next != instance {
		t.Fatal("expected managed render to reuse the same instance across mode changes")
	}

	if len(stdout.writes) < 4 {
		t.Fatalf("expected clear plus replayed static and dynamic output, got %#v", stdout.writes)
	}

	if !strings.Contains(stdout.writes[0], ansiEraseLines(3)) {
		t.Fatalf("expected mode switch to clear previous output, got %#v", stdout.writes)
	}

	hasStaticReplay := false
	for _, write := range stdout.writes {
		if write == "A\n" {
			hasStaticReplay = true
			break
		}
	}

	if !hasStaticReplay {
		t.Fatalf("expected mode switch to replay static output, got %#v", stdout.writes)
	}

	if stdout.writes[len(stdout.writes)-1] != "Live\n" {
		t.Fatalf("expected mode switch to replay dynamic output, got %#v", stdout.writes)
	}
}
