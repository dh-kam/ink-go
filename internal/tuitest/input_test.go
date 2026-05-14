package tuitest

import "testing"

func TestNormalizeInputShiftTab(t *testing.T) {
	got, err := NormalizeInput(InputSpec{Key: "shift-tab"})
	if err != nil {
		t.Fatalf("NormalizeInput shift-tab failed: %v", err)
	}
	if got != "\x1b[Z" {
		t.Fatalf("NormalizeInput shift-tab = %q, want %q", got, "\x1b[Z")
	}
}

func TestNormalizeInputHex(t *testing.T) {
	got, err := NormalizeInput(InputSpec{Hex: "ea b0 80"})
	if err != nil {
		t.Fatalf("NormalizeInput hex failed: %v", err)
	}
	if got != "가" {
		t.Fatalf("NormalizeInput hex = %q, want %q", got, "가")
	}
}

func TestNormalizeInputRejectsMixedHex(t *testing.T) {
	if _, err := NormalizeInput(InputSpec{Text: "x", Hex: "78"}); err == nil {
		t.Fatalf("expected mixed hex/text input to fail")
	}
}
