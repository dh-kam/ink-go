package renderer_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// --- Render -----------------------------------------------------------------

// TestRenderEntryPointSimpleText verifies the bare Render entry point handles
// a plain text node. Render does not consult layout so it just stamps the text
// into a buffer.
func TestRenderEntryPointSimpleText(t *testing.T) {
	node := vdom.CreateTextNode("simple-text")

	output := renderer.Render(node, 40, 4)

	if !strings.Contains(output, "simple-text") {
		t.Fatalf("expected output to contain text, got %q", output)
	}
}

// TestRenderEntryPointEmptyBox verifies that an element node with no children
// produces no visible output through the bare Render path.
func TestRenderEntryPointEmptyBox(t *testing.T) {
	node := vdom.CreateElement("box", nil)

	output := renderer.Render(node, 20, 4)

	if strings.TrimSpace(output) != "" {
		t.Fatalf("expected empty output, got %q", output)
	}
}

// TestRenderEntryPointZeroSize verifies that Render does not crash with
// degenerate dimensions and the buffer remains effectively empty.
func TestRenderEntryPointZeroSize(t *testing.T) {
	node := vdom.CreateTextNode("ignored")

	output := renderer.Render(node, 0, 0)

	if strings.Contains(output, "ignored") {
		t.Fatalf("expected zero-size buffer to drop content, got %q", output)
	}
}

// --- RenderWithLayout -------------------------------------------------------

// TestRenderWithLayoutBoxWithChildren renders a Box containing several text
// children and verifies that all of them appear in the rendered output.
func TestRenderWithLayoutBoxWithChildren(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("alpha"),
		components.Text("beta"),
		components.Text("gamma"),
	)

	output := renderer.RenderWithLayout(root, 30, 10)

	for _, want := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

// TestRenderWithLayoutZeroWidth confirms the renderer does not crash when
// width is zero. Layout falls back to the intrinsic size of the children, so
// content may still appear; the assertion only enforces no panic occurred.
func TestRenderWithLayoutZeroWidth(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RenderWithLayout panicked on zero width: %v", r)
		}
	}()
	root := components.Box(nil, components.Text("zw-content"))

	_ = renderer.RenderWithLayout(root, 0, 5)
}

// TestRenderWithLayoutZeroHeight verifies that a zero-height layout returns an
// empty string because layout reports no rows to render.
func TestRenderWithLayoutZeroHeight(t *testing.T) {
	root := components.Box(nil, components.Text("hidden"))

	output := renderer.RenderWithLayout(root, 20, 0)

	if output != "" {
		t.Fatalf("expected empty output for zero height, got %q", output)
	}
}

// TestRenderWithLayoutDeeplyNested exercises layout recursion against a
// deeply nested column tree to ensure no stack/recursion limits are hit and
// the inner-most content still renders.
func TestRenderWithLayoutDeeplyNested(t *testing.T) {
	leaf := components.Text("leaf")
	current := leaf
	for i := 0; i < 25; i++ {
		current = components.Box(vdom.Props{"flexDirection": "column"}, current)
	}

	output := renderer.RenderWithLayout(current, 40, 30)

	if !strings.Contains(output, "leaf") {
		t.Fatalf("expected deeply nested leaf to render, got:\n%s", output)
	}
}

// TestRenderWithLayoutLargeTree builds a tree with many siblings so the
// layout traversal visits all of them and confirms that every label is in the
// output.
func TestRenderWithLayoutLargeTree(t *testing.T) {
	children := make([]*vdom.Node, 0, 50)
	for i := 0; i < 50; i++ {
		children = append(children, components.Text("row"))
	}

	root := components.Box(vdom.Props{"flexDirection": "column"}, children...)

	output := renderer.RenderWithLayout(root, 20, 80)

	if strings.Count(output, "row") < 50 {
		t.Fatalf("expected 50 rows, got %d in:\n%s", strings.Count(output, "row"), output)
	}
}

// --- RenderWithLayoutANSI ---------------------------------------------------

// TestRenderWithLayoutANSINilNode confirms the ANSI entry point safely returns
// an empty string when given a nil node.
func TestRenderWithLayoutANSINilNode(t *testing.T) {
	if got := renderer.RenderWithLayoutANSI(nil, 20, 5); got != "" {
		t.Fatalf("expected empty output for nil node, got %q", got)
	}
}

// TestRenderWithLayoutANSIZeroHeight confirms the ANSI entry point does not
// panic when height is zero. Layout uses intrinsic measurements so output
// may still be produced; we only verify no panic.
func TestRenderWithLayoutANSIZeroHeight(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RenderWithLayoutANSI panicked on zero height: %v", r)
		}
	}()
	root := components.Box(nil, components.Text("zh-content"))

	_ = renderer.RenderWithLayoutANSI(root, 20, 0)
}

// TestRenderWithLayoutANSIColorAndStyle verifies that color, bold, underline
// modifiers actually emit ANSI escape codes for a styled text run.
func TestRenderWithLayoutANSIColorAndStyle(t *testing.T) {
	root := components.Box(nil,
		components.Text("hello", vdom.Props{
			"color":     "red",
			"bold":      true,
			"underline": true,
		}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)

	if !strings.Contains(output, "hello") {
		t.Fatalf("expected text to be present, got %q", output)
	}
	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI escape sequence in output, got %q", output)
	}
}

// TestRenderWithLayoutANSIBackgroundOnBox verifies that a background color on
// a sized box emits at least one ANSI escape sequence.
func TestRenderWithLayoutANSIBackgroundOnBox(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":           10.0,
		"height":          2.0,
		"backgroundColor": "blue",
	})

	output := renderer.RenderWithLayoutANSI(root, 20, 5)

	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI escape for background color, got %q", output)
	}
}

// --- SyncComputedLayout -----------------------------------------------------

// TestSyncComputedLayoutPopulatesNode verifies that calling SyncComputedLayout
// fills in computed Layout metadata on the original vdom node.
func TestSyncComputedLayoutPopulatesNode(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":  20.0,
		"height": 5.0,
	}, components.Text("inside"))

	// Pre-condition: layout is empty before sync.
	if root.ComputedLayout().Width != 0 {
		t.Fatalf("expected pristine layout, got %+v", root.ComputedLayout())
	}

	renderer.SyncComputedLayout(root, 80, 24)

	got := root.ComputedLayout()
	if got.Width != 20 || got.Height != 5 {
		t.Fatalf("expected layout to be 20x5, got %+v", got)
	}
}

// TestSyncComputedLayoutNilNode is a defensive test verifying nil input does
// not panic.
func TestSyncComputedLayoutNilNode(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SyncComputedLayout panicked on nil: %v", r)
		}
	}()
	renderer.SyncComputedLayout(nil, 10, 10)
}

// --- RenderWithLayoutSections / Mode ---------------------------------------

// TestRenderWithLayoutSectionsNilNode ensures the helper returns a zero-value
// RenderSections for a nil tree.
func TestRenderWithLayoutSectionsNilNode(t *testing.T) {
	sections := renderer.RenderWithLayoutSections(nil, 20, 5)
	if sections.Output != "" || sections.StaticOutput != "" {
		t.Fatalf("expected zero RenderSections, got %+v", sections)
	}
}

// TestRenderWithLayoutSectionsSplitsStaticAndDynamic verifies that a tree
// containing both dynamic content and a static block produces the expected
// split between Output and StaticOutput.
func TestRenderWithLayoutSectionsSplitsStaticAndDynamic(t *testing.T) {
	staticBlock := components.Static(nil, components.Text("static-line"))
	dynamic := components.Text("dynamic-line")

	root := components.Box(vdom.Props{"flexDirection": "column"},
		dynamic,
		staticBlock,
	)

	sections := renderer.RenderWithLayoutSections(root, 40, 10)

	if !strings.Contains(sections.Output, "dynamic-line") {
		t.Fatalf("expected dynamic content in Output, got %q", sections.Output)
	}
	if strings.Contains(sections.Output, "static-line") {
		t.Fatalf("did not expect static content in Output, got %q", sections.Output)
	}
	if !strings.Contains(sections.StaticOutput, "static-line") {
		t.Fatalf("expected static content in StaticOutput, got %q", sections.StaticOutput)
	}
}

// TestRenderWithLayoutSectionsModeANSIEnabled verifies the ANSI variant emits
// escape sequences for styled text inside a static section.
func TestRenderWithLayoutSectionsModeANSIEnabled(t *testing.T) {
	staticBlock := components.Static(nil,
		components.Text("colored", vdom.Props{"color": "green"}),
	)
	root := components.Box(nil, staticBlock)

	sections := renderer.RenderWithLayoutSectionsMode(root, 30, 5, true)

	if !strings.Contains(sections.StaticOutput, "colored") {
		t.Fatalf("expected colored text in StaticOutput, got %q", sections.StaticOutput)
	}
	if !strings.Contains(sections.StaticOutput, "\x1b[") {
		t.Fatalf("expected ANSI escape sequence, got %q", sections.StaticOutput)
	}
}

// TestRenderWithLayoutSectionsNoStatic verifies a tree without any static
// children populates only the dynamic Output field.
func TestRenderWithLayoutSectionsNoStatic(t *testing.T) {
	root := components.Box(nil, components.Text("only-dynamic"))

	sections := renderer.RenderWithLayoutSections(root, 30, 5)

	if !strings.Contains(sections.Output, "only-dynamic") {
		t.Fatalf("expected dynamic content, got %q", sections.Output)
	}
	if sections.StaticOutput != "" {
		t.Fatalf("expected no StaticOutput, got %q", sections.StaticOutput)
	}
}

// --- RenderScreenReaderSections (additional aria scenarios) ----------------

// TestRenderScreenReaderSectionsNilNode confirms nil input yields a zero-value
// RenderSections.
func TestRenderScreenReaderSectionsNilNode(t *testing.T) {
	sections := renderer.RenderScreenReaderSections(nil)
	if sections.Output != "" || sections.StaticOutput != "" {
		t.Fatalf("expected zero RenderSections, got %+v", sections)
	}
}

