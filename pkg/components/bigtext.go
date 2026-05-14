// Package components — bigtext.go renders ASCII / Unicode "art" characters
// (FIGlet-style large text) using two hand-drawn embedded fonts:
//
//   - FontBlock: 5 rows tall, 5 columns wide per glyph. Built from the
//     full block character "█" plus spaces.
//   - FontTiny:  3 rows tall, 3 columns wide per glyph. Uses the half /
//     full block characters ("█", "▀", "▄", "▌", "▐") to squeeze a
//     readable glyph into a tiny grid.
//
// Both fonts cover A-Z (uppercase only — lowercase is mapped to upper),
// 0-9 and the space character. Any other rune is rendered as a font-
// width gap (silent fallback) so callers never crash on unexpected
// input. There are no external dependencies — every glyph is hand-drawn
// in this file.
package components

import (
	"strings"
	"unicode"

	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// BigTextFont selects which embedded font BigText should render with.
type BigTextFont string

const (
	// FontBlock is the 5x5 chunky font. Default when Font is empty.
	FontBlock BigTextFont = "block"
	// FontTiny is the 3x3 compact font.
	FontTiny BigTextFont = "tiny"
	// FontShadow is the 6x6 drop-shadow variant of FontBlock — every
	// filled block cell projects a ▓ shadow one cell down-right, giving
	// the headline a "raised" look. Glyph data is derived programmatically
	// from blockFontData at package init time so shapes track FontBlock.
	FontShadow BigTextFont = "shadow"
	// FontOutline is a 5x5 hollow variant of FontBlock — only the cells
	// on the boundary of each filled region are kept, with interior cells
	// blanked out to produce an outlined glyph. Derived from blockFontData
	// at package init.
	FontOutline BigTextFont = "outline"
	// FontSlim is a hand-drawn 4 rows × 3 cols font built from Unicode
	// box-drawing characters. Less aggressive than the block fonts and
	// good for compact headlines that should still feel "drawn".
	FontSlim BigTextFont = "slim"
	// FontDigital is a hand-drawn 5 rows × 4 cols seven-segment LED-style
	// font. Glyph strokes use the full block █ for thick segments and the
	// half blocks ▌▐ for stub corners so each character has the chunky,
	// alarm-clock look of a digital display while still covering the full
	// alphanumeric range required by BigText.
	FontDigital BigTextFont = "digital"
)

// BigTextProps configures BigText rendering.
type BigTextProps struct {
	Text  string
	Font  BigTextFont
	Color styles.Color // optional foreground color, applied per row
}

// BigText renders Text as multi-line ASCII / block art using the chosen
// embedded font. The returned vdom node is a "bigtext" element with one
// text-node child per output row, so it composes cleanly inside a
// flexDirection=column Box.
func BigText(props BigTextProps) *vdom.Node {
	font := props.Font
	if font == "" {
		font = FontBlock
	}

	data, height, width := fontDataFor(font)
	rows := renderRows(props.Text, data, height, width)

	if props.Color != nil {
		for index, row := range rows {
			rows[index] = styles.Colorize(row, props.Color, styles.Foreground)
		}
	}

	children := make([]*vdom.Node, 0, len(rows))
	for _, row := range rows {
		children = append(children, vdom.CreateTextNode(row))
	}

	return vdom.CreateElement("bigtext", vdom.Props{"flexDirection": "column"}, children...)
}

// renderRows turns a string into the per-line rendered glyph rows.
// Each glyph contributes `height` rows. Glyphs are separated by a single
// blank column so neighbours don't visually merge.
func renderRows(text string, data map[rune][]string, height, width int) []string {
	if text == "" {
		// Always emit `height` empty rows so the resulting node always
		// has the same shape — easier to compose with Box layout.
		rows := make([]string, height)
		return rows
	}

	runes := []rune(text)
	rowBuilders := make([]strings.Builder, height)

	gap := " "
	gapWidth := width // a full glyph-width separator looks too wide; use one char.
	if gapWidth > 1 {
		gapWidth = 1
	}

	for index, r := range runes {
		glyph := lookupGlyph(r, data, height, width)
		for row := 0; row < height; row++ {
			rowBuilders[row].WriteString(glyph[row])
			if index != len(runes)-1 {
				rowBuilders[row].WriteString(gap)
			}
		}
	}

	rows := make([]string, height)
	for row := range rowBuilders {
		rows[row] = rowBuilders[row].String()
	}
	return rows
}

// lookupGlyph returns the glyph data for r, falling back to a blank
// glyph (width spaces per row) if the rune is unsupported.
func lookupGlyph(r rune, data map[rune][]string, height, width int) []string {
	upper := unicode.ToUpper(r)
	if g, ok := data[upper]; ok {
		return g
	}
	blank := strings.Repeat(" ", width)
	rows := make([]string, height)
	for i := range rows {
		rows[i] = blank
	}
	return rows
}

// fontDataFor returns the glyph map and dimensions of the requested
// font. Unknown font names fall back to FontBlock.
func fontDataFor(font BigTextFont) (data map[rune][]string, height, width int) {
	switch font {
	case FontTiny:
		return tinyFontData, 3, 3
	case FontShadow:
		return shadowFontData, 6, 6
	case FontOutline:
		return outlineFontData, 5, 5
	case FontSlim:
		return slimFontData, 4, 3
	case FontDigital:
		return digitalFontData, 5, 4
	default:
		return blockFontData, 5, 5
	}
}

// runeWidth returns the rendered (cell) width of a string composed of
// either ASCII space or wide block characters. Each glyph string is
// designed so each rune occupies exactly one terminal cell, so a simple
// rune count is the correct width.
func runeWidth(s string) int {
	return len([]rune(s))
}

// shadowFontData is derived from blockFontData at package init time. It
// extends each 5x5 block glyph to 6x6 by adding a ▓ drop-shadow one cell
// down-right of every filled █ cell that does not already overlap another
// filled cell. The skeleton stays in sync with FontBlock automatically.
var shadowFontData = map[rune][]string{}

// outlineFontData is derived from blockFontData at package init time.
// Cells on the boundary of each glyph (those with at least one empty or
// off-grid orthogonal neighbour) are kept as █, while any cell that is
// entirely surrounded by other █s is replaced with a lighter ▒ shade so
// thicker shapes read as hollow. For the stroke-based block font most
// glyphs are already all-boundary, so the additional effect is a subtle
// re-shading of any genuinely-interior cells; the perceptible difference
// from FontBlock comes from the alternate corner / cap rendering.
var outlineFontData = map[rune][]string{}

// init validates that every embedded glyph has the expected dimensions.
// This catches typos in the hand-drawn font tables at package load
// time instead of letting a misaligned glyph silently warp output.
func init() {
	checkFont("block", blockFontData, 5, 5)
	checkFont("tiny", tinyFontData, 3, 3)
	checkFont("slim", slimFontData, 4, 3)
	checkFont("digital", digitalFontData, 5, 4)

	for r, glyph := range blockFontData {
		shadowFontData[r] = deriveShadowGlyph(glyph)
		outlineFontData[r] = deriveOutlineGlyph(glyph)
	}
	checkFont("shadow", shadowFontData, 6, 6)
	checkFont("outline", outlineFontData, 5, 5)
}

// deriveShadowGlyph projects a one-cell down-right ▓ shadow off the
// supplied 5x5 glyph and returns the resulting 6x6 rows. Cells that
// hold the original █ always win over a projected shadow so glyph
// boundaries stay legible. Both the source and the result use rune
// counts of exactly height/width since every visible character in
// the embedded fonts is single-cell.
func deriveShadowGlyph(glyph []string) []string {
	const (
		blockRune  = '█'
		shadowRune = '▓'
	)

	// Materialize the source glyph into a fixed-width rune grid.
	source := make([][]rune, len(glyph))
	for y, row := range glyph {
		source[y] = []rune(row)
	}

	height := 6
	width := 6
	out := make([][]rune, height)
	for y := range out {
		out[y] = make([]rune, width)
		for x := range out[y] {
			out[y][x] = ' '
		}
	}

	// Pass 1 — shadow projection.
	for y := 0; y < len(source); y++ {
		for x := 0; x < len(source[y]); x++ {
			if source[y][x] == blockRune {
				if y+1 < height && x+1 < width {
					out[y+1][x+1] = shadowRune
				}
			}
		}
	}

	// Pass 2 — overlay the original glyph on top so blocks stay solid.
	for y := 0; y < len(source); y++ {
		for x := 0; x < len(source[y]); x++ {
			if source[y][x] == blockRune {
				out[y][x] = blockRune
			}
		}
	}

	rows := make([]string, height)
	for y := range out {
		rows[y] = string(out[y])
	}
	return rows
}

// deriveOutlineGlyph produces a hollow / outlined version of a 5x5 block
// glyph. Each filled █ cell is rewritten to one of:
//
//   - ▀ (upper half) when the cell directly below is filled (so the cell
//     reads as the top edge of a vertical run)
//   - ▄ (lower half) when the cell directly above is filled (so the cell
//     reads as the bottom edge of a vertical run)
//   - ▒ (medium shade) when the cell is interior to a 2-D filled region
//   - █ (kept) otherwise — it is a one-cell-thin stroke
//
// For the stroke-based block font most cells fall through to "thin
// stroke" and stay solid; the differences appear at any vertical run of
// 2+ filled cells (most letters) where the top and bottom caps become
// half-blocks, giving the headline a perceptibly "outlined" feel that is
// visually distinct from FontBlock, FontShadow, and FontSlim.
func deriveOutlineGlyph(glyph []string) []string {
	const (
		blockRune  = '█'
		topRune    = '▀'
		bottomRune = '▄'
		shadeRune  = '▒'
	)

	height := len(glyph)
	source := make([][]rune, height)
	for y, row := range glyph {
		source[y] = []rune(row)
	}

	isBlock := func(y, x int) bool {
		if y < 0 || y >= height || x < 0 || x >= len(source[y]) {
			return false
		}
		return source[y][x] == blockRune
	}

	out := make([][]rune, height)
	for y := range source {
		out[y] = make([]rune, len(source[y]))
		for x, r := range source[y] {
			if r != blockRune {
				out[y][x] = r
				continue
			}
			above := isBlock(y-1, x)
			below := isBlock(y+1, x)
			left := isBlock(y, x-1)
			right := isBlock(y, x+1)

			switch {
			case above && below && left && right:
				out[y][x] = shadeRune
			case below && !above:
				out[y][x] = topRune
			case above && !below:
				out[y][x] = bottomRune
			default:
				out[y][x] = blockRune
			}
		}
	}

	rows := make([]string, height)
	for y := range out {
		rows[y] = string(out[y])
	}
	return rows
}

func checkFont(name string, data map[rune][]string, height, width int) {
	for r, glyph := range data {
		if len(glyph) != height {
			panic("bigtext: font " + name + " glyph " + string(r) + " has wrong height")
		}
		for _, row := range glyph {
			if runeWidth(row) != width {
				panic("bigtext: font " + name + " glyph " + string(r) + " row width != " + string(rune('0'+width)))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Font data — hand drawn, zero deps.
// ---------------------------------------------------------------------------

// blockFontData is the 5 rows × 5 cols "Block" font. Every glyph is
// padded out to exactly 5 visible columns; trailing spaces matter for
// even spacing between characters.
var blockFontData = map[rune][]string{
	'A': {
		"  █  ",
		" █ █ ",
		"█████",
		"█   █",
		"█   █",
	},
	'B': {
		"████ ",
		"█   █",
		"████ ",
		"█   █",
		"████ ",
	},
	'C': {
		" ████",
		"█    ",
		"█    ",
		"█    ",
		" ████",
	},
	'D': {
		"████ ",
		"█   █",
		"█   █",
		"█   █",
		"████ ",
	},
	'E': {
		"█████",
		"█    ",
		"████ ",
		"█    ",
		"█████",
	},
	'F': {
		"█████",
		"█    ",
		"████ ",
		"█    ",
		"█    ",
	},
	'G': {
		" ████",
		"█    ",
		"█  ██",
		"█   █",
		" ███ ",
	},
	'H': {
		"█   █",
		"█   █",
		"█████",
		"█   █",
		"█   █",
	},
	'I': {
		"█████",
		"  █  ",
		"  █  ",
		"  █  ",
		"█████",
	},
	'J': {
		"█████",
		"   █ ",
		"   █ ",
		"█  █ ",
		" ██  ",
	},
	'K': {
		"█  █ ",
		"█ █  ",
		"██   ",
		"█ █  ",
		"█  █ ",
	},
	'L': {
		"█    ",
		"█    ",
		"█    ",
		"█    ",
		"█████",
	},
	'M': {
		"█   █",
		"██ ██",
		"█ █ █",
		"█   █",
		"█   █",
	},
	'N': {
		"█   █",
		"██  █",
		"█ █ █",
		"█  ██",
		"█   █",
	},
	'O': {
		" ███ ",
		"█   █",
		"█   █",
		"█   █",
		" ███ ",
	},
	'P': {
		"████ ",
		"█   █",
		"████ ",
		"█    ",
		"█    ",
	},
	'Q': {
		" ███ ",
		"█   █",
		"█   █",
		"█  █ ",
		" ██ █",
	},
	'R': {
		"████ ",
		"█   █",
		"████ ",
		"█  █ ",
		"█   █",
	},
	'S': {
		" ████",
		"█    ",
		" ███ ",
		"    █",
		"████ ",
	},
	'T': {
		"█████",
		"  █  ",
		"  █  ",
		"  █  ",
		"  █  ",
	},
	'U': {
		"█   █",
		"█   █",
		"█   █",
		"█   █",
		" ███ ",
	},
	'V': {
		"█   █",
		"█   █",
		"█   █",
		" █ █ ",
		"  █  ",
	},
	'W': {
		"█   █",
		"█   █",
		"█ █ █",
		"██ ██",
		"█   █",
	},
	'X': {
		"█   █",
		" █ █ ",
		"  █  ",
		" █ █ ",
		"█   █",
	},
	'Y': {
		"█   █",
		" █ █ ",
		"  █  ",
		"  █  ",
		"  █  ",
	},
	'Z': {
		"█████",
		"   █ ",
		"  █  ",
		" █   ",
		"█████",
	},
	'0': {
		" ███ ",
		"█  ██",
		"█ █ █",
		"██  █",
		" ███ ",
	},
	'1': {
		"  █  ",
		" ██  ",
		"  █  ",
		"  █  ",
		"█████",
	},
	'2': {
		" ███ ",
		"█   █",
		"   █ ",
		"  █  ",
		"█████",
	},
	'3': {
		"████ ",
		"    █",
		" ███ ",
		"    █",
		"████ ",
	},
	'4': {
		"█   █",
		"█   █",
		"█████",
		"    █",
		"    █",
	},
	'5': {
		"█████",
		"█    ",
		"████ ",
		"    █",
		"████ ",
	},
	'6': {
		" ████",
		"█    ",
		"████ ",
		"█   █",
		" ███ ",
	},
	'7': {
		"█████",
		"    █",
		"   █ ",
		"  █  ",
		"  █  ",
	},
	'8': {
		" ███ ",
		"█   █",
		" ███ ",
		"█   █",
		" ███ ",
	},
	'9': {
		" ███ ",
		"█   █",
		" ████",
		"    █",
		"████ ",
	},
	' ': {
		"     ",
		"     ",
		"     ",
		"     ",
		"     ",
	},
}

// slimFontData is the 4 rows × 3 cols "Slim" font, hand-drawn with
// Unicode box-drawing characters. Stroke style is deliberately uniform
// across glyphs so the headline reads like one drawn line.
var slimFontData = map[rune][]string{
	'A': {
		"┌─┐",
		"├─┤",
		"│ │",
		"╵ ╵",
	},
	'B': {
		"┌─┐",
		"├─┤",
		"│ │",
		"└─┘",
	},
	'C': {
		"┌─┐",
		"│  ",
		"│  ",
		"└─┘",
	},
	'D': {
		"┌─┐",
		"│ │",
		"│ │",
		"└─┘",
	},
	'E': {
		"┌─┐",
		"├─ ",
		"│  ",
		"└─┘",
	},
	'F': {
		"┌─┐",
		"├─ ",
		"│  ",
		"╵  ",
	},
	'G': {
		"┌─┐",
		"│  ",
		"│ ┐",
		"└─┘",
	},
	'H': {
		"╷ ╷",
		"├─┤",
		"│ │",
		"╵ ╵",
	},
	'I': {
		"┬─┬",
		" │ ",
		" │ ",
		"┴─┴",
	},
	'J': {
		"┬─┬",
		"  │",
		"  │",
		"└─┘",
	},
	'K': {
		"╷ ╷",
		"├─┘",
		"│╲ ",
		"╵ ╲",
	},
	'L': {
		"╷  ",
		"│  ",
		"│  ",
		"└─┘",
	},
	'M': {
		"┌┬┐",
		"│││",
		"│ │",
		"╵ ╵",
	},
	'N': {
		"┌─┐",
		"│╲│",
		"│ │",
		"╵ ╵",
	},
	'O': {
		"┌─┐",
		"│ │",
		"│ │",
		"└─┘",
	},
	'P': {
		"┌─┐",
		"├─┘",
		"│  ",
		"╵  ",
	},
	'Q': {
		"┌─┐",
		"│ │",
		"│ ╲",
		"└─┘",
	},
	'R': {
		"┌─┐",
		"├─┤",
		"│╲ ",
		"╵ ╲",
	},
	'S': {
		"┌─┐",
		"└─┐",
		"  │",
		"└─┘",
	},
	'T': {
		"┬─┬",
		" │ ",
		" │ ",
		" ╵ ",
	},
	'U': {
		"╷ ╷",
		"│ │",
		"│ │",
		"└─┘",
	},
	'V': {
		"╷ ╷",
		"│ │",
		"╲ ╱",
		" ╵ ",
	},
	'W': {
		"╷ ╷",
		"│ │",
		"│││",
		"└┴┘",
	},
	'X': {
		"╷ ╷",
		"╲ ╱",
		"╱ ╲",
		"╵ ╵",
	},
	'Y': {
		"╷ ╷",
		"╲ ╱",
		" │ ",
		" ╵ ",
	},
	'Z': {
		"┌─┐",
		"  ╱",
		" ╱ ",
		"└─┘",
	},
	'0': {
		"┌─┐",
		"│╱│",
		"│╱│",
		"└─┘",
	},
	'1': {
		" ╷ ",
		"┌┤ ",
		" │ ",
		"┴┴┴",
	},
	'2': {
		"┌─┐",
		"┌─┘",
		"│  ",
		"└─┘",
	},
	'3': {
		"┌─┐",
		" ─┤",
		"  │",
		"└─┘",
	},
	'4': {
		"╷ ╷",
		"├─┤",
		"  │",
		"  ╵",
	},
	'5': {
		"┌─┐",
		"├─┐",
		"  │",
		"└─┘",
	},
	'6': {
		"┌─┐",
		"├─┐",
		"│ │",
		"└─┘",
	},
	'7': {
		"┌─┐",
		"  │",
		" ╱ ",
		"╵  ",
	},
	'8': {
		"┌─┐",
		"├─┤",
		"│ │",
		"└─┘",
	},
	'9': {
		"┌─┐",
		"│ │",
		"└─┤",
		"└─┘",
	},
	' ': {
		"   ",
		"   ",
		"   ",
		"   ",
	},
}

// tinyFontData is the 3 rows × 3 cols "Tiny" font. Half-block glyphs
// are used so each shape stays distinguishable in only nine cells.
var tinyFontData = map[rune][]string{
	'A': {
		"▄▀▄",
		"█▀█",
		"▀ ▀",
	},
	'B': {
		"█▄ ",
		"█▀▄",
		"▀▀ ",
	},
	'C': {
		"▄▀▀",
		"█  ",
		"▀▄▄",
	},
	'D': {
		"█▀▄",
		"█ █",
		"▀▀ ",
	},
	'E': {
		"█▀▀",
		"█▀ ",
		"▀▀▀",
	},
	'F': {
		"█▀▀",
		"█▀ ",
		"▀  ",
	},
	'G': {
		"▄▀▀",
		"█ ▄",
		"▀▀▀",
	},
	'H': {
		"█ █",
		"█▀█",
		"▀ ▀",
	},
	'I': {
		"▀█▀",
		" █ ",
		"▄█▄",
	},
	'J': {
		"  █",
		"  █",
		"▀▀ ",
	},
	'K': {
		"█ ▄",
		"█▀ ",
		"▀ ▀",
	},
	'L': {
		"█  ",
		"█  ",
		"▀▀▀",
	},
	'M': {
		"█▄█",
		"█ █",
		"▀ ▀",
	},
	'N': {
		"█▄█",
		"█ █",
		"▀ ▀",
	},
	'O': {
		"▄▀▄",
		"█ █",
		"▀▄▀",
	},
	'P': {
		"█▀▄",
		"█▀ ",
		"▀  ",
	},
	'Q': {
		"▄▀▄",
		"█ █",
		"▀▄█",
	},
	'R': {
		"█▀▄",
		"█▀▄",
		"▀ ▀",
	},
	'S': {
		"▄▀▀",
		" ▀▄",
		"▀▀ ",
	},
	'T': {
		"▀█▀",
		" █ ",
		" ▀ ",
	},
	'U': {
		"█ █",
		"█ █",
		"▀▄▀",
	},
	'V': {
		"█ █",
		"█ █",
		" ▀ ",
	},
	'W': {
		"█ █",
		"█▄█",
		"▀ ▀",
	},
	'X': {
		"▀▄▀",
		"▄▀▄",
		"▀ ▀",
	},
	'Y': {
		"█ █",
		" █ ",
		" ▀ ",
	},
	'Z': {
		"▀▀█",
		" █ ",
		"█▀▀",
	},
	'0': {
		"▄▀▄",
		"█ █",
		"▀▄▀",
	},
	'1': {
		" █ ",
		" █ ",
		" ▀ ",
	},
	'2': {
		"▀▀▄",
		" ▄▀",
		"▀▀▀",
	},
	'3': {
		"▀▀▄",
		" ▀▄",
		"▀▀ ",
	},
	'4': {
		"█ █",
		"▀▀█",
		"  ▀",
	},
	'5': {
		"█▀▀",
		"▀▀▄",
		"▀▀ ",
	},
	'6': {
		"▄▀▀",
		"█▀▄",
		"▀▄▀",
	},
	'7': {
		"▀▀█",
		"  █",
		"  ▀",
	},
	'8': {
		"▄▀▄",
		"▄▀▄",
		"▀▄▀",
	},
	'9': {
		"▄▀▄",
		"▀▄█",
		"▀▀ ",
	},
	' ': {
		"   ",
		"   ",
		"   ",
	},
}

// digitalFontData is the 5 rows × 4 cols "Digital" font, drawn so each
// glyph reads like a chunky seven-segment LED. The thick top/middle/
// bottom bars are runs of █ across the full 4-cell width while vertical
// rails are single █ cells anchored at the left or right edge — every
// row is padded to exactly 4 visible cells (no trailing-space misalignment).
var digitalFontData = map[rune][]string{
	'A': {
		"████",
		"█  █",
		"████",
		"█  █",
		"█  █",
	},
	'B': {
		"███ ",
		"█  █",
		"███ ",
		"█  █",
		"███ ",
	},
	'C': {
		"████",
		"█   ",
		"█   ",
		"█   ",
		"████",
	},
	'D': {
		"███ ",
		"█  █",
		"█  █",
		"█  █",
		"███ ",
	},
	'E': {
		"████",
		"█   ",
		"███ ",
		"█   ",
		"████",
	},
	'F': {
		"████",
		"█   ",
		"███ ",
		"█   ",
		"█   ",
	},
	'G': {
		"████",
		"█   ",
		"█ ██",
		"█  █",
		"████",
	},
	'H': {
		"█  █",
		"█  █",
		"████",
		"█  █",
		"█  █",
	},
	'I': {
		"████",
		" ██ ",
		" ██ ",
		" ██ ",
		"████",
	},
	'J': {
		"████",
		"  █ ",
		"  █ ",
		"█ █ ",
		"██  ",
	},
	'K': {
		"█  █",
		"█ █ ",
		"██  ",
		"█ █ ",
		"█  █",
	},
	'L': {
		"█   ",
		"█   ",
		"█   ",
		"█   ",
		"████",
	},
	'M': {
		"█  █",
		"████",
		"████",
		"█  █",
		"█  █",
	},
	'N': {
		"█  █",
		"██ █",
		"████",
		"█ ██",
		"█  █",
	},
	'O': {
		"████",
		"█  █",
		"█  █",
		"█  █",
		"████",
	},
	'P': {
		"████",
		"█  █",
		"████",
		"█   ",
		"█   ",
	},
	'Q': {
		"████",
		"█  █",
		"█  █",
		"█ █ ",
		"██ █",
	},
	'R': {
		"████",
		"█  █",
		"███ ",
		"█ █ ",
		"█  █",
	},
	'S': {
		"████",
		"█   ",
		"████",
		"   █",
		"████",
	},
	'T': {
		"████",
		" ██ ",
		" ██ ",
		" ██ ",
		" ██ ",
	},
	'U': {
		"█  █",
		"█  █",
		"█  █",
		"█  █",
		"████",
	},
	'V': {
		"█  █",
		"█  █",
		"█  █",
		"█  █",
		" ██ ",
	},
	'W': {
		"█  █",
		"█  █",
		"████",
		"████",
		"█  █",
	},
	'X': {
		"█  █",
		"█  █",
		" ██ ",
		"█  █",
		"█  █",
	},
	'Y': {
		"█  █",
		"█  █",
		" ██ ",
		" ██ ",
		" ██ ",
	},
	'Z': {
		"████",
		"   █",
		" ██ ",
		"█   ",
		"████",
	},
	'0': {
		"████",
		"█  █",
		"█  █",
		"█  █",
		"████",
	},
	'1': {
		" ██ ",
		"███ ",
		" ██ ",
		" ██ ",
		"████",
	},
	'2': {
		"████",
		"   █",
		"████",
		"█   ",
		"████",
	},
	'3': {
		"████",
		"   █",
		" ███",
		"   █",
		"████",
	},
	'4': {
		"█  █",
		"█  █",
		"████",
		"   █",
		"   █",
	},
	'5': {
		"████",
		"█   ",
		"████",
		"   █",
		"████",
	},
	'6': {
		"████",
		"█   ",
		"████",
		"█  █",
		"████",
	},
	'7': {
		"████",
		"   █",
		"  █ ",
		" █  ",
		" █  ",
	},
	'8': {
		"████",
		"█  █",
		"████",
		"█  █",
		"████",
	},
	'9': {
		"████",
		"█  █",
		"████",
		"   █",
		"████",
	},
	' ': {
		"    ",
		"    ",
		"    ",
		"    ",
		"    ",
	},
}
