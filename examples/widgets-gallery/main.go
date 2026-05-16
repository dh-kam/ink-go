// widgets-gallery is a single-shot demo that renders every public widget
// shipped in pkg/components so a user can see the full catalogue at a
// glance. Each section is preceded by a Divider with a title — the title
// doubles as the widget's name in the gallery output.
package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

const (
	galleryWidth  = 60
	galleryHeight = 200
	canvasWidth   = 100
)

func main() {
	sections := []func() *vdom.Node{
		widgetTextInput,
		widgetPasswordInput,
		widgetProgressBar,
		widgetSpinner,
		widgetTable,
		widgetSelect,
		widgetMultiSelect,
		widgetConfirm,
		widgetDivider,
		widgetAlert,
		widgetTabs,
		widgetQuickSearch,
		widgetLink,
		widgetGradient,
		widgetBigText,
		widgetSyntax,
		widgetImage,
	}

	children := make([]*vdom.Node, 0, len(sections)+2)
	children = append(children, components.Text("ink-go widgets gallery"))
	children = append(children, components.Newline())
	for _, section := range sections {
		children = append(children, section())
		children = append(children, components.Newline())
	}

	root := components.Box(vdom.Props{"flexDirection": "column"}, children...)
	// We use RenderWithLayout (rather than ink.RenderToString) so the column
	// flex layout actually stacks each section on its own row — the simple
	// renderer concatenates inline text and would smush the gallery into a
	// single line. galleryHeight is set generous on purpose so every widget
	// fits vertically.
	fmt.Print(renderer.RenderWithLayout(root, canvasWidth, galleryHeight))
}

// stackTextRows takes a widget node whose direct children are text nodes
// (Select / MultiSelect / Tabs header / QuickSearch / BigText all share
// this shape) and rebuilds it as a column-flex Box so each text-node row
// renders on its own line. The renderer's flexbox layout only stacks
// Box / Static children — text-node siblings of any other element type
// get concatenated inline, which is what causes the unwrapped widgets
// to collapse into a single garbled row.
//
// nestedSelectIndex (>=0) marks one child that is itself a multi-row
// widget element (e.g. QuickSearch contains a Select); we recursively
// stack it instead of wrapping it as Text.
func stackTextRows(node *vdom.Node, nestedSelectIndex int) *vdom.Node {
	if node == nil {
		return node
	}

	rows := make([]*vdom.Node, 0, len(node.Children))
	for index, child := range node.Children {
		if child == nil {
			continue
		}
		if index == nestedSelectIndex && child.Type == vdom.ElementNode {
			rows = append(rows, stackTextRows(child, -1))
			continue
		}
		if child.Type == vdom.TextNode {
			rows = append(rows, components.Text(child.Text))
			continue
		}
		rows = append(rows, child)
	}

	return components.Box(vdom.Props{"flexDirection": "column"}, rows...)
}

// section wraps a widget body with a titled Divider header so every
// section in the gallery has a consistent shape.
func section(title string, body ...*vdom.Node) *vdom.Node {
	children := make([]*vdom.Node, 0, len(body)+1)
	children = append(children, components.Divider(components.DividerProps{
		Title:   title,
		Padding: 1,
		Width:   galleryWidth,
	}))
	children = append(children, body...)
	return components.Box(vdom.Props{"flexDirection": "column"}, children...)
}

func widgetTextInput() *vdom.Node {
	return section("TextInput",
		components.TextInput(components.TextInputProps{
			Value: "sample text",
			Width: 24,
		}),
		components.TextInput(components.TextInputProps{
			Placeholder: "your name…",
			Width:       24,
		}),
	)
}

func widgetPasswordInput() *vdom.Node {
	return section("PasswordInput",
		components.PasswordInput(components.TextInputProps{
			Value: "secret123",
			Width: 24,
		}),
	)
}

func widgetProgressBar() *vdom.Node {
	return section("ProgressBar",
		components.ProgressBar(components.ProgressBarProps{
			Percent:     50,
			Width:       40,
			ShowPercent: true,
		}),
	)
}

func widgetSpinner() *vdom.Node {
	return section("Spinner",
		components.Spinner(components.SpinnerProps{
			Type: components.DotsSpinner,
			Text: "loading…",
		}),
	)
}

func widgetTable() *vdom.Node {
	table := components.SimpleTable(
		[]string{"id", "name", "role"},
		[][]string{
			{"1", "alice", "admin"},
			{"2", "bob", "user"},
			{"3", "carol", "guest"},
		},
	)
	return section("Table", table)
}

