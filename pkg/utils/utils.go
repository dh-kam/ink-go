package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
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

// GraphemeClusters splits a string into user-perceived characters.
func GraphemeClusters(s string) []string {
	if s == "" {
		return nil
	}

	clusters := make([]string, 0, len(s))
	graphemes := uniseg.NewGraphemes(s)
	for graphemes.Next() {
		clusters = append(clusters, graphemes.Str())
	}

	return clusters
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

// StringWidth returns the display width of a string.
func StringWidth(s string) int {
	s = stripANSIForWidth(s)

	width := 0
	graphemes := uniseg.NewGraphemes(s)
	for graphemes.Next() {
		width += stringWidthCluster(graphemes.Str(), graphemes.Width())
	}
	return width
}

// TruncateWidth truncates a string to a maximum display width
func TruncateWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	width := 0
	var builder strings.Builder
	graphemes := uniseg.NewGraphemes(s)
	for graphemes.Next() {
		cluster := graphemes.Str()
		clusterWidth := stringWidthCluster(cluster, graphemes.Width())
		if width+clusterWidth > maxWidth {
			// Check if we have room for ellipsis (3 display width)
			ellipsisWidth := 3
			if width+ellipsisWidth <= maxWidth && builder.Len() > 0 {
				return builder.String() + "..."
			}
			// No room for ellipsis, just return what fits
			if builder.Len() > 0 {
				return builder.String()
			}
			return ""
		}
		builder.WriteString(cluster)
		width += clusterWidth
	}
	return s
}

func stringWidthCluster(cluster string, defaultWidth int) int {
	runes := []rune(cluster)
	if len(runes) == 0 {
		return 0
	}

	if isWidthZeroCluster(runes) {
		return 0
	}

	if isWidthRegionalIndicatorCluster(runes) {
		if isWidthRecognizedRegionalIndicatorCluster(runes) {
			return 2
		}

		return 1
	}

	if len(runes) == 1 && isWidthOneTerminalSymbol(runes[0]) {
		return 1
	}

	switch {
	case isWidthKeycapCluster(runes):
		return 2
	case isWidthEmojiZWJSequenceCluster(runes):
		return 2
	case isWidthSimpleEmojiPresentationCluster(runes):
		return 2
	case strings.ContainsRune(cluster, 0x200D):
		return widthBaseCluster(runes)
	default:
		return defaultWidth
	}
}

func isWidthOneTerminalSymbol(r rune) bool {
	return (r >= 0x2500 && r <= 0x259F) || // box drawing + block elements
		(r >= 0x2190 && r <= 0x21FF) // arrows used by cli-boxes border styles
}

func stripANSIForWidth(text string) string {
	if text == "" {
		return ""
	}

	const (
		ansiPlain = iota
		ansiEscape
		ansiCSI
		ansiOSC
		ansiOSCEscape
	)

	state := ansiPlain
	var builder strings.Builder
	builder.Grow(len(text))

	for index := 0; index < len(text); index++ {
		ch := text[index]

		switch state {
		case ansiPlain:
			if ch >= 0x80 {
				r, size := utf8.DecodeRuneInString(text[index:])
				if r != utf8.RuneError || size > 1 {
					builder.WriteString(text[index : index+size])
					index += size - 1
					continue
				}
			}

			switch ch {
			case 0x1b:
				state = ansiEscape
			case 0x9b:
				state = ansiCSI
			default:
				builder.WriteByte(ch)
			}
		case ansiEscape:
			switch ch {
			case '[':
				state = ansiCSI
			case ']':
				state = ansiOSC
			default:
				state = ansiPlain
			}
		case ansiCSI:
			if ch >= 0x40 && ch <= 0x7e {
				state = ansiPlain
			}
		case ansiOSC:
			switch ch {
			case 0x07:
				state = ansiPlain
			case 0x1b:
				state = ansiOSCEscape
			}
		case ansiOSCEscape:
			if ch == '\\' {
				state = ansiPlain
				continue
			}

			if ch == 0x1b {
				continue
			}

			state = ansiOSC
		}
	}

	return builder.String()
}

func isWidthZeroControl(r rune) bool {
	return (r >= 0 && r < 0x20) || (r >= 0x7f && r < 0xa0)
}

func isWidthZeroCluster(cluster []rune) bool {
	for _, r := range cluster {
		if !isWidthZeroControl(r) && !isWidthZeroMark(r) {
			return false
		}
	}

	return true
}

