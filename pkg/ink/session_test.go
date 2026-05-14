package ink

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type recordingWriter struct {
	writes []string
}

func (writer *recordingWriter) Write(data []byte) (int, error) {
	writer.writes = append(writer.writes, string(data))
	return len(data), nil
}

func (writer *recordingWriter) last() string {
	if len(writer.writes) == 0 {
		return ""
	}

	return writer.writes[len(writer.writes)-1]
}

func (writer *recordingWriter) joined() string {
	return strings.Join(writer.writes, "")
}

type ttyRecordingWriter struct {
	recordingWriter
}

func (writer *ttyRecordingWriter) IsTTY() bool {
	return true
}

type ttySizedRecordingWriter struct {
	ttyRecordingWriter
	rows int
}

func (writer *ttySizedRecordingWriter) Rows() int {
	return writer.rows
}

type bufferedFlushWriter struct {
	pending    strings.Builder
	flushes    []string
	flushCount int
	flushErr   error
}

func (writer *bufferedFlushWriter) Write(data []byte) (int, error) {
	_, _ = writer.pending.Write(data)
	return len(data), nil
}

func (writer *bufferedFlushWriter) Flush() error {
	writer.flushCount++
	if writer.flushErr != nil {
		return writer.flushErr
	}

	if writer.pending.Len() > 0 {
		writer.flushes = append(writer.flushes, writer.pending.String())
		writer.pending.Reset()
	}

	return nil
}

func (writer *bufferedFlushWriter) joined() string {
	return strings.Join(writer.flushes, "")
}

type barrierWaitWriter struct {
	recordingWriter
	waitErr     error
	waitCalls   int
	releaseCh   chan struct{}
	waitStarted chan struct{}
	startOnce   sync.Once
}

func newBarrierWaitWriter() *barrierWaitWriter {
	return &barrierWaitWriter{
		releaseCh:   make(chan struct{}),
		waitStarted: make(chan struct{}),
	}
}

func (writer *barrierWaitWriter) Wait() error {
	writer.waitCalls++
	writer.startOnce.Do(func() {
		close(writer.waitStarted)
	})
	<-writer.releaseCh
	return writer.waitErr
}

func (writer *barrierWaitWriter) Release() {
	select {
	case <-writer.releaseCh:
	default:
		close(writer.releaseCh)
	}
}

type inputRecordingStdin struct {
	nextID    int
	listeners map[int]func(string)
}

type rawModeRecordingStdin struct {
	inputRecordingStdin
}

func (stdin *rawModeRecordingStdin) Fd() int {
	return 0
}

func (stdin *rawModeRecordingStdin) IsTTY() bool {
	return true
}

func (stdin *inputRecordingStdin) Read(data []byte) (int, error) {
	return 0, io.EOF
}

func (stdin *inputRecordingStdin) SubscribeInput(listener func(string)) func() {
	if stdin.listeners == nil {
		stdin.listeners = make(map[int]func(string))
	}

	id := stdin.nextID
	stdin.nextID++
	stdin.listeners[id] = listener

	return func() {
		delete(stdin.listeners, id)
	}
}

func (stdin *inputRecordingStdin) EmitInput(data string) {
	callbacks := make([]func(string), 0, len(stdin.listeners))
	for _, listener := range stdin.listeners {
		callbacks = append(callbacks, listener)
	}

	for _, listener := range callbacks {
		listener(data)
	}
}

func (stdin *inputRecordingStdin) InputListenerCount() int {
	return len(stdin.listeners)
}

type fakeScheduledTimer struct {
	when    time.Time
	fn      func()
	stopped bool
	fired   bool
}

func (timer *fakeScheduledTimer) Stop() bool {
	if timer.stopped || timer.fired {
		return false
	}

	timer.stopped = true
	return true
}

type fakeThrottleClock struct {
	now    time.Time
	timers []*fakeScheduledTimer
}

func newFakeThrottleClock() *fakeThrottleClock {
	return &fakeThrottleClock{now: time.Unix(0, 0)}
}

func (clock *fakeThrottleClock) Now() time.Time {
	return clock.now
}

func (clock *fakeThrottleClock) AfterFunc(delay time.Duration, fn func()) scheduledTimer {
	timer := &fakeScheduledTimer{
		when: clock.now.Add(delay),
		fn:   fn,
	}
	clock.timers = append(clock.timers, timer)
	return timer
}

func (clock *fakeThrottleClock) Advance(delay time.Duration) {
	clock.now = clock.now.Add(delay)

	for {
		fired := false
		for _, timer := range clock.timers {
			if timer.stopped || timer.fired || timer.when.After(clock.now) {
				continue
			}

			timer.fired = true
			fired = true
			timer.fn()
		}

		if !fired {
			return
		}
	}
}

func attachFakeThrottleClock(instance *Instance, clock *fakeThrottleClock) {
	instance.now = clock.Now
	instance.afterFunc = clock.AfterFunc
	instance.lastRenderAt = clock.Now()
}

func fpsLimit(value int) *int {
	return &value
}

type lastWriteRecorder interface {
	last() string
}

func mountSessionForTest(t *testing.T, stdout io.Writer, render func() *vdom.Node) *Instance {
	t.Helper()

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	return instance
}

func assertLastWriteMatches(t *testing.T, actual lastWriteRecorder, expected lastWriteRecorder, label string) {
	t.Helper()

	if actual.last() != expected.last() {
		t.Fatalf("%s mismatch\nactual:   %q\nexpected: %q", label, actual.last(), expected.last())
	}
}

func TestMountWritesInitialOutput(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) != 1 {
		t.Fatalf("expected 1 write, got %d", len(stdout.writes))
	}

	if stdout.writes[0] != "Hello\n" {
		t.Fatalf("expected first write to contain output, got %q", stdout.writes[0])
	}
}

func TestMountEmptyOutputSkipsInitialWrite(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return nil
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) != 0 {
		t.Fatalf("expected empty initial render to avoid writes, got %#v", stdout.writes)
	}
}

func TestClearErasesOutputLikeUpstreamFixture(t *testing.T) {
	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("B"),
			components.Text("C"),
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil

	if err := instance.Clear(); err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	if len(stdout.writes) != 1 {
		t.Fatalf("expected a single erase payload, got %#v", stdout.writes)
	}

	if stdout.writes[0] != ansiEraseLines(4) {
		t.Fatalf("expected upstream-style eraseLines(4) payload, got %q", stdout.writes[0])
	}
}

// TestMountedRerenderUnchangedTreeHitsCache exercises the reconciler tracker
// short-circuit: a Rerender() with a structurally identical component must
// not emit any new bytes to stdout, and the tracker's sections cache must
// record a hit. This is the core integration point for the renderer cache.
func TestMountedRerenderUnchangedTreeHitsCache(t *testing.T) {
	stdout := &recordingWriter{}

	component := func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}

	instance, err := MountWithOptions(component, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	initialWrites := len(stdout.writes)
	if initialWrites == 0 {
		t.Fatalf("expected at least one initial write")
	}

	hitsBefore, missesBefore := instance.renderCache.SectionsStats()

	if err := instance.Rerender(component); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	hitsAfter, missesAfter := instance.renderCache.SectionsStats()
	if hitsAfter != hitsBefore+1 {
		t.Fatalf("expected one cache hit on idle rerender, hits before=%d after=%d", hitsBefore, hitsAfter)
	}
	if missesAfter != missesBefore {
		t.Fatalf("expected miss count unchanged on idle rerender, before=%d after=%d", missesBefore, missesAfter)
	}

	if len(stdout.writes) != initialWrites {
		t.Fatalf("expected no extra writes on idle rerender, before=%d after=%d (extra=%q)",
			initialWrites, len(stdout.writes), stdout.writes[initialWrites:])
	}
}

func TestInstanceRerenderClearsPreviousOutput(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("World")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(stdout.writes))
	}

	if !strings.Contains(stdout.writes[1], ansiEraseLines(2)) {
		t.Fatalf("expected rerender to erase previous output, got %q", stdout.writes[1])
	}

	if !strings.Contains(stdout.writes[1], "World") {
		t.Fatalf("expected rerender to write new output, got %q", stdout.writes[1])
	}
}

func TestReconcilerUpdateChildMatchesDirectRender(t *testing.T) {
	stdoutActual := &recordingWriter{}
	stdoutExpected := &recordingWriter{}

	actual := mountSessionForTest(t, stdoutActual, func() *vdom.Node {
		return components.Text("A")
	})
	defer actual.Unmount()

	expected := mountSessionForTest(t, stdoutExpected, func() *vdom.Node {
		return components.Text("A")
	})
	defer expected.Unmount()

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "initial render")

	if err := actual.Rerender(func() *vdom.Node {
		return components.Text("B")
	}); err != nil {
		t.Fatalf("actual rerender failed: %v", err)
	}

	if err := expected.Rerender(func() *vdom.Node {
		return components.Text("B")
	}); err != nil {
		t.Fatalf("expected rerender failed: %v", err)
	}

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "updated child render")
}

func TestReconcilerUpdateTextNodeMatchesCombinedText(t *testing.T) {
	stdoutActual := &recordingWriter{}
	stdoutExpected := &recordingWriter{}

	actual := mountSessionForTest(t, stdoutActual, func() *vdom.Node {
		return components.Box(nil,
			components.Text("Hello "),
			components.Text("A"),
		)
	})
	defer actual.Unmount()

	expected := mountSessionForTest(t, stdoutExpected, func() *vdom.Node {
		return components.Text("Hello A")
	})
	defer expected.Unmount()

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "initial text-node render")

	if err := actual.Rerender(func() *vdom.Node {
		return components.Box(nil,
			components.Text("Hello "),
			components.Text("B"),
		)
	}); err != nil {
		t.Fatalf("actual rerender failed: %v", err)
	}

	if err := expected.Rerender(func() *vdom.Node {
		return components.Text("Hello B")
	}); err != nil {
		t.Fatalf("expected rerender failed: %v", err)
	}

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "updated text-node render")
}

func TestReconcilerAppendChildMatchesDirectRender(t *testing.T) {
	stdoutActual := &recordingWriter{}
	stdoutExpected := &recordingWriter{}

	actual := mountSessionForTest(t, stdoutActual, func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
		)
	})
	defer actual.Unmount()

	expected := mountSessionForTest(t, stdoutExpected, func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
		)
	})
	defer expected.Unmount()

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "initial appended render")

	if err := actual.Rerender(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("B"),
		)
	}); err != nil {
		t.Fatalf("actual rerender failed: %v", err)
	}

	if err := expected.Rerender(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("B"),
		)
	}); err != nil {
		t.Fatalf("expected rerender failed: %v", err)
	}

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "append child render")
}

func TestReconcilerInsertAndRemoveChildrenMatchDirectRender(t *testing.T) {
	stdoutActual := &recordingWriter{}
	stdoutExpected := &recordingWriter{}

	actual := mountSessionForTest(t, stdoutActual, func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text(vdom.Props{"key": "a"}, "A"),
			components.Text(vdom.Props{"key": "c"}, "C"),
		)
	})
	defer actual.Unmount()

	expected := mountSessionForTest(t, stdoutExpected, func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("C"),
		)
	})
	defer expected.Unmount()

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "initial insert/remove render")

	if err := actual.Rerender(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text(vdom.Props{"key": "a"}, "A"),
			components.Text(vdom.Props{"key": "b"}, "B"),
			components.Text(vdom.Props{"key": "c"}, "C"),
		)
	}); err != nil {
		t.Fatalf("actual insert rerender failed: %v", err)
	}

	if err := expected.Rerender(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("B"),
			components.Text("C"),
		)
	}); err != nil {
		t.Fatalf("expected insert rerender failed: %v", err)
	}

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "insert child render")

	if err := actual.Rerender(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text(vdom.Props{"key": "a"}, "A"),
		)
	}); err != nil {
		t.Fatalf("actual remove rerender failed: %v", err)
	}

	if err := expected.Rerender(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
		)
	}); err != nil {
		t.Fatalf("expected remove rerender failed: %v", err)
	}

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "remove child render")
}

func TestReconcilerReordersKeyedChildren(t *testing.T) {
	stdoutActual := &recordingWriter{}
	stdoutExpected := &recordingWriter{}

	actual := mountSessionForTest(t, stdoutActual, func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text(vdom.Props{"key": "a"}, "A"),
			components.Text(vdom.Props{"key": "b"}, "B"),
		)
	})
	defer actual.Unmount()

	expected := mountSessionForTest(t, stdoutExpected, func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("B"),
		)
	})
	defer expected.Unmount()

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "initial keyed render")

	if err := actual.Rerender(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text(vdom.Props{"key": "b"}, "B"),
			components.Text(vdom.Props{"key": "a"}, "A"),
		)
	}); err != nil {
		t.Fatalf("actual reorder rerender failed: %v", err)
	}

	if err := expected.Rerender(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("B"),
			components.Text("A"),
		)
	}); err != nil {
		t.Fatalf("expected reorder rerender failed: %v", err)
	}

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "reordered child render")
}

func TestReconcilerReplacesStyledChildWithText(t *testing.T) {
	stdoutActual := &ttyRecordingWriter{}
	stdoutExpected := &ttyRecordingWriter{}

	actual := mountSessionForTest(t, stdoutActual, func() *vdom.Node {
		return components.Text(
			components.Text(vdom.Props{"color": "green"}, "test"),
		)
	})
	defer actual.Unmount()

	expected := mountSessionForTest(t, stdoutExpected, func() *vdom.Node {
		return components.Text(vdom.Props{"color": "green"}, "test")
	})
	defer expected.Unmount()

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "initial styled child render")

	if err := actual.Rerender(func() *vdom.Node {
		return components.Text("x")
	}); err != nil {
		t.Fatalf("actual replace rerender failed: %v", err)
	}

	if err := expected.Rerender(func() *vdom.Node {
		return components.Text("x")
	}); err != nil {
		t.Fatalf("expected replace rerender failed: %v", err)
	}

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "replaced child render")
}

