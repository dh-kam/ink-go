package renderer_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/internal/renderer"
	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/layout"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TestRenderWithLayout tests rendering with layout positioning
func TestRenderWithLayout(t *testing.T) {
	// Create a box with specific position
	box := vdom.CreateElement("box", vdom.Props{
		"width":  100.0,
		"height": 50.0,
	})

	text := vdom.CreateTextNode("Positioned Text")
	box.Children = append(box.Children, text)

	output := renderer.RenderWithLayout(box, 200, 100)

	if !strings.Contains(output, "Positioned Text") {
		t.Errorf("Expected output to contain text, got:\n%s", output)
	}
}

// TestRenderBoxWithPadding tests box with padding
func TestRenderBoxWithPadding(t *testing.T) {
	box := vdom.CreateElement("box", vdom.Props{
		"width":   100.0,
		"height":  50.0,
		"padding": 10.0,
	})

	text := vdom.CreateTextNode("X")
	box.Children = append(box.Children, text)

	output := renderer.RenderWithLayout(box, 200, 100)

	// Text should be positioned at padding offset
	lines := strings.Split(output, "\n")

	// Check that there's whitespace before the X (padding)
	found := false
	for _, line := range lines {
		if strings.Contains(line, "X") {
			// Should have at least some leading spaces
			trimmed := strings.TrimLeft(line, " ")
			if len(line) > len(trimmed) {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Expected text to have padding offset")
	}
}

// TestRenderFlexRow tests flex row layout
func TestRenderFlexRow(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": layout.FlexDirectionRow,
		"width":         200.0,
		"height":        50.0,
	})

	child1 := vdom.CreateElement("box", vdom.Props{
		"width":  50.0,
		"height": 30.0,
	})
	child1.Children = append(child1.Children, vdom.CreateTextNode("A"))

	child2 := vdom.CreateElement("box", vdom.Props{
		"width":  50.0,
		"height": 30.0,
	})
	child2.Children = append(child2.Children, vdom.CreateTextNode("B"))

	container.Children = append(container.Children, child1, child2)

	output := renderer.RenderWithLayout(container, 200, 100)

	// Both A and B should be in the output, on the same line (row layout)
	if !strings.Contains(output, "A") || !strings.Contains(output, "B") {
		t.Errorf("Expected output to contain both A and B, got:\n%s", output)
	}
}

// TestRenderFlexColumn tests flex column layout
func TestRenderFlexColumn(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": layout.FlexDirectionColumn,
		"width":         100.0,
		"height":        100.0,
	})

	child1 := vdom.CreateElement("box", vdom.Props{
		"width":  50.0,
		"height": 20.0,
	})
	child1.Children = append(child1.Children, vdom.CreateTextNode("Top"))

	child2 := vdom.CreateElement("box", vdom.Props{
		"width":  50.0,
		"height": 20.0,
	})
	child2.Children = append(child2.Children, vdom.CreateTextNode("Bottom"))

	container.Children = append(container.Children, child1, child2)

	output := renderer.RenderWithLayout(container, 200, 100)

	lines := strings.Split(output, "\n")

	// Find which lines contain Top and Bottom
	topLine := -1
	bottomLine := -1

	for i, line := range lines {
		if strings.Contains(line, "Top") {
			topLine = i
		}
		if strings.Contains(line, "Bottom") {
			bottomLine = i
		}
	}

	// Bottom should be on a later line than Top
	if topLine == -1 || bottomLine == -1 {
		t.Errorf("Expected to find both Top and Bottom in output:\n%s", output)
	} else if bottomLine <= topLine {
		t.Error("Expected Bottom to be on a line after Top (column layout)")
	}
}

func TestRenderTextElementsDoNotOverlapInRowLayout(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": layout.FlexDirectionRow,
		"width":         20.0,
		"height":        1.0,
	}, components.Text("A"), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 2)

	if !strings.Contains(output, "AB") {
		t.Errorf("Expected row text elements to render sequentially, got:\n%s", output)
	}
}

