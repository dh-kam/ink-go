package components

import (
	"strings"

	"github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// SelectItem is one row in a Select list.
type SelectItem struct {
	Label string
	Value string
}

// SelectProps configures the pure-render Select component. Selected,
// scrolling, and key handling all live on SelectState — this component is
// intentionally controlled (TextInput pattern), so the parent owns the
// state and feeds props.Selected each render.
type SelectProps struct {
	Items     []SelectItem
	Selected  int    // currently highlighted index (clamped to range)
	Limit     int    // max visible rows; 0 = show all
	Indicator string // gutter glyph for the selected row; defaults to "❯"
	Focused   bool   // when false, indicator/highlighting are dimmed
}

// DefaultSelectIndicator is the gutter glyph drawn next to the highlighted row.
const DefaultSelectIndicator = "❯"

// Select renders a list with the highlighted row marked. Pure render —
// connect SelectState to your input handler to drive Selected updates.
func Select(props SelectProps) *vdom.Node {
	if props.Indicator == "" {
		props.Indicator = DefaultSelectIndicator
	}

	if len(props.Items) == 0 {
		return vdom.CreateElement("select", nil)
	}

	state := &SelectState{
		Items:    props.Items,
		Selected: props.Selected,
		Limit:    props.Limit,
	}
	state.Clamp()

	visible, offset := state.Visible()
	rows := make([]*vdom.Node, 0, len(visible)+2)

	if props.Limit > 0 && offset > 0 {
		rows = append(rows, vdom.CreateTextNode(" ↑"))
	}
	for i, item := range visible {
		actualIdx := offset + i
		row := renderRow(item, actualIdx == state.Selected, props.Indicator, props.Focused)
		rows = append(rows, vdom.CreateTextNode(row))
	}
	if props.Limit > 0 && offset+len(visible) < len(props.Items) {
		rows = append(rows, vdom.CreateTextNode(" ↓"))
	}

	return vdom.CreateElement("select", nil, rows...)
}

func renderRow(item SelectItem, selected bool, indicator string, focused bool) string {
	var b strings.Builder
	if selected {
		if focused {
			b.WriteString(styles.Colorize(indicator, styles.Cyan, styles.Foreground))
		} else {
			b.WriteString(styles.Dim(indicator))
		}
		b.WriteString(" ")
		if focused {
			b.WriteString(styles.Bold(item.Label))
		} else {
			b.WriteString(item.Label)
		}
	} else {
		b.WriteString(strings.Repeat(" ", visualLen(indicator)+1))
		b.WriteString(item.Label)
	}
	return b.String()
}

// visualLen returns the visible length of s for indicator padding.
// Approximation: count runes (good enough for single-glyph indicators).
func visualLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// SelectState is the controller half: it tracks the highlighted index, the
// item list, and the visible-window limit. Wire its movement methods to
// arrow-key handlers and feed props.Selected back into the Select render.
type SelectState struct {
	Items    []SelectItem
	Selected int
	Limit    int
}

// NewSelectState constructs a state with sensible defaults.
func NewSelectState(items []SelectItem, limit int) *SelectState {
	s := &SelectState{Items: items, Selected: 0, Limit: limit}
	s.Clamp()
	return s
}

// MoveUp moves the highlight up, wrapping to the last item.
func (s *SelectState) MoveUp() {
	if len(s.Items) == 0 {
		return
	}
	if s.Selected <= 0 {
		s.Selected = len(s.Items) - 1
	} else {
		s.Selected--
	}
}

// MoveDown moves the highlight down, wrapping to the first item.
func (s *SelectState) MoveDown() {
	if len(s.Items) == 0 {
		return
	}
	if s.Selected >= len(s.Items)-1 {
		s.Selected = 0
	} else {
		s.Selected++
	}
}

// MoveTo jumps the highlight to idx (clamped).
func (s *SelectState) MoveTo(idx int) {
	s.Selected = idx
	s.Clamp()
}

// Confirm returns the currently selected item.
func (s *SelectState) Confirm() (SelectItem, bool) {
	return s.Current()
}

// Current returns (item, true) when the selection is in range.
func (s *SelectState) Current() (SelectItem, bool) {
	if s.Selected < 0 || s.Selected >= len(s.Items) {
		return SelectItem{}, false
	}
	return s.Items[s.Selected], true
}

// Clamp pulls Selected back into the valid range. Empty list → 0.
func (s *SelectState) Clamp() {
	if len(s.Items) == 0 {
		s.Selected = 0
		return
	}
	if s.Selected < 0 {
		s.Selected = 0
	}
	if s.Selected >= len(s.Items) {
		s.Selected = len(s.Items) - 1
	}
}

// Visible returns the slice of items that should currently be drawn given
// Limit, plus the offset (start index) of that slice into Items. Limit ≤ 0
// returns the full list with offset 0.
func (s *SelectState) Visible() ([]SelectItem, int) {
	if s.Limit <= 0 || s.Limit >= len(s.Items) {
		return s.Items, 0
	}
	// Center the selection in the window when possible.
	half := s.Limit / 2
	offset := s.Selected - half
	if offset < 0 {
		offset = 0
	}
	if offset+s.Limit > len(s.Items) {
		offset = len(s.Items) - s.Limit
	}
	return s.Items[offset : offset+s.Limit], offset
}
