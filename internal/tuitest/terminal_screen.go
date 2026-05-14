package tuitest

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dh-kam/goink.go/pkg/utils"
)

// TerminalScreen projects a PTY byte stream into a plain terminal viewport.
// It intentionally keeps only text cells; style-aware projections can be added
// beside this without changing scenario semantics.
type TerminalScreen struct {
	width    int
	height   int
	cells    [][]string
	row      int
	col      int
	savedRow int
	savedCol int
}

func NewTerminalScreen(width int, height int) *TerminalScreen {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}

	screen := &TerminalScreen{
		width:  width,
		height: height,
		cells:  make([][]string, height),
	}
	for row := range screen.cells {
		screen.cells[row] = make([]string, width)
		screen.clearRow(row)
	}
	return screen
}

func (screen *TerminalScreen) Resize(width int, height int) {
	if screen == nil {
		return
	}
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	if width == screen.width && height == screen.height {
		return
	}

	cells := make([][]string, height)
	for row := range cells {
		cells[row] = make([]string, width)
		for col := range cells[row] {
			cells[row][col] = " "
		}
	}

	copyRows := minInt(height, screen.height)
	copyCols := minInt(width, screen.width)
	for row := 0; row < copyRows; row++ {
		copy(cells[row][:copyCols], screen.cells[row][:copyCols])
	}

	screen.width = width
	screen.height = height
	screen.cells = cells
	screen.clampCursor()
}

func (screen *TerminalScreen) Apply(text string) {
	for index := 0; index < len(text); {
		ch := text[index]
		switch ch {
		case 0x1b:
			index = screen.consumeEscape(text, index)
		case '\r':
			screen.col = 0
			index++
		case '\n':
			screen.lineFeed()
			index++
		case '\b':
			if screen.col > 0 {
				screen.col--
			}
			index++
		case '\t':
			next := ((screen.col / 8) + 1) * 8
			if next >= screen.width {
				next = screen.width - 1
			}
			screen.col = next
			index++
		default:
			if ch < 0x20 || ch == 0x7f {
				index++
				continue
			}

			r, size := utf8.DecodeRuneInString(text[index:])
			if r == utf8.RuneError && size == 1 {
				index++
				continue
			}
			screen.putRune(r)
			index += size
		}
	}
}

func (screen *TerminalScreen) PlainString() string {
	lines := make([]string, len(screen.cells))
	for row, cells := range screen.cells {
		var builder strings.Builder
		for _, cell := range cells {
			if cell == "" {
				continue
			}
			builder.WriteString(cell)
		}
		lines[row] = strings.TrimRight(builder.String(), " ")
	}

	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

func (screen *TerminalScreen) consumeEscape(text string, index int) int {
	if index+1 >= len(text) {
		return len(text)
	}

	switch text[index+1] {
	case '[':
		return screen.consumeCSI(text, index+2)
	case ']':
		return consumeOSC(text, index+2)
	case '7':
		screen.saveCursor()
		return index + 2
	case '8':
		screen.restoreCursor()
		return index + 2
	case 'c':
		screen.clearDisplay(2)
		screen.row = 0
		screen.col = 0
		return index + 2
	case '(', ')', '*', '+':
		if index+2 < len(text) {
			return index + 3
		}
		return len(text)
	default:
		return index + 2
	}
}

func (screen *TerminalScreen) consumeCSI(text string, start int) int {
	for index := start; index < len(text); index++ {
		ch := text[index]
		if ch >= 0x40 && ch <= 0x7e {
			screen.applyCSI(text[start:index], ch)
			return index + 1
		}
	}
	return len(text)
}

func consumeOSC(text string, start int) int {
	for index := start; index < len(text); index++ {
		switch text[index] {
		case 0x07:
			return index + 1
		case 0x1b:
			if index+1 < len(text) && text[index+1] == '\\' {
				return index + 2
			}
		}
	}
	return len(text)
}

func (screen *TerminalScreen) applyCSI(payload string, final byte) {
	private := strings.HasPrefix(payload, "?")
	params := parseCSIParams(payload)

	switch final {
	case 'A':
		screen.row -= csiParam(params, 0, 1)
	case 'B':
		screen.row += csiParam(params, 0, 1)
	case 'C':
		screen.col += csiParam(params, 0, 1)
	case 'D':
		screen.col -= csiParam(params, 0, 1)
	case 'E':
		screen.row += csiParam(params, 0, 1)
		screen.col = 0
	case 'F':
		screen.row -= csiParam(params, 0, 1)
		screen.col = 0
	case 'G':
		screen.col = csiParam(params, 0, 1) - 1
	case 'H', 'f':
		screen.row = csiParam(params, 0, 1) - 1
		screen.col = csiParam(params, 1, 1) - 1
	case 'J':
		screen.clearDisplay(csiParam(params, 0, 0))
	case 'K':
		screen.clearLine(csiParam(params, 0, 0))
	case 'P':
		screen.deleteChars(csiParam(params, 0, 1))
	case 'S':
		screen.scrollUp(csiParam(params, 0, 1))
	case 'T':
		screen.scrollDown(csiParam(params, 0, 1))
	case 'X':
		screen.eraseChars(csiParam(params, 0, 1))
	case 'd':
		screen.row = csiParam(params, 0, 1) - 1
	case 's':
		screen.saveCursor()
	case 'u':
		screen.restoreCursor()
	case 'h':
		if private && hasCSIParam(params, 1049) {
			screen.clearDisplay(2)
			screen.row = 0
			screen.col = 0
		}
	case 'l':
		if private && hasCSIParam(params, 1049) {
			screen.clearDisplay(2)
			screen.row = 0
			screen.col = 0
		}
	}

	screen.clampCursor()
}

func parseCSIParams(payload string) []int {
	payload = strings.TrimLeft(payload, "? >!=$")
	if payload == "" {
		return nil
	}

	parts := strings.FieldsFunc(payload, func(r rune) bool {
		return r == ';' || r == ':'
	})
	params := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			params = append(params, 0)
			continue
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			params = append(params, 0)
			continue
		}
		params = append(params, value)
	}
	return params
}