func TestReconcilerNestedTextGrowthMatchesDirectRender(t *testing.T) {
	stdoutActual := &recordingWriter{}
	stdoutExpected := &recordingWriter{}

	actual := mountSessionForTest(t, stdoutActual, func() *vdom.Node {
		return components.Box(nil,
			components.Text(
				"abc",
			),
		)
	})
	defer actual.Unmount()

	expected := mountSessionForTest(t, stdoutExpected, func() *vdom.Node {
		return components.Box(nil,
			components.Text("abc"),
		)
	})
	defer expected.Unmount()

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "initial nested text render")

	if err := actual.Rerender(func() *vdom.Node {
		return components.Box(nil,
			components.Text(
				"abc",
				components.Text("x"),
			),
		)
	}); err != nil {
		t.Fatalf("actual nested-text rerender failed: %v", err)
	}

	if err := expected.Rerender(func() *vdom.Node {
		return components.Box(nil,
			components.Text("abcx"),
		)
	}); err != nil {
		t.Fatalf("expected nested-text rerender failed: %v", err)
	}

	assertLastWriteMatches(t, stdoutActual, stdoutExpected, "grown nested text render")
}

func TestRerenderClearsTerminalWhenPreviousOutputFillsViewport(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 3}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("A\nB\nC")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 3,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("X\nY\nZ")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := stdout.last()
	if !strings.Contains(update, clearTerminalEscape) {
		t.Fatalf("expected fullscreen rerender to clear terminal, got %q", update)
	}

	if !strings.Contains(update, "X\nY\nZ") {
		t.Fatalf("expected fullscreen rerender to write next output, got %q", update)
	}
}

func TestDoNotEraseScreenLikeUpstreamFixture(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 4}
	value := "A\nB\nC"

	render := func() *vdom.Node {
		return vdom.CreateTextNode(value)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 4,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	value = "X\nY\nZ"

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := stdout.last()
	if strings.Contains(update, clearTerminalEscape) {
		t.Fatalf("expected non-fullscreen rerender to avoid clearTerminal, got %q", update)
	}

	if !strings.Contains(update, "X\nY\nZ") {
		t.Fatalf("expected non-fullscreen rerender to write next output, got %q", update)
	}
}

func TestStaticDoesNotTriggerClearTerminalWhenOnlyTotalOutputExceedsViewport(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 4}
	value := "D\nE\nF"

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"A", "B", "C"}, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text(value),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 4,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	value = "X\nY\nZ"

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	for _, write := range stdout.writes {
		if strings.Contains(write, clearTerminalEscape) {
			t.Fatalf("expected static rerender to avoid fullscreen clear, got %#v", stdout.writes)
		}
	}
}

func TestDoNotEraseScreenWhereStaticIsTallerThanViewportLikeUpstreamFixture(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 4}
	value := "done"

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"A", "B", "C", "D", "E"}, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text(value),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 4,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	value = "updated"

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	for _, write := range stdout.writes {
		if strings.Contains(write, clearTerminalEscape) {
			t.Fatalf("expected rerender with tall static history to avoid clearTerminal, got %#v", stdout.writes)
		}
	}

	if !strings.Contains(stdout.last(), "updated") {
		t.Fatalf("expected rerender to include updated interactive output, got %#v", stdout.writes)
	}
}

func TestEffectStateChangeTriggersFollowUpRerender(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		width, setWidth := UseState(0)
		ref := UseRef((*DOMElement)(nil))

		measuredNode := vdom.CreateElement("box", vdom.Props{
			"ref":    ref,
			"width":  float64(12),
			"height": float64(1),
		}, components.Text("body"))

		UseEffect(func() func() {
			current, _ := ref.Current().(*DOMElement)
			setWidth(MeasureElement(current).Width)
			return nil
		}, []interface{}{"measure"})

		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("Width: ", vdom.CreateTextNode(fmt.Sprintf("%d", width.(int)))),
			measuredNode,
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	joined := stdout.joined()
	if !strings.Contains(joined, "Width: 0") {
		t.Fatalf("expected initial render output in debug stream, got %q", joined)
	}
	if !strings.Contains(joined, "Width: 12") {
		t.Fatalf("expected follow-up rerender output in debug stream, got %q", joined)
	}
}

func TestOnRenderReportsInitialManualAndStateDrivenRenders(t *testing.T) {
	stdout := &recordingWriter{}
	stdin := &inputRecordingStdin{}
	renderTimes := make([]time.Duration, 0, 3)

	instance, err := MountWithOptions(func() *vdom.Node {
		count, setCount := UseState(0)

		UseInput(func(input interface{}, keys []string) bool {
			if input == "a" {
				setCount(count.(int) + 1)
			}

			return false
		})

		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("Count: ", vdom.CreateTextNode(fmt.Sprintf("%d", count.(int)))),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
		OnRender: func(metrics RenderMetrics) {
			renderTimes = append(renderTimes, metrics.RenderTime)
		},
		Debug: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(renderTimes) != 1 {
		t.Fatalf("expected initial onRender call, got %d", len(renderTimes))
	}
	if renderTimes[0] < 0 {
		t.Fatalf("expected non-negative initial render time, got %v", renderTimes[0])
	}

	if err := instance.Rerender(func() *vdom.Node {
		count, setCount := UseState(0)

		UseInput(func(input interface{}, keys []string) bool {
			if input == "a" {
				setCount(count.(int) + 1)
			}

			return false
		})

		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("Count: ", vdom.CreateTextNode(fmt.Sprintf("%d", count.(int)))),
			components.Text("Updated"),
		)
	}); err != nil {
		t.Fatalf("manual rerender failed: %v", err)
	}

	if len(renderTimes) != 2 {
		t.Fatalf("expected onRender after manual rerender, got %d", len(renderTimes))
	}
	if renderTimes[1] < 0 {
		t.Fatalf("expected non-negative manual rerender time, got %v", renderTimes[1])
	}

	if err := instance.HandleInput("a"); err != nil {
		t.Fatalf("input rerender failed: %v", err)
	}

	if len(renderTimes) != 3 {
		t.Fatalf("expected onRender after state-driven rerender, got %d", len(renderTimes))
	}
	if renderTimes[2] < 0 {
		t.Fatalf("expected non-negative state-driven render time, got %v", renderTimes[2])
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "Updated") {
		t.Fatalf("expected manual rerender output in debug stream, got %q", joined)
	}
	if !strings.Contains(joined, "Count: 1") {
		t.Fatalf("expected state-driven rerender output in debug stream, got %q", joined)
	}
}

func TestOnRenderUsesInstanceClockForRenderMetrics(t *testing.T) {
	stdout := &recordingWriter{}
	renderTimes := []time.Duration{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		OnRender: func(metrics RenderMetrics) {
			renderTimes = append(renderTimes, metrics.RenderTime)
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	renderTimes = nil
	base := time.Unix(10, 0)
	clockValues := []time.Time{
		base,
		base.Add(25 * time.Millisecond),
		base.Add(25 * time.Millisecond),
	}
	callIndex := 0
	instance.now = func() time.Time {
		if callIndex >= len(clockValues) {
			return clockValues[len(clockValues)-1]
		}

		value := clockValues[callIndex]
		callIndex++
		return value
	}

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("World")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(renderTimes) != 1 {
		t.Fatalf("expected one onRender call after rerender, got %d", len(renderTimes))
	}

	if renderTimes[0] != 25*time.Millisecond {
		t.Fatalf("expected renderTime to use injected clock, got %v", renderTimes[0])
	}
}

func TestHandleInputTriggersStateUpdateRerender(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		count, setCount := UseState(0)

		UseInput(func(input interface{}, keys []string) bool {
			if input == "a" {
				setCount(count.(int) + 1)
			}

			return false
		})

		return components.Text(fmt.Sprintf("Count: %d", count.(int)))
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("a"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "Count: 0") {
		t.Fatalf("expected initial output in debug stream, got %q", joined)
	}
	if !strings.Contains(joined, "Count: 1") {
		t.Fatalf("expected rerendered count in debug stream, got %q", joined)
	}
}

func TestUseTransitionSchedulesDeferredWorkAfterUrgentRender(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		query, setQuery := UseState("")
		deferred, setDeferred := UseState("")
		isPending, startTransition := UseTransition()

		UseInput(func(input string, key InputKey) {
			if input != "a" {
				return
			}

			setQuery(input)
			startTransition(func() {
				setDeferred(input)
			})
		})

		return components.Text(fmt.Sprintf("query=%s deferred=%s pending=%t", query.(string), deferred.(string), isPending))
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	clock := newFakeThrottleClock()
	attachFakeThrottleClock(instance, clock)

	if err := instance.HandleInput("a"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "query=a deferred= pending=true") {
		t.Fatalf("expected urgent render to keep deferred value stale and mark pending, got %q", joined)
	}
	if strings.Contains(joined, "query=a deferred=a pending=false") {
		t.Fatalf("deferred transition committed before scheduler tick: %q", joined)
	}

	clock.Advance(time.Millisecond)

	joined = stdout.joined()
	if !strings.Contains(joined, "query=a deferred=a pending=false") {
		t.Fatalf("expected scheduled transition commit after timer tick, got %q", joined)
	}
}

func TestUseDeferredValueLagsUntilSchedulerTick(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		value, setValue := UseState("")
		deferred := UseDeferredValue(value.(string))

		UseInput(func(input string, key InputKey) {
			if input == "a" {
				setValue(input)
			}
		})

		return components.Text(fmt.Sprintf("value=%s deferred=%s", value.(string), deferred))
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	clock := newFakeThrottleClock()
	attachFakeThrottleClock(instance, clock)

	if err := instance.HandleInput("a"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "value=a deferred=") {
		t.Fatalf("expected urgent render with stale deferred value, got %q", joined)
	}
	if strings.Contains(joined, "value=a deferred=a") {
		t.Fatalf("deferred value committed before scheduler tick: %q", joined)
	}

	clock.Advance(time.Millisecond)

	joined = stdout.joined()
	if !strings.Contains(joined, "value=a deferred=a") {
		t.Fatalf("expected deferred value commit after timer tick, got %q", joined)
	}
}

func TestSuspenseRendersFallbackAndRerendersWhenDoneCloses(t *testing.T) {
	stdout := &recordingWriter{}
	done := make(chan struct{})

	var resolvedMu sync.Mutex
	resolved := false
	isResolved := func() bool {
		resolvedMu.Lock()
		defer resolvedMu.Unlock()
		return resolved
	}
	markResolved := func() {
		resolvedMu.Lock()
		resolved = true
		resolvedMu.Unlock()
	}

	instance, err := MountWithOptions(func() *vdom.Node {
		return Suspense(components.Text("loading"), func() *vdom.Node {
			if !isResolved() {
				SuspendUntil(done)
			}

			return components.Text("ready")
		})
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if joined := stdout.joined(); !strings.Contains(joined, "loading") {
		t.Fatalf("expected suspense fallback before promise resolves, got %q", joined)
	}

	markResolved()
	close(done)

	deadline := time.Now().Add(200 * time.Millisecond)
	for {
		if strings.Contains(stdout.joined(), "ready") {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("expected suspense to rerender after done closes, got %q", stdout.joined())
		}

		time.Sleep(5 * time.Millisecond)
	}
}

func TestSuspendedRenderDoesNotLeakCursorPositionToFallback(t *testing.T) {
	stdout := &recordingWriter{}
	done := make(chan struct{})

	instance, err := MountWithOptions(func() *vdom.Node {
		return Suspense(components.Text("loading"), func() *vdom.Node {
			UseCursor().SetCursorPosition(&CursorPosition{X: 5, Y: 0})
			SuspendUntil(done)
			return components.Text("loaded")
		})
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if joined := stdout.joined(); strings.Contains(joined, showCursorEscape) {
		t.Fatalf("expected fallback output without leaked cursor escape, got %q", joined)
	}
}

func TestHandleInputParsesSpecialKeysForHooks(t *testing.T) {
	stdout := &recordingWriter{}
	var receivedInput interface{}
	var receivedKeys []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input interface{}, keys []string) bool {
			receivedInput = input
			receivedKeys = append([]string(nil), keys...)
			return false
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b[A"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if receivedInput != "" {
		t.Fatalf("expected special key input text to be empty, got %#v", receivedInput)
	}
	if strings.Join(receivedKeys, ",") != "up" {
		t.Fatalf("expected up-arrow key payload, got %#v", receivedKeys)
	}
}

func TestHandleInputProvidesKeyObjectSignature(t *testing.T) {
	stdout := &recordingWriter{}
	var receivedInput string
	var receivedKey InputKey

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			receivedInput = input
			receivedKey = key
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b[D"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if receivedInput != "" {
		t.Fatalf("expected special key text input to be empty, got %q", receivedInput)
	}
	if !receivedKey.LeftArrow {
		t.Fatal("expected leftArrow flag to be set")
	}
	if receivedKey.RightArrow || receivedKey.Ctrl || receivedKey.Tab {
		t.Fatalf("unexpected extra key flags: %+v", receivedKey)
	}
}

func TestHandleInputProvidesMetaModifierFlags(t *testing.T) {
	stdout := &recordingWriter{}
	var receivedInput string
	var receivedKey InputKey

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			receivedInput = input
			receivedKey = key
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b\x1b[A"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if receivedInput != "" {
		t.Fatalf("expected empty input text for meta+arrow, got %q", receivedInput)
	}
	if !receivedKey.Meta || !receivedKey.UpArrow {
		t.Fatalf("expected meta+up-arrow flags, got %+v", receivedKey)
	}
}

func TestHandleInputProvidesRemainingKeyObjectMatrix(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		assertKey func(t *testing.T, key InputKey)
	}{
		{
			name: "page-down",
			raw:  "\x1b[6~",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.PageDown || key.PageUp || key.Meta || key.Ctrl {
					t.Fatalf("expected pageDown-only flags, got %+v", key)
				}
			},
		},
		{
			name: "page-up",
			raw:  "\x1b[5~",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.PageUp || key.PageDown || key.Meta || key.Ctrl {
					t.Fatalf("expected pageUp-only flags, got %+v", key)
				}
			},
		},
		{
			name: "home",
			raw:  "\x1b[H",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Home || key.End || key.Meta || key.Ctrl {
					t.Fatalf("expected home-only flags, got %+v", key)
				}
			},
		},
		{
			name: "end",
			raw:  "\x1b[F",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.End || key.Home || key.Meta || key.Ctrl {
					t.Fatalf("expected end-only flags, got %+v", key)
				}
			},
		},
		{
			name: "meta-down",
			raw:  "\x1b\x1b[B",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Meta || !key.DownArrow || key.UpArrow || key.Ctrl {
					t.Fatalf("expected meta+down-arrow flags, got %+v", key)
				}
			},
		},
		{
			name: "meta-left",
			raw:  "\x1b\x1b[D",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Meta || !key.LeftArrow || key.RightArrow || key.Ctrl {
					t.Fatalf("expected meta+left-arrow flags, got %+v", key)
				}
			},
		},
		{
			name: "meta-right",
			raw:  "\x1b\x1b[C",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Meta || !key.RightArrow || key.LeftArrow || key.Ctrl {
					t.Fatalf("expected meta+right-arrow flags, got %+v", key)
				}
			},
		},
		{
			name: "ctrl-down",
			raw:  "\x1b[1;5B",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Ctrl || !key.DownArrow || key.UpArrow || key.Meta {
					t.Fatalf("expected ctrl+down-arrow flags, got %+v", key)
				}
			},
		},
		{
			name: "ctrl-left",
			raw:  "\x1b[1;5D",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Ctrl || !key.LeftArrow || key.RightArrow || key.Meta {
					t.Fatalf("expected ctrl+left-arrow flags, got %+v", key)
				}
			},
		},
		{
			name: "ctrl-right",
			raw:  "\x1b[1;5C",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Ctrl || !key.RightArrow || key.LeftArrow || key.Meta {
					t.Fatalf("expected ctrl+right-arrow flags, got %+v", key)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			var receivedInput string
			var receivedKey InputKey

			instance, err := MountWithOptions(func() *vdom.Node {
				UseInput(func(input string, key InputKey) {
					receivedInput = input
					receivedKey = key
				})

				return components.Text("Input")
			}, RenderOptions{
				AppOptions: AppOptions{Stdout: stdout},
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.HandleInput(testCase.raw); err != nil {
				t.Fatalf("handle input failed: %v", err)
			}

			if receivedInput != "" {
				t.Fatalf("expected special key text input to be empty, got %q", receivedInput)
			}
			testCase.assertKey(t, receivedKey)
		})
	}
}

