package devtools_test

import (
	"strings"
	"testing"
	"time"

	"github.com/dh-kam/ink-go/pkg/devtools"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// collectText flattens all text nodes underneath a vdom subtree into a single
// string. The non-compact panel renders each metric as its own Text element,
// so tests assert against the joined output rather than fishing through the
// child tree by index.
func collectText(node *vdom.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == vdom.TextNode {
		return node.Text
	}
	var b strings.Builder
	for _, child := range node.Children {
		if child == nil {
			continue
		}
		b.WriteString(collectText(child))
		b.WriteByte('\n')
	}
	return b.String()
}

// findProp walks the subtree until it finds an element with the supplied
// prop key set, returning the first match (BFS order).
func findProp(node *vdom.Node, key string) (interface{}, bool) {
	if node == nil {
		return nil, false
	}
	queue := []*vdom.Node{node}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == nil {
			continue
		}
		if current.Type == vdom.ElementNode {
			if v, ok := current.Props[key]; ok {
				return v, true
			}
		}
		queue = append(queue, current.Children...)
	}
	return nil, false
}

func TestDebugPanelEmptyData(t *testing.T) {
	node := devtools.DebugPanel(devtools.DebugPanelProps{})
	if node == nil {
		t.Fatal("DebugPanel returned nil")
	}

	out := collectText(node)
	required := []string{
		"Renders: 0 (0 hit)",
		"Last:    0s",
		"Avg:     0s",
		"State slots: 0",
		"Effects: 0",
		"Patches/frame: 0",
	}
	for _, r := range required {
		if !strings.Contains(out, r) {
			t.Errorf("zero-data panel missing %q in:\n%s", r, out)
		}
	}

	if strings.Contains(out, "Custom") {
		t.Errorf("zero-data panel should not include Custom divider:\n%s", out)
	}
}

func TestDebugPanelAllFields(t *testing.T) {
	data := devtools.PanelData{
		Title:        "MyDebug",
		Renders:      42,
		CacheHits:    12,
		LastDuration: 3400 * time.Microsecond,
		AvgDuration:  2100 * time.Microsecond,
		StateCount:   5,
		EffectCount:  3,
		PatchCount:   7,
		Custom: map[string]string{
			"alpha": "one",
			"beta":  "two",
		},
	}

	node := devtools.DebugPanel(devtools.DebugPanelProps{Data: data, Width: 40})
	out := collectText(node)

	wantContains := []string{
		"Renders: 42 (12 hit)",
		"Last:    3.4ms",
		"Avg:     2.1ms",
		"State slots: 5",
		"Effects: 3",
		"Patches/frame: 7",
		"─ Custom ─",
		"alpha: one",
		"beta: two",
	}
	for _, w := range wantContains {
		if !strings.Contains(out, w) {
			t.Errorf("expected output to contain %q, got:\n%s", w, out)
		}
	}

	// alpha sorts before beta deterministically
	idxA := strings.Index(out, "alpha:")
	idxB := strings.Index(out, "beta:")
	if idxA < 0 || idxB < 0 || idxA > idxB {
		t.Errorf("custom keys not sorted alphabetically: alpha@%d beta@%d", idxA, idxB)
	}

	// Title flows through to the border label.
	label, ok := findProp(node, "borderLabel")
	if !ok {
		t.Fatalf("expected borderLabel prop on rendered panel")
	}
	if got, _ := label.(string); got != "MyDebug" {
		t.Errorf("borderLabel = %q, want %q", got, "MyDebug")
	}
}

func TestDebugPanelDefaultTitle(t *testing.T) {
	node := devtools.DebugPanel(devtools.DebugPanelProps{})
	label, ok := findProp(node, "borderLabel")
	if !ok {
		t.Fatalf("default title should still set borderLabel")
	}
	if got, _ := label.(string); got != "DEBUG" {
		t.Errorf("default borderLabel = %q, want %q", got, "DEBUG")
	}
}

func TestDebugPanelCompact(t *testing.T) {
	data := devtools.PanelData{
		LastDuration: 3400 * time.Microsecond,
		StateCount:   5,
		PatchCount:   7,
		Custom:       map[string]string{"mode": "dev"},
	}
	node := devtools.DebugPanel(devtools.DebugPanelProps{Data: data, Compact: true})
	if node == nil {
		t.Fatal("compact DebugPanel returned nil")
	}

	out := collectText(node)
	if strings.Contains(out, "\n") {
		// collectText appends a newline per child wrapper; the compact panel
		// is a single Text node so there should be no inner newline boundary
		// before the trailing wrap.
		trimmed := strings.TrimRight(out, "\n")
		if strings.Contains(trimmed, "\n") {
			t.Errorf("compact panel should be a single line, got:\n%q", out)
		}
	}

	for _, want := range []string{"3.4ms", "5 states", "7 patches", "mode=dev"} {
		if !strings.Contains(out, want) {
			t.Errorf("compact line missing %q, got %q", want, out)
		}
	}
}

