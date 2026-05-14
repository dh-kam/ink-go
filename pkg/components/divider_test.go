package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/styles"
)

func dividerText(t *testing.T, p components.DividerProps) string {
	t.Helper()
	node := components.Divider(p)
	if len(node.Children) != 1 {
		t.Fatalf("Divider produced %d children, want 1", len(node.Children))
	}
	return node.Children[0].Text
}

func TestDividerDefaultLine(t *testing.T) {
	got := dividerText(t, components.DividerProps{})
	want := strings.Repeat(components.DefaultDividerChar, components.DefaultDividerWidth)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDividerCustomChar(t *testing.T) {
	got := dividerText(t, components.DividerProps{Char: "=", Width: 10})
	if got != strings.Repeat("=", 10) {
		t.Fatalf("got %q, want %q", got, strings.Repeat("=", 10))
	}
}

func TestDividerWithTitle(t *testing.T) {
	got := dividerText(t, components.DividerProps{Title: "X", Width: 10, Char: "-"})
	if !strings.Contains(got, "X") {
		t.Fatalf("Divider with title missing title: %q", got)
	}
	if len(got) < 10 {
		t.Fatalf("got %q, expected width >=10", got)
	}
}

func TestDividerTitleWithPadding(t *testing.T) {
	got := dividerText(t, components.DividerProps{Title: "T", Padding: 2, Width: 20, Char: "-"})
	if !strings.Contains(got, "  T  ") {
		t.Fatalf("padding not applied: %q", got)
	}
}

func TestDividerTitleWiderThanWidth(t *testing.T) {
	got := dividerText(t, components.DividerProps{Title: "long-title", Width: 4, Char: "-"})
	if !strings.Contains(got, "long-title") {
		t.Fatalf("expected title to survive when wider than Width: %q", got)
	}
}

func TestDividerWithColorWraps(t *testing.T) {
	got := dividerText(t, components.DividerProps{Width: 5, Char: "-", Color: styles.Red})
	// Colorized output contains ANSI escape code for red (31)
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("expected ANSI escape in colored divider, got %q", got)
	}
}

func TestDividerNegativePaddingNormalized(t *testing.T) {
	got := dividerText(t, components.DividerProps{Title: "X", Padding: -5, Width: 10, Char: "-"})
	if !strings.Contains(got, "X") {
		t.Fatalf("title lost with negative padding: %q", got)
	}
}
