package components_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

func mkMultiItems(n int) []components.MultiSelectItem {
	out := make([]components.MultiSelectItem, n)
	for i := range out {
		out[i] = components.MultiSelectItem{
			Label: string(rune('A' + i)),
			Value: string(rune('a' + i)),
		}
	}
	return out
}

func TestMultiSelectEmpty(t *testing.T) {
	node := components.MultiSelect(components.MultiSelectProps{Items: nil})
	if len(node.Children) != 0 {
		t.Fatalf("rows = %d, want 0 for empty", len(node.Children))
	}
}

func TestMultiSelectRendersAllRows(t *testing.T) {
	items := mkMultiItems(3)
	node := components.MultiSelect(components.MultiSelectProps{
		Items:    items,
		Selected: []string{"b"},
		Cursor:   1,
		Focused:  true,
	})
	if len(node.Children) != 3 {
		t.Fatalf("rows = %d, want 3", len(node.Children))
	}

	var rendered strings.Builder
	for _, c := range node.Children {
		rendered.WriteString(c.Text)
		rendered.WriteString("\n")
	}
	out := rendered.String()
	if !strings.Contains(out, "[x]") {
		t.Fatalf("expected at least one [x] glyph, got:\n%s", out)
	}
	if !strings.Contains(out, "[ ]") {
		t.Fatalf("expected at least one [ ] glyph, got:\n%s", out)
	}
}

func TestMultiSelectLimitWithScrollIndicators(t *testing.T) {
	items := mkMultiItems(10)
	node := components.MultiSelect(components.MultiSelectProps{
		Items:  items,
		Cursor: 9,
		Limit:  3,
	})
	// 3 visible rows + 1 ↑ indicator
	if len(node.Children) != 4 {
		t.Fatalf("rows = %d, want 4 (3 items + ↑)", len(node.Children))
	}

	// Cursor at top → ↓ indicator
	node = components.MultiSelect(components.MultiSelectProps{
		Items:  items,
		Cursor: 0,
		Limit:  3,
	})
	if len(node.Children) != 4 {
		t.Fatalf("rows = %d, want 4 (3 items + ↓)", len(node.Children))
	}
}

func TestMultiSelectStateMoveUpDownWrap(t *testing.T) {
	s := components.NewMultiSelectState(mkMultiItems(3), 0)
	s.MoveUp()
	if s.Cursor != 2 {
		t.Fatalf("MoveUp from 0 = %d, want 2 (wrap)", s.Cursor)
	}
	s.MoveDown()
	if s.Cursor != 0 {
		t.Fatalf("MoveDown from 2 = %d, want 0 (wrap)", s.Cursor)
	}
	s.MoveDown()
	if s.Cursor != 1 {
		t.Fatalf("MoveDown from 0 = %d, want 1", s.Cursor)
	}
}

func TestMultiSelectStateEmptyMovesNoOp(t *testing.T) {
	s := components.NewMultiSelectState(nil, 0)
	s.MoveUp()
	s.MoveDown()
	s.Toggle()
	if s.Cursor != 0 {
		t.Fatalf("Cursor = %d, want 0", s.Cursor)
	}
	if len(s.Selected) != 0 {
		t.Fatalf("Selected len = %d, want 0", len(s.Selected))
	}
}

func TestMultiSelectStateToggle(t *testing.T) {
	items := mkMultiItems(3)
	s := components.NewMultiSelectState(items, 0)
	s.Toggle() // check 'a'
	if !s.Selected["a"] {
		t.Fatal("expected 'a' to be checked")
	}
	s.Toggle() // uncheck 'a'
	if s.Selected["a"] {
		t.Fatal("expected 'a' to be unchecked after second Toggle")
	}
}

func TestMultiSelectStateToggleSkipsDisabled(t *testing.T) {
	items := []components.MultiSelectItem{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b", Disabled: true},
	}
	s := components.NewMultiSelectState(items, 0)
	s.MoveDown() // cursor → b (disabled)
	s.Toggle()
	if s.Selected["b"] {
		t.Fatal("disabled item must not be toggled")
	}
}

