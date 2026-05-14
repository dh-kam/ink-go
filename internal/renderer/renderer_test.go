package renderer_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// TestRenderTextNode tests rendering a simple text node
func TestRenderTextNode(t *testing.T) {
	node := vdom.CreateTextNode("Hello, World!")

	output := renderer.Render(node, 80, 24)

	expected := "Hello, World!"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
	}
}

// TestRenderEmptyNode tests rendering an empty node
func TestRenderEmptyNode(t *testing.T) {
	node := vdom.CreateElement("box", nil)

	output := renderer.Render(node, 80, 24)

	if output != "" {
		t.Errorf("Expected empty output, got:\n%q", output)
	}
}

func TestRenderWithLayoutNilNode(t *testing.T) {
	output := renderer.RenderWithLayout(nil, 80, 24)
	if output != "" {
		t.Errorf("Expected empty output, got %q", output)
	}
}

// TestRenderBoxWithText tests rendering a box containing text
func TestRenderBoxWithText(t *testing.T) {
	text := vdom.CreateTextNode("Inside Box")
	box := vdom.CreateElement("box", nil, text)

	output := renderer.Render(box, 80, 24)

	expected := "Inside Box"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
	}
}

func TestRenderTextIgnoresBoxChildren(t *testing.T) {
	root := components.Text(
		"Hello",
		components.Box(nil, components.Text("Ignored")),
		components.Text(" World"),
	)

	output := renderer.Render(root, 80, 24)
	expected := "Hello World"
	if !strings.Contains(output, expected) {
		t.Fatalf("expected text output to ignore Box children, got:\n%s", output)
	}
	if strings.Contains(output, "Ignored") {
		t.Fatalf("expected Box child text to be ignored, got:\n%s", output)
	}
}

// TestRenderNestedBoxes tests rendering nested boxes
func TestRenderNestedBoxes(t *testing.T) {
	innerText := vdom.CreateTextNode("Inner")
	innerBox := vdom.CreateElement("box", nil, innerText)

	outerText := vdom.CreateTextNode("Outer ")
	outerBox := vdom.CreateElement("box", nil, outerText, innerBox)

	output := renderer.Render(outerBox, 80, 24)

	if !strings.Contains(output, "Outer") {
		t.Error("Expected output to contain 'Outer'")
	}
	if !strings.Contains(output, "Inner") {
		t.Error("Expected output to contain 'Inner'")
	}
}

// TestRenderMultipleChildren tests rendering multiple children
func TestRenderMultipleChildren(t *testing.T) {
	child1 := vdom.CreateTextNode("First ")
	child2 := vdom.CreateTextNode("Second ")
	child3 := vdom.CreateTextNode("Third")

	box := vdom.CreateElement("box", nil, child1, child2, child3)

	output := renderer.Render(box, 80, 24)

	expected := "First Second Third"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
	}
}

func TestRenderTextElementSequence(t *testing.T) {
	root := vdom.CreateElement("box", nil,
		components.Text("First"),
		vdom.CreateTextNode(" "),
		components.Text("Second"),
	)

	output := renderer.Render(root, 80, 24)

	if !strings.Contains(output, "First Second") {
		t.Errorf("Expected rendered text sequence, got:\n%s", output)
	}
}

func TestRenderTransform(t *testing.T) {
	root := vdom.CreateElement("box", nil,
		components.Transform(func(children string, index int) string {
			return strings.ToUpper(children)
		}, vdom.CreateTextNode("hello")),
	)

	output := renderer.Render(root, 80, 24)

	if !strings.Contains(output, "HELLO") {
		t.Errorf("Expected transformed output, got:\n%s", output)
	}
}

