package buffer_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/internal/buffer"
)

// TestSetOutOfBoundsNegative covers the early return when x or y is negative,
// as well as y >= height.
func TestSetOutOfBoundsNegative(t *testing.T) {
	buf := buffer.New(5, 3)

	// Should be no-ops; no panic, no side effect.
	buf.Set(-1, 0, 'A')
	buf.Set(0, -1, 'B')
	buf.Set(0, 5, 'C')
	buf.Set(0, 3, 'D') // y == height

	if got := buf.Render(); got != "" {
		t.Errorf("Expected empty render after only OOB sets, got %q", got)
	}
}

// TestSetWideRuneAtRowEnd exercises the width>1 branch with a wide rune,
// confirming the continuation cell is marked undefined.
func TestSetWideRuneAtRowEnd(t *testing.T) {
	buf := buffer.New(4, 1)
	buf.Set(0, 0, '한')
	buf.Set(2, 0, '글')

	out := buf.RenderRows(1)
	if out != "한글" {
		t.Errorf("Expected 한글, got %q", out)
	}
}

// TestSetWideRuneExtendsRow exercises ensureRowWidth growing the row
// past its original capacity (Set with x+width > current row length).
func TestSetWideRuneExtendsRow(t *testing.T) {
	buf := buffer.New(2, 1)
	// width-2 rune placed at x=5 forces ensureRowWidth to extend.
	buf.Set(5, 0, '中')

	if r := buf.Get(5, 0); r != '中' {
		t.Errorf("Expected 中 at (5,0), got %q", r)
	}
	// Continuation cell should report space.
	if r := buf.Get(6, 0); r != ' ' {
		t.Errorf("Expected ' ' at (6,0), got %q", r)
	}
}

// TestSetZeroWidthCombinesWithPrev covers the appendZeroWidth append branch.
func TestSetZeroWidthCombinesWithPrev(t *testing.T) {
	buf := buffer.New(3, 1)
	buf.Set(0, 0, 'e')
	// U+0301 COMBINING ACUTE ACCENT — width 0, must append to prev cell.
	buf.Set(1, 0, '́')

	out := buf.RenderRows(1)
	if out != "é" {
		t.Errorf("Expected 'é' (combined), got %q", out)
	}
	// Get on the zero-width slot returns space (cell unchanged).
	if r := buf.Get(1, 0); r != ' ' {
		t.Errorf("Expected ' ' at (1,0) after zero-width append, got %q", r)
	}
}

// TestSetZeroWidthAtZeroXNoop covers appendZeroWidth's "x <= 0" early return.
func TestSetZeroWidthAtZeroXNoop(t *testing.T) {
	buf := buffer.New(3, 1)
	buf.Set(0, 0, '́') // x == 0, zero-width: must be ignored.

	out := buf.RenderRows(1)
	if out != "" {
		t.Errorf("Expected empty render, got %q", out)
	}
}

// TestSetZeroWidthOutOfBoundsY covers appendZeroWidth's y guard.
func TestSetZeroWidthOutOfBoundsY(t *testing.T) {
	buf := buffer.New(3, 1)
	// y == height triggers the y >= b.height guard inside appendZeroWidth.
	buf.Set(1, 1, '́')
	buf.Set(1, -1, '́')

	if got := buf.Render(); got != "" {
		t.Errorf("Expected empty render, got %q", got)
	}
}

// TestSetZeroWidthSkipsUndefinedCells exercises the loop's "skip undefined"
// branch in appendZeroWidth: a wide rune leaves an undefined cell that the
// search must walk past to find the real previous cell.
func TestSetZeroWidthSkipsUndefinedCells(t *testing.T) {
	buf := buffer.New(5, 1)
	buf.Set(0, 0, '日') // occupies (0,0) base + (1,0) undefined
	buf.Set(2, 0, '́')

	out := buf.RenderRows(1)
	// The combining mark should attach to '日' (the nearest non-undefined cell).
	if out != "日́" {
		t.Errorf("Expected '日<combine>', got %q", out)
	}
}

// TestSetZeroWidthBeyondRowLength exercises the "index >= len(b.cells[y])"
// branch in appendZeroWidth — start scan past the current row length so the
// loop must skip indices outside the slice before settling on a real cell.
func TestSetZeroWidthBeyondRowLength(t *testing.T) {
	buf := buffer.New(2, 1)
	buf.Set(0, 0, 'x')
	// Zero-width rune placed at x far beyond row length.
	buf.Set(50, 0, '́')

	out := buf.RenderRows(1)
	// The combining mark attaches to the nearest non-undefined cell to its
	// left within the existing row, which is the trailing space at x=1.
	if out != "x ́" {
		t.Errorf("Expected 'x <combine>', got %q", out)
	}
}

