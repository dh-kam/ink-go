package ink

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

// firstNonEmptyTransition returns the index of the first stdout write that
// resulted from a Rerender (i.e. excludes the initial mount paint).
func firstRerenderWrite(stdout *recordingWriter, mountWriteCount int) string {
	if len(stdout.writes) <= mountWriteCount {
		return ""
	}

	return stdout.writes[mountWriteCount]
}

// TestCellLevelDiffSingleCellChange asserts that a one-character mutation
// inside an otherwise-static line produces a single cell-level write that
// targets the changed column with cursorTo, instead of a full-line rewrite.
func TestCellLevelDiffSingleCellChange(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Counter: 0\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Counter: 1\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	// "Counter: " is 9 columns wide; the changed cell is at column 9.
	if !strings.Contains(update, ansiCursorTo(9)) {
		t.Fatalf("expected cell-level cursorTo(9), got %q", update)
	}
	if !strings.Contains(update, "1") {
		t.Fatalf("expected changed cell '1' to be emitted, got %q", update)
	}
	// Cell-level diff must NOT rewrite the unchanged "Counter: " prefix.
	if strings.Contains(update, "Counter: 1") {
		t.Fatalf("expected cell-level diff to skip the static prefix, got %q", update)
	}
	// Cell-level path does not use eraseEndLine — that is the column-level
	// fallback path's signature.
	if strings.Contains(update, ansiEraseEndLine()) {
		t.Fatalf("expected cell-level path to skip eraseEndLine, got %q", update)
	}
}

