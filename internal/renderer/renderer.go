package renderer

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dh-kam/goink.go/internal/buffer"
	"github.com/dh-kam/goink.go/pkg/layout"
	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/utils"
	"github.com/dh-kam/goink.go/pkg/vdom"
	"github.com/rivo/uniseg"
)

type refSetter interface {
	SetCurrent(interface{})
}

// RenderSections separates dynamic output from newly rendered static blocks.
type RenderSections struct {
	Output            string
	StaticOutput      string
	StaticDeltaOutput string
	StaticCounts      []int
}

type clipRect struct {
	left   int
	top    int
	right  int
	bottom int
}

type ansiStyle struct {
	fg            string
	bg            string
	bold          bool
	dim           bool
	italic        bool
	underline     bool
	inverse       bool
	strikethrough bool
}

func roundLayoutValue(value float64) int {
	return int(math.Round(value))
}

type ansiCell struct {
	text         string
	style        ansiStyle
	continuation bool
	// prefix holds raw escape bytes emitted immediately after the SGR
	// transition for this cell and before its glyph. Used to carry OSC 8
	// hyperlink enter/exit sequences through the canvas write/diff pipeline
	// without disturbing column accounting.
	prefix string
}

type plainCell struct {
	text         string
	visible      string
	continuation bool
}

type styledRune struct {
	ch    rune
	style ansiStyle
	// prefix carries raw escape bytes (typically OSC sequences such as the
	// OSC 8 hyperlink enter/exit pair) that must be emitted verbatim
	// immediately before this rune. When ch == 0 the entry is a zero-width
	// sentinel that contributes no visible glyph and exists solely to carry
	// trailing escape bytes (e.g. the closing OSC after the last visible
	// rune).
	prefix string
}

type ansiCanvas struct {
	width  int
	height int
	cells  [][]ansiCell
}

type plainCanvas struct {
	width  int
	height int
	cells  [][]plainCell
}

func newANSICanvas(width, height int) *ansiCanvas {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	cells := make([][]ansiCell, height)
	for row := range cells {
		cells[row] = make([]ansiCell, width)
		for column := range cells[row] {
			cells[row][column].text = " "
		}
	}

	return &ansiCanvas{
		width:  width,
		height: height,
		cells:  cells,
	}
}

func newPlainCanvas(width, height int) *plainCanvas {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	cells := make([][]plainCell, height)
	for row := range cells {
		cells[row] = make([]plainCell, width)
		for column := range cells[row] {
			cells[row][column] = plainCell{
				text:    " ",
				visible: " ",
			}
		}
	}

	return &plainCanvas{
		width:  width,
		height: height,
		cells:  cells,
	}
}

func (c *plainCanvas) ensureRowWidth(y, width int) {
	if c == nil || y < 0 || y >= c.height || width <= len(c.cells[y]) {
		return
	}

	extra := make([]plainCell, width-len(c.cells[y]))
	for index := range extra {
		extra[index] = plainCell{
			text:    " ",
			visible: " ",
		}
	}

	c.cells[y] = append(c.cells[y], extra...)
}

func (c *plainCanvas) setCellText(x, y int, text string, visible string, width int) {
	if c == nil || x < 0 || y < 0 || y >= c.height || width <= 0 {
		return
	}

	c.ensureRowWidth(y, x+width)
	c.cells[y][x] = plainCell{
		text:    text,
		visible: visible,
	}

	for offset := 1; offset < width; offset++ {
		c.cells[y][x+offset] = plainCell{
			continuation: true,
		}
	}
}

func (c *plainCanvas) setCell(x, y int, ch rune) {
	if c == nil || x < 0 || y < 0 || y >= c.height {
		return
	}

	width := utils.RuneWidth(ch)
	if width == 0 {
		c.appendToPreviousVisible(x, y, string(ch))
		return
	}

	c.setCellText(x, y, string(ch), string(ch), width)
}

func (c *plainCanvas) appendToPreviousVisible(x, y int, suffix string) bool {
	if c == nil || suffix == "" || y < 0 || y >= c.height || x <= 0 {
		return false
	}

	for index := x - 1; index >= 0; index-- {
		if index >= len(c.cells[y]) {
			continue
		}

		cell := &c.cells[y][index]
		if cell.continuation || cell.text == "" {
			continue
		}

		cell.text += suffix
		return true
	}

	return false
}

