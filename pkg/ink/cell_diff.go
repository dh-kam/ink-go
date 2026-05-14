package ink

import (
	"strings"

	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/utils"
)

// cellDiffResetSGR resets all active SGR attributes — emitted whenever the
// cursor state needs to be cleared between styled runs of changed cells.
const cellDiffResetSGR = "\x1b[0m"

// sgrAttrs captures the visible-state attributes encoded in a concatenated
// SGR sequence string (e.g. "\x1b[31m\x1b[1m"). The fg/bg fields hold the
// raw enable sequence (e.g. "\x1b[31m" or "\x1b[38;5;208m") so that we can
// re-emit them verbatim — this avoids any color-mapping ambiguity. The
// boolean flags track the on/off state of toggle attributes.
//
// parseOK is false when the input contains a code we don't model (and we
// must therefore use the safe "\x1b[0m + full SGR" fallback emission).
type sgrAttrs struct {
	fg            string
	bg            string
	bold          bool
	dim           bool
	italic        bool
	underline     bool
	inverse       bool
	strikethrough bool
	parseOK       bool
}

// parseSGR walks a concatenated SGR string and returns its accumulated
// attribute state. The string must consist solely of CSI 'm' sequences —
// callers obtain it from parseLine which already enforces that invariant.
//
// We support 0 (reset), 1/2/3/4/7/9 enable, 22/23/24/27/29 disable, and the
// foreground/background color spaces (30-37, 39, 40-47, 49, 90-97, 100-107,
// 38;5;n, 48;5;n, 38;2;r;g;b, 48;2;r;g;b). Unknown numeric codes set
// parseOK = false so the caller can fall back to the conservative path.
func parseSGR(seq string) sgrAttrs {
	attrs := sgrAttrs{parseOK: true}
	if seq == "" {
		return attrs
	}

	for index := 0; index < len(seq); {
		if seq[index] != 0x1b {
			attrs.parseOK = false
			return attrs
		}
		if index+1 >= len(seq) || seq[index+1] != '[' {
			attrs.parseOK = false
			return attrs
		}
		end := index + 2
		for end < len(seq) && seq[end] != 'm' {
			end++
		}
		if end >= len(seq) {
			attrs.parseOK = false
			return attrs
		}
		body := seq[index+2 : end]
		fullSeq := seq[index : end+1]
		index = end + 1

		// Empty body ("\x1b[m") is equivalent to reset.
		if body == "" {
			attrs = sgrAttrs{parseOK: true}
			continue
		}

		params := strings.Split(body, ";")
		i := 0
		for i < len(params) {
			code := params[i]
			switch code {
			case "0":
				attrs = sgrAttrs{parseOK: true}
			case "1":
				attrs.bold = true
			case "2":
				attrs.dim = true
			case "3":
				attrs.italic = true
			case "4":
				attrs.underline = true
			case "7":
				attrs.inverse = true
			case "9":
				attrs.strikethrough = true
			case "22":
				attrs.bold = false
				attrs.dim = false
			case "23":
				attrs.italic = false
			case "24":
				attrs.underline = false
			case "27":
				attrs.inverse = false
			case "29":
				attrs.strikethrough = false
			case "39":
				attrs.fg = ""
			case "49":
				attrs.bg = ""
			case "30", "31", "32", "33", "34", "35", "36", "37",
				"90", "91", "92", "93", "94", "95", "96", "97":
				attrs.fg = "\x1b[" + code + "m"
			case "40", "41", "42", "43", "44", "45", "46", "47",
				"100", "101", "102", "103", "104", "105", "106", "107":
				attrs.bg = "\x1b[" + code + "m"
			case "38", "48":
				// Extended color spaces. Either 38;5;n / 48;5;n (ANSI-256)
				// or 38;2;r;g;b / 48;2;r;g;b (truecolor). We stash the
				// whole originating sequence as the fg/bg payload and skip
				// the consumed sub-parameters.
				if i+1 >= len(params) {
					attrs.parseOK = false
					return attrs
				}
				switch params[i+1] {
				case "5":
					if code == "38" {
						attrs.fg = fullSeq
					} else {
						attrs.bg = fullSeq
					}
					i = len(params)
					continue
				case "2":
					if code == "38" {
						attrs.fg = fullSeq
					} else {
						attrs.bg = fullSeq
					}
					i = len(params)
					continue
				default:
					attrs.parseOK = false
					return attrs
				}
			default:
				attrs.parseOK = false
				return attrs
			}
			i++
		}
	}

	return attrs
}