func isWidthZeroMark(r rune) bool {
	return unicode.Is(unicode.Cf, r) || isWidthVariationSelector(r) || unicode.Is(unicode.M, r)
}

func isWidthVariationSelector(r rune) bool {
	return (r >= 0xFE00 && r <= 0xFE0F) || (r >= 0xE0100 && r <= 0xE01EF)
}

func isWidthEmojiModifier(r rune) bool {
	return r >= 0x1F3FB && r <= 0x1F3FF
}

func isWidthEmojiModifierBase(r rune) bool {
	if isWidthEmojiBase(r) {
		return true
	}

	switch {
	case r == 0x261D || r == 0x26F9:
		return true
	case r >= 0x270A && r <= 0x270D:
		return true
	default:
		return false
	}
}

func isWidthRegionalIndicator(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
}

func isWidthRecognizedRegionalIndicatorCluster(cluster []rune) bool {
	return len(cluster) == 2 &&
		isWidthRegionalIndicator(cluster[0]) &&
		isWidthRegionalIndicator(cluster[1]) &&
		isWidthRecognizedRegionalIndicatorPair(cluster[0], cluster[1])
}

func isWidthRecognizedRegionalIndicatorPair(left rune, right rune) bool {
	code := string([]rune{
		rune('A' + (left - 0x1F1E6)),
		rune('A' + (right - 0x1F1E6)),
	})

	return strings.Contains(widthSupportedFlagRegionCodes, " "+code+" ")
}

func isWidthRegionalIndicatorCluster(cluster []rune) bool {
	baseIndex := firstWidthVisibleRuneIndex(cluster)
	return baseIndex >= 0 && isWidthRegionalIndicator(cluster[baseIndex])
}

func isWidthKeycapBase(r rune) bool {
	return r == '#' || r == '*' || (r >= '0' && r <= '9')
}

func isWidthEmojiBase(r rune) bool {
	if isWidthRegionalIndicator(r) {
		return true
	}

	return (r >= 0x1F300 && r <= 0x1FAFF) ||
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
		(r >= 0x2B1B && r <= 0x2B1C) || r == 0x2B50 || r == 0x2B55
}

func isWidthEmojiVariationBase(r rune) bool {
	if isWidthEmojiBase(r) {
		return true
	}

	return r == 0x00A9 || r == 0x00AE ||
		r == 0x203C || r == 0x2049 || r == 0x2122 || r == 0x2139 ||
		(r >= 0x2194 && r <= 0x2199) || (r >= 0x21A9 && r <= 0x21AA) ||
		r == 0x2328 || r == 0x23CF || (r >= 0x23ED && r <= 0x23EF) || (r >= 0x23F1 && r <= 0x23F2) ||
		r == 0x24C2 ||
		(r >= 0x25AA && r <= 0x25AB) || r == 0x25B6 || r == 0x25C0 || (r >= 0x25FB && r <= 0x25FE) ||
		(r >= 0x2600 && r <= 0x2604) || r == 0x260E || r == 0x2611 || r == 0x2618 || r == 0x261D ||
		r == 0x2620 || (r >= 0x2622 && r <= 0x2623) || r == 0x2626 || r == 0x262A ||
		(r >= 0x262E && r <= 0x262F) || (r >= 0x2638 && r <= 0x263A) || r == 0x2640 || r == 0x2642 ||
		r == 0x265F || r == 0x2660 || r == 0x2663 || (r >= 0x2665 && r <= 0x2666) || r == 0x2668 ||
		r == 0x267B || (r >= 0x267E && r <= 0x267F) || (r >= 0x2692 && r <= 0x2697) || r == 0x2699 ||
		(r >= 0x269B && r <= 0x269C) || r == 0x26A7 || (r >= 0x26B0 && r <= 0x26B1) ||
		r == 0x26C8 || r == 0x26CF || r == 0x26D1 || r == 0x26D3 ||
		(r >= 0x26E9 && r <= 0x26EA) || (r >= 0x26F0 && r <= 0x26F1) || (r >= 0x26F7 && r <= 0x26FA) ||
		r == 0x2702 || (r >= 0x2708 && r <= 0x270D) || r == 0x270F || r == 0x2712 || r == 0x2714 ||
		r == 0x2716 || r == 0x271D || r == 0x2721 || (r >= 0x2733 && r <= 0x2734) ||
		r == 0x2744 || r == 0x2747 || r == 0x2763 || r == 0x2764 || r == 0x27A1 ||
		(r >= 0x2934 && r <= 0x2935) || (r >= 0x2B05 && r <= 0x2B07) ||
		r == 0x3030 || r == 0x303D || r == 0x3297 || r == 0x3299
}

