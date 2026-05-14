package ink_test

import (
	"testing"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func TestMeasureElementNilReturnsZero(t *testing.T) {
	if got := ink.MeasureElement(nil); got != (ink.ElementDimensions{}) {
		t.Fatalf("expected zero dimensions for nil node, got %+v", got)
	}
}

func TestMeasureTextMatchesUpstreamConstructorWidth(t *testing.T) {
	if got := ink.MeasureText("constructor"); got != (ink.ElementDimensions{Width: 11, Height: 1}) {
		t.Fatalf("expected text measurement {11 1}, got %+v", got)
	}
}

func TestMeasureTextUsesWidestLineAndLineCount(t *testing.T) {
	if got := ink.MeasureText("go\nconstructor\n日"); got != (ink.ElementDimensions{Width: 11, Height: 3}) {
		t.Fatalf("expected text measurement {11 3}, got %+v", got)
	}
}

func TestMeasureTextTreatsWideRunesAsDoubleWidth(t *testing.T) {
	if got := ink.MeasureText("日"); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected wide rune measurement {2 1}, got %+v", got)
	}
}

func TestMeasureTextIgnoresANSIEscapeSequences(t *testing.T) {
	if got := ink.MeasureText("\x1b[31mgo\x1b[0m\n\x1b]8;;https://example.com\x07日\x1b]8;;\x07"); got != (ink.ElementDimensions{Width: 2, Height: 2}) {
		t.Fatalf("expected ANSI-aware text measurement {2 2}, got %+v", got)
	}
}

func TestMeasureTextIgnoresANSIOSCSTEscapeSequences(t *testing.T) {
	if got := ink.MeasureText("\x1b]8;;https://example.com\x1b\\ok\x1b]8;;\x1b\\"); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected OSC-ST-aware text measurement {2 1}, got %+v", got)
	}
}

func TestMeasureTextIgnoresControlCharactersInWidth(t *testing.T) {
	if got := ink.MeasureText("a\r\nb\tc\x00"); got != (ink.ElementDimensions{Width: 2, Height: 2}) {
		t.Fatalf("expected control-aware text measurement {2 2}, got %+v", got)
	}
}

func TestMeasureTextIgnoresUnicodeFormatCharactersInWidth(t *testing.T) {
	if got := ink.MeasureText("go\u200e\n🏴\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F"); got != (ink.ElementDimensions{Width: 2, Height: 2}) {
		t.Fatalf("expected format-aware text measurement {2 2}, got %+v", got)
	}
}

func TestMeasureTextCollapsesEmojiSequencesToSingleClusterWidth(t *testing.T) {
	if got := ink.MeasureText("👨‍👩‍👧‍👦\n👍🏽"); got != (ink.ElementDimensions{Width: 2, Height: 2}) {
		t.Fatalf("expected emoji-sequence measurement {2 2}, got %+v", got)
	}
}

func TestMeasureTextDoesNotCollapseZWJSeparatedPlainText(t *testing.T) {
	if got := ink.MeasureText("a\u200db"); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected non-emoji ZWJ text measurement {2 1}, got %+v", got)
	}
}

func TestMeasureTextDoesNotCollapseMixedEmojiAndPlainTextAcrossZWJ(t *testing.T) {
	if got := ink.MeasureText("a\u200d🙂\n🙂\u200db"); got != (ink.ElementDimensions{Width: 3, Height: 2}) {
		t.Fatalf("expected mixed ZWJ text measurement {3 2}, got %+v", got)
	}
}

func TestMeasureTextCountsAdjacentEmojiClustersIndividually(t *testing.T) {
	if got := ink.MeasureText("👨‍👩‍👧‍👦👍🏽"); got != (ink.ElementDimensions{Width: 4, Height: 1}) {
		t.Fatalf("expected adjacent emoji clusters to measure {4 1}, got %+v", got)
	}
}

func TestMeasureTextTreatsEmojiVariationSequencesAsWide(t *testing.T) {
	if got := ink.MeasureText("♥️✈️™️©️"); got != (ink.ElementDimensions{Width: 8, Height: 1}) {
		t.Fatalf("expected emoji-variation measurement {8 1}, got %+v", got)
	}
}