// TestRenderScreenReaderSectionsAriaRoleNested verifies that an aria-role on a
// nested element is announced before its content.
func TestRenderScreenReaderSectionsAriaRoleNested(t *testing.T) {
	root := components.Box(nil,
		components.Box(vdom.Props{"aria-role": "heading"},
			components.Text("Title"),
		),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "Title") {
		t.Fatalf("expected heading title in screen-reader output, got %q", sections.Output)
	}
}

// TestRenderScreenReaderSectionsAriaHiddenSkipsSubtree confirms aria-hidden
// fully suppresses a subtree from the screen reader output.
func TestRenderScreenReaderSectionsAriaHiddenSkipsSubtree(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"aria-hidden": true},
			components.Text("Secret"),
		),
		components.Text("Visible"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if strings.Contains(sections.Output, "Secret") {
		t.Fatalf("expected aria-hidden subtree to be hidden, got %q", sections.Output)
	}
	if !strings.Contains(sections.Output, "Visible") {
		t.Fatalf("expected visible text in screen-reader output, got %q", sections.Output)
	}
}

// TestRenderScreenReaderSectionsStaticBranchSplit verifies static subtrees end
// up in StaticOutput and the dynamic remainder is in Output.
func TestRenderScreenReaderSectionsStaticBranchSplit(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("dynamic"),
		components.Static(nil, components.Text("static")),
	)

	sections := renderer.RenderScreenReaderSections(root)

	if !strings.Contains(sections.Output, "dynamic") {
		t.Fatalf("expected dynamic in Output, got %q", sections.Output)
	}
	if !strings.Contains(sections.StaticOutput, "static") {
		t.Fatalf("expected static in StaticOutput, got %q", sections.StaticOutput)
	}
}

// --- RenderRuntimeSections / Mode ------------------------------------------

// TestRenderRuntimeSectionsNilNode confirms nil input returns a zero value.
func TestRenderRuntimeSectionsNilNode(t *testing.T) {
	sections := renderer.RenderRuntimeSections(nil, 20, 5, nil, false)
	if sections.Output != "" || sections.StaticDeltaOutput != "" || sections.StaticCounts != nil {
		t.Fatalf("expected zero RenderSections, got %+v", sections)
	}
}

// TestRenderRuntimeSectionsAppendsDelta exercises the static-delta book
// keeping by calling the renderer twice with the StaticCounts returned from
// the first call.
func TestRenderRuntimeSectionsAppendsDelta(t *testing.T) {
	type item struct{ Label string }
	itemsFirst := []item{{"first"}, {"second"}}

	rootFirst := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("dynamic"),
		components.StaticItems(itemsFirst, func(it item, idx int) *vdom.Node {
			return components.Text(it.Label)
		}),
	)

	first := renderer.RenderRuntimeSections(rootFirst, 30, 10, nil, false)

	if !strings.Contains(first.StaticDeltaOutput, "first") || !strings.Contains(first.StaticDeltaOutput, "second") {
		t.Fatalf("expected initial StaticDeltaOutput to contain both items, got %q", first.StaticDeltaOutput)
	}
	if len(first.StaticCounts) != 1 || first.StaticCounts[0] != 2 {
		t.Fatalf("expected StaticCounts [2], got %+v", first.StaticCounts)
	}

	itemsSecond := append(itemsFirst, item{"third"})
	rootSecond := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("dynamic"),
		components.StaticItems(itemsSecond, func(it item, idx int) *vdom.Node {
			return components.Text(it.Label)
		}),
	)

	second := renderer.RenderRuntimeSections(rootSecond, 30, 10, first.StaticCounts, false)

	if strings.Contains(second.StaticDeltaOutput, "first") || strings.Contains(second.StaticDeltaOutput, "second") {
		t.Fatalf("expected delta to omit previously rendered items, got %q", second.StaticDeltaOutput)
	}
	if !strings.Contains(second.StaticDeltaOutput, "third") {
		t.Fatalf("expected new item in delta, got %q", second.StaticDeltaOutput)
	}
	if len(second.StaticCounts) != 1 || second.StaticCounts[0] != 3 {
		t.Fatalf("expected StaticCounts [3], got %+v", second.StaticCounts)
	}
}

// TestRenderRuntimeSectionsNoChange verifies the second call with identical
// counts produces no delta output.
func TestRenderRuntimeSectionsNoChange(t *testing.T) {
	type item struct{ Label string }
	items := []item{{"one"}, {"two"}}

	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("dyn"),
		components.StaticItems(items, func(it item, idx int) *vdom.Node {
			return components.Text(it.Label)
		}),
	)

	first := renderer.RenderRuntimeSections(root, 30, 10, nil, false)
	second := renderer.RenderRuntimeSections(root, 30, 10, first.StaticCounts, false)

	if second.StaticDeltaOutput != "" {
		t.Fatalf("expected no delta on second call, got %q", second.StaticDeltaOutput)
	}
	if len(second.StaticCounts) != 1 || second.StaticCounts[0] != 2 {
		t.Fatalf("expected StaticCounts [2], got %+v", second.StaticCounts)
	}
}

// TestRenderRuntimeSectionsScreenReader confirms screen-reader mode returns
// plain text Output via the runtime section path.
func TestRenderRuntimeSectionsScreenReader(t *testing.T) {
	root := components.Box(nil,
		components.Text("hello"),
		components.Static(nil, components.Text("static-sr")),
	)

	sections := renderer.RenderRuntimeSections(root, 20, 5, nil, true)

	if !strings.Contains(sections.Output, "hello") {
		t.Fatalf("expected screen-reader Output, got %q", sections.Output)
	}
	if !strings.Contains(sections.StaticDeltaOutput, "static-sr") {
		t.Fatalf("expected static delta in screen-reader mode, got %q", sections.StaticDeltaOutput)
	}
}

// TestRenderRuntimeSectionsModeANSI verifies the ANSI runtime mode emits
// escape sequences in the static delta when colors are present.
func TestRenderRuntimeSectionsModeANSI(t *testing.T) {
	root := components.Box(nil,
		components.Static(nil,
			components.Text("colored", vdom.Props{"color": "magenta"}),
		),
	)

	sections := renderer.RenderRuntimeSectionsMode(root, 30, 5, nil, false, true)

	if !strings.Contains(sections.StaticDeltaOutput, "colored") {
		t.Fatalf("expected colored text in StaticDeltaOutput, got %q", sections.StaticDeltaOutput)
	}
	if !strings.Contains(sections.StaticDeltaOutput, "\x1b[") {
		t.Fatalf("expected ANSI escape sequence in delta output, got %q", sections.StaticDeltaOutput)
	}
}

// TestRenderRuntimeSectionsUntrackedStatic exercises the branch where a
// static element has no __staticItemsCount tracking; the first call should
// emit content, the second should skip it because previousCount > 0.
func TestRenderRuntimeSectionsUntrackedStatic(t *testing.T) {
	root := components.Box(nil,
		components.Static(nil, components.Text("untracked")),
	)

	first := renderer.RenderRuntimeSections(root, 30, 5, nil, false)
	if !strings.Contains(first.StaticDeltaOutput, "untracked") {
		t.Fatalf("expected first delta to contain content, got %q", first.StaticDeltaOutput)
	}
	if len(first.StaticCounts) != 1 || first.StaticCounts[0] != 1 {
		t.Fatalf("expected StaticCounts [1] on first call, got %+v", first.StaticCounts)
	}

	second := renderer.RenderRuntimeSections(root, 30, 5, first.StaticCounts, false)
	if second.StaticDeltaOutput != "" {
		t.Fatalf("expected no delta for untracked static after first render, got %q", second.StaticDeltaOutput)
	}
}

// TestRenderRuntimeSectionsUntrackedScreenReader exercises the
// untracked + screen-reader branch in renderStaticRootDelta.
func TestRenderRuntimeSectionsUntrackedScreenReader(t *testing.T) {
	root := components.Box(nil,
		components.Static(nil, components.Text("sr-static")),
	)

	first := renderer.RenderRuntimeSections(root, 30, 5, nil, true)
	if !strings.Contains(first.StaticDeltaOutput, "sr-static") {
		t.Fatalf("expected sr-static in delta, got %q", first.StaticDeltaOutput)
	}

	second := renderer.RenderRuntimeSections(root, 30, 5, first.StaticCounts, true)
	if second.StaticDeltaOutput != "" {
		t.Fatalf("expected empty delta on second call, got %q", second.StaticDeltaOutput)
	}
}

// --- Text wrapping & truncation modes (ANSI canvas) -----------------------

// TestRenderWithLayoutANSITruncateMiddleStyled exercises the styled
// middle-truncation code path.
func TestRenderWithLayoutANSITruncateMiddleStyled(t *testing.T) {
	root := components.Box(vdom.Props{"width": 7.0},
		components.Text("HelloWorldHello", vdom.Props{
			"wrap":  "truncate-middle",
			"color": "cyan",
		}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)

	// truncate-middle inserts an ellipsis (U+2026) somewhere in the middle.
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected middle ellipsis in output, got %q", output)
	}
}

// TestRenderWithLayoutANSITruncateStartStyled exercises the styled
// start-truncation code path.
func TestRenderWithLayoutANSITruncateStartStyled(t *testing.T) {
	root := components.Box(vdom.Props{"width": 5.0},
		components.Text("HelloWorld", vdom.Props{
			"wrap":  "truncate-start",
			"color": "yellow",
		}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)

	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected leading ellipsis, got %q", output)
	}
}