// sgrTransition computes the minimal byte sequence that takes the terminal
// from current SGR state to next SGR state. It mirrors the strategy used by
// internal/renderer's emitANSITransition: prefer targeted enable/disable
// codes over a full reset whenever the transition only adds attributes or
// flips colors.
//
// When a toggle attribute needs to be turned off (e.g. bold off), the
// terminal command 22 also disables dim — so we emit 22 then re-enable any
// remaining bold/dim that are still active. This is the same logic as
// emitANSITransition.
//
// When parseOK is false on either side, the caller must avoid this
// optimization and emit the safe "\x1b[0m + raw SGR string" fallback.
func sgrTransition(current, next sgrAttrs) string {
	if !current.parseOK || !next.parseOK {
		return ""
	}

	if current == next {
		return ""
	}

	// Fast path: returning to a fully-plain state is often shortest with a
	// single \x1b[0m reset. We compute the targeted form and pick whichever
	// is shorter (preferring \x1b[0m on tie, since it leaves no residual
	// state that downstream emissions could accidentally inherit).
	clean := sgrAttrs{parseOK: true}
	if next == clean {
		targeted := sgrTransitionTargeted(current, next)
		if len(targeted) >= len(cellDiffResetSGR) {
			return cellDiffResetSGR
		}
		return targeted
	}

	return sgrTransitionTargeted(current, next)
}

