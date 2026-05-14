package ink_test

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

type upstreamMeasureStdout struct {
	mu      sync.Mutex
	writes  []string
	columns int
	rows    int
}

func (stdout *upstreamMeasureStdout) Write(data []byte) (int, error) {
	stdout.mu.Lock()
	defer stdout.mu.Unlock()

	stdout.writes = append(stdout.writes, string(data))
	return len(data), nil
}

func (stdout *upstreamMeasureStdout) Columns() int {
	return stdout.columns
}

func (stdout *upstreamMeasureStdout) Rows() int {
	if stdout.rows > 0 {
		return stdout.rows
	}

	return 24
}

func (stdout *upstreamMeasureStdout) First() string {
	stdout.mu.Lock()
	defer stdout.mu.Unlock()

	if len(stdout.writes) == 0 {
		return ""
	}

	return stdout.writes[0]
}

func (stdout *upstreamMeasureStdout) Last() string {
	stdout.mu.Lock()
	defer stdout.mu.Unlock()

	if len(stdout.writes) == 0 {
		return ""
	}

	return stdout.writes[len(stdout.writes)-1]
}

func (stdout *upstreamMeasureStdout) Reset() {
	stdout.mu.Lock()
	defer stdout.mu.Unlock()

	stdout.writes = nil
}

func (stdout *upstreamMeasureStdout) Snapshot() []string {
	stdout.mu.Lock()
	defer stdout.mu.Unlock()

	snapshot := make([]string, len(stdout.writes))
	copy(snapshot, stdout.writes)
	return snapshot
}

var ansiSequencePattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func stripANSISequences(text string) string {
	return ansiSequencePattern.ReplaceAllString(text, "")
}

func waitForMeasuredWrite(t *testing.T, stdout *upstreamMeasureStdout, want string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		if got := stdout.Last(); got == want {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("expected last write %q, got %q from %#v", want, stdout.Last(), stdout.Snapshot())
		}

		time.Sleep(5 * time.Millisecond)
	}
}

func waitForMeasuredRenderedText(t *testing.T, stdout *upstreamMeasureStdout, want string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		got := strings.TrimSpace(stripANSISequences(stdout.Last()))
		if got == want {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("expected last rendered text %q, got %q from %#v", want, got, stdout.Snapshot())
		}

		time.Sleep(5 * time.Millisecond)
	}
}

func TestDOMRefExposesInkHostNodeNamesAndTraversal(t *testing.T) {
	var current *ink.DOMElement

	app := ink.NewApp(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ = ref.Current().(*ink.DOMElement)
			return nil
		}, []interface{}{"capture"})

		return vdom.CreateElement("box", vdom.Props{
			"data-id":                "measured-box",
			"data-index":             2,
			"data-extra":             []string{"hidden"},
			"ref":                    ref,
			"width":                  float64(12),
			"internal_accessibility": "hidden",
		},
			vdom.CreateTextNode("prefix"),
			nil,
			vdom.CreateElement("text", nil, vdom.CreateTextNode("suffix")),
		)
	})

	app.RenderOnce()

	if current == nil {
		t.Fatal("expected ref current to capture the rendered DOM element")
	}
	if got := current.NodeName(); got != "ink-box" {
		t.Fatalf("expected ref node name %q, got %q", "ink-box", got)
	}
	if got := ink.MeasureElement(current); got != (ink.ElementDimensions{Width: 12, Height: 1}) {
		t.Fatalf("expected measured ref dimensions {12 1}, got %+v", got)
	}
	if got := current.Attributes()["data-id"]; got != "measured-box" {
		t.Fatalf("expected data-id attribute %q, got %v", "measured-box", got)
	}
	if got := current.Attributes()["data-index"]; got != 2 {
		t.Fatalf("expected data-index attribute %d, got %v", 2, got)
	}
	if _, ok := current.Attributes()["width"]; ok {
		t.Fatal("expected width style prop to be hidden from ref-facing DOM attributes")
	}
	if _, ok := current.Attributes()["data-extra"]; ok {
		t.Fatal("expected slice-valued prop to be hidden from ref-facing DOM attributes")
	}
	if _, ok := current.Attributes()["ref"]; ok {
		t.Fatal("expected ref prop to be hidden from ref-facing DOM attributes")
	}
	if _, ok := current.Attributes()["internal_accessibility"]; ok {
		t.Fatal("expected internal accessibility prop to be hidden from ref-facing DOM attributes")
	}

	children := current.ChildNodes()
	if len(children) != 2 {
		t.Fatalf("expected 2 child nodes, got %d", len(children))
	}
	if got := children[0].NodeName(); got != "#text" {
		t.Fatalf("expected first child node name %q, got %q", "#text", got)
	}
	if got := children[1].NodeName(); got != "ink-text" {
		t.Fatalf("expected second child node name %q, got %q", "ink-text", got)
	}
	if children[0].ParentNode() != current || children[1].ParentNode() != current {
		t.Fatal("expected child nodes to point back at the ref parent")
	}
	if children[0].NextSibling() != children[1] {
		t.Fatal("expected first child next sibling to be the nested text node")
	}
	if children[1].PreviousSibling() != children[0] {
		t.Fatal("expected nested text node previous sibling to be the first child")
	}
	if got := current.TextContent(); got != "prefixsuffix" {
		t.Fatalf("expected ref text content %q, got %q", "prefixsuffix", got)
	}
}