func TestRenderSpacerInRowLayout(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": layout.FlexDirectionRow,
		"width":         10.0,
		"height":        1.0,
	}, components.Text("L"), components.Spacer(), components.Text("R"))

	output := renderer.RenderWithLayout(container, 10, 2)
	lines := strings.Split(output, "\n")

	found := false
	for _, line := range lines {
		if strings.Contains(line, "L") && strings.Contains(line, "R") {
			if strings.Index(line, "R") > strings.Index(line, "L")+1 {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected spacer to push right text, got:\n%s", output)
	}
}

func TestRenderMarginAtRoot(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"marginTop":  1.0,
		"marginLeft": 2.0,
	}, components.Text("X"))

	output := renderer.RenderWithLayout(container, 10, 10)
	expected := "\n  X"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderAlignItemsCenter(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": layout.FlexDirectionRow,
		"alignItems":    "center",
		"height":        3.0,
	}, components.Text("X"))

	output := renderer.RenderWithLayout(container, 10, 10)
	expected := "\nX\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderAlignItemsCenterWithMultipleTextNodes(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"alignItems": "center",
		"height":     3.0,
	}, components.Text("A"), components.Text("B"))

	output := renderer.RenderWithLayout(container, 10, 10)
	expected := "\nAB\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderAlignSelfCenter(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"height": 3.0,
	}, components.Box(vdom.Props{
		"alignSelf": "center",
	}, components.Text("Test")))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\nTest\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderAlignSelfCenterWithMultipleTextNodes(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"height": 3.0,
	}, components.Box(vdom.Props{
		"alignSelf": "center",
	}, components.Text("A"), components.Text("B")))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\nAB\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderAlignSelfEndInColumn(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"width":         10.0,
	}, components.Box(vdom.Props{
		"alignSelf": "flex-end",
	}, components.Text("Test")))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "      Test"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDisplayNone(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
	},
		components.Box(vdom.Props{"display": "none"}, components.Text("Kitty!")),
		components.Text("Doggo"),
	)

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "Doggo"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexDirectionRowReverse(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "row-reverse",
		"width":         4.0,
	}, components.Text("A"), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "  BA"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexDirectionColumnReverse(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column-reverse",
		"height":        4.0,
	}, components.Text("A"), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\n\nB\nA"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderGapWrap(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"gap":      1.0,
		"width":    3.0,
		"flexWrap": "wrap",
	}, components.Text("A"), components.Text("B"), components.Text("C"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A B\n\nC"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderColumnGap(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"columnGap": 1.0,
	}, components.Text("A"), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRowGap(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"rowGap":        1.0,
	}, components.Text("A"), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A\n\nB"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRowGapWrap(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"columnGap": 1.0,
		"rowGap":    2.0,
		"width":     3.0,
		"flexWrap":  "wrap",
	}, components.Text("A"), components.Text("B"), components.Text("C"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A B\n\n\nC"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderColumnGapWrap(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"columnGap":     2.0,
		"height":        2.0,
		"flexWrap":      "wrap",
	}, components.Text("A"), components.Text("B"), components.Text("C"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A  C\nB"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexWrapRow(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":    2.0,
		"flexWrap": "wrap",
	}, components.Text("A"), components.Text("BC"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A\nBC"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexWrapColumnNoWrap(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"height":        2.0,
	}, components.Text("A"), components.Text("B"), components.Text("C"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "B\nC"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexWrapColumn(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"height":        2.0,
		"flexWrap":      "wrap",
	}, components.Text("A"), components.Text("B"), components.Text("C"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "AC\nB"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexWrapColumnReverse(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"height":        2.0,
		"width":         3.0,
		"flexWrap":      "wrap-reverse",
	}, components.Text("A"), components.Text("B"), components.Text("C"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := " CA\n  B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexWrapRowReverse(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"height":   3.0,
		"width":    2.0,
		"flexWrap": "wrap-reverse",
	}, components.Text("A"), components.Text("B"), components.Text("C"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\nC\nAB"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderJustifySpaceAround(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"justifyContent": "space-around",
		"width":          5.0,
	}, components.Text("A"), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A  B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderJustifySpaceEvenlyRoundsStartLikeYoga(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"justifyContent": "space-evenly",
		"width":          10.0,
	}, components.Text("A"), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "  A   B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBorderJustifyCenterRoundsStartLikeYoga(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"borderStyle":    "round",
		"justifyContent": "center",
		"width":          20.0,
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 30, 10)
	expected := "╭──────────────────╮\n│   Hello World    │\n╰──────────────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBorderAlignCenterRoundsStartLikeYoga(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"alignItems":  "center",
		"alignSelf":   "flex-start",
		"borderStyle": "round",
		"height":      20.0,
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 30, 30)
	expected := "╭───────────╮\n│           │\n│           │\n│           │\n│           │\n│           │\n│           │\n│           │\n│           │\n│Hello World│\n│           │\n│           │\n│           │\n│           │\n│           │\n│           │\n│           │\n│           │\n│           │\n╰───────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRootUsesAvailableWidthForJustifyContent(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"justifyContent": "center",
	}, components.Text("Test"))

	output := renderer.RenderWithLayout(container, 10, 5)
	expected := "   Test"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexBasisRow(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 6.0,
	}, components.Box(vdom.Props{
		"flexBasis": 3.0,
	}, components.Text("A")), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A  B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexBasisColumnPercent(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"height":        6.0,
	}, components.Box(vdom.Props{
		"flexBasis": "50%",
	}, components.Text("A")), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A\n\n\nB\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderWidthPercentRow(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	}, components.Box(vdom.Props{
		"width": "50%",
	}, components.Text("A")), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A    B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderWidthPercentRoundsHalfColumnLikeYoga(t *testing.T) {
	container := components.Box(vdom.Props{"flexDirection": "column"},
		components.Box(vdom.Props{
			"width":       "50%",
			"borderStyle": "single",
		}, components.Text("X")),
	)

	output := renderer.RenderWithLayout(container, 127, 20)
	topBorder := strings.Split(output, "\n")[0]
	if got := len([]rune(topBorder)); got != 64 {
		t.Fatalf("expected 50%% of 127 columns to render as 64 columns like Yoga, got %d in %q", got, topBorder)
	}
}

