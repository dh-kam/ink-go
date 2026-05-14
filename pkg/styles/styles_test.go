package styles_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/styles"
)

// TestColorBasic tests basic color creation
func TestColorBasic(t *testing.T) {
	tests := []struct {
		name     string
		color    styles.Color
		expected string
	}{
		{"Red", styles.Red, "red"},
		{"Green", styles.Green, "green"},
		{"Blue", styles.Blue, "blue"},
		{"Yellow", styles.Yellow, "yellow"},
		{"Magenta", styles.Magenta, "magenta"},
		{"Cyan", styles.Cyan, "cyan"},
		{"White", styles.White, "white"},
		{"Black", styles.Black, "black"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color.Name() != tt.expected {
				t.Errorf("Expected color name %q, got %q", tt.expected, tt.color.Name())
			}
		})
	}
}

// TestColorToANSI tests ANSI code generation
func TestColorToANSI(t *testing.T) {
	tests := []struct {
		name     string
		color    styles.Color
		mode     styles.ColorMode
		contains string
	}{
		{"Red foreground", styles.Red, styles.Foreground, "31"},
		{"Green foreground", styles.Green, styles.Foreground, "32"},
		{"Blue background", styles.Blue, styles.Background, "44"},
		{"Yellow background", styles.Yellow, styles.Background, "43"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ansi := tt.color.ToANSI(tt.mode)
			if !strings.Contains(ansi, tt.contains) {
				t.Errorf("Expected ANSI code to contain %q, got %q", tt.contains, ansi)
			}
		})
	}
}

// TestStyleText tests applying styles to text
func TestStyleText(t *testing.T) {
	text := "Hello"

	// Test bold
	bold := styles.Bold(text)
	if !strings.Contains(bold, "1m") { // Bold ANSI code
		t.Errorf("Expected bold ANSI code in %q", bold)
	}
	if !strings.Contains(bold, text) {
		t.Errorf("Expected text %q in styled output", text)
	}

	// Test with color
	colored := styles.Colorize(text, styles.Red, styles.Foreground)
	if !strings.Contains(colored, "31m") { // Red foreground
		t.Errorf("Expected red color code in %q", colored)
	}
}

// TestStyleCombination tests combining multiple styles
func TestStyleCombination(t *testing.T) {
	text := "Styled"

	// Bold + Red
	styled := styles.Bold(styles.Colorize(text, styles.Red, styles.Foreground))

	if !strings.Contains(styled, "1m") {
		t.Error("Expected bold code")
	}
	if !strings.Contains(styled, "31m") {
		t.Error("Expected red color code")
	}
	if !strings.Contains(styled, text) {
		t.Error("Expected original text")
	}
}

// TestStyleReset tests that styles are properly reset
func TestStyleReset(t *testing.T) {
	text := "Test"
	styled := styles.Bold(text)

	// Should end with reset code
	if !strings.HasSuffix(styled, "\x1b[0m") {
		t.Errorf("Expected styled text to end with reset code, got: %q", styled)
	}
}

// TestRGBColor tests RGB color support
func TestRGBColor(t *testing.T) {
	rgb := styles.RGB(255, 0, 0) // Pure red

	ansi := rgb.ToANSI(styles.Foreground)
	// RGB colors use 38;2;r;g;b format
	if !strings.Contains(ansi, "38;2") {
		t.Errorf("Expected RGB ANSI format, got %q", ansi)
	}
	if !strings.Contains(ansi, "255;0;0") {
		t.Errorf("Expected RGB values 255;0;0, got %q", ansi)
	}
}

// TestDim tests dim style
func TestDim(t *testing.T) {
	text := "dimmed"
	dimmed := styles.Dim(text)

	if !strings.Contains(dimmed, "2m") { // Dim ANSI code
		t.Errorf("Expected dim ANSI code in %q", dimmed)
	}
	if !strings.Contains(dimmed, text) {
		t.Errorf("Expected text %q in styled output", text)
	}
	if !strings.HasSuffix(dimmed, "\x1b[0m") {
		t.Errorf("Expected reset code at end, got %q", dimmed)
	}
}

// TestItalic tests italic style
func TestItalic(t *testing.T) {
	text := "italicized"
	italic := styles.Italic(text)

	if !strings.Contains(italic, "3m") { // Italic ANSI code
		t.Errorf("Expected italic ANSI code in %q", italic)
	}
	if !strings.Contains(italic, text) {
		t.Errorf("Expected text %q in styled output", text)
	}
}

// TestUnderline tests underline style
func TestUnderline(t *testing.T) {
	text := "underlined"
	underlined := styles.Underline(text)

	if !strings.Contains(underlined, "4m") { // Underline ANSI code
		t.Errorf("Expected underline ANSI code in %q", underlined)
	}
	if !strings.Contains(underlined, text) {
		t.Errorf("Expected text %q in styled output", text)
	}
}

// TestStrikethrough tests strikethrough style
func TestStrikethrough(t *testing.T) {
	text := "struck"
	struck := styles.Strikethrough(text)

	if !strings.Contains(struck, "9m") { // Strikethrough ANSI code
		t.Errorf("Expected strikethrough ANSI code in %q", struck)
	}
	if !strings.Contains(struck, text) {
		t.Errorf("Expected text %q in styled output", text)
	}
}

func TestInverse(t *testing.T) {
	text := "inverse"
	inversed := styles.Inverse(text)

	if !strings.Contains(inversed, "7m") {
		t.Errorf("Expected inverse ANSI code in %q", inversed)
	}

	if !strings.Contains(inversed, text) {
		t.Errorf("Expected text %q in styled output", text)
	}
}

