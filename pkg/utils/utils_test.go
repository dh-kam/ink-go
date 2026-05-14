package utils_test

import (
	"testing"

	"github.com/dh-kam/goink.go/pkg/utils"
)

// TestTruncate tests string truncation
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		expected string
	}{
		{"no truncation", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate", "hello world", 8, "hello..."},
		{"no truncation short", "hi", 3, "hi"},
		{"truncate very short", "test", 3, "..."},
		{"empty string", "", 5, ""},
		{"zero max", "hello", 0, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Truncate(tt.s, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestPadLeft tests left padding
func TestPadLeft(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		width    int
		expected string
	}{
		{"no padding", "hello", 5, "hello"},
		{"pad", "hi", 5, "   hi"},
		{"empty", "", 3, "   "},
		{"longer than width", "hello", 3, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.PadLeft(tt.s, tt.width)
			if result != tt.expected {
				t.Errorf("PadLeft(%q, %d) = %q, want %q", tt.s, tt.width, result, tt.expected)
			}
		})
	}
}

// TestPadRight tests right padding
func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		width    int
		expected string
	}{
		{"no padding", "hello", 5, "hello"},
		{"pad", "hi", 5, "hi   "},
		{"empty", "", 3, "   "},
		{"longer than width", "hello", 3, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.PadRight(tt.s, tt.width)
			if result != tt.expected {
				t.Errorf("PadRight(%q, %d) = %q, want %q", tt.s, tt.width, result, tt.expected)
			}
		})
	}
}