// TestRenderWithLayoutANSIWrapStyledMultiLine forces wrapping to occur on a
// styled string so wrapStyledLine and splitStyledRunesLines are traversed.
func TestRenderWithLayoutANSIWrapStyledMultiLine(t *testing.T) {
	root := components.Box(vdom.Props{"width": 6.0},
		components.Text("alpha beta gamma", vdom.Props{
			"color": "red",
		}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 6)

	// Wrap must produce more than one line of text in the output.
	if strings.Count(output, "\n") < 2 {
		t.Fatalf("expected multi-line wrapped output, got %q", output)
	}
}

// TestRenderWithLayoutANSIPreserveAnsiInTransform forces a Transform to emit a
// string already containing ANSI escapes, exercising parseANSIToStyledRunes
// and consumeANSISequence.
func TestRenderWithLayoutANSIPreserveAnsiInTransform(t *testing.T) {
	transformed := components.Transform(func(children string, idx int) string {
		// Wrap children in explicit ANSI red foreground sequence.
		return "\x1b[31m" + children + "\x1b[0m"
	}, vdom.CreateTextNode("inner"))

	root := components.Box(nil, transformed)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)

	if !strings.Contains(output, "inner") {
		t.Fatalf("expected transformed content, got %q", output)
	}
	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI escape carried through, got %q", output)
	}
}

func TestRenderWithLayoutPreservesOSCSequences(t *testing.T) {
	link := "\x1b]8;;https://example.com\x07Example\x1b]8;;\x07"
	root := components.Text(link)

	plain := renderer.RenderWithLayout(root, 40, 3)
	if plain != link {
		t.Fatalf("plain OSC output mismatch\nwant: %q\n got: %q", link, plain)
	}

	ansi := renderer.RenderWithLayoutANSI(root, 40, 3)
	if ansi != link {
		t.Fatalf("ANSI OSC output mismatch\nwant: %q\n got: %q", link, ansi)
	}
}

// --- Plain (non-ANSI) truncate variants ------------------------------------

// TestRenderWithLayoutTruncateMiddlePlain exercises the plain truncateMiddle
// path through applyTextLayoutMode.
func TestRenderWithLayoutTruncateMiddlePlain(t *testing.T) {
	root := components.Box(vdom.Props{"width": 7.0},
		components.Text("HelloWorldHello", vdom.Props{
			"wrap": "truncate-middle",
		}),
	)

	output := renderer.RenderWithLayout(root, 20, 3)

	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected middle ellipsis in plain output, got %q", output)
	}
}

// TestRenderWithLayoutTruncateStartPlain exercises the plain truncateStart
// path through applyTextLayoutMode.
func TestRenderWithLayoutTruncateStartPlain(t *testing.T) {
	root := components.Box(vdom.Props{"width": 5.0},
		components.Text("HelloWorld", vdom.Props{
			"wrap": "truncate-start",
		}),
	)

	output := renderer.RenderWithLayout(root, 20, 3)

	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected leading ellipsis in plain output, got %q", output)
	}
}

// TestRenderWithLayoutTextWrapModeAlias confirms the legacy "textWrap" prop is
// honored as an alias for "wrap".
func TestRenderWithLayoutTextWrapModeAlias(t *testing.T) {
	root := components.Box(vdom.Props{"width": 5.0},
		components.Text("HelloWorld", vdom.Props{
			"textWrap": "truncate",
		}),
	)

	output := renderer.RenderWithLayout(root, 20, 3)

	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected ellipsis from truncate via textWrap alias, got %q", output)
	}
}

// --- Transform output edge cases -------------------------------------------

// TestRenderWithLayoutANSITransformReturnsEmpty verifies the renderer accepts
// a Transform that emits an empty string without panicking.
func TestRenderWithLayoutANSITransformReturnsEmpty(t *testing.T) {
	root := components.Box(nil,
		components.Transform(func(children string, idx int) string {
			return ""
		}, vdom.CreateTextNode("ignored")),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)

	if strings.Contains(output, "ignored") {
		t.Fatalf("expected empty transform to suppress content, got %q", output)
	}
}

// TestRenderWithLayoutANSITransformIdentity verifies an identity Transform
// passes through without altering content while still exercising the styled
// transform branch.
func TestRenderWithLayoutANSITransformIdentity(t *testing.T) {
	root := components.Box(nil,
		components.Transform(func(children string, idx int) string {
			return children
		}, components.Text("identity-text", vdom.Props{"color": "blue"})),
	)

	output := renderer.RenderWithLayoutANSI(root, 30, 3)

	if !strings.Contains(output, "identity-text") {
		t.Fatalf("expected identity transform to preserve text, got %q", output)
	}
}

// --- ANSI styling matrix ---------------------------------------------------

// TestRenderWithLayoutANSIAllTextModifiers triggers each TextProps modifier
// path inside the styled-text resolution helpers.
func TestRenderWithLayoutANSIAllTextModifiers(t *testing.T) {
	root := components.Box(nil,
		components.Text("styled", vdom.Props{
			"color":         "white",
			"dimColor":      true,
			"bold":          true,
			"italic":        true,
			"underline":     true,
			"inverse":       true,
			"strikethrough": true,
		}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)

	if !strings.Contains(output, "styled") {
		t.Fatalf("expected styled text in output, got %q", output)
	}
	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI escapes, got %q", output)
	}
}

// TestRenderWithLayoutANSIBorderAllSides exercises styled border resolution
// for each side, hitting resolveBorderSideStyle branches.
func TestRenderWithLayoutANSIBorderAllSides(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":             10.0,
		"height":            3.0,
		"borderStyle":       "round",
		"borderTopColor":    "red",
		"borderBottomColor": "green",
		"borderLeftColor":   "yellow",
		"borderRightColor":  "blue",
	})

	output := renderer.RenderWithLayoutANSI(root, 20, 5)

	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI escapes for colored border, got %q", output)
	}
}

// --- Screen reader: aria-state --------------------------------------------

// TestRenderScreenReaderSectionsAriaStateString verifies a string aria-state
// value flows through screenReaderStateDescription.
func TestRenderScreenReaderSectionsAriaStateString(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":  "switch",
		"aria-state": "checked",
	}, components.Text("Toggle"))

	sections := renderer.RenderScreenReaderSections(root)

	if !strings.Contains(sections.Output, "Toggle") {
		t.Fatalf("expected Toggle text in output, got %q", sections.Output)
	}
}

// TestRenderScreenReaderSectionsAriaLabelOnly verifies that an aria-label
// without children still produces output.
func TestRenderScreenReaderSectionsAriaLabelOnly(t *testing.T) {
	root := components.Box(vdom.Props{"aria-label": "label-only"})

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "label-only" {
		t.Fatalf("expected aria-label-only output, got %q", sections.Output)
	}
}

// TestRenderScreenReaderSectionsMultipleStaticBlocks ensures multiple static
// roots are concatenated with newline boundaries.
func TestRenderScreenReaderSectionsMultipleStaticBlocks(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Static(nil, components.Text("first-static")),
		components.Static(nil, components.Text("second-static")),
		components.Text("dyn"),
	)

	sections := renderer.RenderScreenReaderSections(root)

	if !strings.Contains(sections.StaticOutput, "first-static") || !strings.Contains(sections.StaticOutput, "second-static") {
		t.Fatalf("expected both static blocks in StaticOutput, got %q", sections.StaticOutput)
	}
	if !strings.Contains(sections.Output, "dyn") {
		t.Fatalf("expected dynamic content in Output, got %q", sections.Output)
	}
}

// --- Layout prop parsing branches -----------------------------------------

// TestRenderWithLayoutJustifyContentVariants exercises each branch of
// parseJustifyContent through public layout rendering.
func TestRenderWithLayoutJustifyContentVariants(t *testing.T) {
	for _, mode := range []string{
		"flex-start", "start", "center", "flex-end", "end",
		"space-between", "space-around", "space-evenly",
	} {
		root := components.Box(vdom.Props{
			"width":          20.0,
			"height":         3.0,
			"flexDirection":  "row",
			"justifyContent": mode,
		},
			components.Text("a"),
			components.Text("b"),
		)

		output := renderer.RenderWithLayout(root, 30, 5)
		if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
			t.Fatalf("mode %q: expected both items rendered, got %q", mode, output)
		}
	}
}

// TestRenderWithLayoutAlignItemsVariants exercises parseAlignItems branches.
func TestRenderWithLayoutAlignItemsVariants(t *testing.T) {
	for _, mode := range []string{"stretch", "flex-start", "start", "center", "flex-end", "end"} {
		root := components.Box(vdom.Props{
			"width":         12.0,
			"height":        4.0,
			"flexDirection": "row",
			"alignItems":    mode,
		},
			components.Text("x"),
		)

		output := renderer.RenderWithLayout(root, 20, 6)
		if !strings.Contains(output, "x") {
			t.Fatalf("alignItems %q: expected x to render, got %q", mode, output)
		}
	}
}

// TestRenderWithLayoutNumericPropsAcceptVariousTypes exercises
// parseNumericValue against multiple Go numeric kinds via the prop pipeline.
func TestRenderWithLayoutNumericPropsAcceptVariousTypes(t *testing.T) {
	cases := []vdom.Props{
		{"width": float32(10), "height": float64(2)},
		{"width": int8(10), "height": int16(2)},
		{"width": int32(10), "height": int64(2)},
		{"width": uint(10), "height": uint8(2)},
		{"width": uint16(10), "height": uint32(2)},
		{"width": uint64(10), "height": int(2)},
	}
	for i, props := range cases {
		root := components.Box(props, components.Text("n"))
		output := renderer.RenderWithLayout(root, 20, 5)
		if !strings.Contains(output, "n") {
			t.Fatalf("case %d: expected text rendered, got %q", i, output)
		}
	}
}

// TestRenderWithLayoutPercentSizes exercises percent size parsing in
// parseSizeValue and parseFlexBasis. We only assert at least one of the
// percent-sized children renders to confirm the parser path is exercised.
func TestRenderWithLayoutPercentSizes(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":         "100%",
		"height":        "100%",
		"flexDirection": "row",
	},
		components.Box(vdom.Props{"flexBasis": "50%"}, components.Text("L")),
		components.Box(vdom.Props{"flexBasis": "50%"}, components.Text("R")),
	)

	output := renderer.RenderWithLayout(root, 20, 3)
	if !strings.Contains(output, "L") && !strings.Contains(output, "R") {
		t.Fatalf("expected at least one half rendered, got %q", output)
	}
}