func csiParam(params []int, index int, fallback int) int {
	if index >= len(params) || params[index] == 0 {
		return fallback
	}
	return params[index]
}

func hasCSIParam(params []int, value int) bool {
	for _, param := range params {
		if param == value {
			return true
		}
	}
	return false
}

func (screen *TerminalScreen) putRune(r rune) {
	width := utils.RuneWidth(r)
	if width == 0 {
		screen.appendCombiningRune(r)
		return
	}
	if width < 0 {
		width = 1
	}
	if width > screen.width {
		width = 1
	}
	if width > 1 && screen.col == screen.width-1 {
		screen.lineFeed()
		screen.col = 0
	}
	if screen.col >= screen.width {
		screen.lineFeed()
		screen.col = 0
	}

	screen.cells[screen.row][screen.col] = string(r)
	for offset := 1; offset < width && screen.col+offset < screen.width; offset++ {
		screen.cells[screen.row][screen.col+offset] = ""
	}
	screen.col += width
	if screen.col >= screen.width {
		screen.lineFeed()
		screen.col = 0
	}
}

func (screen *TerminalScreen) appendCombiningRune(r rune) {
	for col := screen.col - 1; col >= 0; col-- {
		if screen.cells[screen.row][col] != "" {
			screen.cells[screen.row][col] += string(r)
			return
		}
	}
}

func (screen *TerminalScreen) lineFeed() {
	if screen.row == screen.height-1 {
		screen.scrollUp(1)
		return
	}
	screen.row++
}

func (screen *TerminalScreen) clearDisplay(mode int) {
	switch mode {
	case 1:
		for row := 0; row < screen.row; row++ {
			screen.clearRow(row)
		}
		for col := 0; col <= screen.col && col < screen.width; col++ {
			screen.cells[screen.row][col] = " "
		}
	case 2, 3:
		for row := 0; row < screen.height; row++ {
			screen.clearRow(row)
		}
	default:
		for col := screen.col; col < screen.width; col++ {
			screen.cells[screen.row][col] = " "
		}
		for row := screen.row + 1; row < screen.height; row++ {
			screen.clearRow(row)
		}
	}
}

func (screen *TerminalScreen) clearLine(mode int) {
	switch mode {
	case 1:
		for col := 0; col <= screen.col && col < screen.width; col++ {
			screen.cells[screen.row][col] = " "
		}
	case 2:
		screen.clearRow(screen.row)
	default:
		for col := screen.col; col < screen.width; col++ {
			screen.cells[screen.row][col] = " "
		}
	}
}

func (screen *TerminalScreen) eraseChars(count int) {
	if count < 1 {
		count = 1
	}
	for col := screen.col; col < screen.col+count && col < screen.width; col++ {
		screen.cells[screen.row][col] = " "
	}
}

func (screen *TerminalScreen) deleteChars(count int) {
	if count < 1 {
		count = 1
	}
	if count > screen.width-screen.col {
		count = screen.width - screen.col
	}
	row := screen.cells[screen.row]
	copy(row[screen.col:], row[screen.col+count:])
	for col := screen.width - count; col < screen.width; col++ {
		row[col] = " "
	}
}

func (screen *TerminalScreen) scrollUp(count int) {
	if count < 1 {
		count = 1
	}
	if count > screen.height {
		count = screen.height
	}
	copy(screen.cells, screen.cells[count:])
	for row := screen.height - count; row < screen.height; row++ {
		screen.cells[row] = make([]string, screen.width)
		screen.clearRow(row)
	}
}

func (screen *TerminalScreen) scrollDown(count int) {
	if count < 1 {
		count = 1
	}
	if count > screen.height {
		count = screen.height
	}
	copy(screen.cells[count:], screen.cells[:screen.height-count])
	for row := 0; row < count; row++ {
		screen.cells[row] = make([]string, screen.width)
		screen.clearRow(row)
	}
}

func (screen *TerminalScreen) clearRow(row int) {
	for col := 0; col < screen.width; col++ {
		screen.cells[row][col] = " "
	}
}

func (screen *TerminalScreen) saveCursor() {
	screen.savedRow = screen.row
	screen.savedCol = screen.col
}

func (screen *TerminalScreen) restoreCursor() {
	screen.row = screen.savedRow
	screen.col = screen.savedCol
	screen.clampCursor()
}

func (screen *TerminalScreen) clampCursor() {
	if screen.row < 0 {
		screen.row = 0
	}
	if screen.row >= screen.height {
		screen.row = screen.height - 1
	}
	if screen.col < 0 {
		screen.col = 0
	}
	if screen.col >= screen.width {
		screen.col = screen.width - 1
	}
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