// sgrTransitionTargeted is the per-attribute disable/enable emission used
// by sgrTransition. Splitting it out lets the to-plain fast path compare
// against the full reset form.
func sgrTransitionTargeted(current, next sgrAttrs) string {
	var builder strings.Builder

	// If any attribute that's currently on must be turned off, and the
	// targeted disable codes aren't enough to express that change cleanly,
	// reset and re-emit. This handles the "disabling bold mid-color" corner
	// case: 22 disables both bold and dim, so if current has both and next
	// has only one, we reset and re-enable to stay safe.
	if needsResetForTransition(current, next) {
		builder.WriteString(cellDiffResetSGR)
		writeEnableSGR(&builder, next)
		return builder.String()
	}

	// Disable bold/dim with 22 if either flips off; then re-enable the
	// survivor. This matches emitANSITransition's behavior.
	if (current.bold && !next.bold) || (current.dim && !next.dim) {
		builder.WriteString("\x1b[22m")
		if next.bold {
			builder.WriteString(styles.BoldCode())
		}
		if next.dim {
			builder.WriteString(styles.DimCode())
		}
	} else {
		if !current.bold && next.bold {
			builder.WriteString(styles.BoldCode())
		}
		if !current.dim && next.dim {
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

	return builder.String()
}

// needsResetForTransition returns true when computing a targeted transition
// is unsafe or would not actually save bytes — in those cases the caller
// should emit "\x1b[0m" then the next style's full enable. Currently this
// only triggers when both bold and dim are mixed across the transition in a
// way that the 22 disable can't model cleanly without losing one of them
// — but our 22+re-enable code already covers that, so this remains a
// future-proof escape hatch.
func needsResetForTransition(current, next sgrAttrs) bool {
	_ = current
	_ = next
	return false
}

// writeEnableSGR appends the full enable sequence for attrs — used after
// emitting "\x1b[0m" to bring the cursor up to the desired state.
func writeEnableSGR(builder *strings.Builder, attrs sgrAttrs) {
	if attrs.fg != "" {
		builder.WriteString(attrs.fg)
	}
	if attrs.bg != "" {
		builder.WriteString(attrs.bg)
	}
	if attrs.bold {
		builder.WriteString(styles.BoldCode())
	}
	if attrs.dim {
		builder.WriteString(styles.DimCode())
	}
	if attrs.italic {
		builder.WriteString(styles.ItalicCode())
	}
	if attrs.underline {
		builder.WriteString(styles.UnderlineCode())
	}
	if attrs.inverse {
		builder.WriteString(styles.InverseCode())
	}
	if attrs.strikethrough {
		builder.WriteString(styles.StrikethroughCode())
	}
}

// renderedCell describes one terminal cell as parsed from a rendered line.
//
// width is the visible column width: 1 for the typical case and 2 for wide
// CJK / emoji clusters. cells[i] for a wide cluster's trailing column carries
// width == 0 and an empty text — these are never written individually because
// the wide cluster lays them down as part of its leading cell.
//
// style is the cumulative ANSI SGR sequence active at this cell (raw bytes,
// e.g. "\x1b[31m" or "\x1b[31m\x1b[1m"). Empty string means "no styling".
// Style is only ever the concatenation of CSI introducer sequences ending in
// 'm' that preceded the cell on the same line — anything more exotic (OSC,
// non-SGR CSI) trips the parser's safe-fallback bit, which makes the caller
// drop back to line-level diff.
type renderedCell struct {
	text  string
	width int
	style string
}

// renderedFrame is a parsed rendered output, keyed by row index. Each row is
// a slice of cells exactly as they would land on screen. When the parser
// encounters anything it cannot represent reliably (non-SGR CSI sequences,
// OSC sequences, raw control bytes other than newline), parseOK is false and
// the caller must not consume cells.
type renderedFrame struct {
	rows    [][]renderedCell
	parseOK bool
}

// parseFrame walks output one rune at a time, tracking the active SGR style
// stack as a single accumulated escape string. It is intentionally narrow:
// only CSI sequences ending in 'm' (SGR) are tracked; other CSI/OSC/control
// sequences cause parseOK to be set false. This keeps cell-level diff opt-in
// with a hard fallback for anything not strictly text + SGR.
func parseFrame(output string) renderedFrame {
	frame := renderedFrame{parseOK: true}
	if output == "" {
		return frame
	}

	// Strip the trailing newline that ensureTrailingNewline added — we want
	// rows to mirror visible lines, not the appended terminator.
	trimmed := output
	if strings.HasSuffix(trimmed, "\n") {
		trimmed = trimmed[:len(trimmed)-1]
	}

	if trimmed == "" {
		return frame
	}

	for _, line := range strings.Split(trimmed, "\n") {
		row, ok := parseLine(line)
		if !ok {
			frame.parseOK = false
			return frame
		}
		frame.rows = append(frame.rows, row)
	}

	return frame
}

// parseLine tokenizes a single line into cells. Active SGR state is rebuilt
// from scratch each line (terminals reset SGR at line boundaries in upstream
// Ink's incremental path because each emitted line is preceded by an explicit
// cursor move + erase — we mirror that assumption).
func parseLine(line string) ([]renderedCell, bool) {
	cells := make([]renderedCell, 0, len(line))
	style := ""

	// Walk byte-by-byte to detect ESC/CSI; for non-escape regions we fall
	// back to grapheme cluster iteration so wide runes, ZWJ sequences, and
	// emoji modifiers all land in a single cell.
	index := 0
	for index < len(line) {
		ch := line[index]
		if ch == 0x1b {
			// CSI introducer: ESC [ ... final-byte. Anything else (single
			// shift, ESC P / ESC ] / ESC \) is not something we handle, so
			// fall back.
			if index+1 >= len(line) || line[index+1] != '[' {
				return nil, false
			}
			end := index + 2
			for end < len(line) {
				b := line[end]
				if b >= 0x40 && b <= 0x7e {
					break
				}
				end++
			}
			if end >= len(line) {
				return nil, false
			}
			final := line[end]
			seq := line[index : end+1]
			if final == 'm' {
				if seq == cellDiffResetSGR || seq == "\x1b[m" {
					style = ""
				} else {
					style += seq
				}
			} else {
				// Non-SGR CSI (cursor move, erase, etc) inside a rendered
				// frame should not happen — bail out to fallback.
				return nil, false
			}
			index = end + 1
			continue
		}

		if ch < 0x20 && ch != '\t' {
			// Stray control byte we cannot represent; bail.
			return nil, false
		}

		// Find the next escape boundary so we can hand a slice to the
		// grapheme iterator.
		runStart := index
		for index < len(line) && line[index] != 0x1b {
			index++
		}
		segment := line[runStart:index]

		for _, cluster := range utils.GraphemeClusters(segment) {
			width := utils.StringWidth(cluster)
			cells = append(cells, renderedCell{text: cluster, width: width, style: style})
			// Pad trailing columns of a wide cluster with empty 0-width
			// markers so cell index == column index.
			for k := 1; k < width; k++ {
				cells = append(cells, renderedCell{width: 0, style: style})
			}
		}
	}

	return cells, true
}

// rowsEqual reports whether two parsed rows would render identically.
func rowsEqual(a, b []renderedCell) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].text != b[i].text || a[i].width != b[i].width || a[i].style != b[i].style {
			return false
		}
	}
	return true
}