func TestHandleInputPreservesPastedTextPayload(t *testing.T) {
	stdout := &recordingWriter{}
	var receivedInput string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			receivedInput = input
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\rtest"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if receivedInput != "\rtest" {
		t.Fatalf("expected pasted payload to be preserved, got %q", receivedInput)
	}
}

func TestHandleInputIgnoresInactiveUseInputHook(t *testing.T) {
	stdout := &recordingWriter{}
	called := false

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			called = true
		}, InputOptions{IsActive: false})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("a"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if called {
		t.Fatal("expected inactive useInput hook not to receive input")
	}
}

func TestHandleInputEscapeBlursFocusedComponent(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("input-focus-1", true)
		secondFocused, _, _ := UseFocus("input-focus-2", false)

		label := "none"
		switch {
		case firstFocused():
			label = "first"
		case secondFocused():
			label = "second"
		}

		return components.Text(label)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout, Stdin: rawModeTestStdin{}},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "first") {
		t.Fatalf("expected initial focused output, got %q", joined)
	}
	if !strings.Contains(joined, "none") {
		t.Fatalf("expected escape to blur focus and rerender, got %q", joined)
	}
}

func TestHandleInputTabRerendersFocusedComponent(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("input-focus-1", true)
		secondFocused, _, _ := UseFocus("input-focus-2", false)

		label := "first"
		if secondFocused() {
			label = "second"
		} else if !firstFocused() {
			label = "none"
		}

		return components.Text(label)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "first") {
		t.Fatalf("expected initial focused label in debug stream, got %q", joined)
	}
	if !strings.Contains(joined, "second") {
		t.Fatalf("expected tab navigation to rerender second focus target, got %q", joined)
	}
}

func TestHandleInputShiftTabWrapsToLastFocusTarget(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("input-focus-1", true)
		secondFocused, _, _ := UseFocus("input-focus-2", false)
		thirdFocused, _, _ := UseFocus("input-focus-3", false)

		label := "none"
		switch {
		case firstFocused():
			label = "first"
		case secondFocused():
			label = "second"
		case thirdFocused():
			label = "third"
		}

		return components.Text(label)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout, Stdin: rawModeTestStdin{}},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b[Z"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "first") {
		t.Fatalf("expected initial focused output, got %q", joined)
	}
	if !strings.Contains(joined, "third") {
		t.Fatalf("expected shift-tab to wrap to the last focus target, got %q", joined)
	}
}

func TestHandleInputTabSkipsInactiveFocusComponents(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		firstFocused, _, _ := UseFocus(FocusOptions{
			ID:        "input-focus-1",
			AutoFocus: true,
		})
		secondFocused, _, _ := UseFocus(FocusOptions{
			ID:       "input-focus-2",
			IsActive: boolPtr(false),
		})
		thirdFocused, _, _ := UseFocus(FocusOptions{
			ID: "input-focus-3",
		})

		label := "none"
		switch {
		case firstFocused():
			label = "first"
		case secondFocused():
			label = "second"
		case thirdFocused():
			label = "third"
		}

		return components.Text(label)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout, Stdin: rawModeTestStdin{}},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if !strings.Contains(joined, "first") {
		t.Fatalf("expected initial focus output, got %q", joined)
	}
	if !strings.Contains(joined, "third") {
		t.Fatalf("expected tab navigation to skip inactive focus target, got %q", joined)
	}
}

func TestHandleInputTabDoesNotMoveFocusWhileFocusIsDisabled(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		firstFocused, _, _ := UseFocus("input-focus-1", true)
		secondFocused, _, _ := UseFocus("input-focus-2", false)
		manager := UseFocusManager()
		manager.DisableFocus()

		label := "none"
		switch {
		case firstFocused():
			label = "first"
		case secondFocused():
			label = "second"
		}

		return components.Text(label)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\t"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	joined := stdout.joined()
	if strings.Contains(joined, "second") {
		t.Fatalf("expected disabled focus management to block tab navigation, got %q", joined)
	}
}

func TestInputSubscriberAutoRerendersAndCleansUp(t *testing.T) {
	stdin := &inputRecordingStdin{}
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		count, setCount := UseState(0)

		UseInput(func(input interface{}, keys []string) bool {
			if input == "a" {
				setCount(count.(int) + 1)
			}

			return false
		})

		return components.Text(fmt.Sprintf("Count: %d", count.(int)))
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdin:  stdin,
			Stdout: stdout,
		},
		Debug: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if stdin.InputListenerCount() != 1 {
		t.Fatalf("expected one input listener after mount, got %d", stdin.InputListenerCount())
	}

	stdin.EmitInput("a")
	if !strings.Contains(stdout.joined(), "Count: 1") {
		t.Fatalf("expected auto input subscription to rerender output, got %q", stdout.joined())
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if stdin.InputListenerCount() != 0 {
		t.Fatalf("expected input listener cleanup on unmount, got %d", stdin.InputListenerCount())
	}
}

func TestFullscreenClearRestoresAccumulatedStaticOutput(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 3}
	value := "B\nC\nD"

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"A"}, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text(value),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 3,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	value = "X\nY\nZ"

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := stdout.last()
	if !strings.Contains(update, clearTerminalEscape) {
		t.Fatalf("expected fullscreen rerender to clear terminal, got %q", update)
	}

	if !strings.Contains(update, "A\nX\nY\nZ") {
		t.Fatalf("expected fullscreen rerender to restore static history and dynamic output, got %q", update)
	}
}

func TestEraseScreenWhereStaticExistsButInteractivePartIsTallerThanViewportLikeUpstreamFixture(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 3}
	value := "B\nC\nD"

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"A"}, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text(value),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 3,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	value = "X\nY\nZ"

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := stdout.last()
	if !strings.Contains(update, clearTerminalEscape) {
		t.Fatalf("expected rerender with static plus oversized interactive output to clear terminal, got %q", update)
	}

	if !strings.Contains(update, "A\nX\nY\nZ") {
		t.Fatalf("expected rerender to restore static history and updated interactive output, got %q", update)
	}
}

func TestFullscreenInitialRenderDoesNotAddTrailingNewline(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 5}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{
			"height":        5.0,
			"flexDirection": "column",
		},
			components.Box(vdom.Props{"flexGrow": 1.0},
				components.Text("Full-screen: top"),
			),
			components.Text("Bottom line (should be usable)"),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 5,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) != 2 {
		t.Fatalf("expected initial fullscreen render to produce 2 writes, got %#v", stdout.writes)
	}

	output := stdout.writes[1]
	if strings.HasSuffix(output, "\n") {
		t.Fatalf("expected fullscreen render to avoid trailing newline, got %q", output)
	}

	lines := strings.Split(output, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected fullscreen render to occupy exactly 5 lines, got %d in %q", len(lines), output)
	}

	if !strings.Contains(lines[4], "Bottom line (should be usable)") {
		t.Fatalf("expected bottom line to remain usable, got %q", output)
	}
}

