package components

import (
	"strings"

	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TabItem is a single tab — a header label plus the panel content shown when
// it is the active tab. Content may be nil for header-only tabs.
type TabItem struct {
	Label   string
	Content *vdom.Node
}

// TabsProps configures the pure-render Tabs component. Active is the index of
// the currently shown tab and is clamped into range; Focused toggles the
// colored / bold highlight on the active header.
type TabsProps struct {
	Items   []TabItem
	Active  int
	Focused bool
}

// tabSeparator is the glyph drawn between tab labels in the header row.
const tabSeparator = " │ "

// Tabs renders a header row of tab labels above the active tab's panel
// content. The active label is wrapped in [brackets] and (when Focused)
// rendered bold + cyan; inactive labels are dim. Pure render — pair with
// TabsState to drive Active updates from input handlers.
func Tabs(props TabsProps) *vdom.Node {
	if len(props.Items) == 0 {
		return vdom.CreateElement("tabs", vdom.Props{"flexDirection": "column"})
	}

	active := clampIndex(props.Active, len(props.Items))

	header := vdom.CreateTextNode(renderTabHeader(props.Items, active, props.Focused))
	children := []*vdom.Node{header}

	if panel := props.Items[active].Content; panel != nil {
		children = append(children, panel)
	}

	return vdom.CreateElement("tabs", vdom.Props{"flexDirection": "column"}, children...)
}

// renderTabHeader builds the " Label1 │ [Label2] │ Label3 " header line. The
// active label is bracketed; when focused it is also bold + cyan, otherwise
// just bold. Inactive labels are dimmed for visual contrast.
func renderTabHeader(items []TabItem, active int, focused bool) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = renderTabLabel(item.Label, i == active, focused)
	}
	return " " + strings.Join(parts, tabSeparator) + " "
}

func renderTabLabel(label string, active bool, focused bool) string {
	if active {
		marked := "[" + label + "]"
		if focused {
			return styles.Bold(styles.Colorize(marked, styles.Cyan, styles.Foreground))
		}
		return styles.Bold(marked)
	}
	return styles.Dim(label)
}

func clampIndex(i, n int) int {
	if n <= 0 {
		return 0
	}
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}

// TabsState is the controller half: it owns the active index and the item
// list. Wire Next / Prev / SetActive to your key handler (typically Tab /
// Shift-Tab) and feed Active back into Tabs each render.
type TabsState struct {
	Items  []TabItem
	Active int
}

// NewTabsState constructs a state starting at index 0.
func NewTabsState(items []TabItem) *TabsState {
	return &TabsState{Items: items, Active: 0}
}

// Next advances to the next tab, wrapping back to the first after the last.
// No-op on an empty item list.
func (s *TabsState) Next() {
	if len(s.Items) == 0 {
		return
	}
	if s.Active >= len(s.Items)-1 {
		s.Active = 0
	} else {
		s.Active++
	}
}

// Prev moves to the previous tab, wrapping to the last when at the first.
// No-op on an empty item list.
func (s *TabsState) Prev() {
	if len(s.Items) == 0 {
		return
	}
	if s.Active <= 0 {
		s.Active = len(s.Items) - 1
	} else {
		s.Active--
	}
}

// SetActive jumps directly to idx, clamped into the valid range. Empty list
// pins Active at 0.
func (s *TabsState) SetActive(i int) {
	s.Active = clampIndex(i, len(s.Items))
}

// Current returns (item, true) when Active points at a real tab. Empty list
// returns the zero TabItem and false.
func (s *TabsState) Current() (TabItem, bool) {
	if len(s.Items) == 0 || s.Active < 0 || s.Active >= len(s.Items) {
		return TabItem{}, false
	}
	return s.Items[s.Active], true
}
