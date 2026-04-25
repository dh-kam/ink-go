package renderer

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dh-kam/goink.go/internal/buffer"
	"github.com/dh-kam/goink.go/pkg/layout"
	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/utils"
	"github.com/dh-kam/goink.go/pkg/vdom"
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

type ansiCell struct {
	text         string
	style        ansiStyle
	continuation bool
}

type plainCell struct {
	text         string
	visible      string
	continuation bool
}

type styledRune struct {
	ch    rune
	style ansiStyle
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
	if c == nil || x < 0 || x >= c.width || y < 0 || y >= c.height {
		return
	}

	width := utils.RuneWidth(ch)
	if width == 0 {
		c.appendZeroWidth(x, y, ch)
		return
	}

	c.cells[y][x] = ansiCell{
		text:  string(ch),
		style: style,
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
		if cell.continuation || cell.text != " " || cell.style.fg != "" || cell.style.bg != "" {
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
	"multiline",
	"multiselectable",
	"readonly",
	"required",
	"selected",
}

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

	width := int(layoutNode.GetComputedWidth())
	height := int(layoutNode.GetComputedHeight())
	if width <= 0 || height <= 0 {
		return
	}

	x := int(layoutNode.GetComputedLeft())
	y := int(layoutNode.GetComputedTop())

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

	width := int(layoutNode.GetComputedWidth())
	height := int(layoutNode.GetComputedHeight())
	if width <= 0 || height <= 0 {
		return
	}

	x := int(layoutNode.GetComputedLeft())
	y := int(layoutNode.GetComputedTop())

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

	width := int(layoutNode.GetComputedWidth())
	height := int(layoutNode.GetComputedHeight())
	if width <= 0 || height <= 0 {
		return
	}

	topStyle := resolveBorderSideStyle(vnode.Props, "borderTopColor", "borderTopDimColor")
	bottomStyle := resolveBorderSideStyle(vnode.Props, "borderBottomColor", "borderBottomDimColor")
	leftStyle := resolveBorderSideStyle(vnode.Props, "borderLeftColor", "borderLeftDimColor")
	rightStyle := resolveBorderSideStyle(vnode.Props, "borderRightColor", "borderRightDimColor")
	x := int(layoutNode.GetComputedLeft())
	y := int(layoutNode.GetComputedTop())

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
	x := int(layoutNode.GetComputedLeft()) + int(leftInset)
	y := int(layoutNode.GetComputedTop()) + int(topInset)
	width := int(layoutNode.GetComputedWidth()) - int(leftInset) - int(rightInset)
	height := int(layoutNode.GetComputedHeight()) - int(topInset) - int(bottomInset)
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

	left := int(layoutNode.GetComputedLeft())
	top := int(layoutNode.GetComputedTop())
	right := left + int(layoutNode.GetComputedWidth())
	bottom := top + int(layoutNode.GetComputedHeight())

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
				case 0x07:
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
				buf.Set(currentX, currentY, ch)
			}
			continue
		}

		if currentY >= clip.top && currentY < clip.bottom && currentX < clip.right && currentX+width > clip.left {
			buf.Set(currentX, currentY, ch)
		}

		currentX += width
	}
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

	return transform(text, index)
}

func applyLineTransform(node *vdom.Node, text string) string {
	if node == nil || node.Props == nil || text == "" {
		return text
	}

	transform, ok := node.Props["transform"].(func(string, int) string)
	if !ok || transform == nil {
		return text
	}

	lines := strings.Split(text, "\n")
	for lineIndex, line := range lines {
		lines[lineIndex] = transform(line, lineIndex)
	}

	return strings.Join(lines, "\n")
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

	runes := []rune(line)
	return string(runes[:fitRunesToWidth(runes, maxWidth-1)]) + "…"
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

	runes := []rune(line)
	keepWidth := maxWidth - 1
	start := len(runes)
	width := 0
	for start > 0 {
		runeWidth := utils.RuneWidth(runes[start-1])
		if width+runeWidth > keepWidth {
			break
		}
		start--
		width += runeWidth
	}

	return "…" + string(runes[start:])
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
	left := []rune(line)
	right := []rune(line)

	left = left[:fitRunesToWidth(left, leftWidth)]

	start := len(right)
	width := 0
	for start > 0 {
		runeWidth := utils.RuneWidth(right[start-1])
		if width+runeWidth > rightWidth {
			break
		}
		start--
		width += runeWidth
	}

	return string(left) + "…" + string(right[start:])
}

