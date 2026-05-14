package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

func qsItems() []components.SelectItem {
	return []components.SelectItem{
		{Label: "Apple", Value: "apple"},
		{Label: "Banana", Value: "banana"},
		{Label: "Cherry", Value: "cherry"},
		{Label: "Date", Value: "date"},
		{Label: "Apricot", Value: "apricot"},
	}
}

func TestQuickSearchRendersPromptAndList(t *testing.T) {
	node := components.QuickSearch(components.QuickSearchProps{
		Items:   qsItems(),
		Query:   "",
		Focused: true,
	})
	// children: [prompt text node, select element]
	if len(node.Children) != 2 {
		t.Fatalf("children = %d, want 2 (prompt + select)", len(node.Children))
	}
	if !strings.Contains(node.Children[0].Text, components.DefaultQuickSearchPrompt) {
		t.Fatalf("prompt missing %q in %q", components.DefaultQuickSearchPrompt, node.Children[0].Text)
	}
	// empty query -> all 5 items rendered
	if got := len(node.Children[1].Children); got != 5 {
		t.Fatalf("filtered rows = %d, want 5", got)
	}
}

func TestQuickSearchPromptIncludesQuery(t *testing.T) {
	node := components.QuickSearch(components.QuickSearchProps{
		Items: qsItems(),
		Query: "ap",
	})
	if !strings.Contains(node.Children[0].Text, "ap") {
		t.Fatalf("prompt %q missing query 'ap'", node.Children[0].Text)
	}
}

func TestQuickSearchFilteredRowCount(t *testing.T) {
	node := components.QuickSearch(components.QuickSearchProps{
		Items:    qsItems(),
		Query:    "ap",
		Selected: 0,
		Focused:  true,
	})
	// "Apple" + "Apricot" -> 2 matches
	if got := len(node.Children[1].Children); got != 2 {
		t.Fatalf("filtered rows = %d, want 2", got)
	}
}

func TestQuickSearchEmptyMatch(t *testing.T) {
	node := components.QuickSearch(components.QuickSearchProps{
		Items: qsItems(),
		Query: "xyz",
	})
	if got := len(node.Children[1].Children); got != 0 {
		t.Fatalf("filtered rows = %d, want 0", got)
	}
}

func TestQuickSearchCustomIndicator(t *testing.T) {
	node := components.QuickSearch(components.QuickSearchProps{
		Items:     qsItems(),
		Query:     "",
		Indicator: ">",
		Focused:   true,
	})
	if len(node.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(node.Children))
	}
}

func TestQuickSearchUnfocusedPrompt(t *testing.T) {
	node := components.QuickSearch(components.QuickSearchProps{
		Items:   qsItems(),
		Query:   "a",
		Focused: false,
	})
	// prompt still renders, just dimmed cursor
	if !strings.Contains(node.Children[0].Text, "Search: a") {
		t.Fatalf("prompt missing 'Search: a' in %q", node.Children[0].Text)
	}
}

func TestQuickSearchStateEmptyQueryReturnsAll(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	if got := len(s.Filtered()); got != 5 {
		t.Fatalf("Filtered() = %d, want 5", got)
	}
}

func TestQuickSearchStatePartialMatch(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.SetQuery("an")
	// "Banana" only
	got := s.Filtered()
	if len(got) != 1 {
		t.Fatalf("Filtered len = %d, want 1", len(got))
	}
	if got[0].Label != "Banana" {
		t.Fatalf("Filtered[0] = %q, want Banana", got[0].Label)
	}
}

func TestQuickSearchStateCaseInsensitive(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.SetQuery("CHERRY")
	got := s.Filtered()
	if len(got) != 1 || got[0].Value != "cherry" {
		t.Fatalf("case-insensitive lookup failed: %+v", got)
	}

	s.SetQuery("ap")
	got = s.Filtered()
	if len(got) != 2 {
		t.Fatalf("expected 2 matches for 'ap', got %d", len(got))
	}
}

func TestQuickSearchStateSetQueryResetsSelected(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.MoveDown()
	s.MoveDown()
	if s.Selected != 2 {
		t.Fatalf("pre-condition: Selected = %d, want 2", s.Selected)
	}
	s.SetQuery("a")
	if s.Selected != 0 {
		t.Fatalf("SetQuery should reset Selected to 0, got %d", s.Selected)
	}
}