func TestDOMRefMeasureElementAfterStateUpdate(t *testing.T) {
	var setItems ink.SetStateFunc

	app := ink.NewApp(func() *vdom.Node {
		itemsValue, setItemsState := ink.UseState([]string{})
		heightValue, setHeight := ink.UseState(0)
		ref := ink.UseRef((*ink.DOMElement)(nil))

		setItems = setItemsState
		items := itemsValue.([]string)
		height := heightValue.(int)

		ink.UseEffect(func() func() {
			current, _ := ref.Current().(*ink.DOMElement)
			if current == nil {
				return nil
			}

			setHeight(ink.MeasureElement(current).Height)
			return nil
		}, []interface{}{len(items)})

		measuredChildren := make([]*vdom.Node, 0, len(items))
		for _, item := range items {
			measuredChildren = append(measuredChildren, vdom.CreateElement("text", nil, vdom.CreateTextNode(item)))
		}

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			vdom.CreateElement("box", vdom.Props{"ref": ref, "flexDirection": "column"}, measuredChildren...),
			vdom.CreateElement("text", nil, vdom.CreateTextNode(fmt.Sprintf("Height: %d", height))),
		)
	})

	if got := app.RenderOnce(); got != "Height: 0" {
		t.Fatalf("expected initial measured output %q, got %q", "Height: 0", got)
	}

	setItems([]string{"line 1", "line 2", "line 3"})
	app.RenderOnce()

	if got := app.RenderOnce(); got != "line 1\nline 2\nline 3\nHeight: 3" {
		t.Fatalf("expected measured output after update %q, got %q", "line 1\nline 2\nline 3\nHeight: 3", got)
	}
}

func TestDOMRefMeasureElementAfterMultipleStateUpdates(t *testing.T) {
	var setItems ink.SetStateFunc

	app := ink.NewApp(func() *vdom.Node {
		itemsValue, setItemsState := ink.UseState([]string{})
		heightValue, setHeight := ink.UseState(0)
		ref := ink.UseRef((*ink.DOMElement)(nil))

		setItems = setItemsState
		items := itemsValue.([]string)
		height := heightValue.(int)

		ink.UseEffect(func() func() {
			current, _ := ref.Current().(*ink.DOMElement)
			if current == nil {
				return nil
			}

			setHeight(ink.MeasureElement(current).Height)
			return nil
		}, []interface{}{len(items)})

		measuredChildren := make([]*vdom.Node, 0, len(items))
		for _, item := range items {
			measuredChildren = append(measuredChildren, vdom.CreateElement("text", nil, vdom.CreateTextNode(item)))
		}

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			vdom.CreateElement("box", vdom.Props{"ref": ref, "flexDirection": "column"}, measuredChildren...),
			vdom.CreateElement("text", nil, vdom.CreateTextNode(fmt.Sprintf("Height: %d", height))),
		)
	})

	app.RenderOnce()

	setItems([]string{"line 1", "line 2", "line 3"})
	app.RenderOnce()
	app.RenderOnce()

	setItems([]string{"line 1"})
	app.RenderOnce()

	if got := app.RenderOnce(); got != "line 1\nHeight: 1" {
		t.Fatalf("expected measured output after multiple updates %q, got %q", "line 1\nHeight: 1", got)
	}
}

