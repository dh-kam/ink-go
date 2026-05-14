package ttyinput

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// UTF8Decoder mirrors Node's setEncoding("utf8") behavior for TTY reads: a
// read ending in the middle of a multi-byte rune is buffered until the next
// read completes it. Invalid high-bit bytes are still passed through once they
// are known not to be an incomplete suffix.
type UTF8Decoder struct {
	pending []byte
}

func (decoder *UTF8Decoder) Write(chunk []byte) string {
	if len(chunk) == 0 {
		return ""
	}

	data := make([]byte, 0, len(decoder.pending)+len(chunk))
	data = append(data, decoder.pending...)
	data = append(data, chunk...)

	complete, pending := splitCompleteUTF8(data)
	decoder.pending = append(decoder.pending[:0], pending...)
	return string(complete)
}

func (decoder *UTF8Decoder) Flush() string {
	if len(decoder.pending) == 0 {
		return ""
	}

	output := string(decoder.pending)
	decoder.pending = nil
	return output
}

func splitCompleteUTF8(data []byte) ([]byte, []byte) {
	if len(data) == 0 || utf8.Valid(data) {
		return data, nil
	}

	maxSuffix := utf8.UTFMax - 1
	if len(data) < maxSuffix {
		maxSuffix = len(data)
	}

	for suffixLen := 1; suffixLen <= maxSuffix; suffixLen++ {
		suffix := data[len(data)-suffixLen:]
		if !isIncompleteUTF8Suffix(suffix) {
			continue
		}

		prefix := data[:len(data)-suffixLen]
		if utf8.Valid(prefix) {
			return prefix, suffix
		}
	}

	return data, nil
}

func isIncompleteUTF8Suffix(suffix []byte) bool {
	if len(suffix) == 0 {
		return false
	}

	expected := expectedUTF8Length(suffix[0])
	if expected == 0 || len(suffix) >= expected {
		return false
	}

	for _, b := range suffix[1:] {
		if b&0xc0 != 0x80 {
			return false
		}
	}
	return true
}

func expectedUTF8Length(first byte) int {
	switch {
	case first < utf8.RuneSelf:
		return 1
	case first >= 0xc2 && first <= 0xdf:
		return 2
	case first >= 0xe0 && first <= 0xef:
		return 3
	case first >= 0xf0 && first <= 0xf4:
		return 4
	default:
		return 0
	}
}

func Run(reader io.Reader, handleInput func(string) error, shouldExit func(string) bool) error {
	if reader == nil {
		return fmt.Errorf("ttyinput: reader is nil")
	}
	if handleInput == nil {
		return fmt.Errorf("ttyinput: input handler is nil")
	}

	buffer := make([]byte, 1024)
	decoder := UTF8Decoder{}

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			exited, err := dispatchInput(decoder.Write(buffer[:n]), handleInput, shouldExit)
			if err != nil {
				return err
			}
			if exited {
				return nil
			}
		}

		if err == nil {
			continue
		}
		if err == io.EOF {
			_, err := dispatchInput(decoder.Flush(), handleInput, shouldExit)
			return err
		}
		return err
	}
}

func dispatchInput(input string, handleInput func(string) error, shouldExit func(string) bool) (bool, error) {
	if input == "" {
		return false, nil
	}
	if err := handleInput(input); err != nil {
		return false, err
	}
	if shouldExit != nil && shouldExit(input) {
		return true, nil
	}
	return false, nil
}
