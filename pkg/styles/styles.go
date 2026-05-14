package styles

import (
	"fmt"
	"math"
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
	Black         = basicColor{"black", 30, 40}
	Red           = basicColor{"red", 31, 41}
	Green         = basicColor{"green", 32, 42}
	Yellow        = basicColor{"yellow", 33, 43}
	Blue          = basicColor{"blue", 34, 44}
	Magenta       = basicColor{"magenta", 35, 45}
	Cyan          = basicColor{"cyan", 36, 46}
	White         = basicColor{"white", 37, 47}
	BlackBright   = basicColor{"blackBright", 90, 100}
	Gray          = basicColor{"gray", 90, 100}
	Grey          = Gray
	RedBright     = basicColor{"redBright", 91, 101}
	GreenBright   = basicColor{"greenBright", 92, 102}
	YellowBright  = basicColor{"yellowBright", 93, 103}
	BlueBright    = basicColor{"blueBright", 94, 104}
	MagentaBright = basicColor{"magentaBright", 95, 105}
	CyanBright    = basicColor{"cyanBright", 96, 106}
	WhiteBright   = basicColor{"whiteBright", 97, 107}
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

// RGBToANSI256Index maps an RGB color to the nearest xterm 256-color palette
// index. Chalk uses this representation when the output color level is 256.
func RGBToANSI256Index(r, g, b uint8) uint8 {
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}

		return uint8(math.Round(((float64(r)-8)/247)*24) + 232)
	}

	red := int(math.Round(float64(r) / 255 * 5))
	green := int(math.Round(float64(g) / 255 * 5))
	blue := int(math.Round(float64(b) / 255 * 5))

	return uint8(16 + 36*red + 6*green + blue)
}

// DowngradeTruecolorANSIToANSI256 rewrites SGR truecolor sequences to their
// xterm 256-color equivalents while leaving all other escape sequences intact.
func DowngradeTruecolorANSIToANSI256(text string) string {
	if text == "" {
		return text
	}

	var builder strings.Builder
	for index := 0; index < len(text); {
		if text[index] != '\x1b' || index+2 >= len(text) || text[index+1] != '[' {
			builder.WriteByte(text[index])
			index++
			continue
		}

		end := strings.IndexByte(text[index+2:], 'm')
		if end < 0 {
			builder.WriteString(text[index:])
			break
		}

		end += index + 2
		sequence := text[index : end+1]
		body := text[index+2 : end]
		rewritten, ok := downgradeTruecolorSGRBody(body)
		if !ok {
			builder.WriteString(sequence)
		} else {
			builder.WriteString("\x1b[")
			builder.WriteString(rewritten)
			builder.WriteByte('m')
		}

		index = end + 1
	}

	return builder.String()
}

func downgradeTruecolorSGRBody(body string) (string, bool) {
	parts := strings.Split(body, ";")
	rewritten := make([]string, 0, len(parts))
	changed := false

	for index := 0; index < len(parts); index++ {
		if (parts[index] == "38" || parts[index] == "48") && index+4 < len(parts) && parts[index+1] == "2" {
			r, okR := parseSGRByte(parts[index+2])
			g, okG := parseSGRByte(parts[index+3])
			b, okB := parseSGRByte(parts[index+4])
			if okR && okG && okB {
				rewritten = append(rewritten, parts[index], "5", strconv.Itoa(int(RGBToANSI256Index(r, g, b))))
				index += 4
				changed = true
				continue
			}
		}

		rewritten = append(rewritten, parts[index])
	}

	if !changed {
		return "", false
	}

	return strings.Join(rewritten, ";"), true
}

func parseSGRByte(part string) (uint8, bool) {
	value, err := strconv.ParseUint(part, 10, 8)
	if err != nil {
		return 0, false
	}

	return uint8(value), true
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
	case "blackbright":
		return BlackBright, true
	case "gray", "grey":
		return Gray, true
	case "redbright":
		return RedBright, true
	case "greenbright":
		return GreenBright, true
	case "yellowbright":
		return YellowBright, true
	case "bluebright":
		return BlueBright, true
	case "magentabright":
		return MagentaBright, true
	case "cyanbright":
		return CyanBright, true
	case "whitebright":
		return WhiteBright, true
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