func TestMeasureTextTreatsEmojiModifierPresentationSequencesAsWide(t *testing.T) {
	if got := ink.MeasureText("✌🏽"); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected emoji-modifier measurement {2 1}, got %+v", got)
	}
}

func TestMeasureTextTreatsRecognizedFlagsAsWideButLoneIndicatorsAsSingleWidth(t *testing.T) {
	if got := ink.MeasureText("🇺🇸🇨"); got != (ink.ElementDimensions{Width: 3, Height: 1}) {
		t.Fatalf("expected recognized-flag measurement {3 1}, got %+v", got)
	}
}

func TestMeasureTextTreatsUnsupportedRegionalIndicatorPairsAsSingleWidthCluster(t *testing.T) {
	if got := ink.MeasureText("🇦🇧🇨"); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected unsupported-flag measurement {2 1}, got %+v", got)
	}
}

func TestMeasureTextTreatsUnsupportedRegionalIndicatorPairItselfAsSingleWidth(t *testing.T) {
	if got := ink.MeasureText("🇦🇧"); got != (ink.ElementDimensions{Width: 1, Height: 1}) {
		t.Fatalf("expected unsupported-flag pair measurement {1 1}, got %+v", got)
	}
}

func TestMeasureTextTreatsDeprecatedRegionalIndicatorPairsAsSingleWidth(t *testing.T) {
	if got := ink.MeasureText("🇸🇺🇨🇸🇧🇺"); got != (ink.ElementDimensions{Width: 3, Height: 1}) {
		t.Fatalf("expected deprecated-flag measurement {3 1}, got %+v", got)
	}
}

func TestMeasureTextTreatsCurrentRegionalIndicatorExceptionsAsWide(t *testing.T) {
	if got := ink.MeasureText("🇪🇺🇺🇳🇽🇰"); got != (ink.ElementDimensions{Width: 6, Height: 1}) {
		t.Fatalf("expected current-flag exception measurement {6 1}, got %+v", got)
	}
}

func TestMeasureTextTreatsRegionalIndicatorClustersWithExtraMarksAsSingleWidth(t *testing.T) {
	if got := ink.MeasureText("🇺🇸️\n🇦️"); got != (ink.ElementDimensions{Width: 1, Height: 2}) {
		t.Fatalf("expected marked regional-indicator measurement {1 2}, got %+v", got)
	}
}

func TestMeasureTextDropsEmojiPresentationForDanglingZWJVariationClusters(t *testing.T) {
	if got := ink.MeasureText("❤️\u200d\n™️\u200d"); got != (ink.ElementDimensions{Width: 1, Height: 2}) {
		t.Fatalf("expected dangling-ZWJ variation measurement {1 2}, got %+v", got)
	}
}

func TestMeasureTextDoesNotCollapseEmojiToRegionalIndicatorAcrossZWJ(t *testing.T) {
	if got := ink.MeasureText("🙂\u200d🇨\n👨\u200d🇨"); got != (ink.ElementDimensions{Width: 3, Height: 2}) {
		t.Fatalf("expected emoji-to-regional-indicator measurement {3 2}, got %+v", got)
	}
}

func TestMeasureTextEmptyStringReturnsZero(t *testing.T) {
	if got := ink.MeasureText(""); got != (ink.ElementDimensions{}) {
		t.Fatalf("expected zero text measurement, got %+v", got)
	}
}