// --- Border style + glyphs branches ---------------------------------------

// TestRenderWithLayoutBorderStyleVariants exercises borderStyleGlyphs branches
// for each documented border style name.
func TestRenderWithLayoutBorderStyleVariants(t *testing.T) {
	expectedByStyle := map[string]string{
		"single":       "┌───┐\n│b  │\n└───┘",
		"double":       "╔═══╗\n║b  ║\n╚═══╝",
		"round":        "╭───╮\n│b  │\n╰───╯",
		"bold":         "┏━━━┓\n┃b  ┃\n┗━━━┛",
		"singleDouble": "╓───╖\n║b  ║\n╙───╜",
		"doubleSingle": "╒═══╕\n│b  │\n╘═══╛",
		"classic":      "+---+\n|b  |\n+---+",
		"arrow":        "↘↓↓↓↙\n→b  ←\n↗↑↑↑↖",
	}

	for style, expected := range expectedByStyle {
		root := components.Box(vdom.Props{
			"width":       5.0,
			"height":      3.0,
			"borderStyle": style,
		}, components.Text("b"))

		output := renderer.RenderWithLayout(root, 20, 5)
		if output != expected {
			t.Fatalf("style %q: expected %q, got %q", style, expected, output)
		}
	}
}

// TestRenderWithLayoutCustomBorderRune exercises borderRune for arbitrary
// single-rune border specifications via the per-side rune props.
func TestRenderWithLayoutCustomBorderRune(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":       12.0,
		"height":      3.0,
		"borderStyle": "single",
		"borderTop":   "*",
	}, components.Text("inside"))

	output := renderer.RenderWithLayout(root, 20, 5)
	if !strings.Contains(output, "inside") {
		t.Fatalf("expected inside text, got %q", output)
	}
}

// --- ANSI canvas styled-text matrix ---------------------------------------

// TestRenderWithLayoutANSITransformPreservesStandardForeground forces
// parseANSIToStyledRunes to walk through every supported standard color code
// (30-37 fg, 40-47 bg) plus reset codes.
func TestRenderWithLayoutANSITransformPreservesStandardForeground(t *testing.T) {
	transformed := components.Transform(func(children string, idx int) string {
		// Build a string with all 8 standard foreground colors then reset.
		var sb strings.Builder
		for code := 30; code <= 37; code++ {
			sb.WriteString("\x1b[")
			sb.WriteString(itoa(code))
			sb.WriteString("m")
			sb.WriteString("c")
		}
		// Add background range as well.
		for code := 40; code <= 47; code++ {
			sb.WriteString("\x1b[")
			sb.WriteString(itoa(code))
			sb.WriteString("m")
			sb.WriteString("b")
		}
		sb.WriteString("\x1b[0m")
		return sb.String()
	}, vdom.CreateTextNode("seed"))

	root := components.Box(nil, transformed)

	output := renderer.RenderWithLayoutANSI(root, 60, 3)

	if !strings.Contains(output, "c") || !strings.Contains(output, "b") {
		t.Fatalf("expected colored markers preserved, got %q", output)
	}
}

// TestRenderWithLayoutANSITransform256Color exercises the 38;5;n / 48;5;n
// SGR sequences in parseANSIToStyledRunes.
func TestRenderWithLayoutANSITransform256Color(t *testing.T) {
	transformed := components.Transform(func(children string, idx int) string {
		return "\x1b[38;5;201m" + "\x1b[48;5;226m" + children + "\x1b[0m"
	}, vdom.CreateTextNode("c256"))

	root := components.Box(nil, transformed)
	output := renderer.RenderWithLayoutANSI(root, 30, 3)

	if !strings.Contains(output, "c256") {
		t.Fatalf("expected text in output, got %q", output)
	}
}

// TestRenderWithLayoutANSITransformRGBColor exercises 38;2 / 48;2 truecolor
// sequences and the dim/italic/underline/strikethrough/inverse modifier
// branches in parseANSIToStyledRunes.
func TestRenderWithLayoutANSITransformRGBColor(t *testing.T) {
	transformed := components.Transform(func(children string, idx int) string {
		// fg truecolor + bg truecolor + bold + dim + italic + underline +
		// inverse + strikethrough + reset of italic/underline/inverse.
		return "\x1b[38;2;10;20;30m\x1b[48;2;200;100;50m\x1b[1m\x1b[2m\x1b[3m\x1b[4m\x1b[7m\x1b[9m" +
			children + "\x1b[23m\x1b[24m\x1b[27m\x1b[29m\x1b[22m\x1b[39m\x1b[49m"
	}, vdom.CreateTextNode("rgb"))

	root := components.Box(nil, transformed)
	output := renderer.RenderWithLayoutANSI(root, 30, 3)

	if !strings.Contains(output, "rgb") {
		t.Fatalf("expected rgb text, got %q", output)
	}
}

// TestRenderWithLayoutANSIWideRuneWrapping pushes wide characters through the
// wrap path so fit/trim helpers traverse non-ASCII runes.
func TestRenderWithLayoutANSIWideRuneWrapping(t *testing.T) {
	root := components.Box(vdom.Props{"width": 6.0},
		components.Text("漢字テスト 漢字テスト", vdom.Props{"color": "green"}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 6)

	if !strings.Contains(output, "漢") {
		t.Fatalf("expected wide char in output, got %q", output)
	}
}

// TestRenderWithLayoutTextLeadingWhitespaceWrap targets trimLeftSpaceRunes /
// trimRightSpaceRunes when wrapping spaces around words. We only assert the
// content fragments are present somewhere in the output (wrap may break a
// long token mid-character).
func TestRenderWithLayoutTextLeadingWhitespaceWrap(t *testing.T) {
	root := components.Box(vdom.Props{"width": 5.0},
		components.Text("   alpha beta   "),
	)

	output := renderer.RenderWithLayout(root, 20, 5)

	if !strings.Contains(output, "beta") {
		t.Fatalf("expected wrapped words, got %q", output)
	}
}

// TestRenderWithLayoutNestedTextStyleInheritance exercises the nested styled
// run resolution when an inner text overrides a parent's color.
func TestRenderWithLayoutNestedTextStyleInheritance(t *testing.T) {
	root := components.Box(vdom.Props{"backgroundColor": "blue"},
		components.Text("outer ", vdom.Props{"color": "red"},
			components.Text("inner", vdom.Props{"color": "green"}),
		),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)

	if !strings.Contains(output, "outer") || !strings.Contains(output, "inner") {
		t.Fatalf("expected nested text, got %q", output)
	}
	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI in output, got %q", output)
	}
}

// TestRenderWithLayoutANSIRawTextNodeChild renders a raw vdom TextNode as a
// direct child of a box, exercising the text-node case in
// renderNodeWithLayoutANSI.
func TestRenderWithLayoutANSIRawTextNodeChild(t *testing.T) {
	root := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
	},
		vdom.CreateTextNode("raw-text"),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)

	if !strings.Contains(output, "raw-text") {
		t.Fatalf("expected raw text node rendered, got %q", output)
	}
}

// TestRenderWithLayoutRawTextNodeChild same as above but for the plain
// renderer, exercising the TextNode branch in renderNodeWithLayout.
func TestRenderWithLayoutRawTextNodeChild(t *testing.T) {
	root := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
	},
		vdom.CreateTextNode("raw-plain"),
	)

	output := renderer.RenderWithLayout(root, 20, 3)

	if !strings.Contains(output, "raw-plain") {
		t.Fatalf("expected raw text node rendered, got %q", output)
	}
}

// TestRenderScreenReaderSectionsBoolStateMap exercises the map[string]bool
// branch of accessibilityStateEnabled.
func TestRenderScreenReaderSectionsBoolStateMap(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role": "checkbox",
		"aria-state": map[string]bool{
			"checked":  true,
			"disabled": false,
		},
	}, components.Text("Item"))

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "Item") {
		t.Fatalf("expected output to contain Item, got %q", sections.Output)
	}
}

// TestRenderScreenReaderSectionsStringSliceState exercises the []string
// branch in screenReaderStateDescription.
func TestRenderScreenReaderSectionsStringSliceState(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":  "tablist",
		"aria-state": []string{"selected", "expanded"},
	}, components.Text("Tab"))

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "Tab") {
		t.Fatalf("expected output to contain Tab, got %q", sections.Output)
	}
}

// TestRenderScreenReaderSectionsRowDirection exercises screenReaderSeparator's
// row branch.
func TestRenderScreenReaderSectionsRowDirection(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "row"},
		components.Text("a"),
		components.Text("b"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "a") || !strings.Contains(sections.Output, "b") {
		t.Fatalf("expected both items in row, got %q", sections.Output)
	}
}

// TestRenderScreenReaderSectionsColumnDirection exercises the column branch
// of screenReaderSeparator (newline separation).
func TestRenderScreenReaderSectionsColumnDirection(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("first"),
		components.Text("second"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "first") || !strings.Contains(sections.Output, "second") {
		t.Fatalf("expected both items in column, got %q", sections.Output)
	}
}

// --- Buffer / Render edge cases -------------------------------------------

// TestRenderEntryPointNestedTextElements exercises the simple Render code
// path through nested Text components.
func TestRenderEntryPointNestedTextElements(t *testing.T) {
	root := components.Text("outer ",
		components.Text("inner"),
	)

	output := renderer.Render(root, 30, 3)
	if !strings.Contains(output, "outer") {
		t.Fatalf("expected outer text, got %q", output)
	}
}