// rowVisibleWidth sums column widths of a parsed row.
func rowVisibleWidth(row []renderedCell) int {
	total := 0
	for _, cell := range row {
		total += cell.width
	}
	return total
}

// buildCellDiffPayload produces the byte sequence that takes the terminal
// from showing prevFrame to showing nextFrame, assuming the cursor is parked
// at the bottom-left of the previous frame on entry. The payload begins by
// moving the cursor to the top-left of the previous frame's region, emits
// per-cell jumps and SGR transitions for changed cells only, then leaves the
// cursor at row visibleLines - 1 (consistent with how writeIncrementalRender
// finishes today, so buildCursorSuffix at the call site can reposition).
//
// Returns ok == false when frame sizes differ or fall-back conditions apply
// — callers must then defer to the line-level diff.
func buildCellDiffPayload(prevFrame, nextFrame renderedFrame) (string, bool) {
	if !prevFrame.parseOK || !nextFrame.parseOK {
		return "", false
	}

	if len(prevFrame.rows) != len(nextFrame.rows) {
		return "", false
	}

	rowCount := len(prevFrame.rows)
	if rowCount == 0 {
		return "", true
	}

	// Frame width must match per-row; a row whose width changed forces a
	// fall-back since we no longer know what's at the trailing columns.
	for i := 0; i < rowCount; i++ {
		if rowVisibleWidth(prevFrame.rows[i]) != rowVisibleWidth(nextFrame.rows[i]) {
			return "", false
		}
	}

	var builder strings.Builder
	// Move the cursor up to the top of the previous frame. The line-level
	// path enters with the cursor at the bottom of the prior render (one
	// row past the last line), so move up rowCount times to land at the
	// top-left of row 0.
	builder.WriteString(ansiCursorUp(rowCount))
	currentRow := 0
	currentCol := 0
	// Track both the raw style string (so we can detect "no change") and
	// the parsed attrs (so we can compute byte-minimal transitions). We
	// also remember whether anything has been emitted yet so the very
	// first styled cell can skip an unnecessary reset.
	currentStyleRaw := ""
	currentAttrs := sgrAttrs{parseOK: true}
	styleEmitted := false

	for rowIndex := 0; rowIndex < rowCount; rowIndex++ {
		prevRow := prevFrame.rows[rowIndex]
		nextRow := nextFrame.rows[rowIndex]
		if rowsEqual(prevRow, nextRow) {
			continue
		}

		// Walk columns; skip identical cells, emit changed runs together.
		col := 0
		for col < len(nextRow) {
			cell := nextRow[col]
			if cell.width == 0 {
				col++
				continue
			}

			same := col < len(prevRow) &&
				prevRow[col].text == cell.text &&
				prevRow[col].width == cell.width &&
				prevRow[col].style == cell.style
			if same {
				col += cell.width
				if cell.width == 0 {
					col++
				}
				continue
			}

			// Position the cursor on the changed cell.
			if rowIndex != currentRow {
				delta := rowIndex - currentRow
				if delta > 0 {
					builder.WriteString(ansiCursorDown(delta))
				} else {
					builder.WriteString(ansiCursorUp(-delta))
				}
				currentRow = rowIndex
				currentCol = 0
			}
			if col != currentCol {
				builder.WriteString(ansiCursorTo(col))
				currentCol = col
			}

			// Emit SGR transition if needed. Adjacent changed cells with
			// the same style will share a single transition emission —
			// styleEmitted + currentStyleRaw track that across the loop.
			if !styleEmitted || cell.style != currentStyleRaw {
				nextAttrs := parseSGR(cell.style)
				if !currentAttrs.parseOK || !nextAttrs.parseOK {
					// Safe fallback: full reset + raw enable string. This
					// is what the round-2 implementation did for every
					// transition; we only take this branch when the
					// minimal-diff parser cannot model one of the styles.
					if styleEmitted {
						builder.WriteString(cellDiffResetSGR)
					}
					if cell.style != "" {
						builder.WriteString(cell.style)
					}
				} else {
					// Compute byte-minimal transition. From the initial
					// "no style emitted yet" state, currentAttrs is the
					// default sgrAttrs{parseOK:true} clean slate — the
					// transition function emits only enable codes for
					// what's needed.
					builder.WriteString(sgrTransition(currentAttrs, nextAttrs))
				}
				currentStyleRaw = cell.style
				currentAttrs = nextAttrs
				styleEmitted = true
			}

			builder.WriteString(cell.text)
			currentCol += cell.width
			col += cell.width
		}
	}

	// Always restore default SGR before parking the cursor so subsequent
	// writes don't inherit a leftover color. Use the targeted disable
	// codes when possible; otherwise fall back to the full reset.
	if styleEmitted && currentStyleRaw != "" {
		if currentAttrs.parseOK {
			builder.WriteString(sgrTransition(currentAttrs, sgrAttrs{parseOK: true}))
		} else {
			builder.WriteString(cellDiffResetSGR)
		}
	}

	// Leave the cursor at row rowCount - 1 (the row buildCursorSuffix's
	// caller expects), regardless of where we last wrote.
	finalRow := rowCount - 1
	if currentRow != finalRow {
		delta := finalRow - currentRow
		if delta > 0 {
			builder.WriteString(ansiCursorDown(delta))
		} else {
			builder.WriteString(ansiCursorUp(-delta))
		}
	}
	builder.WriteString(ansiCursorTo(0))

	return builder.String(), true
}

