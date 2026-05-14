package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func bigTextRows(t *testing.T, props components.BigTextProps) []string {
	t.Helper()
	node := components.BigText(props)
	if node == nil {
		t.Fatal("BigText returned nil node")
	}
	if node.ElementType != "bigtext" {
		t.Fatalf("element type = %q, want bigtext", node.ElementType)
	}
	rows := make([]string, 0, len(node.Children))
	for _, child := range node.Children {
		if child == nil || child.Type != vdom.TextNode {
			t.Fatalf("expected text-node children, got %#v", child)
		}
		rows = append(rows, child.Text)
	}
	return rows
}

func TestBigTextDefaultsToBlockFont(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A"})
	if len(rows) != 5 {
		t.Fatalf("default font row count = %d, want 5", len(rows))
	}
}

func TestBigTextBlockSingleCharacterShape(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A", Font: components.FontBlock})
	if len(rows) != 5 {
		t.Fatalf("block 'A' rows = %d, want 5", len(rows))
	}
	want := []string{
		"  █  ",
		" █ █ ",
		"█████",
		"█   █",
		"█   █",
	}
	for i, r := range rows {
		if r != want[i] {
			t.Fatalf("row %d = %q, want %q", i, r, want[i])
		}
	}
}

func TestBigTextTinySingleCharacterShape(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A", Font: components.FontTiny})
	if len(rows) != 3 {
		t.Fatalf("tiny 'A' rows = %d, want 3", len(rows))
	}
	want := []string{"▄▀▄", "█▀█", "▀ ▀"}
	for i, r := range rows {
		if r != want[i] {
			t.Fatalf("tiny row %d = %q, want %q", i, r, want[i])
		}
	}
}

func TestBigTextMultiCharacterCompositionBlock(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "HI", Font: components.FontBlock})
	if len(rows) != 5 {
		t.Fatalf("HI block rows = %d, want 5", len(rows))
	}
	// Each glyph is 5 cols and a 1-col gap between glyphs ⇒ 5+1+5 = 11.
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 11 {
			t.Fatalf("HI row %d width = %d, want 11 — got %q", i, rw, r)
		}
	}
	// First row of H starts with "█   █" and I starts with "█████".
	if !strings.HasPrefix(rows[0], "█   █") {
		t.Fatalf("HI first row missing H glyph: %q", rows[0])
	}
	if !strings.HasSuffix(rows[0], "█████") {
		t.Fatalf("HI first row missing I glyph: %q", rows[0])
	}
}

func TestBigTextMultiCharacterCompositionTiny(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "OK", Font: components.FontTiny})
	if len(rows) != 3 {
		t.Fatalf("tiny OK rows = %d, want 3", len(rows))
	}
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 7 {
			t.Fatalf("tiny OK row %d width = %d, want 7 — got %q", i, rw, r)
		}
	}
}

func TestBigTextLowercaseFoldsToUppercase(t *testing.T) {
	upper := bigTextRows(t, components.BigTextProps{Text: "G", Font: components.FontBlock})
	lower := bigTextRows(t, components.BigTextProps{Text: "g", Font: components.FontBlock})
	if len(upper) != len(lower) {
		t.Fatalf("upper rows %d != lower rows %d", len(upper), len(lower))
	}
	for i := range upper {
		if upper[i] != lower[i] {
			t.Fatalf("row %d differs: upper=%q lower=%q", i, upper[i], lower[i])
		}
	}
}

func TestBigTextUnsupportedRuneFallsBackToBlank(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "@", Font: components.FontBlock})
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}
	for i, r := range rows {
		if strings.TrimSpace(r) != "" {
			t.Fatalf("row %d should be blank, got %q", i, r)
		}
		if rw := len([]rune(r)); rw != 5 {
			t.Fatalf("row %d width = %d, want 5", i, rw)
		}
	}
}

