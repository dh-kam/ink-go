package devtools

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

func TestInspect_NilNode(t *testing.T) {
	if got := Inspect(nil); got != nil {
		t.Fatalf("Inspect(nil) = %#v, want nil", got)
	}
}

func TestInspect_TextNode(t *testing.T) {
	node := vdom.CreateTextNode("Hello")
	node.Key = "greeting"

	snap := Inspect(node)
	if snap == nil {
		t.Fatal("Inspect returned nil for text node")
	}
	if snap.Type != "text" {
		t.Errorf("snap.Type = %q, want %q", snap.Type, "text")
	}
	if snap.Text != "Hello" {
		t.Errorf("snap.Text = %q, want %q", snap.Text, "Hello")
	}
	if snap.Key != "greeting" {
		t.Errorf("snap.Key = %q, want %q", snap.Key, "greeting")
	}
	if snap.Children != nil {
		t.Errorf("snap.Children = %v, want nil", snap.Children)
	}
	if snap.Props != nil {
		t.Errorf("snap.Props = %v, want nil", snap.Props)
	}
}

func TestInspect_NestedChildren(t *testing.T) {
	leaf1 := vdom.CreateTextNode("Hello")
	leaf2 := vdom.CreateTextNode("World")
	inner := vdom.CreateElement("box", vdom.Props{"flexDirection": "column"}, leaf2)
	root := vdom.CreateElement("box", nil, leaf1, inner)

	snap := Inspect(root)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Type != "box" {
		t.Errorf("root.Type = %q, want box", snap.Type)
	}
	if len(snap.Children) != 2 {
		t.Fatalf("root.Children len = %d, want 2", len(snap.Children))
	}
	if snap.Children[0].Type != "text" || snap.Children[0].Text != "Hello" {
		t.Errorf("first child mismatch: %+v", snap.Children[0])
	}
	if snap.Children[1].Type != "box" {
		t.Errorf("second child type = %q, want box", snap.Children[1].Type)
	}
	if got, want := snap.Children[1].Props["flexDirection"], "column"; got != want {
		t.Errorf("inner flexDirection = %v, want %v", got, want)
	}
	if len(snap.Children[1].Children) != 1 || snap.Children[1].Children[0].Text != "World" {
		t.Errorf("inner children mismatch: %+v", snap.Children[1].Children)
	}
}

func TestInspect_DeepCopyIndependence(t *testing.T) {
	root := vdom.CreateElement("box", vdom.Props{"k": "v"}, vdom.CreateTextNode("x"))
	snap := Inspect(root)

	// Mutate the snapshot's props/children and confirm the source is unchanged.
	snap.Props["k"] = "changed"
	snap.Children = append(snap.Children, &Snapshot{Type: "text", Text: "added"})

	if root.Props["k"] != "v" {
		t.Errorf("source props mutated: %v", root.Props["k"])
	}
	if len(root.Children) != 1 {
		t.Errorf("source children mutated: len=%d", len(root.Children))
	}
}

func TestInspect_NilChildSkipped(t *testing.T) {
	root := vdom.CreateElement("box", nil)
	root.Children = append(root.Children, nil, vdom.CreateTextNode("only"))

	snap := Inspect(root)
	if len(snap.Children) != 1 {
		t.Fatalf("snap.Children len = %d, want 1", len(snap.Children))
	}
	if snap.Children[0].Text != "only" {
		t.Errorf("kept child = %+v", snap.Children[0])
	}
}

func TestInspect_EmptyElementType(t *testing.T) {
	node := &vdom.Node{Type: vdom.ElementNode}
	snap := Inspect(node)
	if snap.Type != "element" {
		t.Errorf("snap.Type = %q, want element", snap.Type)
	}
}

func TestInspect_LayoutPopulatedAndOmitted(t *testing.T) {
	root := vdom.CreateElement("box", nil)
	root.Layout = vdom.Layout{Left: 1, Top: 2, Width: 30, Height: 40}

	bare := vdom.CreateElement("box", nil)

	snap := Inspect(root)
	if snap.Layout == nil {
		t.Fatal("expected layout to be populated")
	}
	if *snap.Layout != (LayoutInfo{Left: 1, Top: 2, Width: 30, Height: 40}) {
		t.Errorf("unexpected layout: %+v", snap.Layout)
	}

	bareSnap := Inspect(bare)
	if bareSnap.Layout != nil {
		t.Errorf("expected nil layout for unlaid-out node, got %+v", bareSnap.Layout)
	}
}

