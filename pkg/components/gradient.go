package components

import (
	"strings"

	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// GradientProps configures a Gradient text rendering.
//
// From and To are explicit 24-bit RGB tuples ([R, G, B]). We use raw
// triples — rather than the styles.Color interface — because the styles
// package keeps its rgbColor type unexported, leaving us no way to
// recover the channels needed for interpolation. Concrete tuples also
// keep the API allocation-free for the most common case.
type GradientProps struct {
	Text string
	From [3]uint8
	To   [3]uint8
}

// Gradient renders Text with a per-rune linear RGB color gradient
// interpolated between From and To.
//
// Empty Text → empty text node. Single-rune Text uses From verbatim
// (avoids a divide-by-zero in the interpolation). Equal From and To
// still produce colored output (just constant), matching the JS
// ink-gradient behavior.
func Gradient(props GradientProps) *vdom.Node {
	if props.Text == "" {
		return vdom.CreateTextNode("")
	}

	// Iterate by rune so multi-byte glyphs are colored as a unit.
	runes := []rune(props.Text)
	if len(runes) == 1 {
		return vdom.CreateTextNode(rgbForeground(props.From) + string(runes) + styles.Reset())
	}

	var b strings.Builder
	// Rough preallocation: each colored rune adds ~20 bytes of escape codes.
	b.Grow(len(props.Text) + len(runes)*24)

	denom := float64(len(runes) - 1)
	for index, r := range runes {
		t := float64(index) / denom
		color := lerpRGB(props.From, props.To, t)
		b.WriteString(rgbForeground(color))
		b.WriteRune(r)
		b.WriteString(styles.Reset())
	}

	return vdom.CreateTextNode(b.String())
}

// lerpRGB linearly interpolates between two RGB tuples at parameter t
// (0 → from, 1 → to). t is clamped implicitly by the caller's index range.
func lerpRGB(from, to [3]uint8, t float64) [3]uint8 {
	return [3]uint8{
		lerpChannel(from[0], to[0], t),
		lerpChannel(from[1], to[1], t),
		lerpChannel(from[2], to[2], t),
	}
}

func lerpChannel(from, to uint8, t float64) uint8 {
	// Cast to float64, interpolate, then round to nearest integer.
	value := float64(from) + (float64(to)-float64(from))*t
	if value < 0 {
		value = 0
	}
	if value > 255 {
		value = 255
	}
	// + 0.5 for round-half-up — gives us a symmetric gradient endpoint.
	return uint8(value + 0.5)
}

// rgbForeground builds the SGR truecolor foreground escape directly,
// avoiding a styles.Color allocation per rune. Equivalent to
// styles.RGB(r,g,b).ToANSI(styles.Foreground).
func rgbForeground(c [3]uint8) string {
	// "\x1b[38;2;R;G;Bm" — assemble manually to keep this hot path tight.
	var b strings.Builder
	b.Grow(19)
	b.WriteString("\x1b[38;2;")
	writeUint8(&b, c[0])
	b.WriteByte(';')
	writeUint8(&b, c[1])
	b.WriteByte(';')
	writeUint8(&b, c[2])
	b.WriteByte('m')
	return b.String()
}

func writeUint8(b *strings.Builder, v uint8) {
	if v >= 100 {
		b.WriteByte('0' + v/100)
		b.WriteByte('0' + (v/10)%10)
		b.WriteByte('0' + v%10)
		return
	}
	if v >= 10 {
		b.WriteByte('0' + v/10)
		b.WriteByte('0' + v%10)
		return
	}
	b.WriteByte('0' + v)
}