func TestStandardStateChangeToEmptyErasesPreviousLines(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 4}
	show := true

	render := func() *vdom.Node {
		if !show {
			return components.Box(vdom.Props{"flexDirection": "column"})
		}

		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("B"),
			components.Text("C"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 4,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	show = false

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 1 {
		t.Fatalf("expected empty rerender to emit a single erase payload, got %#v", stdout.writes)
	}

	update := stdout.last()
	if update != ansiEraseLines(4)+"\n" {
		t.Fatalf("expected empty rerender to emit erase payload plus normalized blank line, got %q", update)
	}
}

func TestFullscreenStateChangeToEmptyUsesClearTerminal(t *testing.T) {
	stdout := &ttySizedRecordingWriter{rows: 3}
	show := true

	render := func() *vdom.Node {
		if !show {
			return components.Box(vdom.Props{"flexDirection": "column"})
		}

		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("A"),
			components.Text("B"),
			components.Text("C"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Height: 3,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	show = false

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 1 {
		t.Fatalf("expected fullscreen empty rerender to emit a single clear-terminal payload, got %#v", stdout.writes)
	}

	update := stdout.last()
	if update != clearTerminalEscape {
		t.Fatalf("expected fullscreen empty rerender to emit only clear-terminal payload, got %q", update)
	}
}

func TestMountWaitUntilExitReturnsExitError(t *testing.T) {
	exitErr := errors.New("exit now")
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		UseApp().Exit(exitErr)
		return vdom.CreateTextNode("bye")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if err := instance.WaitUntilExit(); !errors.Is(err, exitErr) {
		t.Fatalf("expected waitUntilExit to return %v, got %v", exitErr, err)
	}

	if strings.Contains(stdout.joined(), showCursorEscape) {
		t.Fatalf("expected non-TTY unmount not to restore cursor, got %#v", stdout.writes)
	}
}

func TestUseAppExitFromAsyncEffectUnmountsSession(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		app := UseApp()
		UseEffect(func() func() {
			timer := time.AfterFunc(10*time.Millisecond, func() {
				app.Exit()
			})
			return func() {
				timer.Stop()
			}
		}, []interface{}{"async-exit"})

		return vdom.CreateTextNode("bye")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- instance.WaitUntilExit()
	}()

	select {
	case err := <-waitDone:
		if err != nil {
			t.Fatalf("expected nil exit error, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected async useApp exit to unmount the session")
	}
}

func TestRuntimeRenderAppendsNewlineEvenWhenOutputEndsWithBlankRow(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return Box(vdom.Props{"flexDirection": "column"},
			Text("Use arrow keys to move the face. Press “q” to exit."),
			Box(vdom.Props{
				"height":      12.0,
				"paddingLeft": 20.0,
				"paddingTop":  10.0,
			}, Text("^_^")),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Width:  100,
			Height: 30,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if got := stdout.last(); !strings.HasSuffix(got, "                    ^_^\n\n") {
		t.Fatalf("expected non-fullscreen render to append a newline after the output blank row, got %q", got)
	}
}

func TestUnmountFlushesBufferedStdoutBeforeExitCompletes(t *testing.T) {
	stdout := &bufferedFlushWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		t.Fatalf("expected nil exit error, got %v", err)
	}

	if stdout.flushCount == 0 {
		t.Fatal("expected unmount to flush buffered stdout")
	}

	if !strings.Contains(stdout.joined(), "Hello") {
		t.Fatalf("expected flushed output to contain the final frame, got %#v", stdout.flushes)
	}
}

func TestUnmountReturnsFlushErrorFromBufferedStdout(t *testing.T) {
	flushErr := errors.New("flush failed")
	stdout := &bufferedFlushWriter{flushErr: flushErr}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if err := instance.Unmount(); !errors.Is(err, flushErr) {
		t.Fatalf("expected unmount to return %v, got %v", flushErr, err)
	}

	if err := instance.WaitUntilExit(); !errors.Is(err, flushErr) {
		t.Fatalf("expected waitUntilExit to return %v, got %v", flushErr, err)
	}
}

func TestUnmountFlushesBufferedStderrBeforeExitCompletes(t *testing.T) {
	stdout := &recordingWriter{}
	stderr := &bufferedFlushWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		stderrCtx := UseStderr()
		UseEffect(func() func() {
			if _, err := stderrCtx.Write("stderr\n"); err != nil {
				t.Fatalf("unexpected stderr write error: %v", err)
			}

			return nil
		}, []interface{}{"stderr-flush"})

		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		t.Fatalf("expected nil exit error, got %v", err)
	}

	if stderr.flushCount == 0 {
		t.Fatal("expected unmount to flush buffered stderr")
	}

	if !strings.Contains(stderr.joined(), "stderr\n") {
		t.Fatalf("expected flushed stderr to contain hook output, got %#v", stderr.flushes)
	}
}

func TestUnmountReturnsFlushErrorFromBufferedStderr(t *testing.T) {
	flushErr := errors.New("stderr flush failed")
	stdout := &recordingWriter{}
	stderr := &bufferedFlushWriter{flushErr: flushErr}

	instance, err := MountWithOptions(func() *vdom.Node {
		stderrCtx := UseStderr()
		UseEffect(func() func() {
			if _, err := stderrCtx.Write("stderr\n"); err != nil {
				t.Fatalf("unexpected stderr write error: %v", err)
			}

			return nil
		}, []interface{}{"stderr-flush-error"})

		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if err := instance.Unmount(); !errors.Is(err, flushErr) {
		t.Fatalf("expected unmount to return %v, got %v", flushErr, err)
	}

	if err := instance.WaitUntilExit(); !errors.Is(err, flushErr) {
		t.Fatalf("expected waitUntilExit to return %v, got %v", flushErr, err)
	}
}

func TestWaitUntilExitWaitsForAsyncStdoutBarrier(t *testing.T) {
	stdout := newBarrierWaitWriter()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- instance.WaitUntilExit()
	}()

	unmountDone := make(chan error, 1)
	go func() {
		unmountDone <- instance.Unmount()
	}()

	select {
	case <-stdout.waitStarted:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected unmount to start waiting for stdout barrier")
	}

	select {
	case err := <-waitDone:
		t.Fatalf("expected waitUntilExit to block on stdout barrier, got %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	stdout.Release()

	select {
	case err := <-unmountDone:
		if err != nil {
			t.Fatalf("expected nil unmount error, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected unmount to complete after stdout barrier release")
	}

	select {
	case err := <-waitDone:
		if err != nil {
			t.Fatalf("expected nil waitUntilExit error, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected waitUntilExit to resolve after stdout barrier release")
	}

	if stdout.waitCalls == 0 {
		t.Fatal("expected unmount to wait for stdout barrier")
	}
}

func TestWaitUntilExitReturnsAsyncStdoutBarrierError(t *testing.T) {
	waitErr := errors.New("async stdout wait failed")
	stdout := newBarrierWaitWriter()
	stdout.waitErr = waitErr

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- instance.WaitUntilExit()
	}()

	unmountDone := make(chan error, 1)
	go func() {
		unmountDone <- instance.Unmount()
	}()

	select {
	case <-stdout.waitStarted:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected unmount to start waiting for stdout barrier")
	}

	stdout.Release()

	select {
	case err := <-unmountDone:
		if !errors.Is(err, waitErr) {
			t.Fatalf("expected unmount to return %v, got %v", waitErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected unmount to complete after stdout barrier release")
	}

	select {
	case err := <-waitDone:
		if !errors.Is(err, waitErr) {
			t.Fatalf("expected waitUntilExit to return %v, got %v", waitErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected waitUntilExit to resolve after stdout barrier release")
	}
}

func TestWaitUntilExitRequiresExplicitUnmountInGoRuntime(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("complete")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- instance.WaitUntilExit()
	}()

	select {
	case err := <-waitDone:
		t.Fatalf("expected waitUntilExit to wait for explicit unmount, got %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	select {
	case err := <-waitDone:
		if err != nil {
			t.Fatalf("expected nil waitUntilExit error, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected waitUntilExit to resolve after explicit unmount")
	}
}

func TestMountUseStdoutWritePreservesRenderOutput(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		if _, err := UseStdout().Write("outside\n"); err != nil {
			t.Fatalf("unexpected stdout write error: %v", err)
		}

		return vdom.CreateTextNode("inside")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(stdout.writes))
	}

	if stdout.writes[0] != "outside\n" {
		t.Fatalf("expected direct stdout write first, got %q", stdout.writes[0])
	}

	if stdout.writes[1] != "inside\n" {
		t.Fatalf("expected render output last, got %q", stdout.writes[1])
	}
}

func TestMountedUseStdoutWriteRestoresManagedOutputAndCursorAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			phase := 0

			render := func() *vdom.Node {
				currentPhase := phase
				stdoutCtx := UseStdout()
				UseCursor().SetCursorPosition(&CursorPosition{X: 2, Y: 0})
				UseEffect(func() func() {
					if currentPhase == 1 {
						if _, err := stdoutCtx.Write("from stdout hook\n"); err != nil {
							t.Fatalf("unexpected stdout hook write error: %v", err)
						}
					}

					return nil
				}, []interface{}{currentPhase})

				return vdom.CreateTextNode("Hello")
			}

			instance, err := MountWithOptions(render, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil
			phase = 1

			if err := instance.Rerender(render); err != nil {
				t.Fatalf("rerender failed: %v", err)
			}

			if len(stdout.writes) != 3 {
				t.Fatalf("expected clear, hook write, and restore on stdout; got %#v", stdout.writes)
			}

			if stdout.writes[1] != "from stdout hook\n" {
				t.Fatalf("expected hook payload in the middle, got %#v", stdout.writes)
			}

			if !strings.Contains(stdout.last(), "Hello") {
				t.Fatalf("expected managed output to be restored, got %q", stdout.last())
			}

			if !strings.Contains(stdout.last(), showCursorEscape) {
				t.Fatalf("expected restored output to show cursor again, got %q", stdout.last())
			}

			joined := stdout.joined()
			if strings.LastIndex(joined, showCursorEscape) < strings.LastIndex(joined, hideCursorEscape) {
				t.Fatalf("expected cursor to end visible after stdout hook write, got %#v", stdout.writes)
			}
		})
	}
}

func TestMountedUseStderrWriteRestoresManagedOutputAndCursorAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			stderr := &recordingWriter{}
			phase := 0

			render := func() *vdom.Node {
				currentPhase := phase
				stderrCtx := UseStderr()
				UseCursor().SetCursorPosition(&CursorPosition{X: 2, Y: 0})
				UseEffect(func() func() {
					if currentPhase == 1 {
						if _, err := stderrCtx.Write("from stderr hook\n"); err != nil {
							t.Fatalf("unexpected stderr hook write error: %v", err)
						}
					}

					return nil
				}, []interface{}{currentPhase})

				return vdom.CreateTextNode("Hello")
			}

			instance, err := MountWithOptions(render, RenderOptions{
				AppOptions: AppOptions{
					Stdout: stdout,
					Stderr: stderr,
				},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil
			stderr.writes = nil
			phase = 1

			if err := instance.Rerender(render); err != nil {
				t.Fatalf("rerender failed: %v", err)
			}

			if len(stderr.writes) != 1 || stderr.writes[0] != "from stderr hook\n" {
				t.Fatalf("expected hook payload on stderr, got %#v", stderr.writes)
			}

			if len(stdout.writes) != 2 {
				t.Fatalf("expected clear and restore on stdout, got %#v", stdout.writes)
			}

			if !strings.Contains(stdout.last(), "Hello") {
				t.Fatalf("expected managed output to be restored on stdout, got %q", stdout.last())
			}

			if !strings.Contains(stdout.last(), showCursorEscape) {
				t.Fatalf("expected restored stderr flow to show cursor again, got %q", stdout.last())
			}

			joined := stdout.joined()
			if strings.LastIndex(joined, showCursorEscape) < strings.LastIndex(joined, hideCursorEscape) {
				t.Fatalf("expected cursor to end visible after stderr hook write, got %#v", stdout.writes)
			}
		})
	}
}

func TestCIModeUnmountWritesOnlyFinalFrame(t *testing.T) {
	t.Setenv("CI", "true")

	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("first")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected CI mode to defer dynamic output until unmount, got %#v", stdout.writes)
	}

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("second")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected CI rerender to keep deferring output, got %#v", stdout.writes)
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if len(stdout.writes) != 1 || stdout.writes[0] != "second\n" {
		t.Fatalf("expected CI unmount to emit only the last frame, got %#v", stdout.writes)
	}
}

func TestCIModeWriteStdoutDoesNotRestoreManagedOutput(t *testing.T) {
	t.Setenv("CI", "true")

	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("inside")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil

	if _, err := instance.WriteStdout("outside\n"); err != nil {
		t.Fatalf("stdout write failed: %v", err)
	}

	if len(stdout.writes) != 1 || stdout.writes[0] != "outside\n" {
		t.Fatalf("expected CI stdout write to bypass clear-and-restore, got %#v", stdout.writes)
	}
}

func TestCIModeEmptyUnmountDoesNotWriteBlankFinalFrame(t *testing.T) {
	t.Setenv("CI", "true")

	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return nil
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected CI empty mount to avoid writes, got %#v", stdout.writes)
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		t.Fatalf("expected nil exit error, got %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected CI empty unmount to avoid blank final frame, got %#v", stdout.writes)
	}
}

func TestCIModeStaticOutputStreamsBeforeFinalDynamicFrame(t *testing.T) {
	t.Setenv("CI", "true")

	stdout := &ttyRecordingWriter{}
	items := []string{"A"}

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems(items, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("X"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if len(stdout.writes) != 1 || stdout.writes[0] != "A\n" {
		t.Fatalf("expected CI mode to stream static output immediately, got %#v", stdout.writes)
	}

	stdout.writes = nil
	items = []string{"A", "B"}

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 1 || stdout.writes[0] != "B\n" {
		t.Fatalf("expected CI rerender to stream only appended static output, got %#v", stdout.writes)
	}

	stdout.writes = nil

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if len(stdout.writes) != 1 || stdout.writes[0] != "X\n" {
		t.Fatalf("expected CI unmount to emit the final dynamic frame after static writes, got %#v", stdout.writes)
	}
}

func TestResetPropWhenRemovedFromElementLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}

	render := func(remove bool) func() *vdom.Node {
		return func() *vdom.Node {
			props := vdom.Props{
				"flexDirection":  "column",
				"justifyContent": "flex-end",
			}
			if !remove {
				props["height"] = float64(4)
			}

			return components.Box(props,
				components.Text("x"),
			)
		}
	}

	instance, err := MountWithOptions(render(false), RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.last() != "\n\n\nx" {
		t.Fatalf("expected initial fixed-height render to bottom-align text, got %q", stdout.last())
	}

	if err := instance.Rerender(render(true)); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if stdout.last() != "x" {
		t.Fatalf("expected rerender without height prop to reset layout, got %q", stdout.last())
	}
}

func TestRerenderRemeasuresTextChangeLikeUpstreamFixture(t *testing.T) {
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

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.last() != "abc" {
		t.Fatalf("expected initial text-change output %q, got %#v", "abc", stdout.writes)
	}

	add = true
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if stdout.last() != "abcx" {
		t.Fatalf("expected rerendered text-change output %q, got %#v", "abcx", stdout.writes)
	}
}

func TestRerenderRemeasuresNestedTextNodeChangeLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}
	add := false

	render := func() *vdom.Node {
		children := []any{"abc"}
		if add {
			children = append(children, components.Text("x"))
		}

		return components.Box(nil,
			components.Text(children...),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.last() != "abc" {
		t.Fatalf("expected initial nested text-node output %q, got %#v", "abc", stdout.writes)
	}

	add = true
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if stdout.last() != "abcx" {
		t.Fatalf("expected rerendered nested text-node output %q, got %#v", "abcx", stdout.writes)
	}
}

func TestCursorOnlyUpdateDoesNotRewriteOutput(t *testing.T) {
	stdout := &recordingWriter{}
	cursorX := 0

	instance, err := MountWithOptions(func() *vdom.Node {
		UseCursor().SetCursorPosition(&CursorPosition{X: cursorX, Y: 0})
		return vdom.CreateTextNode("cursor")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	cursorX = 2
	if err := instance.Rerender(func() *vdom.Node {
		UseCursor().SetCursorPosition(&CursorPosition{X: cursorX, Y: 0})
		return vdom.CreateTextNode("cursor")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(stdout.writes))
	}

	if strings.Contains(stdout.writes[1], "cursor") {
		t.Fatalf("expected cursor-only update without rewriting output, got %q", stdout.writes[1])
	}

	if !strings.Contains(stdout.writes[1], showCursorEscape) {
		t.Fatalf("expected cursor-only update to show cursor, got %q", stdout.writes[1])
	}
}

func TestCursorRemainsVisibleAfterFirstRenderEffects(t *testing.T) {
	stdout := &recordingWriter{}
	effectRan := false

	instance, err := MountWithOptions(func() *vdom.Node {
		UseCursor().SetCursorPosition(&CursorPosition{X: 2, Y: 0})

		UseEffect(func() func() {
			effectRan = true
			return nil
		}, []interface{}{"cursor-effect"})

		return components.Box(nil,
			components.Text("> "),
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if !effectRan {
		t.Fatal("expected first-render effect to run")
	}

	lastVisibilityChange := ""
	for _, write := range stdout.writes {
		switch {
		case strings.Contains(write, showCursorEscape):
			lastVisibilityChange = showCursorEscape
		case write == hideCursorEscape:
			lastVisibilityChange = hideCursorEscape
		}
	}

	if lastVisibilityChange != showCursorEscape {
		t.Fatalf("expected last cursor visibility change to show cursor, got %#v", stdout.writes)
	}
}

func TestCursorTracksTypedInputLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}
	stdin := &inputRecordingStdin{}

	instance, err := MountWithOptions(func() *vdom.Node {
		text, setText := UseState("")

		UseInput(func(input string, key InputKey) {
			current := text.(string)
			if key.Backspace || key.Delete {
				if current != "" {
					setText(current[:len(current)-1])
				}
				return
			}

			if !key.Ctrl && !key.Meta && input != "" {
				setText(current + input)
			}
		})

		current := text.(string)
		UseCursor().SetCursorPosition(&CursorPosition{X: 2 + len(current), Y: 0})

		return components.Box(nil,
			components.Text(fmt.Sprintf("> %s", current)),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if !strings.Contains(stdout.last(), showCursorEscape) {
		t.Fatalf("expected initial render to show cursor, got %q", stdout.last())
	}

	if !strings.Contains(stdout.last(), ansiCursorTo(2)) {
		t.Fatalf("expected initial render to place cursor after prompt, got %q", stdout.last())
	}

	writesAfterMount := len(stdout.writes)
	stdin.EmitInput("a")

	if len(stdout.writes) <= writesAfterMount {
		t.Fatalf("expected input rerender after typing, got %#v", stdout.writes)
	}

	if !strings.Contains(stdout.last(), showCursorEscape) {
		t.Fatalf("expected typed input rerender to end with a visible cursor, got %q", stdout.last())
	}

	if !strings.Contains(stdout.last(), ansiCursorTo(3)) {
		t.Fatalf("expected cursor to advance after typing \"a\", got %q", stdout.last())
	}

	writesAfterLetter := len(stdout.writes)
	stdin.EmitInput(" ")

	if len(stdout.writes) <= writesAfterLetter {
		t.Fatalf("expected trailing-space input to trigger a rerender, got %#v", stdout.writes)
	}

	if !strings.Contains(stdout.last(), showCursorEscape) {
		t.Fatalf("expected trailing-space rerender to keep cursor visible, got %q", stdout.last())
	}

	if !strings.Contains(stdout.last(), ansiCursorTo(4)) {
		t.Fatalf("expected cursor to advance after typing a trailing space, got %q", stdout.last())
	}
}

func TestSubsequentCursorRendersReturnToBottomBeforeErasing(t *testing.T) {
	stdout := &recordingWriter{}
	stdin := &inputRecordingStdin{}

	instance, err := MountWithOptions(func() *vdom.Node {
		text, setText := UseState("")

		UseInput(func(input string, key InputKey) {
			current := text.(string)
			if !key.Ctrl && !key.Meta && input != "" {
				setText(current + input)
			}
		})

		current := text.(string)
		UseCursor().SetCursorPosition(&CursorPosition{X: 2 + len(current), Y: 1})

		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("Header"),
			components.Text(fmt.Sprintf("> %s", current)),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	writesAfterMount := len(stdout.writes)
	stdin.EmitInput("x")

	if len(stdout.writes) <= writesAfterMount {
		t.Fatalf("expected rerender after typed input, got %#v", stdout.writes)
	}

	update := stdout.last()
	expectedPrefix := hideCursorEscape + ansiCursorDown(1) + ansiCursorTo(0) + ansiEraseLines(3)

	if !strings.HasPrefix(update, expectedPrefix) {
		t.Fatalf("expected rerender to return to the bottom before erasing, got %q", update)
	}

	if !strings.Contains(update, "Header\n> x\n") {
		t.Fatalf("expected rerender to include updated multiline output, got %q", update)
	}
}

func TestCursorIsClearedWhenUseCursorComponentUnmountsAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			showChild := true

			render := func() *vdom.Node {
				if showChild {
					return components.Box(nil,
						func() *vdom.Node {
							UseCursor().SetCursorPosition(&CursorPosition{X: 5, Y: 0})
							return components.Text("child")
						}(),
					)
				}

				return components.Text("no cursor")
			}

			instance, err := MountWithOptions(render, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil
			showChild = false

			if err := instance.Rerender(render); err != nil {
				t.Fatalf("rerender failed: %v", err)
			}

			update := stdout.last()
			if strings.Contains(update, showCursorEscape) {
				t.Fatalf("expected rerender without useCursor child to keep cursor hidden, got %q", update)
			}
			if !strings.Contains(update, "no cursor") {
				t.Fatalf("expected rerender to write updated content, got %q", update)
			}
		})
	}
}

func TestUnmountRunsEffectCleanup(t *testing.T) {
	stdout := &recordingWriter{}
	cleanupCalled := false

	instance, err := MountWithOptions(func() *vdom.Node {
		UseEffect(func() func() {
			return func() {
				cleanupCalled = true
			}
		}, []interface{}{"cleanup"})

		return vdom.CreateTextNode("cleanup")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if !cleanupCalled {
		t.Fatal("expected effect cleanup to run on unmount")
	}

	if strings.Contains(stdout.joined(), showCursorEscape) {
		t.Fatalf("expected non-TTY unmount not to restore cursor, got %#v", stdout.writes)
	}
}

func TestExitFullyDisablesNestedManualRawMode(t *testing.T) {
	stdout := &recordingWriter{}
	stdin := &rawModeRecordingStdin{}
	rawModeUsersBeforeExit := 0

	instance, err := MountWithOptions(func() *vdom.Node {
		stdinCtx := UseStdin()
		app := UseApp()

		UseEffect(func() func() {
			if err := stdinCtx.SetRawMode(true); err != nil {
				t.Fatalf("enable raw mode failed: %v", err)
			}
			if err := stdinCtx.SetRawMode(true); err != nil {
				t.Fatalf("nested raw mode failed: %v", err)
			}

			rawModeUsersBeforeExit = stdinCtx.app.rawModeUsers
			app.Exit()
			return nil
		}, []interface{}{"raw-mode-exit"})

		return vdom.CreateTextNode("bye")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if rawModeUsersBeforeExit != 2 {
		t.Fatalf("expected nested raw mode users before exit, got %d", rawModeUsersBeforeExit)
	}

	if err := instance.WaitUntilExit(); err != nil {
		t.Fatalf("expected nil exit error, got %v", err)
	}

	if instance.app.rawModeUsers != 0 {
		t.Fatalf("expected exit to fully clear raw mode users, got %d", instance.app.rawModeUsers)
	}

	if instance.app.rawState != nil {
		t.Fatalf("expected exit to restore raw terminal state, got %#v", instance.app.rawState)
	}
}

func TestExitWithErrorFullyDisablesRawModeLikeUpstreamFixture(t *testing.T) {
	exitErr := errors.New("raw mode exit failed")
	stdout := &recordingWriter{}
	stdin := &rawModeRecordingStdin{}
	rawModeUsersBeforeExit := 0

	instance, err := MountWithOptions(func() *vdom.Node {
		stdinCtx := UseStdin()
		app := UseApp()

		UseEffect(func() func() {
			if err := stdinCtx.SetRawMode(true); err != nil {
				t.Fatalf("enable raw mode failed: %v", err)
			}

			rawModeUsersBeforeExit = stdinCtx.app.rawModeUsers
			app.Exit(exitErr)
			return nil
		}, []interface{}{"raw-mode-exit-error"})

		return vdom.CreateTextNode("bye")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if rawModeUsersBeforeExit != 1 {
		t.Fatalf("expected mounted app to hold one raw mode user before exit, got %d", rawModeUsersBeforeExit)
	}

	if err := instance.WaitUntilExit(); !errors.Is(err, exitErr) {
		t.Fatalf("expected waitUntilExit to return %v, got %v", exitErr, err)
	}

	if instance.app.rawModeUsers != 0 {
		t.Fatalf("expected exit with error to fully clear raw mode users, got %d", instance.app.rawModeUsers)
	}

	if instance.app.rawState != nil {
		t.Fatalf("expected exit with error to restore raw terminal state, got %#v", instance.app.rawState)
	}
}

func TestUnmountDisablesRawModeLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}
	stdin := &rawModeRecordingStdin{}

	instance, err := MountWithOptions(func() *vdom.Node {
		stdinCtx := UseStdin()

		UseEffect(func() func() {
			if err := stdinCtx.SetRawMode(true); err != nil {
				t.Fatalf("enable raw mode failed: %v", err)
			}

			return nil
		}, []interface{}{"raw-mode-unmount"})

		return vdom.CreateTextNode("Hello World")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if instance.app.rawModeUsers != 1 {
		t.Fatalf("expected mounted app to hold one raw mode user, got %d", instance.app.rawModeUsers)
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		t.Fatalf("expected nil exit error after raw-mode unmount, got %v", err)
	}

	if instance.app.rawModeUsers != 0 {
		t.Fatalf("expected unmount to fully clear raw mode users, got %d", instance.app.rawModeUsers)
	}

	if instance.app.rawState != nil {
		t.Fatalf("expected unmount to restore raw terminal state, got %#v", instance.app.rawState)
	}
}

func TestDebugRerenderWritesSeparateOutputs(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("abc")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("abcx")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(stdout.writes))
	}

	if stdout.writes[0] != "abc" || stdout.writes[1] != "abcx" {
		t.Fatalf("expected separate debug outputs, got %#v", stdout.writes)
	}
}

func TestIncrementalRerenderUsesSurgicalUpdate(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\nLine 2\nLine 3\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\nUpdated\nLine 3\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected 3 writes, got %d", len(stdout.writes))
	}

	update := stdout.writes[2]
	if !strings.Contains(update, ansiCursorNextLine()) {
		t.Fatalf("expected incremental update to skip unchanged lines, got %q", update)
	}
	if !strings.Contains(update, "Updated") {
		t.Fatalf("expected incremental update to write changed line, got %q", update)
	}
	if strings.Contains(update, "Line 1") || strings.Contains(update, "Line 3") {
		t.Fatalf("expected incremental update to avoid rewriting unchanged lines, got %q", update)
	}
}

func TestIncrementalRerenderShrinksOutput(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\nLine 2\nLine 3\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected 3 writes, got %d", len(stdout.writes))
	}

	update := stdout.writes[2]
	if !strings.Contains(update, ansiEraseLines(2)) {
		t.Fatalf("expected shrink update to erase extra lines, got %q", update)
	}
}

// TestIncrementalRerenderUsesColumnLevelDirtyDiff verifies the per-line
// column-level dirty-rect optimization: when only the trailing portion of
// a line changes, the writer emits cursor positioning to the divergence
// column plus the differing tail, instead of rewriting the whole line.
func TestIncrementalRerenderUsesColumnLevelDirtyDiff(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Counter: 0\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Counter: 1\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected 3 writes, got %d", len(stdout.writes))
	}

	update := stdout.writes[2]
	// "Counter: " is the 9-column shared prefix; the writer should jump
	// straight to column 9, erase the rest of the line, and emit "1".
	if !strings.Contains(update, ansiCursorTo(9)) {
		t.Fatalf("expected column-level diff to position cursor at column 9, got %q", update)
	}
	if !strings.Contains(update, "1") {
		t.Fatalf("expected column-level diff to write the changed tail, got %q", update)
	}
	if strings.Contains(update, "Counter: 1") {
		t.Fatalf("expected column-level diff to skip rewriting the shared prefix, got %q", update)
	}
}

// TestIncrementalRerenderColumnDiffFallsBackForANSIEscapes verifies the
// optimization is gated to plain text. When either line contains ANSI
// escape sequences we must not naively count columns — the writer falls
// back to the cursorTo(0) + full-line rewrite path.
func TestIncrementalRerenderColumnDiffFallsBackForANSIEscapes(t *testing.T) {
	stdout := &ttyRecordingWriter{}

	first := "\x1b[31mhello\x1b[0m\n"
	second := "\x1b[32mhello\x1b[0m\n"

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode(first)
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode(second)
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	last := stdout.writes[len(stdout.writes)-1]
	// ANSI-bearing line — full-line rewrite must reposition to column 0.
	if !strings.Contains(last, ansiCursorTo(0)) {
		t.Fatalf("expected ANSI line to fall back to cursorTo(0) full-line rewrite, got %q", last)
	}
	if !strings.Contains(last, "\x1b[32m") {
		t.Fatalf("expected new ANSI sequence to be emitted in full-line rewrite, got %q", last)
	}
}

func TestIncrementalClearResetsState(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\nLine 2\nLine 3\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Clear(); err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 4 {
		t.Fatalf("expected 4 writes, got %d", len(stdout.writes))
	}

	if stdout.writes[3] != "Line 1\n\n" {
		t.Fatalf("expected fresh output after clear, got %q", stdout.writes[3])
	}
}

func TestIncrementalMultipleClearCallsAreHarmless(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\nLine 2\nLine 3\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Clear(); err != nil {
		t.Fatalf("first clear failed: %v", err)
	}
	if err := instance.Clear(); err != nil {
		t.Fatalf("second clear failed: %v", err)
	}
	if err := instance.Clear(); err != nil {
		t.Fatalf("third clear failed: %v", err)
	}

	if len(stdout.writes) != 5 {
		t.Fatalf("expected 5 writes after repeated clears, got %d", len(stdout.writes))
	}

	if stdout.writes[3] != "" || stdout.writes[4] != "" {
		t.Fatalf("expected empty no-op writes after state reset, got %#v", stdout.writes[3:])
	}
}

func TestIncrementalTrailingToNoTrailingTransition(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("A\nB\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("A\nB")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected trailing blank-row transition to clear the extra row, got %#v", stdout.writes)
	}
	if !containsANSIEscape(stdout.writes[2]) {
		t.Fatalf("expected trailing blank-row transition to use cursor/erase escapes, got %q", stdout.writes[2])
	}
}

func TestIncrementalNoTrailingGrowAndStableLastLine(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("A")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("A\nB\nC")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := stdout.last()
	if !strings.Contains(update, "B") || !strings.Contains(update, "C") {
		t.Fatalf("expected grow update to include new lines, got %q", update)
	}
	if !strings.HasSuffix(update, "\n") {
		t.Fatalf("expected grow payload to preserve normalized trailing newline, got %q", update)
	}
}

func TestIncrementalNoTrailingUnchangedLastLineDoesNotOvershoot(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("A\nB")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("X\nB")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := stdout.last()
	if !strings.Contains(update, "X") {
		t.Fatalf("expected changed first line to be written, got %q", update)
	}
	if !strings.Contains(update, ansiCursorNextLine()) {
		t.Fatalf("expected normalized trailing-empty line to allow cursor-next-line, got %q", update)
	}
}

func TestIncrementalRenderToEmptyLineSkipsSecondIdenticalRender(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\nLine 2\nLine 3\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("\n")
	}); err != nil {
		t.Fatalf("first rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected 3 writes after first empty-line render, got %d", len(stdout.writes))
	}

	firstEmpty := stdout.last()
	if strings.Count(firstEmpty, "\n") != 2 {
		t.Fatalf("expected empty-line render to preserve output newline plus render newline, got %q", firstEmpty)
	}
	if !strings.Contains(firstEmpty, ansiEraseLines(3)) {
		t.Fatalf("expected empty-line render to erase previous output, got %q", firstEmpty)
	}

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("\n")
	}); err != nil {
		t.Fatalf("second rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected identical empty-line rerender to be skipped, got %d writes", len(stdout.writes))
	}
}

func TestSyncWithoutCursorDoesNotWriteToStream(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(nil, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil

	if err := instance.Sync("Line 1\nLine 2\nLine 3\n", nil); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected sync without cursor to avoid writes, got %#v", stdout.writes)
	}
}

func TestSyncWritesCursorSuffixAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}

			instance, err := MountWithOptions(nil, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil

			cursor := &CursorPosition{X: 5, Y: 1}
			if err := instance.Sync("Line 1\nLine 2\nLine 3\n", cursor); err != nil {
				t.Fatalf("sync failed: %v", err)
			}

			if len(stdout.writes) != 1 {
				t.Fatalf("expected one sync write, got %#v", stdout.writes)
			}

			expected := buildCursorSuffix(visibleLineCount("Line 1\nLine 2\nLine 3\n"), cursor)
			if stdout.writes[0] != expected {
				t.Fatalf("expected %q, got %q", expected, stdout.writes[0])
			}
		})
	}
}

