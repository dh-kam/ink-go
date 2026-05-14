package components

import (
	"strings"

	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// DividerProps configures Divider rendering.
type DividerProps struct {
	Title   string       // optional centered label
	Char    string       // single rune used to draw the rule (default: "─")
	Padding int          // spaces around Title when present
	Color   styles.Color // optional color for the rule and title
	Width   int          // total visible width (default: 40)
}

// DefaultDividerChar is the rule character used when DividerProps.Char is empty.
const DefaultDividerChar = "─"

// DefaultDividerWidth is used when DividerProps.Width <= 0.
const DefaultDividerWidth = 40

// Divider renders a single horizontal rule, optionally with a centered
// title. Useful between sections of dashboards / forms.
func Divider(props DividerProps) *vdom.Node {
	if props.Char == "" {
		props.Char = DefaultDividerChar
	}
	if props.Width <= 0 {
		props.Width = DefaultDividerWidth
	}
	if props.Padding < 0 {
		props.Padding = 0
	}

	var line string
	if props.Title == "" {
		line = strings.Repeat(props.Char, props.Width)
	} else {
		pad := strings.Repeat(" ", props.Padding)
		title := pad + props.Title + pad
		if len(title) >= props.Width {
			line = title
		} else {
			fill := props.Width - len(title)
			leftFill := fill / 2
			rightFill := fill - leftFill
			line = strings.Repeat(props.Char, leftFill) + title + strings.Repeat(props.Char, rightFill)
		}
	}

	if props.Color != nil {
		line = styles.Colorize(line, props.Color, styles.Foreground)
	}
	return vdom.CreateElement("divider", nil, vdom.CreateTextNode(line))
}