func TestRenderScreenReaderUsesAriaLabel(t *testing.T) {
	root := components.Box(vdom.Props{"aria-label": "Hello World"},
		components.Text("Not visible to screen readers"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "Hello World" {
		t.Fatalf("expected screen-reader label output, got %q", sections.Output)
	}
}

func TestRenderScreenReaderRespectsAriaHiddenAndDisplayNone(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{"display": "none"}, components.Text("Hidden")),
		components.Box(vdom.Props{"aria-hidden": true}, components.Text("Hidden too")),
		components.Text("Visible"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "Visible" {
		t.Fatalf("expected only visible screen-reader output, got %q", sections.Output)
	}
}

func TestRenderScreenReaderDefaultBoxDirectionUsesSpaces(t *testing.T) {
	root := components.Box(vdom.Props{},
		components.Text("Hello"),
		components.Text("World"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "Hello World" {
		t.Fatalf("expected default box screen-reader output to be row-based, got %q", sections.Output)
	}
}

func TestRenderScreenReaderDefaultBoxDirectionPreservesRoleNarration(t *testing.T) {
	root := components.Box(vdom.Props{"aria-role": "button"},
		components.Text("Click"),
		components.Text("me"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "button: Click me" {
		t.Fatalf("expected role narration with default row separator, got %q", sections.Output)
	}
}

func TestRenderScreenReaderAppliesRoleAndState(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":     "list",
		"flexDirection": "column",
	},
		components.Text("Select a color:"),
		components.Box(vdom.Props{
			"aria-label": "1. Red",
			"aria-role":  "listitem",
		}),
		components.Box(vdom.Props{
			"aria-label": "2. Green",
			"aria-role":  "listitem",
			"aria-state": vdom.Props{"selected": true},
		}),
	)

	sections := renderer.RenderScreenReaderSections(root)
	expected := "list: Select a color:\nlistitem: 1. Red\nlistitem: (selected) 2. Green"
	if sections.Output != expected {
		t.Fatalf("expected %q, got %q", expected, sections.Output)
	}
}

func TestRenderScreenReaderAcceptsBoolStateMaps(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":     "listbox",
		"aria-state":    map[string]bool{"multiselectable": true},
		"flexDirection": "column",
	},
		components.Box(vdom.Props{
			"aria-role":  "option",
			"aria-state": map[string]bool{"selected": true},
		}, components.Text("Option 1")),
		components.Box(vdom.Props{
			"aria-role":  "option",
			"aria-state": map[string]bool{"selected": false},
		}, components.Text("Option 2")),
	)

	sections := renderer.RenderScreenReaderSections(root)
	expected := "listbox: (multiselectable) option: (selected) Option 1\noption: Option 2"
	if sections.Output != expected {
		t.Fatalf("expected %q, got %q", expected, sections.Output)
	}
}

func TestRenderScreenReaderSplitsStaticOutput(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.StaticItems([]string{"A", "B"}, func(item string, index int) *vdom.Node {
			return components.Text(item)
		}),
		components.Text("X"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "X" {
		t.Fatalf("expected dynamic output only, got %q", sections.Output)
	}
	if sections.StaticOutput != "A\nB\n" {
		t.Fatalf("expected static output with trailing newline, got %q", sections.StaticOutput)
	}
}

func TestRenderScreenReaderSkipsNilChildren(t *testing.T) {
	root := components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("Hello"),
		nil,
		components.Text("World"),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "Hello\nWorld" {
		t.Fatalf("expected nil child to be ignored, got %q", sections.Output)
	}
}

func TestRenderScreenReaderTransformAccessibilityLabel(t *testing.T) {
	root := vdom.CreateElement("transform", vdom.Props{
		"transform":          func(children string, index int) string { return strings.ToUpper(children) },
		"accessibilityLabel": "spoken value",
	}, components.Text("hidden visual value"))

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "spoken value" {
		t.Fatalf("expected transform accessibility label output, got %q", sections.Output)
	}
}

func TestRenderScreenReaderTransformAccessibilityLabelWithinRole(t *testing.T) {
	root := components.Box(vdom.Props{"aria-role": "button"},
		vdom.CreateElement("transform", vdom.Props{
			"transform":          func(children string, index int) string { return strings.ToUpper(children) },
			"accessibilityLabel": "spoken value",
		}, components.Text("hidden visual value")),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "button: spoken value" {
		t.Fatalf("expected role-prefixed transform accessibility label output, got %q", sections.Output)
	}
}

func TestRenderScreenReaderRoleWithoutContentKeepsPrefixSpacing(t *testing.T) {
	root := components.Box(vdom.Props{"aria-role": "button"})

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "button: " {
		t.Fatalf("expected empty role narration to retain upstream spacing, got %q", sections.Output)
	}
}

func TestRenderScreenReaderStateWithoutContentKeepsPrefixSpacing(t *testing.T) {
	root := components.Box(vdom.Props{"aria-state": vdom.Props{"busy": true}})

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "(busy) " {
		t.Fatalf("expected empty state narration to retain upstream spacing, got %q", sections.Output)
	}
}

func TestRenderScreenReaderRoleAndStateWithoutContentKeepPrefixSpacing(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":  "option",
		"aria-state": vdom.Props{"selected": true},
	})

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "option: (selected) " {
		t.Fatalf("expected empty role/state narration to retain upstream spacing, got %q", sections.Output)
	}
}

func TestRenderScreenReaderStateOrderPreservesInputOrder(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":  "option",
		"aria-state": []string{"selected", "disabled", "expanded"},
	})

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "option: (selected, disabled, expanded) " {
		t.Fatalf("expected ordered state narration, got %q", sections.Output)
	}
}