func TestSyncWithCursorMakesNextRenderHideCursorAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}

			instance, err := MountWithOptions(nil, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.Sync("Line 1\nLine 2\nLine 3\n", &CursorPosition{X: 5, Y: 1}); err != nil {
				t.Fatalf("sync failed: %v", err)
			}

			stdout.writes = nil

			if err := instance.Rerender(func() *vdom.Node {
				return vdom.CreateTextNode("Updated\n")
			}); err != nil {
				t.Fatalf("rerender failed: %v", err)
			}

			if len(stdout.writes) == 0 {
				t.Fatal("expected rerender output after sync")
			}

			if !strings.HasPrefix(stdout.last(), hideCursorEscape) {
				t.Fatalf("expected rerender after cursor sync to begin by hiding cursor, got %q", stdout.last())
			}
		})
	}
}

func TestSyncHidesCursorWhenPreviousRenderShowedCursorAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}

			instance, err := MountWithOptions(func() *vdom.Node {
				UseCursor().SetCursorPosition(&CursorPosition{X: 5, Y: 1})
				return vdom.CreateTextNode("Line 1\nLine 2\nLine 3\n")
			}, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil

			if err := instance.Sync("Fresh output\n", nil); err != nil {
				t.Fatalf("sync failed: %v", err)
			}

			if len(stdout.writes) != 1 {
				t.Fatalf("expected one sync write, got %#v", stdout.writes)
			}

			if stdout.writes[0] != hideCursorEscape {
				t.Fatalf("expected sync to hide cursor, got %q", stdout.writes[0])
			}
		})
	}
}

func TestSyncResetsCursorStateForNextRenderAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}

			instance, err := MountWithOptions(func() *vdom.Node {
				UseCursor().SetCursorPosition(&CursorPosition{X: 5, Y: 0})
				return vdom.CreateTextNode("Line 1\nLine 2\nLine 3\n")
			}, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.Sync("Fresh output\n", nil); err != nil {
				t.Fatalf("sync failed: %v", err)
			}

			stdout.writes = nil

			if err := instance.Rerender(func() *vdom.Node {
				return vdom.CreateTextNode("Updated output\n")
			}); err != nil {
				t.Fatalf("rerender failed: %v", err)
			}

			if len(stdout.writes) == 0 {
				t.Fatal("expected rerender output after sync")
			}

			if strings.Contains(stdout.last(), hideCursorEscape) {
				t.Fatalf("expected sync to reset cursor state before rerender, got %q", stdout.last())
			}
		})
	}
}

func TestWriteStdoutRestoresManagedOutputAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}

			instance, err := MountWithOptions(func() *vdom.Node {
				UseCursor().SetCursorPosition(&CursorPosition{X: 2, Y: 0})
				return vdom.CreateTextNode("Hello")
			}, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil

			if _, err := instance.WriteStdout("outside\n"); err != nil {
				t.Fatalf("stdout write failed: %v", err)
			}

			if len(stdout.writes) != 3 {
				t.Fatalf("expected clear, external write, and restore; got %#v", stdout.writes)
			}

			if stdout.writes[1] != "outside\n" {
				t.Fatalf("expected external stdout payload in the middle, got %q", stdout.writes[1])
			}

			if !strings.Contains(stdout.last(), "Hello") {
				t.Fatalf("expected managed output to be restored, got %q", stdout.last())
			}

			if !strings.Contains(stdout.last(), showCursorEscape) {
				t.Fatalf("expected restored output to show cursor again, got %q", stdout.last())
			}

			if strings.LastIndex(stdout.joined(), showCursorEscape) < strings.LastIndex(stdout.joined(), hideCursorEscape) {
				t.Fatalf("expected cursor to end visible after stdout write, got %#v", stdout.writes)
			}
		})
	}
}

func TestWriteStdoutRestoresCursorWithoutOutputAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			cursor := &CursorPosition{X: 2, Y: 0}

			instance, err := MountWithOptions(func() *vdom.Node {
				UseCursor().SetCursorPosition(cursor)
				return nil
			}, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil

			if _, err := instance.WriteStdout("outside\n"); err != nil {
				t.Fatalf("stdout write failed: %v", err)
			}

			if len(stdout.writes) != 3 {
				t.Fatalf("expected clear, external write, and cursor restore; got %#v", stdout.writes)
			}

			if stdout.writes[1] != "outside\n" {
				t.Fatalf("expected external stdout payload in the middle, got %#v", stdout.writes)
			}

			expectedRestore := buildCursorSuffix(visibleLineCount(""), cursor)
			expectedRestore = "\n" + buildCursorSuffix(visibleLineCount("\n"), cursor)
			if stdout.writes[2] != expectedRestore {
				t.Fatalf("expected cursor restore %q, got %#v", expectedRestore, stdout.writes)
			}

			if strings.LastIndex(stdout.joined(), showCursorEscape) < strings.LastIndex(stdout.joined(), hideCursorEscape) {
				t.Fatalf("expected cursor to end visible after stdout write, got %#v", stdout.writes)
			}
		})
	}
}

func TestWriteStdoutWrapsRestoreCycleWithSyncEscapesOnTTYAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &ttyRecordingWriter{}

			instance, err := MountWithOptions(func() *vdom.Node {
				return vdom.CreateTextNode("Hello")
			}, RenderOptions{
				AppOptions:           AppOptions{Stdout: stdout},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil

			if _, err := instance.WriteStdout("outside\n"); err != nil {
				t.Fatalf("stdout write failed: %v", err)
			}

			if len(stdout.writes) != 5 {
				t.Fatalf("expected wrapped clear, external write, restore, and sync escapes; got %#v", stdout.writes)
			}

			if stdout.writes[0] != bsu {
				t.Fatalf("expected stdout write flow to start with bsu, got %#v", stdout.writes)
			}

			if stdout.writes[1] == "" {
				t.Fatalf("expected stdout write flow to clear the previous frame, got %#v", stdout.writes)
			}

			if stdout.writes[2] != "outside\n" {
				t.Fatalf("expected external stdout payload in the wrapped sequence, got %#v", stdout.writes)
			}

			if !strings.Contains(stdout.writes[3], "Hello") {
				t.Fatalf("expected wrapped stdout flow to restore managed output, got %#v", stdout.writes)
			}

			if stdout.writes[4] != esu {
				t.Fatalf("expected stdout write flow to end with esu, got %#v", stdout.writes)
			}
		})
	}
}

func TestWriteStderrRestoresManagedOutputAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			stderr := &recordingWriter{}

			instance, err := MountWithOptions(func() *vdom.Node {
				UseCursor().SetCursorPosition(&CursorPosition{X: 2, Y: 0})
				return vdom.CreateTextNode("Hello")
			}, RenderOptions{
				AppOptions: AppOptions{
					Stdout: stdout,
					Stderr: stderr,
				},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil
			stderr.writes = nil

			if _, err := instance.WriteStderr("outside\n"); err != nil {
				t.Fatalf("stderr write failed: %v", err)
			}

			if len(stderr.writes) != 1 || stderr.writes[0] != "outside\n" {
				t.Fatalf("expected external payload on stderr, got %#v", stderr.writes)
			}

			if len(stdout.writes) != 2 {
				t.Fatalf("expected clear and restore on stdout, got %#v", stdout.writes)
			}

			if !strings.Contains(stdout.last(), "Hello") {
				t.Fatalf("expected managed output to be restored after stderr write, got %q", stdout.last())
			}

			if !strings.Contains(stdout.last(), showCursorEscape) {
				t.Fatalf("expected restored stderr flow to show cursor again, got %q", stdout.last())
			}

			if strings.LastIndex(stdout.joined(), showCursorEscape) < strings.LastIndex(stdout.joined(), hideCursorEscape) {
				t.Fatalf("expected cursor to end visible after stderr write, got %#v", stdout.writes)
			}
		})
	}
}

func TestWriteStderrRestoresCursorWithoutOutputAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			stderr := &recordingWriter{}
			cursor := &CursorPosition{X: 2, Y: 0}

			instance, err := MountWithOptions(func() *vdom.Node {
				UseCursor().SetCursorPosition(cursor)
				return nil
			}, RenderOptions{
				AppOptions: AppOptions{
					Stdout: stdout,
					Stderr: stderr,
				},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil
			stderr.writes = nil

			if _, err := instance.WriteStderr("outside\n"); err != nil {
				t.Fatalf("stderr write failed: %v", err)
			}

			if len(stderr.writes) != 1 || stderr.writes[0] != "outside\n" {
				t.Fatalf("expected external payload on stderr, got %#v", stderr.writes)
			}

			if len(stdout.writes) != 2 {
				t.Fatalf("expected clear and cursor restore on stdout, got %#v", stdout.writes)
			}

			expectedRestore := buildCursorSuffix(visibleLineCount(""), cursor)
			expectedRestore = "\n" + buildCursorSuffix(visibleLineCount("\n"), cursor)
			if stdout.writes[1] != expectedRestore {
				t.Fatalf("expected cursor restore %q, got %#v", expectedRestore, stdout.writes)
			}

			if strings.LastIndex(stdout.joined(), showCursorEscape) < strings.LastIndex(stdout.joined(), hideCursorEscape) {
				t.Fatalf("expected cursor to end visible after stderr write, got %#v", stdout.writes)
			}
		})
	}
}

