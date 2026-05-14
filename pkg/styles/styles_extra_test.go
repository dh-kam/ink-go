package styles

import (
	"strings"
	"testing"
)

// --- RGB color tests ---

func TestRGB_ToANSI_Foreground(t *testing.T) {
	c := RGB(10, 20, 30)
	got := c.ToANSI(Foreground)
	want := "\x1b[38;2;10;20;30m"
	if got != want {
		t.Fatalf("RGB foreground ToANSI = %q, want %q", got, want)
	}
}

func TestRGB_ToANSI_Background(t *testing.T) {
	c := RGB(10, 20, 30)
	got := c.ToANSI(Background)
	want := "\x1b[48;2;10;20;30m"
	if got != want {
		t.Fatalf("RGB background ToANSI = %q, want %q", got, want)
	}
}

func TestRGB_Name(t *testing.T) {
	c := RGB(255, 128, 64)
	if got, want := c.Name(), "rgb(255,128,64)"; got != want {
		t.Fatalf("RGB Name = %q, want %q", got, want)
	}
}

func TestRGB_Boundary(t *testing.T) {
	// 0 and 255 boundaries
	c := RGB(0, 0, 0)
	if got, want := c.ToANSI(Foreground), "\x1b[38;2;0;0;0m"; got != want {
		t.Fatalf("RGB(0,0,0) fg = %q, want %q", got, want)
	}
	c = RGB(255, 255, 255)
	if got, want := c.ToANSI(Background), "\x1b[48;2;255;255;255m"; got != want {
		t.Fatalf("RGB(255,255,255) bg = %q, want %q", got, want)
	}
}

// --- ANSI256 color tests ---

func TestANSI256_ToANSI_Foreground(t *testing.T) {
	c := ANSI256(123)
	got := c.ToANSI(Foreground)
	want := "\x1b[38;5;123m"
	if got != want {
		t.Fatalf("ANSI256 fg = %q, want %q", got, want)
	}
}

func TestANSI256_ToANSI_Background(t *testing.T) {
	c := ANSI256(200)
	got := c.ToANSI(Background)
	want := "\x1b[48;5;200m"
	if got != want {
		t.Fatalf("ANSI256 bg = %q, want %q", got, want)
	}
}

func TestANSI256_Name(t *testing.T) {
	c := ANSI256(42)
	if got, want := c.Name(), "ansi256(42)"; got != want {
		t.Fatalf("ANSI256 Name = %q, want %q", got, want)
	}
}