func TestBigTextSpaceCharacterRendersBlankGlyph(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A B", Font: components.FontBlock})
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}
	// Each char (5) + 2 gaps (1 each) = 5+1+5+1+5 = 17.
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 17 {
			t.Fatalf("row %d width = %d, want 17 — got %q", i, rw, r)
		}
	}
}

func TestBigTextEmptyStringYieldsBlankRows(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "", Font: components.FontBlock})
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}
	for i, r := range rows {
		if r != "" {
			t.Fatalf("row %d = %q, want empty", i, r)
		}
	}
}

func TestBigTextEmptyStringTinyYieldsBlankRows(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "", Font: components.FontTiny})
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
}

func TestBigTextColorAppliedToEachRow(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "X", Font: components.FontBlock, Color: styles.Red})
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}
	for i, r := range rows {
		if !strings.Contains(r, "\x1b[31m") {
			t.Fatalf("row %d missing red ANSI code: %q", i, r)
		}
		if !strings.Contains(r, "\x1b[0m") {
			t.Fatalf("row %d missing ANSI reset: %q", i, r)
		}
	}
}

func TestBigTextColorAppliedToTinyFont(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "Z", Font: components.FontTiny, Color: styles.Green})
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	for _, r := range rows {
		if !strings.Contains(r, "\x1b[32m") {
			t.Fatalf("row missing green ANSI: %q", r)
		}
	}
}

func TestBigTextShadowFontDimensions(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A", Font: components.FontShadow})
	if len(rows) != 6 {
		t.Fatalf("shadow rows = %d, want 6", len(rows))
	}
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 6 {
			t.Fatalf("shadow row %d width = %d, want 6 — got %q", i, rw, r)
		}
	}
}

func TestBigTextShadowFontHasShadowGlyph(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A", Font: components.FontShadow})
	joined := strings.Join(rows, "")
	if !strings.ContainsRune(joined, '▓') {
		t.Fatalf("expected shadow font to contain ▓ shadow runes, got %q", joined)
	}
	if !strings.ContainsRune(joined, '█') {
		t.Fatalf("expected shadow font to retain █ block runes, got %q", joined)
	}
}

func TestBigTextShadowSupportsAllAlphanumeric(t *testing.T) {
	const all = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, r := range all {
		rows := bigTextRows(t, components.BigTextProps{Text: string(r), Font: components.FontShadow})
		if len(rows) != 6 {
			t.Fatalf("shadow rune=%q rendered %d rows, want 6", r, len(rows))
		}
		joined := strings.Join(rows, "")
		if strings.TrimSpace(joined) == "" {
			t.Fatalf("shadow rune=%q rendered as blank", r)
		}
	}
}

func TestBigTextFontsHaveDifferentRowCounts(t *testing.T) {
	block := bigTextRows(t, components.BigTextProps{Text: "GO", Font: components.FontBlock})
	tiny := bigTextRows(t, components.BigTextProps{Text: "GO", Font: components.FontTiny})
	if len(block) == len(tiny) {
		t.Fatalf("expected differing row counts: block=%d tiny=%d", len(block), len(tiny))
	}
	if len(block) != 5 || len(tiny) != 3 {
		t.Fatalf("expected block=5/tiny=3, got block=%d tiny=%d", len(block), len(tiny))
	}
}

func TestBigTextSupportsAllAlphanumericPlusSpace(t *testing.T) {
	// Confirm all 37 supported runes render non-blank rows under both
	// fonts (space is the only blank-glyph case, so we exclude it).
	const all = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, font := range []components.BigTextFont{components.FontBlock, components.FontTiny} {
		for _, r := range all {
			rows := bigTextRows(t, components.BigTextProps{Text: string(r), Font: font})
			joined := strings.Join(rows, "")
			if strings.TrimSpace(joined) == "" {
				t.Fatalf("font=%s rune=%q rendered as blank", font, r)
			}
		}
	}

	// Space must render a fully blank glyph (correct width, no ink).
	for _, font := range []components.BigTextFont{components.FontBlock, components.FontTiny} {
		rows := bigTextRows(t, components.BigTextProps{Text: " ", Font: font})
		joined := strings.Join(rows, "")
		if strings.TrimSpace(joined) != "" {
			t.Fatalf("font=%s space rendered ink: %q", font, joined)
		}
	}
}