// TestReset tests Reset function
func TestReset(t *testing.T) {
	reset := styles.Reset()

	if reset != "\x1b[0m" {
		t.Errorf("Expected reset code \\x1b[0m, got %q", reset)
	}
}

// TestRGBBackground tests RGB background mode
func TestRGBBackground(t *testing.T) {
	rgb := styles.RGB(100, 150, 200)

	ansi := rgb.ToANSI(styles.Background)
	// RGB background uses 48;2;r;g;b format
	if !strings.Contains(ansi, "48;2") {
		t.Errorf("Expected RGB background ANSI format, got %q", ansi)
	}
	if !strings.Contains(ansi, "100;150;200") {
		t.Errorf("Expected RGB values 100;150;200, got %q", ansi)
	}
}

// TestRGBName tests RGB color name
func TestRGBName(t *testing.T) {
	rgb := styles.RGB(128, 64, 192)

	name := rgb.Name()
	expected := "rgb(128,64,192)"
	if name != expected {
		t.Errorf("Expected name %q, got %q", expected, name)
	}
}

// TestColorizeBackground tests colorize with background mode
func TestColorizeBackground(t *testing.T) {
	text := "highlighted"
	colored := styles.Colorize(text, styles.Blue, styles.Background)

	if !strings.Contains(colored, "44m") { // Blue background
		t.Errorf("Expected blue background code in %q", colored)
	}
	if !strings.Contains(colored, text) {
		t.Errorf("Expected text %q in colored output", text)
	}
	if !strings.HasSuffix(colored, "\x1b[0m") {
		t.Errorf("Expected reset code at end, got %q", colored)
	}
}

func TestWrapWithANSI(t *testing.T) {
	text := styles.WrapWithANSI("Test", styles.BoldCode(), styles.UnderlineCode())

	if !strings.Contains(text, styles.BoldCode()) {
		t.Error("Expected bold code")
	}

	if !strings.Contains(text, styles.UnderlineCode()) {
		t.Error("Expected underline code")
	}

	if !strings.HasSuffix(text, styles.Reset()) {
		t.Error("Expected reset code")
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"green", "green"},
		{"#FF8800", "rgb(255,136,0)"},
		{"rgb(255, 136, 0)", "rgb(255,136,0)"},
		{"ansi256(194)", "ansi256(194)"},
	}

	for _, tt := range tests {
		color, ok := styles.ParseColor(tt.input)
		if !ok {
			t.Fatalf("expected color %q to parse", tt.input)
		}

		if color.Name() != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, color.Name())
		}
	}
}

func TestParseColorInvalid(t *testing.T) {
	if _, ok := styles.ParseColor("not-a-color"); ok {
		t.Fatal("expected invalid color to fail parsing")
	}
}

// TestEmptyTextStyles tests styling empty text
func TestEmptyTextStyles(t *testing.T) {
	empty := ""

	// All should handle empty text gracefully
	bold := styles.Bold(empty)
	dim := styles.Dim(empty)
	italic := styles.Italic(empty)
	underline := styles.Underline(empty)
	strike := styles.Strikethrough(empty)
	colored := styles.Colorize(empty, styles.Red, styles.Foreground)

	// All should contain the reset code
	resetCode := "\x1b[0m"
	if !strings.HasSuffix(bold, resetCode) {
		t.Error("Bold empty text should end with reset")
	}
	if !strings.HasSuffix(dim, resetCode) {
		t.Error("Dim empty text should end with reset")
	}
	if !strings.HasSuffix(italic, resetCode) {
		t.Error("Italic empty text should end with reset")
	}
	if !strings.HasSuffix(underline, resetCode) {
		t.Error("Underline empty text should end with reset")
	}
	if !strings.HasSuffix(strike, resetCode) {
		t.Error("Strikethrough empty text should end with reset")
	}
	if !strings.HasSuffix(colored, resetCode) {
		t.Error("Colored empty text should end with reset")
	}
}

// TestColorModeValues tests ColorMode constants
func TestColorModeValues(t *testing.T) {
	if styles.Foreground != 0 {
		t.Errorf("Expected Foreground to be 0, got %d", styles.Foreground)
	}
	if styles.Background != 1 {
		t.Errorf("Expected Background to be 1, got %d", styles.Background)
	}
}

// TestAllBasicColorANSICodes tests all basic colors generate correct ANSI codes
func TestAllBasicColorANSICodes(t *testing.T) {
	tests := []struct {
		color  styles.Color
		fgCode string
		bgCode string
	}{
		{styles.Black, "30", "40"},
		{styles.Red, "31", "41"},
		{styles.Green, "32", "42"},
		{styles.Yellow, "33", "43"},
		{styles.Blue, "34", "44"},
		{styles.Magenta, "35", "45"},
		{styles.Cyan, "36", "46"},
		{styles.White, "37", "47"},
	}

	for _, tt := range tests {
		t.Run(tt.color.Name(), func(t *testing.T) {
			fgANSI := tt.color.ToANSI(styles.Foreground)
			if !strings.Contains(fgANSI, tt.fgCode+"m") {
				t.Errorf("Expected foreground code %q in %q", tt.fgCode, fgANSI)
			}

			bgANSI := tt.color.ToANSI(styles.Background)
			if !strings.Contains(bgANSI, tt.bgCode+"m") {
				t.Errorf("Expected background code %q in %q", tt.bgCode, bgANSI)
			}
		})
	}
}
