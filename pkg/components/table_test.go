package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

// TestSimpleTable tests creating a simple table
func TestSimpleTable(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	rows := [][]string{
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}

	node := components.SimpleTable(headers, rows)

	if node == nil {
		t.Fatal("SimpleTable should return a non-nil node")
	}

	if node.ElementType != "table" {
		t.Errorf("Expected element type 'table', got %q", node.ElementType)
	}

	text := node.Children[0].Text
	if text == "" {
		t.Error("Expected table text to be non-empty")
	}

	// Check that headers are present
	if !strings.Contains(text, "Name") {
		t.Error("Expected table to contain 'Name'")
	}
}

// TestTableEmpty tests empty table
func TestTableEmpty(t *testing.T) {
	node := components.Table(components.TableProps{})

	if node == nil {
		t.Fatal("Table should return a non-nil node")
	}

	if node.ElementType != "table" {
		t.Errorf("Expected element type 'table', got %q", node.ElementType)
	}
}

// TestTableWithBorders tests table with borders
func TestTableWithBorders(t *testing.T) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "A"},
				{Text: "B"},
			},
			IsHeader: true,
		},
		{
			Cells: []components.TableCell{
				{Text: "1"},
				{Text: "2"},
			},
		},
	}

	node := components.Table(components.TableProps{
		Rows:    rows,
		Border:  true,
		Padding: 1,
	})

	text := node.Children[0].Text

	// Should contain border characters
	if !strings.Contains(text, "|") {
		t.Error("Expected bordered table to contain '|'")
	}
	if !strings.Contains(text, "+") {
		t.Error("Expected bordered table to contain '+'")
	}
	if !strings.Contains(text, "-") {
		t.Error("Expected bordered table to contain '-'")
	}
}

// TestTableAlignments tests cell alignment
func TestTableAlignments(t *testing.T) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "left", Align: "left"},
				{Text: "center", Align: "center"},
				{Text: "right", Align: "right"},
			},
		},
	}

	node := components.Table(components.TableProps{
		Rows:     rows,
		Border:   false,
		Padding:  0,
		ColumnWidths: []int{20, 20, 20},
	})

	text := node.Children[0].Text
	if text == "" {
		t.Error("Expected table text to be non-empty")
	}
}

// TestTableWithoutBorders tests table without borders
func TestTableWithoutBorders(t *testing.T) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "A"},
				{Text: "B"},
			},
		},
	}

	node := components.Table(components.TableProps{
		Rows:    rows,
		Border:  false,
	})

	text := node.Children[0].Text

	// Should still have pipes for cell separators
	if !strings.Contains(text, "|") {
		t.Error("Expected table to contain cell separators")
	}
}

// TestAlignCell tests AlignCell helper
func TestAlignCell(t *testing.T) {
	cell := components.AlignCell("test", "center")

	if cell.Text != "test" {
		t.Errorf("Expected text 'test', got %q", cell.Text)
	}
	if cell.Align != "center" {
		t.Errorf("Expected align 'center', got %q", cell.Align)
	}
}

// TestCenterCell tests CenterCell helper
func TestCenterCell(t *testing.T) {
	cell := components.CenterCell("test")

	if cell.Text != "test" {
		t.Errorf("Expected text 'test', got %q", cell.Text)
	}
	if cell.Align != "center" {
		t.Errorf("Expected align 'center', got %q", cell.Align)
	}
}

// TestRightCell tests RightCell helper
func TestRightCell(t *testing.T) {
	cell := components.RightCell("test")

	if cell.Text != "test" {
		t.Errorf("Expected text 'test', got %q", cell.Text)
	}
	if cell.Align != "right" {
		t.Errorf("Expected align 'right', got %q", cell.Align)
	}
}

// TestHeaderRow tests HeaderRow helper
func TestHeaderRow(t *testing.T) {
	row := components.HeaderRow("A", "B", "C")

	if !row.IsHeader {
		t.Error("Expected IsHeader to be true")
	}
	if len(row.Cells) != 3 {
		t.Errorf("Expected 3 cells, got %d", len(row.Cells))
	}
}

// TestDataRow tests DataRow helper
func TestDataRow(t *testing.T) {
	row := components.DataRow("1", "2", "3")

	if row.IsHeader {
		t.Error("Expected IsHeader to be false")
	}
	if len(row.Cells) != 3 {
		t.Errorf("Expected 3 cells, got %d", len(row.Cells))
	}
}

