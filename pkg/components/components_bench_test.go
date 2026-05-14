package components_test

import (
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

// BenchmarkSpinner benchmarks spinner creation
func BenchmarkSpinner(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = components.Spinner(components.SpinnerProps{
			Type: components.DotsSpinner,
			Text: "Loading",
		})
	}
}

// BenchmarkSpinnerSimple benchmarks simple spinner creation
func BenchmarkSpinnerSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = components.SpinnerWithText("Loading...")
	}
}

// BenchmarkProgressBar benchmarks progress bar creation
func BenchmarkProgressBar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = components.ProgressBar(components.ProgressBarProps{
			Percent:     50,
			Width:       40,
			ShowPercent: true,
		})
	}
}

// BenchmarkProgressBarSimple benchmarks simple progress bar
func BenchmarkProgressBarSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = components.ProgressBarSimple(75)
	}
}

// BenchmarkTextInput benchmarks text input creation
func BenchmarkTextInput(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = components.TextInput(components.TextInputProps{
			Value:       "test input",
			Placeholder: "Enter text",
			Width:       30,
		})
	}
}

// BenchmarkTable benchmarks table creation
func BenchmarkTable(b *testing.B) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "Name"},
				{Text: "Age"},
			},
			IsHeader: true,
		},
		{
			Cells: []components.TableCell{
				{Text: "Alice"},
				{Text: "30"},
			},
		},
	}

	props := components.TableProps{
		Rows:    rows,
		Border:  true,
		Padding: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = components.Table(props)
	}
}

// BenchmarkSimpleTable benchmarks simple table creation
func BenchmarkSimpleTable(b *testing.B) {
	headers := []string{"Name", "Age", "City"}
	rows := [][]string{
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = components.SimpleTable(headers, rows)
	}
}

// BenchmarkBox benchmarks box component creation
func BenchmarkBox(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = components.Box(nil)
	}
}

// BenchmarkBorder benchmarks border component creation
func BenchmarkBorder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = components.Border(components.BorderProps{
			Style: "double",
		}, nil)
	}
}

// BenchmarkTextInputStateInsert benchmarks text insertion
func BenchmarkTextInputStateInsert(b *testing.B) {
	state := components.NewTextInputState("")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.Insert("x")
	}
}

// BenchmarkTextInputStateDelete benchmarks text deletion
func BenchmarkTextInputStateDelete(b *testing.B) {
	state := components.NewTextInputState("hello world this is a test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.Delete()
	}
}

// BenchmarkTextInputStateMove benchmarks cursor movement
func BenchmarkTextInputStateMove(b *testing.B) {
	state := components.NewTextInputState("hello world")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.MoveLeft()
		state.MoveRight()
	}
}

// BenchmarkSpinnerTypeChange benchmarks switching spinner types
func BenchmarkSpinnerTypeChange(b *testing.B) {
	types := []*components.SpinnerFrames{
		components.DotsSpinner,
		components.LineSpinner,
		components.ArrowSpinner,
		components.PlusMinusSpinner,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, t := range types {
			_ = components.Spinner(components.SpinnerProps{
				Type: t,
			})
		}
	}
}

// BenchmarkProgressBarUpdate benchmarks updating progress bar
func BenchmarkProgressBarUpdate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		percent := (i * 100) / b.N
		_ = components.ProgressBar(components.ProgressBarProps{
			Percent:     percent,
			Width:       40,
			ShowPercent: true,
		})
	}
}

// BenchmarkTableWithManyRows benchmarks table with many rows
func BenchmarkTableWithManyRows(b *testing.B) {
	rows := []components.TableRow{
		{
			Cells: []components.TableCell{
				{Text: "Col1"},
				{Text: "Col2"},
				{Text: "Col3"},
			},
			IsHeader: true,
		},
	}

	// Add 100 data rows
	for i := 0; i < 100; i++ {
		rows = append(rows, components.TableRow{
			Cells: []components.TableCell{
				{Text: "Data1"},
				{Text: "Data2"},
				{Text: "Data3"},
			},
		})
	}

	props := components.TableProps{
		Rows:    rows,
		Border:  true,
		Padding: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = components.Table(props)
	}
}