func TestRGBToANSI256Index(t *testing.T) {
	tests := []struct {
		name string
		r    uint8
		g    uint8
		b    uint8
		want uint8
	}{
		{name: "ink box backgrounds hex orange", r: 255, g: 136, b: 0, want: 214},
		{name: "ink box backgrounds rgb green", r: 0, g: 255, b: 0, want: 46},
		{name: "black", r: 0, g: 0, b: 0, want: 16},
		{name: "white", r: 255, g: 255, b: 255, want: 231},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RGBToANSI256Index(tt.r, tt.g, tt.b); got != tt.want {
				t.Fatalf("RGBToANSI256Index(%d,%d,%d) = %d, want %d", tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}

func TestDowngradeTruecolorANSIToANSI256(t *testing.T) {
	input := "\x1b[48;2;255;136;0mHex\x1b[39m \x1b[38;2;0;255;0;1mRGB\x1b[0m"
	want := "\x1b[48;5;214mHex\x1b[39m \x1b[38;5;46;1mRGB\x1b[0m"

	if got := DowngradeTruecolorANSIToANSI256(input); got != want {
		t.Fatalf("DowngradeTruecolorANSIToANSI256 = %q, want %q", got, want)
	}
}

func TestDowngradeTruecolorANSIToANSI256LeavesOtherSequences(t *testing.T) {
	input := "\x1b[48;5;214mHex\x1b[?25l\x1b[31mRed\x1b[0m"
	if got := DowngradeTruecolorANSIToANSI256(input); got != input {
		t.Fatalf("DowngradeTruecolorANSIToANSI256 changed unrelated sequences: %q", got)
	}
}

// --- Colorize edge cases ---

func TestColorize_EmptyString(t *testing.T) {
	got := Colorize("", Red, Foreground)
	want := "\x1b[31m" + "" + "\x1b[0m"
	if got != want {
		t.Fatalf("Colorize empty = %q, want %q", got, want)
	}
}

func TestColorize_BackgroundMode(t *testing.T) {
	got := Colorize("hi", Blue, Background)
	want := "\x1b[44mhi\x1b[0m"
	if got != want {
		t.Fatalf("Colorize bg = %q, want %q", got, want)
	}
}

func TestColorize_RGBColor(t *testing.T) {
	got := Colorize("x", RGB(1, 2, 3), Foreground)
	want := "\x1b[38;2;1;2;3mx\x1b[0m"
	if got != want {
		t.Fatalf("Colorize rgb = %q, want %q", got, want)
	}
}

// --- WrapWithANSI tests ---

func TestWrapWithANSI_Empty(t *testing.T) {
	got := WrapWithANSI("hello")
	if got != "hello" {
		t.Fatalf("WrapWithANSI no codes = %q, want %q", got, "hello")
	}
}

func TestWrapWithANSI_SingleCode(t *testing.T) {
	got := WrapWithANSI("hi", BoldCode())
	want := "\x1b[1mhi\x1b[0m"
	if got != want {
		t.Fatalf("WrapWithANSI single = %q, want %q", got, want)
	}
}

func TestWrapWithANSI_MultipleCodes(t *testing.T) {
	got := WrapWithANSI("hi", BoldCode(), UnderlineCode(), ItalicCode())
	want := "\x1b[1m\x1b[4m\x1b[3mhi\x1b[0m"
	if got != want {
		t.Fatalf("WrapWithANSI multi = %q, want %q", got, want)
	}
}

func TestWrapWithANSI_EmptyText(t *testing.T) {
	got := WrapWithANSI("", BoldCode())
	want := "\x1b[1m\x1b[0m"
	if got != want {
		t.Fatalf("WrapWithANSI empty text = %q, want %q", got, want)
	}
}

// --- ParseColor: known names ---

func TestParseColor_KnownNames(t *testing.T) {
	cases := map[string]basicColor{
		"black":   Black,
		"red":     Red,
		"green":   Green,
		"yellow":  Yellow,
		"blue":    Blue,
		"magenta": Magenta,
		"cyan":    Cyan,
		"white":   White,
	}
	for name, want := range cases {
		got, ok := ParseColor(name)
		if !ok {
			t.Errorf("ParseColor(%q) not ok", name)
			continue
		}
		if got.Name() != want.Name() {
			t.Errorf("ParseColor(%q) = %s, want %s", name, got.Name(), want.Name())
		}
	}
}

func TestParseColor_KnownNames_CaseAndSpace(t *testing.T) {
	got, ok := ParseColor("  RED  ")
	if !ok {
		t.Fatalf("ParseColor case/space not ok")
	}
	if got.Name() != Red.Name() {
		t.Fatalf("ParseColor(' RED ') = %s, want %s", got.Name(), Red.Name())
	}
}

func TestParseColor_ChalkBrightAndGrayNames(t *testing.T) {
	cases := map[string]Color{
		"gray":          Gray,
		"grey":          Gray,
		"blackBright":   BlackBright,
		"redBright":     RedBright,
		"greenBright":   GreenBright,
		"yellowBright":  YellowBright,
		"blueBright":    BlueBright,
		"magentaBright": MagentaBright,
		"cyanBright":    CyanBright,
		"whiteBright":   WhiteBright,
	}

	for name, want := range cases {
		got, ok := ParseColor(name)
		if !ok {
			t.Errorf("ParseColor(%q) not ok", name)
			continue
		}
		if got.ToANSI(Foreground) != want.ToANSI(Foreground) {
			t.Errorf("ParseColor(%q) = %q, want %q", name, got.ToANSI(Foreground), want.ToANSI(Foreground))
		}
	}
}

// --- ParseColor: hex ---

func TestParseColor_Hex(t *testing.T) {
	got, ok := ParseColor("#FF8040")
	if !ok {
		t.Fatalf("ParseColor(#FF8040) not ok")
	}
	rgb, ok := got.(rgbColor)
	if !ok {
		t.Fatalf("ParseColor(#FF8040) returned %T, want rgbColor", got)
	}
	if rgb.r != 0xFF || rgb.g != 0x80 || rgb.b != 0x40 {
		t.Fatalf("ParseColor(#FF8040) = (%d,%d,%d), want (255,128,64)", rgb.r, rgb.g, rgb.b)
	}
}

func TestParseColor_HexLowercase(t *testing.T) {
	// The current implementation lowercases the spec via ToLower for the name switch,
	// but for hex it passes the original spec. Verify lowercase hex parses too.
	got, ok := ParseColor("#abcdef")
	if !ok {
		t.Fatalf("ParseColor(#abcdef) not ok")
	}
	rgb := got.(rgbColor)
	if rgb.r != 0xab || rgb.g != 0xcd || rgb.b != 0xef {
		t.Fatalf("ParseColor(#abcdef) = (%d,%d,%d)", rgb.r, rgb.g, rgb.b)
	}
}

func TestParseColor_HexBadLength(t *testing.T) {
	if _, ok := ParseColor("#FFF"); ok {
		t.Errorf("ParseColor(#FFF) unexpectedly ok")
	}
	if _, ok := ParseColor("#FFFFFFF"); ok {
		t.Errorf("ParseColor(#FFFFFFF) unexpectedly ok")
	}
}

func TestParseColor_HexBadDigits(t *testing.T) {
	if _, ok := ParseColor("#GGGGGG"); ok {
		t.Errorf("ParseColor(#GGGGGG) unexpectedly ok")
	}
	if _, ok := ParseColor("#ZZ1122"); ok {
		t.Errorf("ParseColor(#ZZ1122) unexpectedly ok")
	}
}

// --- ParseColor: rgb(...) ---

func TestParseColor_RGB(t *testing.T) {
	got, ok := ParseColor("rgb(10, 20, 30)")
	if !ok {
		t.Fatalf("ParseColor(rgb(10,20,30)) not ok")
	}
	rgb := got.(rgbColor)
	if rgb.r != 10 || rgb.g != 20 || rgb.b != 30 {
		t.Fatalf("ParseColor rgb = (%d,%d,%d)", rgb.r, rgb.g, rgb.b)
	}
}

func TestParseColor_RGB_NoSpaces(t *testing.T) {
	got, ok := ParseColor("rgb(0,0,0)")
	if !ok {
		t.Fatalf("ParseColor(rgb(0,0,0)) not ok")
	}
	if got.Name() != "rgb(0,0,0)" {
		t.Fatalf("Name = %s", got.Name())
	}
}

func TestParseColor_RGB_BadParts(t *testing.T) {
	cases := []string{
		"rgb(1,2)",     // too few
		"rgb(1,2,3,4)", // too many
		"rgb(a,b,c)",   // not numeric
		"rgb(256,0,0)", // out of uint8 range
		"rgb(-1,0,0)",  // negative
		"rgb(1,2,3",    // missing close paren — won't match suffix
		"rgb 1,2,3)",   // missing prefix
	}
	for _, c := range cases {
		if _, ok := ParseColor(c); ok {
			t.Errorf("ParseColor(%q) unexpectedly ok", c)
		}
	}
}

// --- ParseColor: ansi256(...) ---

func TestParseColor_ANSI256(t *testing.T) {
	got, ok := ParseColor("ansi256(42)")
	if !ok {
		t.Fatalf("ParseColor(ansi256(42)) not ok")
	}
	a := got.(ansi256Color)
	if a.index != 42 {
		t.Fatalf("ParseColor ansi256 index = %d, want 42", a.index)
	}
}

func TestParseColor_ANSI256_BadValue(t *testing.T) {
	cases := []string{
		"ansi256(256)", // out of uint8 range
		"ansi256(-1)",  // negative
		"ansi256(abc)", // not numeric
		"ansi256()",    // empty inside
	}
	for _, c := range cases {
		if _, ok := ParseColor(c); ok {
			t.Errorf("ParseColor(%q) unexpectedly ok", c)
		}
	}
}

// --- ParseColor: invalid input ---

func TestParseColor_Invalid(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"not-a-color",
		"hsl(0,0,0)",
		"#",
		"rgb",
		"ansi256",
		"@@@",
	}
	for _, c := range cases {
		if _, ok := ParseColor(c); ok {
			t.Errorf("ParseColor(%q) unexpectedly ok", c)
		}
	}
}

// --- Style functions ---

func TestBold(t *testing.T) {
	got := Bold("x")
	want := "\x1b[1mx\x1b[0m"
	if got != want {
		t.Fatalf("Bold = %q, want %q", got, want)
	}
}

func TestDim(t *testing.T) {
	got := Dim("x")
	want := "\x1b[2mx\x1b[0m"
	if got != want {
		t.Fatalf("Dim = %q, want %q", got, want)
	}
}

func TestItalic(t *testing.T) {
	got := Italic("x")
	want := "\x1b[3mx\x1b[0m"
	if got != want {
		t.Fatalf("Italic = %q, want %q", got, want)
	}
}

func TestUnderline(t *testing.T) {
	got := Underline("x")
	want := "\x1b[4mx\x1b[0m"
	if got != want {
		t.Fatalf("Underline = %q, want %q", got, want)
	}
}

func TestInverse(t *testing.T) {
	got := Inverse("x")
	want := "\x1b[7mx\x1b[0m"
	if got != want {
		t.Fatalf("Inverse = %q, want %q", got, want)
	}
}

func TestStrikethrough(t *testing.T) {
	got := Strikethrough("x")
	want := "\x1b[9mx\x1b[0m"
	if got != want {
		t.Fatalf("Strikethrough = %q, want %q", got, want)
	}
}

// --- Code accessors ---

func TestCodeAccessors(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"BoldCode", BoldCode(), "\x1b[1m"},
		{"DimCode", DimCode(), "\x1b[2m"},
		{"ItalicCode", ItalicCode(), "\x1b[3m"},
		{"UnderlineCode", UnderlineCode(), "\x1b[4m"},
		{"InverseCode", InverseCode(), "\x1b[7m"},
		{"StrikethroughCode", StrikethroughCode(), "\x1b[9m"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// --- Reset ---

func TestReset_Exact(t *testing.T) {
	if got := Reset(); got != "\x1b[0m" {
		t.Fatalf("Reset = %q, want %q", got, "\x1b[0m")
	}
}

// --- Ensure all style funcs end with reset ---

func TestStyleFuncs_AllEndWithReset(t *testing.T) {
	funcs := map[string]func(string) string{
		"Bold":          Bold,
		"Dim":           Dim,
		"Italic":        Italic,
		"Underline":     Underline,
		"Inverse":       Inverse,
		"Strikethrough": Strikethrough,
	}
	for name, fn := range funcs {
		out := fn("test")
		if !strings.HasSuffix(out, Reset()) {
			t.Errorf("%s output does not end with Reset(): %q", name, out)
		}
		if !strings.Contains(out, "test") {
			t.Errorf("%s output does not contain text: %q", name, out)
		}
	}
}