func TestWriteStderrWrapsRestoreCycleWithSyncEscapesOnTTYAcrossModes(t *testing.T) {
	modes := []struct {
		name        string
		incremental bool
	}{
		{name: "standard"},
		{name: "incremental", incremental: true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			stdout := &ttyRecordingWriter{}
			stderr := &recordingWriter{}

			instance, err := MountWithOptions(func() *vdom.Node {
				return vdom.CreateTextNode("Hello")
			}, RenderOptions{
				AppOptions: AppOptions{
					Stdout: stdout,
					Stderr: stderr,
				},
				IncrementalRendering: mode.incremental,
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			stdout.writes = nil
			stderr.writes = nil

			if _, err := instance.WriteStderr("outside\n"); err != nil {
				t.Fatalf("stderr write failed: %v", err)
			}

			if len(stdout.writes) != 4 {
				t.Fatalf("expected wrapped clear, restore, and sync escapes on stdout; got %#v", stdout.writes)
			}

			if stdout.writes[0] != bsu {
				t.Fatalf("expected stderr write flow to start with bsu on stdout, got %#v", stdout.writes)
			}

			if stdout.writes[1] == "" {
				t.Fatalf("expected stderr write flow to clear the previous frame, got %#v", stdout.writes)
			}

			if !strings.Contains(stdout.writes[2], "Hello") {
				t.Fatalf("expected stderr write flow to restore managed output on stdout, got %#v", stdout.writes)
			}

			if stdout.writes[3] != esu {
				t.Fatalf("expected stderr write flow to end with esu on stdout, got %#v", stdout.writes)
			}

			if len(stderr.writes) != 1 || stderr.writes[0] != "outside\n" {
				t.Fatalf("expected wrapped flow to write external payload on stderr, got %#v", stderr.writes)
			}
		})
	}
}

func TestDebugWriteStderrReplaysManagedOutput(t *testing.T) {
	stdout := &recordingWriter{}
	stderr := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		Debug: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	stderr.writes = nil

	if _, err := instance.WriteStderr("outside\n"); err != nil {
		t.Fatalf("stderr write failed: %v", err)
	}

	if len(stderr.writes) != 1 || stderr.writes[0] != "outside\n" {
		t.Fatalf("expected external payload on stderr, got %#v", stderr.writes)
	}

	if len(stdout.writes) != 1 || stdout.writes[0] != "Hello" {
		t.Fatalf("expected debug stderr write to replay managed output, got %#v", stdout.writes)
	}
}

func TestDebugWriteStdoutReplaysManagedOutput(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil

	if _, err := instance.WriteStdout("outside\n"); err != nil {
		t.Fatalf("stdout write failed: %v", err)
	}

	if len(stdout.writes) != 2 {
		t.Fatalf("expected debug stdout write plus managed replay, got %#v", stdout.writes)
	}

	if stdout.writes[0] != "outside\n" {
		t.Fatalf("expected external payload on stdout first, got %#v", stdout.writes)
	}

	if stdout.writes[1] != "Hello" {
		t.Fatalf("expected debug stdout write to replay managed output, got %#v", stdout.writes)
	}
}

func TestDebugMountedUseStderrWriteReplaysManagedOutput(t *testing.T) {
	stdout := &recordingWriter{}
	stderr := &recordingWriter{}
	phase := 0

	render := func() *vdom.Node {
		currentPhase := phase
		stderrCtx := UseStderr()

		UseEffect(func() func() {
			if currentPhase == 1 {
				if _, err := stderrCtx.Write("from stderr hook\n"); err != nil {
					t.Fatalf("unexpected stderr hook write error: %v", err)
				}
			}

			return nil
		}, []interface{}{currentPhase})

		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"A"}, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("Live"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		Debug: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	stderr.writes = nil
	phase = 1

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stderr.writes) != 1 || stderr.writes[0] != "from stderr hook\n" {
		t.Fatalf("expected hook payload on stderr, got %#v", stderr.writes)
	}

	// In debug mode upstream Ink writes once per render PLUS replays the
	// managed output after every external stderr/stdout hook write (see
	// ink.tsx writeToStderr). The hook fires during the rerender's effect,
	// producing one replay write; the rerender itself produces a second
	// debug append (debug mode emits one write per render unconditionally).
	if len(stdout.writes) != 2 {
		t.Fatalf("expected debug hook replay plus post-render append, got %#v", stdout.writes)
	}
	for index, write := range stdout.writes {
		if write != "A\nLive" {
			t.Fatalf("expected debug stdout write #%d to replay managed output %q, got %q", index, "A\nLive", write)
		}
	}
}

func TestDebugMountedUseStdoutWriteReplaysManagedOutput(t *testing.T) {
	stdout := &recordingWriter{}
	phase := 0

	render := func() *vdom.Node {
		currentPhase := phase
		stdoutCtx := UseStdout()

		UseEffect(func() func() {
			if currentPhase == 1 {
				if _, err := stdoutCtx.Write("from stdout hook\n"); err != nil {
					t.Fatalf("unexpected stdout hook write error: %v", err)
				}
			}

			return nil
		}, []interface{}{currentPhase})

		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"A"}, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("Live"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	phase = 1

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	// Upstream Ink debug mode writes the hook payload, replays managed
	// output, then emits one append for the render itself — three writes.
	// See ink.tsx writeToStdout (debug branch) plus the unconditional
	// per-render write in the debug branch of onRender.
	if len(stdout.writes) != 3 {
		t.Fatalf("expected hook payload, managed replay, and post-render append on stdout, got %#v", stdout.writes)
	}

	if stdout.writes[0] != "from stdout hook\n" {
		t.Fatalf("expected hook payload on stdout first, got %#v", stdout.writes)
	}

	if stdout.writes[1] != "A\nLive" || stdout.writes[2] != "A\nLive" {
		t.Fatalf("expected debug stdout writes [1..2] to replay managed output %q, got %#v", "A\nLive", stdout.writes)
	}
}

func TestStaticOutputWritesOnlyNewItemsAboveDynamicBlock(t *testing.T) {
	stdout := &recordingWriter{}
	items := []string{"A"}

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems(items, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("X"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	items = []string{"A", "B"}

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected clear, static delta, and dynamic restore; got %#v", stdout.writes)
	}

	if stdout.writes[1] != "B\n" {
		t.Fatalf("expected only new static item to be written, got %q", stdout.writes[1])
	}

	if stdout.last() != "X\n" {
		t.Fatalf("expected dynamic output to be restored without static items, got %q", stdout.last())
	}
}

func TestStaticOutputIgnoresReplacementWithoutAppend(t *testing.T) {
	stdout := &recordingWriter{}
	items := []string{"A"}

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems(items, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("X"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	items = []string{"B"}

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected no output for non-appending static replacement, got %#v", stdout.writes)
	}
}

func TestDebugStaticOutputKeepsHistoryAndIgnoresReplacement(t *testing.T) {
	stdout := &recordingWriter{}
	items := []string{}

	render := func() *vdom.Node {
		return components.StaticItems(items, func(item string, index int) *vdom.Node {
			return components.Text(item)
		})
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.last() != "" {
		t.Fatalf("expected initial empty debug static output, got %q", stdout.last())
	}

	items = []string{"A"}
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("first rerender failed: %v", err)
	}
	if stdout.last() != "A\n" {
		t.Fatalf("expected first appended static item, got %q", stdout.last())
	}

	stdout.writes = nil
	items = []string{"B"}
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("second rerender failed: %v", err)
	}

	// Upstream Ink debug mode emits one append per render, replaying
	// `fullStaticOutput + output` regardless of whether the tree changed
	// (see ink.tsx onRender debug branch). Static-list replacement at the
	// same index is intentionally NOT mirrored above the dynamic block —
	// `fullStaticOutput` keeps the originally appended "A\n" — but the
	// render itself still produces a debug write.
	if len(stdout.writes) != 1 || stdout.writes[0] != "A\n" {
		t.Fatalf("expected debug rerender to replay accumulated static history, got %#v", stdout.writes)
	}
}

func TestThrottleStaticAppendBypassesDelay(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()
	items := []string{"A"}

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems(items, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("X"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil
	items = []string{"A", "B"}

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected throttled static append to render immediately, got %#v", stdout.writes)
	}

	if stdout.writes[1] != "B\n" {
		t.Fatalf("expected only appended static item to be written, got %q", stdout.writes[1])
	}

	if stdout.last() != "X\n" {
		t.Fatalf("expected dynamic block to be restored immediately, got %q", stdout.last())
	}

	writeCount := len(stdout.writes)
	clock.Advance(1 * time.Second)
	if len(stdout.writes) != writeCount {
		t.Fatalf("expected static append to cancel trailing throttled write, got %#v", stdout.writes)
	}
}

func TestMaxFPSLimitCanSelectUpstreamDefaultWithoutChangingLegacyZero(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("initial")
	}, RenderOptions{
		AppOptions:  AppOptions{Stdout: stdout},
		MaxFPSLimit: fpsLimit(DefaultMaxFPS),
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if instance.maxFPS != DefaultMaxFPS {
		t.Fatalf("expected upstream default max FPS %d, got %d", DefaultMaxFPS, instance.maxFPS)
	}

	if instance.renderThrottle != 34*time.Millisecond {
		t.Fatalf("expected upstream default throttle window 34ms, got %s", instance.renderThrottle)
	}

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("updated")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected upstream default max FPS to throttle immediate rerender, got %#v", stdout.writes)
	}

	clock.Advance(33 * time.Millisecond)
	if len(stdout.writes) != 0 {
		t.Fatalf("expected no trailing render before default throttle window, got %#v", stdout.writes)
	}

	clock.Advance(1 * time.Millisecond)
	if !strings.Contains(stdout.last(), "updated") {
		t.Fatalf("expected trailing render after default throttle window, got %#v", stdout.writes)
	}
}

func TestLegacyZeroMaxFPSRemainsUnthrottled(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("initial")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     0,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if instance.maxFPS != 0 || instance.renderThrottle != 0 {
		t.Fatalf("expected legacy MaxFPS: 0 to disable throttling, got fps=%d throttle=%s", instance.maxFPS, instance.renderThrottle)
	}

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("updated")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if !strings.Contains(stdout.last(), "updated") {
		t.Fatalf("expected legacy MaxFPS: 0 rerender to write immediately, got %#v", stdout.writes)
	}

	if len(clock.timers) != 0 {
		t.Fatalf("expected legacy MaxFPS: 0 to avoid scheduling throttle timers, got %#v", clock.timers)
	}
}

func TestMaxFPSLimitZeroOverridesLegacyMaxFPS(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("initial")
	}, RenderOptions{
		AppOptions:  AppOptions{Stdout: stdout},
		MaxFPS:      1,
		MaxFPSLimit: fpsLimit(0),
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if instance.maxFPS != 0 || instance.renderThrottle != 0 {
		t.Fatalf("expected MaxFPSLimit: 0 to disable throttling, got fps=%d throttle=%s", instance.maxFPS, instance.renderThrottle)
	}

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("updated")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if !strings.Contains(stdout.last(), "updated") {
		t.Fatalf("expected MaxFPSLimit: 0 rerender to write immediately, got %#v", stdout.writes)
	}
}

func TestThrottleStaticAppendReplacesPendingDynamicRender(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()
	items := []string{"A"}
	value := "X"

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems(items, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text(value),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil
	value = "Y"

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("first rerender failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected dynamic-only rerender to stay pending, got %#v", stdout.writes)
	}

	items = []string{"A", "B"}
	value = "Z"

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("second rerender failed: %v", err)
	}

	if len(stdout.writes) != 3 {
		t.Fatalf("expected static append to flush latest frame immediately, got %#v", stdout.writes)
	}

	if stdout.writes[1] != "B\n" {
		t.Fatalf("expected appended static item to be written, got %q", stdout.writes[1])
	}

	if stdout.last() != "Z\n" {
		t.Fatalf("expected latest dynamic output to win over pending frame, got %q", stdout.last())
	}

	writeCount := len(stdout.writes)
	clock.Advance(1 * time.Second)
	if len(stdout.writes) != writeCount {
		t.Fatalf("expected pending dynamic frame to be canceled after static flush, got %#v", stdout.writes)
	}
}

func TestThrottleExitRequestBypassesDelay(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()
	exitErr := errors.New("bye")
	shouldExit := false

	render := func() *vdom.Node {
		if shouldExit {
			UseApp().Exit(exitErr)
			return vdom.CreateTextNode("Final")
		}

		return vdom.CreateTextNode("Hello")
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil
	shouldExit = true

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if !instance.unmounted {
		t.Fatal("expected exit-triggering rerender to bypass throttle and unmount immediately")
	}

	if !strings.Contains(stdout.joined(), "Final") {
		t.Fatalf("expected final frame to be rendered before exit, got %#v", stdout.writes)
	}

	if err := instance.WaitUntilExit(); !errors.Is(err, exitErr) {
		t.Fatalf("expected waitUntilExit to return %v, got %v", exitErr, err)
	}

	writeCount := len(stdout.writes)
	clock.Advance(1 * time.Second)
	if len(stdout.writes) != writeCount {
		t.Fatalf("expected exit-triggering rerender to cancel trailing throttled writes, got %#v", stdout.writes)
	}
}

func TestThrottleRerenderCoalescesWithinWindow(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("World")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected throttled rerender to defer output, got %#v", stdout.writes)
	}

	clock.Advance(999 * time.Millisecond)
	if len(stdout.writes) != 0 {
		t.Fatalf("expected no trailing render before window end, got %#v", stdout.writes)
	}

	clock.Advance(1 * time.Millisecond)
	if len(stdout.writes) != 1 {
		t.Fatalf("expected one trailing render, got %#v", stdout.writes)
	}

	if !strings.Contains(stdout.last(), "World") {
		t.Fatalf("expected trailing render to output final content, got %q", stdout.last())
	}
}

func TestThrottleUnmountFlushesPendingRender(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Final")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected rerender to stay pending, got %#v", stdout.writes)
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if !strings.Contains(stdout.joined(), "Final") {
		t.Fatalf("expected unmount to flush final throttled frame, got %#v", stdout.writes)
	}
}

func TestThrottleUnmountCancelsPendingTimer(t *testing.T) {
	stdout := &recordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Foo")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Bar")
	}); err != nil {
		t.Fatalf("first rerender failed: %v", err)
	}
	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Baz")
	}); err != nil {
		t.Fatalf("second rerender failed: %v", err)
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	callCountAfterUnmount := len(stdout.writes)
	clock.Advance(2 * time.Second)
	if len(stdout.writes) != callCountAfterUnmount {
		t.Fatalf("expected no trailing render after unmount, got %#v", stdout.writes)
	}
}

func TestThrottleTTYWrapsImmediateAndTrailingWritesWithSyncEscapes(t *testing.T) {
	stdout := &ttyRecordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) < 3 || stdout.writes[0] != bsu || stdout.writes[len(stdout.writes)-1] != esu {
		t.Fatalf("expected initial throttled TTY render to be wrapped with sync escapes, got %#v", stdout.writes)
	}

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("World")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected trailing TTY render to wait, got %#v", stdout.writes)
	}

	clock.Advance(1 * time.Second)

	if len(stdout.writes) < 3 {
		t.Fatalf("expected wrapped trailing render writes, got %#v", stdout.writes)
	}

	if stdout.writes[0] != bsu || stdout.writes[len(stdout.writes)-1] != esu {
		t.Fatalf("expected sync escapes around trailing render, got %#v", stdout.writes)
	}

	if !strings.Contains(strings.Join(stdout.writes[1:len(stdout.writes)-1], ""), "World") {
		t.Fatalf("expected wrapped trailing writes to include content, got %#v", stdout.writes)
	}
}

func TestThrottleTTYUnchangedTrailingRenderDoesNotEmitSyncEscapes(t *testing.T) {
	stdout := &ttyRecordingWriter{}
	clock := newFakeThrottleClock()

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Hello")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	clock.Advance(1 * time.Second)

	if len(stdout.writes) != 0 {
		t.Fatalf("expected unchanged trailing render to avoid writes entirely, got %#v", stdout.writes)
	}
}

func TestThrottleTTYUnchangedTrailingRenderWithUnchangedCursorDoesNotEmitSyncEscapes(t *testing.T) {
	stdout := &ttyRecordingWriter{}
	clock := newFakeThrottleClock()

	render := func() *vdom.Node {
		UseCursor().SetCursorPosition(&CursorPosition{X: 0, Y: 0})
		return vdom.CreateTextNode("Hello")
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		MaxFPS:     1,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	attachFakeThrottleClock(instance, clock)
	stdout.writes = nil

	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	clock.Advance(1 * time.Second)

	if len(stdout.writes) != 0 {
		t.Fatalf("expected unchanged trailing render with unchanged cursor to avoid writes entirely, got %#v", stdout.writes)
	}
}

func TestTTYSessionRendersANSITextStyles(t *testing.T) {
	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{
			"backgroundColor": "green",
			"alignSelf":       "flex-start",
		}, components.Text(vdom.Props{
			"color":     "red",
			"bold":      true,
			"underline": true,
		}, "Hello"))
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) < 2 {
		t.Fatalf("expected rendered output writes, got %#v", stdout.writes)
	}

	output := stdout.writes[1]
	if !strings.Contains(output, "\x1b[31m") {
		t.Fatalf("expected ANSI foreground color in %q", output)
	}
	if !strings.Contains(output, "\x1b[42m") {
		t.Fatalf("expected ANSI background color in %q", output)
	}
	if !strings.Contains(output, "\x1b[1m") || !strings.Contains(output, "\x1b[4m") {
		t.Fatalf("expected ANSI text modifiers in %q", output)
	}
	if !strings.Contains(output, "Hello") {
		t.Fatalf("expected text content in %q", output)
	}
}