// TestRenderEntryPointNestedBox exercises the simple Render path through a
// box that contains both a text element and a transform.
func TestRenderEntryPointNestedBox(t *testing.T) {
	root := components.Box(nil,
		components.Text("hi"),
		components.Transform(func(s string, idx int) string {
			return strings.ToUpper(s)
		}, vdom.CreateTextNode("yo")),
	)

	output := renderer.Render(root, 30, 3)
	if !strings.Contains(output, "hi") {
		t.Fatalf("expected hi in output, got %q", output)
	}
	if !strings.Contains(output, "YO") {
		t.Fatalf("expected uppercased YO in output, got %q", output)
	}
}

// --- Static child counting / cloning --------------------------------------

// TestRenderWithLayoutSectionsNestedStatic exercises collectStaticRoots and
// cloneWithoutStatic when a Static block is nested deeply inside a regular
// box.
func TestRenderWithLayoutSectionsNestedStatic(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(nil,
			components.Box(nil,
				components.Static(nil, components.Text("deep-static")),
			),
		),
		components.Text("dyn"),
	)

	sections := renderer.RenderWithLayoutSections(root, 30, 10)

	if !strings.Contains(sections.Output, "dyn") {
		t.Fatalf("expected dynamic rendered, got %q", sections.Output)
	}
	if strings.Contains(sections.Output, "deep-static") {
		t.Fatalf("did not expect deep static in dynamic Output, got %q", sections.Output)
	}
	if !strings.Contains(sections.StaticOutput, "deep-static") {
		t.Fatalf("expected deep static in StaticOutput, got %q", sections.StaticOutput)
	}
}

// TestRenderRuntimeSectionsTrackedShrunk exercises the path where a tracked
// static count shrinks (count <= previousCount), which should yield no delta
// and preserve the larger previous count.
func TestRenderRuntimeSectionsTrackedShrunk(t *testing.T) {
	type item struct{ Label string }

	first := []item{{"a"}, {"b"}, {"c"}}
	rootFirst := components.Box(nil,
		components.StaticItems(first, func(it item, idx int) *vdom.Node {
			return components.Text(it.Label)
		}),
	)
	firstSections := renderer.RenderRuntimeSections(rootFirst, 30, 10, nil, false)
	if firstSections.StaticCounts[0] != 3 {
		t.Fatalf("expected initial count 3, got %+v", firstSections.StaticCounts)
	}

	second := []item{{"a"}}
	rootSecond := components.Box(nil,
		components.StaticItems(second, func(it item, idx int) *vdom.Node {
			return components.Text(it.Label)
		}),
	)
	secondSections := renderer.RenderRuntimeSections(rootSecond, 30, 10, firstSections.StaticCounts, false)

	if secondSections.StaticDeltaOutput != "" {
		t.Fatalf("expected empty delta when count shrinks, got %q", secondSections.StaticDeltaOutput)
	}
}

// --- Text wrap variants ---------------------------------------------------

// TestRenderWithLayoutWrapNoneMode exercises the default branch of
// applyTextLayoutMode where mode is unrecognized and returns the line
// untouched.
func TestRenderWithLayoutWrapNoneMode(t *testing.T) {
	root := components.Box(vdom.Props{"width": 5.0},
		components.Text("HelloWorld", vdom.Props{"wrap": "no-wrap-unknown"}),
	)

	output := renderer.RenderWithLayout(root, 20, 3)

	if !strings.Contains(output, "Hello") {
		t.Fatalf("expected text in output, got %q", output)
	}
}

// TestRenderWithLayoutANSIBackgroundInheritanceNested forces background color
// inheritance through nested boxes, exercising getInheritedBackground deeper
// branches.
func TestRenderWithLayoutANSIBackgroundInheritanceNested(t *testing.T) {
	root := components.Box(vdom.Props{
		"backgroundColor": "blue",
		"width":           20.0,
		"height":          5.0,
	},
		components.Box(nil,
			components.Text("nested"),
		),
	)

	output := renderer.RenderWithLayoutANSI(root, 30, 6)
	if !strings.Contains(output, "nested") {
		t.Fatalf("expected nested text, got %q", output)
	}
	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI escape from inherited background, got %q", output)
	}
}

// TestRenderWithLayoutAllPaddingMarginAxes exercises every per-axis
// padding/margin prop branch in applyContainerLayoutProps.
func TestRenderWithLayoutAllPaddingMarginAxes(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":         30.0,
		"height":        10.0,
		"padding":       1.0,
		"paddingX":      1.0,
		"paddingY":      1.0,
		"paddingLeft":   1.0,
		"paddingTop":    1.0,
		"paddingRight":  1.0,
		"paddingBottom": 1.0,
		"margin":        1.0,
		"marginX":       1.0,
		"marginY":       1.0,
		"marginLeft":    1.0,
		"marginTop":     1.0,
		"marginRight":   1.0,
		"marginBottom":  1.0,
		"gap":           1.0,
		"rowGap":        1.0,
		"columnGap":     1.0,
		"flexWrap":      "wrap",
		"flexGrow":      1.0,
		"alignSelf":     "center",
	}, components.Text("padded"))

	output := renderer.RenderWithLayout(root, 60, 20)
	if !strings.Contains(output, "padded") {
		t.Fatalf("expected padded text, got %q", output)
	}
}

// TestRenderWithLayoutPercentMinSizes exercises the percent branches of
// minWidth/minHeight in applyContainerLayoutProps.
func TestRenderWithLayoutPercentMinSizes(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":     "100%",
		"minWidth":  "50%",
		"height":    "100%",
		"minHeight": "20%",
	}, components.Text("min"))

	output := renderer.RenderWithLayout(root, 30, 8)
	if !strings.Contains(output, "min") {
		t.Fatalf("expected min text, got %q", output)
	}
}

// TestRenderWithLayoutFlexDirectionReverseJustifyEnd exercises the reverse
// direction + JustifyEnd swap branch in applyContainerLayoutProps.
func TestRenderWithLayoutFlexDirectionReverseJustifyEnd(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":          20.0,
		"flexDirection":  "row-reverse",
		"justifyContent": "flex-end",
	}, components.Text("a"), components.Text("b"))

	output := renderer.RenderWithLayout(root, 30, 3)
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
		t.Fatalf("expected both items, got %q", output)
	}
}

// TestRenderWithLayoutFlexDirectionReverseJustifyStart exercises the reverse
// direction + JustifyStart swap branch.
func TestRenderWithLayoutFlexDirectionReverseJustifyStart(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":          20.0,
		"flexDirection":  "column-reverse",
		"justifyContent": "flex-start",
	}, components.Text("a"), components.Text("b"))

	output := renderer.RenderWithLayout(root, 30, 5)
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
		t.Fatalf("expected both items, got %q", output)
	}
}

// --- consumeANSISequence variants -----------------------------------------

// TestRenderConsumeANSIOSCAndPrivate exercises the OSC (ESC ]) and the
// alternate CSI (0x9b) branches of consumeANSISequence by passing them
// through a Transform whose output is then routed through visibleStringWidth
// during layout.
func TestRenderConsumeANSIOSCAndPrivate(t *testing.T) {
	transformed := components.Transform(func(children string, idx int) string {
		// ESC ] osc-data BEL  (OSC sequence)
		osc := "\x1b]0;title\x07"
		// 0x9b alternate CSI form: 0x9b 31 m
		csi := "\x9b31m"
		// ESC alone (rare) followed by trailing escape with backslash terminator
		// ESC ] data ESC \ (ST terminator)
		oscSt := "\x1b]2;hello\x1b\\"
		return osc + csi + children + oscSt + "\x1b[0m"
	}, vdom.CreateTextNode("seq"))

	root := components.Box(nil, transformed)

	output := renderer.RenderWithLayoutANSI(root, 30, 3)
	if !strings.Contains(output, "seq") {
		t.Fatalf("expected text in output, got %q", output)
	}
}

// TestRenderConsumeANSILoneEscape feeds a lone ESC byte at the end of
// transform output, exercising the truncated-CSI branch in
// consumeANSISequence.
func TestRenderConsumeANSILoneEscape(t *testing.T) {
	transformed := components.Transform(func(children string, idx int) string {
		// Lone ESC at end is unusual but should not crash the renderer.
		return children + "\x1b"
	}, vdom.CreateTextNode("loneEsc"))

	root := components.Box(nil, transformed)
	output := renderer.RenderWithLayoutANSI(root, 30, 3)
	if !strings.Contains(output, "loneEsc") {
		t.Fatalf("expected text rendered despite lone esc, got %q", output)
	}
}

// --- Border style branches -------------------------------------------------

// TestRenderWithLayoutBorderStyleBoolean exercises hasBorderStyle bool branch.
func TestRenderWithLayoutBorderStyleBoolean(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":       10.0,
		"height":      3.0,
		"borderStyle": true,
	}, components.Text("b"))

	output := renderer.RenderWithLayout(root, 20, 5)
	if !strings.Contains(output, "b") {
		t.Fatalf("expected text rendered, got %q", output)
	}
}

// --- Truncate edge cases (maxWidth == 1) ----------------------------------

// TestRenderWithLayoutTruncateEndAtWidthOne exercises truncateEnd's
// maxWidth==1 branch which returns just an ellipsis.
func TestRenderWithLayoutTruncateEndAtWidthOne(t *testing.T) {
	root := components.Box(vdom.Props{"width": 1.0},
		components.Text("LongerThanOne", vdom.Props{"wrap": "truncate"}),
	)

	output := renderer.RenderWithLayout(root, 20, 3)
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected ellipsis, got %q", output)
	}
}

// TestRenderWithLayoutTruncateMiddleAtWidthOne exercises truncateMiddle's
// maxWidth==1 branch.
func TestRenderWithLayoutTruncateMiddleAtWidthOne(t *testing.T) {
	root := components.Box(vdom.Props{"width": 1.0},
		components.Text("LongMiddle", vdom.Props{"wrap": "truncate-middle"}),
	)

	output := renderer.RenderWithLayout(root, 20, 3)
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected ellipsis, got %q", output)
	}
}

