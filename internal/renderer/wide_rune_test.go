package renderer_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/internal/renderer"
	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/utils"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// TestWideRuneClippedAtFixedBoxWidthNoOverflow verifies that a wide CJK
// cluster at the right edge of a fixed-width Box does not render past the
// box boundary. With width=3 and content "ABC我" the wide cluster cannot
// fit (would require column 3+4) so it must be either dropped or wrapped.
//
// With the default wrap mode "wrap", the wide cluster wraps to its own
// line. The first rendered line must have width <= 3.
func TestWideRuneClippedAtFixedBoxWidthNoOverflow(t *testing.T) {
	root := components.Box(vdom.Props{"width": 3.0},
		components.Text("ABC我"),
	)

	output := renderer.RenderWithLayout(root, 20, 20)
	lines := strings.Split(output, "\n")
	if len(lines) < 1 {
		t.Fatalf("expected at least one line, got %q", output)
	}
	if w := utils.StringWidth(lines[0]); w > 3 {
		t.Fatalf("first line width %d exceeds box width 3, line=%q", w, lines[0])
	}
}

// TestTruncateEndDropsTrailingWideRune verifies truncate-end of "abc我" at
// width 3 produces "abc" — the wide cluster is clipped, not painted past
// the boundary as "ab我" (which would be width 4).
func TestTruncateEndDropsTrailingWideRune(t *testing.T) {
	root := components.Box(vdom.Props{"width": 3.0},
		components.Text(vdom.Props{"wrap": "truncate-end"}, "abc我"),
	)

	output := renderer.RenderWithLayout(root, 20, 20)
	if got := strings.Split(output, "\n")[0]; utils.StringWidth(got) > 3 {
		t.Fatalf("truncate-end produced overflow %q (width %d)", got, utils.StringWidth(got))
	}
}

// TestZWJEmojiWrapsAtClusterBoundary verifies a ZWJ family emoji is treated
// as one width-2 cluster and never split mid-cluster across a wrap point.
// In a width-7 box the family + "wraph" fits on the first line (2+5=7) and
// the rest wraps. The ZWJ cluster must appear intact on whichever line it
// lands on.
func TestZWJEmojiWrapsAtClusterBoundary(t *testing.T) {
	family := "\U0001F468‍\U0001F469‍\U0001F467‍\U0001F466"
	text := family + "wraphere"

	root := components.Box(vdom.Props{"width": 7.0},
		components.Text(text),
	)

	output := renderer.RenderWithLayout(root, 20, 20)
	lines := strings.Split(output, "\n")

	// The family cluster must appear intact — joined by ZWJs (U+200D) — on
	// whichever line it lands on. We check that any line containing
	// any rune of the family contains the full ZWJ-joined sequence.
	foundIntact := false
	for _, line := range lines {
		if strings.Contains(line, family) {
			foundIntact = true
			break
		}
	}
	if !foundIntact {
		t.Fatalf("ZWJ family was split across wrap point; lines=%q", lines)
	}

	// And no line may exceed the box width.
	for i, line := range lines {
		if w := utils.StringWidth(line); w > 7 {
			t.Fatalf("line %d %q width %d > 7", i, line, w)
		}
	}
}

// TestStyledWideRuneKeepsStyleAcrossBothColumns verifies that ANSI styling
// applied to a wide CJK cluster covers the whole rendered cluster (i.e. the
// reset doesn't land between the two columns of the wide cluster).
func TestStyledWideRuneKeepsStyleAcrossBothColumns(t *testing.T) {
	// Box with a styled wide cluster.
	root := components.Box(vdom.Props{
		"alignSelf": "flex-start",
	}, components.Text(vdom.Props{"color": "red"}, "我"))

	output := renderer.RenderWithLayoutANSI(root, 20, 20)

	// The wide cluster must be wrapped by SGR open + close, not split.
	if !strings.Contains(output, "\x1b[31m我") {
		t.Fatalf("expected red-styled wide cluster, got %q", output)
	}
	// The default-color reset must come AFTER the wide cluster.
	idxRune := strings.Index(output, "我")
	idxReset := strings.Index(output, "\x1b[39m")
	if idxRune == -1 {
		t.Fatalf("wide cluster missing in output %q", output)
	}
	if idxReset != -1 && idxReset < idxRune {
		t.Fatalf("style reset appears before wide cluster — style split mid-cluster: %q", output)
	}
}

// TestWideRuneAtYogaShrunkBoundaryStaysWithinBox covers the case where a
// flex container shrinks below the text's natural width. The text inside
// must clip at the new width — including a wide cluster that would
// otherwise straddle the boundary.
func TestWideRuneAtYogaShrunkBoundaryStaysWithinBox(t *testing.T) {
	// Outer fixed at 4, inner text "ab我cd" — natural width 6.
	root := components.Box(vdom.Props{
		"width":     4.0,
		"alignSelf": "flex-start",
	}, components.Text(vdom.Props{"wrap": "truncate-end"}, "ab我cd"))

	output := renderer.RenderWithLayout(root, 20, 20)
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		t.Fatalf("expected rendered output, got empty")
	}
	if w := utils.StringWidth(lines[0]); w > 4 {
		t.Fatalf("yoga-shrunk wide-rune line %q width %d > 4", lines[0], w)
	}
}