func TestQuickSearchStateAppendAndBackspace(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.AppendQuery('a')
	s.AppendQuery('p')
	if s.Query != "ap" {
		t.Fatalf("Query = %q, want ap", s.Query)
	}
	if got := len(s.Filtered()); got != 2 {
		t.Fatalf("Filtered len = %d, want 2", got)
	}

	s.MoveDown() // move within filtered
	if s.Selected != 1 {
		t.Fatalf("MoveDown: Selected = %d, want 1", s.Selected)
	}

	s.BackspaceQuery()
	if s.Query != "a" {
		t.Fatalf("after backspace Query = %q, want a", s.Query)
	}
	if s.Selected != 0 {
		t.Fatalf("backspace should reset Selected, got %d", s.Selected)
	}

	s.BackspaceQuery()
	if s.Query != "" {
		t.Fatalf("after second backspace Query = %q, want empty", s.Query)
	}

	// safe on empty
	s.BackspaceQuery()
	if s.Query != "" {
		t.Fatalf("backspace on empty should be no-op, got %q", s.Query)
	}
}

func TestQuickSearchStateAppendQueryUnicode(t *testing.T) {
	s := components.NewQuickSearchState(
		[]components.SelectItem{
			{Label: "한국어", Value: "ko"},
			{Label: "English", Value: "en"},
		},
		0,
	)
	s.AppendQuery('한')
	if got := len(s.Filtered()); got != 1 {
		t.Fatalf("Filtered len = %d, want 1", got)
	}
	s.BackspaceQuery()
	if s.Query != "" {
		t.Fatalf("Backspace on multi-byte rune left %q", s.Query)
	}
}

func TestQuickSearchStateMoveWrapOnFiltered(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.SetQuery("ap") // 2 matches: Apple, Apricot
	s.MoveUp()       // wrap to last (1)
	if s.Selected != 1 {
		t.Fatalf("MoveUp wrap: Selected = %d, want 1", s.Selected)
	}
	s.MoveDown() // wrap to first (0)
	if s.Selected != 0 {
		t.Fatalf("MoveDown wrap: Selected = %d, want 0", s.Selected)
	}
	s.MoveDown()
	if s.Selected != 1 {
		t.Fatalf("MoveDown: Selected = %d, want 1", s.Selected)
	}
}

func TestQuickSearchStateMoveOnEmptyFilteredNoOp(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.SetQuery("zzzz")
	s.MoveUp()
	s.MoveDown()
	if s.Selected != 0 {
		t.Fatalf("Selected on empty filter = %d, want 0", s.Selected)
	}
}

func TestQuickSearchStateMoveOnEmptyItemsNoOp(t *testing.T) {
	s := components.NewQuickSearchState(nil, 0)
	s.MoveUp()
	s.MoveDown()
	if s.Selected != 0 {
		t.Fatalf("Selected on empty items = %d, want 0", s.Selected)
	}
}

func TestQuickSearchStateConfirmHighlighted(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.SetQuery("ap")
	s.MoveDown() // select Apricot (index 1 in filtered)
	item, ok := s.Confirm()
	if !ok {
		t.Fatal("Confirm ok=false on valid filtered list")
	}
	if item.Label != "Apricot" {
		t.Fatalf("Confirm = %q, want Apricot", item.Label)
	}
}

func TestQuickSearchStateConfirmEmptyFilteredFalse(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.SetQuery("zzz")
	if _, ok := s.Confirm(); ok {
		t.Fatal("Confirm should return false on empty filtered list")
	}
}

func TestQuickSearchStateConfirmOutOfRangeFalse(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 0)
	s.SetQuery("zzz") // empty filter -> Confirm false
	if _, ok := s.Confirm(); ok {
		t.Fatal("expected false on empty filtered list")
	}

	// Set up an out-of-range Selected manually.
	s.SetQuery("")
	s.Selected = 999
	if _, ok := s.Confirm(); ok {
		t.Fatal("Confirm with Selected=999 should be false")
	}
	s.Selected = -1
	if _, ok := s.Confirm(); ok {
		t.Fatal("Confirm with Selected=-1 should be false")
	}
}

func TestQuickSearchStateLimitPreserved(t *testing.T) {
	s := components.NewQuickSearchState(qsItems(), 3)
	if s.Limit != 3 {
		t.Fatalf("Limit = %d, want 3", s.Limit)
	}
}