func TestRenderPaddingAppliesToTextWithNewlines(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"padding": 1.0,
	}, components.Text("Hello\nWorld"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\n Hello\n World\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderPaddingEdgeOverridesShorthandLikeYoga(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"padding":     1.0,
		"paddingLeft": 3.0,
	}, components.Text("X"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\n   X\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderPaddingAppliesToWrappedText(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"padding": 1.0,
		"width":   5.0,
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\n Hel\n lo\n Wor\n ld\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderMarginAppliesToTextWithNewlines(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"margin": 1.0,
	}, components.Text("Hello\nWorld"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\n Hello\n World\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderMarginAppliesToWrappedText(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"margin": 1.0,
		"width":  6.0,
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\n Hello\n World\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderMinWidthRow(t *testing.T) {
	container := vdom.CreateElement("box", nil,
		components.Box(vdom.Props{"minWidth": 5.0}, components.Text("A")),
		components.Text("B"),
	)

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A    B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderMinWidthPercentRow(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	},
		components.Box(vdom.Props{"minWidth": "50%"}, components.Text("A")),
		components.Text("B"),
	)

	output := renderer.RenderWithLayout(container, 100, 20)
	expected := "A                                                 B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderHeightPercentColumn(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"height":        6.0,
	}, components.Box(vdom.Props{
		"height": "50%",
	}, components.Text("A")), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A\n\n\nB\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderHeightCutsText(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"height": 2.0,
	}, components.Text("AAAABBBBCCCC"))

	output := renderer.RenderWithLayout(container, 4, 20)
	expected := "AAAA\nBBBB"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderMinHeightSmall(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"minHeight": 4.0,
	}, components.Text("A"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A\n\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderMinHeightLarge(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"minHeight": 2.0,
	}, components.Box(vdom.Props{
		"height": 4.0,
	}, components.Text("A")))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A\n\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderWideCharacterInsideFixedWidthBox(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
	},
		vdom.CreateElement("box", nil,
			vdom.CreateElement("box", vdom.Props{"width": 2.0}, components.Text("🍔")),
			components.Text("|"),
		),
		vdom.CreateElement("box", nil,
			vdom.CreateElement("box", vdom.Props{"width": 2.0}, components.Text("⏳")),
			components.Text("|"),
		),
	)

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "🍔|\n⏳|"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexShrinkNone(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 16.0,
	}, components.Box(vdom.Props{
		"flexShrink": 0.0,
		"width":      6.0,
	}, components.Text("A")), components.Box(vdom.Props{
		"flexShrink": 0.0,
		"width":      6.0,
	}, components.Text("B")), components.Box(vdom.Props{
		"width": 6.0,
	}, components.Text("C")))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A     B     C"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderFlexShrinkEqual(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	}, components.Box(vdom.Props{
		"flexShrink": 1.0,
		"width":      6.0,
	}, components.Text("A")), components.Box(vdom.Props{
		"flexShrink": 1.0,
		"width":      6.0,
	}, components.Text("B")), components.Text("C"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A    B   C"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderIntegerFlexGrowProp(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 6,
	}, components.Box(vdom.Props{
		"flexGrow": 1,
	}, components.Text("A")), components.Text("B"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A    B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderIntegerMarginProp(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"margin": 1,
	}, components.Text("X"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\n X\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderIntegerPaddingPropWrapsText(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"padding": 1,
		"width":   5,
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "\n Hel\n lo\n Wor\n ld\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderSkipsNilChildrenInLayout(t *testing.T) {
	container := components.Box(vdom.Props{
		"gap": 1.0,
	},
		components.Text("A"),
		nil,
		components.Text("B"),
	)

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "A B"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXHiddenSingleText(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":     6.0,
		"overflowX": "hidden",
	}, components.Box(vdom.Props{
		"width":      16.0,
		"flexShrink": 0.0,
	}, components.Text("Hello World")))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "Hello"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXHiddenLeftIntersection(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":     6.0,
		"overflowX": "hidden",
	}, components.Box(vdom.Props{
		"marginLeft": -3.0,
		"width":      12.0,
		"flexShrink": 0.0,
	}, components.Text("Hello World")))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "lo Wor"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenSingleText(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"height":    1.0,
		"overflowY": "hidden",
	}, components.Text("Hello\nWorld"))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "Hello"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowHiddenMultiAxis(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":    6.0,
		"height":   1.0,
		"overflow": "hidden",
	}, components.Box(vdom.Props{
		"width":      12.0,
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld")))

	output := renderer.RenderWithLayout(container, 20, 20)
	expected := "Hello"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRoundBorderFitContent(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"borderStyle": "round",
		"alignSelf":   "flex-start",
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭───────────╮\n│Hello World│\n╰───────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRoundBorderFixedWidthWrap(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"borderStyle": "round",
		"width":       10.0,
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭────────╮\n│Hello   │\n│World   │\n╰────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRoundBorderFixedWidthWrapsMultipleTextChildrenLikeUpstream(t *testing.T) {
	testCases := []struct {
		name     string
		children []any
		expected string
	}{
		{
			name:     "long-first-node",
			children: []any{"Helloooooo", " World"},
			expected: "╭────────╮\n│Helloooo│\n│oo World│\n╰────────╯",
		},
		{
			name:     "very-long-first-node",
			children: []any{"Hellooooooooooooo", " World"},
			expected: "╭────────╮\n│Helloooo│\n│oooooooo│\n│o World │\n╰────────╯",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			container := vdom.CreateElement("box", vdom.Props{
				"borderStyle": "round",
				"width":       10.0,
			}, components.Text(testCase.children...))

			output := renderer.RenderWithLayout(container, 40, 20)
			if output != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, output)
			}
		})
	}
}

func TestRenderRoundBorderWithPadding(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"borderStyle": "round",
		"padding":     1.0,
		"alignSelf":   "flex-start",
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭─────────────╮\n│             │\n│ Hello World │\n│             │\n╰─────────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRoundBorderFixedWidth(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"borderStyle": "round",
		"width":       20.0,
	}, components.Text("Hello World"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭──────────────────╮\n│Hello World       │\n╰──────────────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRoundBorderHideTop(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"alignItems":    "flex-start",
	},
		components.Text("Above"),
		components.Box(vdom.Props{"borderStyle": "round", "borderTop": false}, components.Text("Content")),
		components.Text("Below"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "Above\n│Content│\n╰───────╯\nBelow"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRoundBorderHideLeftRight(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"alignItems":    "flex-start",
	},
		components.Text("Above"),
		components.Box(vdom.Props{
			"borderStyle": "round",
			"borderLeft":  false,
			"borderRight": false,
		}, components.Text("Content")),
		components.Text("Below"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "Above\n───────\nContent\n───────\nBelow"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXHiddenWithBorder(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       6.0,
		"overflowX":   "hidden",
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"width":      16.0,
		"flexShrink": 0.0,
	}, components.Text("Hello World")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭────╮\n│Hell│\n╰────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenWithBorder(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       20.0,
		"height":      3.0,
		"overflowY":   "hidden",
		"borderStyle": "round",
	}, components.Text("Hello\nWorld"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭──────────────────╮\n│Hello             │\n╰──────────────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBorderDrawsBeforeNegativeMarginChild(t *testing.T) {
	container := components.Box(vdom.Props{
		"width":       6.0,
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"marginLeft": -1.0,
		"width":      4.0,
		"height":     1.0,
		"flexShrink": 0.0,
	}, components.Text("ABCD")))

	expected := "╭────╮\nABCD │\n╰────╯"
	assertLayoutAndANSIOutput(t, container, expected)
}

func TestRenderOverflowHiddenWithBorderClipsChildrenToContent(t *testing.T) {
	container := components.Box(vdom.Props{
		"width":       6.0,
		"height":      3.0,
		"overflow":    "hidden",
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"marginLeft": -1.0,
		"marginTop":  -1.0,
		"width":      6.0,
		"height":     3.0,
		"flexShrink": 0.0,
	}, components.Text("AAAAAA\nBBBBBB\nCCCCCC")))

	expected := "╭────╮\n│BBBB│\n╰────╯"
	assertLayoutAndANSIOutput(t, container, expected)
}

func assertLayoutAndANSIOutput(t *testing.T, node *vdom.Node, expected string) {
	t.Helper()

	if output := renderer.RenderWithLayout(node, 40, 20); output != expected {
		t.Fatalf("RenderWithLayout expected %q, got %q", expected, output)
	}

	if output := renderer.RenderWithLayoutANSI(node, 40, 20); output != expected {
		t.Fatalf("RenderWithLayoutANSI expected %q, got %q", expected, output)
	}
}

func TestRenderRoundBorderHideBottom(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"alignItems":    "flex-start",
	},
		components.Text("Above"),
		components.Box(vdom.Props{"borderStyle": "round", "borderBottom": false}, components.Text("Content")),
		components.Text("Below"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "Above\n╭───────╮\n│Content│\nBelow"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderTopOnlyBorderStretchesEmptyChildAcrossParentWidth(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":         100.0,
		"flexDirection": "column",
	},
		components.Box(vdom.Props{
			"borderStyle":  "single",
			"borderTop":    true,
			"borderBottom": false,
			"borderLeft":   false,
			"borderRight":  false,
		}),
		components.Box(nil, components.Text("  Showing detailed transcript · ctrl+o to toggle")),
	)

	output := renderer.RenderWithLayout(container, 120, 20)
	expected := "────────────────────────────────────────────────────────────────────────────────────────────────────\n  Showing detailed transcript · ctrl+o to toggle"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderRoundBorderHideAll(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"flexDirection": "column",
		"alignItems":    "flex-start",
	},
		components.Text("Above"),
		components.Box(vdom.Props{
			"borderStyle":  "round",
			"borderTop":    false,
			"borderBottom": false,
			"borderLeft":   false,
			"borderRight":  false,
		}, components.Text("Content")),
		components.Text("Below"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "Above\nContent\nBelow"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderCustomBorderStyle(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"alignSelf": "flex-start",
		"borderStyle": vdom.Props{
			"topLeft":     "↘",
			"top":         "↓",
			"topRight":    "↙",
			"left":        "→",
			"bottomLeft":  "↗",
			"bottom":      "↑",
			"bottomRight": "↖",
			"right":       "←",
		},
	}, components.Text("Content"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "↘↓↓↓↓↓↓↓↙\n→Content←\n↗↑↑↑↑↑↑↑↖"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBackgroundFixedArea(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"backgroundColor": "red",
		"width":           10.0,
		"height":          3.0,
		"alignSelf":       "flex-start",
	}, components.Text("Hello"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "Hello\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBackgroundWithBorder(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"backgroundColor": "cyan",
		"borderStyle":     "round",
		"width":           10.0,
		"height":          5.0,
		"alignSelf":       "flex-start",
	}, components.Text("Hi"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭────────╮\n│Hi      │\n│        │\n│        │\n╰────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBackgroundWithPadding(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"backgroundColor": "magenta",
		"padding":         1.0,
		"width":           10.0,
		"height":          5.0,
		"alignSelf":       "flex-start",
	}, components.Text("Hi"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "\n Hi\n\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBackgroundWithCenterAlignment(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"backgroundColor": "blue",
		"width":           10.0,
		"height":          3.0,
		"justifyContent":  "center",
		"alignSelf":       "flex-start",
	}, components.Text("Hi"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "    Hi\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBackgroundColumnLayout(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"backgroundColor": "green",
		"flexDirection":   "column",
		"width":           10.0,
		"height":          5.0,
		"alignSelf":       "flex-start",
	}, components.Text("Line 1"), components.Text("Line 2"))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "Line 1\nLine 2\n\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXHiddenChildBorder(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":     6.0,
		"overflowX": "hidden",
	}, components.Box(vdom.Props{
		"width":       16.0,
		"flexShrink":  0.0,
		"borderStyle": "round",
	}, components.Text("Hello World")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭─────\n│Hello\n╰─────"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXBeforeLeftEdge(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":     6.0,
		"overflowX": "hidden",
	}, components.Box(vdom.Props{
		"marginLeft": -12.0,
		"width":      6.0,
		"flexShrink": 0.0,
	}, components.Text("Hello")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := ""

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXAfterRightEdge(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":     6.0,
		"overflowX": "hidden",
	}, components.Box(vdom.Props{
		"marginLeft": 6.0,
		"width":      6.0,
		"flexShrink": 0.0,
	}, components.Text("Hello")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := ""

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXRightIntersection(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":     6.0,
		"overflowX": "hidden",
	}, components.Box(vdom.Props{
		"marginLeft": 3.0,
		"width":      6.0,
		"flexShrink": 0.0,
	}, components.Text("Hello")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "   Hel"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXHiddenMultipleTextBorderContainer(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       8.0,
		"overflowX":   "hidden",
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"width":      12.0,
		"flexShrink": 0.0,
	}, components.Text("Hello "), components.Text("World")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭──────╮\n│Hello │\n╰──────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenMultipleBoxesBorderContainer(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":         9.0,
		"height":        4.0,
		"overflowY":     "hidden",
		"flexDirection": "column",
		"borderStyle":   "round",
	},
		components.Box(vdom.Props{"flexShrink": 0.0}, components.Text("Line #1")),
		components.Box(vdom.Props{"flexShrink": 0.0}, components.Text("Line #2")),
		components.Box(vdom.Props{"flexShrink": 0.0}, components.Text("Line #3")),
		components.Box(vdom.Props{"flexShrink": 0.0}, components.Text("Line #4")),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭───────╮\n│Line #1│\n│Line #2│\n╰───────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenTopIntersectionBorderContainer(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       7.0,
		"height":      3.0,
		"overflowY":   "hidden",
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"marginTop":  -1.0,
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭─────╮\n│World│\n╰─────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenAboveTop(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"height":    1.0,
		"overflowY": "hidden",
	}, components.Box(vdom.Props{
		"marginTop":  -2.0,
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := ""

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenBelowBottom(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"height":    1.0,
		"overflowY": "hidden",
	}, components.Box(vdom.Props{
		"marginTop":  1.0,
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := ""

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenBottomIntersection(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"height":    1.0,
		"overflowY": "hidden",
	}, components.Box(vdom.Props{
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "Hello"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenAboveTopBorderContainer(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       7.0,
		"height":      3.0,
		"overflowY":   "hidden",
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"marginTop":  -3.0,
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭─────╮\n│     │\n╰─────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenBelowBottomBorderContainer(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       7.0,
		"height":      3.0,
		"overflowY":   "hidden",
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"marginTop":  2.0,
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭─────╮\n│     │\n╰─────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowYHiddenBottomIntersectionBorderContainer(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       7.0,
		"height":      3.0,
		"overflowY":   "hidden",
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭─────╮\n│Hello│\n╰─────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowHiddenSingleTextBorderContainer(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"paddingBottom": 1.0,
	}, components.Box(vdom.Props{
		"width":       8.0,
		"height":      3.0,
		"overflow":    "hidden",
		"borderStyle": "round",
	}, components.Box(vdom.Props{
		"width":      12.0,
		"height":     2.0,
		"flexShrink": 0.0,
	}, components.Text("Hello\nWorld"))))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭──────╮\n│Hello │\n╰──────╯\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowHiddenMultiBoxesBorderContainer(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"paddingBottom": 1.0,
	}, components.Box(vdom.Props{
		"width":       6.0,
		"height":      3.0,
		"overflow":    "hidden",
		"borderStyle": "round",
	},
		components.Box(vdom.Props{"width": 2.0, "height": 2.0, "flexShrink": 0.0}, components.Text("TL\nBL")),
		components.Box(vdom.Props{"width": 2.0, "height": 2.0, "flexShrink": 0.0}, components.Text("TR\nBR")),
	))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭────╮\n│TLTR│\n╰────╯\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowXHiddenMultipleTextChildBorder(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":     8.0,
		"overflowX": "hidden",
	}, components.Box(vdom.Props{
		"width":       12.0,
		"flexShrink":  0.0,
		"borderStyle": "round",
	}, components.Text("Hello "), components.Text("World")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭───────\n│HelloWo\n│\n╰───────"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBoxDefaultFlexShrinkEqual(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	},
		components.Box(vdom.Props{"width": 6.0}, components.Text("A")),
		components.Box(vdom.Props{"width": 6.0}, components.Text("B")),
		components.Text("C"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "A    B   C"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderBoxDefaultFlexShrinkEqualBordered(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       12.0,
		"borderStyle": "round",
	},
		components.Box(vdom.Props{"width": 6.0}, components.Text("A")),
		components.Box(vdom.Props{"width": 6.0}, components.Text("B")),
		components.Text("C"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭──────────╮\n│A    B   C│\n╰──────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowHiddenTopLeftIntersection(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":    4.0,
		"height":   4.0,
		"overflow": "hidden",
	}, components.Box(vdom.Props{
		"marginTop":  -2.0,
		"marginLeft": -2.0,
		"width":      4.0,
		"height":     4.0,
		"flexShrink": 0.0,
	}, components.Text("AAAA\nBBBB\nCCCC\nDDDD")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "CC\nDD\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowHiddenTopRightIntersection(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":    4.0,
		"height":   4.0,
		"overflow": "hidden",
	}, components.Box(vdom.Props{
		"marginTop":  -2.0,
		"marginLeft": 2.0,
		"width":      4.0,
		"height":     4.0,
		"flexShrink": 0.0,
	}, components.Text("AAAA\nBBBB\nCCCC\nDDDD")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "  CC\n  DD\n\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowHiddenBottomLeftIntersection(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":    4.0,
		"height":   4.0,
		"overflow": "hidden",
	}, components.Box(vdom.Props{
		"marginTop":  2.0,
		"marginLeft": -2.0,
		"width":      4.0,
		"height":     4.0,
		"flexShrink": 0.0,
	}, components.Text("AAAA\nBBBB\nCCCC\nDDDD")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "\n\nAA\nBB"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOverflowHiddenBottomRightIntersection(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":    4.0,
		"height":   4.0,
		"overflow": "hidden",
	}, components.Box(vdom.Props{
		"marginTop":  2.0,
		"marginLeft": 2.0,
		"width":      4.0,
		"height":     4.0,
		"flexShrink": 0.0,
	}, components.Text("AAAA\nBBBB\nCCCC\nDDDD")))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "\n\n  AA\n  BB"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderNestedOverflow(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"paddingBottom": 1.0,
	}, components.Box(vdom.Props{
		"width":         4.0,
		"height":        4.0,
		"overflow":      "hidden",
		"flexDirection": "column",
	},
		components.Box(vdom.Props{
			"width":    2.0,
			"height":   2.0,
			"overflow": "hidden",
		}, components.Box(vdom.Props{
			"width":      4.0,
			"height":     4.0,
			"flexShrink": 0.0,
		}, components.Text("AAAA\nBBBB\nCCCC\nDDDD"))),
		components.Box(vdom.Props{
			"width":  4.0,
			"height": 3.0,
		}, components.Text("XXXX\nYYYY\nZZZZ")),
	))

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "AA\nBB\nXXXX\nYYYY\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderOutOfBoundsWritesDoNotCrash(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       12.0,
		"height":      10.0,
		"borderStyle": "round",
	})

	output := renderer.RenderWithLayout(container, 10, 20)
	expected := "╭──────────╮\n│         │\n│         │\n│         │\n│         │\n│         │\n│         │\n│         │\n│         │\n╰──────────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkTextBoxText(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	},
		components.Text("ABCD"),
		components.Box(vdom.Props{"width": 4.0}, components.Text("EFGH")),
		components.Text("IJKL"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "ABCEFGIJKL\n   H"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkMixedBordered(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width":       9.0,
		"borderStyle": "round",
	},
		components.Text("AAAA"),
		components.Box(vdom.Props{"width": 6.0}, components.Text("B")),
		components.Text("CC"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "╭───────╮\n│AAB  CC│\n│A      │\n╰───────╯"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkBoxTextPlain(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 9.0,
	},
		components.Box(vdom.Props{"width": 6.0}, components.Text("A")),
		components.Text("BBBB"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "A    BBBB\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkTextBoxPlain(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 9.0,
	},
		components.Text("AAAA"),
		components.Box(vdom.Props{"width": 6.0}, components.Text("B")),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "AAAAB\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkTextTriple(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	},
		components.Text("ABCD"),
		components.Text("EFGH"),
		components.Text("IJKL"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "ABCEFGIJKL\n"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkBoxTextBox(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	},
		components.Box(vdom.Props{"width": 4.0}, components.Text("ABCD")),
		components.Text("EFGH"),
		components.Box(vdom.Props{"width": 4.0}, components.Text("IJKL")),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "ABCEFGHIJK\nD      L"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkTextBoxBox(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	},
		components.Text("ABCD"),
		components.Box(vdom.Props{"width": 4.0}, components.Text("EFGH")),
		components.Box(vdom.Props{"width": 4.0}, components.Text("IJKL")),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "ABCEFG IJK\n   H   L"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkBoxBoxText(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 10.0,
	},
		components.Box(vdom.Props{"width": 4.0}, components.Text("ABCD")),
		components.Box(vdom.Props{"width": 4.0}, components.Text("EFGH")),
		components.Text("IJKL"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "ABCEFGIJKL\nD  H"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestRenderDefaultFlexShrinkWideMiddle(t *testing.T) {
	container := vdom.CreateElement("box", vdom.Props{
		"width": 11.0,
	},
		components.Text("ABCD"),
		components.Box(vdom.Props{"width": 5.0}, components.Text("EFGHI")),
		components.Text("JKLM"),
	)

	output := renderer.RenderWithLayout(container, 40, 20)
	expected := "ABCEFGHJKLM\n   I"

	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}