func wrapLine(line string, maxWidth int) []string {
	if maxWidth <= 0 || utils.StringWidth(line) <= maxWidth {
		return []string{line}
	}

	runes := []rune(line)
	lines := make([]string, 0, 2)
	for len(runes) > 0 {
		if utils.StringWidth(string(runes)) <= maxWidth {
			lines = append(lines, string(runes))
			break
		}

		fit := fitRunesToWidth(runes, maxWidth)
		breakIndex := -1
		seenNonSpace := false
		for index := 0; index < fit; index++ {
			if !unicode.IsSpace(runes[index]) {
				seenNonSpace = true
			}
			if seenNonSpace && unicode.IsSpace(runes[index]) {
				breakIndex = index
			}
		}

		if breakIndex > 0 {
			segment := string(trimRightSpaceRunes(runes[:breakIndex]))
			runes = trimLeftSpaceRunes(runes[breakIndex:])
			if segment != "" {
				lines = append(lines, segment)
				continue
			}
		}

		segment := string(runes[:fit])
		lines = append(lines, segment)
		runes = runes[fit:]
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

func applyTextLayoutMode(text string, mode string, maxWidth int) string {
	if text == "" || maxWidth <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	processed := make([]string, 0, len(lines))
	for _, line := range lines {
		switch mode {
		case "truncate":
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
		builder.WriteRune(r.ch)
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
			builder.WriteRune('\n')
			continue
		}

		emitANSITransition(&builder, current, r.style)
		current = r.style
		builder.WriteRune(r.ch)
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

	for index := 0; index < len(text); {
		if text[index] != 0x1b || index+1 >= len(text) || text[index+1] != '[' {
			r, size := rune(text[index]), 1
			if text[index] >= 0x80 {
				r, size = utf8.DecodeRuneInString(text[index:])
			}
			result = append(result, styledRune{ch: r, style: current})
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
		runes := []rune(node.Text)
		collected := make([]styledRune, 0, len(runes))
		for _, r := range runes {
			collected = append(collected, styledRune{ch: r, style: inheritedStyle})
		}

		return collected, true
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
			segment := trimRightSpaceStyledRunes(runes[:breakIndex])
			runes = trimLeftSpaceStyledRunes(runes[breakIndex:])
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
		case "truncate":
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
				if column > clip.left && column <= clip.right {
					canvas.setCell(column, row, styled.ch, styled.style)
				}
				continue
			}

			if column >= clip.left && column < clip.right {
				canvas.setCell(column, row, styled.ch, styled.style)
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

	paddingLeft, paddingTop, paddingRight, paddingBottom := borderInsets(props)
	if p, ok := parseNumericValue(props["padding"]); ok {
		paddingLeft += p
		paddingTop += p
		paddingRight += p
		paddingBottom += p
	}
	if p, ok := parseNumericValue(props["paddingX"]); ok {
		paddingLeft += p
		paddingRight += p
	}
	if p, ok := parseNumericValue(props["paddingY"]); ok {
		paddingTop += p
		paddingBottom += p
	}
	if p, ok := parseNumericValue(props["paddingLeft"]); ok {
		paddingLeft += p
	}
	if p, ok := parseNumericValue(props["paddingTop"]); ok {
		paddingTop += p
	}
	if p, ok := parseNumericValue(props["paddingRight"]); ok {
		paddingRight += p
	}
	if p, ok := parseNumericValue(props["paddingBottom"]); ok {
		paddingBottom += p
	}
	node.SetPadding(layout.EdgeLeft, paddingLeft)
	node.SetPadding(layout.EdgeTop, paddingTop)
	node.SetPadding(layout.EdgeRight, paddingRight)
	node.SetPadding(layout.EdgeBottom, paddingBottom)

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

	var staticBuilder strings.Builder
	for _, staticRoot := range staticRoots {
		staticBuilder.WriteString(renderWithLayoutMode(staticRoot, width, height, enableANSI))
	}

	sections.StaticOutput = staticBuilder.String()
	return sections
}

// RenderScreenReaderSections renders a tree as plain accessibility text.
func RenderScreenReaderSections(node *vdom.Node) RenderSections {
	if node == nil {
		return RenderSections{}
	}

	staticRoots := collectStaticRoots(node)
	mainNode := cloneWithoutStatic(node)

	sections := RenderSections{}
	if mainNode != nil {
		sections.Output = renderScreenReaderNode(mainNode, "", false)
	}

	if len(staticRoots) == 0 {
		return sections
	}

	var staticBuilder strings.Builder
	for _, staticRoot := range staticRoots {
		rendered := renderScreenReaderNode(staticRoot, "", false)
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
			sections.Output = renderScreenReaderNode(mainNode, "", false)
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

	cloned.Children = filteredChildren
	return cloned
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
			return renderScreenReaderNode(root, "", false), 1
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
		return renderScreenReaderNode(deltaRoot, "", false), count
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

func renderScreenReaderNode(node *vdom.Node, parentRole string, skipStatic bool) string {
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
		}

		output := ""
		switch node.ElementType {
		case "text":
			output = screenReaderLabelOrChildren(node, parentRole, "")
		case "box", "static":
			output = screenReaderLabelOrChildren(node, parentRole, screenReaderSeparator(node))
		case "transform":
			if node.Props != nil {
				if label, _ := node.Props["accessibilityLabel"].(string); label != "" {
					output = label
					break
				}
			}

			output = renderScreenReaderChildren(node, parentRole, "")
		default:
			output = renderScreenReaderChildren(node, parentRole, "")
		}

		return applyScreenReaderAccessibility(node, output, parentRole)
	default:
		return ""
	}
}

func screenReaderLabelOrChildren(node *vdom.Node, parentRole string, separator string) string {
	if node == nil {
		return ""
	}

	if node.Props != nil {
		if label, _ := node.Props["aria-label"].(string); label != "" {
			return label
		}
	}

	return renderScreenReaderChildren(node, parentRole, separator)
}

func renderScreenReaderChildren(node *vdom.Node, parentRole string, separator string) string {
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
	role, _ := node.Props["aria-role"].(string)
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
		output := renderScreenReaderNode(child, role, false)
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

	if state := screenReaderStateDescription(node.Props["aria-state"]); state != "" {
		output = "(" + state + ") " + output
	}

	role, _ := node.Props["aria-role"].(string)
	if role != "" && role != parentRole {
		return role + ": " + output
	}

	return output
}

func screenReaderStateDescription(value interface{}) string {
	if value == nil {
		return ""
	}

	switch typed := value.(type) {
	case []string:
		return strings.Join(typed, ", ")
	}

	states := make([]string, 0, len(accessibilityStateOrder))
	for _, key := range accessibilityStateOrder {
		enabled, ok := accessibilityStateEnabled(value, key)
		if ok && enabled {
			states = append(states, key)
		}
	}

	return strings.Join(states, ", ")
}

func accessibilityStateEnabled(value interface{}, key string) (bool, bool) {
	switch typed := value.(type) {
	case map[string]interface{}:
		enabled, ok := typed[key].(bool)
		return enabled, ok
	case vdom.Props:
		enabled, ok := typed[key].(bool)
		return enabled, ok
	case map[string]bool:
		enabled, ok := typed[key]
		return enabled, ok
	default:
		return false, false
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

		paddingLeft := 0
		paddingRight := 0
		if padding, ok := parseNumericValue(vnode.Props["padding"]); ok {
			paddingLeft += int(padding)
			paddingRight += int(padding)
		}
		if paddingX, ok := parseNumericValue(vnode.Props["paddingX"]); ok {
			paddingLeft += int(paddingX)
			paddingRight += int(paddingX)
		}
		if left, ok := parseNumericValue(vnode.Props["paddingLeft"]); ok {
			paddingLeft += int(left)
		}
		if right, ok := parseNumericValue(vnode.Props["paddingRight"]); ok {
			paddingRight += int(right)
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
		node.SetMeasureHeightFunc(func(width float64) float64 {
			text := measureTextLikeNode(vnode, textMeasurementWidth(width))
			_, measuredHeight := measureTextBlock(text)
			return measuredHeight
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
		Left:   int(layoutNode.GetComputedLeft()),
		Top:    int(layoutNode.GetComputedTop()),
		Width:  int(layoutNode.GetComputedWidth()),
		Height: int(layoutNode.GetComputedHeight()),
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

	x := int(layoutNode.GetComputedLeft())
	y := int(layoutNode.GetComputedTop())
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
			writeStringClipped(buf, x, y, renderTextLikeNode(vnode, inheritedBackground, enableANSI, textWidth), currentClip)
			return
		}

		// Render children with their layout
		childAvailableWidth := childWidthConstraint(vnode, availableWidth)
		for i, child := range orderedLayoutChildren(vnode) {
			if i < layoutNode.GetChildCount() {
				childLayout := layoutNode.GetChild(i)
				renderNodeWithLayout(buf, child, childLayout, currentBackground, enableANSI, childAvailableWidth, currentClip)
			}
		}

		if vnode.ElementType == "box" {
			drawBoxBorder(buf, vnode, layoutNode, currentClip)
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

	x := int(layoutNode.GetComputedLeft())
	y := int(layoutNode.GetComputedTop())
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
		}

		childAvailableWidth := childWidthConstraint(vnode, availableWidth)
		for index, child := range orderedLayoutChildren(vnode) {
			if index < layoutNode.GetChildCount() {
				childLayout := layoutNode.GetChild(index)
				renderNodeWithLayoutANSI(canvas, child, childLayout, currentBackground, childAvailableWidth, currentClip)
			}
		}

		if vnode.ElementType == "box" {
			drawBoxBorderANSI(canvas, vnode, layoutNode, currentClip)
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
