package styles

import (
	"fmt"
	"strconv"
	"strings"
)

// ColorMode specifies whether color is for foreground or background
type ColorMode int

const (
	Foreground ColorMode = iota
	Background
)

// Color represents a terminal color
type Color interface {
	Name() string
	ToANSI(mode ColorMode) string
}

// basicColor represents a standard 16-color terminal color
type basicColor struct {
	name string
	fg   int // foreground code
	bg   int // background code
}

func (c basicColor) Name() string {
	return c.name
}

func (c basicColor) ToANSI(mode ColorMode) string {
	if mode == Foreground {
		return fmt.Sprintf("\x1b[%dm", c.fg)
	}
	return fmt.Sprintf("\x1b[%dm", c.bg)
}

// Standard colors
var (
	Black   = basicColor{"black", 30, 40}
	Red     = basicColor{"red", 31, 41}
	Green   = basicColor{"green", 32, 42}
	Yellow  = basicColor{"yellow", 33, 43}
	Blue    = basicColor{"blue", 34, 44}
	Magenta = basicColor{"magenta", 35, 45}
	Cyan    = basicColor{"cyan", 36, 46}
	White   = basicColor{"white", 37, 47}
)

// rgbColor represents a 24-bit RGB color
type rgbColor struct {
	r, g, b uint8
}

func (c rgbColor) Name() string {
	return fmt.Sprintf("rgb(%d,%d,%d)", c.r, c.g, c.b)
}

func (c rgbColor) ToANSI(mode ColorMode) string {
	if mode == Foreground {
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.r, c.g, c.b)
	}
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c.r, c.g, c.b)
}

// RGB creates an RGB color
func RGB(r, g, b uint8) Color {
	return rgbColor{r, g, b}
}

// ansi256Color represents an ANSI 256 color.
type ansi256Color struct {
	index uint8
}

func (c ansi256Color) Name() string {
	return fmt.Sprintf("ansi256(%d)", c.index)
}

func (c ansi256Color) ToANSI(mode ColorMode) string {
	if mode == Foreground {
		return fmt.Sprintf("\x1b[38;5;%dm", c.index)
	}

	return fmt.Sprintf("\x1b[48;5;%dm", c.index)
}

// ANSI256 creates an ANSI 256 color.
func ANSI256(index uint8) Color {
	return ansi256Color{index: index}
}

// ANSI codes
const (
	ansiReset     = "\x1b[0m"
	ansiBold      = "\x1b[1m"
	ansiDim       = "\x1b[2m"
	ansiItalic    = "\x1b[3m"
	ansiUnderline = "\x1b[4m"
	ansiInverse   = "\x1b[7m"
	ansiStrike    = "\x1b[9m"
)

// Colorize applies a color to text
func Colorize(text string, color Color, mode ColorMode) string {
	return color.ToANSI(mode) + text + ansiReset
}

// WrapWithANSI applies multiple ANSI codes and a single trailing reset.
func WrapWithANSI(text string, codes ...string) string {
	if len(codes) == 0 {
		return text
	}

	var builder strings.Builder
	for _, code := range codes {
		builder.WriteString(code)
	}

	builder.WriteString(text)
	builder.WriteString(ansiReset)

	return builder.String()
}

// Bold applies bold style to text
func Bold(text string) string {
	return ansiBold + text + ansiReset
}

// Dim applies dim style to text
func Dim(text string) string {
	return ansiDim + text + ansiReset
}

// Italic applies italic style to text
func Italic(text string) string {
	return ansiItalic + text + ansiReset
}

// Underline applies underline style to text
func Underline(text string) string {
	return ansiUnderline + text + ansiReset
}

// Inverse applies inverse video style to text.
func Inverse(text string) string {
	return ansiInverse + text + ansiReset
}

// Strikethrough applies strikethrough style to text
func Strikethrough(text string) string {
	return ansiStrike + text + ansiReset
}

// BoldCode returns the ANSI sequence for bold text.
func BoldCode() string {
	return ansiBold
}

// DimCode returns the ANSI sequence for dim text.
func DimCode() string {
	return ansiDim
}

// ItalicCode returns the ANSI sequence for italic text.
func ItalicCode() string {
	return ansiItalic
}

// UnderlineCode returns the ANSI sequence for underlined text.
func UnderlineCode() string {
	return ansiUnderline
}

// InverseCode returns the ANSI sequence for inverse text.
func InverseCode() string {
	return ansiInverse
}

// StrikethroughCode returns the ANSI sequence for strikethrough text.
func StrikethroughCode() string {
	return ansiStrike
}

// ParseColor parses an Ink-compatible color string for the subset supported by this port.
func ParseColor(spec string) (Color, bool) {
	switch strings.ToLower(strings.TrimSpace(spec)) {
	case "black":
		return Black, true
	case "red":
		return Red, true
	case "green":
		return Green, true
	case "yellow":
		return Yellow, true
	case "blue":
		return Blue, true
	case "magenta":
		return Magenta, true
	case "cyan":
		return Cyan, true
	case "white":
		return White, true
	}

	if strings.HasPrefix(spec, "#") && len(spec) == 7 {
		r, errR := strconv.ParseUint(spec[1:3], 16, 8)
		g, errG := strconv.ParseUint(spec[3:5], 16, 8)
		b, errB := strconv.ParseUint(spec[5:7], 16, 8)
		if errR == nil && errG == nil && errB == nil {
			return RGB(uint8(r), uint8(g), uint8(b)), true
		}
	}

	if strings.HasPrefix(spec, "rgb(") && strings.HasSuffix(spec, ")") {
		body := strings.TrimSuffix(strings.TrimPrefix(spec, "rgb("), ")")
		parts := strings.Split(body, ",")
		if len(parts) != 3 {
			return nil, false
		}

		values := [3]uint8{}
		for index, part := range parts {
			value, err := strconv.ParseUint(strings.TrimSpace(part), 10, 8)
			if err != nil {
				return nil, false
			}

			values[index] = uint8(value)
		}

		return RGB(values[0], values[1], values[2]), true
	}

	if strings.HasPrefix(spec, "ansi256(") && strings.HasSuffix(spec, ")") {
		body := strings.TrimSuffix(strings.TrimPrefix(spec, "ansi256("), ")")
		value, err := strconv.ParseUint(strings.TrimSpace(body), 10, 8)
		if err == nil {
			return ANSI256(uint8(value)), true
		}
	}

	return nil, false
}

// Reset returns the ANSI reset code
func Reset() string {
	return ansiReset
}