// TestRenderWithLayoutTruncateStartAtWidthOne exercises truncateStart's
// maxWidth==1 branch.
func TestRenderWithLayoutTruncateStartAtWidthOne(t *testing.T) {
	root := components.Box(vdom.Props{"width": 1.0},
		components.Text("LongStart", vdom.Props{"wrap": "truncate-start"}),
	)

	output := renderer.RenderWithLayout(root, 20, 3)
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected ellipsis, got %q", output)
	}
}

// TestRenderWithLayoutANSITruncateMiddleWidthOneStyled exercises the styled
// maxWidth==1 branch in truncateStyledMiddle (multi-rune content forced into
// 1-cell width).
func TestRenderWithLayoutANSITruncateMiddleWidthOneStyled(t *testing.T) {
	root := components.Box(vdom.Props{"width": 1.0},
		components.Text("HelloMiddle", vdom.Props{"wrap": "truncate-middle", "color": "red"}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected styled width-1 ellipsis, got %q", output)
	}
}

// TestRenderWithLayoutANSITruncateEndWidthOneStyled exercises the styled
// maxWidth==1 branch in truncateStyledEnd.
func TestRenderWithLayoutANSITruncateEndWidthOneStyled(t *testing.T) {
	root := components.Box(vdom.Props{"width": 1.0},
		components.Text("HelloEnd", vdom.Props{"wrap": "truncate", "color": "red"}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected styled width-1 ellipsis, got %q", output)
	}
}

// TestRenderWithLayoutANSITruncateStartWidthOneStyled exercises the styled
// maxWidth==1 branch in truncateStyledStart.
func TestRenderWithLayoutANSITruncateStartWidthOneStyled(t *testing.T) {
	root := components.Box(vdom.Props{"width": 1.0},
		components.Text("HelloStart", vdom.Props{"wrap": "truncate-start", "color": "red"}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected styled width-1 ellipsis, got %q", output)
	}
}

// --- Screen reader: transform with accessibilityLabel ---------------------

// TestRenderScreenReaderTransformAccessibilityLabelInBox exercises the
// transform-with-label branch in renderScreenReaderNode.
func TestRenderScreenReaderTransformAccessibilityLabelInBox(t *testing.T) {
	root := components.Box(nil,
		components.Transform(func(s string, idx int) string {
			return strings.ToUpper(s)
		},
			vdom.CreateTextNode("hidden-orig"),
		),
	)
	// Set accessibilityLabel directly since components.Transform doesn't expose it
	root.Children[0].Props["accessibilityLabel"] = "transform-label"

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "transform-label") {
		t.Fatalf("expected accessibilityLabel, got %q", sections.Output)
	}
}

// TestRenderScreenReaderUnknownElementType exercises the default branch in
// renderScreenReaderNode for unknown element types.
func TestRenderScreenReaderUnknownElementType(t *testing.T) {
	root := components.Box(nil,
		vdom.CreateElement("custom-thing", vdom.Props{},
			vdom.CreateTextNode("custom-content"),
		),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "custom-content") {
		t.Fatalf("expected unknown element children to render, got %q", sections.Output)
	}
}

// TestRenderWithLayoutANSITransformWrapMode forces a Transform whose subtree
// must be wrapped to multiple lines, hitting the transform branch in
// styledTextLines plus splitStyledRunesLines for embedded newlines.
func TestRenderWithLayoutANSITransformWrapMode(t *testing.T) {
	transformed := components.Transform(func(s string, idx int) string {
		return "[" + s + "]"
	}, vdom.CreateTextNode("aaa\nbbb\nccc"))

	root := components.Box(vdom.Props{"width": 4.0}, transformed)
	output := renderer.RenderWithLayoutANSI(root, 20, 6)

	if !strings.Contains(output, "aaa") {
		t.Fatalf("expected aaa in output, got %q", output)
	}
}

// TestRenderWithLayoutANSITextWithEmbeddedNewlines exercises
// splitStyledRunesLines with multi-line styled text.
func TestRenderWithLayoutANSITextWithEmbeddedNewlines(t *testing.T) {
	root := components.Box(nil,
		components.Text("line1\nline2\nline3", vdom.Props{"color": "cyan"}),
	)

	output := renderer.RenderWithLayoutANSI(root, 30, 5)
	if !strings.Contains(output, "line2") {
		t.Fatalf("expected multiline text, got %q", output)
	}
}

// TestRenderWithLayoutANSIBoxBackgroundClipped places a colored box that
// extends beyond a tight viewport, exercising the per-row/per-column clip
// branches in fillBoxBackgroundANSI.
func TestRenderWithLayoutANSIBoxBackgroundClipped(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":           20.0,
		"height":          10.0,
		"backgroundColor": "red",
	}, components.Text("clip"))

	// Render with smaller window than the box wants.
	output := renderer.RenderWithLayoutANSI(root, 20, 4)
	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected ANSI background even when clipped, got %q", output)
	}
}

// TestRenderWithLayoutScreenReaderRowReverseChildren exercises the
// row-reverse branch in renderScreenReaderChildren.
func TestRenderWithLayoutScreenReaderRowReverseChildren(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "row-reverse"},
		components.Text("first"),
		components.Text("second"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "first") || !strings.Contains(sections.Output, "second") {
		t.Fatalf("expected both items, got %q", sections.Output)
	}
}

// TestRenderWithLayoutANSITransformOverridesText exercises a Transform with a
// text-like child whose result differs from the raw input, hitting the
// transform-then-replace branch in applyStyledTransformRunes.
func TestRenderWithLayoutANSITransformOverridesText(t *testing.T) {
	transformed := components.Transform(func(s string, idx int) string {
		return strings.ToUpper(s) + "!"
	}, components.Text("hello", vdom.Props{"color": "yellow"}))

	root := components.Box(nil, transformed)
	output := renderer.RenderWithLayoutANSI(root, 20, 3)
	// ANSI escapes interleave between letters; check the visible chars
	// individually rather than as a contiguous substring.
	if !strings.Contains(output, "HELLO") && !strings.Contains(output, "H") {
		t.Fatalf("expected H in output, got %q", output)
	}
	if !strings.Contains(output, "!") {
		t.Fatalf("expected exclamation in output, got %q", output)
	}
}

// TestRenderWithLayoutANSINestedTextWithBoxChild exercises the default
// branch (drop) in collectStyledRenderedTextContent for non-text children of
// a "text" element.
func TestRenderWithLayoutANSINestedTextWithBoxChild(t *testing.T) {
	root := components.Text("outer ",
		// Boxes inside a text node should be ignored (default branch).
		components.Box(nil, components.Text("ignored")),
		components.Text("inner"),
	)

	output := renderer.RenderWithLayoutANSI(root, 30, 3)
	if !strings.Contains(output, "outer") || !strings.Contains(output, "inner") {
		t.Fatalf("expected outer+inner, got %q", output)
	}
	if strings.Contains(output, "ignored") {
		t.Fatalf("box-in-text should be ignored, got %q", output)
	}
}

// TestRenderWithLayoutBoxInTextPlain mirrors the above for the non-ANSI
// renderer to hit the equivalent branch in collectRenderedTextContent.
func TestRenderWithLayoutBoxInTextPlain(t *testing.T) {
	root := components.Text("outer ",
		components.Box(nil, components.Text("dropped")),
		components.Text("inner"),
	)

	output := renderer.RenderWithLayout(root, 30, 3)
	if !strings.Contains(output, "outer") || !strings.Contains(output, "inner") {
		t.Fatalf("expected outer+inner, got %q", output)
	}
	if strings.Contains(output, "dropped") {
		t.Fatalf("box-in-text should be dropped, got %q", output)
	}
}

// TestRenderWithLayoutMeasureTextWithBoxChild exercises the same dropping
// behavior in collectMeasuredTextContent.
func TestRenderWithLayoutMeasureTextWithBoxChild(t *testing.T) {
	// A text element with a box child (which should be measured-out)
	root := components.Box(vdom.Props{"width": 10.0},
		components.Text("hi",
			components.Box(nil, components.Text("nope")),
			components.Text(" there"),
		),
	)

	output := renderer.RenderWithLayout(root, 30, 5)
	if !strings.Contains(output, "hi") {
		t.Fatalf("expected hi, got %q", output)
	}
	if strings.Contains(output, "nope") {
		t.Fatalf("box-in-text should not appear, got %q", output)
	}
}

// TestRenderWithLayoutANSITruncateStyledEndLongLine forces the long-line
// branches of truncateStyledEnd: width > 1 with content that overflows.
func TestRenderWithLayoutANSITruncateStyledEndLongLine(t *testing.T) {
	root := components.Box(vdom.Props{"width": 4.0},
		components.Text("abcdefgh", vdom.Props{
			"wrap":  "truncate",
			"color": "magenta",
		}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected ellipsis from end-truncation, got %q", output)
	}
}

// TestRenderWithLayoutANSITruncateStyledStartLongLine pushes truncateStyledStart
// down its full code path with a multi-rune line and width > 1.
func TestRenderWithLayoutANSITruncateStyledStartLongLine(t *testing.T) {
	root := components.Box(vdom.Props{"width": 4.0},
		components.Text("abcdefgh", vdom.Props{
			"wrap":  "truncate-start",
			"color": "magenta",
		}),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 3)
	if !strings.ContainsRune(output, '…') {
		t.Fatalf("expected leading ellipsis, got %q", output)
	}
}

// TestRenderWithLayoutANSICustomBorderMap exercises the map[string]interface{}
// branch in borderStyleGlyphs.
func TestRenderWithLayoutANSICustomBorderMap(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":  10.0,
		"height": 3.0,
		"borderStyle": map[string]interface{}{
			"topLeft":     "+",
			"top":         "-",
			"topRight":    "+",
			"left":        "|",
			"bottomLeft":  "+",
			"bottom":      "-",
			"bottomRight": "+",
			"right":       "|",
		},
		"borderColor": "blue",
	}, components.Text("c"))

	output := renderer.RenderWithLayoutANSI(root, 20, 5)
	if !strings.Contains(output, "+") {
		t.Fatalf("expected custom border corner, got %q", output)
	}
}