// TestWordWrap tests word wrapping
func TestWordWrap(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string
	}{
		{"single word", "hello", 10, []string{"hello"}},
		{"multiple words fit", "hello world", 20, []string{"hello world"}},
		{"wrap needed", "hello world foo", 10, []string{"hello", "world foo"}},
		{"long word", "supercalifragilistic", 10, []string{"supercalifragilistic"}},
		{"empty", "", 10, []string{""}},
		{"zero width", "test", 0, []string{"test"}},
		{"multiple lines", "one two three four", 8, []string{"one two", "three", "four"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.WordWrap(tt.text, tt.width)
			if len(result) != len(tt.expected) {
				t.Errorf("WordWrap(%q, %d) returned %d lines, want %d", tt.text, tt.width, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("WordWrap(%q, %d)[%d] = %q, want %q", tt.text, tt.width, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestSplitLines tests line splitting
func TestSplitLines(t *testing.T) {
	input := "line1\nline2\nline3"
	expected := []string{"line1", "line2", "line3"}
	result := utils.SplitLines(input)

	if len(result) != len(expected) {
		t.Fatalf("Expected %d lines, got %d", len(expected), len(result))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Line %d: expected %q, got %q", i, expected[i], result[i])
		}
	}
}

// TestJoinLines tests line joining
func TestJoinLines(t *testing.T) {
	input := []string{"line1", "line2", "line3"}
	expected := "line1\nline2\nline3"
	result := utils.JoinLines(input)

	if result != expected {
		t.Errorf("JoinLines(%v) = %q, want %q", input, result, expected)
	}
}

// TestTrimSpace tests trimming whitespace from each line
func TestTrimSpace(t *testing.T) {
	input := "  line1  \n  line2  \nline3"
	expected := "line1\nline2\nline3"
	result := utils.TrimSpace(input)

	if result != expected {
		t.Errorf("TrimSpace(%q) = %q, want %q", input, result, expected)
	}
}

// TestMin tests Min function
func TestMin(t *testing.T) {
	if utils.Min(5, 10) != 5 {
		t.Error("Min(5, 10) should be 5")
	}
	if utils.Min(10, 5) != 5 {
		t.Error("Min(10, 5) should be 5")
	}
	if utils.Min(5, 5) != 5 {
		t.Error("Min(5, 5) should be 5")
	}
	if utils.Min(-5, 10) != -5 {
		t.Error("Min(-5, 10) should be -5")
	}
}

// TestMax tests Max function
func TestMax(t *testing.T) {
	if utils.Max(5, 10) != 10 {
		t.Error("Max(5, 10) should be 10")
	}
	if utils.Max(10, 5) != 10 {
		t.Error("Max(10, 5) should be 10")
	}
	if utils.Max(5, 5) != 5 {
		t.Error("Max(5, 5) should be 5")
	}
	if utils.Max(-5, 10) != 10 {
		t.Error("Max(-5, 10) should be 10")
	}
}

// TestClamp tests Clamp function
func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected int
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 0, 0},
		{5, 10, 20, 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := utils.Clamp(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

// TestAbs tests Abs function
func TestAbs(t *testing.T) {
	if utils.Abs(5) != 5 {
		t.Error("Abs(5) should be 5")
	}
	if utils.Abs(-5) != 5 {
		t.Error("Abs(-5) should be 5")
	}
	if utils.Abs(0) != 0 {
		t.Error("Abs(0) should be 0")
	}
}

// TestIsInBounds tests bounds checking
func TestIsInBounds(t *testing.T) {
	tests := []struct {
		x, y, width, height int
		expected            bool
	}{
		{5, 5, 10, 10, true},
		{0, 0, 10, 10, true},
		{9, 9, 10, 10, true},
		{-1, 5, 10, 10, false},
		{5, -1, 10, 10, false},
		{10, 5, 10, 10, false},
		{5, 10, 10, 10, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := utils.IsInBounds(tt.x, tt.y, tt.width, tt.height)
			if result != tt.expected {
				t.Errorf("IsInBounds(%d, %d, %d, %d) = %v, want %v",
					tt.x, tt.y, tt.width, tt.height, result, tt.expected)
			}
		})
	}
}

// TestIsValidWidth tests width validation
func TestIsValidWidth(t *testing.T) {
	if !utils.IsValidWidth(10) {
		t.Error("IsValidWidth(10) should be true")
	}
	if utils.IsValidWidth(0) {
		t.Error("IsValidWidth(0) should be false")
	}
	if utils.IsValidWidth(-5) {
		t.Error("IsValidWidth(-5) should be false")
	}
}

// TestIsValidHeight tests height validation
func TestIsValidHeight(t *testing.T) {
	if !utils.IsValidHeight(10) {
		t.Error("IsValidHeight(10) should be true")
	}
	if utils.IsValidHeight(0) {
		t.Error("IsValidHeight(0) should be false")
	}
	if utils.IsValidHeight(-5) {
		t.Error("IsValidHeight(-5) should be false")
	}
}

// TestRuneWidth tests rune width calculation
func TestRuneWidth(t *testing.T) {
	if utils.RuneWidth('a') != 1 {
		t.Error("RuneWidth('a') should be 1")
	}
	if utils.RuneWidth('日') != 2 {
		t.Error("RuneWidth('日') should be 2 (wide character)")
	}
	if utils.RuneWidth('🍔') != 2 {
		t.Error("RuneWidth('🍔') should be 2 (emoji)")
	}
	if utils.RuneWidth('⏳') != 2 {
		t.Error("RuneWidth('⏳') should be 2 (emoji symbol)")
	}
	if utils.RuneWidth('⚠') != 2 {
		t.Error("RuneWidth('⚠') should be 2 (emoji presentation base)")
	}
	if utils.RuneWidth('\u200d') != 0 {
		t.Error("RuneWidth('\\u200d') should be 0 (ZWJ)")
	}
	if utils.RuneWidth('\ufe0f') != 0 {
		t.Error("RuneWidth('\\ufe0f') should be 0 (variation selector)")
	}
}

// TestStringWidth tests string width calculation
func TestStringWidth(t *testing.T) {
	if utils.StringWidth("hello") != 5 {
		t.Error("StringWidth('hello') should be 5")
	}
	if utils.StringWidth("hello日") != 7 {
		t.Error("StringWidth('hello日') should be 7")
	}
	if utils.StringWidth("🍔|") != 3 {
		t.Error("StringWidth('🍔|') should be 3")
	}
	if utils.StringWidth("🌡️⚠️✅") != 6 {
		t.Error("StringWidth('🌡️⚠️✅') should be 6")
	}
}

func TestStringWidthMatchesStringWidthGraphemeClusters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"family emoji", "👨‍👩‍👧‍👦", 2},
		{"skin tone emoji", "👍🏽", 2},
		{"rainbow flag", "🏳️‍🌈", 2},
		{"text variation selector", "♥︎", 1},
		{"emoji variation selector", "♥️", 2},
		{"plain emoji-capable symbols stay narrow", "🌡⚠✈", 3},
		{"emoji variation symbols are wide", "🌡️⚠️✈️", 6},
		{"emoji variation legacy symbols are wide", "™️©️®️ℹ️", 8},
		{"regional indicator flag", "🇺🇸", 2},
		{"lone regional indicator", "🇨", 1},
		{"unsupported regional indicator pair", "🇦🇧", 1},
		{"recognized flag plus lone regional indicator", "🇺🇸🇨", 3},
		{"marked regional indicator pair is not a flag", "🇺🇸️", 1},
		{"keycap sequence", "1️⃣", 2},
		{"dangling zwj variation sequence", "⚠️‍", 1},
		{"box drawing glyphs stay narrow", "╘════════════╛ +", 16},
		{"block and arrow glyphs stay narrow", "▀→←↘↙", 5},
		{"combining mark cluster", "é", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := utils.StringWidth(tt.input); result != tt.expected {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStringWidthMatchesWidestLineGraphemeClusters(t *testing.T) {
	text := "a\n👨‍👩‍👧‍👦👍🏽\n🇺🇸🇨"
	widest := 0
	for _, line := range utils.SplitLines(text) {
		widest = utils.Max(widest, utils.StringWidth(line))
	}

	if widest != 4 {
		t.Fatalf("widest grapheme-aware line width = %d, want 4", widest)
	}
}

func TestStringWidthIgnoresANSIAndControlCharacters(t *testing.T) {
	input := "\x1b[31mred\x1b[0m\n\t\x00"
	if result := utils.StringWidth(input); result != 3 {
		t.Errorf("StringWidth(%q) = %d, want 3", input, result)
	}
}

// TestTruncateWidth tests width-based truncation
func TestTruncateWidth(t *testing.T) {
	tests := []struct {
		s        string
		maxWidth int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello wo"},
		{"hello日", 7, "hello日"},
		{"hello日", 6, "hello"},
		{"hello日", 5, "hello"},
		{"", 5, ""},
		{"test", 0, ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := utils.TruncateWidth(tt.s, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("TruncateWidth(%q, %d) = %q, want %q", tt.s, tt.maxWidth, result, tt.expected)
			}
		})
	}
}

func TestTruncateWidthPreservesGraphemeClusters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"family emoji fits as one cluster", "a👨‍👩‍👧‍👦b", 3, "a👨‍👩‍👧‍👦"},
		{"family emoji does not split", "a👨‍👩‍👧‍👦b", 2, "a"},
		{"variation selector stays attached", "✈️x", 2, "✈️"},
		{"variation selector cluster too wide", "✈️x", 1, ""},
		{"skin tone modifier stays attached", "👍🏽ok", 2, "👍🏽"},
		{"flag stays attached", "🇺🇸x", 2, "🇺🇸"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := utils.TruncateWidth(tt.input, tt.width); result != tt.expected {
				t.Errorf("TruncateWidth(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
			}
		})
	}
}

// TestContains tests string contains with case sensitivity
func TestContains(t *testing.T) {
	if !utils.Contains("hello world", "lo", false) {
		t.Error("Contains('hello world', 'lo', false) should be true")
	}
	if !utils.Contains("hello world", "LO", true) {
		t.Error("Contains('hello world', 'LO', true) should be true")
	}
	if utils.Contains("hello world", "LO", false) {
		t.Error("Contains('hello world', 'LO', false) should be false")
	}
}

// TestIsPrintable tests printable character check
func TestIsPrintable(t *testing.T) {
	if !utils.IsPrintable('a') {
		t.Error("IsPrintable('a') should be true")
	}
	if utils.IsPrintable('\n') {
		t.Error("IsPrintable('\\n') should be false")
	}
	if utils.IsPrintable('\x00') {
		t.Error("IsPrintable('\\x00') should be false")
	}
}

// TestFilterPrintable tests filtering printable characters
func TestFilterPrintable(t *testing.T) {
	input := "hello\x00world\n"
	expected := "helloworld"
	result := utils.FilterPrintable(input)

	if result != expected {
		t.Errorf("FilterPrintable(%q) = %q, want %q", input, result, expected)
	}
}
