package ttyinput

import (
	"io"
	"strings"
	"testing"
)

func TestUTF8DecoderBuffersSplitHangulRune(t *testing.T) {
	decoder := UTF8Decoder{}
	hangul := []byte("가")

	if got := decoder.Write(append([]byte("hello "), hangul[:1]...)); got != "hello " {
		t.Fatalf("first write = %q, want %q", got, "hello ")
	}
	if got := decoder.Write(hangul[1:]); got != "가" {
		t.Fatalf("second write = %q, want %q", got, "가")
	}
}

func TestUTF8DecoderBuffersTwoByteIncompleteSuffix(t *testing.T) {
	decoder := UTF8Decoder{}
	input := []byte("abc가나다")
	splitAt := len([]byte("abc가나")) + 2

	if got := decoder.Write(input[:splitAt]); got != "abc가나" {
		t.Fatalf("first write = %q, want %q", got, "abc가나")
	}
	if got := decoder.Write(input[splitAt:]); got != "다" {
		t.Fatalf("second write = %q, want %q", got, "다")
	}
}

func TestUTF8DecoderPassesInvalidHighByteAfterItIsNotIncomplete(t *testing.T) {
	decoder := UTF8Decoder{}

	if got := decoder.Write([]byte{0xe9}); got != "" {
		t.Fatalf("first write = %q, want empty pending output", got)
	}
	if got := decoder.Write([]byte("x")); got != string([]byte{0xe9, 'x'}) {
		t.Fatalf("second write = %q, want raw invalid byte followed by x", got)
	}
}

func TestRunDispatchesOnlyCompleteUTF8Input(t *testing.T) {
	hangul := []byte("가")
	reader := io.MultiReader(
		strings.NewReader("hello "),
		strings.NewReader(string(hangul[:1])),
		strings.NewReader(string(hangul[1:])),
		strings.NewReader("\x03"),
	)

	var inputs []string
	err := Run(reader, func(input string) error {
		inputs = append(inputs, input)
		return nil
	}, func(input string) bool {
		return strings.Contains(input, "\x03")
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	want := []string{"hello ", "가", "\x03"}
	if strings.Join(inputs, "|") != strings.Join(want, "|") {
		t.Fatalf("inputs = %#v, want %#v", inputs, want)
	}
}
