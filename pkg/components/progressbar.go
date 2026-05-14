package components

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// ProgressBarProps defines the properties for a ProgressBar component
type ProgressBarProps struct {
	// Percent is the completion percentage (0-100)
	Percent int
	// Width is the total width of the progress bar in characters
	Width int
	// Character is the character to use for filled portions
	Character string
	// ShowPercent controls whether to show the percentage text
	ShowPercent bool
}

// DefaultProgressBarWidth is the default width for progress bars
const DefaultProgressBarWidth = 40

// DefaultProgressBarChar is the default character for filled portions
const DefaultProgressBarChar = "█"

// ProgressBar creates a progress bar component
func ProgressBar(props ProgressBarProps) *vdom.Node {
	// Set defaults
	if props.Width <= 0 {
		props.Width = DefaultProgressBarWidth
	}
	if props.Character == "" {
		props.Character = DefaultProgressBarChar
	}

	// Clamp percent to valid range
	percent := props.Percent
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	// Calculate filled and empty widths
	filledWidth := (percent * props.Width) / 100
	emptyWidth := props.Width - filledWidth

	// Build the bar
	bar := strings.Repeat(props.Character, filledWidth)
	bar += strings.Repeat(" ", emptyWidth)

	// Add percentage display if requested
	text := bar
	if props.ShowPercent {
		text = fmt.Sprintf("%s %d%%", bar, percent)
	}

	return vdom.CreateElement("progress", nil, vdom.CreateTextNode(text))
}

// ProgressBarSimple creates a simple progress bar with just the percentage
func ProgressBarSimple(percent int) *vdom.Node {
	return ProgressBar(ProgressBarProps{
		Percent: percent,
	})
}

// ProgressBarWithPercent creates a progress bar with percentage display
func ProgressBarWithPercent(percent int, width int) *vdom.Node {
	return ProgressBar(ProgressBarProps{
		Percent:     percent,
		Width:       width,
		ShowPercent: true,
	})
}

// ProgressBarIndeterminate creates an indeterminate progress bar (animated)
func ProgressBarIndeterminate(width int) *vdom.Node {
	if width <= 0 {
		width = DefaultProgressBarWidth
	}

	// Use spinner frame for animation
	frame := getCurrentFrame(DotsSpinner)

	// Create a bar with a moving indicator
	bar := strings.Repeat(" ", width)
	midPoint := width / 2

	// Place the spinner in the middle
	runes := []rune(bar)
	if midPoint < len(runes) {
		runes[midPoint] = []rune(frame)[0]
	}
	bar = string(runes)

	return vdom.CreateElement("progress", nil, vdom.CreateTextNode(bar))
}
