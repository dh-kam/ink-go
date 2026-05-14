package ink

import (
	"strings"
	"sync"
	"unicode"

	"github.com/dh-kam/ink-go/pkg/utils"
	"github.com/dh-kam/ink-go/pkg/vdom"
	"github.com/rivo/uniseg"
)

// DOMElement is the current Go-level handle for a rendered Ink element.
type DOMElement = vdom.Node

// ElementDimensions contains the measured size of an element.
type ElementDimensions struct {
	Width  int
	Height int
}

// ElementPosition contains the measured top/left coordinates of an element
// relative to its rendered surface.
type ElementPosition struct {
	Left int
	Top  int
}

// MeasureElementPosition returns the top/left coordinates that the layout
// engine has assigned to the element. Returns the zero value when the element
// has not yet been laid out (for example, before the first render or for refs
// attached to detached nodes).
func MeasureElementPosition(node *DOMElement) ElementPosition {
	if node == nil {
		return ElementPosition{}
	}

	layout := node.ComputedLayout()
	return ElementPosition{Left: layout.Left, Top: layout.Top}
}

var measureTextCache sync.Map

// MeasureElement returns the latest computed layout dimensions for an element ref.
func MeasureElement(node *DOMElement) ElementDimensions {
	if node == nil {
		return ElementDimensions{}
	}

	layout := node.ComputedLayout()
	if node.Type == vdom.TextNode && layout.Width == 0 && layout.Height == 0 {
		return MeasureText(node.NodeValue())
	}

	return ElementDimensions{
		Width:  layout.Width,
		Height: layout.Height,
	}
}

// MeasureText returns the display width and height of a text block.
func MeasureText(text string) ElementDimensions {
	if text == "" {
		return ElementDimensions{}
	}

	if cached, ok := measureTextCache.Load(text); ok {
		return cached.(ElementDimensions)
	}

	lines := strings.Split(text, "\n")
	maxWidth := 0
	for _, line := range lines {
		lineWidth := measureVisibleLineWidth(stripANSI(line))
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	dimensions := ElementDimensions{
		Width:  maxWidth,
		Height: len(lines),
	}

	measureTextCache.Store(text, dimensions)
	return dimensions
}

func measureVisibleLineWidth(text string) int {
	width := 0
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		width += measureVisibleClusterWidth([]rune(graphemes.Str()))
	}

	return width
}

func measureVisibleClusterWidth(cluster []rune) int {
	if len(cluster) == 0 {
		return 0
	}

	if isZeroWidthCluster(cluster) {
		return 0
	}

	if isRegionalIndicatorCluster(cluster) {
		if isRecognizedRegionalIndicatorCluster(cluster) {
			return 2
		}

		return 1
	}

	if isDoubleWidthEmojiCluster(cluster) {
		return 2
	}

	return measureBaseVisibleClusterWidth(cluster)
}

func isZeroWidthControl(r rune) bool {
	return (r >= 0 && r < 0x20) || (r >= 0x7f && r < 0xa0)
}

func isZeroWidthCluster(cluster []rune) bool {
	for _, r := range cluster {
		if !isZeroWidthControl(r) && !isZeroWidthMark(r) {
			return false
		}
	}

	return true
}

func isZeroWidthMark(r rune) bool {
	return unicode.Is(unicode.Cf, r) || isVariationSelector(r) || isCombiningMark(r)
}

func isVariationSelector(r rune) bool {
	return (r >= 0xFE00 && r <= 0xFE0F) || (r >= 0xE0100 && r <= 0xE01EF)
}

func isCombiningMark(r rune) bool {
	return unicode.Is(unicode.M, r)
}

func isEmojiModifier(r rune) bool {
	return r >= 0x1F3FB && r <= 0x1F3FF
}

func isEmojiModifierBase(r rune) bool {
	if isEmojiBase(r) {
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

func isRegionalIndicator(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
}

func isRecognizedRegionalIndicatorCluster(cluster []rune) bool {
	return len(cluster) == 2 &&
		isRegionalIndicator(cluster[0]) &&
		isRegionalIndicator(cluster[1]) &&
		isRecognizedRegionalIndicatorPair(cluster[0], cluster[1])
}

func isRecognizedRegionalIndicatorPair(left rune, right rune) bool {
	code := string([]rune{
		rune('A' + (left - 0x1F1E6)),
		rune('A' + (right - 0x1F1E6)),
	})

	_, ok := supportedFlagRegionCodes[code]
	return ok
}

func isRegionalIndicatorCluster(cluster []rune) bool {
	baseIndex := firstVisibleRuneIndex(cluster)
	return baseIndex >= 0 && isRegionalIndicator(cluster[baseIndex])
}

func isKeycapBase(r rune) bool {
	return r == '#' || r == '*' || (r >= '0' && r <= '9')
}

func isEmojiBase(r rune) bool {
	if isRegionalIndicator(r) {
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

func isEmojiVariationBase(r rune) bool {
	if isEmojiBase(r) {
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

func measureBaseVisibleClusterWidth(cluster []rune) int {
	baseIndex := firstVisibleRuneIndex(cluster)
	if baseIndex < 0 {
		return 0
	}

	width := utils.RuneWidth(cluster[baseIndex])
	for _, r := range cluster[baseIndex+1:] {
		if r >= 0xFF00 && r <= 0xFFEF {
			width += utils.RuneWidth(r)
		}
	}

	return width
}

func firstVisibleRuneIndex(cluster []rune) int {
	for index, r := range cluster {
		if !isZeroWidthControl(r) && !isZeroWidthMark(r) {
			return index
		}
	}

	return -1
}

func isDoubleWidthEmojiCluster(cluster []rune) bool {
	switch {
	case isRecognizedRegionalIndicatorCluster(cluster):
		return true
	case isKeycapCluster(cluster):
		return true
	case isEmojiZWJSequenceCluster(cluster):
		return true
	case isSimpleEmojiPresentationCluster(cluster):
		return true
	default:
		return false
	}
}

func isKeycapCluster(cluster []rune) bool {
	if len(cluster) < 2 || len(cluster) > 3 {
		return false
	}

	if !isKeycapBase(cluster[0]) {
		return false
	}

	if len(cluster) == 2 {
		return cluster[1] == 0x20E3
	}

	return cluster[1] == 0xFE0F && cluster[2] == 0x20E3
}

func isEmojiZWJSequenceCluster(cluster []rune) bool {
	pictographics := 0
	containsZWJ := false

	for _, r := range cluster {
		switch {
		case r == 0x200D:
			containsZWJ = true
		case isEmojiZWJJoinableBase(r):
			pictographics++
		}
	}

	return containsZWJ && pictographics >= 2
}

func isEmojiZWJJoinableBase(r rune) bool {
	return isEmojiVariationBase(r) && !isRegionalIndicator(r) && !isEmojiModifier(r)
}

func isSimpleEmojiPresentationCluster(cluster []rune) bool {
	baseIndex := firstVisibleRuneIndex(cluster)
	if baseIndex < 0 || baseIndex >= len(cluster) {
		return false
	}

	base := cluster[baseIndex]
	if !isEmojiVariationBase(base) {
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
		case isEmojiModifier(r) && isEmojiModifierBase(base):
			// Emoji modifiers keep the cluster in emoji presentation.
			sawEmojiModifier = true
		case isZeroWidthMark(r):
			return false
		default:
			return false
		}
	}

	return sawVariationSelector || sawEmojiModifier
}

func stripANSI(text string) string {
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