func TestBigTextUnknownFontFallsBackToBlock(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A", Font: components.BigTextFont("nonsense")})
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5 (block fallback)", len(rows))
	}
}

func TestBigTextNodeCarriesColumnFlexDirection(t *testing.T) {
	node := components.BigText(components.BigTextProps{Text: "A"})
	if node.Props["flexDirection"] != "column" {
		t.Fatalf("expected flexDirection=column, got %v", node.Props["flexDirection"])
	}
}

// --- Outline font ----------------------------------------------------------

func TestBigTextOutlineFontDimensions(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A", Font: components.FontOutline})
	if len(rows) != 5 {
		t.Fatalf("outline rows = %d, want 5", len(rows))
	}
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 5 {
			t.Fatalf("outline row %d width = %d, want 5 — got %q", i, rw, r)
		}
	}
}

func TestBigTextOutlineSupportsAllAlphanumeric(t *testing.T) {
	const all = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, r := range all {
		rows := bigTextRows(t, components.BigTextProps{Text: string(r), Font: components.FontOutline})
		if len(rows) != 5 {
			t.Fatalf("outline rune=%q rendered %d rows, want 5", r, len(rows))
		}
		joined := strings.Join(rows, "")
		if strings.TrimSpace(joined) == "" {
			t.Fatalf("outline rune=%q rendered as blank", r)
		}
		if !strings.ContainsRune(joined, '█') {
			t.Fatalf("outline rune=%q lost all block runes (boundary should remain): %q", r, joined)
		}
	}
}

func TestBigTextOutlineDiffersFromBlock(t *testing.T) {
	// At least one alphanumeric glyph must visibly differ from the
	// block source — that's the whole point of the derivation. Half-
	// block caps (▀ / ▄) are emitted whenever a vertical run of 2+
	// filled cells exists, which holds for nearly every letter / digit.
	differs := false
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" {
		blockRows := bigTextRows(t, components.BigTextProps{Text: string(r), Font: components.FontBlock})
		outlineRows := bigTextRows(t, components.BigTextProps{Text: string(r), Font: components.FontOutline})
		if strings.Join(blockRows, "\n") != strings.Join(outlineRows, "\n") {
			differs = true
			break
		}
	}
	if !differs {
		t.Fatalf("outline font produced output identical to block for every glyph — derivation never fired")
	}
}

func TestBigTextOutlineEmitsHalfBlockCaps(t *testing.T) {
	// Letters with vertical bars (e.g. 'L', 'I', 'H') must have at
	// least one cell rendered as ▀ or ▄ in the outline font.
	for _, r := range "HILUT" {
		rows := bigTextRows(t, components.BigTextProps{Text: string(r), Font: components.FontOutline})
		joined := strings.Join(rows, "")
		if !strings.ContainsRune(joined, '▀') && !strings.ContainsRune(joined, '▄') {
			t.Fatalf("outline rune=%q expected ▀/▄ caps, got %q", r, joined)
		}
	}
}

// --- Slim font -------------------------------------------------------------

func TestBigTextSlimFontDimensions(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A", Font: components.FontSlim})
	if len(rows) != 4 {
		t.Fatalf("slim rows = %d, want 4", len(rows))
	}
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 3 {
			t.Fatalf("slim row %d width = %d, want 3 — got %q", i, rw, r)
		}
	}
}