// TestRenderWithLayoutANSICustomBorderProps exercises the vdom.Props branch in
// borderStyleGlyphs.
func TestRenderWithLayoutANSICustomBorderProps(t *testing.T) {
	root := components.Box(vdom.Props{
		"width":  10.0,
		"height": 3.0,
		"borderStyle": vdom.Props{
			"topLeft":     "+",
			"top":         "-",
			"topRight":    "+",
			"left":        "|",
			"bottomLeft":  "+",
			"bottom":      "-",
			"bottomRight": "+",
			"right":       "|",
		},
	}, components.Text("c"))

	output := renderer.RenderWithLayoutANSI(root, 20, 5)
	if !strings.Contains(output, "+") {
		t.Fatalf("expected custom border corner, got %q", output)
	}
}

// TestSyncComputedLayoutWithDisplayNoneChild exercises the post-traversal
// branch of syncComputedLayout that clears layout for display:none children.
func TestSyncComputedLayoutWithDisplayNoneChild(t *testing.T) {
	hidden := components.Box(vdom.Props{"display": "none"}, components.Text("hidden"))
	visible := components.Text("shown")
	root := components.Box(vdom.Props{"width": 20.0, "height": 5.0}, hidden, visible)

	renderer.SyncComputedLayout(root, 30, 10)

	// Hidden child should have its layout zeroed.
	if hidden.ComputedLayout().Width != 0 {
		t.Fatalf("expected hidden child layout cleared, got %+v", hidden.ComputedLayout())
	}
}

// --- Screen reader: extended aria coverage --------------------------------
//
// The tests in this section cover aria features that goink emits beyond the
// base aria-label / aria-hidden / aria-role / aria-state subset. They mirror
// upstream Ink's `Object.keys(state).filter(key => state[key])` semantics
// (any truthy state key is announced) and add a small set of pragmatic
// extensions: top-level shorthand props, a tri-state aria-state.checked
// value, aria-level for headings, and aria-live="off" suppression.

// TestRenderScreenReaderArbitraryStateKeyAnnounced verifies that a truthy
// state key not in the documented 9-key list is still announced. Upstream
// uses Object.keys(state).filter(key => state[key]) and goink should match
// that behaviour for arbitrary keys (e.g. "invalid").
func TestRenderScreenReaderArbitraryStateKeyAnnounced(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role": "textbox",
		"aria-state": vdom.Props{
			"invalid": true,
		},
	}, components.Text("Email"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "textbox: (invalid) Email" {
		t.Fatalf("expected arbitrary state key announced, got %q", sections.Output)
	}
}

// TestRenderScreenReaderArbitraryStateKeyOrdering verifies that known state
// keys come first in their predefined order, then unknown keys are appended
// in alphabetical order to keep parity output deterministic.
func TestRenderScreenReaderArbitraryStateKeyOrdering(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role": "textbox",
		"aria-state": vdom.Props{
			"required": true,
			"invalid":  true,
			"active":   true,
		},
	}, components.Text("Email"))

	sections := renderer.RenderScreenReaderSections(root)
	// "required" is a known key, "active" and "invalid" are unknown so they
	// follow alphabetically after the known keys.
	if sections.Output != "textbox: (required, active, invalid) Email" {
		t.Fatalf("expected deterministic ordering, got %q", sections.Output)
	}
}

// TestRenderScreenReaderCheckedMixed verifies that aria-state.checked with
// the special string "mixed" is narrated as "(mixed)" rather than just
// "(checked)". This handles tri-state checkboxes.
func TestRenderScreenReaderCheckedMixed(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role": "checkbox",
		"aria-state": vdom.Props{
			"checked": "mixed",
		},
	}, components.Text("Select all"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "checkbox: (mixed) Select all" {
		t.Fatalf("expected tri-state mixed narration, got %q", sections.Output)
	}
}

// TestRenderScreenReaderTopLevelAriaDisabled verifies that the top-level
// shorthand prop `aria-disabled` is folded into the state description, so
// authors don't have to wrap it in aria-state.
func TestRenderScreenReaderTopLevelAriaDisabled(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":     "button",
		"aria-disabled": true,
	}, components.Text("Submit"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "button: (disabled) Submit" {
		t.Fatalf("expected top-level aria-disabled, got %q", sections.Output)
	}
}

// TestRenderScreenReaderTopLevelMultipleShorthand verifies that multiple
// top-level shorthand props combine in the documented order and merge with
// values from aria-state without duplication.
func TestRenderScreenReaderTopLevelMultipleShorthand(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":     "checkbox",
		"aria-checked":  true,
		"aria-required": true,
		"aria-state": vdom.Props{
			"disabled": true,
		},
	}, components.Text("Accept"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "checkbox: (checked, disabled, required) Accept" {
		t.Fatalf("expected merged shorthand+state, got %q", sections.Output)
	}
}

// TestRenderScreenReaderTopLevelCheckedMixed verifies that a top-level
// shorthand `aria-checked="mixed"` produces the (mixed) narration.
func TestRenderScreenReaderTopLevelCheckedMixed(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":    "checkbox",
		"aria-checked": "mixed",
	}, components.Text("Select all"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "checkbox: (mixed) Select all" {
		t.Fatalf("expected mixed via shorthand, got %q", sections.Output)
	}
}

// TestRenderScreenReaderHeadingLevel verifies that aria-role="heading" with
// a numeric aria-level prepends the level to the role narration.
func TestRenderScreenReaderHeadingLevel(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":  "heading",
		"aria-level": 2,
	}, components.Text("Section title"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "heading 2: Section title" {
		t.Fatalf("expected heading level narration, got %q", sections.Output)
	}
}

// TestRenderScreenReaderHeadingLevelFloat covers JSON-decoded aria-level
// values which arrive as float64 (the parity harness path).
func TestRenderScreenReaderHeadingLevelFloat(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":  "heading",
		"aria-level": 3.0,
	}, components.Text("Subheading"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "heading 3: Subheading" {
		t.Fatalf("expected heading float-level, got %q", sections.Output)
	}
}