// writeCellLevelRenderLocked is the cell-by-cell repaint path. It opts in
// only when CellLevelDiff is true on RenderOptions; otherwise the line-level
// path is used. On any parse failure or shape change it falls back to the
// existing incremental path so behavior remains conservative.
//
// The caller (writeRenderLocked) gates this branch on
// instance.cellLevelDiff && instance.incrementalRendering.
func (instance *Instance) writeCellLevelRenderLocked(logicalOutput, output string, cursorPosition *CursorPosition) error {
	if instance.stdout == nil {
		return nil
	}

	// First render or full reset — defer to the line-level path which knows
	// how to do an initial paint.
	if instance.previousOutput == "" {
		return instance.writeIncrementalRenderLocked(logicalOutput, output, cursorPosition)
	}

	// Identical frames or cursor-only changes — let the line-level path
	// take care of those small cases.
	if output == instance.previousOutput {
		return instance.writeIncrementalRenderLocked(logicalOutput, output, cursorPosition)
	}

	prevFrame := parseFrame(instance.previousOutput)
	nextFrame := parseFrame(output)
	if !prevFrame.parseOK || !nextFrame.parseOK {
		return instance.writeIncrementalRenderLocked(logicalOutput, output, cursorPosition)
	}

	if len(prevFrame.rows) != len(nextFrame.rows) {
		return instance.writeIncrementalRenderLocked(logicalOutput, output, cursorPosition)
	}

	for i := 0; i < len(prevFrame.rows); i++ {
		if rowVisibleWidth(prevFrame.rows[i]) != rowVisibleWidth(nextFrame.rows[i]) {
			return instance.writeIncrementalRenderLocked(logicalOutput, output, cursorPosition)
		}
	}

	if !instance.cursorHidden {
		if err := writePayload(instance.stdout, hideCursorEscape); err != nil {
			return err
		}
		instance.cursorHidden = true
	}

	payload, ok := buildCellDiffPayload(prevFrame, nextFrame)
	if !ok {
		return instance.writeIncrementalRenderLocked(logicalOutput, output, cursorPosition)
	}

	visibleLines := visibleLineCount(output)
	returnPrefix := buildReturnToBottomPrefix(
		instance.cursorWasShown,
		len(instance.previousLines),
		instance.previousCursorPosition,
	)
	full := returnPrefix + payload + buildCursorSuffix(visibleLines, cursorPosition)
	if err := writePayload(instance.stdout, full); err != nil {
		return err
	}

	instance.previousLogicalOutput = logicalOutput
	instance.previousOutput = output
	instance.previousLines = splitOutputLines(output)
	instance.previousLineCount = outputLineCount(output)
	instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
	instance.cursorWasShown = cursorPosition != nil
	return nil
}