func widthBaseCluster(cluster []rune) int {
	baseIndex := firstWidthVisibleRuneIndex(cluster)
	if baseIndex < 0 {
		return 0
	}

	width := uniseg.StringWidth(string(cluster[baseIndex]))
	for _, r := range cluster[baseIndex+1:] {
		if r >= 0xFF00 && r <= 0xFFEF {
			width += uniseg.StringWidth(string(r))
		}
	}

	return width
}

func firstWidthVisibleRuneIndex(cluster []rune) int {
	for index, r := range cluster {
		if !isWidthZeroControl(r) && !isWidthZeroMark(r) {
			return index
		}
	}

	return -1
}

func isWidthKeycapCluster(cluster []rune) bool {
	if len(cluster) < 2 || len(cluster) > 3 {
		return false
	}

	if !isWidthKeycapBase(cluster[0]) {
		return false
	}

	if len(cluster) == 2 {
		return cluster[1] == 0x20E3
	}

	return cluster[1] == 0xFE0F && cluster[2] == 0x20E3
}

func isWidthEmojiZWJSequenceCluster(cluster []rune) bool {
	pictographics := 0
	containsZWJ := false

	for _, r := range cluster {
		switch {
		case r == 0x200D:
			containsZWJ = true
		case isWidthEmojiZWJJoinableBase(r):
			pictographics++
		}
	}

	return containsZWJ && pictographics >= 2
}

func isWidthEmojiZWJJoinableBase(r rune) bool {
	return isWidthEmojiVariationBase(r) && !isWidthRegionalIndicator(r) && !isWidthEmojiModifier(r)
}

func isWidthSimpleEmojiPresentationCluster(cluster []rune) bool {
	baseIndex := firstWidthVisibleRuneIndex(cluster)
	if baseIndex < 0 || baseIndex >= len(cluster) {
		return false
	}

	base := cluster[baseIndex]
	if !isWidthEmojiVariationBase(base) {
		return false
	}

	if strings.ContainsRune(string(cluster), 0x200D) {
		return false
	}

	sawVariationSelector := false
	sawEmojiModifier := false
	for _, r := range cluster[baseIndex+1:] {
		switch {
		case r == 0xFE0F:
			if sawVariationSelector {
				return false
			}

			sawVariationSelector = true
		case isWidthEmojiModifier(r) && isWidthEmojiModifierBase(base):
			sawEmojiModifier = true
		case isWidthZeroMark(r):
			return false
		default:
			return false
		}
	}

	return sawVariationSelector || sawEmojiModifier
}

const widthSupportedFlagRegionCodes = " " +
	"AC AD AE AF AG AI AL AM AO AQ AR AS AT AU AW AX AZ BA BB BD BE BF BG BH BI BJ BL BM BN BO BQ BR BS BT BV BW BY BZ " +
	"CA CC CD CF CG CH CI CK CL CM CN CO CP CQ CR CU CV CW CX CY CZ DE DG DJ DK DM DO DZ EA EC EE EG EH ER ES ET EU " +
	"FI FJ FK FM FO FR GA GB GD GE GF GG GH GI GL GM GN GP GQ GR GS GT GU GW GY HK HM HN HR HT HU IC ID IE IL IM IN IO " +
	"IQ IR IS IT JE JM JO JP KE KG KH KI KM KN KP KR KW KY KZ LA LB LC LI LK LR LS LT LU LV LY MA MC MD ME MF MG MH MK " +
	"ML MM MN MO MP MQ MR MS MT MU MV MW MX MY MZ NA NC NE NF NG NI NL NO NP NR NU NZ OM PA PE PF PG PH PK PL PM PN PR " +
	"PS PT PW PY QA RE RO RS RU RW SA SB SC SD SE SG SH SI SJ SK SL SM SN SO SR SS ST SV SX SY SZ TA TC TD TF TG TH TJ " +
	"TK TL TM TN TO TR TT TV TW TZ UA UG UM UN US UY UZ VA VC VE VG VI VN VU WF WS XK YE YT ZA ZM ZW "

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