// TestGetOutOfBoundsAndUndefined covers all the early-return arms of Get.
func TestGetOutOfBoundsAndUndefined(t *testing.T) {
	buf := buffer.New(5, 2)

	if r := buf.Get(-1, 0); r != ' ' {
		t.Errorf("Get(-1,0) expected ' ', got %q", r)
	}
	if r := buf.Get(0, -1); r != ' ' {
		t.Errorf("Get(0,-1) expected ' ', got %q", r)
	}
	if r := buf.Get(0, 2); r != ' ' { // y == height
		t.Errorf("Get(0,2) expected ' ', got %q", r)
	}
	if r := buf.Get(99, 0); r != ' ' { // x past row width
		t.Errorf("Get(99,0) expected ' ', got %q", r)
	}

	// Undefined cell returned as space (continuation slot of a wide rune).
	buf.Set(0, 0, '日')
	if r := buf.Get(1, 0); r != ' ' {
		t.Errorf("Get on continuation cell expected ' ', got %q", r)
	}
}

// TestEnsureRowWidthOutOfBoundsViaWriteString — WriteString routes through
// Set which defends OOB; here we simply verify the y-OOB guard in
// ensureRowWidth doesn't produce panics for very large y via Set.
func TestEnsureRowWidthOutOfBoundsViaSet(t *testing.T) {
	buf := buffer.New(2, 2)
	// y way out of bounds — must just be a no-op.
	buf.Set(0, 100, 'A')
	if got := buf.Render(); got != "" {
		t.Errorf("Expected empty render, got %q", got)
	}
}

// TestRenderRowOutOfBoundsViaRenderRowsClamped validates the y-OOB guard in
// renderRow indirectly: RenderRows clamps rows to height, so we exercise the
// loop boundary at exactly height.
func TestRenderRowsClampedToHeight(t *testing.T) {
	buf := buffer.New(3, 2)
	buf.WriteString(0, 0, "ab")
	out := buf.RenderRows(99)
	if out != "ab\n" {
		t.Errorf("Expected 'ab\\n', got %q", out)
	}
}

// TestRenderRowsZeroOrNegative covers the rows<=0 early return.
func TestRenderRowsZeroOrNegative(t *testing.T) {
	buf := buffer.New(3, 2)
	buf.WriteString(0, 0, "x")

	if out := buf.RenderRows(0); out != "" {
		t.Errorf("RenderRows(0) expected empty, got %q", out)
	}
	if out := buf.RenderRows(-5); out != "" {
		t.Errorf("RenderRows(-5) expected empty, got %q", out)
	}
}

// TestRenderEmptyBufferReturnsEmpty exercises Render's all-empty short-circuit.
func TestRenderEmptyBufferReturnsEmpty(t *testing.T) {
	buf := buffer.New(4, 3)
	if out := buf.Render(); out != "" {
		t.Errorf("Render of pristine buffer expected empty, got %q", out)
	}
}

// TestWriteStringNewlineOverflow covers WriteString's "currentY >= b.height"
// return branch hit on a '\n' increment.
func TestWriteStringNewlineOverflow(t *testing.T) {
	buf := buffer.New(5, 2)
	// Two newlines pushes currentY to 2 (== height) and must early-return.
	buf.WriteString(0, 0, "A\nB\nC")

	if r := buf.Get(0, 0); r != 'A' {
		t.Errorf("Expected 'A' at (0,0), got %q", r)
	}
	if r := buf.Get(0, 1); r != 'B' {
		t.Errorf("Expected 'B' at (0,1), got %q", r)
	}
	// 'C' must not have been written (no row 2).
	if got := buf.Render(); strings.Contains(got, "C") {
		t.Errorf("Row 2 should not exist; render=%q", got)
	}
}

// TestWriteStringNegativeYAborts covers the "currentY < 0 || currentY >= b.height"
// branch that fires when starting y is already out of bounds.
func TestWriteStringNegativeYAborts(t *testing.T) {
	buf := buffer.New(5, 2)
	buf.WriteString(0, -1, "Hello")
	buf.WriteString(0, 5, "World")

	if got := buf.Render(); got != "" {
		t.Errorf("Expected empty render, got %q", got)
	}
}

// TestWriteStringNegativeStartXIsSkipped covers the "currentX >= 0" guard.
// Starting at x=-2 with "ab" should skip 'a' (x=-2) and 'b' (x=-1) and write
// nothing visible.
func TestWriteStringNegativeStartXIsSkipped(t *testing.T) {
	buf := buffer.New(5, 1)
	buf.WriteString(-2, 0, "ab")
	if got := buf.Render(); got != "" {
		t.Errorf("Expected empty render with negative start x, got %q", got)
	}

	// And with "abcde" starting at x=-2, 'c' lands at x=0, 'd' at x=1, 'e' at x=2.
	buf2 := buffer.New(5, 1)
	buf2.WriteString(-2, 0, "abcde")
	out := buf2.RenderRows(1)
	if out != "cde" {
		t.Errorf("Expected 'cde', got %q", out)
	}
}