func TestBigTextSlimSupportsAllAlphanumericPlusSpace(t *testing.T) {
	const all = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, r := range all {
		rows := bigTextRows(t, components.BigTextProps{Text: string(r), Font: components.FontSlim})
		if len(rows) != 4 {
			t.Fatalf("slim rune=%q rendered %d rows, want 4", r, len(rows))
		}
		joined := strings.Join(rows, "")
		if strings.TrimSpace(joined) == "" {
			t.Fatalf("slim rune=%q rendered as blank", r)
		}
	}
	// Space must render fully blank.
	rows := bigTextRows(t, components.BigTextProps{Text: " ", Font: components.FontSlim})
	if strings.TrimSpace(strings.Join(rows, "")) != "" {
		t.Fatalf("slim space rendered ink: %q", rows)
	}
}

func TestBigTextSlimMultiCharacterComposition(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "HI", Font: components.FontSlim})
	if len(rows) != 4 {
		t.Fatalf("slim HI rows = %d, want 4", len(rows))
	}
	// Each glyph is 3 cols, gap is 1 col ⇒ 3+1+3 = 7.
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 7 {
			t.Fatalf("slim HI row %d width = %d, want 7 — got %q", i, rw, r)
		}
	}
}

func TestBigTextSlimColorAppliedToEachRow(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "X", Font: components.FontSlim, Color: styles.Magenta})
	if len(rows) != 4 {
		t.Fatalf("slim rows = %d, want 4", len(rows))
	}
	for i, r := range rows {
		if !strings.Contains(r, "\x1b[35m") {
			t.Fatalf("slim row %d missing magenta ANSI: %q", i, r)
		}
	}
}

// --- Digital font ----------------------------------------------------------

func TestBigTextDigitalFontDimensions(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "A", Font: components.FontDigital})
	if len(rows) != 5 {
		t.Fatalf("digital rows = %d, want 5", len(rows))
	}
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 4 {
			t.Fatalf("digital row %d width = %d, want 4 — got %q", i, rw, r)
		}
	}
}

func TestBigTextDigitalSupportsAllAlphanumericPlusSpace(t *testing.T) {
	const all = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, r := range all {
		rows := bigTextRows(t, components.BigTextProps{Text: string(r), Font: components.FontDigital})
		if len(rows) != 5 {
			t.Fatalf("digital rune=%q rendered %d rows, want 5", r, len(rows))
		}
		for i, row := range rows {
			if rw := len([]rune(row)); rw != 4 {
				t.Fatalf("digital rune=%q row %d width = %d, want 4 — got %q", r, i, rw, row)
			}
		}
		joined := strings.Join(rows, "")
		if strings.TrimSpace(joined) == "" {
			t.Fatalf("digital rune=%q rendered as blank", r)
		}
		if !strings.ContainsRune(joined, '█') {
			t.Fatalf("digital rune=%q lost all █ segment runes: %q", r, joined)
		}
	}
	rows := bigTextRows(t, components.BigTextProps{Text: " ", Font: components.FontDigital})
	if strings.TrimSpace(strings.Join(rows, "")) != "" {
		t.Fatalf("digital space rendered ink: %q", rows)
	}
}

func TestBigTextDigitalMultiCharacterComposition(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "12", Font: components.FontDigital})
	if len(rows) != 5 {
		t.Fatalf("digital 12 rows = %d, want 5", len(rows))
	}
	// Each glyph is 4 cols, gap is 1 col ⇒ 4+1+4 = 9.
	for i, r := range rows {
		if rw := len([]rune(r)); rw != 9 {
			t.Fatalf("digital 12 row %d width = %d, want 9 — got %q", i, rw, r)
		}
	}
}

func TestBigTextDigitalColorAppliedToEachRow(t *testing.T) {
	rows := bigTextRows(t, components.BigTextProps{Text: "0", Font: components.FontDigital, Color: styles.Red})
	if len(rows) != 5 {
		t.Fatalf("digital rows = %d, want 5", len(rows))
	}
	for i, r := range rows {
		if !strings.Contains(r, "\x1b[31m") {
			t.Fatalf("digital row %d missing red ANSI: %q", i, r)
		}
	}
}