func TestMultiSelectStateValuesPreservesItemOrder(t *testing.T) {
	items := mkMultiItems(5) // a..e
	s := components.NewMultiSelectState(items, 0)
	// Toggle in mixed order.
	s.MoveDown()
	s.MoveDown()
	s.Toggle() // c
	s.MoveUp()
	s.MoveUp()
	s.Toggle() // a
	s.MoveDown()
	s.MoveDown()
	s.MoveDown()
	s.MoveDown()
	s.Toggle() // e
	got := s.Values()
	want := []string{"a", "c", "e"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Values = %v, want %v", got, want)
	}
}

func TestMultiSelectStateValuesEmpty(t *testing.T) {
	s := components.NewMultiSelectState(mkMultiItems(3), 0)
	got := s.Values()
	if len(got) != 0 {
		t.Fatalf("Values len = %d, want 0", len(got))
	}
	if got == nil {
		t.Fatal("Values must return a non-nil slice")
	}
}

func TestMultiSelectStateSelectAllSkipsDisabled(t *testing.T) {
	items := []components.MultiSelectItem{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b", Disabled: true},
		{Label: "C", Value: "c"},
	}
	s := components.NewMultiSelectState(items, 0)
	s.SelectAll()
	if !s.Selected["a"] || !s.Selected["c"] {
		t.Fatalf("expected a+c selected, got %v", s.Selected)
	}
	if s.Selected["b"] {
		t.Fatal("disabled 'b' must not be selected")
	}
	got := s.Values()
	want := []string{"a", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Values = %v, want %v", got, want)
	}
}

func TestMultiSelectStateClearAll(t *testing.T) {
	s := components.NewMultiSelectState(mkMultiItems(3), 0)
	s.SelectAll()
	if len(s.Selected) == 0 {
		t.Fatal("setup: SelectAll didn't populate Selected")
	}
	s.ClearAll()
	if len(s.Selected) != 0 {
		t.Fatalf("Selected len = %d after ClearAll, want 0", len(s.Selected))
	}
}

func TestMultiSelectStateVisibleNoLimit(t *testing.T) {
	items := mkMultiItems(5)
	s := components.NewMultiSelectState(items, 0)
	v, off := s.Visible()
	if off != 0 || len(v) != 5 {
		t.Fatalf("Visible() = (len=%d, off=%d), want (5, 0)", len(v), off)
	}
}

func TestMultiSelectStateVisibleWindowCenters(t *testing.T) {
	items := mkMultiItems(10)
	s := components.NewMultiSelectState(items, 3)

	// At top — window starts at 0
	v, off := s.Visible()
	if off != 0 || len(v) != 3 {
		t.Errorf("at top: off=%d len=%d, want 0/3", off, len(v))
	}

	// At bottom — window pinned to end
	for i := 0; i < 9; i++ {
		s.MoveDown()
	}
	v, off = s.Visible()
	if off != 7 || len(v) != 3 {
		t.Errorf("at bottom: off=%d len=%d, want 7/3", off, len(v))
	}

	// Middle — selection centered
	s.Cursor = 5
	_, off = s.Visible()
	if off != 4 {
		t.Errorf("middle: off=%d, want 4 (centered)", off)
	}
}

func TestMultiSelectRenderShowsCheckedGlyphForSelected(t *testing.T) {
	items := mkMultiItems(2)
	node := components.MultiSelect(components.MultiSelectProps{
		Items:    items,
		Selected: []string{"a", "b"},
		Cursor:   0,
		Focused:  true,
	})
	if len(node.Children) != 2 {
		t.Fatalf("rows = %d, want 2", len(node.Children))
	}
	for i, c := range node.Children {
		if !strings.Contains(c.Text, "[x]") {
			t.Errorf("row %d missing [x]: %q", i, c.Text)
		}
	}
}

func TestMultiSelectRendersDisabledDimmed(t *testing.T) {
	items := []components.MultiSelectItem{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b", Disabled: true},
	}
	node := components.MultiSelect(components.MultiSelectProps{
		Items:   items,
		Cursor:  0,
		Focused: true,
	})
	if len(node.Children) != 2 {
		t.Fatalf("rows = %d, want 2", len(node.Children))
	}
	// Second row should include the dim ANSI code (ESC[2m).
	if !strings.Contains(node.Children[1].Text, "\x1b[2m") {
		t.Fatalf("disabled row not dimmed: %q", node.Children[1].Text)
	}
}