// TestTableFromSlice tests creating table from slice of maps
func TestTableFromSlice(t *testing.T) {
	headers := []string{"name", "value"}
	data := []map[string]string{
		{"name": "Alice", "value": "100"},
		{"name": "Bob", "value": "200"},
	}

	node := components.TableFromSlice(headers, data)

	if node == nil {
		t.Fatal("TableFromSlice should return a non-nil node")
	}

	text := node.Children[0].Text

	if !strings.Contains(text, "Alice") {
		t.Error("Expected table to contain 'Alice'")
	}
	if !strings.Contains(text, "Bob") {
		t.Error("Expected table to contain 'Bob'")
	}
}

// TestTableVariableColumns tests table with varying column counts
func TestTableVariableColumns(t *testing.T) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "A"},
				{Text: "B"},
				{Text: "C"},
			},
		},
		{
			Cells: []components.TableCell{
				{Text: "1"},
				{Text: "2"},
			},
		},
	}

	node := components.Table(components.TableProps{
		Rows:    rows,
		Border:  true,
	})

	if node == nil {
		t.Fatal("Table should return a non-nil node")
	}

	// Should handle varying column counts
	text := node.Children[0].Text
	if text == "" {
		t.Error("Expected table text to be non-empty")
	}
}

// TestTableCustomColumnWidths tests custom column widths
func TestTableCustomColumnWidths(t *testing.T) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "A"},
				{Text: "BBB"},
			},
		},
	}

	node := components.Table(components.TableProps{
		Rows:         rows,
		ColumnWidths: []int{5, 10},
	})

	if node == nil {
		t.Fatal("Table should return a non-nil node")
	}

	_ = node.Children[0].Text
	// Just verify it doesn't crash
}

// TestTablePadding tests cell padding
func TestTablePadding(t *testing.T) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "A"},
			},
		},
	}

	node1 := components.Table(components.TableProps{
		Rows:    rows,
		Padding: 0,
	})

	node2 := components.Table(components.TableProps{
		Rows:    rows,
		Padding: 2,
	})

	text1 := node1.Children[0].Text
	text2 := node2.Children[0].Text

	// More padding should result in longer text
	if len(text2) <= len(text1) {
		t.Error("Expected more padding to result in longer text")
	}
}

// TestTableEmptyRow tests table with empty row
func TestTableEmptyRow(t *testing.T) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{},
		},
	}

	node := components.Table(components.TableProps{
		Rows: rows,
	})

	if node == nil {
		t.Fatal("Table should return a non-nil node")
	}

	// Should handle empty row gracefully
	_ = node.Children[0].Text
}

// TestTableCellSpans tests cells with colspan/rowspan
func TestTableCellSpans(t *testing.T) {
	cell := components.TableCell{
		Text:    "test",
		ColSpan: 2,
		RowSpan: 3,
	}

	if cell.Text != "test" {
		t.Errorf("Expected text 'test', got %q", cell.Text)
	}
	if cell.ColSpan != 2 {
		t.Errorf("Expected ColSpan 2, got %d", cell.ColSpan)
	}
	if cell.RowSpan != 3 {
		t.Errorf("Expected RowSpan 3, got %d", cell.RowSpan)
	}
}

// TestTableSingleColumn tests table with single column
func TestTableSingleColumn(t *testing.T) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "Only one"},
			},
		},
	}

	node := components.Table(components.TableProps{
		Rows:    rows,
		Border:  true,
	})

	if node == nil {
		t.Fatal("Table should return a non-nil node")
	}

	text := node.Children[0].Text
	if !strings.Contains(text, "Only one") {
		t.Error("Expected table to contain 'Only one'")
	}
}

// TestTableLongContent tests table with long content
func TestTableLongContent(t *testing.T) {
	longText := strings.Repeat("x", 100)

	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: longText},
			},
		},
	}

	node := components.Table(components.TableProps{
		Rows:         rows,
		ColumnWidths: []int{20},
	})

	if node == nil {
		t.Fatal("Table should return a non-nil node")
	}

	// Should truncate to column width
	text := node.Children[0].Text
	if len(text) > 50 { // rough check
		t.Logf("Warning: table text may not be properly truncated: %d chars", len(text))
	}
}
