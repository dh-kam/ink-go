package components_test

import (
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

func mkItems(n int) []components.SelectItem {
	out := make([]components.SelectItem, n)
	for i := range out {
		out[i].Label = string(rune('A' + i))
		out[i].Value = string(rune('a' + i))
	}
	return out
}

func TestSelectRendersChildrenPerItem(t *testing.T) {
	items := mkItems(3)
	node := components.Select(components.SelectProps{Items: items, Selected: 1, Focused: true})
	if got := len(node.Children); got != 3 {
		t.Fatalf("rows = %d, want 3", got)
	}
}

func TestSelectEmpty(t *testing.T) {
	node := components.Select(components.SelectProps{Items: nil})
	if got := len(node.Children); got != 0 {
		t.Fatalf("rows = %d, want 0 for empty", got)
	}
}

func TestSelectLimitWithScrollIndicators(t *testing.T) {
	items := mkItems(10)
	// Selected near the bottom — should show ↑ indicator (offset > 0)
	node := components.Select(components.SelectProps{Items: items, Selected: 9, Limit: 3})
	// 3 visible rows + 1 ↑ indicator = 4
	if got := len(node.Children); got != 4 {
		t.Fatalf("rows = %d, want 4 (3 items + ↑)", got)
	}
}

func TestSelectStateMoveUpDownWrap(t *testing.T) {
	s := components.NewSelectState(mkItems(3), 0)
	s.MoveUp()
	if s.Selected != 2 {
		t.Fatalf("MoveUp from 0 = %d, want 2 (wrap)", s.Selected)
	}
	s.MoveDown()
	if s.Selected != 0 {
		t.Fatalf("MoveDown from 2 = %d, want 0 (wrap)", s.Selected)
	}
	s.MoveDown()
	if s.Selected != 1 {
		t.Fatalf("MoveDown from 0 = %d, want 1", s.Selected)
	}
}

func TestSelectStateEmptyMovesNoOp(t *testing.T) {
	s := components.NewSelectState(nil, 0)
	s.MoveUp()
	s.MoveDown()
	if s.Selected != 0 {
		t.Fatalf("Selected = %d, want 0", s.Selected)
	}
}

func TestSelectStateClamp(t *testing.T) {
	s := components.NewSelectState(mkItems(3), 0)
	s.MoveTo(99)
	if s.Selected != 2 {
		t.Fatalf("Selected = %d, want 2 after clamp", s.Selected)
	}
	s.MoveTo(-5)
	if s.Selected != 0 {
		t.Fatalf("Selected = %d, want 0 after clamp", s.Selected)
	}
}

func TestSelectStateConfirm(t *testing.T) {
	s := components.NewSelectState(mkItems(3), 0)
	s.MoveTo(1)
	item, ok := s.Confirm()
	if !ok {
		t.Fatal("Confirm returned ok=false on valid state")
	}
	if item.Label != "B" {
		t.Fatalf("Confirm label = %q, want B", item.Label)
	}
}

func TestSelectStateVisibleNoLimit(t *testing.T) {
	items := mkItems(5)
	s := components.NewSelectState(items, 0)
	v, off := s.Visible()
	if off != 0 || len(v) != 5 {
		t.Fatalf("Visible() = (len=%d, off=%d), want (5, 0)", len(v), off)
	}
}

func TestSelectStateVisibleWindow(t *testing.T) {
	items := mkItems(10)
	s := components.NewSelectState(items, 3)

	// Selection at top — window starts at 0
	v, off := s.Visible()
	if off != 0 || len(v) != 3 {
		t.Errorf("at top: off=%d len=%d, want 0/3", off, len(v))
	}

	// Selection at bottom — window pinned to end
	s.MoveTo(9)
	v, off = s.Visible()
	if off != 7 || len(v) != 3 {
		t.Errorf("at bottom: off=%d len=%d, want 7/3", off, len(v))
	}

	// Middle — selection centered
	s.MoveTo(5)
	_, off = s.Visible()
	if off != 4 {
		t.Errorf("middle: off=%d, want 4 (centered)", off)
	}
}

func TestSelectStateCurrentEmpty(t *testing.T) {
	s := components.NewSelectState(nil, 0)
	if _, ok := s.Current(); ok {
		t.Fatal("Current() ok=true on empty state")
	}
}

func TestSelectIndicatorDefault(t *testing.T) {
	items := mkItems(2)
	node := components.Select(components.SelectProps{Items: items})
	// Just verify it renders without panicking and has rows
	if len(node.Children) == 0 {
		t.Fatal("default indicator path produced no rows")
	}
}

func TestSelectCustomIndicator(t *testing.T) {
	items := mkItems(2)
	node := components.Select(components.SelectProps{
		Items:     items,
		Selected:  0,
		Indicator: ">",
		Focused:   true,
	})
	if len(node.Children) != 2 {
		t.Fatalf("rows = %d, want 2", len(node.Children))
	}
}
