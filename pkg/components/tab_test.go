package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func mkTabs(labels ...string) []components.TabItem {
	out := make([]components.TabItem, len(labels))
	for i, l := range labels {
		out[i] = components.TabItem{
			Label:   l,
			Content: vdom.CreateTextNode("panel-" + l),
		}
	}
	return out
}

// headerText extracts the rendered header line (first child text node) from
// the Tabs node so tests can assert on the visible label string.
func headerText(t *testing.T, node *vdom.Node) string {
	t.Helper()
	if len(node.Children) == 0 {
		t.Fatalf("Tabs node had no children — expected at least a header")
	}
	header := node.Children[0]
	if header.Type != vdom.TextNode {
		t.Fatalf("first child is %v, want TextNode", header.Type)
	}
	return header.Text
}

func TestTabsRendersHeaderAndPanel(t *testing.T) {
	items := mkTabs("Logs", "Settings", "About")
	node := components.Tabs(components.TabsProps{Items: items, Active: 1, Focused: true})

	// header + panel
	if got := len(node.Children); got != 2 {
		t.Fatalf("children = %d, want 2 (header + panel)", got)
	}

	header := headerText(t, node)
	for _, label := range []string{"Logs", "Settings", "About"} {
		if !strings.Contains(header, label) {
			t.Errorf("header missing label %q: %q", label, header)
		}
	}

	// Active label should be bracketed in the header.
	if !strings.Contains(header, "[Settings]") {
		t.Errorf("active label not bracketed: %q", header)
	}

	// Panel content for the active tab should be the second child.
	panel := node.Children[1]
	if panel.Type != vdom.TextNode || panel.Text != "panel-Settings" {
		t.Errorf("active panel = %+v, want text panel-Settings", panel)
	}
}

func TestTabsActiveBracketAndBoldFocused(t *testing.T) {
	items := mkTabs("One", "Two")
	node := components.Tabs(components.TabsProps{Items: items, Active: 0, Focused: true})
	header := headerText(t, node)

	if !strings.Contains(header, "[One]") {
		t.Errorf("focused header missing bracketed active: %q", header)
	}
	// Bold ANSI code should be present when focused.
	if !strings.Contains(header, "\x1b[1m") {
		t.Errorf("focused header missing bold ANSI: %q", header)
	}
	// Cyan foreground should be present when focused.
	if !strings.Contains(header, "\x1b[36m") {
		t.Errorf("focused header missing cyan ANSI: %q", header)
	}
}

func TestTabsActiveBoldOnlyWhenUnfocused(t *testing.T) {
	items := mkTabs("One", "Two")
	node := components.Tabs(components.TabsProps{Items: items, Active: 1, Focused: false})
	header := headerText(t, node)

	if !strings.Contains(header, "[Two]") {
		t.Errorf("unfocused header missing bracketed active: %q", header)
	}
	// Bold should still apply for active.
	if !strings.Contains(header, "\x1b[1m") {
		t.Errorf("unfocused active should still be bold: %q", header)
	}
	// But no cyan colorize when unfocused.
	if strings.Contains(header, "\x1b[36m") {
		t.Errorf("unfocused header should not have cyan ANSI: %q", header)
	}
}

func TestTabsHeaderHasSeparator(t *testing.T) {
	items := mkTabs("A", "B", "C")
	node := components.Tabs(components.TabsProps{Items: items, Active: 0})
	header := headerText(t, node)

	if strings.Count(header, "│") != 2 {
		t.Errorf("expected 2 separators in header for 3 tabs, got %q", header)
	}
}

func TestTabsActiveClampedHigh(t *testing.T) {
	items := mkTabs("A", "B", "C")
	// Active out of range should clamp to last.
	node := components.Tabs(components.TabsProps{Items: items, Active: 99})
	header := headerText(t, node)
	if !strings.Contains(header, "[C]") {
		t.Errorf("Active=99 should clamp to last (C), header=%q", header)
	}

	panel := node.Children[1]
	if panel.Text != "panel-C" {
		t.Errorf("panel = %q, want panel-C", panel.Text)
	}
}