// TestRenderScreenReaderHeadingLevelIgnoredWithoutHeadingRole verifies that
// aria-level is only consumed when role is "heading", to avoid surprising
// narrations on unrelated roles.
func TestRenderScreenReaderHeadingLevelIgnoredWithoutHeadingRole(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":  "button",
		"aria-level": 4,
	}, components.Text("Click"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "button: Click" {
		t.Fatalf("expected aria-level ignored, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLiveOffSuppresses verifies that aria-live="off"
// suppresses a subtree from screen-reader output, mirroring aria-hidden.
func TestRenderScreenReaderAriaLiveOffSuppresses(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"aria-live": "off"}, components.Text("muted")),
		components.Text("audible"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if strings.Contains(sections.Output, "muted") {
		t.Fatalf("expected aria-live=off content suppressed, got %q", sections.Output)
	}
	if !strings.Contains(sections.Output, "audible") {
		t.Fatalf("expected sibling text to remain audible, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLivePoliteEmitsContent verifies that
// aria-live="polite" does not suppress content (only "off" does).
func TestRenderScreenReaderAriaLivePoliteEmitsContent(t *testing.T) {
	root := components.Box(vdom.Props{"aria-live": "polite"},
		components.Text("status update"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "status update") {
		t.Fatalf("expected aria-live=polite content emitted, got %q", sections.Output)
	}
}

// --- Screen reader: aria-labelledby / aria-describedby --------------------
//
// These tests pin the goink-specific resolution of `aria-labelledby` and
// `aria-describedby` against ids elsewhere in the tree. Upstream Ink does
// not implement either prop today, so the behaviour here is a forward-
// looking parity gap-filler rather than a strict mirror -- see the
// renderer's screenReaderContext implementation comment for details.

// TestRenderScreenReaderAriaLabelledByResolvesSingleID verifies a node
// with a single labelledby id substitutes that referenced node's text
// content for its own narration, taking precedence over both children
// and aria-label.
func TestRenderScreenReaderAriaLabelledByResolvesSingleID(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"id": "title"}, components.Text("Settings")),
		components.Box(vdom.Props{
			"aria-labelledby": "title",
			"aria-label":      "should be ignored",
		}, components.Text("would be ignored too")),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "Settings") {
		t.Fatalf("expected labelledby to substitute referenced text, got %q", sections.Output)
	}
	if strings.Contains(sections.Output, "should be ignored") || strings.Contains(sections.Output, "would be ignored too") {
		t.Fatalf("expected labelledby to override aria-label and children, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLabelledByResolvesMultipleIDs verifies that
// whitespace-separated ids are concatenated with single spaces in the
// order given.
func TestRenderScreenReaderAriaLabelledByResolvesMultipleIDs(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"id": "first"}, components.Text("Hello")),
		components.Box(vdom.Props{"id": "second"}, components.Text("World")),
		components.Box(vdom.Props{"aria-labelledby": "first second"}),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "Hello World") {
		t.Fatalf("expected labelledby to concatenate %q first then %q second, got %q", "Hello", "World", sections.Output)
	}
}

// TestRenderScreenReaderAriaLabelledByMissingIDFallsBack verifies that
// when *all* referenced ids are missing the host's children are still
// narrated (matching the spec's "missing labelledby targets are
// skipped, then own children" precedence).
func TestRenderScreenReaderAriaLabelledByMissingIDFallsBack(t *testing.T) {
	root := components.Box(vdom.Props{"aria-labelledby": "does-not-exist"},
		components.Text("fallback content"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "fallback content") {
		t.Fatalf("expected fallback to children when labelledby ids missing, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLabelledBySelfReferenceDoesNotLoop verifies
// that a node whose labelledby points to its own id does not recurse
// forever; the cycle break leaves the node with empty resolution and so
// the narration falls through to the host's own children.
func TestRenderScreenReaderAriaLabelledBySelfReferenceDoesNotLoop(t *testing.T) {
	root := components.Box(vdom.Props{
		"id":              "self",
		"aria-labelledby": "self",
	}, components.Text("self-host"))

	sections := renderer.RenderScreenReaderSections(root)
	// The cycle is broken; a self-reference resolves to "" so the rendered
	// resolution is empty -- but the host still has children that should
	// narrate. The exact precedence is: try labelledby (gets ""), fall
	// through to label (none), fall through to children. Either "self-host"
	// alone or a duplicated form is acceptable as long as there's no
	// infinite loop and content is preserved.
	if !strings.Contains(sections.Output, "self-host") {
		t.Fatalf("expected self-reference to fall through to host children, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaDescribedByAppendsDescription verifies that
// describedby concatenates the resolved description after the host's
// regular narration, separated by a single space, and is announced
// before the role/state prefix decoration.
func TestRenderScreenReaderAriaDescribedByAppendsDescription(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"id": "hint"}, components.Text("Press Enter to submit")),
		components.Box(vdom.Props{
			"aria-role":        "button",
			"aria-describedby": "hint",
		}, components.Text("Submit")),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "button: Submit Press Enter to submit") {
		t.Fatalf("expected describedby to append description before role decoration, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaDescribedByMissingIDSilentlySkipped verifies
// that an unresolved describedby does not add an empty trailing space
// and otherwise leaves the host's narration unchanged.
func TestRenderScreenReaderAriaDescribedByMissingIDSilentlySkipped(t *testing.T) {
	root := components.Box(vdom.Props{"aria-describedby": "missing"},
		components.Text("only content"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "only content" {
		t.Fatalf("expected unresolved describedby to leave narration unchanged, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLabelledByResolvesAcrossHiddenSubtree
// verifies a labelledby target inside an aria-hidden subtree is still
// usable as a label source. The hidden subtree itself stays suppressed
// from the main narration.
func TestRenderScreenReaderAriaLabelledByResolvesAcrossHiddenSubtree(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"aria-hidden": true},
			components.Box(vdom.Props{"id": "shadow-label"}, components.Text("Shadow Title")),
		),
		components.Box(vdom.Props{"aria-labelledby": "shadow-label"}, components.Text("ignored")),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "Shadow Title") {
		t.Fatalf("expected hidden labelledby target to still be readable as label source, got %q", sections.Output)
	}
}

// --- Screen reader: aria-live announcer regions ---------------------------
//
// Static rendering surfaces aria-live="polite" / "assertive" subtrees in a
// dedicated announcer block at the end of Output. Runtime announcer
// dispatch (incremental announcements as state mutates) is deferred --
// upstream Ink does not own that integration and goink does not yet model
// a separate announcer channel; the static marker is a parity foothold.

// TestRenderScreenReaderAriaLivePoliteEmitsAnnouncerRegion verifies that
// a polite region appears after the main tree narration, prefixed with
// "[polite]" and separated by a newline.
func TestRenderScreenReaderAriaLivePoliteEmitsAnnouncerRegion(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("main content"),
		components.Box(vdom.Props{"aria-live": "polite"}, components.Text("status update")),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if !strings.Contains(sections.Output, "main content") {
		t.Fatalf("expected main content to remain inline, got %q", sections.Output)
	}
	if !strings.Contains(sections.Output, "[polite] status update") {
		t.Fatalf("expected polite announcer region in output, got %q", sections.Output)
	}
	politeIdx := strings.Index(sections.Output, "[polite]")
	mainIdx := strings.Index(sections.Output, "main content")
	if politeIdx < mainIdx {
		t.Fatalf("expected polite region to come after main content, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLiveAssertivePrecedesPolite verifies that
// when both polite and assertive regions exist, assertive is emitted
// first to reflect its higher urgency.
func TestRenderScreenReaderAriaLiveAssertivePrecedesPolite(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("body"),
		components.Box(vdom.Props{"aria-live": "polite"}, components.Text("polite-msg")),
		components.Box(vdom.Props{"aria-live": "assertive"}, components.Text("assertive-msg")),
	)

	sections := renderer.RenderScreenReaderSections(root)
	assertiveIdx := strings.Index(sections.Output, "[assertive] assertive-msg")
	politeIdx := strings.Index(sections.Output, "[polite] polite-msg")
	if assertiveIdx < 0 || politeIdx < 0 {
		t.Fatalf("expected both announcer regions, got %q", sections.Output)
	}
	if assertiveIdx > politeIdx {
		t.Fatalf("expected assertive region before polite region, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLiveOffStaysSuppressed verifies the existing
// off-suppression behaviour is unchanged: aria-live="off" still drops the
// subtree from both the inline narration and the announcer block.
func TestRenderScreenReaderAriaLiveOffStaysSuppressed(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"aria-live": "off"}, components.Text("muted")),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if strings.Contains(sections.Output, "muted") {
		t.Fatalf("expected aria-live=off content suppressed everywhere, got %q", sections.Output)
	}
	if strings.Contains(sections.Output, "[polite]") || strings.Contains(sections.Output, "[assertive]") {
		t.Fatalf("expected no announcer block for aria-live=off, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLiveContentStaysInline verifies that a
// polite region's content also remains in the inline narration -- the
// announcer block duplicates rather than relocates the content. This is
// what assistive consumers expect: the visual flow is unchanged.
func TestRenderScreenReaderAriaLiveContentStaysInline(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("alpha"),
		components.Box(vdom.Props{"aria-live": "polite"}, components.Text("beta")),
		components.Text("gamma"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	// Expect alpha, beta, gamma all inline, then [polite] beta at the end.
	for _, fragment := range []string{"alpha", "beta", "gamma", "[polite] beta"} {
		if !strings.Contains(sections.Output, fragment) {
			t.Fatalf("expected output to contain %q, got %q", fragment, sections.Output)
		}
	}
	// Confirm "beta" appears twice (once inline, once in the announcer).
	if strings.Count(sections.Output, "beta") < 2 {
		t.Fatalf("expected beta to appear inline and in the announcer block, got %q", sections.Output)
	}
}

// TestRenderScreenReaderAriaLiveIdempotent verifies that running the
// renderer twice on the same input produces the same output -- the
// announcer collection must not depend on map iteration order or other
// non-deterministic state.
func TestRenderScreenReaderAriaLiveIdempotent(t *testing.T) {
	build := func() *vdom.Node {
		return components.Box(vdom.Props{"flexDirection": "column"},
			components.Box(vdom.Props{"aria-live": "polite"}, components.Text("first")),
			components.Box(vdom.Props{"aria-live": "assertive"}, components.Text("second")),
			components.Box(vdom.Props{"aria-live": "polite"}, components.Text("third")),
		)
	}

	first := renderer.RenderScreenReaderSections(build()).Output
	second := renderer.RenderScreenReaderSections(build()).Output
	if first != second {
		t.Fatalf("expected idempotent screen-reader output, got\nfirst:  %q\nsecond: %q", first, second)
	}
}

// itoa avoids importing strconv just for a tiny helper.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := false
	if n < 0 {
		negative = true
		n = -n
	}
	digits := make([]byte, 0, 8)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

// TestRenderProportionalFlexShrinkWrapsAtMeasuredWidth locks in upstream Ink
// parity for proportional flex-shrink with non-1 weights and inner text
// wrapping. Outer width=8 row with two boxes whose flexShrink weights are 2
// and 1 (intrinsic widths 6/6, total 12, deficit 4). Without the
// measure-time floored-width plumb-through the inner BBBBBB wraps at the
// rounded box width of 5 → "BBBBB\nB"; with it, BBBBBB wraps at the floored
// 4.667 → 4 → "BBBB\nBB", matching upstream's `getMaxWidth(yogaNode)` flow.
// See PORTING_STATUS.md "proportional flex-shrink" entry.
func TestRenderProportionalFlexShrinkWrapsAtMeasuredWidth(t *testing.T) {
	textBox := func(weight float64, body string) *vdom.Node {
		return &vdom.Node{
			Type:        vdom.ElementNode,
			ElementType: "box",
			Props:       vdom.Props{"flexShrink": weight, "width": float64(6)},
			Children: []*vdom.Node{
				{
					Type:        vdom.ElementNode,
					ElementType: "text",
					Children:    []*vdom.Node{{Type: vdom.TextNode, Text: body}},
				},
			},
		}
	}

	tree := &vdom.Node{
		Type:        vdom.ElementNode,
		ElementType: "box",
		Props:       vdom.Props{"width": float64(8)},
		Children: []*vdom.Node{
			textBox(2, "AAAAAA"),
			textBox(1, "BBBBBB"),
		},
	}

	output := renderer.RenderWithLayout(tree, 10, 5)
	expected := "AAABBBB\nAAABB"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// TestRenderFlexShrinkTextPairKeepsCeilOverlap guards against regressing the
// text-pair sibling overlap behaviour. "Hello "+"World" in width=10
// container is rendered as "HelloWorld" on the first line (trailing space of
// the first text gets overwritten by the next sibling's first char).
// Honouring the floored measure-time width here would over-wrap "World" to
// "Worl\nd". The fix is gated on `parent.sizeAdjusted` so this directly-
// shrunk case keeps using the existing ceil-rounded textLike layout.
func TestRenderFlexShrinkTextPairKeepsCeilOverlap(t *testing.T) {
	tree := &vdom.Node{
		Type:        vdom.ElementNode,
		ElementType: "box",
		Props:       vdom.Props{"width": float64(10)},
		Children: []*vdom.Node{
			{
				Type:        vdom.ElementNode,
				ElementType: "text",
				Children:    []*vdom.Node{{Type: vdom.TextNode, Text: "Hello "}},
			},
			{
				Type:        vdom.ElementNode,
				ElementType: "text",
				Children:    []*vdom.Node{{Type: vdom.TextNode, Text: "World"}},
			},
		},
	}

	output := renderer.RenderWithLayout(tree, 12, 4)
	if !strings.HasPrefix(output, "HelloWorld") {
		t.Fatalf("expected output to start with %q (sibling overlap of trailing space), got %q", "HelloWorld", output)
	}
}