func TestDebugPanelWidthDefault(t *testing.T) {
	node := devtools.DebugPanel(devtools.DebugPanelProps{})
	w, ok := findProp(node, "width")
	if !ok {
		t.Fatalf("expected width prop")
	}
	if got, _ := w.(int); got != 30 {
		t.Errorf("default width = %v, want 30", got)
	}
}

func TestDebugPanelWidthExplicit(t *testing.T) {
	node := devtools.DebugPanel(devtools.DebugPanelProps{Width: 50})
	w, ok := findProp(node, "width")
	if !ok {
		t.Fatalf("expected width prop")
	}
	if got, _ := w.(int); got != 50 {
		t.Errorf("explicit width = %v, want 50", got)
	}
}

func TestWithDebugRightLayout(t *testing.T) {
	app := func() *vdom.Node { return vdom.CreateTextNode("APP") }
	node := devtools.WithDebug(devtools.WithDebugProps{
		App:   app,
		Panel: devtools.PanelData{Renders: 1},
		Side:  "right",
	})
	if node == nil {
		t.Fatal("WithDebug returned nil")
	}
	dir, ok := node.Props["flexDirection"]
	if !ok {
		t.Fatalf("expected flexDirection on outer Box")
	}
	if dir != "row" {
		t.Errorf("flexDirection = %v, want row", dir)
	}
	// Two children: app + panel.
	if len(node.Children) != 2 {
		t.Errorf("expected 2 children (app + panel), got %d", len(node.Children))
	}
	if !strings.Contains(collectText(node), "APP") {
		t.Errorf("layout missing app content")
	}
	if !strings.Contains(collectText(node), "Renders: 1") {
		t.Errorf("layout missing panel content")
	}
}

func TestWithDebugBottomLayout(t *testing.T) {
	app := func() *vdom.Node { return vdom.CreateTextNode("APP") }
	node := devtools.WithDebug(devtools.WithDebugProps{
		App:   app,
		Panel: devtools.PanelData{},
		Side:  "bottom",
	})
	if dir := node.Props["flexDirection"]; dir != "column" {
		t.Errorf("Side=bottom flexDirection = %v, want column", dir)
	}
}

func TestWithDebugDefaultSideIsRow(t *testing.T) {
	app := func() *vdom.Node { return vdom.CreateTextNode("APP") }
	node := devtools.WithDebug(devtools.WithDebugProps{App: app})
	if dir := node.Props["flexDirection"]; dir != "row" {
		t.Errorf("default Side flexDirection = %v, want row", dir)
	}
}

func TestWithDebugNilApp(t *testing.T) {
	node := devtools.WithDebug(devtools.WithDebugProps{
		App:   nil,
		Panel: devtools.PanelData{Renders: 9},
	})
	if node == nil {
		t.Fatal("WithDebug nil App returned nil")
	}
	if len(node.Children) != 1 {
		t.Errorf("nil App should leave just the panel, got %d children", len(node.Children))
	}
	if !strings.Contains(collectText(node), "Renders: 9") {
		t.Errorf("expected panel to still render with nil app")
	}
}

func TestWithDebugAppReturnsNil(t *testing.T) {
	// An App fn that yields nil should be skipped exactly like a nil App.
	node := devtools.WithDebug(devtools.WithDebugProps{
		App:   func() *vdom.Node { return nil },
		Panel: devtools.PanelData{},
	})
	if len(node.Children) != 1 {
		t.Errorf("App returning nil should be skipped, got %d children", len(node.Children))
	}
}

func TestDebugPanelDurationFormatting(t *testing.T) {
	cases := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "0s"},
		{"nanoseconds", 250 * time.Nanosecond, "250ns"},
		{"microseconds", 1500 * time.Nanosecond, "1.5µs"},
		{"milliseconds", 3400 * time.Microsecond, "3.4ms"},
		{"seconds", 2500 * time.Millisecond, "2.50s"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			node := devtools.DebugPanel(devtools.DebugPanelProps{
				Data: devtools.PanelData{LastDuration: tc.d},
			})
			out := collectText(node)
			if !strings.Contains(out, tc.want) {
				t.Errorf("duration %v -> want %q in output, got:\n%s", tc.d, tc.want, out)
			}
		})
	}
}

func TestDebugPanelCustomMapEmpty(t *testing.T) {
	// An empty (but non-nil) Custom map must not produce a divider row.
	node := devtools.DebugPanel(devtools.DebugPanelProps{
		Data: devtools.PanelData{Custom: map[string]string{}},
	})
	out := collectText(node)
	if strings.Contains(out, "Custom") {
		t.Errorf("empty Custom map should not render divider, got:\n%s", out)
	}
}

func TestDebugPanelCompactZeroData(t *testing.T) {
	node := devtools.DebugPanel(devtools.DebugPanelProps{Compact: true})
	out := collectText(node)
	for _, want := range []string{"0s", "0 states", "0 patches"} {
		if !strings.Contains(out, want) {
			t.Errorf("compact zero data missing %q, got %q", want, out)
		}
	}
}