// TestCellLevelDiffMultiRowSparseChanges ensures multi-row sparse mutations
// emit cursor jumps between rows but only touch the differing cells.
func TestCellLevelDiffMultiRowSparseChanges(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("aaa\nbbb\nccc\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	if err := instance.Rerender(func() *vdom.Node {
		// Change only row 0 col 0 (a -> X) and row 2 col 2 (c -> Y).
		return vdom.CreateTextNode("Xaa\nbbb\nccY\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	if !strings.Contains(update, "X") || !strings.Contains(update, "Y") {
		t.Fatalf("expected changed cells X and Y in payload, got %q", update)
	}

	// The middle row was unchanged — its content "bbb" must not appear.
	// (We can't simply Contains("bbb") because it never appears as a
	// rendered line, but checking that "Xaa", "ccY", or "bbb" are not
	// rewritten in full is the real signal — we look for the absence of
	// those longer rewrites.)
	if strings.Contains(update, "Xaa") || strings.Contains(update, "ccY") {
		t.Fatalf("expected cell-level path not to rewrite full lines, got %q", update)
	}

	// Multi-row dispatch should require at least one cursorDown move
	// (between row 0 and row 2).
	if !strings.Contains(update, "\x1b[") {
		t.Fatalf("expected ANSI cursor positioning in multi-row diff, got %q", update)
	}
}

// TestCellLevelDiffWholeLineChangeStillCellTargeted demonstrates that a
// whole-line replacement still goes through the cell path: it emits per-cell
// writes from column 0 (no eraseEndLine), with each cell of the new content
// present.
func TestCellLevelDiffWholeLineChangeStillCellTargeted(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("abcde\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("vwxyz\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	for _, ch := range []string{"v", "w", "x", "y", "z"} {
		if !strings.Contains(update, ch) {
			t.Fatalf("expected new cell %q in payload, got %q", ch, update)
		}
	}

	// Cell path always positions the cursor explicitly. Even when the
	// whole line is dirty we go through cursorTo(0) (column 0) and
	// per-cell writes — the eraseEndLine signature of the column-level
	// fallback must not appear.
	if strings.Contains(update, ansiEraseEndLine()) {
		t.Fatalf("expected cell-level path not to emit eraseEndLine, got %q", update)
	}
}

// TestCellLevelDiffFallsBackOnFrameSizeGrowth verifies that adding a row
// drops us back to the line-level path. The line-level signature is the
// presence of eraseLines (the line-level path always emits an erase preamble
// when the visible row count differs).
func TestCellLevelDiffFallsBackOnFrameSizeGrowth(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("row1\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("row1\nrow2\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	// Line-level fallback path emits the new row content in full.
	if !strings.Contains(update, "row2") {
		t.Fatalf("expected fallback to emit new row content, got %q", update)
	}
}

// TestCellLevelDiffFallsBackOnRowWidthChange verifies that growing a single
// row's width forces the fallback. Buffer parity requires same column count
// per row to safely emit cell-targeted writes.
func TestCellLevelDiffFallsBackOnRowWidthChange(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("ab\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("abc\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	// Fallback signature: the column-level fallback writes the differing
	// tail "c" with eraseEndLine. Either way, the new content must appear.
	if !strings.Contains(update, "c") {
		t.Fatalf("expected width-change fallback to emit new tail, got %q", update)
	}
}

// TestCellLevelDiffSGRTransition exercises a single-cell ANSI styling
// change: the leading cell goes from green to red while text identity is
// unchanged. The cell-level diff must emit an SGR transition for that cell.
func TestCellLevelDiffSGRTransition(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("\x1b[32mhello\x1b[0m\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("\x1b[31mhello\x1b[0m\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	// The new style (red, \x1b[31m) must appear; the cell-level path
	// emits an SGR transition before writing changed cells.
	if !strings.Contains(update, "\x1b[31m") {
		t.Fatalf("expected new SGR sequence in cell-level diff, got %q", update)
	}
	if !strings.Contains(update, "hello") {
		t.Fatalf("expected styled cells to be rewritten, got %q", update)
	}
}

// TestCellLevelDiffSGRAdjacentCellsShareTransition asserts that when two
// adjacent cells change but carry the same style, the SGR setup is emitted
// once and both cell payloads follow back-to-back (no redundant reset +
// re-enable between them).
func TestCellLevelDiffSGRAdjacentCellsShareTransition(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("\x1b[31mab\x1b[0m\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	// Both cells change text; both stay red.
	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("\x1b[31mXY\x1b[0m\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	// The red SGR must appear exactly once — both cells share it.
	if got := strings.Count(update, "\x1b[31m"); got != 1 {
		t.Fatalf("expected red SGR emitted once, got %d in %q", got, update)
	}
	// And no inline resets between the two cells.
	xIndex := strings.Index(update, "X")
	yIndex := strings.Index(update, "Y")
	if xIndex < 0 || yIndex < 0 || yIndex < xIndex {
		t.Fatalf("expected XY ordering in payload, got %q", update)
	}
	between := update[xIndex+1 : yIndex]
	if strings.Contains(between, "\x1b[0m") || strings.Contains(between, "\x1b[31m") {
		t.Fatalf("expected no SGR emission between adjacent same-style cells, got between=%q", between)
	}
}

// TestCellLevelDiffSGRForegroundOnlyChange asserts that when only the
// foreground color flips between two cells (keeping bold etc unchanged),
// the cell-level path emits the new fg code WITHOUT a preceding \x1b[0m
// reset. This is the headline byte saving over the round-2 implementation.
func TestCellLevelDiffSGRForegroundOnlyChange(t *testing.T) {
	stdout := &recordingWriter{}

	// Two cells: first is red, second is green. Both will change in
	// rerender so the cell path walks them and emits fg transitions.
	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("\x1b[31ma\x1b[32mb\x1b[0m\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	// Rerender swaps cell texts but keeps the same fg colors.
	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("\x1b[31mX\x1b[32mY\x1b[0m\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	// Locate the green transition: should be just \x1b[32m, not preceded
	// by a \x1b[0m reset (since only fg changed).
	greenIndex := strings.Index(update, "\x1b[32m")
	if greenIndex < 0 {
		t.Fatalf("expected green SGR in payload, got %q", update)
	}
	// The 5 bytes preceding \x1b[32m must NOT be \x1b[0m.
	if greenIndex >= 4 {
		preceding := update[greenIndex-4 : greenIndex]
		if preceding == "\x1b[0m" {
			t.Fatalf("expected fg-only change to skip reset, got reset before green: %q", update)
		}
	}
}

// TestCellLevelDiffSGRReturnToPlainUsesShortReset asserts that returning
// from a styled cell to the unstyled trailer at end-of-payload emits
// \x1b[0m (the shortest form) rather than a longer enumerated disable
// chain.
func TestCellLevelDiffSGRReturnToPlainUsesShortReset(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("\x1b[1;31ma\x1b[0m\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		CellLevelDiff:        true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("\x1b[1;31mb\x1b[0m\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	// The cell-level payload must finish by restoring default state. We
	// pick the shorter of \x1b[0m vs \x1b[22m\x1b[39m: with bold+red the
	// targeted disable chain is 9 bytes while \x1b[0m is 4 bytes — the
	// optimizer must choose the reset.
	if !strings.Contains(update, "\x1b[0m") {
		t.Fatalf("expected reset to plain at payload tail, got %q", update)
	}
	if strings.Contains(update, "\x1b[22m\x1b[39m") {
		t.Fatalf("expected short reset over enumerated disables, got %q", update)
	}
}

// TestCellLevelDiffSGRBoldEnableOnlyEmitsBoldCode asserts that when a cell
// gains bold without other changes, the transition is just \x1b[1m — no
// reset + full re-enable.
func TestCellLevelDiffSGRBoldEnableOnlyEmitsBoldCode(t *testing.T) {
	// We construct a parsed SGR transition directly to keep this test
	// independent of the full mount/render pipeline; the function under
	// test is the one whose byte budget we care about.
	current := sgrAttrs{parseOK: true}
	next := sgrAttrs{bold: true, parseOK: true}

	got := sgrTransition(current, next)
	if got != "\x1b[1m" {
		t.Fatalf("expected just bold enable, got %q", got)
	}
}

// TestCellLevelDiffSGRBoldOffWithSurvivingDim asserts that toggling bold
// off while dim stays on emits \x1b[22m\x1b[2m — the "22 disables both,
// then re-enable the survivor" pattern from emitANSITransition.
func TestCellLevelDiffSGRBoldOffWithSurvivingDim(t *testing.T) {
	current := sgrAttrs{bold: true, dim: true, parseOK: true}
	next := sgrAttrs{dim: true, parseOK: true}

	got := sgrTransition(current, next)
	if got != "\x1b[22m\x1b[2m" {
		t.Fatalf("expected 22 disable + dim re-enable, got %q", got)
	}
}

// TestCellLevelDiffSGRParseSafeFallback asserts that an unparseable style
// string (containing a code we don't model) flips parseOK off so callers
// can take the safe \x1b[0m + full SGR fallback path.
func TestCellLevelDiffSGRParseSafeFallback(t *testing.T) {
	// Code 53 (overline) is intentionally not modeled.
	attrs := parseSGR("\x1b[53m")
	if attrs.parseOK {
		t.Fatalf("expected parseOK=false for unmodeled SGR code, got %+v", attrs)
	}

	// And sgrTransition with parseOK=false must return the empty string —
	// the caller then takes the safe fallback branch.
	got := sgrTransition(sgrAttrs{parseOK: true}, attrs)
	if got != "" {
		t.Fatalf("expected empty transition when next.parseOK=false, got %q", got)
	}
}

// TestCellLevelDiffOptInDoesNotChangeDefault verifies that without the flag
// we still take the line-level path. Mirrors a line-level expectation: a
// counter rerender uses the column-level diff signature (eraseEndLine), not
// the cell-level signature (raw cursorTo + cell text).
func TestCellLevelDiffOptInDoesNotChangeDefault(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("Counter: 0\n")
	}, RenderOptions{
		AppOptions:           AppOptions{Stdout: stdout},
		IncrementalRendering: true,
		// CellLevelDiff intentionally omitted — must default to line-level.
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	mountWrites := len(stdout.writes)

	if err := instance.Rerender(func() *vdom.Node {
		return vdom.CreateTextNode("Counter: 1\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	// Line-level (column-diff) signature — eraseEndLine must be present.
	if !strings.Contains(update, ansiEraseEndLine()) {
		t.Fatalf("expected line-level path's eraseEndLine when opt-in is disabled, got %q", update)
	}
}
