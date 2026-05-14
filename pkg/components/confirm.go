package components

import (
	"strings"

	"github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// ConfirmProps configures the pure-render Confirm component, a port of
// ink-confirm-input. The rendered prompt is:
//
//	"<Question> [Y/n]"  when Default == true
//	"<Question> [y/N]"  when Default == false
//
// Once Value is non-nil, the chosen letter ("y" or "n") is appended in a
// dimmed style to echo the user's answer. Like the other components in
// this package the props are controlled — pair Confirm with ConfirmState
// and feed the answer back on every render.
type ConfirmProps struct {
	// Question is the prompt text shown before the [Y/n] hint.
	Question string
	// Default is the answer used when the user hits Enter without typing
	// either letter. true → "[Y/n]", false → "[y/N]".
	Default bool
	// Value is the current answer. nil means the question has not been
	// answered yet; non-nil values render the chosen letter after the
	// hint.
	Value *bool
}

// Confirm renders a yes/no prompt. Pure render — drive answer changes via
// ConfirmState.HandleKey.
func Confirm(props ConfirmProps) *vdom.Node {
	var b strings.Builder
	b.WriteString(props.Question)
	if props.Question != "" {
		b.WriteString(" ")
	}
	b.WriteString(confirmHint(props.Default))

	if props.Value != nil {
		b.WriteString(" ")
		if *props.Value {
			b.WriteString(styles.Colorize("y", styles.Green, styles.Foreground))
		} else {
			b.WriteString(styles.Colorize("n", styles.Red, styles.Foreground))
		}
	}

	return vdom.CreateElement("confirm", nil, vdom.CreateTextNode(b.String()))
}

// confirmHint returns the bracketed Y/n hint, with the default answer
// uppercased so the user can spot it at a glance.
func confirmHint(def bool) string {
	if def {
		return "[Y/n]"
	}
	return "[y/N]"
}

// ConfirmState is the controller half of the Confirm pattern. HandleKey
// translates a single rune (typically from your input loop) into a
// resolved (true/false) answer or a no-op return when the rune is
// ignored.
type ConfirmState struct {
	Question string
	Default  bool
	Answer   *bool
}

// NewConfirmState builds a fresh ConfirmState with no answer recorded
// yet.
func NewConfirmState(question string, def bool) *ConfirmState {
	return &ConfirmState{
		Question: question,
		Default:  def,
		Answer:   nil,
	}
}

// HandleKey processes a key press and returns (resolved, value).
//
//	'y' / 'Y' → resolved=true, value=true
//	'n' / 'N' → resolved=true, value=false
//	'\r' / '\n' → resolved=true, value=Default
//	anything else → resolved=false, value=false (state unchanged)
//
// On a resolved key the Answer field is populated so subsequent renders
// echo the choice via ConfirmProps.Value.
func (s *ConfirmState) HandleKey(ch rune) (resolved bool, value bool) {
	switch ch {
	case 'y', 'Y':
		v := true
		s.Answer = &v
		return true, true
	case 'n', 'N':
		v := false
		s.Answer = &v
		return true, false
	case '\r', '\n':
		v := s.Default
		s.Answer = &v
		return true, v
	}
	return false, false
}

// Reset clears any recorded answer so the prompt can be re-asked.
func (s *ConfirmState) Reset() {
	s.Answer = nil
}