func TestDOMRefMeasureElementMatchesUpstreamAvailableWidth(t *testing.T) {
	stdout := &upstreamMeasureStdout{columns: 100, rows: 24}

	instance, err := ink.RenderWithOptions(func() *vdom.Node {
		widthValue, setWidth := ink.UseState(0)
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ := ref.Current().(*ink.DOMElement)
			if current != nil {
				setWidth(ink.MeasureElement(current).Width)
			}

			return nil
		}, []interface{}{"measure"})

		return vdom.CreateElement("box", vdom.Props{"ref": ref},
			vdom.CreateElement("text", nil, vdom.CreateTextNode(fmt.Sprintf("Width: %d", widthValue.(int)))),
		)
	}, ink.RenderOptions{
		AppOptions: ink.AppOptions{Stdout: stdout},
		Debug:      true,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	defer instance.Unmount()

	waitForMeasuredWrite(t, stdout, "Width: 100", 200*time.Millisecond)

	if got := stdout.First(); got != "Width: 0" {
		t.Fatalf("expected initial measured output %q, got %q", "Width: 0", got)
	}
}

func TestDOMRefMeasureElementMatchesUpstreamAvailableWidthWhileThrottled(t *testing.T) {
	stdout := &upstreamMeasureStdout{columns: 100, rows: 24}

	instance, err := ink.RenderWithOptions(nil, ink.RenderOptions{
		AppOptions: ink.AppOptions{Stdout: stdout},
		Debug:      true,
		MaxFPS:     20,
	})
	if err != nil {
		t.Fatalf("initial nil render failed: %v", err)
	}
	defer instance.Unmount()

	stdout.Reset()

	_, err = ink.RenderWithOptions(func() *vdom.Node {
		widthValue, setWidth := ink.UseState(0)
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ := ref.Current().(*ink.DOMElement)
			if current != nil {
				setWidth(ink.MeasureElement(current).Width)
			}

			return nil
		}, []interface{}{"measure"})

		return vdom.CreateElement("box", vdom.Props{"ref": ref},
			vdom.CreateElement("text", nil, vdom.CreateTextNode(fmt.Sprintf("Width: %d", widthValue.(int)))),
		)
	}, ink.RenderOptions{
		AppOptions: ink.AppOptions{Stdout: stdout},
		Debug:      true,
		MaxFPS:     20,
	})
	if err != nil {
		t.Fatalf("throttled measured rerender failed: %v", err)
	}

	waitForMeasuredWrite(t, stdout, "Width: 100", 300*time.Millisecond)
}

func TestDOMRefMeasureElementMatchesUpstreamThrottledRerenderOutput(t *testing.T) {
	stdout := &upstreamMeasureStdout{columns: 100, rows: 24}

	instance, err := ink.RenderWithOptions(nil, ink.RenderOptions{
		AppOptions: ink.AppOptions{Stdout: stdout},
		MaxFPS:     20,
	})
	if err != nil {
		t.Fatalf("initial nil render failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.Rerender(func() *vdom.Node {
		widthValue, setWidth := ink.UseState(0)
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ := ref.Current().(*ink.DOMElement)
			if current != nil {
				setWidth(ink.MeasureElement(current).Width)
			}

			return nil
		}, []interface{}{"measure"})

		return vdom.CreateElement("box", vdom.Props{"ref": ref},
			vdom.CreateElement("text", nil, vdom.CreateTextNode(fmt.Sprintf("Width: %d", widthValue.(int)))),
		)
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	waitForMeasuredRenderedText(t, stdout, "Width: 100", 300*time.Millisecond)
}