func widgetSelect() *vdom.Node {
	items := []components.SelectItem{
		{Label: "Option A", Value: "a"},
		{Label: "Option B", Value: "b"},
		{Label: "Option C", Value: "c"},
	}
	return section("Select",
		stackTextRows(components.Select(components.SelectProps{
			Items:    items,
			Selected: 1,
			Focused:  true,
		}), -1),
	)
}

func widgetMultiSelect() *vdom.Node {
	items := []components.MultiSelectItem{
		{Label: "Apples", Value: "apples"},
		{Label: "Bananas", Value: "bananas"},
		{Label: "Cherries", Value: "cherries"},
	}
	return section("MultiSelect",
		stackTextRows(components.MultiSelect(components.MultiSelectProps{
			Items:    items,
			Selected: []string{"bananas"},
			Cursor:   0,
			Focused:  true,
		}), -1),
	)
}

func widgetConfirm() *vdom.Node {
	return section("Confirm",
		components.Confirm(components.ConfirmProps{
			Question: "Continue?",
			Default:  true,
		}),
	)
}

func widgetDivider() *vdom.Node {
	return section("Divider",
		components.Divider(components.DividerProps{
			Title:   "section break",
			Padding: 1,
			Width:   galleryWidth,
		}),
	)
}

func widgetAlert() *vdom.Node {
	// Alert (the bordered form) wraps its body in an "alert-body" element
	// the renderer treats as inline text — title and message collide. We
	// fall back to AlertText, the package's intentional inline variant,
	// and stack one alert per line.
	variants := []components.AlertProps{
		{Variant: components.AlertInfo, Title: "Info", Message: "This is an informational message."},
		{Variant: components.AlertSuccess, Title: "Success", Message: "Operation completed successfully."},
		{Variant: components.AlertWarning, Title: "Warning", Message: "Heads up — disk almost full."},
		{Variant: components.AlertError, Title: "Error", Message: "Something went wrong."},
	}
	rows := make([]*vdom.Node, 0, len(variants))
	for _, p := range variants {
		rows = append(rows, components.Text(components.AlertText(p)))
	}
	return section("Alert", rows...)
}

func widgetTabs() *vdom.Node {
	tabs := []components.TabItem{
		{Label: "Overview", Content: components.Text("Tab 1 panel: overview content.")},
		{Label: "Details", Content: components.Text("Tab 2 panel: details content.")},
		{Label: "Logs", Content: components.Text("Tab 3 panel: log content.")},
	}
	return section("Tabs",
		stackTextRows(components.Tabs(components.TabsProps{
			Items:   tabs,
			Active:  0,
			Focused: true,
		}), 1),
	)
}

func widgetQuickSearch() *vdom.Node {
	items := []components.SelectItem{
		{Label: "apple", Value: "apple"},
		{Label: "apricot", Value: "apricot"},
		{Label: "banana", Value: "banana"},
		{Label: "cherry", Value: "cherry"},
	}
	return section("QuickSearch",
		stackTextRows(components.QuickSearch(components.QuickSearchProps{
			Items:    items,
			Query:    "ap",
			Selected: 0,
			Focused:  true,
		}), 1),
	)
}

func widgetLink() *vdom.Node {
	return section("Link",
		components.Link(components.LinkProps{
			URL:  "https://github.com/dh-kam/ink-go",
			Text: "ink-go on GitHub",
		}),
	)
}

func widgetGradient() *vdom.Node {
	return section("Gradient",
		components.Gradient(components.GradientProps{
			Text: "HELLO",
			From: [3]uint8{255, 80, 80},
			To:   [3]uint8{80, 120, 255},
		}),
	)
}

func widgetBigText() *vdom.Node {
	return section("BigText",
		stackTextRows(components.BigText(components.BigTextProps{
			Text: "GO",
			Font: components.FontTiny,
		}), -1),
	)
}

func widgetSyntax() *vdom.Node {
	const goSnippet = `package main

// greet prints a friendly hello.
func greet(name string) {
    fmt.Println("hello,", name)
}`
	return section("Syntax",
		components.Syntax(components.SyntaxProps{
			Code:     goSnippet,
			Language: components.SyntaxGo,
		}),
	)
}

func widgetImage() *vdom.Node {
	return section("Image",
		components.Image(components.ImageProps{Image: build3x3Raster()}),
	)
}

// build3x3Raster makes a tiny 3x3 RGB image: a diagonal of bright cells
// against a darker background, just enough to see the half-block encoding.
func build3x3Raster() *components.ImageData {
	red := [3]uint8{220, 60, 60}
	green := [3]uint8{60, 200, 100}
	blue := [3]uint8{60, 120, 220}
	bg := [3]uint8{30, 30, 30}

	pixels := []([3]uint8){
		red, bg, bg,
		bg, green, bg,
		bg, bg, blue,
	}
	return &components.ImageData{Width: 3, Height: 3, Pixels: pixels}
}