func TestTTYSessionRendersANSIBorderAndBackground(t *testing.T) {
	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{
			"borderStyle":     "round",
			"borderColor":     "green",
			"backgroundColor": "cyan",
			"width":           8.0,
			"height":          3.0,
			"alignSelf":       "flex-start",
		}, components.Text("Hi"))
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) < 2 {
		t.Fatalf("expected rendered output writes, got %#v", stdout.writes)
	}

	output := stdout.writes[1]
	if !strings.Contains(output, "\x1b[32m") {
		t.Fatalf("expected ANSI border color in %q", output)
	}
	if !strings.Contains(output, "\x1b[46m") {
		t.Fatalf("expected ANSI box background in %q", output)
	}
	if !strings.Contains(output, "╭") || !strings.Contains(output, "╯") {
		t.Fatalf("expected border glyphs in %q", output)
	}
}

func TestTTYSessionRendersANSIStaticOutput(t *testing.T) {
	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems([]string{"Static"}, func(item string, index int) *vdom.Node {
				return components.Text(vdom.Props{"color": "red"}, item)
			}),
			components.Text("Live"),
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	joined := stdout.joined()
	if !strings.Contains(joined, "\x1b[31mStatic") {
		t.Fatalf("expected ANSI-styled static output in %q", joined)
	}
	if !strings.Contains(joined, "Live") {
		t.Fatalf("expected dynamic output in %q", joined)
	}
}

func TestTTYSessionRendersANSINestedTextStyleResumption(t *testing.T) {
	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Text(
			vdom.Props{"color": "green"},
			"A ",
			components.Text(vdom.Props{"color": "blue"}, "B"),
			" C",
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) < 2 {
		t.Fatalf("expected rendered output writes, got %#v", stdout.writes)
	}

	output := stdout.writes[1]
	expected := "\x1b[32mA \x1b[34mB\x1b[32m C\x1b[39m\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestTTYSessionRendersANSINestedTruncateTransitions(t *testing.T) {
	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{"width": 4.0},
			components.Text(
				vdom.Props{"color": "green", "wrap": "truncate"},
				"A ",
				components.Text(vdom.Props{"color": "blue"}, "BC"),
				"D",
			),
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) < 2 {
		t.Fatalf("expected rendered output writes, got %#v", stdout.writes)
	}

	output := stdout.writes[1]
	expected := "\x1b[32mA \x1b[34mB…\x1b[39m\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestTTYSessionRendersANSINestedTransformTransitions(t *testing.T) {
	stdout := &ttyRecordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Text(
			vdom.Props{"color": "green"},
			"A",
			components.Transform(func(children string, index int) string {
				return "<" + children + ">"
			}, components.Text(vdom.Props{"color": "blue"}, "B")),
			"C",
		)
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) < 2 {
		t.Fatalf("expected rendered output writes, got %#v", stdout.writes)
	}

	output := stdout.writes[1]
	expected := "\x1b[32mA<\x1b[34mB\x1b[32m>C\x1b[39m\n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestScreenReaderSessionWritesPlainAccessibleOutput(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{"aria-role": "button"},
			components.Text(vdom.Props{"bold": true}, "Click me"),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout:              stdout,
			ScreenReaderEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) != 1 {
		t.Fatalf("expected single plain write in screen-reader mode, got %#v", stdout.writes)
	}

	if stdout.writes[0] != "button: Click me" {
		t.Fatalf("expected accessible plain output, got %q", stdout.writes[0])
	}

	if strings.Contains(stdout.writes[0], "\x1b[") {
		t.Fatalf("expected no ANSI escapes in screen-reader output, got %q", stdout.writes[0])
	}
}

func TestScreenReaderSessionRendersSelectInputLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		items := []string{"Red", "Green", "Blue"}
		children := []*vdom.Node{
			components.Text("Select a color:"),
		}

		for index, item := range items {
			label := fmt.Sprintf("%d. %s", index+1, item)
			props := vdom.Props{
				"aria-label": label,
				"aria-role":  "listitem",
			}
			if index == 1 {
				props["aria-state"] = map[string]bool{"selected": true}
			}

			children = append(children, components.Box(props, components.Text(item)))
		}

		return components.Box(vdom.Props{
			"aria-role":     "list",
			"flexDirection": "column",
		}, children...)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout:              stdout,
			ScreenReaderEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	expected := "list: Select a color:\nlistitem: 1. Red\nlistitem: (selected) 2. Green\nlistitem: 3. Blue"
	if stdout.last() != expected {
		t.Fatalf("expected %q, got %q", expected, stdout.last())
	}
}

func TestScreenReaderSessionRendersNestedRowLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Box(vdom.Props{"flexDirection": "row"},
				components.Text("Line 1"),
				components.Text("Line 2"),
			),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout:              stdout,
			ScreenReaderEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.last() != "Line 1 Line 2" {
		t.Fatalf("expected nested row screen-reader output %q, got %q", "Line 1 Line 2", stdout.last())
	}
}

func TestScreenReaderSessionRendersComponentReturningNilLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}

	nullComponent := func() *vdom.Node {
		return nil
	}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Text("Hello"),
			nullComponent(),
			components.Text("World"),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout:              stdout,
			ScreenReaderEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if stdout.last() != "Hello\nWorld" {
		t.Fatalf("expected screen-reader output without nil child gap %q, got %q", "Hello\nWorld", stdout.last())
	}
}

func TestScreenReaderUnmountDoesNotEraseAccessibleOutput(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return components.Box(vdom.Props{"aria-role": "button"},
			components.Text("Click me"),
		)
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout:              stdout,
			ScreenReaderEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if len(stdout.writes) != 1 {
		t.Fatalf("expected single initial screen-reader write, got %#v", stdout.writes)
	}

	stdout.writes = nil

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		t.Fatalf("expected nil waitUntilExit error, got %v", err)
	}

	if len(stdout.writes) != 0 {
		t.Fatalf("expected screen-reader unmount to avoid erase writes, got %#v", stdout.writes)
	}
}

// TestDebugTakesPrecedenceOverScreenReaderRerender locks upstream Ink's render
// branch precedence: when both `debug` and `isScreenReaderEnabled` are set,
// the debug append-only path wins (see ink/src/ink.tsx onRender — debug check
// fires before the screen-reader branch). A second render must therefore
// produce a single append-only write of the new logical output, with no
// `eraseLines` escape sequences from the screen-reader rewrite path.
func TestDebugTakesPrecedenceOverScreenReaderRerender(t *testing.T) {
	stdout := &recordingWriter{}

	makeRoot := func(label string) func() *vdom.Node {
		return func() *vdom.Node {
			return components.Box(vdom.Props{"aria-role": "button"},
				components.Text(label),
			)
		}
	}

	instance, err := MountWithOptions(makeRoot("Click"), RenderOptions{
		AppOptions: AppOptions{
			Stdout:              stdout,
			ScreenReaderEnabled: true,
		},
		Debug: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) != 1 || stdout.writes[0] != "button: Click" {
		t.Fatalf("expected single debug-mode initial write of plain accessible output, got %#v", stdout.writes)
	}

	stdout.writes = nil
	if err := instance.Rerender(makeRoot("Submit")); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	if len(stdout.writes) != 1 {
		t.Fatalf("expected single debug append on rerender, got %#v", stdout.writes)
	}

	if stdout.writes[0] != "button: Submit" {
		t.Fatalf("expected debug append %q, got %q", "button: Submit", stdout.writes[0])
	}

	if strings.Contains(stdout.writes[0], "\x1b[2K") || strings.Contains(stdout.writes[0], "\x1b[1A") {
		t.Fatalf("expected debug rerender to avoid screen-reader erase escapes, got %q", stdout.writes[0])
	}
}

// TestDebugRerenderEmitsAppendEvenWhenOutputUnchanged locks upstream Ink's
// debug-mode contract: every render produces exactly one append, even when
// the rendered string is byte-identical to the previous frame. Upstream
// onRender always writes `fullStaticOutput + output` in the debug branch
// (no equality short-circuit), so a stream of identical renders is faithfully
// duplicated in the debug log. Goink previously skipped identical writes,
// breaking the append-only timeline contract.
func TestDebugRerenderEmitsAppendEvenWhenOutputUnchanged(t *testing.T) {
	stdout := &recordingWriter{}

	makeRoot := func() *vdom.Node {
		return vdom.CreateTextNode("hello")
	}

	instance, err := MountWithOptions(makeRoot, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if len(stdout.writes) != 1 || stdout.writes[0] != "hello" {
		t.Fatalf("expected single initial debug write %q, got %#v", "hello", stdout.writes)
	}

	stdout.writes = nil
	if err := instance.Rerender(makeRoot); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}
	if err := instance.Rerender(makeRoot); err != nil {
		t.Fatalf("second rerender failed: %v", err)
	}

	if len(stdout.writes) != 2 {
		t.Fatalf("expected two append-only debug writes for two identical rerenders, got %#v", stdout.writes)
	}
	for index, write := range stdout.writes {
		if write != "hello" {
			t.Fatalf("expected debug append #%d to equal %q, got %q", index, "hello", write)
		}
	}
}

// TestStaticDeltaOfPureNewlineDoesNotRewriteDynamic locks upstream Ink's
// `hasStaticOutput = staticOutput && staticOutput !== '\n'` filter (see
// ink/src/ink.tsx onRender). Any static delta that consists solely of a
// trailing newline must be ignored: the dynamic block is left intact and
// `fullStaticOutput` does not grow. Goink historically treated any non-empty
// staticOutput as a real append, which fired clearLocked + write(static)
// even when the renderer produced just a newline.
func TestStaticDeltaOfPureNewlineDoesNotRewriteDynamic(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("dyn")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	previousFullStatic := instance.fullStaticOutput

	prepared := preparedRender{
		logicalOutput:  "dyn",
		renderedOutput: "dyn\n",
		staticOutput:   "\n",
		shouldWrite:    true,
	}
	if err := instance.commitPreparedRenderLocked(prepared); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	if instance.fullStaticOutput != previousFullStatic {
		t.Fatalf("expected fullStaticOutput to remain %q after pure-newline static delta, got %q", previousFullStatic, instance.fullStaticOutput)
	}

	for _, write := range stdout.writes {
		if write == "\n" {
			t.Fatalf("expected pure-newline static delta to be filtered, got writes %#v", stdout.writes)
		}
	}
}

// TestIncrementalConsecutiveClearsAreIdempotent locks the upstream invariant
// that calling clear() multiple times in a row leaves incremental
// rendering's internal state in a fully-reset shape — a subsequent render
// must do a fresh full write (eraseLines(0) + payload), not a surgical
// diff against stale `previousLines`. Mirrors upstream test
// 'incremental rendering - multiple consecutive clear() calls (should be harmless no-ops)'.
func TestIncrementalConsecutiveClearsAreIdempotent(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Line 1\nLine 2\nLine 3\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	for index := 0; index < 3; index++ {
		if err := instance.Clear(); err != nil {
			t.Fatalf("clear #%d failed: %v", index, err)
		}
	}

	if len(instance.previousLines) != 0 {
		t.Fatalf("expected previousLines reset after consecutive clears, got %#v", instance.previousLines)
	}
	if instance.previousOutput != "" {
		t.Fatalf("expected previousOutput reset after consecutive clears, got %q", instance.previousOutput)
	}

	stdout.writes = nil
	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("New content\n")
	}); err != nil {
		t.Fatalf("rerender after clears failed: %v", err)
	}

	combined := strings.Join(stdout.writes, "")
	if !strings.Contains(combined, "New content\n") {
		t.Fatalf("expected fresh full write of new content after consecutive clears, got %#v", stdout.writes)
	}

	// A surgical diff against stale `previousLines` would emit ansiCursorTo +
	// per-line tail rewrites; a fresh write goes through the
	// `previousOutput == ""` branch and emits eraseLines(0) (empty) + payload.
	if strings.Contains(combined, ansiCursorNextLine()) {
		t.Fatalf("expected fresh write after clears, but found surgical cursor escapes in %#v", stdout.writes)
	}
}

// TestScreenReaderStaticAppendReemitsDynamicBlock locks upstream Ink's
// screen-reader path with a non-empty staticOutput: the dynamic block must
// be re-emitted alongside the new static delta even when the dynamic
// rendered string is byte-identical to the previous frame. Upstream's
// `if (output === this.lastOutput && !hasStaticOutput)` short-circuit
// intentionally requires `!hasStaticOutput` to skip — when static grows we
// always rewrite the dynamic part below it (see ink.tsx onRender).
func TestScreenReaderStaticAppendReemitsDynamicBlock(t *testing.T) {
	stdout := &recordingWriter{}
	items := []string{"A"}

	render := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.StaticItems(items, func(item string, index int) *vdom.Node {
				return components.Text(item)
			}),
			components.Text("dyn"),
		)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout:              stdout,
			ScreenReaderEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	stdout.writes = nil
	items = []string{"A", "B"}
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	combined := strings.Join(stdout.writes, "")
	if !strings.Contains(combined, "B\n") {
		t.Fatalf("expected appended static delta %q in screen-reader output, got %#v", "B\n", stdout.writes)
	}

	if !strings.Contains(combined, "dyn") {
		t.Fatalf("expected dynamic block to be re-emitted alongside static append, got %#v", stdout.writes)
	}
}
