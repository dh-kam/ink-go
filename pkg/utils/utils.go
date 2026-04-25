package utils

import (
	"strings"
	"unicode"
)

// String utilities

// Truncate truncates a string to a maximum length, adding an ellipsis if truncated
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

// PadLeft pads a string on the left with spaces to reach the target width
func PadLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// PadRight pads a string on the right with spaces to reach the target width
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// WordWrap wraps text to a specified width, preserving word boundaries
func WordWrap(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""
	currentLength := 0

	for _, word := range words {
		wordLen := len(word)

		if currentLength == 0 {
			// First word on line
			currentLine = word
			currentLength = wordLen
		} else if currentLength+1+wordLen <= width {
			// Word fits on current line
			currentLine += " " + word
			currentLength += 1 + wordLen
		} else {
			// Word doesn't fit, start new line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
			currentLength = wordLen
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		lines = append(lines, "")
	}

	return lines
}

// SplitLines splits a string into lines, handling both \n and \r\n
func SplitLines(s string) []string {
	return strings.Split(s, "\n")
}

// JoinLines joins lines with newline characters
func JoinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// TrimSpace trims leading and trailing whitespace from each line
func TrimSpace(s string) string {
	lines := SplitLines(s)
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return JoinLines(lines)
}

// Math utilities

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp clamps a value between a minimum and maximum
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Abs returns the absolute value of an integer
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Validation utilities

// IsInBounds checks if coordinates are within bounds
func IsInBounds(x, y, width, height int) bool {
	return x >= 0 && x < width && y >= 0 && y < height
}

// IsValidWidth checks if a width value is valid
func IsValidWidth(width int) bool {
	return width > 0
}

// IsValidHeight checks if a height value is valid
func IsValidHeight(height int) bool {
	return height > 0
}

// Rune utilities

// RuneWidth returns the display width of a rune (1 for most, 2 for wide characters)
func RuneWidth(r rune) int {
	if r == 0 {
		return 0
	}

	// Zero-width joiners, variation selectors, and combining marks don't advance the cursor.
	if r == 0x200D || (r >= 0xFE00 && r <= 0xFE0F) || unicode.Is(unicode.Mn, r) {
		return 0
	}

	// Simple check for wide unicode characters
	// This covers common East Asian wide ranges plus the main emoji/pictograph blocks.
	if r >= 0x1100 && (r <= 0x115F || r == 0x2329 || r == 0x232A ||
		(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE10 && r <= 0xFE19) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x20000 && r <= 0x2FFFD) ||
		(r >= 0x30000 && r <= 0x3FFFD)) {
		return 2
	}

	if (r >= 0x1F300 && r <= 0x1FAFF) ||
		r == 0x231A || r == 0x231B || r == 0x23F0 || r == 0x23F3 ||
		(r >= 0x23E9 && r <= 0x23EC) ||
		(r >= 0x25FD && r <= 0x25FE) ||
		(r >= 0x2614 && r <= 0x2615) ||
		r == 0x26A0 ||
		(r >= 0x2648 && r <= 0x2653) ||
		r == 0x267F || r == 0x2693 || r == 0x26A1 ||
		(r >= 0x26AA && r <= 0x26AB) ||
		(r >= 0x26BD && r <= 0x26BE) ||
		(r >= 0x26C4 && r <= 0x26C5) ||
		r == 0x26CE || r == 0x26D4 || r == 0x26EA ||
		(r >= 0x26F2 && r <= 0x26F3) ||
		r == 0x26F5 || r == 0x26FA || r == 0x26FD ||
		r == 0x2705 || (r >= 0x270A && r <= 0x270B) ||
		r == 0x2728 || r == 0x274C || r == 0x274E ||
		(r >= 0x2753 && r <= 0x2755) || r == 0x2757 ||
		(r >= 0x2795 && r <= 0x2797) || r == 0x27B0 || r == 0x27BF ||
		(r >= 0x2B1B && r <= 0x2B1C) || r == 0x2B50 || r == 0x2B55 {
		return 2
	}

	return 1
}

// StringWidth returns the display width of a string
func StringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += RuneWidth(r)
	}
	return width
}

// TruncateWidth truncates a string to a maximum display width
func TruncateWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	width := 0
	runes := []rune(s)
	for i, r := range runes {
		rw := RuneWidth(r)
		if width+rw > maxWidth {
			// Check if we have room for ellipsis (3 display width)
			ellipsisWidth := 3
			if width+ellipsisWidth <= maxWidth && i > 0 {
				return string(runes[:i]) + "..."
			}
			// No room for ellipsis, just return what fits
			if i > 0 {
				return string(runes[:i])
			}
			return ""
		}
		width += rw
	}
	return s
}

// Contains checks if a string contains a substring (case-insensitive option)
func Contains(s, substr string, caseInsensitive bool) bool {
	if caseInsensitive {
		return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
	}
	return strings.Contains(s, substr)
}

// IsPrintable checks if a rune is printable
func IsPrintable(r rune) bool {
	return unicode.IsPrint(r) && !unicode.IsControl(r)
}

// FilterPrintable removes non-printable characters from a string
func FilterPrintable(s string) string {
	var result []rune
	for _, r := range s {
		if IsPrintable(r) {
			result = append(result, r)
		}
	}
	return string(result)
}
