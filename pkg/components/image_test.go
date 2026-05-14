package components_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
)

const upperHalfBlock = "▀"

func imageText(t *testing.T, p components.ImageProps) string {
	t.Helper()
	node := components.Image(p)
	if node == nil {
		t.Fatalf("Image returned nil")
	}
	if node.ElementType != "image" {
		t.Fatalf("Image element type = %q, want \"image\"", node.ElementType)
	}
	if len(node.Children) != 1 {
		t.Fatalf("Image produced %d children, want 1", len(node.Children))
	}
	return node.Children[0].Text
}

func TestImageNilProducesEmptyNode(t *testing.T) {
	got := imageText(t, components.ImageProps{Image: nil})
	if got != "" {
		t.Fatalf("nil image should render empty text, got %q", got)
	}
}

func TestImageZeroWidthProducesEmpty(t *testing.T) {
	got := imageText(t, components.ImageProps{Image: &components.ImageData{Width: 0, Height: 4}})
	if got != "" {
		t.Fatalf("zero-width image should render empty, got %q", got)
	}
}

func TestImageZeroHeightProducesEmpty(t *testing.T) {
	got := imageText(t, components.ImageProps{Image: &components.ImageData{Width: 2, Height: 0}})
	if got != "" {
		t.Fatalf("zero-height image should render empty, got %q", got)
	}
}

func TestImage1x2RedOverBlue(t *testing.T) {
	img := &components.ImageData{
		Width:  1,
		Height: 2,
		Pixels: [][3]uint8{
			{255, 0, 0}, // top: red
			{0, 0, 255}, // bottom: blue
		},
	}
	got := imageText(t, components.ImageProps{Image: img})

	wantFG := "\x1b[38;2;255;0;0m"
	wantBG := "\x1b[48;2;0;0;255m"
	if !strings.Contains(got, wantFG) {
		t.Fatalf("expected red foreground %q in %q", wantFG, got)
	}
	if !strings.Contains(got, wantBG) {
		t.Fatalf("expected blue background %q in %q", wantBG, got)
	}
	if !strings.Contains(got, upperHalfBlock) {
		t.Fatalf("expected upper-half-block glyph in %q", got)
	}
	if strings.Contains(got, "\n") {
		t.Fatalf("single-row output should not contain newline: %q", got)
	}
}

func TestImage2x2ProducesSingleRowTwoGlyphs(t *testing.T) {
	img := &components.ImageData{
		Width:  2,
		Height: 2,
		Pixels: [][3]uint8{
			{1, 2, 3}, {4, 5, 6},
			{7, 8, 9}, {10, 11, 12},
		},
	}
	got := imageText(t, components.ImageProps{Image: img})

	if got == "" {
		t.Fatalf("expected non-empty output")
	}
	if strings.Contains(got, "\n") {
		t.Fatalf("2x2 image collapses to one half-block row, should have no newline: %q", got)
	}
	if count := strings.Count(got, upperHalfBlock); count != 2 {
		t.Fatalf("expected 2 half-block glyphs, got %d in %q", count, got)
	}
	// Verify the second pixel column carries the right top FG color.
	if !strings.Contains(got, "\x1b[38;2;4;5;6m") {
		t.Fatalf("missing top color for second column: %q", got)
	}
	// Verify the second pixel column's bottom row uses (10,11,12) as BG.
	if !strings.Contains(got, "\x1b[48;2;10;11;12m") {
		t.Fatalf("missing bottom color for second column: %q", got)
	}
}

func TestImageOddHeightPadsLastRowWithBlack(t *testing.T) {
	img := &components.ImageData{
		Width:  1,
		Height: 3,
		Pixels: [][3]uint8{
			{10, 20, 30},
			{40, 50, 60},
			{70, 80, 90}, // alone in last half-block row, paired with black.
		},
	}
	got := imageText(t, components.ImageProps{Image: img})

	// Two output half-rows -> exactly one newline between them.
	if count := strings.Count(got, "\n"); count != 1 {
		t.Fatalf("expected 1 newline (2 half-rows), got %d in %q", count, got)
	}

	// Last row's top FG must be (70,80,90) and BG must be (0,0,0).
	wantTop := "\x1b[38;2;70;80;90m"
	wantBottom := "\x1b[48;2;0;0;0m"
	idx := strings.LastIndex(got, wantTop)
	if idx < 0 {
		t.Fatalf("expected top FG %q in %q", wantTop, got)
	}
	tail := got[idx:]
	if !strings.Contains(tail, wantBottom) {
		t.Fatalf("expected black BG pad %q after last top color in tail %q", wantBottom, tail)
	}
}

func TestImageMultiRowHasCorrectNewlineCount(t *testing.T) {
	// 4x4 (4 rows -> 2 half-block rows -> 1 newline).
	img := &components.ImageData{Width: 4, Height: 4, Pixels: solidRaster(4, 4, [3]uint8{255, 255, 255})}
	got := imageText(t, components.ImageProps{Image: img})
	if count := strings.Count(got, "\n"); count != 1 {
		t.Fatalf("expected 1 newline for 4x4, got %d", count)
	}
	if count := strings.Count(got, upperHalfBlock); count != 8 {
		t.Fatalf("expected 8 glyphs (4 cols * 2 rows), got %d", count)
	}
}

func TestImageEncodingPerCellShape(t *testing.T) {
	img := &components.ImageData{
		Width:  1,
		Height: 2,
		Pixels: [][3]uint8{{1, 2, 3}, {4, 5, 6}},
	}
	got := imageText(t, components.ImageProps{Image: img})
	want := fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%s\x1b[0m", 1, 2, 3, 4, 5, 6, upperHalfBlock)
	if got != want {
		t.Fatalf("unexpected encoding\n got: %q\nwant: %q", got, want)
	}
}

func solidRaster(w, h int, c [3]uint8) [][3]uint8 {
	pixels := make([][3]uint8, w*h)
	for i := range pixels {
		pixels[i] = c
	}
	return pixels
}