func TestSnapshot_JSON_RoundTripProps(t *testing.T) {
	props := vdom.Props{
		"name":    "title",
		"count":   42,
		"enabled": true,
	}
	root := vdom.CreateElement("box", props, vdom.CreateTextNode("content"))

	snap := Inspect(root)
	raw := snap.JSON()
	if !strings.HasPrefix(raw, "{") {
		t.Fatalf("JSON output not an object: %s", raw)
	}

	var decoded Snapshot
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Type != "box" {
		t.Errorf("decoded.Type = %q", decoded.Type)
	}
	if decoded.Props["name"] != "title" {
		t.Errorf("decoded.Props[name] = %v", decoded.Props["name"])
	}
	// JSON numbers decode as float64.
	if got, ok := decoded.Props["count"].(float64); !ok || got != 42 {
		t.Errorf("decoded.Props[count] = %v (%T)", decoded.Props["count"], decoded.Props["count"])
	}
	if decoded.Props["enabled"] != true {
		t.Errorf("decoded.Props[enabled] = %v", decoded.Props["enabled"])
	}
	if len(decoded.Children) != 1 || decoded.Children[0].Text != "content" {
		t.Errorf("decoded children mismatch: %+v", decoded.Children)
	}
}

func TestSnapshot_JSON_NilSnapshot(t *testing.T) {
	var s *Snapshot
	if got := s.JSON(); got != "null" {
		t.Errorf("nil snapshot JSON = %q, want null", got)
	}
}

func TestSnapshot_JSON_MarshalError(t *testing.T) {
	s := &Snapshot{
		Type: "box",
		Props: map[string]interface{}{
			"bad": make(chan int), // not JSON-serializable
		},
	}
	out := s.JSON()
	if !strings.Contains(out, "error") {
		t.Errorf("expected error indicator in JSON output, got %q", out)
	}
}

func TestSnapshot_Tree_IndentationAndStructure(t *testing.T) {
	tree := vdom.CreateElement("box", nil,
		vdom.CreateTextNode("Hello"),
		vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			vdom.CreateTextNode("World"),
		),
	)

	snap := Inspect(tree)
	out := snap.Tree()

	want := []string{
		"<box>\n",
		"  <text>\"Hello\"</text>\n",
		"  <box flexDirection=column>\n",
		"    <text>\"World\"</text>\n",
		"  </box>\n",
		"</box>\n",
	}
	for _, line := range want {
		if !strings.Contains(out, line) {
			t.Errorf("Tree output missing %q\nfull output:\n%s", line, out)
		}
	}
}

func TestSnapshot_Tree_KeyShownAndSkipped(t *testing.T) {
	keyed := vdom.CreateElement("box", nil)
	keyed.Key = "main"
	keyless := vdom.CreateElement("box", nil)

	if got := Inspect(keyed).Tree(); !strings.Contains(got, `key="main"`) {
		t.Errorf("expected key in output, got %q", got)
	}
	if got := Inspect(keyless).Tree(); strings.Contains(got, "key=") {
		t.Errorf("did not expect key in output, got %q", got)
	}

	textKeyed := vdom.CreateTextNode("hi")
	textKeyed.Key = "t1"
	if got := Inspect(textKeyed).Tree(); !strings.Contains(got, `key="t1"`) {
		t.Errorf("expected key on text node, got %q", got)
	}
}

func TestSnapshot_Tree_SelfClosing(t *testing.T) {
	snap := Inspect(vdom.CreateElement("box", vdom.Props{"width": 5}))
	out := snap.Tree()
	if !strings.HasSuffix(strings.TrimSpace(out), "/>") {
		t.Errorf("expected self-closing tag, got %q", out)
	}
}

func TestSnapshot_Tree_NilSnapshot(t *testing.T) {
	var s *Snapshot
	if got := s.Tree(); got != "" {
		t.Errorf("nil tree = %q, want empty", got)
	}
}

func TestSnapshot_Tree_PropsSortedDeterministic(t *testing.T) {
	root := vdom.CreateElement("box", vdom.Props{
		"zeta":  1,
		"alpha": 2,
		"mid":   3,
	})
	out := Inspect(root).Tree()
	idxAlpha := strings.Index(out, "alpha=")
	idxMid := strings.Index(out, "mid=")
	idxZeta := strings.Index(out, "zeta=")
	if !(idxAlpha < idxMid && idxMid < idxZeta) {
		t.Errorf("props not sorted: %s", out)
	}
}

func TestPrintTree(t *testing.T) {
	root := vdom.CreateElement("box", nil, vdom.CreateTextNode("hi"))

	var buf bytes.Buffer
	if err := PrintTree(&buf, root); err != nil {
		t.Fatalf("PrintTree returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "<box>") || !strings.Contains(buf.String(), `"hi"`) {
		t.Errorf("PrintTree output unexpected: %s", buf.String())
	}
}

func TestPrintTree_NilWriter(t *testing.T) {
	if err := PrintTree(nil, vdom.CreateTextNode("x")); err == nil {
		t.Error("expected error for nil writer")
	}
}

type errWriter struct{}

func (errWriter) Write(_ []byte) (int, error) { return 0, errors.New("boom") }

func TestPrintTree_PropagatesWriterError(t *testing.T) {
	root := vdom.CreateElement("box", nil, vdom.CreateTextNode("hi"))
	if err := PrintTree(errWriter{}, root); err == nil {
		t.Error("expected writer error to propagate")
	}
}
