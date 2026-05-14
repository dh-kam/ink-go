package components

import (
	"fmt"
	"time"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// SpinnerFrames contains the animation frames for different spinner styles
type SpinnerFrames struct {
	frames   []string
	interval time.Duration
}

// Frames returns the animation frames
func (s *SpinnerFrames) Frames() []string {
	if s == nil {
		return []string{}
	}
	return s.frames
}

// Interval returns the animation interval
func (s *SpinnerFrames) Interval() time.Duration {
	if s == nil {
		return 100 * time.Millisecond
	}
	return s.interval
}

// Common spinner styles
var (
	// DotsSpinner shows rotating dots: ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏
	DotsSpinner = &SpinnerFrames{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		interval: 80 * time.Millisecond,
	}

	// LineSpinner shows a rotating line: | / - \
	LineSpinner = &SpinnerFrames{
		frames: []string{"|", "/", "-", "\\"},
		interval: 100 * time.Millisecond,
	}

	// ArrowSpinner shows rotating arrows: ← ↑ → ↓
	ArrowSpinner = &SpinnerFrames{
		frames: []string{"←", "↑", "→", "↓"},
		interval: 100 * time.Millisecond,
	}

	// PlusMinusSpinner shows rotating +/x
	PlusMinusSpinner = &SpinnerFrames{
		frames: []string{"+", "x"},
		interval: 150 * time.Millisecond,
	}

	// DefaultSpinner is the default spinner style
	DefaultSpinner = DotsSpinner
)

// SpinnerProps defines the properties for a Spinner component
type SpinnerProps struct {
	// Type is the spinner animation style
	Type *SpinnerFrames
	// Text is optional text to show next to the spinner
	Text string
}

// Spinner creates a loading spinner component
// The spinner will animate when rendered in a loop
func Spinner(props SpinnerProps) *vdom.Node {
	spinnerType := props.Type
	if spinnerType == nil {
		spinnerType = DefaultSpinner
	}

	// Get current frame based on time
	frame := getCurrentFrame(spinnerType)

	// Build the spinner display
	text := frame
	if props.Text != "" {
		text = fmt.Sprintf("%s %s", frame, props.Text)
	}

	return vdom.CreateElement("spinner", nil, vdom.CreateTextNode(text))
}

// getCurrentFrame returns the current frame based on time
func getCurrentFrame(spinner *SpinnerFrames) string {
	if spinner == nil || len(spinner.frames) == 0 {
		return "⠋"
	}

	// Calculate frame index based on current time
	elapsed := time.Now().UnixNano() / int64(spinner.interval)
	frameIndex := int(elapsed) % len(spinner.frames)

	return spinner.frames[frameIndex]
}

// SpinnerWithText creates a spinner with accompanying text
func SpinnerWithText(text string) *vdom.Node {
	return Spinner(SpinnerProps{
		Text: text,
	})
}

// SpinnerWithType creates a spinner with a specific type and text
func SpinnerWithType(spinnerType *SpinnerFrames, text string) *vdom.Node {
	return Spinner(SpinnerProps{
		Type: spinnerType,
		Text: text,
	})
}