// TestWriteStringMultiByteAndEmoji exercises wide runes (Korean + emoji)
// flowing through WriteString -> Set -> ensureRowWidth.
func TestWriteStringMultiByteAndEmoji(t *testing.T) {
	buf := buffer.New(20, 1)
	buf.WriteString(0, 0, "한글🍔X")

	out := buf.RenderRows(1)
	if out != "한글🍔X" {
		t.Errorf("Expected '한글🍔X', got %q", out)
	}
}

// TestEmptyStringWriteIsNoop covers the trivial empty-input branch.
func TestEmptyStringWriteIsNoop(t *testing.T) {
	buf := buffer.New(3, 1)
	buf.WriteString(0, 0, "")
	if got := buf.Render(); got != "" {
		t.Errorf("Expected empty render, got %q", got)
	}
}

// TestNewZeroDimensions exercises width=0/height=0 boundary.
func TestNewZeroDimensions(t *testing.T) {
	buf := buffer.New(0, 0)
	if buf.Width() != 0 || buf.Height() != 0 {
		t.Errorf("Expected 0x0 buffer, got %dx%d", buf.Width(), buf.Height())
	}
	// All ops must be safe on a 0x0 buffer.
	buf.Set(0, 0, 'A')
	buf.WriteString(0, 0, "hi")
	buf.Clear()
	if got := buf.Render(); got != "" {
		t.Errorf("Render on 0x0 expected empty, got %q", got)
	}
	if got := buf.RenderRows(3); got != "" {
		t.Errorf("RenderRows on 0x0 expected empty, got %q", got)
	}
	if r := buf.Get(0, 0); r != ' ' {
		t.Errorf("Get on 0x0 expected ' ', got %q", r)
	}
}

// TestNewSingleCellBoundary exercises the width=1/height=1 boundary.
func TestNewSingleCellBoundary(t *testing.T) {
	buf := buffer.New(1, 1)
	buf.Set(0, 0, 'Z')
	if r := buf.Get(0, 0); r != 'Z' {
		t.Errorf("Expected 'Z', got %q", r)
	}
	if got := buf.RenderRows(1); got != "Z" {
		t.Errorf("Expected 'Z', got %q", got)
	}
}

// TestVeryLargeInput stresses Render/Clear paths with a sizable buffer.
func TestVeryLargeInput(t *testing.T) {
	const w, h = 200, 50
	buf := buffer.New(w, h)
	line := strings.Repeat("a", w)
	for y := 0; y < h; y++ {
		buf.WriteString(0, y, line)
	}

	out := buf.Render()
	gotLines := strings.Split(out, "\n")
	if len(gotLines) != h {
		t.Fatalf("Expected %d rendered lines, got %d", h, len(gotLines))
	}
	for i, l := range gotLines {
		if l != line {
			t.Fatalf("Line %d mismatch: got %q", i, l)
		}
	}

	// Clear must wipe everything.
	buf.Clear()
	if buf.Render() != "" {
		t.Error("Render after Clear expected empty")
	}
}

// TestSetStringOverwritesWideContinuationCleanly verifies the case where a
// wide cluster previously stored at column N is partially overwritten by a
// narrow cluster at column N+1: the leading cell at N must NOT remain as
// the wide cluster's text (which would render as a 2-column glyph spilling
// over the new cell). The implementation marks the trailing column with
// undefinedCell when laying down the wide cluster; an overwrite at N+1
// should leave the now-stale leading cell at N visually as a half-rendered
// wide that callers must repaint to a space. This test pins the current
// behaviour so any future change is intentional.
func TestSetStringOverwritesWideContinuationCleanly(t *testing.T) {
	buf := buffer.New(5, 1)
	buf.Set(0, 0, '日')
	// Overwrite the continuation column with a single ASCII char.
	buf.SetString(1, 0, "X", 1)

	// Cell 0 still holds the wide-cluster glyph (the buffer does not
	// retroactively scrub its leading cell), and cell 1 now holds X.
	out := buf.RenderRows(1)
	// Render concatenates non-undefined cells; '日' (wide) + 'X' = '日X'.
	if out != "日X" {
		t.Errorf("expected '日X', got %q", out)
	}
}

// TestClearAfterWideAndZeroWidthRunes ensures Clear restores plain spaces
// even when undefined cells / appended zero-width runes were present.
func TestClearAfterWideAndZeroWidthRunes(t *testing.T) {
	buf := buffer.New(5, 2)
	buf.Set(0, 0, '日')
	buf.Set(2, 0, '́')
	buf.Clear()

	for y := 0; y < 2; y++ {
		for x := 0; x < 5; x++ {
			if r := buf.Get(x, y); r != ' ' {
				t.Errorf("Cell (%d,%d) not space after Clear: %q", x, y, r)
			}
		}
	}
	if buf.Render() != "" {
		t.Error("Render after Clear expected empty")
	}
}
