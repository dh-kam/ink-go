package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

func TestAlertVariants(t *testing.T) {
	cases := []struct {
		name string
		v    components.AlertVariant
		icon string
	}{
		{"info", components.AlertInfo, "i"},
		{"success", components.AlertSuccess, "✓"},
		{"warning", components.AlertWarning, "⚠"},
		{"error", components.AlertError, "✗"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := components.AlertText(components.AlertProps{
				Variant: tc.v,
				Title:   "Test",
				Message: "msg",
			})
			if !strings.Contains(got, tc.icon) {
				t.Fatalf("AlertText for %s missing icon %q: %q", tc.name, tc.icon, got)
			}
			if !strings.Contains(got, "Test") {
				t.Fatalf("AlertText missing title: %q", got)
			}
			if !strings.Contains(got, "msg") {
				t.Fatalf("AlertText missing message: %q", got)
			}
		})
	}
}

func TestAlertTextNoTitle(t *testing.T) {
	got := components.AlertText(components.AlertProps{
		Variant: components.AlertInfo,
		Message: "just message",
	})
	if !strings.Contains(got, "just message") {
		t.Fatalf("AlertText missing message: %q", got)
	}
}

func TestAlertTextNoMessage(t *testing.T) {
	got := components.AlertText(components.AlertProps{
		Variant: components.AlertSuccess,
		Title:   "Done",
	})
	if !strings.Contains(got, "Done") {
		t.Fatalf("AlertText missing title: %q", got)
	}
}

func TestAlertReturnsBorderedNode(t *testing.T) {
	node := components.Alert(components.AlertProps{
		Variant: components.AlertInfo,
		Title:   "Hello",
		Message: "World",
	})
	if node == nil {
		t.Fatal("Alert returned nil node")
	}
	// Border wraps a single body element.
	if len(node.Children) == 0 {
		t.Fatalf("Alert had no children")
	}
}

func TestAlertEmptyMessage(t *testing.T) {
	node := components.Alert(components.AlertProps{
		Variant: components.AlertWarning,
		Title:   "Heads up",
	})
	if node == nil {
		t.Fatal("Alert with no message returned nil")
	}
}

func TestAlertUnknownVariantFallsBackToInfo(t *testing.T) {
	got := components.AlertText(components.AlertProps{
		Variant: components.AlertVariant("bogus"),
		Title:   "X",
	})
	if !strings.Contains(got, "i") { // info icon
		t.Fatalf("unknown variant should fall back to info: %q", got)
	}
}
