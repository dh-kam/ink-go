package components

import (
	"strings"

	"github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// QuickSearchProps configures the pure-render QuickSearch component. Like
// Select it is intentionally controlled — the parent owns query/selection
// state via QuickSearchState and feeds props on each render.
//
// Layout: line 1 is "Search: <query>_" (cursor glyph appended when Focused).
// Subsequent lines are the Select-style filtered list. Selected indexes into
// the filtered list, not the source Items.
type QuickSearchProps struct {
	Items     []SelectItem // source list (re-uses pkg/components SelectItem)
	Query     string       // current search text
	Selected  int          // index in the FILTERED list
	Limit     int          // max visible filtered rows; 0 = show all
	Focused   bool         // when false, indicator/highlight are dimmed
	Indicator string       // gutter glyph; defaults to DefaultSelectIndicator
}

// DefaultQuickSearchPrompt is the prefix drawn before the query text.
const DefaultQuickSearchPrompt = "Search: "

// QuickSearch renders the search prompt followed by a filtered Select. Pure
// render — call into QuickSearchState from your input handler.
func QuickSearch(props QuickSearchProps) *vdom.Node {
	if props.Indicator == "" {
		props.Indicator = DefaultSelectIndicator
	}

	prompt := buildSearchPrompt(props.Query, props.Focused)
	children := make([]*vdom.Node, 0, 2)
	children = append(children, vdom.CreateTextNode(prompt))

	filtered := filterItems(props.Items, props.Query)
	selectNode := Select(SelectProps{
		Items:     filtered,
		Selected:  props.Selected,
		Limit:     props.Limit,
		Focused:   props.Focused,
		Indicator: props.Indicator,
	})
	children = append(children, selectNode)

	return vdom.CreateElement("quicksearch", nil, children...)
}

// buildSearchPrompt formats "Search: <query>_" with a trailing cursor glyph
// when focused. The cursor is dimmed when the input is not focused so the
// caller still gets a stable layout.
func buildSearchPrompt(query string, focused bool) string {
	var b strings.Builder
	b.WriteString(DefaultQuickSearchPrompt)
	b.WriteString(query)
	if focused {
		b.WriteString(styles.Colorize("_", styles.Cyan, styles.Foreground))
	} else {
		b.WriteString(styles.Dim("_"))
	}
	return b.String()
}

// filterItems returns the subset of items whose Label contains query
// (case-insensitive). Empty query matches everything.
func filterItems(items []SelectItem, query string) []SelectItem {
	if query == "" {
		return items
	}
	needle := strings.ToLower(query)
	out := make([]SelectItem, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Label), needle) {
			out = append(out, item)
		}
	}
	return out
}

// QuickSearchState is the controller half: it tracks the source items, the
// current query string, the highlighted index in the filtered view, and the
// visible-window limit. Wire it to a key handler to drive the Select.
type QuickSearchState struct {
	Items    []SelectItem
	Query    string
	Selected int // index into Filtered(), not Items
	Limit    int
}

// NewQuickSearchState returns a state with empty query and selection at 0.
func NewQuickSearchState(items []SelectItem, limit int) *QuickSearchState {
	return &QuickSearchState{
		Items:    items,
		Query:    "",
		Selected: 0,
		Limit:    limit,
	}
}

// SetQuery replaces the query and resets the selection to the top of the
// new filtered list. Resetting matches ink-quicksearch-input's UX — a fresh
// query starts the cursor at the first match.
func (s *QuickSearchState) SetQuery(q string) {
	s.Query = q
	s.Selected = 0
}

// AppendQuery appends a single rune to the query and resets selection.
func (s *QuickSearchState) AppendQuery(ch rune) {
	s.Query += string(ch)
	s.Selected = 0
}

// BackspaceQuery removes the last rune of the query and resets selection.
// Safe to call on an empty query.
func (s *QuickSearchState) BackspaceQuery() {
	if s.Query == "" {
		return
	}
	runes := []rune(s.Query)
	s.Query = string(runes[:len(runes)-1])
	s.Selected = 0
}

// MoveUp moves the highlight up within the filtered list, wrapping. No-op on
// an empty filtered list.
func (s *QuickSearchState) MoveUp() {
	n := len(s.Filtered())
	if n == 0 {
		return
	}
	if s.Selected <= 0 {
		s.Selected = n - 1
	} else {
		s.Selected--
	}
}

// MoveDown moves the highlight down within the filtered list, wrapping.
func (s *QuickSearchState) MoveDown() {
	n := len(s.Filtered())
	if n == 0 {
		return
	}
	if s.Selected >= n-1 {
		s.Selected = 0
	} else {
		s.Selected++
	}
}

// Filtered returns items whose Label contains Query (case-insensitive).
// Empty query returns the full list.
func (s *QuickSearchState) Filtered() []SelectItem {
	return filterItems(s.Items, s.Query)
}

// Confirm returns the currently highlighted filtered item. Returns
// (zero, false) when the filtered list is empty or Selected is out of range.
func (s *QuickSearchState) Confirm() (SelectItem, bool) {
	filtered := s.Filtered()
	if len(filtered) == 0 {
		return SelectItem{}, false
	}
	if s.Selected < 0 || s.Selected >= len(filtered) {
		return SelectItem{}, false
	}
	return filtered[s.Selected], true
}
