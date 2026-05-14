// Package devtools — DebugPanel renders an in-app stats overlay alongside or
// underneath a running goink application.
//
// The panel is a pure component: callers supply pre-aggregated PanelData (the
// runtime is responsible for producing the numbers) and DebugPanel turns it
// into a vdom tree that can be composed like any other component. WithDebug
// is the convenience wrapper that lays the panel out next to a real app
// using a row (Side="right") or column (Side="bottom") flex container.
package devtools

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// PanelData contains the metrics surfaced by DebugPanel. All fields are
// optional — zero values render as "0" / "0s" so the panel can be wired up
// before the runtime has produced any real numbers.
type PanelData struct {
	Title        string            // Header label (defaults to "DEBUG")
	Renders      int               // Total render count
	CacheHits    int               // Cached re-renders served from memo cache
	LastDuration time.Duration     // Wall time of the most recent render
	AvgDuration  time.Duration     // Rolling average render duration
	StateCount   int               // Hook state slot count for the current render
	EffectCount  int               // Effect slot count
	PatchCount   int               // Patch ops applied for the most recent frame
	Custom       map[string]string // Free-form key=value rows appended after the built-in stats
}

// DebugPanelProps configures a DebugPanel.
type DebugPanelProps struct {
	Data    PanelData
	Width   int  // 0 = use defaultPanelWidth (30)
	Compact bool // when true, render a single status-bar line instead of a bordered block
}

// WithDebugProps wires a real application together with a DebugPanel.
type WithDebugProps struct {
	App   func() *vdom.Node // Real app render fn. nil App skips the app slot and only renders the panel.
	Panel PanelData
	Side  string // "right" (default) or "bottom"
}

const (
	defaultPanelWidth = 30
	defaultPanelTitle = "DEBUG"
)

// DebugPanel renders the stats panel as a vdom node. In compact mode the
// output is a single text node ready to drop into a status bar; otherwise the
// output is a Border-wrapped column with one row per metric.
func DebugPanel(props DebugPanelProps) *vdom.Node {
	if props.Compact {
		return components.Text(compactLine(props.Data))
	}

	width := props.Width
	if width <= 0 {
		width = defaultPanelWidth
	}

	title := props.Data.Title
	if title == "" {
		title = defaultPanelTitle
	}

	rows := buildRows(props.Data)
	children := make([]*vdom.Node, 0, len(rows))
	for _, row := range rows {
		children = append(children, components.Text(row))
	}

	body := vdom.CreateElement("debug-panel-body", vdom.Props{
		"flexDirection": "column",
		"width":         width,
	}, children...)

	return components.Border(
		components.BorderProps{Style: components.BorderRounded, Title: title},
		vdom.Props{"width": width},
		body,
	)
}

// WithDebug composes a real app and a DebugPanel into a single layout. When
// Side is "bottom" the layout is a column (panel under the app); any other
// value (including "" or "right") yields a row (panel to the right of the
// app). A nil App skips the app slot so the panel can render on its own.
func WithDebug(props WithDebugProps) *vdom.Node {
	direction := "row"
	if props.Side == "bottom" {
		direction = "column"
	}

	panel := DebugPanel(DebugPanelProps{Data: props.Panel})

	children := make([]*vdom.Node, 0, 2)
	if props.App != nil {
		if appNode := props.App(); appNode != nil {
			children = append(children, appNode)
		}
	}
	children = append(children, panel)

	return components.Box(vdom.Props{"flexDirection": direction}, children...)
}

// buildRows produces the ordered text lines for the non-compact panel.
func buildRows(data PanelData) []string {
	rows := []string{
		fmt.Sprintf("Renders: %d (%d hit)", data.Renders, data.CacheHits),
		fmt.Sprintf("Last:    %s", formatDuration(data.LastDuration)),
		fmt.Sprintf("Avg:     %s", formatDuration(data.AvgDuration)),
		fmt.Sprintf("State slots: %d", data.StateCount),
		fmt.Sprintf("Effects: %d", data.EffectCount),
		fmt.Sprintf("Patches/frame: %d", data.PatchCount),
	}

	if len(data.Custom) > 0 {
		rows = append(rows, "─ Custom ─")
		keys := make([]string, 0, len(data.Custom))
		for k := range data.Custom {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			rows = append(rows, fmt.Sprintf("%s: %s", k, data.Custom[k]))
		}
	}

	return rows
}

// compactLine produces the single-line status bar form. We deliberately use
// short ASCII-friendly tokens with unicode glyphs so the line stays roughly
// terminal-width independent.
func compactLine(data PanelData) string {
	var b strings.Builder
	b.WriteString(" ⏱ ")
	b.WriteString(formatDuration(data.LastDuration))
	b.WriteString("  📦 ")
	fmt.Fprintf(&b, "%d states", data.StateCount)
	b.WriteString("  🔄 ")
	fmt.Fprintf(&b, "%d patches", data.PatchCount)

	if len(data.Custom) > 0 {
		keys := make([]string, 0, len(data.Custom))
		for k := range data.Custom {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "  %s=%s", k, data.Custom[k])
		}
	}

	b.WriteByte(' ')
	return b.String()
}

// formatDuration prints a duration in a millisecond-friendly form. We keep it
// inline rather than depending on time.Duration.String() because the standard
// formatting produces noisy output ("3.4ms" vs "3.412345ms") for sub-second
// values that change every frame.
func formatDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return "0s"
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.1fµs", float64(d.Nanoseconds())/float64(time.Microsecond))
	case d < time.Second:
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/float64(time.Millisecond))
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