func TestTabsActiveClampedLow(t *testing.T) {
	items := mkTabs("A", "B")
	node := components.Tabs(components.TabsProps{Items: items, Active: -10})
	header := headerText(t, node)
	if !strings.Contains(header, "[A]") {
		t.Errorf("Active=-10 should clamp to first (A), header=%q", header)
	}
}

func TestTabsEmptyItems(t *testing.T) {
	node := components.Tabs(components.TabsProps{Items: nil})
	if node == nil {
		t.Fatal("Tabs returned nil for empty items")
	}
	if len(node.Children) != 0 {
		t.Errorf("empty items should render no children, got %d", len(node.Children))
	}
}

func TestTabsNilContentSkipsPanel(t *testing.T) {
	items := []components.TabItem{
		{Label: "Header-only", Content: nil},
		{Label: "WithPanel", Content: vdom.CreateTextNode("body")},
	}
	node := components.Tabs(components.TabsProps{Items: items, Active: 0})
	// Header only — no panel child appended when Content is nil.
	if got := len(node.Children); got != 1 {
		t.Fatalf("nil content should skip panel, children = %d, want 1", got)
	}
}

func TestTabsStateNextWraps(t *testing.T) {
	s := components.NewTabsState(mkTabs("A", "B", "C"))
	s.Next()
	if s.Active != 1 {
		t.Fatalf("Next from 0 = %d, want 1", s.Active)
	}
	s.Next()
	s.Next()
	if s.Active != 0 {
		t.Fatalf("Next x3 from 0 = %d, want 0 (wrap)", s.Active)
	}
}

func TestTabsStatePrevWraps(t *testing.T) {
	s := components.NewTabsState(mkTabs("A", "B", "C"))
	s.Prev()
	if s.Active != 2 {
		t.Fatalf("Prev from 0 = %d, want 2 (wrap)", s.Active)
	}
	s.Prev()
	if s.Active != 1 {
		t.Fatalf("Prev again = %d, want 1", s.Active)
	}
}

func TestTabsStateNextPrevEmpty(t *testing.T) {
	s := components.NewTabsState(nil)
	s.Next()
	s.Prev()
	if s.Active != 0 {
		t.Fatalf("Active = %d on empty, want 0", s.Active)
	}
	s.SetActive(5)
	if s.Active != 0 {
		t.Fatalf("SetActive on empty = %d, want 0", s.Active)
	}
}

func TestTabsStateSetActiveClamp(t *testing.T) {
	s := components.NewTabsState(mkTabs("A", "B", "C"))

	s.SetActive(99)
	if s.Active != 2 {
		t.Fatalf("SetActive(99) = %d, want 2", s.Active)
	}

	s.SetActive(-3)
	if s.Active != 0 {
		t.Fatalf("SetActive(-3) = %d, want 0", s.Active)
	}

	s.SetActive(1)
	if s.Active != 1 {
		t.Fatalf("SetActive(1) = %d, want 1", s.Active)
	}
}

func TestTabsStateCurrent(t *testing.T) {
	items := mkTabs("Logs", "Settings", "About")
	s := components.NewTabsState(items)

	cur, ok := s.Current()
	if !ok {
		t.Fatal("Current() ok=false on fresh state")
	}
	if cur.Label != "Logs" {
		t.Errorf("Current label = %q, want Logs", cur.Label)
	}

	s.Next()
	s.Next()
	cur, ok = s.Current()
	if !ok || cur.Label != "About" {
		t.Errorf("Current after Next x2 = (%q, %v), want (About, true)", cur.Label, ok)
	}
}

func TestTabsStateCurrentEmpty(t *testing.T) {
	s := components.NewTabsState(nil)
	if _, ok := s.Current(); ok {
		t.Fatal("Current() ok=true on empty state")
	}
}

func TestTabsInactiveLabelsDimmed(t *testing.T) {
	items := mkTabs("A", "B", "C")
	node := components.Tabs(components.TabsProps{Items: items, Active: 1})
	header := headerText(t, node)
	// Dim ANSI code (\x1b[2m) should appear for inactive labels.
	if !strings.Contains(header, "\x1b[2m") {
		t.Errorf("inactive labels should be dim, header=%q", header)
	}
}
