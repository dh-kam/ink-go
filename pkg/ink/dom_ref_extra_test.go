package ink_test

import (
	"testing"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TestDOMRefGetAttributeReturnsExposedValues exercises the DOM-like
// `getAttribute(key)` accessor against a mounted tree, verifying that the
// behaviour matches Attributes() for both visible and hidden keys.
func TestDOMRefGetAttributeReturnsExposedValues(t *testing.T) {
	var current *ink.DOMElement

	app := ink.NewApp(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ = ref.Current().(*ink.DOMElement)
			return nil
		}, []interface{}{"capture"})

		return vdom.CreateElement("box", vdom.Props{
			"data-id":                "ref-target",
			"data-count":             7,
			"width":                  float64(20),
			"ref":                    ref,
			"internal_accessibility": "hidden",
		},
			vdom.CreateTextNode("hi"),
		)
	})

	app.RenderOnce()
	if current == nil {
		t.Fatal("expected ref to capture mounted element")
	}

	if value, ok := current.GetAttribute("data-id"); !ok || value != "ref-target" {
		t.Fatalf("expected data-id %q, got %v (ok=%v)", "ref-target", value, ok)
	}
	if value, ok := current.GetAttribute("data-count"); !ok || value != 7 {
		t.Fatalf("expected data-count 7, got %v (ok=%v)", value, ok)
	}
	if _, ok := current.GetAttribute("width"); ok {
		t.Fatal("expected width style prop to be hidden from GetAttribute")
	}
	if _, ok := current.GetAttribute("ref"); ok {
		t.Fatal("expected ref prop to be hidden from GetAttribute")
	}
	if _, ok := current.GetAttribute("internal_accessibility"); ok {
		t.Fatal("expected internal_accessibility to be hidden from GetAttribute")
	}
	if _, ok := current.GetAttribute("missing"); ok {
		t.Fatal("expected GetAttribute on missing key to return ok=false")
	}
}

// TestDOMRefStyleExposesLayoutProps verifies that the ref handle exposes a
// style map analogous to upstream's `DOMElement.style`, returning layout
// props that Attributes() filters out.
func TestDOMRefStyleExposesLayoutProps(t *testing.T) {
	var current *ink.DOMElement

	app := ink.NewApp(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ = ref.Current().(*ink.DOMElement)
			return nil
		}, []interface{}{"capture"})

		return vdom.CreateElement("box", vdom.Props{
			"ref":           ref,
			"width":         float64(15),
			"height":        float64(3),
			"flexDirection": "row",
			"data-id":       "styled-box",
		}, vdom.CreateTextNode("x"))
	})

	app.RenderOnce()
	if current == nil {
		t.Fatal("expected ref to capture mounted element")
	}

	style := current.Style()
	if got := style["width"]; got != float64(15) {
		t.Fatalf("expected style.width 15, got %v", got)
	}
	if got := style["height"]; got != float64(3) {
		t.Fatalf("expected style.height 3, got %v", got)
	}
	if got := style["flexDirection"]; got != "row" {
		t.Fatalf("expected style.flexDirection 'row', got %v", got)
	}
	if _, ok := style["data-id"]; ok {
		t.Fatal("expected data attribute to be excluded from style map")
	}
	if _, ok := style["ref"]; ok {
		t.Fatal("expected ref to be excluded from style map")
	}
}

// TestDOMRefInternalStaticReportsStaticAncestors exercises the `internal_static`
// flag exposure on the ref handle.
func TestDOMRefInternalStaticReportsStaticAncestors(t *testing.T) {
	staticNode := vdom.CreateElement("static", nil, vdom.CreateTextNode("logged"))
	box := vdom.CreateElement("box", vdom.Props{"internal_static": true}, vdom.CreateTextNode("alt"))
	plain := vdom.CreateElement("box", nil, vdom.CreateTextNode("dynamic"))

	if !staticNode.InternalStatic() {
		t.Fatal("expected <static> element to report internal_static=true")
	}
	if !box.InternalStatic() {
		t.Fatal("expected explicit internal_static prop to be honoured")
	}
	if plain.InternalStatic() {
		t.Fatal("expected default box to report internal_static=false")
	}

	textChild := staticNode.FirstChild()
	if textChild == nil {
		t.Fatal("expected static node to have child")
	}
	if textChild.InternalStatic() {
		t.Fatal("expected text child not to report internal_static")
	}
}

// TestDOMRefChildrenSkipsTextNodes verifies the `Element.children` accessor
// returns only element children, mirroring the DOM split between childNodes
// and children.
func TestDOMRefChildrenSkipsTextNodes(t *testing.T) {
	var current *ink.DOMElement

	app := ink.NewApp(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ = ref.Current().(*ink.DOMElement)
			return nil
		}, []interface{}{"capture"})

		return vdom.CreateElement("box", vdom.Props{"ref": ref},
			vdom.CreateTextNode("plain"),
			vdom.CreateElement("text", nil, vdom.CreateTextNode("a")),
			vdom.CreateElement("text", nil, vdom.CreateTextNode("b")),
		)
	})

	app.RenderOnce()
	if current == nil {
		t.Fatal("expected ref to capture mounted element")
	}

	if got := len(current.ChildNodes()); got != 3 {
		t.Fatalf("expected 3 child nodes, got %d", got)
	}

	elements := current.ElementChildren()
	if got := len(elements); got != 2 {
		t.Fatalf("expected 2 element children, got %d", got)
	}
	if elements[0].NodeName() != "ink-text" || elements[1].NodeName() != "ink-text" {
		t.Fatalf("expected element children to be ink-text nodes, got %v", elements)
	}
}