func TestMeasureElementUsesComputedLayout(t *testing.T) {
	node := vdom.CreateElement("box", nil)
	node.Layout = vdom.Layout{Width: 12, Height: 3}

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 12, Height: 3}) {
		t.Fatalf("expected measured layout {12 3}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallsBackToIntrinsicTextMeasurement(t *testing.T) {
	node := vdom.CreateTextNode("constructor")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 11, Height: 1}) {
		t.Fatalf("expected intrinsic text measurement {11 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackUsesPublicMeasureTextSemantics(t *testing.T) {
	node := vdom.CreateTextNode("go\n日")

	if got := ink.MeasureElement(node); got != ink.MeasureText("go\n日") {
		t.Fatalf("expected text-node fallback to match MeasureText, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackIgnoresANSIEscapeSequences(t *testing.T) {
	node := vdom.CreateTextNode("\x1b[32mok\x1b[0m")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected ANSI-aware intrinsic text measurement {2 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackIgnoresANSIOSCSTEscapeSequences(t *testing.T) {
	node := vdom.CreateTextNode("\x1b]8;;https://example.com\x1b\\ok\x1b]8;;\x1b\\")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected OSC-ST-aware intrinsic text measurement {2 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackCollapsesEmojiSequences(t *testing.T) {
	node := vdom.CreateTextNode("👨‍👩‍👧‍👦")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected emoji-aware intrinsic text measurement {2 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackDoesNotCollapseMixedZWJText(t *testing.T) {
	node := vdom.CreateTextNode("A\u200d👨")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 3, Height: 1}) {
		t.Fatalf("expected mixed ZWJ intrinsic text measurement {3 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackTreatsEmojiVariationSequencesAsWide(t *testing.T) {
	node := vdom.CreateTextNode("♥️")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected emoji-variation intrinsic text measurement {2 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackTreatsEmojiModifierPresentationSequencesAsWide(t *testing.T) {
	node := vdom.CreateTextNode("✌🏽")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 2, Height: 1}) {
		t.Fatalf("expected emoji-modifier intrinsic text measurement {2 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackMatchesRegionalIndicatorSemantics(t *testing.T) {
	node := vdom.CreateTextNode("🇺🇸🇨")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 3, Height: 1}) {
		t.Fatalf("expected regional-indicator intrinsic text measurement {3 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackMatchesUnsupportedRegionalIndicatorPairSemantics(t *testing.T) {
	node := vdom.CreateTextNode("🇦🇧")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 1, Height: 1}) {
		t.Fatalf("expected unsupported regional-indicator pair measurement {1 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackTreatsDeprecatedRegionalIndicatorPairsAsSingleWidth(t *testing.T) {
	node := vdom.CreateTextNode("🇸🇺🇨🇸🇧🇺")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 3, Height: 1}) {
		t.Fatalf("expected deprecated regional-indicator measurement {3 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackTreatsMarkedRegionalIndicatorClustersAsSingleWidth(t *testing.T) {
	node := vdom.CreateTextNode("🇺🇸️")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 1, Height: 1}) {
		t.Fatalf("expected marked regional-indicator measurement {1 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodeFallbackDropsEmojiPresentationForDanglingZWJVariationClusters(t *testing.T) {
	node := vdom.CreateTextNode("❤️\u200d")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 1, Height: 1}) {
		t.Fatalf("expected dangling-ZWJ variation intrinsic text measurement {1 1}, got %+v", got)
	}
}

func TestMeasureElementTextNodePrefersComputedLayout(t *testing.T) {
	node := vdom.CreateTextNode("constructor")
	node.Layout = vdom.Layout{Width: 4, Height: 2}

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 4, Height: 2}) {
		t.Fatalf("expected computed layout {4 2}, got %+v", got)
	}
}

func TestMeasureElementTextNodeRecomputesAfterNodeValueChanges(t *testing.T) {
	node := vdom.CreateTextNode("go")
	node.Layout = vdom.Layout{Width: 2, Height: 1}
	node.SetNodeValue("constructor")

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{Width: 11, Height: 1}) {
		t.Fatalf("expected updated intrinsic text measurement {11 1}, got %+v", got)
	}
}

func TestMeasureElementTextElementWithoutLayoutRemainsZero(t *testing.T) {
	node := vdom.CreateElement("text", nil, vdom.CreateTextNode("constructor"))

	if got := ink.MeasureElement(node); got != (ink.ElementDimensions{}) {
		t.Fatalf("expected zero dimensions before element layout sync, got %+v", got)
	}
}
