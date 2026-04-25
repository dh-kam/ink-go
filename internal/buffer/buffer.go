package buffer

import (
	"strings"
	"unicode/utf8"

	"github.com/dh-kam/goink.go/pkg/utils"
)

// Buffer represents a 2D character buffer for terminal rendering
type Buffer struct {
	width  int
	height int
	cells  [][]string
}

const undefinedCell = "\x00"

// New creates a new buffer with the specified dimensions
func New(width, height int) *Buffer {
	cells := make([][]string, height)
	for i := range cells {
		cells[i] = make([]string, width)
		// Initialize with spaces
		for j := range cells[i] {
			cells[i][j] = " "
		}
	}

	return &Buffer{
		width:  width,
		height: height,
		cells:  cells,
	}
}

func (b *Buffer) ensureRowWidth(y, width int) {
	if y < 0 || y >= b.height {
		return
	}

	if width <= len(b.cells[y]) {
		return
	}

	extra := make([]string, width-len(b.cells[y]))
	b.cells[y] = append(b.cells[y], extra...)
}

func (b *Buffer) renderRow(y int) string {
	if y < 0 || y >= b.height {
		return ""
	}

	var row strings.Builder
	for _, ch := range b.cells[y] {
		if ch == undefinedCell {
			continue
		}

		row.WriteString(ch)
	}

	return strings.TrimRight(row.String(), " ")
}

// Width returns the buffer width
func (b *Buffer) Width() int {
	return b.width
}

// Height returns the buffer height
func (b *Buffer) Height() int {
	return b.height
}

// Set sets a character at the specified position
func (b *Buffer) Set(x, y int, ch rune) {
	if x < 0 || y < 0 || y >= b.height {
		return // Out of bounds
	}

	width := utils.RuneWidth(ch)
	if width == 0 {
		b.appendZeroWidth(x, y, ch)
		return
	}

	b.ensureRowWidth(y, x+width)
	b.cells[y][x] = string(ch)
	for offset := 1; offset < width; offset++ {
		b.cells[y][x+offset] = undefinedCell
	}
}

func (b *Buffer) appendZeroWidth(x, y int, ch rune) {
	if y < 0 || y >= b.height || x <= 0 {
		return
	}

	for index := x - 1; index >= 0; index-- {
		if index >= len(b.cells[y]) {
			continue
		}

		if b.cells[y][index] == undefinedCell {
			continue
		}

		b.cells[y][index] += string(ch)
		return
	}
}

// Get gets a character at the specified position
func (b *Buffer) Get(x, y int) rune {
	if x < 0 || y < 0 || y >= b.height || x >= len(b.cells[y]) {
		return ' '
	}

	if b.cells[y][x] == undefinedCell {
		return ' '
	}

	if b.cells[y][x] == "" {
		return ' '
	}

	r, _ := utf8.DecodeRuneInString(b.cells[y][x])
	if r == utf8.RuneError {
		return ' '
	}

	return r
}

// Clear fills the buffer with spaces
func (b *Buffer) Clear() {
	for y := 0; y < b.height; y++ {
		b.cells[y] = make([]string, b.width)
		for x := range b.cells[y] {
			b.cells[y][x] = " "
		}
	}
}

// WriteString writes a string starting at the specified position
func (b *Buffer) WriteString(x, y int, s string) {
	currentX := x
	currentY := y

	for _, ch := range s {
		if ch == '\n' {
			currentY++
			currentX = x
			if currentY >= b.height {
				return
			}
			continue
		}

		if currentY < 0 || currentY >= b.height {
			return
		}

		if currentX >= 0 {
			b.Set(currentX, currentY, ch)
		}

		currentX += utils.RuneWidth(ch)
	}
}

// Render converts the buffer to a string for output
func (b *Buffer) Render() string {
	var sb strings.Builder
	lastNonEmptyLine := -1

	// Find the last non-empty line
	for y := 0; y < b.height; y++ {
		line := b.renderRow(y)
		if strings.TrimSpace(line) != "" {
			lastNonEmptyLine = y
		}
	}

	// If all lines are empty, return empty string
	if lastNonEmptyLine == -1 {
		return ""
	}

	// Render up to the last non-empty line
	for y := 0; y <= lastNonEmptyLine; y++ {
		line := b.renderRow(y)
		sb.WriteString(line)
		if y < lastNonEmptyLine {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// RenderRows renders exactly the requested number of rows, preserving empty lines.
func (b *Buffer) RenderRows(rows int) string {
	if rows <= 0 {
		return ""
	}

	if rows > b.height {
		rows = b.height
	}

	var sb strings.Builder
	for y := 0; y < rows; y++ {
		line := b.renderRow(y)
		sb.WriteString(line)
		if y < rows-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