func TestRenderScreenReaderStateOrderVariesWithInputOrder(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":  "option",
		"aria-state": []string{"expanded", "selected", "disabled"},
	})

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "option: (expanded, selected, disabled) " {
		t.Fatalf("expected state narration to follow input order, got %q", sections.Output)
	}
}

func TestRenderScreenReaderNestedSameRoleViaNeutralWrapperStillNarratesInnerRole(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":     "list",
		"flexDirection": "column",
	},
		components.Box(nil,
			components.Box(vdom.Props{"aria-role": "list"}, components.Text("Nested")),
		),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "list: list: Nested" {
		t.Fatalf("expected inner role to be preserved across neutral wrapper, got %q", sections.Output)
	}
}

func TestRenderScreenReaderNestedSameRoleViaNeutralRowWrapperStillNarratesInnerRole(t *testing.T) {
	root := components.Box(vdom.Props{"aria-role": "button"},
		components.Box(nil,
			components.Box(vdom.Props{"aria-role": "button"}, components.Text("Nested")),
		),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "button: button: Nested" {
		t.Fatalf("expected nested same role through neutral row wrapper, got %q", sections.Output)
	}
}

func TestRenderScreenReaderNestedSameRoleWithStateViaNeutralWrapperStillNarratesInnerRole(t *testing.T) {
	root := components.Box(vdom.Props{
		"aria-role":     "listbox",
		"flexDirection": "column",
	},
		components.Box(nil,
			components.Box(vdom.Props{
				"aria-role":  "listbox",
				"aria-state": vdom.Props{"multiselectable": true},
			}, components.Text("Options")),
		),
	)

	sections := renderer.RenderScreenReaderSections(root)
	if sections.Output != "listbox: listbox: (multiselectable) Options" {
		t.Fatalf("expected nested same role with state through wrapper, got %q", sections.Output)
	}
}

func TestRenderWithLayoutWrapsTextToBoxWidth(t *testing.T) {
	root := components.Box(vdom.Props{"width": 7.0},
		components.Text("Hello World"),
	)

	output := renderer.RenderWithLayout(root, 20, 20)
	if output != "Hello\nWorld" {
		t.Fatalf("expected wrapped text output, got %q", output)
	}
}

func TestRenderWithLayoutWrapsTextLikeInkTrimFalse(t *testing.T) {
	root := components.Box(vdom.Props{"width": 70.0, "padding": 1.0, "flexDirection": "column"},
		components.Text("Type something and then resize your terminal (drag the edge or press Cmd/Ctrl -/+)"),
	)

	output := renderer.RenderWithLayout(root, 70, 20)
	expected := "\n Type something and then resize your terminal (drag the edge or press\n  Cmd/Ctrl -/+)\n"
	if output != expected {
		t.Fatalf("expected Ink trim=false wrapping, got %q", output)
	}
}

func TestRenderWithLayoutTruncatesTextToBoxWidth(t *testing.T) {
	root := components.Box(vdom.Props{"width": 7.0},
		components.Text(vdom.Props{"wrap": "truncate"}, "Hello World"),
	)

	output := renderer.RenderWithLayout(root, 20, 20)
	if output != "Hello …" {
		t.Fatalf("expected truncated text output, got %q", output)
	}
}

func TestRenderWithLayoutTruncatesTextInMiddle(t *testing.T) {
	root := components.Box(vdom.Props{"width": 7.0},
		components.Text(vdom.Props{"wrap": "truncate-middle"}, "Hello World"),
	)

	output := renderer.RenderWithLayout(root, 20, 20)
	if output != "Hel…rld" {
		t.Fatalf("expected middle-truncated text output, got %q", output)
	}
}

func TestRenderWithLayoutTruncatesTextAtStart(t *testing.T) {
	root := components.Box(vdom.Props{"width": 7.0},
		components.Text(vdom.Props{"wrap": "truncate-start"}, "Hello World"),
	)

	output := renderer.RenderWithLayout(root, 20, 20)
	if output != "… World" {
		t.Fatalf("expected start-truncated text output, got %q", output)
	}
}

func TestRenderWithLayoutANSITextInheritsBoxBackground(t *testing.T) {
	root := components.Box(vdom.Props{
		"backgroundColor": "green",
		"alignSelf":       "flex-start",
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[42mHello World\x1b[49m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSITextBackgroundOverrideAndClear(t *testing.T) {
	root := components.Box(vdom.Props{
		"backgroundColor": "green",
		"alignSelf":       "flex-start",
	},
		components.Text("Inherited "),
		components.Text(vdom.Props{"backgroundColor": ""}, "No BG "),
		components.Text(vdom.Props{"backgroundColor": "red"}, "Red BG"),
	)

	output := renderer.RenderWithLayoutANSI(root, 40, 20)
	expected := "\x1b[42mInherited \x1b[49mNo BG \x1b[41mRed BG\x1b[49m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBoxBackgroundFill(t *testing.T) {
	root := components.Box(vdom.Props{
		"backgroundColor": "red",
		"width":           10.0,
		"height":          3.0,
		"alignSelf":       "flex-start",
	}, components.Text("Hello"))

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[41mHello     \x1b[49m\n\x1b[41m          \x1b[49m\n\x1b[41m          \x1b[49m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIRootBackgroundUsesAvailableWidth(t *testing.T) {
	root := components.Box(vdom.Props{
		"backgroundColor": "red",
	}, components.Text("Hello"))

	output := renderer.RenderWithLayoutANSI(root, 10, 5)
	expected := "\x1b[41mHello     \x1b[49m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBoxBackgroundWithBorder(t *testing.T) {
	root := components.Box(vdom.Props{
		"backgroundColor": "cyan",
		"borderStyle":     "round",
		"width":           10.0,
		"height":          5.0,
		"alignSelf":       "flex-start",
	}, components.Text("Hi"))

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "╭────────╮\n│\x1b[46mHi      \x1b[49m│\n│\x1b[46m        \x1b[49m│\n│\x1b[46m        \x1b[49m│\n╰────────╯"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBorderColor(t *testing.T) {
	root := components.Box(vdom.Props{
		"borderStyle": "round",
		"borderColor": "green",
		"alignSelf":   "flex-start",
	}, components.Text("Hello"))

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[32m╭─────╮\x1b[39m\n\x1b[32m│\x1b[39mHello\x1b[32m│\x1b[39m\n\x1b[32m╰─────╯\x1b[39m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBorderColorUpdatesWithoutStaleState(t *testing.T) {
	root := components.Box(vdom.Props{
		"borderStyle": "round",
	}, components.Text("Hello World"))

	initial := renderer.RenderWithLayoutANSI(root, 20, 20)
	expectedInitial := "╭──────────────────╮\n│Hello World       │\n╰──────────────────╯"
	if initial != expectedInitial {
		t.Fatalf("expected initial %q, got %q", expectedInitial, initial)
	}

	root.Props["borderColor"] = "green"
	updated := renderer.RenderWithLayoutANSI(root, 20, 20)
	expectedUpdated := "\x1b[32m╭──────────────────╮\x1b[39m\n\x1b[32m│\x1b[39mHello World       \x1b[32m│\x1b[39m\n\x1b[32m╰──────────────────╯\x1b[39m"
	if updated != expectedUpdated {
		t.Fatalf("expected updated %q, got %q", expectedUpdated, updated)
	}

	delete(root.Props, "borderColor")
	reset := renderer.RenderWithLayoutANSI(root, 20, 20)
	if reset != expectedInitial {
		t.Fatalf("expected reset %q, got %q", expectedInitial, reset)
	}
}

func TestRenderWithLayoutANSIBorderTopColor(t *testing.T) {
	root := components.Box(vdom.Props{
		"flexDirection": "column",
		"alignItems":    "flex-start",
	},
		components.Text("Above"),
		components.Box(vdom.Props{"borderStyle": "round", "borderTopColor": "green"}, components.Text("Content")),
		components.Text("Below"),
	)

	output := renderer.RenderWithLayoutANSI(root, 40, 20)
	expected := "Above\n\x1b[32m╭───────╮\x1b[39m\n│Content│\n╰───────╯\nBelow"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBorderBottomColor(t *testing.T) {
	root := components.Box(vdom.Props{
		"flexDirection": "column",
		"alignItems":    "flex-start",
	},
		components.Text("Above"),
		components.Box(vdom.Props{"borderStyle": "round", "borderBottomColor": "green"}, components.Text("Content")),
		components.Text("Below"),
	)

	output := renderer.RenderWithLayoutANSI(root, 40, 20)
	expected := "Above\n╭───────╮\n│Content│\n\x1b[32m╰───────╯\x1b[39m\nBelow"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBorderLeftColor(t *testing.T) {
	root := components.Box(vdom.Props{
		"flexDirection": "column",
		"alignItems":    "flex-start",
	},
		components.Text("Above"),
		components.Box(vdom.Props{"borderStyle": "round", "borderLeftColor": "green"}, components.Text("Content")),
		components.Text("Below"),
	)

	output := renderer.RenderWithLayoutANSI(root, 40, 20)
	expected := "Above\n╭───────╮\n\x1b[32m│\x1b[39mContent│\n╰───────╯\nBelow"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBorderRightColor(t *testing.T) {
	root := components.Box(vdom.Props{
		"flexDirection": "column",
		"alignItems":    "flex-start",
	},
		components.Text("Above"),
		components.Box(vdom.Props{"borderStyle": "round", "borderRightColor": "green"}, components.Text("Content")),
		components.Text("Below"),
	)

	output := renderer.RenderWithLayoutANSI(root, 40, 20)
	expected := "Above\n╭───────╮\n│Content\x1b[32m│\x1b[39m\n╰───────╯\nBelow"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBorderDimKeepsChildStyle(t *testing.T) {
	root := components.Box(vdom.Props{
		"borderStyle":    "round",
		"borderDimColor": true,
		"alignSelf":      "flex-start",
	}, components.Text(vdom.Props{"bold": true, "color": "blue"}, "styled text"))

	output := renderer.RenderWithLayoutANSI(root, 40, 20)
	if !strings.Contains(output, "\x1b[2m╭───────────╮\x1b[22m") {
		t.Fatalf("expected dimmed top border, got %q", output)
	}
	if strings.Contains(output, "\x1b[2m\x1b[1m\x1b[34mstyled text") {
		t.Fatalf("expected child text not to inherit dim border style, got %q", output)
	}
	if !strings.Contains(output, "\x1b[1m\x1b[34mstyled text\x1b[39m\x1b[22m") {
		t.Fatalf("expected child text styling to remain non-dimmed, got %q", output)
	}
	expected := "\x1b[2m╭───────────╮\x1b[22m\n\x1b[2m│\x1b[22m\x1b[1m\x1b[34mstyled text\x1b[39m\x1b[22m\x1b[2m│\x1b[22m\n\x1b[2m╰───────────╯\x1b[22m"
	if output != expected {
		t.Fatalf("expected exact border/style transition %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIBorderVariationSelectorEmojiWidth(t *testing.T) {
	root := components.Box(vdom.Props{
		"borderStyle": "round",
		"alignSelf":   "flex-start",
	}, components.Text("🌡️⚠️✅"))

	output := renderer.RenderWithLayout(root, 40, 20)
	expected := "╭──────╮\n│🌡️⚠️✅│\n╰──────╯"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIStyledVariationSelectorEmojiWidth(t *testing.T) {
	root := components.Box(vdom.Props{
		"borderStyle": "round",
		"alignSelf":   "flex-start",
	}, components.Text(vdom.Props{"color": "green"}, "🌡️⚠️✅"))

	output := renderer.RenderWithLayoutANSI(root, 40, 20)
	expected := "╭──────╮\n│\x1b[32m🌡️⚠️✅\x1b[39m│\n╰──────╯"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSITruncatesVariationSelectorEmojiByDisplayWidth(t *testing.T) {
	root := components.Box(vdom.Props{"width": 5.0},
		components.Text(vdom.Props{"color": "green", "wrap": "truncate"}, "🌡️⚠️✅"),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[32m🌡️⚠️…\x1b[39m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSITextStyleModifiers(t *testing.T) {
	root := components.Text(vdom.Props{
		"color":           "red",
		"backgroundColor": "green",
		"bold":            true,
		"italic":          true,
		"underline":       true,
		"inverse":         true,
		"strikethrough":   true,
	}, "Hello")

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[1m\x1b[31m\x1b[42m\x1b[3m\x1b[4m\x1b[7m\x1b[9mHello\x1b[22m\x1b[23m\x1b[24m\x1b[27m\x1b[29m\x1b[39m\x1b[49m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIDimmedColorOrderMatchesInk(t *testing.T) {
	root := components.Text(vdom.Props{
		"color":    "green",
		"dimColor": true,
	}, "Test")

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[32m\x1b[2mTest\x1b[22m\x1b[39m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSINestedTextColorOverrideResumesParent(t *testing.T) {
	root := components.Text(
		vdom.Props{"color": "green"},
		"A ",
		components.Text(vdom.Props{"color": "blue"}, "B"),
		" C",
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[32mA \x1b[34mB\x1b[32m C\x1b[39m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSINestedTextBackgroundClearResumesParent(t *testing.T) {
	root := components.Box(vdom.Props{
		"backgroundColor": "green",
		"alignSelf":       "flex-start",
	},
		components.Text(
			"A ",
			components.Text(vdom.Props{"backgroundColor": ""}, "B"),
			" C",
		),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[42mA \x1b[49mB\x1b[42m C\x1b[49m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSINestedTextBackgroundOverrideResumesParent(t *testing.T) {
	root := components.Text(
		vdom.Props{"backgroundColor": "blue"},
		"A ",
		components.Text(vdom.Props{"backgroundColor": "red"}, "B"),
		" C",
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[44mA \x1b[41mB\x1b[44m C\x1b[49m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSINestedTextTruncateEndKeepsStyleTransitions(t *testing.T) {
	root := components.Box(vdom.Props{"width": 4.0},
		components.Text(
			vdom.Props{"color": "green", "wrap": "truncate"},
			"A ",
			components.Text(vdom.Props{"color": "blue"}, "BC"),
			"D",
		),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[32mA \x1b[34mB…\x1b[39m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSINestedTextTruncateStartKeepsBackgroundTransitions(t *testing.T) {
	root := components.Box(vdom.Props{"width": 4.0},
		components.Text(
			vdom.Props{"backgroundColor": "blue", "wrap": "truncate-start"},
			"AB",
			components.Text(vdom.Props{"backgroundColor": "red"}, "CD"),
			"E",
		),
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[41m…CD\x1b[44mE\x1b[49m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSINestedTransformWrapKeepsParentAndChildStyles(t *testing.T) {
	root := components.Text(
		vdom.Props{"color": "green"},
		"A",
		components.Transform(func(children string, index int) string {
			return "<" + children + ">"
		}, components.Text(vdom.Props{"color": "blue"}, "B")),
		"C",
	)

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "\x1b[32mA<\x1b[34mB\x1b[32m>C\x1b[39m"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestRenderWithLayoutANSIRootTransformWrapKeepsChildStyle(t *testing.T) {
	root := components.Transform(func(children string, index int) string {
		return "<" + children + ">"
	}, components.Text(vdom.Props{"color": "blue"}, "B"))

	output := renderer.RenderWithLayoutANSI(root, 20, 20)
	expected := "<\x1b[34mB\x1b[39m>"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}