func (c *plainCanvas) Render() string {
	if c == nil {
		return ""
	}

	lastNonEmptyLine := -1
	for row := 0; row < c.height; row++ {
		if renderPlainRow(c.cells[row]) != "" {
			lastNonEmptyLine = row
		}
	}

	if lastNonEmptyLine == -1 {
		return ""
	}

	var builder strings.Builder
	for row := 0; row <= lastNonEmptyLine; row++ {
		builder.WriteString(renderPlainRow(c.cells[row]))
		if row < lastNonEmptyLine {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func (c *plainCanvas) RenderRows(rows int) string {
	if c == nil || rows <= 0 {
		return ""
	}

	if rows > c.height {
		rows = c.height
	}

	var builder strings.Builder
	for row := 0; row < rows; row++ {
		builder.WriteString(renderPlainRow(c.cells[row]))
		if row < rows-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func renderPlainRow(cells []plainCell) string {
	lastVisible := -1
	for index, cell := range cells {
		if cell.continuation {
			continue
		}
		if cell.visible != " " {
			lastVisible = index
		}
	}

	if lastVisible == -1 {
		return ""
	}

	var builder strings.Builder
	for index := 0; index <= lastVisible; index++ {
		if cells[index].continuation {
			continue
		}
		builder.WriteString(cells[index].text)
	}

	return builder.String()
}

func (c *ansiCanvas) fillCell(x, y int, style ansiStyle) {
	if c == nil || x < 0 || x >= c.width || y < 0 || y >= c.height {
		return
	}

	c.cells[y][x] = ansiCell{
		text:  " ",
		style: style,
	}
}

func (c *ansiCanvas) setCell(x, y int, ch rune, style ansiStyle) {
	c.setCellWithPrefix(x, y, ch, style, "")
}

func (c *ansiCanvas) setCellWithPrefix(x, y int, ch rune, style ansiStyle, prefix string) {
	if c == nil || x < 0 || x >= c.width || y < 0 || y >= c.height {
		return
	}

	width := utils.RuneWidth(ch)
	if width == 0 {
		if prefix != "" {
			c.appendZeroWidthRaw(x, y, prefix)
		}
		if ch != 0 {
			c.appendZeroWidth(x, y, ch)
		}
		return
	}

	c.cells[y][x] = ansiCell{
		text:   string(ch),
		style:  style,
		prefix: prefix,
	}

	for offset := 1; offset < width && x+offset < c.width; offset++ {
		c.cells[y][x+offset] = ansiCell{
			style:        style,
			continuation: true,
		}
	}
}

func (c *ansiCanvas) appendZeroWidth(x, y int, ch rune) {
	if c == nil || y < 0 || y >= c.height || x <= 0 {
		return
	}

	for index := x - 1; index >= 0; index-- {
		cell := &c.cells[y][index]
		if cell.continuation {
			continue
		}
		if cell.text == "" {
			continue
		}

		cell.text += string(ch)
		return
	}
}

// appendZeroWidthRaw appends raw escape bytes (typically OSC sequences) to
// the trailing edge of the most recent visible cell on row y so they emit
// after that cell's glyph but before any subsequent SGR transition. This is
// the path used to attach a closing OSC to the last visible rune of a line
// without disturbing column accounting.
func (c *ansiCanvas) appendZeroWidthRaw(x, y int, raw string) {
	if c == nil || y < 0 || y >= c.height || raw == "" {
		return
	}

	for index := x - 1; index >= 0; index-- {
		cell := &c.cells[y][index]
		if cell.continuation {
			continue
		}
		if cell.text == "" {
			continue
		}

		cell.text += raw
		return
	}
}

func (c *ansiCanvas) RenderRows(rows int) string {
	if c == nil || rows <= 0 {
		return ""
	}

	if rows > c.height {
		rows = c.height
	}

	var builder strings.Builder
	for row := 0; row < rows; row++ {
		builder.WriteString(renderANSIRow(c.cells[row]))
		if row < rows-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func renderANSIRow(cells []ansiCell) string {
	lastVisible := -1
	for index, cell := range cells {
		if cell.continuation || cell.text != " " || cell.style.fg != "" || cell.style.bg != "" || cell.prefix != "" {
			lastVisible = index
		}
	}

	if lastVisible == -1 {
		return ""
	}

	var builder strings.Builder
	current := ansiStyle{}
	for index := 0; index <= lastVisible; index++ {
		cell := cells[index]
		if cell.continuation {
			continue
		}

		emitANSITransition(&builder, current, cell.style)
		current = cell.style

		if cell.prefix != "" {
			builder.WriteString(cell.prefix)
		}

		text := cell.text
		if text == "" {
			text = " "
		}
		builder.WriteString(text)
	}

	emitANSITransition(&builder, current, ansiStyle{})
	return builder.String()
}

func emitANSITransition(builder *strings.Builder, current ansiStyle, next ansiStyle) {
	if builder == nil {
		return
	}

	// Match Ink's ordering when a styled segment returns to the default
	// foreground while also switching into a dim/bold border style.
	if current.fg != "" && next.fg == "" && (next.bold || next.dim) {
		builder.WriteString("\x1b[39m")
		current.fg = ""
	}

	// Ink emits foreground color before enabling dim when both change together.
	if next.dim && !next.bold && next.fg != "" && next.fg != current.fg {
		builder.WriteString(next.fg)
		current.fg = next.fg
	}

	if current.bold != next.bold || current.dim != next.dim {
		if current.bold || current.dim {
			builder.WriteString("\x1b[22m")
		}
		if next.bold {
			builder.WriteString(styles.BoldCode())
		}
		if next.dim {
			builder.WriteString(styles.DimCode())
		}
	}
	if current.italic && !next.italic {
		builder.WriteString("\x1b[23m")
	}
	if current.underline && !next.underline {
		builder.WriteString("\x1b[24m")
	}
	if current.inverse && !next.inverse {
		builder.WriteString("\x1b[27m")
	}
	if current.strikethrough && !next.strikethrough {
		builder.WriteString("\x1b[29m")
	}
	if current.fg != "" && next.fg == "" {
		builder.WriteString("\x1b[39m")
	}
	if current.bg != "" && next.bg == "" {
		builder.WriteString("\x1b[49m")
	}
	if next.fg != "" && next.fg != current.fg {
		builder.WriteString(next.fg)
	}
	if next.bg != "" && next.bg != current.bg {
		builder.WriteString(next.bg)
	}
	if !current.italic && next.italic {
		builder.WriteString(styles.ItalicCode())
	}
	if !current.underline && next.underline {
		builder.WriteString(styles.UnderlineCode())
	}
	if !current.inverse && next.inverse {
		builder.WriteString(styles.InverseCode())
	}
	if !current.strikethrough && next.strikethrough {
		builder.WriteString(styles.StrikethroughCode())
	}
}

type borderGlyphs struct {
	topLeft     rune
	topRight    rune
	bottomLeft  rune
	bottomRight rune
	top         rune
	bottom      rune
	left        rune
	right       rune
}

var accessibilityStateOrder = []string{
	"busy",
	"checked",
	"disabled",
	"expanded",
	"readonly",
	"required",
	"selected",
}

// accessibilityShorthandKeys lists the top-level aria-* shorthand props that
// fold into the same state description as `aria-state`. Order here is
// canonical and matches the order in accessibilityStateOrder so the output
// stays deterministic when shorthand and explicit state combine.
var accessibilityShorthandKeys = []string{
	"aria-busy",
	"aria-checked",
	"aria-disabled",
	"aria-expanded",
	"aria-readonly",
	"aria-required",
	"aria-selected",
}

// accessibilityKnownStateSet exists for O(1) membership tests when ordering
// arbitrary state keys after the known set.
var accessibilityKnownStateSet = func() map[string]struct{} {
	known := make(map[string]struct{}, len(accessibilityStateOrder))
	for _, key := range accessibilityStateOrder {
		known[key] = struct{}{}
	}
	return known
}()

func isTextLikeNode(node *vdom.Node) bool {
	if node == nil || node.Type != vdom.ElementNode {
		return false
	}

	return node.ElementType == "text" || node.ElementType == "transform"
}

func isLayoutContainerNode(node *vdom.Node) bool {
	if node == nil || node.Type != vdom.ElementNode {
		return false
	}

	return node.ElementType == "box" || node.ElementType == "static"
}

func parseFlexDirection(value interface{}) (layout.FlexDirection, bool) {
	switch v := value.(type) {
	case layout.FlexDirection:
		return v, true
	case string:
		switch v {
		case "row", "row-reverse":
			return layout.FlexDirectionRow, true
		case "column", "column-reverse":
			return layout.FlexDirectionColumn, true
		}
	}

	return 0, false
}

func parseJustifyContent(value interface{}) (layout.JustifyContent, bool) {
	switch v := value.(type) {
	case layout.JustifyContent:
		return v, true
	case string:
		switch v {
		case "flex-start", "start":
			return layout.JustifyStart, true
		case "center":
			return layout.JustifyCenter, true
		case "flex-end", "end":
			return layout.JustifyEnd, true
		case "space-between":
			return layout.JustifySpaceBetween, true
		case "space-around":
			return layout.JustifySpaceAround, true
		case "space-evenly":
			return layout.JustifySpaceEvenly, true
		}
	}

	return 0, false
}

func parseAlignItems(value interface{}) (layout.AlignItems, bool) {
	switch v := value.(type) {
	case layout.AlignItems:
		return v, true
	case string:
		switch v {
		case "stretch":
			return layout.AlignStretch, true
		case "flex-start", "start":
			return layout.AlignStart, true
		case "center":
			return layout.AlignCenter, true
		case "flex-end", "end":
			return layout.AlignEnd, true
		}
	}

	return 0, false
}

func parseFlexBasis(value interface{}) (float64, bool, bool) {
	if numeric, ok := parseNumericValue(value); ok {
		return numeric, false, true
	}

	switch typed := value.(type) {
	case string:
		if strings.HasSuffix(typed, "%") {
			var percent float64
			if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%f%%", &percent); err == nil {
				return percent, true, true
			}
		}
	}

	return 0, false, false
}

func parseSizeValue(value interface{}) (float64, bool, bool) {
	if numeric, ok := parseNumericValue(value); ok {
		return numeric, false, true
	}

	switch typed := value.(type) {
	case string:
		if strings.HasSuffix(strings.TrimSpace(typed), "%") {
			var percent float64
			if _, err := fmt.Sscanf(strings.TrimSpace(typed), "%f%%", &percent); err == nil {
				return percent, true, true
			}
		}
	}

	return 0, false, false
}

func parseNumericValue(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func parseWrapMode(value interface{}) (layout.WrapMode, bool) {
	switch typed := value.(type) {
	case string:
		switch typed {
		case "", "nowrap", "no-wrap":
			return layout.WrapNoWrap, true
		case "wrap":
			return layout.WrapWrap, true
		case "wrap-reverse":
			return layout.WrapWrapReverse, true
		}
	}

	return 0, false
}

func hasBorderStyle(value interface{}) bool {
	if value == nil {
		return false
	}

	switch typed := value.(type) {
	case string:
		return typed != ""
	case map[string]interface{}:
		return len(typed) > 0
	case vdom.Props:
		return len(typed) > 0
	default:
		_, ok := borderStyleGlyphs(value)
		return ok
	}
}

func borderInsets(props vdom.Props) (left, top, right, bottom float64) {
	if props == nil || !hasBorderStyle(props["borderStyle"]) {
		return 0, 0, 0, 0
	}

	left, top, right, bottom = 1, 1, 1, 1
	if visible, ok := props["borderLeft"].(bool); ok && !visible {
		left = 0
	}
	if visible, ok := props["borderTop"].(bool); ok && !visible {
		top = 0
	}
	if visible, ok := props["borderRight"].(bool); ok && !visible {
		right = 0
	}
	if visible, ok := props["borderBottom"].(bool); ok && !visible {
		bottom = 0
	}

	return left, top, right, bottom
}

func borderRune(value interface{}) (rune, bool) {
	switch typed := value.(type) {
	case string:
		runes := []rune(typed)
		if len(runes) == 0 {
			return 0, false
		}

		return runes[0], true
	}

	return 0, false
}

func borderStyleGlyphs(value interface{}) (borderGlyphs, bool) {
	switch style := value.(type) {
	case string:
		switch style {
		case "round", "rounded":
			return borderGlyphs{
				topLeft:     '╭',
				topRight:    '╮',
				bottomLeft:  '╰',
				bottomRight: '╯',
				top:         '─',
				bottom:      '─',
				left:        '│',
				right:       '│',
			}, true
		case "single":
			return borderGlyphs{
				topLeft:     '┌',
				topRight:    '┐',
				bottomLeft:  '└',
				bottomRight: '┘',
				top:         '─',
				bottom:      '─',
				left:        '│',
				right:       '│',
			}, true
		case "double":
			return borderGlyphs{
				topLeft:     '╔',
				topRight:    '╗',
				bottomLeft:  '╚',
				bottomRight: '╝',
				top:         '═',
				bottom:      '═',
				left:        '║',
				right:       '║',
			}, true
		case "bold":
			return borderGlyphs{
				topLeft:     '┏',
				topRight:    '┓',
				bottomLeft:  '┗',
				bottomRight: '┛',
				top:         '━',
				bottom:      '━',
				left:        '┃',
				right:       '┃',
			}, true
		case "singleDouble":
			return borderGlyphs{
				topLeft:     '╓',
				topRight:    '╖',
				bottomLeft:  '╙',
				bottomRight: '╜',
				top:         '─',
				bottom:      '─',
				left:        '║',
				right:       '║',
			}, true
		case "doubleSingle":
			return borderGlyphs{
				topLeft:     '╒',
				topRight:    '╕',
				bottomLeft:  '╘',
				bottomRight: '╛',
				top:         '═',
				bottom:      '═',
				left:        '│',
				right:       '│',
			}, true
		case "classic":
			return borderGlyphs{
				topLeft:     '+',
				topRight:    '+',
				bottomLeft:  '+',
				bottomRight: '+',
				top:         '-',
				bottom:      '-',
				left:        '|',
				right:       '|',
			}, true
		case "arrow":
			return borderGlyphs{
				topLeft:     '↘',
				topRight:    '↙',
				bottomLeft:  '↗',
				bottomRight: '↖',
				top:         '↓',
				bottom:      '↑',
				left:        '→',
				right:       '←',
			}, true
		}
	case map[string]interface{}:
		topLeft, okTopLeft := borderRune(style["topLeft"])
		top, okTop := borderRune(style["top"])
		topRight, okTopRight := borderRune(style["topRight"])
		left, okLeft := borderRune(style["left"])
		bottomLeft, okBottomLeft := borderRune(style["bottomLeft"])
		bottom, okBottom := borderRune(style["bottom"])
		bottomRight, okBottomRight := borderRune(style["bottomRight"])
		right, okRight := borderRune(style["right"])
		if okTopLeft && okTop && okTopRight && okLeft && okBottomLeft && okBottom && okBottomRight && okRight {
			return borderGlyphs{
				topLeft:     topLeft,
				topRight:    topRight,
				bottomLeft:  bottomLeft,
				bottomRight: bottomRight,
				top:         top,
				bottom:      bottom,
				left:        left,
				right:       right,
			}, true
		}
	case vdom.Props:
		return borderStyleGlyphs(map[string]interface{}(style))
	}

	return borderGlyphs{}, false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func intersectClipRect(base clipRect, left, top, right, bottom int, clipX, clipY bool) clipRect {
	if clipX {
		base.left = maxInt(base.left, left)
		base.right = minInt(base.right, right)
	}

	if clipY {
		base.top = maxInt(base.top, top)
		base.bottom = minInt(base.bottom, bottom)
	}

	if base.right < base.left {
		base.right = base.left
	}

	if base.bottom < base.top {
		base.bottom = base.top
	}

	return base
}

func overflowAxes(node *vdom.Node) (clipX bool, clipY bool) {
	if node == nil || node.Type != vdom.ElementNode || node.Props == nil {
		return false, false
	}

	if overflow, ok := node.Props["overflow"].(string); ok && overflow == "hidden" {
		clipX = true
		clipY = true
	}

	if overflowX, ok := node.Props["overflowX"].(string); ok && overflowX == "hidden" {
		clipX = true
	}

	if overflowY, ok := node.Props["overflowY"].(string); ok && overflowY == "hidden" {
		clipY = true
	}

	return clipX, clipY
}

func shouldUseAvailableRootWidth(node *vdom.Node) bool {
	if node == nil || node.Type != vdom.ElementNode || node.ElementType != "box" || node.Props == nil {
		return false
	}

	if _, hasWidth := node.Props["width"]; hasWidth {
		return false
	}

	switch align := node.Props["alignSelf"].(type) {
	case string:
		return align != "flex-start" && align != "start"
	case layout.AlignItems:
		return align != layout.AlignStart
	default:
		return true
	}
}

func drawHorizontalBorder(buf *buffer.Buffer, x, y, width int, fill rune, leftVisible, rightVisible bool, leftCorner, rightCorner rune, clip clipRect) {
	if width <= 0 {
		return
	}

	contentWidth := width
	if leftVisible {
		contentWidth--
	}
	if rightVisible {
		contentWidth--
	}
	if contentWidth < 0 {
		contentWidth = 0
	}

	var line strings.Builder
	if leftVisible {
		line.WriteRune(leftCorner)
	}
	for i := 0; i < contentWidth; i++ {
		line.WriteRune(fill)
	}
	if rightVisible {
		line.WriteRune(rightCorner)
	}

	writeStringClipped(buf, x, y, line.String(), clip)
}

func drawHorizontalBorderPlain(canvas *plainCanvas, x, y, width int, fill rune, leftVisible, rightVisible bool, leftCorner, rightCorner rune, clip clipRect) {
	if canvas == nil || width <= 0 {
		return
	}

	contentWidth := width
	if leftVisible {
		contentWidth--
	}
	if rightVisible {
		contentWidth--
	}
	if contentWidth < 0 {
		contentWidth = 0
	}

	column := x
	if leftVisible {
		if y >= clip.top && y < clip.bottom && column >= clip.left && column < clip.right {
			canvas.setCell(column, y, leftCorner)
		}
		column++
	}

	for index := 0; index < contentWidth; index++ {
		if y >= clip.top && y < clip.bottom && column >= clip.left && column < clip.right {
			canvas.setCell(column, y, fill)
		}
		column++
	}

	if rightVisible && y >= clip.top && y < clip.bottom && column >= clip.left && column < clip.right {
		canvas.setCell(column, y, rightCorner)
	}
}

func drawBoxBorder(buf *buffer.Buffer, vnode *vdom.Node, layoutNode *layout.Node, clip clipRect) {
	if vnode == nil || layoutNode == nil || vnode.Type != vdom.ElementNode || vnode.Props == nil {
		return
	}

	glyphs, ok := borderStyleGlyphs(vnode.Props["borderStyle"])
	if !ok {
		return
	}

	leftInset, topInset, rightInset, bottomInset := borderInsets(vnode.Props)
	showLeft := leftInset > 0
	showTop := topInset > 0
	showRight := rightInset > 0
	showBottom := bottomInset > 0

	width := roundLayoutValue(layoutNode.GetComputedWidth())
	height := roundLayoutValue(layoutNode.GetComputedHeight())
	if width <= 0 || height <= 0 {
		return
	}

	x := roundLayoutValue(layoutNode.GetComputedLeft())
	y := roundLayoutValue(layoutNode.GetComputedTop())

	if showTop {
		drawHorizontalBorder(buf, x, y, width, glyphs.top, showLeft, showRight, glyphs.topLeft, glyphs.topRight, clip)
	}
	if showBottom {
		drawHorizontalBorder(buf, x, y+height-1, width, glyphs.bottom, showLeft, showRight, glyphs.bottomLeft, glyphs.bottomRight, clip)
	}

	verticalStart := y
	if showTop {
		verticalStart++
	}
	verticalEnd := y + height
	if showBottom {
		verticalEnd--
	}

	for row := verticalStart; row < verticalEnd; row++ {
		if showLeft {
			writeStringClipped(buf, x, row, string(glyphs.left), clip)
		}
		if showRight {
			writeStringClipped(buf, x+width-1, row, string(glyphs.right), clip)
		}
	}
}

func drawBoxBorderPlain(canvas *plainCanvas, vnode *vdom.Node, layoutNode *layout.Node, clip clipRect) {
	if canvas == nil || vnode == nil || layoutNode == nil || vnode.Type != vdom.ElementNode || vnode.Props == nil {
		return
	}

	glyphs, ok := borderStyleGlyphs(vnode.Props["borderStyle"])
	if !ok {
		return
	}

	leftInset, topInset, rightInset, bottomInset := borderInsets(vnode.Props)
	showLeft := leftInset > 0
	showTop := topInset > 0
	showRight := rightInset > 0
	showBottom := bottomInset > 0

	width := roundLayoutValue(layoutNode.GetComputedWidth())
	height := roundLayoutValue(layoutNode.GetComputedHeight())
	if width <= 0 || height <= 0 {
		return
	}

	x := roundLayoutValue(layoutNode.GetComputedLeft())
	y := roundLayoutValue(layoutNode.GetComputedTop())

	if showTop {
		drawHorizontalBorderPlain(canvas, x, y, width, glyphs.top, showLeft, showRight, glyphs.topLeft, glyphs.topRight, clip)
	}
	if showBottom {
		drawHorizontalBorderPlain(canvas, x, y+height-1, width, glyphs.bottom, showLeft, showRight, glyphs.bottomLeft, glyphs.bottomRight, clip)
	}

	verticalStart := y
	if showTop {
		verticalStart++
	}
	verticalEnd := y + height
	if showBottom {
		verticalEnd--
	}

	for row := verticalStart; row < verticalEnd; row++ {
		if showLeft {
			writePlainStringClipped(canvas, x, row, string(glyphs.left), clip)
		}
		if showRight {
			writePlainStringClipped(canvas, x+width-1, row, string(glyphs.right), clip)
		}
	}
}

func drawHorizontalBorderANSI(canvas *ansiCanvas, x, y, width int, fill rune, leftVisible, rightVisible bool, leftCorner, rightCorner rune, style ansiStyle, clip clipRect) {
	if width <= 0 {
		return
	}

	contentWidth := width
	if leftVisible {
		contentWidth--
	}
	if rightVisible {
		contentWidth--
	}
	if contentWidth < 0 {
		contentWidth = 0
	}

	column := x
	if leftVisible {
		if y >= clip.top && y < clip.bottom && column >= clip.left && column < clip.right {
			canvas.setCell(column, y, leftCorner, style)
		}
		column++
	}

	for index := 0; index < contentWidth; index++ {
		if y >= clip.top && y < clip.bottom && column >= clip.left && column < clip.right {
			canvas.setCell(column, y, fill, style)
		}
		column++
	}

	if rightVisible && y >= clip.top && y < clip.bottom && column >= clip.left && column < clip.right {
		canvas.setCell(column, y, rightCorner, style)
	}
}

func drawBoxBorderANSI(canvas *ansiCanvas, vnode *vdom.Node, layoutNode *layout.Node, clip clipRect) {
	if vnode == nil || layoutNode == nil || vnode.Type != vdom.ElementNode || vnode.Props == nil {
		return
	}

	glyphs, ok := borderStyleGlyphs(vnode.Props["borderStyle"])
	if !ok {
		return
	}

	leftInset, topInset, rightInset, bottomInset := borderInsets(vnode.Props)
	showLeft := leftInset > 0
	showTop := topInset > 0
	showRight := rightInset > 0
	showBottom := bottomInset > 0

	width := roundLayoutValue(layoutNode.GetComputedWidth())
	height := roundLayoutValue(layoutNode.GetComputedHeight())
	if width <= 0 || height <= 0 {
		return
	}

	topStyle := resolveBorderSideStyle(vnode.Props, "borderTopColor", "borderTopDimColor")
	bottomStyle := resolveBorderSideStyle(vnode.Props, "borderBottomColor", "borderBottomDimColor")
	leftStyle := resolveBorderSideStyle(vnode.Props, "borderLeftColor", "borderLeftDimColor")
	rightStyle := resolveBorderSideStyle(vnode.Props, "borderRightColor", "borderRightDimColor")
	x := roundLayoutValue(layoutNode.GetComputedLeft())
	y := roundLayoutValue(layoutNode.GetComputedTop())

	if showTop {
		drawHorizontalBorderANSI(canvas, x, y, width, glyphs.top, showLeft, showRight, glyphs.topLeft, glyphs.topRight, topStyle, clip)
	}
	if showBottom {
		drawHorizontalBorderANSI(canvas, x, y+height-1, width, glyphs.bottom, showLeft, showRight, glyphs.bottomLeft, glyphs.bottomRight, bottomStyle, clip)
	}

	verticalStart := y
	if showTop {
		verticalStart++
	}
	verticalEnd := y + height
	if showBottom {
		verticalEnd--
	}

	for row := verticalStart; row < verticalEnd; row++ {
		if showLeft && row >= clip.top && row < clip.bottom && x >= clip.left && x < clip.right {
			canvas.setCell(x, row, glyphs.left, leftStyle)
		}
		if showRight && row >= clip.top && row < clip.bottom && x+width-1 >= clip.left && x+width-1 < clip.right {
			canvas.setCell(x+width-1, row, glyphs.right, rightStyle)
		}
	}
}

func fillBoxBackgroundANSI(canvas *ansiCanvas, vnode *vdom.Node, layoutNode *layout.Node, inheritedBackground string, clip clipRect) {
	if canvas == nil || vnode == nil || layoutNode == nil {
		return
	}

	style := resolveBoxBackgroundStyle(vnode, inheritedBackground)
	if style.bg == "" {
		return
	}

	leftInset, topInset, rightInset, bottomInset := borderInsets(vnode.Props)
	x := roundLayoutValue(layoutNode.GetComputedLeft()) + int(leftInset)
	y := roundLayoutValue(layoutNode.GetComputedTop()) + int(topInset)
	width := roundLayoutValue(layoutNode.GetComputedWidth()) - int(leftInset) - int(rightInset)
	height := roundLayoutValue(layoutNode.GetComputedHeight()) - int(topInset) - int(bottomInset)
	if width <= 0 || height <= 0 {
		return
	}

	for row := 0; row < height; row++ {
		targetY := y + row
		if targetY < clip.top || targetY >= clip.bottom {
			continue
		}

		for column := 0; column < width; column++ {
			targetX := x + column
			if targetX < clip.left || targetX >= clip.right {
				continue
			}

			canvas.fillCell(targetX, targetY, style)
		}
	}
}

func nodeClipRect(node *vdom.Node, layoutNode *layout.Node, parentClip clipRect) clipRect {
	if node == nil || layoutNode == nil {
		return parentClip
	}

	clipX, clipY := overflowAxes(node)
	if !clipX && !clipY {
		return parentClip
	}

	left := roundLayoutValue(layoutNode.GetComputedLeft())
	top := roundLayoutValue(layoutNode.GetComputedTop())
	right := left + roundLayoutValue(layoutNode.GetComputedWidth())
	bottom := top + roundLayoutValue(layoutNode.GetComputedHeight())

	return intersectClipRect(parentClip, left, top, right, bottom, clipX, clipY)
}

func nodeChildClipRect(node *vdom.Node, layoutNode *layout.Node, parentClip clipRect) clipRect {
	if node == nil || layoutNode == nil {
		return parentClip
	}

	clipX, clipY := overflowAxes(node)
	if !clipX && !clipY {
		return parentClip
	}

	leftInset, topInset, rightInset, bottomInset := borderInsets(node.Props)
	left := roundLayoutValue(layoutNode.GetComputedLeft()) + int(leftInset)
	top := roundLayoutValue(layoutNode.GetComputedTop()) + int(topInset)
	right := roundLayoutValue(layoutNode.GetComputedLeft()+layoutNode.GetComputedWidth()) - int(rightInset)
	bottom := roundLayoutValue(layoutNode.GetComputedTop()+layoutNode.GetComputedHeight()) - int(bottomInset)

	if right < left {
		right = left
	}
	if bottom < top {
		bottom = top
	}

	return intersectClipRect(parentClip, left, top, right, bottom, clipX, clipY)
}

func consumeANSISequence(text string, start int) (string, int, bool) {
	if start < 0 || start >= len(text) {
		return "", start, false
	}

	switch text[start] {
	case 0x1b:
		if start+1 >= len(text) {
			return text[start : start+1], start + 1, true
		}

		switch text[start+1] {
		case '[':
			index := start + 2
			for index < len(text) {
				ch := text[index]
				if ch >= 0x40 && ch <= 0x7e {
					return text[start : index+1], index + 1, true
				}
				index++
			}
			return text[start:], len(text), true
		case ']':
			index := start + 2
			for index < len(text) {
				switch text[index] {
				case 0x07, 0x9c:
					return text[start : index+1], index + 1, true
				case 0x1b:
					if index+1 < len(text) && text[index+1] == '\\' {
						return text[start : index+2], index + 2, true
					}
				}
				index++
			}
			return text[start:], len(text), true
		default:
			return text[start : start+1], start + 1, true
		}
	case 0x9b:
		index := start + 1
		for index < len(text) {
			ch := text[index]
			if ch >= 0x40 && ch <= 0x7e {
				return text[start : index+1], index + 1, true
			}
			index++
		}
		return text[start:], len(text), true
	case 0x9d:
		// 8-bit OSC introducer. Terminates on BEL (0x07), 8-bit ST (0x9c),
		// or 7-bit ST (ESC \).
		index := start + 1
		for index < len(text) {
			switch text[index] {
			case 0x07, 0x9c:
				return text[start : index+1], index + 1, true
			case 0x1b:
				if index+1 < len(text) && text[index+1] == '\\' {
					return text[start : index+2], index + 2, true
				}
			}
			index++
		}
		return text[start:], len(text), true
	default:
		return "", start, false
	}
}

func visibleStringWidth(text string) int {
	width := 0
	for index := 0; index < len(text); {
		if _, next, ok := consumeANSISequence(text, index); ok {
			index = next
			continue
		}

		r, size := utf8.DecodeRuneInString(text[index:])
		if r == utf8.RuneError && size == 0 {
			break
		}

		if r != '\n' {
			width += utils.RuneWidth(r)
		}
		index += size
	}

	return width
}

func writeStringClipped(buf *buffer.Buffer, x, y int, s string, clip clipRect) {
	currentX := x
	currentY := y
	pendingPrefix := ""

	for index := 0; index < len(s); {
		if sequence, next, ok := consumeANSISequence(s, index); ok {
			if !buf.AppendToPreviousVisible(currentX, currentY, sequence) {
				pendingPrefix += sequence
			}
			index = next
			continue
		}

		cluster, next := nextGraphemeClusterInString(s, index)
		if cluster == "" {
			return
		}

		if cluster == "\n" {
			currentY++
			currentX = x
			pendingPrefix = ""
			if currentY >= clip.bottom {
				return
			}

			index = next
			continue
		}

		width := utils.StringWidth(cluster)
		if width == 0 {
			if !buf.AppendToPreviousVisible(currentX, currentY, cluster) {
				pendingPrefix += cluster
			}
			index = next
			continue
		}

		if currentY >= clip.top && currentY < clip.bottom && currentX < clip.right && currentX+width > clip.left {
			if currentX >= clip.left {
				buf.SetString(currentX, currentY, pendingPrefix+cluster, width)
				pendingPrefix = ""
			}
		}

		currentX += width
		index = next
	}
}

func nextGraphemeClusterInString(text string, start int) (string, int) {
	if start < 0 || start >= len(text) {
		return "", start
	}

	graphemes := uniseg.NewGraphemes(text[start:])
	if graphemes.Next() {
		cluster := graphemes.Str()
		return cluster, start + len(cluster)
	}

	r, size := utf8.DecodeRuneInString(text[start:])
	if r == utf8.RuneError && size == 0 {
		return "", start
	}

	return text[start : start+size], start + size
}

func writePlainStringClipped(canvas *plainCanvas, x, y int, s string, clip clipRect) {
	currentX := x
	currentY := y
	pendingPrefix := ""

	for index := 0; index < len(s); {
		if sequence, next, ok := consumeANSISequence(s, index); ok {
			if !canvas.appendToPreviousVisible(currentX, currentY, sequence) {
				pendingPrefix += sequence
			}
			index = next
			continue
		}

		ch, size := utf8.DecodeRuneInString(s[index:])
		if ch == utf8.RuneError && size == 0 {
			return
		}

		if ch == '\n' {
			currentY++
			currentX = x
			pendingPrefix = ""
			if currentY >= clip.bottom {
				return
			}

			index += size
			continue
		}

		width := utils.RuneWidth(ch)
		if width == 0 {
			if !canvas.appendToPreviousVisible(currentX, currentY, string(ch)) {
				pendingPrefix += string(ch)
			}
			index += size
			continue
		}

		if currentY >= clip.top && currentY < clip.bottom && currentX < clip.right && currentX+width > clip.left {
			if currentX >= clip.left {
				canvas.setCellText(currentX, currentY, pendingPrefix+string(ch), string(ch), width)
				pendingPrefix = ""
			}
		}

		currentX += width
		index += size
	}
}

func writeStyledStringClipped(canvas *ansiCanvas, x, y int, s string, style ansiStyle, clip clipRect) {
	currentX := x
	currentY := y

	for _, ch := range s {
		if ch == '\n' {
			currentY++
			currentX = x
			if currentY >= clip.bottom {
				return
			}

			continue
		}

		width := utils.RuneWidth(ch)
		if width == 0 {
			if currentY >= clip.top && currentY < clip.bottom && currentX > clip.left && currentX <= clip.right {
				canvas.setCell(currentX, currentY, ch, style)
			}
			continue
		}

		if currentY >= clip.top && currentY < clip.bottom && currentX < clip.right && currentX+width > clip.left {
			if currentX >= clip.left {
				canvas.setCell(currentX, currentY, ch, style)
			}
		}

		currentX += width
	}
}

func applyNodeTransform(node *vdom.Node, text string, index int) string {
	if node == nil || node.Props == nil || text == "" {
		return text
	}

	transform, ok := node.Props["transform"].(func(string, int) string)
	if !ok || transform == nil {
		return text
	}

	if cached, hit := node.LookupTransformCache(text, index); hit {
		return cached
	}

	output := transform(text, index)
	node.StoreTransformCache(text, index, output)
	return output
}

func applyLineTransform(node *vdom.Node, text string) string {
	if node == nil || node.Props == nil || text == "" {
		return text
	}

	transform, ok := node.Props["transform"].(func(string, int) string)
	if !ok || transform == nil {
		return text
	}

	// Cache keyed on the joined input lets the second pass (render after
	// measure) skip the per-line transform invocation. Sentinel index -1
	// distinguishes the line-fan-out cache from the per-child node-transform
	// cache so they cannot collide on a node that's involved in both paths.
	if cached, hit := node.LookupTransformCache(text, -1); hit {
		return cached
	}

	lines := strings.Split(text, "\n")
	for lineIndex, line := range lines {
		lines[lineIndex] = transform(line, lineIndex)
	}

	output := strings.Join(lines, "\n")
	node.StoreTransformCache(text, -1, output)
	return output
}

func collectMeasuredTextContent(node *vdom.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case vdom.TextNode:
		return node.Text
	case vdom.ElementNode:
		if node.ElementType == "text" || node.ElementType == "transform" {
			text := ""
			for childIndex, child := range node.Children {
				if child == nil {
					continue
				}

				switch {
				case isTextLikeNode(child):
					text += measureNestedTextLikeNode(child, childIndex)
				case child.Type == vdom.TextNode:
					text += child.Text
				default:
					// Text-like containers only accept text content.
				}
			}

			return text
		}

		text := ""
		for childIndex, child := range node.Children {
			if child == nil {
				continue
			}

			if isTextLikeNode(child) {
				text += measureNestedTextLikeNode(child, childIndex)
				continue
			}

			text += collectMeasuredTextContent(child)
		}
		return text
	default:
		return ""
	}
}

func measureNestedTextLikeNode(node *vdom.Node, childIndex int) string {
	if node == nil {
		return ""
	}

	text := collectMeasuredTextContent(node)
	return applyNodeTransform(node, text, childIndex)
}

func textWrapMode(node *vdom.Node) string {
	if node == nil || node.Props == nil {
		return "wrap"
	}

	if wrap, ok := node.Props["wrap"].(string); ok && wrap != "" {
		return wrap
	}

	if wrap, ok := node.Props["textWrap"].(string); ok && wrap != "" {
		return wrap
	}

	return "wrap"
}

func fitRunesToWidth(runes []rune, maxWidth int) int {
	if maxWidth <= 0 {
		return 0
	}

	width := 0
	for index, r := range runes {
		runeWidth := utils.RuneWidth(r)
		if width+runeWidth > maxWidth {
			if index == 0 {
				return 1
			}
			return index
		}

		width += runeWidth
	}

	return len(runes)
}

// fitClustersByWidth returns the number of grapheme clusters from the
// front of `clusters` whose cumulative display width fits in maxWidth.
// Unlike fitRunesToWidth this never splits a single cluster (e.g. a ZWJ
// emoji sequence) across the boundary — entire clusters either fit or are
// excluded. If the very first cluster is wider than maxWidth, it is
// excluded (returning 0) so callers can decide how to handle that case.
func fitClustersByWidth(clusters []string, maxWidth int) int {
	if maxWidth <= 0 {
		return 0
	}

	width := 0
	for index, cluster := range clusters {
		clusterWidth := utils.StringWidth(cluster)
		if width+clusterWidth > maxWidth {
			return index
		}
		width += clusterWidth
	}

	return len(clusters)
}

// joinClusterRange joins a slice of grapheme clusters back into a string.
func joinClusterRange(clusters []string) string {
	var sb strings.Builder
	for _, c := range clusters {
		sb.WriteString(c)
	}
	return sb.String()
}

// firstClusterRune returns the first decoded rune of a grapheme cluster.
// Used by wrap-mode helpers that need to test whether a cluster begins
// with a whitespace character (cluster-level alternative to scanning the
// raw rune slice).
func firstClusterRune(cluster string) rune {
	r, _ := utf8.DecodeRuneInString(cluster)
	return r
}

func trimLeftSpaceRunes(runes []rune) []rune {
	index := 0
	for index < len(runes) && unicode.IsSpace(runes[index]) {
		index++
	}

	return runes[index:]
}

func trimRightSpaceRunes(runes []rune) []rune {
	end := len(runes)
	for end > 0 && unicode.IsSpace(runes[end-1]) {
		end--
	}

	return runes[:end]
}

func truncateEnd(line string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if utils.StringWidth(line) <= maxWidth {
		return line
	}
	if maxWidth == 1 {
		return "…"
	}

	// Cluster-based fit ensures wide CJK or ZWJ emoji clusters at the
	// boundary are dropped wholesale rather than split across the cut —
	// for "abc我" with maxWidth=3 we emit "ab…" (width 3) instead of
	// "ab我" (width 4) which would overflow the box.
	clusters := utils.GraphemeClusters(line)
	keep := fitClustersByWidth(clusters, maxWidth-1)
	return joinClusterRange(clusters[:keep]) + "…"
}

func truncateStart(line string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if utils.StringWidth(line) <= maxWidth {
		return line
	}
	if maxWidth == 1 {
		return "…"
	}

	clusters := utils.GraphemeClusters(line)
	keepWidth := maxWidth - 1
	start := len(clusters)
	width := 0
	for start > 0 {
		clusterWidth := utils.StringWidth(clusters[start-1])
		if width+clusterWidth > keepWidth {
			break
		}
		start--
		width += clusterWidth
	}

	return "…" + joinClusterRange(clusters[start:])
}

func truncateMiddle(line string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if utils.StringWidth(line) <= maxWidth {
		return line
	}
	if maxWidth == 1 {
		return "…"
	}

	remainingWidth := maxWidth - 1
	leftWidth := (remainingWidth + 1) / 2
	rightWidth := remainingWidth / 2

	clusters := utils.GraphemeClusters(line)
	leftEnd := fitClustersByWidth(clusters, leftWidth)
	left := joinClusterRange(clusters[:leftEnd])

	start := len(clusters)
	width := 0
	for start > 0 {
		clusterWidth := utils.StringWidth(clusters[start-1])
		if width+clusterWidth > rightWidth {
			break
		}
		start--
		width += clusterWidth
	}

	return left + "…" + joinClusterRange(clusters[start:])
}

func wrapLine(line string, maxWidth int) []string {
	if maxWidth <= 0 || utils.StringWidth(line) <= maxWidth {
		return []string{line}
	}

	clusters := utils.GraphemeClusters(line)
	lines := make([]string, 0, 2)
	for len(clusters) > 0 {
		if utils.StringWidth(joinClusterRange(clusters)) <= maxWidth {
			lines = append(lines, joinClusterRange(clusters))
			break
		}

		if leadingSpaces := leadingSpaceClusterCount(clusters); leadingSpaces > 0 && leadingSpaces < len(clusters) {
			if clusterRangeWidth(clusters[:leadingSpaces])+nextNonSpaceClusterRunWidth(clusters[leadingSpaces:]) > maxWidth {
				if len(lines) == 0 {
					lines = append(lines, "")
				}
				clusters = clusters[leadingSpaces:]
				continue
			}
		}

		fit := fitClustersByWidth(clusters, maxWidth)
		// fit may be 0 when the leading cluster on its own already
		// exceeds maxWidth (e.g. a width-2 ZWJ emoji at maxWidth=1) —
		// emit it on a line of its own to make forward progress and
		// avoid an infinite loop. The emitted line will visibly
		// overflow, but that mirrors upstream behaviour for clusters
		// wider than the available space.
		if fit == 0 {
			fit = 1
		}

		if fit < len(clusters) && unicode.IsSpace(firstClusterRune(clusters[fit])) {
			segment := joinClusterRange(clusters[:fit])
			lines = append(lines, segment)
			clusters = clusters[fit:]
			continue
		}

		breakIndex := -1
		seenNonSpace := false
		for index := 0; index < fit; index++ {
			r := firstClusterRune(clusters[index])
			if !unicode.IsSpace(r) {
				seenNonSpace = true
			}
			if seenNonSpace && unicode.IsSpace(r) {
				breakIndex = index
			}
		}

		if breakIndex > 0 {
			segment := joinClusterRange(clusters[:breakIndex+1])
			clusters = clusters[breakIndex+1:]
			if segment != "" {
				lines = append(lines, segment)
				continue
			}
		}

		segment := joinClusterRange(clusters[:fit])
		lines = append(lines, segment)
		clusters = clusters[fit:]
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

func leadingSpaceClusterCount(clusters []string) int {
	count := 0
	for count < len(clusters) && unicode.IsSpace(firstClusterRune(clusters[count])) {
		count++
	}
	return count
}

func clusterRangeWidth(clusters []string) int {
	width := 0
	for _, cluster := range clusters {
		width += utils.StringWidth(cluster)
	}
	return width
}

func nextNonSpaceClusterRunWidth(clusters []string) int {
	width := 0
	for _, cluster := range clusters {
		if unicode.IsSpace(firstClusterRune(cluster)) {
			break
		}
		width += utils.StringWidth(cluster)
	}
	return width
}

func trimLeftSpaceClusters(clusters []string) []string {
	index := 0
	for index < len(clusters) && unicode.IsSpace(firstClusterRune(clusters[index])) {
		index++
	}
	return clusters[index:]
}

func applyTextLayoutMode(text string, mode string, maxWidth int) string {
	if text == "" || maxWidth <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	processed := make([]string, 0, len(lines))
	for _, line := range lines {
		switch mode {
		case "truncate", "truncate-end":
			processed = append(processed, truncateEnd(line, maxWidth))
		case "truncate-middle":
			processed = append(processed, truncateMiddle(line, maxWidth))
		case "truncate-start":
			processed = append(processed, truncateStart(line, maxWidth))
		case "wrap", "":
			processed = append(processed, wrapLine(line, maxWidth)...)
		default:
			processed = append(processed, line)
		}
	}

	return strings.Join(processed, "\n")
}

func measureTextLikeNode(node *vdom.Node, maxWidth int) string {
	if node == nil {
		return ""
	}

	text := collectMeasuredTextContent(node)
	text = applyTextLayoutMode(text, textWrapMode(node), maxWidth)
	return applyLineTransform(node, text)
}

func textMeasurementWidth(width float64) int {
	if width <= 0 {
		return 0
	}

	if width < 1 {
		return 0
	}

	return int(math.Floor(width + 1e-9))
}

func parseBackgroundANSI(spec string) string {
	if spec == "" {
		return ""
	}

	color, ok := styles.ParseColor(spec)
	if !ok {
		return ""
	}

	return color.ToANSI(styles.Background)
}

func mergeTextANSIStyle(node *vdom.Node, inherited ansiStyle, inheritedBackground string) (ansiStyle, string) {
	style := inherited
	backgroundSpec := inheritedBackground
	if node == nil || node.Props == nil {
		return style, backgroundSpec
	}

	if colorSpec, ok := node.Props["color"].(string); ok {
		if color, ok := styles.ParseColor(colorSpec); ok {
			style.fg = color.ToANSI(styles.Foreground)
		}
	}

	if explicitBackground, ok := node.Props["backgroundColor"].(string); ok {
		backgroundSpec = explicitBackground
		style.bg = parseBackgroundANSI(explicitBackground)
	}

	if dimColor, _ := node.Props["dimColor"].(bool); dimColor {
		style.dim = true
	}
	if bold, _ := node.Props["bold"].(bool); bold {
		style.bold = true
	}
	if italic, _ := node.Props["italic"].(bool); italic {
		style.italic = true
	}
	if underline, _ := node.Props["underline"].(bool); underline {
		style.underline = true
	}
	if inverse, _ := node.Props["inverse"].(bool); inverse {
		style.inverse = true
	}
	if strikethrough, _ := node.Props["strikethrough"].(bool); strikethrough {
		style.strikethrough = true
	}

	return style, backgroundSpec
}

func styledPlainString(runes []styledRune) string {
	if len(runes) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, r := range runes {
		if r.prefix != "" {
			builder.WriteString(r.prefix)
		}
		if r.ch != 0 {
			builder.WriteRune(r.ch)
		}
	}

	return builder.String()
}

func styledRunesToANSIString(runes []styledRune, baseStyle ansiStyle) string {
	if len(runes) == 0 {
		return ""
	}

	var builder strings.Builder
	current := baseStyle
	for _, r := range runes {
		if r.ch == '\n' {
			if r.prefix != "" {
				builder.WriteString(r.prefix)
			}
			builder.WriteRune('\n')
			continue
		}

		emitANSITransition(&builder, current, r.style)
		current = r.style
		if r.prefix != "" {
			builder.WriteString(r.prefix)
		}
		if r.ch != 0 {
			builder.WriteRune(r.ch)
		}
	}

	emitANSITransition(&builder, current, baseStyle)
	return builder.String()
}

func reverseRunesCopy(runes []rune) []rune {
	reversed := make([]rune, len(runes))
	for index := range runes {
		reversed[len(runes)-1-index] = runes[index]
	}

	return reversed
}

func reverseStyledRunesCopy(runes []styledRune) []styledRune {
	reversed := make([]styledRune, len(runes))
	for index := range runes {
		reversed[len(runes)-1-index] = runes[index]
	}

	return reversed
}

func equalRunes(a []rune, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}

	return true
}

func indexOfRuneSubslice(haystack []rune, needle []rune) int {
	if len(needle) == 0 {
		return 0
	}
	if len(needle) > len(haystack) {
		return -1
	}

	for start := 0; start <= len(haystack)-len(needle); start++ {
		matched := true
		for offset := range needle {
			if haystack[start+offset] != needle[offset] {
				matched = false
				break
			}
		}

		if matched {
			return start
		}
	}

	return -1
}

func styledRunesFromString(text string, style ansiStyle) []styledRune {
	runes := []rune(text)
	result := make([]styledRune, 0, len(runes))
	for _, r := range runes {
		result = append(result, styledRune{ch: r, style: style})
	}

	return result
}

func resetANSIStyleToBase(style *ansiStyle, base ansiStyle) {
	if style == nil {
		return
	}

	*style = base
}

func parseANSIToStyledRunes(text string, baseStyle ansiStyle) []styledRune {
	current := baseStyle
	result := make([]styledRune, 0, len(text))
	pendingPrefix := ""

	flushPrefix := func() {
		if pendingPrefix == "" {
			return
		}
		// No further visible rune was emitted; preserve the trailing OSC
		// (or other consumed escape) on a zero-width sentinel so it survives
		// downstream serialization.
		result = append(result, styledRune{ch: 0, style: current, prefix: pendingPrefix})
		pendingPrefix = ""
	}

	for index := 0; index < len(text); {
		// OSC (Operating System Command): ESC ] ... (BEL | ESC \) or 8-bit
		// 0x9d ... terminator. Consume the entire sequence and attach it as
		// a prefix on the next visible rune so the OSC 8 hyperlink pair
		// survives wrap/truncate/clip transformations.
		if text[index] == 0x1b && index+1 < len(text) && text[index+1] == ']' {
			seq, next, ok := consumeANSISequence(text, index)
			if ok {
				pendingPrefix += seq
				index = next
				continue
			}
		}
		if text[index] == 0x9d {
			seq, next, ok := consumeANSISequence(text, index)
			if ok {
				pendingPrefix += seq
				index = next
				continue
			}
		}

		if text[index] != 0x1b || index+1 >= len(text) || text[index+1] != '[' {
			r, size := rune(text[index]), 1
			if text[index] >= 0x80 {
				r, size = utf8.DecodeRuneInString(text[index:])
			}
			result = append(result, styledRune{ch: r, style: current, prefix: pendingPrefix})
			pendingPrefix = ""
			index += size
			continue
		}

		end := index + 2
		for end < len(text) && ((text[end] >= '0' && text[end] <= '9') || text[end] == ';') {
			end++
		}

		if end >= len(text) || text[end] != 'm' {
			index++
			continue
		}

		codesText := text[index+2 : end]
		codes := []int{0}
		if codesText != "" {
			parts := strings.Split(codesText, ";")
			codes = make([]int, 0, len(parts))
			for _, part := range parts {
				if part == "" {
					codes = append(codes, 0)
					continue
				}
				value, err := strconv.Atoi(part)
				if err != nil {
					codes = append(codes, 0)
					continue
				}
				codes = append(codes, value)
			}
		}

		for i := 0; i < len(codes); i++ {
			code := codes[i]
			switch code {
			case 0:
				resetANSIStyleToBase(&current, baseStyle)
			case 1:
				current.bold = true
				current.dim = false
			case 2:
				current.dim = true
				current.bold = false
			case 3:
				current.italic = true
			case 4:
				current.underline = true
			case 7:
				current.inverse = true
			case 9:
				current.strikethrough = true
			case 22:
				current.bold = baseStyle.bold
				current.dim = baseStyle.dim
			case 23:
				current.italic = baseStyle.italic
			case 24:
				current.underline = baseStyle.underline
			case 27:
				current.inverse = baseStyle.inverse
			case 29:
				current.strikethrough = baseStyle.strikethrough
			case 39:
				current.fg = baseStyle.fg
			case 49:
				current.bg = baseStyle.bg
			default:
				switch {
				case code >= 30 && code <= 37:
					current.fg = fmt.Sprintf("\x1b[%dm", code)
				case code >= 40 && code <= 47:
					current.bg = fmt.Sprintf("\x1b[%dm", code)
				case code == 38 && i+1 < len(codes):
					if codes[i+1] == 5 && i+2 < len(codes) {
						current.fg = fmt.Sprintf("\x1b[38;5;%dm", codes[i+2])
						i += 2
					} else if codes[i+1] == 2 && i+4 < len(codes) {
						current.fg = fmt.Sprintf("\x1b[38;2;%d;%d;%dm", codes[i+2], codes[i+3], codes[i+4])
						i += 4
					}
				case code == 48 && i+1 < len(codes):
					if codes[i+1] == 5 && i+2 < len(codes) {
						current.bg = fmt.Sprintf("\x1b[48;5;%dm", codes[i+2])
						i += 2
					} else if codes[i+1] == 2 && i+4 < len(codes) {
						current.bg = fmt.Sprintf("\x1b[48;2;%d;%d;%dm", codes[i+2], codes[i+3], codes[i+4])
						i += 4
					}
				}
			}
		}

		index = end + 1
	}

	flushPrefix()
	return result
}

func applyStyledTransformRunes(node *vdom.Node, runes []styledRune, index int, defaultStyle ansiStyle) ([]styledRune, bool) {
	if node == nil || node.Props == nil {
		return runes, true
	}

	transform, ok := node.Props["transform"].(func(string, int) string)
	if !ok || transform == nil || len(runes) == 0 {
		return runes, true
	}

	input := styledRunesToANSIString(runes, defaultStyle)
	output := transform(input, index)
	if output == "" {
		return []styledRune{}, true
	}
	if output == input {
		return runes, true
	}

	return parseANSIToStyledRunes(output, defaultStyle), true
}

func collectStyledRenderedTextContent(node *vdom.Node, inheritedStyle ansiStyle, inheritedBackground string) ([]styledRune, bool) {
	if node == nil {
		return nil, true
	}

	switch node.Type {
	case vdom.TextNode:
		return parseANSIToStyledRunes(node.Text, inheritedStyle), true
	case vdom.ElementNode:
		currentStyle := inheritedStyle
		currentBackground := inheritedBackground
		if node.ElementType == "text" {
			currentStyle, currentBackground = mergeTextANSIStyle(node, inheritedStyle, inheritedBackground)
		}

		collected := make([]styledRune, 0)
		for childIndex, child := range node.Children {
			if child == nil {
				continue
			}

			if node.ElementType == "text" || node.ElementType == "transform" {
				switch {
				case isTextLikeNode(child):
					childRunes, ok := renderNestedStyledTextLikeNode(child, currentStyle, currentBackground, childIndex)
					if !ok {
						return nil, false
					}
					collected = append(collected, childRunes...)
				case child.Type == vdom.TextNode:
					childRunes, ok := collectStyledRenderedTextContent(child, currentStyle, currentBackground)
					if !ok {
						return nil, false
					}
					collected = append(collected, childRunes...)
				default:
					// Text-like containers only accept text content.
				}

				continue
			}

			switch {
			case isTextLikeNode(child):
				childRunes, ok := renderNestedStyledTextLikeNode(child, currentStyle, currentBackground, childIndex)
				if !ok {
					return nil, false
				}
				collected = append(collected, childRunes...)
			case child.Type == vdom.TextNode:
				childRunes, ok := collectStyledRenderedTextContent(child, currentStyle, currentBackground)
				if !ok {
					return nil, false
				}
				collected = append(collected, childRunes...)
			default:
				childRunes, ok := collectStyledRenderedTextContent(child, currentStyle, currentBackground)
				if !ok {
					return nil, false
				}
				collected = append(collected, childRunes...)
			}
		}

		return collected, true
	default:
		return nil, true
	}
}

func renderNestedStyledTextLikeNode(node *vdom.Node, inheritedStyle ansiStyle, inheritedBackground string, childIndex int) ([]styledRune, bool) {
	if node == nil {
		return nil, true
	}

	runes, ok := collectStyledRenderedTextContent(node, inheritedStyle, inheritedBackground)
	if !ok {
		return nil, false
	}

	return applyStyledTransformRunes(node, runes, childIndex, inheritedStyle)
}

func styledRunesWidth(runes []styledRune) int {
	width := 0
	for _, r := range runes {
		if r.ch == '\n' {
			continue
		}

		width += utils.RuneWidth(r.ch)
	}

	return width
}

// fitStyledRunesToWidth returns the number of runes from the front of the
// slice whose cumulative display width fits in maxWidth. NOTE: this walks
// rune-by-rune so a multi-rune cluster (ZWJ emoji, combining marks) can
// be split across the boundary. The styled-rune render path and width
// calculation are also rune-based, so keeping the boundary aligned with
// that model preserves end-to-end consistency. Cluster-aware wrapping in
// the styled path requires also moving writeStyledLinesClipped onto
// cluster-based column tracking — see deferred work in the wide-rune
// task notes.
func fitStyledRunesToWidth(runes []styledRune, maxWidth int) int {
	if maxWidth <= 0 {
		return 0
	}

	width := 0
	for index, r := range runes {
		runeWidth := utils.RuneWidth(r.ch)
		if width+runeWidth > maxWidth {
			if index == 0 {
				return 1
			}
			return index
		}

		width += runeWidth
	}

	return len(runes)
}

func trimLeftSpaceStyledRunes(runes []styledRune) []styledRune {
	index := 0
	for index < len(runes) && unicode.IsSpace(runes[index].ch) {
		index++
	}

	return runes[index:]
}

func trimRightSpaceStyledRunes(runes []styledRune) []styledRune {
	end := len(runes)
	for end > 0 && unicode.IsSpace(runes[end-1].ch) {
		end--
	}

	return runes[:end]
}

func splitStyledRunesLines(runes []styledRune) [][]styledRune {
	lines := make([][]styledRune, 0, 1)
	current := make([]styledRune, 0, len(runes))
	for _, r := range runes {
		if r.ch == '\n' {
			lines = append(lines, current)
			current = make([]styledRune, 0)
			continue
		}

		current = append(current, r)
	}

	lines = append(lines, current)
	return lines
}

func wrapStyledLine(line []styledRune, maxWidth int) [][]styledRune {
	if maxWidth <= 0 || styledRunesWidth(line) <= maxWidth {
		return [][]styledRune{line}
	}

	runes := append([]styledRune(nil), line...)
	lines := make([][]styledRune, 0, 2)
	for len(runes) > 0 {
		if styledRunesWidth(runes) <= maxWidth {
			lines = append(lines, runes)
			break
		}

		fit := fitStyledRunesToWidth(runes, maxWidth)
		if fit < len(runes) && unicode.IsSpace(runes[fit].ch) {
			segment := append([]styledRune(nil), runes[:fit]...)
			lines = append(lines, segment)
			runes = runes[fit:]
			continue
		}

		breakIndex := -1
		seenNonSpace := false
		for index := 0; index < fit; index++ {
			if !unicode.IsSpace(runes[index].ch) {
				seenNonSpace = true
			}
			if seenNonSpace && unicode.IsSpace(runes[index].ch) {
				breakIndex = index
			}
		}

		if breakIndex > 0 {
			segment := append([]styledRune(nil), runes[:breakIndex+1]...)
			runes = runes[breakIndex+1:]
			if len(segment) > 0 {
				lines = append(lines, segment)
				continue
			}
		}

		segment := append([]styledRune(nil), runes[:fit]...)
		lines = append(lines, segment)
		runes = runes[fit:]
	}

	if len(lines) == 0 {
		return [][]styledRune{{}}
	}

	return lines
}

func ellipsisStyledRune(style ansiStyle) styledRune {
	return styledRune{ch: '…', style: style}
}

func truncateStyledEnd(line []styledRune, maxWidth int) []styledRune {
	if maxWidth <= 0 || styledRunesWidth(line) <= maxWidth {
		return line
	}
	if maxWidth == 1 {
		style := ansiStyle{}
		if len(line) > 0 {
			style = line[0].style
		}
		return []styledRune{ellipsisStyledRune(style)}
	}

	keep := fitStyledRunesToWidth(line, maxWidth-1)
	if keep <= 0 {
		style := ansiStyle{}
		if len(line) > 0 {
			style = line[0].style
		}
		return []styledRune{ellipsisStyledRune(style)}
	}

	segment := append([]styledRune(nil), line[:keep]...)
	ellipsisStyle := segment[len(segment)-1].style
	return append(segment, ellipsisStyledRune(ellipsisStyle))
}

func truncateStyledStart(line []styledRune, maxWidth int) []styledRune {
	if maxWidth <= 0 || styledRunesWidth(line) <= maxWidth {
		return line
	}
	if maxWidth == 1 {
		style := ansiStyle{}
		if len(line) > 0 {
			style = line[len(line)-1].style
		}
		return []styledRune{ellipsisStyledRune(style)}
	}

	keepWidth := maxWidth - 1
	start := len(line)
	width := 0
	for start > 0 {
		runeWidth := utils.RuneWidth(line[start-1].ch)
		if width+runeWidth > keepWidth {
			break
		}
		start--
		width += runeWidth
	}

	segment := append([]styledRune(nil), line[start:]...)
	ellipsisStyle := ansiStyle{}
	if len(segment) > 0 {
		ellipsisStyle = segment[0].style
	}
	return append([]styledRune{ellipsisStyledRune(ellipsisStyle)}, segment...)
}

func truncateStyledMiddle(line []styledRune, maxWidth int) []styledRune {
	if maxWidth <= 0 || styledRunesWidth(line) <= maxWidth {
		return line
	}
	if maxWidth == 1 {
		style := ansiStyle{}
		if len(line) > 0 {
			style = line[0].style
		}
		return []styledRune{ellipsisStyledRune(style)}
	}

	remainingWidth := maxWidth - 1
	leftWidth := (remainingWidth + 1) / 2
	rightWidth := remainingWidth / 2

	leftEnd := fitStyledRunesToWidth(line, leftWidth)
	left := append([]styledRune(nil), line[:leftEnd]...)

	start := len(line)
	width := 0
	for start > 0 {
		runeWidth := utils.RuneWidth(line[start-1].ch)
		if width+runeWidth > rightWidth {
			break
		}
		start--
		width += runeWidth
	}

	right := append([]styledRune(nil), line[start:]...)
	ellipsisStyle := ansiStyle{}
	if len(left) > 0 {
		ellipsisStyle = left[len(left)-1].style
	} else if len(right) > 0 {
		ellipsisStyle = right[0].style
	}

	result := make([]styledRune, 0, len(left)+1+len(right))
	result = append(result, left...)
	result = append(result, ellipsisStyledRune(ellipsisStyle))
	result = append(result, right...)
	return result
}

func styledTextLines(node *vdom.Node, inheritedBackground string, maxWidth int) ([][]styledRune, bool) {
	baseStyle := ansiStyle{}
	baseStyle.bg = parseBackgroundANSI(inheritedBackground)
	runes, ok := collectStyledRenderedTextContent(node, baseStyle, inheritedBackground)
	if !ok {
		return nil, false
	}
	lines := splitStyledRunesLines(runes)
	if maxWidth <= 0 {
		return lines, true
	}

	mode := textWrapMode(node)
	processed := make([][]styledRune, 0, len(lines))
	for _, line := range lines {
		switch mode {
		case "truncate", "truncate-end":
			processed = append(processed, truncateStyledEnd(line, maxWidth))
		case "truncate-middle":
			processed = append(processed, truncateStyledMiddle(line, maxWidth))
		case "truncate-start":
			processed = append(processed, truncateStyledStart(line, maxWidth))
		case "wrap", "":
			processed = append(processed, wrapStyledLine(line, maxWidth)...)
		default:
			processed = append(processed, line)
		}
	}

	if node != nil && node.Type == vdom.ElementNode && node.ElementType == "transform" {
		transformed := make([][]styledRune, 0, len(processed))
		for index, line := range processed {
			lineRunes, ok := applyStyledTransformRunes(node, line, index, baseStyle)
			if !ok {
				return nil, false
			}
			transformed = append(transformed, lineRunes)
		}
		return transformed, true
	}

	return processed, true
}

func writeStyledLinesClipped(canvas *ansiCanvas, x, y int, lines [][]styledRune, clip clipRect) {
	if canvas == nil {
		return
	}

	for rowIndex, line := range lines {
		row := y + rowIndex
		if row < clip.top || row >= clip.bottom {
			continue
		}

		column := x
		for _, styled := range line {
			runeWidth := utils.RuneWidth(styled.ch)
			if runeWidth == 0 {
				// Zero-width entry: combining mark, ZWJ joiner, or an
				// OSC-only sentinel (ch == 0 with prefix). Land it on the
				// preceding visible cell so the OSC bytes survive
				// without affecting column accounting.
				if styled.prefix != "" {
					canvas.appendZeroWidthRaw(column, row, styled.prefix)
				}
				if styled.ch != 0 && column > clip.left && column <= clip.right {
					canvas.setCell(column, row, styled.ch, styled.style)
				}
				continue
			}

			if column >= clip.left && column < clip.right {
				canvas.setCellWithPrefix(column, row, styled.ch, styled.style, styled.prefix)
			}

			column += runeWidth
		}
	}
}

func getInheritedBackground(node *vdom.Node, inherited string) string {
	if node == nil || node.Type != vdom.ElementNode || node.Props == nil {
		return inherited
	}

	if backgroundColor, ok := node.Props["backgroundColor"].(string); ok {
		return backgroundColor
	}

	return inherited
}

func resolveTextBackground(node *vdom.Node, inherited string) string {
	if node == nil || node.Props == nil {
		return inherited
	}

	if explicitBackground, ok := node.Props["backgroundColor"].(string); ok {
		return explicitBackground
	}

	return inherited
}

func resolveTextStyle(node *vdom.Node, inheritedBackground string) ansiStyle {
	if node == nil || node.Props == nil {
		return ansiStyle{}
	}

	style := ansiStyle{}
	if colorSpec, ok := node.Props["color"].(string); ok {
		if color, ok := styles.ParseColor(colorSpec); ok {
			style.fg = color.ToANSI(styles.Foreground)
		}
	}

	backgroundSpec := resolveTextBackground(node, inheritedBackground)
	if backgroundSpec != "" {
		if color, ok := styles.ParseColor(backgroundSpec); ok {
			style.bg = color.ToANSI(styles.Background)
		}
	}

	if dimColor, _ := node.Props["dimColor"].(bool); dimColor {
		style.dim = true
	}
	if bold, _ := node.Props["bold"].(bool); bold {
		style.bold = true
	}
	if italic, _ := node.Props["italic"].(bool); italic {
		style.italic = true
	}
	if underline, _ := node.Props["underline"].(bool); underline {
		style.underline = true
	}
	if inverse, _ := node.Props["inverse"].(bool); inverse {
		style.inverse = true
	}
	if strikethrough, _ := node.Props["strikethrough"].(bool); strikethrough {
		style.strikethrough = true
	}

	return style
}

func resolveBoxBackgroundStyle(node *vdom.Node, inheritedBackground string) ansiStyle {
	backgroundSpec := getInheritedBackground(node, inheritedBackground)
	if backgroundSpec == "" {
		return ansiStyle{}
	}

	color, ok := styles.ParseColor(backgroundSpec)
	if !ok {
		return ansiStyle{}
	}

	return ansiStyle{bg: color.ToANSI(styles.Background)}
}

func resolveBorderSideStyle(props vdom.Props, colorKey string, dimKey string) ansiStyle {
	if props == nil {
		return ansiStyle{}
	}

	style := ansiStyle{}

	colorSpec := ""
	if explicitColor, ok := props[colorKey].(string); ok {
		colorSpec = explicitColor
	} else if borderColor, ok := props["borderColor"].(string); ok {
		colorSpec = borderColor
	}

	if colorSpec != "" {
		if color, ok := styles.ParseColor(colorSpec); ok {
			style.fg = color.ToANSI(styles.Foreground)
		}
	}

	if dim, _ := props[dimKey].(bool); dim {
		style.dim = true
	} else if dim, _ := props["borderDimColor"].(bool); dim {
		style.dim = true
	}

	return style
}

func resolveBorderStyle(node *vdom.Node) ansiStyle {
	if node == nil || node.Props == nil {
		return ansiStyle{}
	}

	return resolveBorderSideStyle(node.Props, "borderTopColor", "borderTopDimColor")
}

func applyTextProps(node *vdom.Node, text string, inheritedBackground string) string {
	if node == nil || node.Props == nil || text == "" {
		return text
	}

	codes := make([]string, 0, 8)

	if colorSpec, ok := node.Props["color"].(string); ok {
		if color, ok := styles.ParseColor(colorSpec); ok {
			codes = append(codes, color.ToANSI(styles.Foreground))
		}
	}

	backgroundSpec := resolveTextBackground(node, inheritedBackground)
	if backgroundSpec != "" {
		if color, ok := styles.ParseColor(backgroundSpec); ok {
			codes = append(codes, color.ToANSI(styles.Background))
		}
	}

	if dimColor, _ := node.Props["dimColor"].(bool); dimColor {
		codes = append(codes, styles.DimCode())
	}
	if bold, _ := node.Props["bold"].(bool); bold {
		codes = append(codes, styles.BoldCode())
	}
	if italic, _ := node.Props["italic"].(bool); italic {
		codes = append(codes, styles.ItalicCode())
	}
	if underline, _ := node.Props["underline"].(bool); underline {
		codes = append(codes, styles.UnderlineCode())
	}
	if inverse, _ := node.Props["inverse"].(bool); inverse {
		codes = append(codes, styles.InverseCode())
	}
	if strikethrough, _ := node.Props["strikethrough"].(bool); strikethrough {
		codes = append(codes, styles.StrikethroughCode())
	}

	return styles.WrapWithANSI(text, codes...)
}

func collectRenderedTextContent(node *vdom.Node, inheritedBackground string, enableANSI bool) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case vdom.TextNode:
		return node.Text
	case vdom.ElementNode:
		currentBackground := getInheritedBackground(node, inheritedBackground)
		if node.ElementType == "text" || node.ElementType == "transform" {
			text := ""
			for childIndex, child := range node.Children {
				if child == nil {
					continue
				}

				switch {
				case isTextLikeNode(child):
					text += renderNestedTextLikeNode(child, currentBackground, childIndex, enableANSI)
				case child.Type == vdom.TextNode:
					text += child.Text
				default:
					// Text-like containers only accept text content.
				}
			}

			return text
		}

		text := ""
		for childIndex, child := range node.Children {
			if child == nil {
				continue
			}

			switch {
			case isTextLikeNode(child):
				text += renderNestedTextLikeNode(child, currentBackground, childIndex, enableANSI)
			case child.Type == vdom.TextNode:
				text += child.Text
			default:
				text += collectRenderedTextContent(child, currentBackground, enableANSI)
			}
		}

		return text
	default:
		return ""
	}
}

func renderNestedTextLikeNode(node *vdom.Node, inheritedBackground string, childIndex int, enableANSI bool) string {
	if node == nil {
		return ""
	}

	text := collectRenderedTextContent(node, inheritedBackground, enableANSI)
	if enableANSI && node.ElementType == "text" {
		text = applyTextProps(node, text, inheritedBackground)
	}

	return applyNodeTransform(node, text, childIndex)
}

func renderTextLikeNode(node *vdom.Node, inheritedBackground string, enableANSI bool, maxWidth int) string {
	if node == nil {
		return ""
	}

	text := collectRenderedTextContent(node, inheritedBackground, enableANSI)
	if enableANSI && node.ElementType == "text" {
		text = applyTextProps(node, text, inheritedBackground)
	}

	text = applyTextLayoutMode(text, textWrapMode(node), maxWidth)
	return applyLineTransform(node, text)
}

func measureTextBlock(text string) (width float64, height float64) {
	if text == "" {
		return 0, 0
	}

	lines := strings.Split(text, "\n")
	maxWidth := 0
	for _, line := range lines {
		lineWidth := visibleStringWidth(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	return float64(maxWidth), float64(len(lines))
}

func applyContainerLayoutProps(node *layout.Node, props vdom.Props, defaultDirection layout.FlexDirection) {
	node.SetFlexDirection(defaultDirection)
	node.SetFlexShrink(1)
	node.SetJustifyContent(layout.JustifyStart)

	if props == nil {
		return
	}

	reverseDirection := false
	if direction, ok := props["flexDirection"].(string); ok {
		reverseDirection = direction == "row-reverse" || direction == "column-reverse"
	}

	if width, percent, ok := parseSizeValue(props["width"]); ok {
		if percent {
			node.SetWidthPercent(width)
		} else {
			node.SetWidth(width)
		}
	}
	if height, percent, ok := parseSizeValue(props["height"]); ok {
		if percent {
			node.SetHeightPercent(height)
		} else {
			node.SetHeight(height)
		}
	}
	if minWidth, percent, ok := parseSizeValue(props["minWidth"]); ok {
		if percent {
			node.SetMinWidthPercent(minWidth)
		} else {
			node.SetMinWidth(minWidth)
		}
	}
	if minHeight, percent, ok := parseSizeValue(props["minHeight"]); ok {
		if percent {
			node.SetMinHeightPercent(minHeight)
		} else {
			node.SetMinHeight(minHeight)
		}
	}

	if fg, ok := parseNumericValue(props["flexGrow"]); ok {
		node.SetFlexGrow(fg)
	}
	if fs, ok := parseNumericValue(props["flexShrink"]); ok {
		node.SetFlexShrink(fs)
	}

	if fd, ok := parseFlexDirection(props["flexDirection"]); ok {
		node.SetFlexDirection(fd)
	}

	justify := layout.JustifyStart
	if jc, ok := parseJustifyContent(props["justifyContent"]); ok {
		justify = jc
	}
	if reverseDirection {
		switch justify {
		case layout.JustifyStart:
			justify = layout.JustifyEnd
		case layout.JustifyEnd:
			justify = layout.JustifyStart
		}
	}
	node.SetJustifyContent(justify)

	if ai, ok := parseAlignItems(props["alignItems"]); ok {
		node.SetAlignItems(ai)
	}
	if as, ok := parseAlignItems(props["alignSelf"]); ok {
		node.SetAlignSelf(as)
	}
	if basis, percent, ok := parseFlexBasis(props["flexBasis"]); ok {
		if percent {
			node.SetFlexBasisPercent(basis)
		} else {
			node.SetFlexBasis(basis)
		}
	}
	if gap, ok := parseNumericValue(props["gap"]); ok {
		node.SetGap(gap)
	}
	if gap, ok := parseNumericValue(props["rowGap"]); ok {
		node.SetRowGap(gap)
	}
	if gap, ok := parseNumericValue(props["columnGap"]); ok {
		node.SetColumnGap(gap)
	}
	if wrapMode, ok := parseWrapMode(props["flexWrap"]); ok {
		node.SetWrapMode(wrapMode)
	}

	borderLeftInset, borderTopInset, borderRightInset, borderBottomInset := borderInsets(props)
	// Resolve padding using Yoga precedence: per-edge values override shorthand.
	// padding < paddingX/Y < paddingLeft/Top/Right/Bottom.
	var paddingLeft, paddingTop, paddingRight, paddingBottom float64
	if p, ok := parseNumericValue(props["padding"]); ok {
		paddingLeft = p
		paddingTop = p
		paddingRight = p
		paddingBottom = p
	}
	if p, ok := parseNumericValue(props["paddingX"]); ok {
		paddingLeft = p
		paddingRight = p
	}
	if p, ok := parseNumericValue(props["paddingY"]); ok {
		paddingTop = p
		paddingBottom = p
	}
	if p, ok := parseNumericValue(props["paddingLeft"]); ok {
		paddingLeft = p
	}
	if p, ok := parseNumericValue(props["paddingTop"]); ok {
		paddingTop = p
	}
	if p, ok := parseNumericValue(props["paddingRight"]); ok {
		paddingRight = p
	}
	if p, ok := parseNumericValue(props["paddingBottom"]); ok {
		paddingBottom = p
	}
	node.SetPadding(layout.EdgeLeft, paddingLeft+borderLeftInset)
	node.SetPadding(layout.EdgeTop, paddingTop+borderTopInset)
	node.SetPadding(layout.EdgeRight, paddingRight+borderRightInset)
	node.SetPadding(layout.EdgeBottom, paddingBottom+borderBottomInset)

	if m, ok := parseNumericValue(props["margin"]); ok {
		node.SetMargin(layout.EdgeAll, m)
	}
	if m, ok := parseNumericValue(props["marginX"]); ok {
		node.SetMargin(layout.EdgeLeft, m)
		node.SetMargin(layout.EdgeRight, m)
	}
	if m, ok := parseNumericValue(props["marginY"]); ok {
		node.SetMargin(layout.EdgeTop, m)
		node.SetMargin(layout.EdgeBottom, m)
	}
	if m, ok := parseNumericValue(props["marginLeft"]); ok {
		node.SetMargin(layout.EdgeLeft, m)
	}
	if m, ok := parseNumericValue(props["marginTop"]); ok {
		node.SetMargin(layout.EdgeTop, m)
	}
	if m, ok := parseNumericValue(props["marginRight"]); ok {
		node.SetMargin(layout.EdgeRight, m)
	}
	if m, ok := parseNumericValue(props["marginBottom"]); ok {
		node.SetMargin(layout.EdgeBottom, m)
	}

	// position prop: Yoga's POSITION_TYPE_ABSOLUTE removes the child from
	// flex flow. We default to relative; only "absolute" is recognized.
	if position, ok := props["position"].(string); ok && position == "absolute" {
		node.SetPosition(layout.PositionAbsolute)
	}
	if v, ok := parseNumericValue(props["top"]); ok {
		node.SetPositionTop(v)
	}
	if v, ok := parseNumericValue(props["left"]); ok {
		node.SetPositionLeft(v)
	}
	if v, ok := parseNumericValue(props["right"]); ok {
		node.SetPositionRight(v)
	}
	if v, ok := parseNumericValue(props["bottom"]); ok {
		node.SetPositionBottom(v)
	}
}

// Render renders a virtual DOM tree to a string (simple, no layout)
func Render(node *vdom.Node, width, height int) string {
	buf := buffer.New(width, height)
	renderNode(buf, node, 0, 0, "", false)
	return buf.Render()
}

// RenderWithLayout renders a virtual DOM tree using flexbox layout
func RenderWithLayout(node *vdom.Node, width, height int) string {
	if node == nil {
		return ""
	}

	buf := buffer.New(width, height)

	// Build layout tree from vdom
	layoutNode := buildLayoutTreeWithWidth(node, width)
	if shouldUseAvailableRootWidth(node) {
		layoutNode.SetWidth(float64(width))
	}

	// Calculate layout
	layoutNode.CalculateLayout()
	syncComputedLayout(node, layoutNode)

	// Render using layout positions
	rootClip := clipRect{left: 0, top: 0, right: int(^uint(0) >> 1), bottom: height}
	renderNodeWithLayout(buf, node, layoutNode, "", false, width, rootClip)

	rows := int(layoutNode.GetOuterComputedHeight())
	if rows <= 0 {
		return buf.Render()
	}

	output := buf.RenderRows(rows)
	if node != nil && node.Type == vdom.ElementNode && node.ElementType == "static" && output != "" {
		return output + "\n"
	}

	return output
}

// RenderWithLayoutANSI renders a virtual DOM tree using layout plus ANSI fg/bg styling.
func RenderWithLayoutANSI(node *vdom.Node, width, height int) string {
	if node == nil {
		return ""
	}

	layoutNode := buildLayoutTreeWithWidth(node, width)
	if shouldUseAvailableRootWidth(node) {
		layoutNode.SetWidth(float64(width))
	}

	layoutNode.CalculateLayout()
	syncComputedLayout(node, layoutNode)

	rows := int(layoutNode.GetOuterComputedHeight())
	if rows <= 0 {
		return ""
	}

	canvasWidth := int(math.Round(layoutNode.GetOuterComputedWidth()))
	if canvasWidth < 0 {
		canvasWidth = 0
	}

	canvas := newANSICanvas(canvasWidth, rows)
	rootClip := clipRect{left: 0, top: 0, right: int(^uint(0) >> 1), bottom: rows}
	renderNodeWithLayoutANSI(canvas, node, layoutNode, "", width, rootClip)

	output := canvas.RenderRows(rows)
	if node != nil && node.Type == vdom.ElementNode && node.ElementType == "static" && output != "" {
		return output + "\n"
	}

	return output
}

// RenderWithLayoutANSI256 renders ANSI output with truecolor SGR sequences
// downgraded to xterm 256-color codes, matching Chalk on 256-color terminals.
func RenderWithLayoutANSI256(node *vdom.Node, width, height int) string {
	return styles.DowngradeTruecolorANSIToANSI256(RenderWithLayoutANSI(node, width, height))
}

// SyncComputedLayout calculates layout for a tree and stores the computed values on the original nodes.
func SyncComputedLayout(node *vdom.Node, width, height int) {
	if node == nil {
		return
	}

	layoutNode := buildLayoutTreeWithWidth(node, width)
	if shouldUseAvailableRootWidth(node) {
		layoutNode.SetWidth(float64(width))
	}
	layoutNode.CalculateLayout()
	syncComputedLayout(node, layoutNode)
}

// RenderWithLayoutSections renders dynamic and static output separately.
func RenderWithLayoutSections(node *vdom.Node, width, height int) RenderSections {
	return RenderWithLayoutSectionsMode(node, width, height, false)
}

// RenderWithLayoutSectionsMode renders dynamic and static output separately, optionally with ANSI styling.
func RenderWithLayoutSectionsMode(node *vdom.Node, width, height int, enableANSI bool) RenderSections {
	if node == nil {
		return RenderSections{}
	}

	staticRoots := collectStaticRoots(node)
	mainNode := cloneWithoutStatic(node)

	sections := RenderSections{}
	if mainNode != nil {
		sections.Output = renderWithLayoutMode(mainNode, width, height, enableANSI)
	}

	if len(staticRoots) == 0 {
		return sections
	}

	// Walk the original tree once to collect cache pointers for the same
	// `<static>` nodes we just cloned. The slice from collectStaticRoots
	// holds clones, but the mutation hooks that flip the dirty flag run
	// on the originals — so cache reads/writes must use the originals.
	originalStaticRoots := collectStaticRootsRef(node)

	var staticBuilder strings.Builder
	for index, staticRoot := range staticRoots {
		var origRoot *vdom.Node
		if index < len(originalStaticRoots) {
			origRoot = originalStaticRoots[index]
		}

		if origRoot != nil {
			if cached, hit := origRoot.LookupStaticOutput(width, height, enableANSI); hit {
				staticBuilder.WriteString(cached)
				continue
			}
		}

		rendered := renderWithLayoutMode(staticRoot, width, height, enableANSI)
		if origRoot != nil {
			origRoot.StoreStaticOutput(rendered, width, height, enableANSI)
		}
		staticBuilder.WriteString(rendered)
	}

	sections.StaticOutput = staticBuilder.String()
	return sections
}

// collectStaticRootsRef walks the tree and returns pointers to the original
// `<static>` nodes (no Clone). Result is parallel-indexed with
// collectStaticRoots so cache lookups land on the same node identity used by
// the mutation hooks. Callers must NOT mutate these pointers.
func collectStaticRootsRef(node *vdom.Node) []*vdom.Node {
	if node == nil {
		return nil
	}

	if node.Type == vdom.ElementNode && node.ElementType == "static" {
		return []*vdom.Node{node}
	}

	roots := make([]*vdom.Node, 0)
	for _, child := range node.Children {
		roots = append(roots, collectStaticRootsRef(child)...)
	}

	return roots
}

// RenderScreenReaderSections renders a tree as plain accessibility text.
func RenderScreenReaderSections(node *vdom.Node) RenderSections {
	if node == nil {
		return RenderSections{}
	}

	staticRoots := collectStaticRoots(node)
	mainNode := cloneWithoutStatic(node)

	// Build the id-map across the *full* tree (including hidden / aria-live=off
	// subtrees) so labelledby / describedby can still resolve into hidden nodes
	// per the spec.
	ctx := newScreenReaderContext(node)

	sections := RenderSections{}
	if mainNode != nil {
		main := renderScreenReaderNode(mainNode, "", false, ctx)
		announcer := renderAnnouncerRegions(mainNode, ctx)
		sections.Output = joinScreenReaderOutput(main, announcer)
	}

	if len(staticRoots) == 0 {
		return sections
	}

	var staticBuilder strings.Builder
	for _, staticRoot := range staticRoots {
		rendered := renderScreenReaderNode(staticRoot, "", false, ctx)
		if rendered == "" {
			continue
		}

		staticBuilder.WriteString(rendered)
		if !strings.HasSuffix(rendered, "\n") {
			staticBuilder.WriteString("\n")
		}
	}

	sections.StaticOutput = staticBuilder.String()
	return sections
}

// joinScreenReaderOutput concatenates the inline narration with the announcer
// block, separated by a single newline when both halves are non-empty.
func joinScreenReaderOutput(main, announcer string) string {
	if announcer == "" {
		return main
	}
	if main == "" {
		return announcer
	}
	return main + "\n" + announcer
}

// RenderRuntimeSections renders dynamic output plus only the newly appended static delta.
func RenderRuntimeSections(node *vdom.Node, width, height int, previousStaticCounts []int, screenReader bool) RenderSections {
	return RenderRuntimeSectionsMode(node, width, height, previousStaticCounts, screenReader, false)
}

// RenderRuntimeSectionsMode renders dynamic output plus only the newly appended static delta, optionally with ANSI styling.
func RenderRuntimeSectionsMode(node *vdom.Node, width, height int, previousStaticCounts []int, screenReader bool, enableANSI bool) RenderSections {
	if node == nil {
		return RenderSections{}
	}

	staticRoots := collectStaticRoots(node)
	mainNode := cloneWithoutStatic(node)

	sections := RenderSections{
		StaticCounts: make([]int, 0, len(staticRoots)),
	}

	if mainNode != nil {
		if screenReader {
			ctx := newScreenReaderContext(node)
			main := renderScreenReaderNode(mainNode, "", false, ctx)
			announcer := renderAnnouncerRegions(mainNode, ctx)
			sections.Output = joinScreenReaderOutput(main, announcer)
		} else {
			sections.Output = renderWithLayoutMode(mainNode, width, height, enableANSI)
		}
	}

	var deltaBuilder strings.Builder
	for index, staticRoot := range staticRoots {
		previousCount := 0
		if index < len(previousStaticCounts) {
			previousCount = previousStaticCounts[index]
		}

		delta, nextCount := renderStaticRootDelta(staticRoot, width, height, previousCount, screenReader, enableANSI)
		sections.StaticCounts = append(sections.StaticCounts, nextCount)
		if delta != "" {
			deltaBuilder.WriteString(delta)
		}
	}

	sections.StaticDeltaOutput = deltaBuilder.String()
	return sections
}

func collectStaticRoots(node *vdom.Node) []*vdom.Node {
	if node == nil {
		return nil
	}

	if node.Type == vdom.ElementNode && node.ElementType == "static" {
		return []*vdom.Node{node.Clone()}
	}

	roots := make([]*vdom.Node, 0)
	for _, child := range node.Children {
		roots = append(roots, collectStaticRoots(child)...)
	}

	return roots
}

func cloneWithoutStatic(node *vdom.Node) *vdom.Node {
	if node == nil {
		return nil
	}

	if node.Type == vdom.ElementNode && node.ElementType == "static" {
		return nil
	}

	cloned := node.Clone()
	if cloned.Type != vdom.ElementNode || len(cloned.Children) == 0 {
		return cloned
	}

	filteredChildren := make([]*vdom.Node, 0, len(cloned.Children))
	for _, child := range node.Children {
		filteredChild := cloneWithoutStatic(child)
		if filteredChild != nil {
			filteredChildren = append(filteredChildren, filteredChild)
		}
	}

	rebuilt := vdom.CreateElement(cloned.ElementType, cloned.Props, filteredChildren...)
	rebuilt.Key = cloned.Key
	return rebuilt
}

func renderStaticRootDelta(root *vdom.Node, width, height int, previousCount int, screenReader bool, enableANSI bool) (string, int) {
	if root == nil {
		return "", 0
	}

	count, tracked := staticRootTrackedCount(root)
	if !tracked {
		if previousCount > 0 {
			return "", previousCount
		}

		if screenReader {
			return renderScreenReaderNode(root, "", false, newScreenReaderContext(root)), 1
		}

		return renderWithLayoutMode(root, width, height, enableANSI), 1
	}

	if count <= previousCount {
		return "", count
	}

	deltaRoot := root.Clone()
	if previousCount < len(root.Children) {
		deltaRoot.Children = make([]*vdom.Node, 0, len(root.Children)-previousCount)
		for _, child := range root.Children[previousCount:] {
			deltaRoot.Children = append(deltaRoot.Children, child.Clone())
		}
	} else {
		deltaRoot.Children = nil
	}

	if screenReader {
		out := renderScreenReaderNode(deltaRoot, "", false, newScreenReaderContext(deltaRoot))
		// Screen-reader static deltas need an explicit boundary so newly
		// appended items don't run into the live region. Mirror upstream's
		// per-item newline append when emitting just the delta.
		if out != "" && !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		return out, count
	}

	return renderWithLayoutMode(deltaRoot, width, height, enableANSI), count
}

func renderWithLayoutMode(node *vdom.Node, width, height int, enableANSI bool) string {
	if enableANSI {
		return RenderWithLayoutANSI(node, width, height)
	}

	return RenderWithLayout(node, width, height)
}

func staticRootTrackedCount(root *vdom.Node) (int, bool) {
	if root == nil || root.Props == nil {
		return 0, false
	}

	value, ok := root.Props["__staticItemsCount"]
	if !ok {
		return 0, false
	}

	switch typed := value.(type) {
	case int:
		return typed, true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}

// screenReaderContext threads accessibility-resolution state across recursive
// renderScreenReaderNode calls. The id-map is built once over the *full* tree
// (including hidden / aria-live=off subtrees) so labelledby/describedby still
// find their targets even when those targets are otherwise suppressed. The
// visited set guards against labelledby cycles (a -> b -> a).
type screenReaderContext struct {
	idMap   map[string]*vdom.Node
	visited map[*vdom.Node]bool
}

func newScreenReaderContext(root *vdom.Node) *screenReaderContext {
	ctx := &screenReaderContext{
		idMap:   make(map[string]*vdom.Node),
		visited: make(map[*vdom.Node]bool),
	}
	indexScreenReaderIDs(root, ctx.idMap)
	return ctx
}

// indexScreenReaderIDs walks the entire tree and records the first node it
// sees for each `id` prop. Hidden / aria-live=off subtrees are still indexed
// so labelledby/describedby can resolve into them.
func indexScreenReaderIDs(node *vdom.Node, idMap map[string]*vdom.Node) {
	if node == nil || node.Type != vdom.ElementNode {
		if node != nil {
			for _, child := range node.Children {
				indexScreenReaderIDs(child, idMap)
			}
		}
		return
	}

	if node.Props != nil {
		if id, _ := node.Props["id"].(string); id != "" {
			if _, exists := idMap[id]; !exists {
				idMap[id] = node
			}
		}
	}

	for _, child := range node.Children {
		indexScreenReaderIDs(child, idMap)
	}
}

func renderScreenReaderNode(node *vdom.Node, parentRole string, skipStatic bool, ctx *screenReaderContext) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case vdom.TextNode:
		return node.Text
	case vdom.ElementNode:
		if skipStatic && node.ElementType == "static" {
			return ""
		}

		if node.Props != nil {
			if hidden, _ := node.Props["aria-hidden"].(bool); hidden {
				return ""
			}

			if display, _ := node.Props["display"].(string); display == "none" {
				return ""
			}

			if live, _ := node.Props["aria-live"].(string); live == "off" {
				return ""
			}
		}

		// aria-labelledby substitutes the node's narration with the joined
		// narration of the referenced nodes. It takes precedence over both
		// aria-label and the host's own children. If resolution yields the
		// empty string (cycle, missing ids) we fall through to the regular
		// label/children narration so authors don't lose their content
		// silently.
		var output string
		if labelledBy := resolveLabelledBy(node, ctx); labelledBy != "" {
			output = labelledBy
		} else {
			switch node.ElementType {
			case "text":
				output = screenReaderLabelOrChildren(node, parentRole, "", ctx)
			case "box", "static":
				output = screenReaderLabelOrChildren(node, parentRole, screenReaderSeparator(node), ctx)
			case "transform":
				if node.Props != nil {
					if label, _ := node.Props["accessibilityLabel"].(string); label != "" {
						output = label
						break
					}
				}

				output = renderScreenReaderChildren(node, parentRole, "", ctx)
			default:
				output = renderScreenReaderChildren(node, parentRole, "", ctx)
			}
		}

		// aria-describedby appends its resolved description AFTER the host's
		// regular narration but BEFORE the role decoration is applied, so the
		// announcer reads "<role>: <content> <description>".
		if description := resolveDescribedBy(node, ctx); description != "" {
			if output == "" {
				output = description
			} else {
				output = output + " " + description
			}
		}

		return applyScreenReaderAccessibility(node, output, parentRole)
	default:
		return ""
	}
}

// resolveLabelledBy looks up every space-separated id in `aria-labelledby`,
// renders the referenced node as if it were the labelledby host (skipping
// hidden / aria-live=off suppression so labels in hidden subtrees are still
// usable), and joins the non-empty results with a single space. Self-
// references and missing ids contribute nothing; the visited set prevents
// infinite recursion through cycles.
func resolveLabelledBy(node *vdom.Node, ctx *screenReaderContext) string {
	if node == nil || node.Props == nil || ctx == nil {
		return ""
	}

	value, _ := node.Props["aria-labelledby"].(string)
	if value == "" {
		return ""
	}

	if ctx.visited[node] {
		return ""
	}
	ctx.visited[node] = true
	defer delete(ctx.visited, node)

	parts := splitIDList(value)
	pieces := make([]string, 0, len(parts))
	for _, id := range parts {
		target := ctx.idMap[id]
		if target == nil || target == node || ctx.visited[target] {
			continue
		}
		ctx.visited[target] = true
		rendered := renderLabelTarget(target, ctx)
		delete(ctx.visited, target)
		if rendered != "" {
			pieces = append(pieces, rendered)
		}
	}

	return strings.Join(pieces, " ")
}

// resolveDescribedBy mirrors resolveLabelledBy but for `aria-describedby`.
// Spec parity: missing ids are silent (no trailing space, no marker).
func resolveDescribedBy(node *vdom.Node, ctx *screenReaderContext) string {
	if node == nil || node.Props == nil || ctx == nil {
		return ""
	}

	value, _ := node.Props["aria-describedby"].(string)
	if value == "" {
		return ""
	}

	parts := splitIDList(value)
	pieces := make([]string, 0, len(parts))
	for _, id := range parts {
		target := ctx.idMap[id]
		if target == nil || target == node || ctx.visited[target] {
			continue
		}
		ctx.visited[target] = true
		rendered := renderLabelTarget(target, ctx)
		delete(ctx.visited, target)
		if rendered != "" {
			pieces = append(pieces, rendered)
		}
	}

	return strings.Join(pieces, " ")
}

// renderLabelTarget renders a labelledby/describedby target for use as a
// label string. It bypasses the aria-hidden / aria-live=off suppression of
// the *target* itself so hidden labels are still usable, but still respects
// suppression deeper in the target's own subtree (so nested hidden content
// inside a label source stays hidden).
func renderLabelTarget(target *vdom.Node, ctx *screenReaderContext) string {
	if target == nil {
		return ""
	}
	if target.Type == vdom.TextNode {
		return target.Text
	}
	if target.Type != vdom.ElementNode {
		return ""
	}

	// Render children directly so the target's own aria-hidden / aria-live=off
	// flag does not blank out an otherwise valid label source. The renderer
	// still drops nested hidden subtrees through the normal recursive path.
	switch target.ElementType {
	case "text":
		return screenReaderLabelOrChildren(target, "", "", ctx)
	case "box", "static":
		return screenReaderLabelOrChildren(target, "", screenReaderSeparator(target), ctx)
	default:
		return renderScreenReaderChildren(target, "", "", ctx)
	}
}

// splitIDList breaks a whitespace-separated id list into individual ids,
// dropping empty fragments produced by repeated spaces.
func splitIDList(value string) []string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func screenReaderLabelOrChildren(node *vdom.Node, parentRole string, separator string, ctx *screenReaderContext) string {
	if node == nil {
		return ""
	}

	if node.Props != nil {
		if label, _ := node.Props["aria-label"].(string); label != "" {
			return label
		}
	}

	return renderScreenReaderChildren(node, parentRole, separator, ctx)
}

func renderScreenReaderChildren(node *vdom.Node, parentRole string, separator string, ctx *screenReaderContext) string {
	if node == nil || len(node.Children) == 0 {
		return ""
	}

	children := node.Children
	if node.Props != nil {
		if direction, _ := node.Props["flexDirection"].(string); direction == "row-reverse" || direction == "column-reverse" {
			children = reverseChildren(children)
		}
	}

	rendered := make([]string, 0, len(children))
	role := ""
	if node.Props != nil {
		role, _ = node.Props["aria-role"].(string)
	}
	for _, child := range children {
		if child == nil {
			continue
		}

		if node.ElementType == "text" || node.ElementType == "transform" {
			switch {
			case isTextLikeNode(child), child.Type == vdom.TextNode:
				// Allowed text content.
			default:
				continue
			}
		}

		// Upstream only suppresses duplicate role narration against the direct
		// parent role. Neutral wrappers must not leak an ancestor role down to
		// grandchildren, or nested same-role nodes stop being announced.
		output := renderScreenReaderNode(child, role, false, ctx)
		if output != "" {
			rendered = append(rendered, output)
		}
	}

	return strings.Join(rendered, separator)
}

func reverseChildren(children []*vdom.Node) []*vdom.Node {
	reversed := make([]*vdom.Node, len(children))
	for index := range children {
		reversed[len(children)-1-index] = children[index]
	}

	return reversed
}

func isDisplayNoneNode(node *vdom.Node) bool {
	if node == nil || node.Type != vdom.ElementNode || node.Props == nil {
		return false
	}

	display, _ := node.Props["display"].(string)
	return display == "none"
}

func orderedLayoutChildren(vnode *vdom.Node) []*vdom.Node {
	if vnode == nil || len(vnode.Children) == 0 {
		return nil
	}

	children := make([]*vdom.Node, 0, len(vnode.Children))
	for _, child := range vnode.Children {
		if child == nil {
			continue
		}

		if isDisplayNoneNode(child) {
			continue
		}

		children = append(children, child)
	}

	if vnode.Props == nil {
		return children
	}

	direction, _ := vnode.Props["flexDirection"].(string)
	if direction == "row-reverse" || direction == "column-reverse" {
		return reverseChildren(children)
	}

	return children
}

func clearComputedLayout(vnode *vdom.Node) {
	if vnode == nil {
		return
	}

	vnode.Layout = vdom.Layout{}
	for _, child := range vnode.Children {
		clearComputedLayout(child)
	}
}

func screenReaderSeparator(node *vdom.Node) string {
	if node == nil {
		return "\n"
	}

	defaultSeparator := "\n"
	if node.ElementType == "box" {
		defaultSeparator = " "
	}

	if node.Props == nil {
		return defaultSeparator
	}

	if direction, _ := node.Props["flexDirection"].(string); direction != "" {
		if direction == "row" || direction == "row-reverse" {
			return " "
		}

		return "\n"
	}

	return defaultSeparator
}

func applyScreenReaderAccessibility(node *vdom.Node, output string, parentRole string) string {
	if node == nil || node.Props == nil {
		return output
	}

	if state := screenReaderStateDescription(node.Props); state != "" {
		output = "(" + state + ") " + output
	}

	role, _ := node.Props["aria-role"].(string)
	if role != "" && role != parentRole {
		if role == "heading" {
			if level, ok := parseAriaLevel(node.Props["aria-level"]); ok {
				return "heading " + strconv.Itoa(level) + ": " + output
			}
		}
		return role + ": " + output
	}

	return output
}

// screenReaderStateDescription builds the parenthesised state string for a
// node by merging `aria-state` (map / Props / map[string]bool / []string)
// with the top-level shorthand props (`aria-busy`, `aria-checked`, etc.).
// Known keys narrate first in accessibilityStateOrder; unknown keys follow
// alphabetically. The special value `"mixed"` for the `checked` key narrates
// as `mixed` (tri-state checkbox parity).
func screenReaderStateDescription(props vdom.Props) string {
	if props == nil {
		return ""
	}

	if slice, ok := props["aria-state"].([]string); ok {
		return strings.Join(slice, ", ")
	}

	merged := mergeAccessibilityState(props)
	if len(merged) == 0 {
		return ""
	}

	knownPart := make([]string, 0, len(accessibilityStateOrder))
	for _, key := range accessibilityStateOrder {
		if label, present := merged[key]; present {
			knownPart = append(knownPart, label)
		}
	}

	unknownKeys := make([]string, 0, len(merged))
	for key := range merged {
		if _, isKnown := accessibilityKnownStateSet[key]; !isKnown {
			unknownKeys = append(unknownKeys, key)
		}
	}
	sort.Strings(unknownKeys)
	for _, key := range unknownKeys {
		knownPart = append(knownPart, merged[key])
	}

	return strings.Join(knownPart, ", ")
}

// mergeAccessibilityState collects the truthy state keys from both the
// `aria-state` prop and the top-level shorthand props, returning a map of
// state-key -> announced-label. The announced label is normally the key
// itself, but `checked == "mixed"` substitutes the literal "mixed" so the
// tri-state narration reads "(mixed)" rather than "(checked)". Explicit
// values from `aria-state` take precedence over shorthand on collision.
func mergeAccessibilityState(props vdom.Props) map[string]string {
	merged := make(map[string]string)

	// First, fold in the top-level shorthand props. These lose to aria-state
	// on collision because they're written in a second pass below.
	for _, shortKey := range accessibilityShorthandKeys {
		raw, present := props[shortKey]
		if !present {
			continue
		}
		stateKey := strings.TrimPrefix(shortKey, "aria-")
		label, ok := accessibilityStateLabel(stateKey, raw)
		if !ok {
			continue
		}
		merged[stateKey] = label
	}

	// Then fold in aria-state. Anything declared here overrides the shorthand
	// and also lets authors use arbitrary keys (e.g. "invalid").
	if state, present := props["aria-state"]; present {
		switch typed := state.(type) {
		case map[string]interface{}:
			for key, raw := range typed {
				if label, ok := accessibilityStateLabel(key, raw); ok {
					merged[key] = label
				} else {
					delete(merged, key)
				}
			}
		case vdom.Props:
			for key, raw := range typed {
				if label, ok := accessibilityStateLabel(key, raw); ok {
					merged[key] = label
				} else {
					delete(merged, key)
				}
			}
		case map[string]bool:
			for key, raw := range typed {
				if label, ok := accessibilityStateLabel(key, raw); ok {
					merged[key] = label
				} else {
					delete(merged, key)
				}
			}
		case string:
			if typed != "" {
				merged[typed] = typed
			}
		}
	}

	return merged
}

// accessibilityStateLabel decides whether a state key is announced and what
// label to use. Truthy values announce the key (e.g. `selected: true` ->
// "selected"). The special string `"mixed"` for the `checked` key swaps the
// announcement to "mixed" for tri-state parity.
func accessibilityStateLabel(key string, value interface{}) (string, bool) {
	if !accessibilityValueTruthy(value) {
		return "", false
	}

	if key == "checked" {
		if str, ok := value.(string); ok && str == "mixed" {
			return "mixed", true
		}
	}

	return key, true
}

// accessibilityValueTruthy returns true for the values upstream Ink's
// `Object.keys(state).filter(key => state[key])` would treat as truthy.
// Numbers count as truthy when non-zero so JSON-decoded shorthand values
// behave the same as their boolean equivalents.
func accessibilityValueTruthy(value interface{}) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case bool:
		return typed
	case string:
		return typed != ""
	case int:
		return typed != 0
	case int64:
		return typed != 0
	case float32:
		return typed != 0
	case float64:
		return typed != 0
	default:
		return true
	}
}

// parseAriaLevel reads aria-level as int / int64 / float32 / float64 (the
// JSON-decoded form arriving from the parity harness) and returns the value
// as an int. Non-numeric values yield ok=false so the caller can omit the
// level prefix.
func parseAriaLevel(value interface{}) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case float32:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}

// renderAnnouncerRegions walks the tree and collects every aria-live="polite"
// or "assertive" subtree, returning a block of one prefixed line per region.
// Assertive regions are emitted before polite to reflect higher urgency.
// aria-hidden / display:none / aria-live=off subtrees are skipped during the
// walk; the polite/assertive subtree contents are still rendered through the
// regular renderScreenReaderNode pipeline so labelledby / describedby and
// aria-state continue to apply inside the region.
func renderAnnouncerRegions(node *vdom.Node, ctx *screenReaderContext) string {
	if node == nil {
		return ""
	}

	var assertive, polite []string
	collectAnnouncerRegions(node, ctx, &assertive, &polite)

	if len(assertive) == 0 && len(polite) == 0 {
		return ""
	}

	pieces := make([]string, 0, len(assertive)+len(polite))
	for _, region := range assertive {
		pieces = append(pieces, "[assertive] "+region)
	}
	for _, region := range polite {
		pieces = append(pieces, "[polite] "+region)
	}
	return strings.Join(pieces, "\n")
}

func collectAnnouncerRegions(node *vdom.Node, ctx *screenReaderContext, assertive, polite *[]string) {
	if node == nil || node.Type != vdom.ElementNode {
		return
	}

	if node.Props != nil {
		if hidden, _ := node.Props["aria-hidden"].(bool); hidden {
			return
		}
		if display, _ := node.Props["display"].(string); display == "none" {
			return
		}
		if live, _ := node.Props["aria-live"].(string); live != "" {
			switch live {
			case "off":
				return
			case "polite", "assertive":
				rendered := renderScreenReaderNode(node, "", false, ctx)
				if rendered != "" {
					if live == "assertive" {
						*assertive = append(*assertive, rendered)
					} else {
						*polite = append(*polite, rendered)
					}
				}
				// Do not descend further: the rendered subtree already
				// captured the polite/assertive content. Nested aria-live
				// regions inside a region are unusual and would just
				// duplicate; mirroring upstream's "outermost wins" intent.
				return
			}
		}
	}

	for _, child := range node.Children {
		collectAnnouncerRegions(child, ctx, assertive, polite)
	}
}

// buildLayoutTree creates a layout tree from a vdom tree
func childWidthConstraint(vnode *vdom.Node, inheritedWidth int) int {
	if vnode == nil {
		return inheritedWidth
	}

	width := inheritedWidth
	if vnode.Props != nil {
		if explicitWidth, percent, ok := parseSizeValue(vnode.Props["width"]); ok {
			if percent {
				width = int(math.Round(float64(width) * explicitWidth / 100))
			} else if explicitWidth > 0 {
				width = int(explicitWidth)
			}
		}

		// Resolve padding with Yoga precedence: per-edge overrides shorthand.
		paddingLeft := 0
		paddingRight := 0
		if padding, ok := parseNumericValue(vnode.Props["padding"]); ok {
			paddingLeft = int(padding)
			paddingRight = int(padding)
		}
		if paddingX, ok := parseNumericValue(vnode.Props["paddingX"]); ok {
			paddingLeft = int(paddingX)
			paddingRight = int(paddingX)
		}
		if left, ok := parseNumericValue(vnode.Props["paddingLeft"]); ok {
			paddingLeft = int(left)
		}
		if right, ok := parseNumericValue(vnode.Props["paddingRight"]); ok {
			paddingRight = int(right)
		}
		borderLeft, _, borderRight, _ := borderInsets(vnode.Props)
		width -= paddingLeft + paddingRight + int(borderLeft) + int(borderRight)
	}

	if width < 0 {
		return 0
	}

	return width
}

func buildLayoutTree(vnode *vdom.Node) *layout.Node {
	return buildLayoutTreeWithWidth(vnode, 0)
}

func buildLayoutTreeWithWidth(vnode *vdom.Node, availableWidth int) *layout.Node {
	node := layout.NewNode()
	if vnode == nil {
		return node
	}

	if availableWidth > 0 {
		node.SetWidthHint(float64(availableWidth))
	}

	if isDisplayNoneNode(vnode) {
		return node
	}

	if isLayoutContainerNode(vnode) {
		defaultDirection := layout.FlexDirectionRow
		if vnode.ElementType == "static" {
			defaultDirection = layout.FlexDirectionColumn
		}

		applyContainerLayoutProps(node, vnode.Props, defaultDirection)

		// Build layout tree for children
		childWidth := childWidthConstraint(vnode, availableWidth)
		for _, child := range orderedLayoutChildren(vnode) {
			childLayout := buildLayoutTreeWithWidth(child, childWidth)
			node.AddChild(childLayout)
		}
	} else if isTextLikeNode(vnode) {
		text := measureTextLikeNode(vnode, availableWidth)
		width, height := measureTextBlock(text)
		node.SetWidth(width)
		node.SetHeight(height)
		node.SetTextLike(true)
		node.SetFlexShrink(1)
		if vnode.Props != nil {
			if flexShrink, ok := parseNumericValue(vnode.Props["flexShrink"]); ok {
				node.SetFlexShrink(flexShrink)
			}
		}
		node.SetMeasureSizeFunc(func(width float64) (float64, float64) {
			// textMeasurementWidth floors the inbound width to the integer wrap
			// budget that applyTextLayoutMode (wrap-ansi parity) will actually
			// use. Upstream Yoga's measureFunc returns these floored
			// dimensions, so the renderer can later wrap at exactly the same
			// budget instead of the parent's allotted (rounded-up) width.
			wrapWidth := textMeasurementWidth(width)
			text := measureTextLikeNode(vnode, wrapWidth)
			measuredW, measuredH := measureTextBlock(text)
			return measuredW, measuredH
		})
	} else if vnode.Type == vdom.TextNode {
		// Text nodes have size based on their content
		width, height := measureTextBlock(vnode.Text)
		node.SetWidth(width)
		node.SetHeight(height)
	} else {
		// Other element types
		for _, child := range orderedLayoutChildren(vnode) {
			childLayout := buildLayoutTreeWithWidth(child, availableWidth)
			node.AddChild(childLayout)
		}
	}

	return node
}

func syncComputedLayout(vnode *vdom.Node, layoutNode *layout.Node) {
	if vnode == nil || layoutNode == nil {
		return
	}

	if isDisplayNoneNode(vnode) {
		clearComputedLayout(vnode)
		return
	}

	vnode.Layout = vdom.Layout{
		Left:   roundLayoutValue(layoutNode.GetComputedLeft()),
		Top:    roundLayoutValue(layoutNode.GetComputedTop()),
		Width:  roundLayoutValue(layoutNode.GetComputedWidth()),
		Height: roundLayoutValue(layoutNode.GetComputedHeight()),
	}

	if vnode.Type == vdom.ElementNode && vnode.Props != nil {
		if ref, ok := vnode.Props["ref"].(refSetter); ok && ref != nil {
			ref.SetCurrent(vnode)
		}
	}

	orderedChildren := orderedLayoutChildren(vnode)
	for childIndex, child := range orderedChildren {
		if childIndex < layoutNode.GetChildCount() {
			syncComputedLayout(child, layoutNode.GetChild(childIndex))
		}
	}

	for _, child := range vnode.Children {
		if isDisplayNoneNode(child) {
			clearComputedLayout(child)
		}
	}
}

// renderNodeWithLayout renders a node using its layout position
func renderNodeWithLayout(buf *buffer.Buffer, vnode *vdom.Node, layoutNode *layout.Node, inheritedBackground string, enableANSI bool, availableWidth int, clip clipRect) {
	if vnode == nil || layoutNode == nil {
		return
	}

	if isDisplayNoneNode(vnode) {
		return
	}

	x := roundLayoutValue(layoutNode.GetComputedLeft())
	y := roundLayoutValue(layoutNode.GetComputedTop())
	currentClip := nodeClipRect(vnode, layoutNode, clip)

	switch vnode.Type {
	case vdom.TextNode:
		// Render text at computed position
		writeStringClipped(buf, x, y, vnode.Text, currentClip)

	case vdom.ElementNode:
		currentBackground := getInheritedBackground(vnode, inheritedBackground)
		if isTextLikeNode(vnode) {
			textWidth := availableWidth
			if layoutNode.HasAdjustedSize() {
				textWidth = int(math.Round(layoutNode.GetComputedWidth()))
			}
			if textWidth <= 0 {
				textWidth = int(math.Round(layoutNode.GetComputedWidth()))
			}
			// When an ancestor flex container shrunk this text node's parent
			// box, the floored measure-time wrap budget can be narrower than
			// the post-rounding computedWidth. Honor it in that case so the
			// renderer wraps at the same budget upstream Ink uses via
			// `getMaxWidth(yogaNode)`. Gated on parent.sizeAdjusted to avoid
			// over-wrapping the directly-shrunk text-pair case where ceil
			// rounding is intentional for sibling overlap of trailing space.
			if layoutNode.ShouldHonorMeasuredWidth() {
				measured := int(layoutNode.GetMeasuredWidth())
				if measured > 0 && measured < textWidth {
					textWidth = measured
				}
			}
			writeStringClipped(buf, x, y, renderTextLikeNode(vnode, inheritedBackground, enableANSI, textWidth), currentClip)
			return
		}

		if vnode.ElementType == "box" {
			drawBoxBorder(buf, vnode, layoutNode, currentClip)
		}

		// Render children with their layout. Ink draws borders before
		// descendants, so negative-margin children can intentionally overlap
		// them unless overflow clipping is enabled.
		childAvailableWidth := childWidthConstraint(vnode, availableWidth)
		childClip := nodeChildClipRect(vnode, layoutNode, currentClip)
		for i, child := range orderedLayoutChildren(vnode) {
			if i < layoutNode.GetChildCount() {
				childLayout := layoutNode.GetChild(i)
				renderNodeWithLayout(buf, child, childLayout, currentBackground, enableANSI, childAvailableWidth, childClip)
			}
		}
	}
}

func renderNodeWithLayoutANSI(canvas *ansiCanvas, vnode *vdom.Node, layoutNode *layout.Node, inheritedBackground string, availableWidth int, clip clipRect) {
	if vnode == nil || layoutNode == nil {
		return
	}

	if isDisplayNoneNode(vnode) {
		return
	}

	x := roundLayoutValue(layoutNode.GetComputedLeft())
	y := roundLayoutValue(layoutNode.GetComputedTop())
	currentClip := nodeClipRect(vnode, layoutNode, clip)

	switch vnode.Type {
	case vdom.TextNode:
		writeStyledStringClipped(canvas, x, y, vnode.Text, ansiStyle{}, currentClip)

	case vdom.ElementNode:
		currentBackground := getInheritedBackground(vnode, inheritedBackground)
		if isTextLikeNode(vnode) {
			textWidth := availableWidth
			if layoutNode.HasAdjustedSize() {
				textWidth = int(math.Round(layoutNode.GetComputedWidth()))
			}
			if textWidth <= 0 {
				textWidth = int(math.Round(layoutNode.GetComputedWidth()))
			}
			// See note in renderNodeWithLayout above.
			if layoutNode.ShouldHonorMeasuredWidth() {
				measured := int(layoutNode.GetMeasuredWidth())
				if measured > 0 && measured < textWidth {
					textWidth = measured
				}
			}

			if styledLines, ok := styledTextLines(vnode, inheritedBackground, textWidth); ok {
				writeStyledLinesClipped(canvas, x, y, styledLines, currentClip)
				return
			}

			text := renderTextLikeNode(vnode, "", false, textWidth)
			style := resolveTextStyle(vnode, inheritedBackground)
			writeStyledStringClipped(canvas, x, y, text, style, currentClip)
			return
		}

		if vnode.ElementType == "box" {
			fillBoxBackgroundANSI(canvas, vnode, layoutNode, inheritedBackground, currentClip)
			drawBoxBorderANSI(canvas, vnode, layoutNode, currentClip)
		}

		childAvailableWidth := childWidthConstraint(vnode, availableWidth)
		childClip := nodeChildClipRect(vnode, layoutNode, currentClip)
		for index, child := range orderedLayoutChildren(vnode) {
			if index < layoutNode.GetChildCount() {
				childLayout := layoutNode.GetChild(index)
				renderNodeWithLayoutANSI(canvas, child, childLayout, currentBackground, childAvailableWidth, childClip)
			}
		}
	}
}

// renderNode recursively renders a node and its children (simple version)
func renderNode(buf *buffer.Buffer, node *vdom.Node, x, y int, inheritedBackground string, enableANSI bool) (nextX, nextY int) {
	if node == nil {
		return x, y
	}

	switch node.Type {
	case vdom.TextNode:
		// Render text
		buf.WriteString(x, y, node.Text)
		// Text advances X position
		return x + len(node.Text), y

	case vdom.ElementNode:
		currentBackground := getInheritedBackground(node, inheritedBackground)
		if isTextLikeNode(node) {
			text := renderTextLikeNode(node, inheritedBackground, enableANSI, 0)
			buf.WriteString(x, y, text)
			return x + len(text), y
		}

		// Render children with their layout
		// For now, just render children in sequence
		currentX, currentY := x, y
		for _, child := range node.Children {
			currentX, currentY = renderNode(buf, child, currentX, currentY, currentBackground, enableANSI)
		}
		return currentX, currentY

	default:
		return x, y
	}
}
