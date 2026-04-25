package components

import (
	"strings"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TableCell represents a single cell in a table
type TableCell struct {
	Text   string
	Align  string // "left", "center", "right"
	ColSpan int
	RowSpan int
}

// TableRow represents a row of table cells
type TableRow struct {
	Cells    []TableCell
	IsHeader bool
}

// TableProps defines the properties for a Table component
type TableProps struct {
	// Rows contains all the table rows
	Rows []TableRow
	// ColumnWidths specifies the width for each column (0 = auto)
	ColumnWidths []int
	// Border controls whether to show borders
	Border bool
	// Padding is the cell padding
	Padding int
}

// Table creates a table component
func Table(props TableProps) *vdom.Node {
	if len(props.Rows) == 0 {
		return vdom.CreateElement("table", nil)
	}

	// Calculate column widths
	colWidths := props.ColumnWidths
	maxCols := 0
	for _, row := range props.Rows {
		if len(row.Cells) > maxCols {
			maxCols = len(row.Cells)
		}
	}

	// Fill in missing column widths
	if len(colWidths) < maxCols {
		for i := len(colWidths); i < maxCols; i++ {
			width := 20 // default auto width
			if i < len(colWidths) {
				width = colWidths[i]
			}
			colWidths = append(colWidths, width)
		}
	}

	// Build the table
	var lines []string

	for i, row := range props.Rows {
		line := buildTableRow(row, colWidths, props.Padding)
		lines = append(lines, line)

		// Add separator after header or between rows (if border is enabled)
		if props.Border && (row.IsHeader || i < len(props.Rows)-1) {
			separator := buildTableSeparator(colWidths, props.Padding)
			lines = append(lines, separator)
		}
	}

	// Join all lines
	tableText := strings.Join(lines, "\n")

	return vdom.CreateElement("table", nil, vdom.CreateTextNode(tableText))
}

// buildTableRow builds a single table row as a string
func buildTableRow(row TableRow, colWidths []int, padding int) string {
	if len(row.Cells) == 0 {
		return ""
	}

	paddingStr := strings.Repeat(" ", padding)

	var parts []string
	for i, cell := range row.Cells {
		width := colWidths[i]
		if width <= 0 {
			width = 20 // default
		}

		// Apply padding
		content := paddingStr + cell.Text + paddingStr

		// Apply alignment and padding to reach desired width
		totalWidth := width + 2*padding
		content = alignText(content, cell.Align, totalWidth)

		parts = append(parts, content)
	}

	return "|" + strings.Join(parts, "|") + "|"
}

// buildTableSeparator builds a separator line for the table
func buildTableSeparator(colWidths []int, padding int) string {
	var parts []string
	for _, width := range colWidths {
		if width <= 0 {
			width = 20
		}
		separator := strings.Repeat("-", width+2*padding)
		parts = append(parts, separator)
	}

	return "+" + strings.Join(parts, "+") + "+"
}

// alignText aligns text within the given width
func alignText(text, align string, width int) string {
	if len(text) >= width {
		return text[:width]
	}

	padding := width - len(text)
	switch align {
	case "center":
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	case "right":
		return strings.Repeat(" ", padding) + text
	default: // left
		return text + strings.Repeat(" ", padding)
	}
}

// SimpleTable creates a simple table from a 2D string array
func SimpleTable(headers []string, rows [][]string) *vdom.Node {
	tableRows := []TableRow{}

	// Add header row
	if len(headers) > 0 {
		headerCells := []TableCell{}
		for _, h := range headers {
			headerCells = append(headerCells, TableCell{
				Text:  h,
				Align: "left",
			})
		}
		tableRows = append(tableRows, TableRow{
			Cells:    headerCells,
			IsHeader: true,
		})
	}

	// Add data rows
	for _, row := range rows {
		cells := []TableCell{}
		for _, cell := range row {
			cells = append(cells, TableCell{
				Text:  cell,
				Align: "left",
			})
		}
		tableRows = append(tableRows, TableRow{
			Cells: cells,
		})
	}

	return Table(TableProps{
		Rows:    tableRows,
		Border:  true,
		Padding: 1,
	})
}

// TableFromSlice creates a table from a slice of maps
func TableFromSlice(headers []string, data []map[string]string) *vdom.Node {
	rows := [][]string{}
	for _, item := range data {
		row := []string{}
		for _, h := range headers {
			row = append(row, item[h])
		}
		rows = append(rows, row)
	}
	return SimpleTable(headers, rows)
}

// AlignCell creates a TableCell with specified alignment
func AlignCell(text string, align string) TableCell {
	return TableCell{
		Text:  text,
		Align: align,
	}
}

// CenterCell creates a centered table cell
func CenterCell(text string) TableCell {
	return TableCell{
		Text:  text,
		Align: "center",
	}
}

// RightCell creates a right-aligned table cell
func RightCell(text string) TableCell {
	return TableCell{
		Text:  text,
		Align: "right",
	}
}

// HeaderRow creates a header row from strings
func HeaderRow(cells ...string) TableRow {
	rowCells := []TableCell{}
	for _, c := range cells {
		rowCells = append(rowCells, TableCell{
			Text:  c,
			Align: "left",
		})
	}
	return TableRow{
		Cells:    rowCells,
		IsHeader: true,
	}
}

// DataRow creates a data row from strings
func DataRow(cells ...string) TableRow {
	rowCells := []TableCell{}
	for _, c := range cells {
		rowCells = append(rowCells, TableCell{
			Text:  c,
			Align: "left",
		})
	}
	return TableRow{
		Cells: rowCells,
	}
}
