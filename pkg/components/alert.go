package components

import (
	"strings"

	"github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// AlertVariant selects the icon + color palette for an Alert.
type AlertVariant string

const (
	AlertInfo    AlertVariant = "info"
	AlertSuccess AlertVariant = "success"
	AlertWarning AlertVariant = "warning"
	AlertError   AlertVariant = "error"
)

// AlertProps configures an Alert. Title is optional; Message is the body.
type AlertProps struct {
	Variant AlertVariant
	Title   string
	Message string
}

// Alert renders an info / success / warning / error block — colored icon
// in front of the title, the message below, wrapped in a Border.
func Alert(props AlertProps) *vdom.Node {
	icon, color := alertVariantStyle(props.Variant)
	header := styles.Colorize(icon, color, styles.Foreground)
	if props.Title != "" {
		header += " " + styles.Bold(props.Title)
	}

	children := []*vdom.Node{
		vdom.CreateTextNode(header),
	}
	if props.Message != "" {
		children = append(children, vdom.CreateTextNode(props.Message))
	}

	body := vdom.CreateElement("alert-body", vdom.Props{"flexDirection": "column"}, children...)

	return Border(BorderProps{Style: BorderRounded}, vdom.Props{"padding": 1}, body)
}

func alertVariantStyle(v AlertVariant) (string, styles.Color) {
	switch v {
	case AlertSuccess:
		return "✓", styles.Green
	case AlertWarning:
		return "⚠", styles.Yellow
	case AlertError:
		return "✗", styles.Red
	default: // info / unknown
		return "i", styles.Blue
	}
}

// AlertText returns just the icon + title + message rendered as plain
// strings (no Border) — handy for inline status messages where a full
// bordered block would be too heavy.
func AlertText(props AlertProps) string {
	icon, color := alertVariantStyle(props.Variant)
	var b strings.Builder
	b.WriteString(styles.Colorize(icon, color, styles.Foreground))
	if props.Title != "" {
		b.WriteString(" ")
		b.WriteString(styles.Bold(props.Title))
	}
	if props.Message != "" {
		b.WriteString(": ")
		b.WriteString(props.Message)
	}
	return b.String()
}