// TestDOMRefOwnerRootWalksToInkRoot verifies the ref can traverse upward to
// the rendered tree's root, mirroring upstream's parent chain.
func TestDOMRefOwnerRootWalksToInkRoot(t *testing.T) {
	var current *ink.DOMElement

	app := ink.NewApp(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ = ref.Current().(*ink.DOMElement)
			return nil
		}, []interface{}{"capture"})

		return vdom.CreateElement("box", nil,
			vdom.CreateElement("box", nil,
				vdom.CreateElement("text", vdom.Props{"ref": ref}, vdom.CreateTextNode("deep")),
			),
		)
	})

	app.RenderOnce()
	if current == nil {
		t.Fatal("expected ref to capture mounted element")
	}

	root := current.OwnerRoot()
	if root == nil {
		t.Fatal("expected OwnerRoot to return a node")
	}
	if root == current {
		t.Fatal("expected OwnerRoot to return an ancestor when the ref is nested")
	}
	if root.ParentNode() != nil {
		t.Fatalf("expected OwnerRoot to be the topmost node, got parent %v", root.ParentNode())
	}
	if !root.Contains(current) {
		t.Fatal("expected OwnerRoot to contain the ref node")
	}
}

// TestDOMRefMeasureElementPositionReturnsLayoutCoords mirrors upstream's
// reliance on the Yoga node's getComputedTop/Left for ref-based measurement
// of nested boxes.
func TestDOMRefMeasureElementPositionReturnsLayoutCoords(t *testing.T) {
	var current *ink.DOMElement

	app := ink.NewAppWithOptions(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ = ref.Current().(*ink.DOMElement)
			return nil
		}, []interface{}{"capture"})

		return vdom.CreateElement("box", vdom.Props{"flexDirection": "column"},
			vdom.CreateElement("text", nil, vdom.CreateTextNode("first line")),
			vdom.CreateElement("box", vdom.Props{"ref": ref},
				vdom.CreateElement("text", nil, vdom.CreateTextNode("second")),
			),
		)
	}, ink.AppOptions{Width: 40, Height: 10})

	app.RenderOnce()
	if current == nil {
		t.Fatal("expected ref to capture mounted element")
	}

	pos := ink.MeasureElementPosition(current)
	if pos.Top == 0 {
		t.Fatalf("expected nested ref to have non-zero top, got %+v", pos)
	}

	// MeasureElementPosition on a nil ref should be the zero value.
	if got := ink.MeasureElementPosition(nil); got != (ink.ElementPosition{}) {
		t.Fatalf("expected zero ElementPosition for nil node, got %+v", got)
	}
}

// TestDOMRefMeasureElementBeforeFirstRender verifies MeasureElement returns
// {0,0} for a ref that has never been attached to a rendered tree, matching
// upstream's behaviour when `ref.current` is read before the first render.
func TestDOMRefMeasureElementBeforeFirstRender(t *testing.T) {
	var captured *ink.DOMElement

	app := ink.NewApp(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))
		// Read ref.Current synchronously - before render commit. Should be nil.
		if current, ok := ref.Current().(*ink.DOMElement); ok {
			captured = current
		}
		return vdom.CreateElement("box", vdom.Props{"ref": ref},
			vdom.CreateTextNode("hi"),
		)
	})

	// Capture before any layout has happened.
	if captured != nil {
		t.Fatalf("expected ref.Current to be nil before render, got %v", captured)
	}

	if got := ink.MeasureElement(captured); got != (ink.ElementDimensions{}) {
		t.Fatalf("expected pre-render measurement to be zero, got %+v", got)
	}

	// After rendering the ref should populate.
	app.RenderOnce()
}

// TestDOMRefMeasureElementInsideFixedWidthParent ports the upstream
// "constrained width" expectation: when a ref'd box is nested inside a parent
// with explicit width, the measured width reflects that constraint rather
// than the terminal viewport.
func TestDOMRefMeasureElementInsideFixedWidthParent(t *testing.T) {
	var current *ink.DOMElement

	app := ink.NewAppWithOptions(func() *vdom.Node {
		ref := ink.UseRef((*ink.DOMElement)(nil))

		ink.UseEffect(func() func() {
			current, _ = ref.Current().(*ink.DOMElement)
			return nil
		}, []interface{}{"capture"})

		return vdom.CreateElement("box", vdom.Props{"width": 20},
			vdom.CreateElement("box", vdom.Props{"ref": ref, "flexGrow": 1},
				vdom.CreateElement("text", nil, vdom.CreateTextNode("inside")),
			),
		)
	}, ink.AppOptions{Width: 100, Height: 24})

	app.RenderOnce()
	if current == nil {
		t.Fatal("expected ref to capture mounted element")
	}

	dims := ink.MeasureElement(current)
	if dims.Width <= 0 {
		t.Fatalf("expected nested ref to have non-zero width, got %+v", dims)
	}
	if dims.Width > 20 {
		t.Fatalf("expected nested ref width to be constrained by parent width 20, got %d", dims.Width)
	}
}
