package ink

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// TestCellLevelDiffWideShiftRepaintsTrailingColumn covers a wide cluster
// moving by one column between frames while the row's total visible width
// is unchanged. The cells that flip must be repainted; any stale trailing
// column adjacent to a wide-cluster move must not be left behind.
//
// prev "X我X" → cells [X][我-lead][我-cont][X]
// next "XX我" → cells [X][X][我-lead][我-cont]
//
// The diff must repaint at column 1 (X) and column 2 (我). The wide write
// at column 2 spans columns 2+3, overwriting prev's wide-cont and X.
func TestCellLevelDiffWideShiftRepaintsTrailingColumn(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("X我X\n")
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
		return vdom.CreateTextNode("XX我\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}

	if !strings.Contains(update, "我") {
		t.Fatalf("expected wide cluster in diff payload, got %q", update)
	}
	if !strings.Contains(update, "X") {
		t.Fatalf("expected new X cell in diff payload, got %q", update)
	}
}

// TestCellLevelDiffWideToNarrowAtTrailingColumn covers prev wide → next
// narrow at the same span. Both columns of the prev wide must be
// independently repainted with their new content.
func TestCellLevelDiffWideToNarrowAtTrailingColumn(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("AB我C\n")
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
		return vdom.CreateTextNode("ABDEC\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}
	if !strings.Contains(update, "D") || !strings.Contains(update, "E") {
		t.Fatalf("expected both narrow cells D and E in payload, got %q", update)
	}
}

// TestCellLevelDiffNarrowToWideAtSameColumn covers prev narrow → next wide
// at the same span; total widths unchanged.
func TestCellLevelDiffNarrowToWideAtSameColumn(t *testing.T) {
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		return vdom.CreateTextNode("AABCC\n")
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
		return vdom.CreateTextNode("AA我CC\n")
	}); err != nil {
		t.Fatalf("rerender failed: %v", err)
	}

	update := firstRerenderWrite(stdout, mountWrites)
	if update == "" {
		t.Fatalf("expected a rerender write, got %#v", stdout.writes)
	}
	if !strings.Contains(update, "我") {
		t.Fatalf("expected wide cluster in payload, got %q", update)
	}
}

// TestParseFrameWideRuneEmitsContinuationCell exercises the parser's
// internal cluster → cell mapping for a wide cluster. The continuation
// cell must be present (width=0, empty text) and the leading cell must
// carry the full cluster text and width=2.
func TestParseFrameWideRuneEmitsContinuationCell(t *testing.T) {
	frame := parseFrame("A我B\n")
	if !frame.parseOK {
		t.Fatalf("parseFrame failed unexpectedly")
	}
	if len(frame.rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(frame.rows))
	}
	row := frame.rows[0]
	if len(row) != 4 {
		t.Fatalf("expected 4 cells in row, got %d: %#v", len(row), row)
	}
	if row[0].text != "A" || row[0].width != 1 {
		t.Fatalf("cell[0] expected A,1 got %#v", row[0])
	}
	if row[1].text != "我" || row[1].width != 2 {
		t.Fatalf("cell[1] expected 我,2 got %#v", row[1])
	}
	if row[2].text != "" || row[2].width != 0 {
		t.Fatalf("cell[2] expected continuation got %#v", row[2])
	}
	if row[3].text != "B" || row[3].width != 1 {
		t.Fatalf("cell[3] expected B,1 got %#v", row[3])
	}
}

// TestParseFrameZWJEmojiEmitsSingleCluster verifies that a ZWJ family
// emoji parses as one width-2 leading cell + one continuation cell — not
// split mid-cluster.
func TestParseFrameZWJEmojiEmitsSingleCluster(t *testing.T) {
	family := "\U0001F468‍\U0001F469‍\U0001F467‍\U0001F466"
	frame := parseFrame(family + "\n")
	if !frame.parseOK {
		t.Fatalf("parseFrame failed unexpectedly")
	}
	if len(frame.rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(frame.rows))
	}
	row := frame.rows[0]
	if len(row) != 2 {
		t.Fatalf("expected 2 cells (lead + cont), got %d: %#v", len(row), row)
	}
	if row[0].text != family || row[0].width != 2 {
		t.Fatalf("cell[0] expected family cluster width 2, got %#v", row[0])
	}
	if row[1].text != "" || row[1].width != 0 {
		t.Fatalf("cell[1] expected continuation, got %#v", row[1])
	}
}

// TestRowVisibleWidthCountsContinuationCellsAsZero exercises the row-width
// helper's invariant: continuation cells have width 0 and so the visible
// width of `A我B` is 4 (1+2+0+1), matching the rendered column count.
func TestRowVisibleWidthCountsContinuationCellsAsZero(t *testing.T) {
	frame := parseFrame("A我B\n")
	if !frame.parseOK {
		t.Fatalf("parseFrame failed unexpectedly")
	}
	if got := rowVisibleWidth(frame.rows[0]); got != 4 {
		t.Fatalf("expected row width 4, got %d", got)
	}
}
