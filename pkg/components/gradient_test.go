package components_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func gradientText(t *testing.T, p components.GradientProps) string {
	t.Helper()
	node := components.Gradient(p)
	if node == nil {
		t.Fatal("Gradient returned nil node")
	}
	if node.Type != vdom.TextNode {
		t.Fatalf("Gradient returned non-text node (type=%v)", node.Type)
	}
	return node.Text
}

func TestGradientEmptyText(t *testing.T) {
	got := gradientText(t, components.GradientProps{From: [3]uint8{0, 0, 0}, To: [3]uint8{255, 255, 255}})
	if got != "" {
		t.Fatalf("expected empty output for empty text, got %q", got)
	}
}

func TestGradientSingleRune(t *testing.T) {
	got := gradientText(t, components.GradientProps{
		Text: "X",
		From: [3]uint8{10, 20, 30},
		To:   [3]uint8{200, 100, 50},
	})
	want := "\x1b[38;2;10;20;30m" + "X" + styles.Reset()
	if got != want {
		t.Fatalf("single-rune mismatch.\n got:  %q\n want: %q", got, want)
	}
}

func TestGradientSameFromAndToStillColors(t *testing.T) {
	got := gradientText(t, components.GradientProps{
		Text: "ab",
		From: [3]uint8{50, 60, 70},
		To:   [3]uint8{50, 60, 70},
	})
	// Both runes should carry the same SGR color and a reset.
	const seq = "\x1b[38;2;50;60;70m"
	if strings.Count(got, seq) != 2 {
		t.Fatalf("expected SGR %q twice, got %q", seq, got)
	}
	if strings.Count(got, styles.Reset()) != 2 {
		t.Fatalf("expected two ANSI resets in %q", got)
	}
	if !strings.Contains(got, "a") || !strings.Contains(got, "b") {
		t.Fatalf("missing visible characters in %q", got)
	}
}

func TestGradientInterpolatesEndpoints(t *testing.T) {
	from := [3]uint8{0, 0, 0}
	to := [3]uint8{255, 255, 255}
	got := gradientText(t, components.GradientProps{Text: "ABC", From: from, To: to})

	// First rune must use From verbatim, last rune must use To verbatim.
	firstSGR := fmt.Sprintf("\x1b[38;2;%d;%d;%dmA", from[0], from[1], from[2])
	lastSGR := fmt.Sprintf("\x1b[38;2;%d;%d;%dmC", to[0], to[1], to[2])
	if !strings.Contains(got, firstSGR) {
		t.Fatalf("expected first-rune sequence %q in %q", firstSGR, got)
	}
	if !strings.Contains(got, lastSGR) {
		t.Fatalf("expected last-rune sequence %q in %q", lastSGR, got)
	}

	// Middle rune (index 1 of 3) sits at t=0.5 → channels round to 128
	// (255*0.5 = 127.5 → +0.5 → 128).
	midSGR := "\x1b[38;2;128;128;128mB"
	if !strings.Contains(got, midSGR) {
		t.Fatalf("expected midpoint sequence %q in %q", midSGR, got)
	}
}

func TestGradientPerRuneResetsApplied(t *testing.T) {
	got := gradientText(t, components.GradientProps{
		Text: "hi!",
		From: [3]uint8{255, 0, 0},
		To:   [3]uint8{0, 0, 255},
	})
	// One reset per rune.
	if want := 3; strings.Count(got, styles.Reset()) != want {
		t.Fatalf("expected %d resets, got %d in %q", want, strings.Count(got, styles.Reset()), got)
	}
	// Also one foreground SGR introducer per rune.
	if want := 3; strings.Count(got, "\x1b[38;2;") != want {
		t.Fatalf("expected %d truecolor SGRs, got %d in %q", want, strings.Count(got, "\x1b[38;2;"), got)
	}
}

func TestGradientDescendingChannels(t *testing.T) {
	// From > To exercises the negative-delta branch in lerpChannel.
	got := gradientText(t, components.GradientProps{
		Text: "ab",
		From: [3]uint8{200, 100, 50},
		To:   [3]uint8{0, 0, 0},
	})
	// Endpoints again: a=From, b=To.
	if !strings.Contains(got, "\x1b[38;2;200;100;50ma") {
		t.Fatalf("missing From endpoint in %q", got)
	}
	if !strings.Contains(got, "\x1b[38;2;0;0;0mb") {
		t.Fatalf("missing To endpoint in %q", got)
	}
}

func TestGradientPreservesUnicodeRunes(t *testing.T) {
	got := gradientText(t, components.GradientProps{
		Text: "한글",
		From: [3]uint8{255, 0, 0},
		To:   [3]uint8{0, 255, 0},
	})
	if !strings.Contains(got, "한") || !strings.Contains(got, "글") {
		t.Fatalf("multi-byte runes lost: %q", got)
	}
	// Two runes → two color groups.
	if strings.Count(got, styles.Reset()) != 2 {
		t.Fatalf("expected 2 resets for 2 runes, got %q", got)
	}
}
