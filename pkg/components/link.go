package components

import (
	"strings"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// LinkProps configures a Link rendering.
//
// URL is the destination passed to the OSC 8 hyperlink escape; supporting
// terminals (iTerm2, WezTerm, Alacritty, kitty, modern GNOME Terminal,
// VS Code's integrated terminal, …) render it as a clickable link, while
// other terminals simply print Text.
//
// Text is the visible label. When omitted (empty string), URL is used as
// the label. When URL itself is empty, the component falls back to a
// plain text node — no escape sequence is emitted, since there is nothing
// for the terminal to link to.
type LinkProps struct {
	URL  string
	Text string
}

// OSC 8 hyperlink delimiters as defined by the de facto terminal spec
// (https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda).
//
// Sequence layout:
//
//	OSC 8 ; params ; URI ST  <visible text>  OSC 8 ; ; ST
//
// We never set params, so the closing form is "OSC 8 ; ; ST".
const (
	osc8LinkPrefixOpen  = "\x1b]8;;"
	osc8StringTerminator = "\x1b\\"
	osc8LinkClose       = "\x1b]8;;\x1b\\"
)

// Link renders a clickable hyperlink using the OSC 8 escape sequence.
//
// The returned node is a plain text node so it composes naturally inside
// any text container — wrapping it in <Text> or printing it directly via
// ink.RenderToString works without further wiring.
func Link(props LinkProps) *vdom.Node {
	label := props.Text
	if label == "" {
		label = props.URL
	}

	if props.URL == "" {
		// No URL → degrade gracefully to plain text. We deliberately do
		// not emit an escape sequence so consumers can safely concatenate
		// the result with surrounding output.
		return vdom.CreateTextNode(label)
	}

	var b strings.Builder
	b.Grow(len(osc8LinkPrefixOpen) + len(props.URL) + len(osc8StringTerminator) + len(label) + len(osc8LinkClose))
	b.WriteString(osc8LinkPrefixOpen)
	b.WriteString(props.URL)
	b.WriteString(osc8StringTerminator)
	b.WriteString(label)
	b.WriteString(osc8LinkClose)

	return vdom.CreateTextNode(b.String())
}
