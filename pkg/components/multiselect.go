package components

import (
	"sort"
	"strings"

	"github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// MultiSelectItem is one row in a MultiSelect list. Items with Disabled
// set are still rendered (in dim style) but the controller skips them
// when toggling and during SelectAll.
type MultiSelectItem struct {
	Label    string
	Value    string
	Disabled bool
}

// MultiSelectProps configures the pure-render MultiSelect component, a
// port of ink-multi-select. Selected, Cursor, and the visible window are
// all controlled — wire them through MultiSelectState.
type MultiSelectProps struct {
	// Items is the full list of choices.
	Items []MultiSelectItem
	// Selected holds the currently checked values (any order). Items
	// whose Value is in this slice render with a "[x]" glyph.
	Selected []string
	// Cursor is the highlighted row index into Items.
	Cursor int
	// Limit is the visible-window size; 0 (or larger than len(Items))
	// shows the full list with no scroll indicators.
	Limit int
	// Focused controls whether the cursor and indicator are drawn at
	// full intensity (true) or dimmed (false).
	Focused bool
}

// Glyphs used by the MultiSelect renderer.
const (
	multiSelectChecked    = "[x]"
	multiSelectUnchecked  = "[ ]"
	multiSelectIndicator  = "❯"
	multiSelectGutterSize = 1 // visible width of the indicator glyph
)

// MultiSelect renders a checkbox list with the highlighted row marked.
// Pure render — pair with MultiSelectState to drive cursor / selection
// updates.
func MultiSelect(props MultiSelectProps) *vdom.Node {
	if len(props.Items) == 0 {
		return vdom.CreateElement("multiselect", nil)
	}

	state := &MultiSelectState{
		Items:    props.Items,
		Selected: selectedSet(props.Selected),
		Cursor:   props.Cursor,
		Limit:    props.Limit,
	}
	state.clampCursor()

	visible, offset := state.Visible()
	rows := make([]*vdom.Node, 0, len(visible)+2)

	if props.Limit > 0 && offset > 0 {
		rows = append(rows, vdom.CreateTextNode(" ↑"))
	}
	for i, item := range visible {
		actualIdx := offset + i
		row := renderMultiRow(item, actualIdx == state.Cursor, state.Selected[item.Value], props.Focused)
		rows = append(rows, vdom.CreateTextNode(row))
	}
	if props.Limit > 0 && offset+len(visible) < len(props.Items) {
		rows = append(rows, vdom.CreateTextNode(" ↓"))
	}

	return vdom.CreateElement("multiselect", nil, rows...)
}

func renderMultiRow(item MultiSelectItem, cursor, checked, focused bool) string {
	var b strings.Builder

	// Gutter (cursor indicator).
	if cursor {
		if focused {
			b.WriteString(styles.Colorize(multiSelectIndicator, styles.Cyan, styles.Foreground))
		} else {
			b.WriteString(styles.Dim(multiSelectIndicator))
		}
	} else {
		b.WriteString(strings.Repeat(" ", multiSelectGutterSize))
	}
	b.WriteString(" ")

	// Checkbox glyph.
	glyph := multiSelectUnchecked
	if checked {
		glyph = multiSelectChecked
	}
	if checked {
		b.WriteString(styles.Colorize(glyph, styles.Green, styles.Foreground))
	} else {
		b.WriteString(glyph)
	}
	b.WriteString(" ")

	// Label.
	switch {
	case item.Disabled:
		b.WriteString(styles.Dim(item.Label))
	case cursor && focused:
		b.WriteString(styles.Bold(item.Label))
	default:
		b.WriteString(item.Label)
	}
	return b.String()
}

func selectedSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, v := range values {
		out[v] = true
	}
	return out
}

// MultiSelectState is the controller half of the MultiSelect pattern. It
// owns the cursor, the set of selected values, and the visible window.
// Wire its movement / toggle methods to your input loop and feed
// MultiSelectProps back into MultiSelect on every render.
type MultiSelectState struct {
	Items    []MultiSelectItem
	Selected map[string]bool
	Cursor   int
	Limit    int
}

// NewMultiSelectState builds a state with no items selected. Limit ≤ 0
// disables the scroll window.
func NewMultiSelectState(items []MultiSelectItem, limit int) *MultiSelectState {
	s := &MultiSelectState{
		Items:    items,
		Selected: map[string]bool{},
		Cursor:   0,
		Limit:    limit,
	}
	s.clampCursor()
	return s
}

// MoveUp moves the cursor up one row, wrapping to the last item when at
// the top. No-op for an empty list.
func (s *MultiSelectState) MoveUp() {
	if len(s.Items) == 0 {
		return
	}
	if s.Cursor <= 0 {
		s.Cursor = len(s.Items) - 1
	} else {
		s.Cursor--
	}
}

// MoveDown moves the cursor down one row, wrapping to the first item
// when at the bottom. No-op for an empty list.
func (s *MultiSelectState) MoveDown() {
	if len(s.Items) == 0 {
		return
	}
	if s.Cursor >= len(s.Items)-1 {
		s.Cursor = 0
	} else {
		s.Cursor++
	}
}

// Toggle flips the checked state of the item under the cursor. Disabled
// items are ignored.
func (s *MultiSelectState) Toggle() {
	if len(s.Items) == 0 {
		return
	}
	item := s.Items[s.Cursor]
	if item.Disabled {
		return
	}
	if s.Selected[item.Value] {
		delete(s.Selected, item.Value)
	} else {
		s.Selected[item.Value] = true
	}
}

// SelectAll marks every non-disabled item as checked.
func (s *MultiSelectState) SelectAll() {
	for _, it := range s.Items {
		if it.Disabled {
			continue
		}
		s.Selected[it.Value] = true
	}
}

// ClearAll unchecks every item.
func (s *MultiSelectState) ClearAll() {
	s.Selected = map[string]bool{}
}

// Values returns the checked values in the order they appear in Items.
// Empty slice (never nil) when nothing is selected.
func (s *MultiSelectState) Values() []string {
	out := make([]string, 0, len(s.Selected))
	for _, it := range s.Items {
		if s.Selected[it.Value] {
			out = append(out, it.Value)
		}
	}
	// Items already define the canonical order; sort.SliceStable is a
	// no-op here but documents the invariant for callers that mutate
	// Items between renders.
	sort.SliceStable(out, func(i, j int) bool { return false })
	return out
}

// Visible returns the slice of items that should currently be drawn
// given Limit, plus the offset (start index) of that slice into Items.
// Limit ≤ 0 (or Limit ≥ len(Items)) returns the full list with offset 0.
// The window is centered on the cursor when possible.
func (s *MultiSelectState) Visible() ([]MultiSelectItem, int) {
	if s.Limit <= 0 || s.Limit >= len(s.Items) {
		return s.Items, 0
	}
	half := s.Limit / 2
	offset := s.Cursor - half
	if offset < 0 {
		offset = 0
	}
	if offset+s.Limit > len(s.Items) {
		offset = len(s.Items) - s.Limit
	}
	return s.Items[offset : offset+s.Limit], offset
}

func (s *MultiSelectState) clampCursor() {
	if len(s.Items) == 0 {
		s.Cursor = 0
		return
	}
	if s.Cursor < 0 {
		s.Cursor = 0
	}
	if s.Cursor >= len(s.Items) {
		s.Cursor = len(s.Items) - 1
	}
}
