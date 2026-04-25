package input

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Key represents a key press event
type Key struct {
	Char rune   // The character for regular keys (0 for special keys)
	Name string // The name of special keys (e.g., "escape", "up", "down")
}

// HookKey matches Ink's boolean-oriented key object shape used by useInput.
type HookKey struct {
	UpArrow    bool
	DownArrow  bool
	LeftArrow  bool
	RightArrow bool
	PageDown   bool
	PageUp     bool
	Home       bool
	End        bool
	Return     bool
	Escape     bool
	Ctrl       bool
	Shift      bool
	Tab        bool
	Backspace  bool
	Delete     bool
	Meta       bool
}

// String returns a string representation of the key
func (k Key) String() string {
	if k.Name != "" {
		return k.Name
	}
	return string(k.Char)
}

// Equals checks if two keys are equal
func (k Key) Equals(other Key) bool {
	return k.Char == other.Char && k.Name == other.Name
}

// Special key names
const (
	KeyEscape    = "escape"
	KeyUp        = "up"
	KeyDown      = "down"
	KeyLeft      = "left"
	KeyRight     = "right"
	KeyClear     = "clear"
	KeyInsert    = "insert"
	KeyReturn    = "return"
	KeyBackspace = "backspace"
	KeyTab       = "tab"
	KeyShiftTab  = "shift-tab"
	KeyDelete    = "delete"
	KeyHome      = "home"
	KeyEnd       = "end"
	KeyPageUp    = "pageup"
	KeyPageDown  = "pagedown"
	KeyF1        = "f1"
	KeyF2        = "f2"
	KeyF3        = "f3"
	KeyF4        = "f4"
	KeyF5        = "f5"
	KeyF6        = "f6"
	KeyF7        = "f7"
	KeyF8        = "f8"
	KeyF9        = "f9"
	KeyF10       = "f10"
	KeyF11       = "f11"
	KeyF12       = "f12"

	bracketedPasteStart = "\x1b[200~"
	bracketedPasteEnd   = "\x1b[201~"
)

var (
	hookMetaKeyCodeRe = regexp.MustCompile("^\x1b([a-zA-Z0-9])$")
	hookFnKeyRe       = regexp.MustCompile(`^(?:\x1b+)(O|N|\[|\[\[)(?:(\d+)(?:;(\d+))?([~^$])|(?:1;)?(\d+)?([a-zA-Z]))`)
	hookKeyNames      = map[string]string{
		"OP":   KeyF1,
		"OQ":   KeyF2,
		"OR":   KeyF3,
		"OS":   KeyF4,
		"[P":   KeyF1,
		"[Q":   KeyF2,
		"[R":   KeyF3,
		"[S":   KeyF4,
		"[11~": KeyF1,
		"[12~": KeyF2,
		"[13~": KeyF3,
		"[14~": KeyF4,
		"[[A":  KeyF1,
		"[[B":  KeyF2,
		"[[C":  KeyF3,
		"[[D":  KeyF4,
		"[[E":  KeyF5,
		"[15~": KeyF5,
		"[17~": KeyF6,
		"[18~": KeyF7,
		"[19~": KeyF8,
		"[20~": KeyF9,
		"[21~": KeyF10,
		"[23~": KeyF11,
		"[24~": KeyF12,
		"[A":   KeyUp,
		"[B":   KeyDown,
		"[C":   KeyRight,
		"[D":   KeyLeft,
		"[E":   KeyClear,
		"[F":   KeyEnd,
		"[H":   KeyHome,
		"OA":   KeyUp,
		"OB":   KeyDown,
		"OC":   KeyRight,
		"OD":   KeyLeft,
		"OE":   KeyClear,
		"OF":   KeyEnd,
		"OH":   KeyHome,
		"[1~":  KeyHome,
		"[2~":  KeyInsert,
		"[3~":  KeyDelete,
		"[4~":  KeyEnd,
		"[5~":  KeyPageUp,
		"[6~":  KeyPageDown,
		"[[5~": KeyPageUp,
		"[[6~": KeyPageDown,
		"[7~":  KeyHome,
		"[8~":  KeyEnd,
		"[a":   KeyUp,
		"[b":   KeyDown,
		"[c":   KeyRight,
		"[d":   KeyLeft,
		"[e":   KeyClear,
		"[2$":  KeyInsert,
		"[3$":  KeyDelete,
		"[5$":  KeyPageUp,
		"[6$":  KeyPageDown,
		"[7$":  KeyHome,
		"[8$":  KeyEnd,
		"Oa":   KeyUp,
		"Ob":   KeyDown,
		"Oc":   KeyRight,
		"Od":   KeyLeft,
		"Oe":   KeyClear,
		"[2^":  KeyInsert,
		"[3^":  KeyDelete,
		"[5^":  KeyPageUp,
		"[6^":  KeyPageDown,
		"[7^":  KeyHome,
		"[8^":  KeyEnd,
		"[Z":   KeyTab,
	}
	hookNonAlphanumericKeyNames = map[string]struct{}{
		KeyUp:        {},
		KeyDown:      {},
		KeyLeft:      {},
		KeyRight:     {},
		KeyPageDown:  {},
		KeyPageUp:    {},
		KeyHome:      {},
		KeyEnd:       {},
		KeyClear:     {},
		KeyInsert:    {},
		KeyF1:        {},
		KeyF2:        {},
		KeyF3:        {},
		KeyF4:        {},
		KeyF5:        {},
		KeyF6:        {},
		KeyF7:        {},
		KeyF8:        {},
		KeyF9:        {},
		KeyF10:       {},
		KeyF11:       {},
		KeyF12:       {},
		KeyTab:       {},
		KeyDelete:    {},
		KeyBackspace: {},
	}
	hookShiftKeyCodes = map[string]struct{}{
		"[a":  {},
		"[b":  {},
		"[c":  {},
		"[d":  {},
		"[e":  {},
		"[2$": {},
		"[3$": {},
		"[5$": {},
		"[6$": {},
		"[7$": {},
		"[8$": {},
		"[Z":  {},
	}
	hookCtrlKeyCodes = map[string]struct{}{
		"Oa":  {},
		"Ob":  {},
		"Oc":  {},
		"Od":  {},
		"Oe":  {},
		"[2^": {},
		"[3^": {},
		"[5^": {},
		"[6^": {},
		"[7^": {},
		"[8^": {},
	}
)

type parsedHookKeypress struct {
	Name     string
	Ctrl     bool
	Meta     bool
	Shift    bool
	Option   bool
	Sequence string
}

func normalizeHighBitMeta(raw string) string {
	if len(raw) != 1 {
		return raw
	}

	if raw[0] <= 127 {
		return raw
	}

	return "\x1b" + string([]byte{raw[0] - 128})
}

// Control key names (Ctrl+A through Ctrl+Z are 0x01 through 0x1a)
func ctrlKeyName(ch byte) string {
	return fmt.Sprintf("ctrl-%c", ch+0x60) // 0x01 + 0x60 = 'a'
}

// IsEscapeSequence checks if the input starts with an escape sequence
func IsEscapeSequence(input string) bool {
	if len(input) == 0 || input[0] != '\x1b' {
		return false
	}
	if len(input) == 1 {
		return true
	}

	parsed := parseHookKeypress(input)
	if parsed.Name != "" {
		return true
	}

	text := strings.TrimPrefix(parsed.Sequence, "\x1b")
	if text == "" || text[0] == '\x1b' || text[0] == '[' || text[0] == 'O' {
		return false
	}

	_, size := utf8.DecodeRuneInString(text)
	return size == len(text)
}

func keyFromParsedHookKeypress(parsed parsedHookKeypress) (Key, bool) {
	switch parsed.Name {
	case KeyEscape:
		return Key{Name: KeyEscape}, true
	case KeyF1:
		return Key{Name: KeyF1}, true
	case KeyF2:
		return Key{Name: KeyF2}, true
	case KeyF3:
		return Key{Name: KeyF3}, true
	case KeyF4:
		return Key{Name: KeyF4}, true
	case KeyF5:
		return Key{Name: KeyF5}, true
	case KeyF6:
		return Key{Name: KeyF6}, true
	case KeyF7:
		return Key{Name: KeyF7}, true
	case KeyF8:
		return Key{Name: KeyF8}, true
	case KeyF9:
		return Key{Name: KeyF9}, true
	case KeyF10:
		return Key{Name: KeyF10}, true
	case KeyF11:
		return Key{Name: KeyF11}, true
	case KeyF12:
		return Key{Name: KeyF12}, true
	case KeyUp:
		return Key{Name: KeyUp}, true
	case KeyDown:
		return Key{Name: KeyDown}, true
	case KeyLeft:
		return Key{Name: KeyLeft}, true
	case KeyRight:
		return Key{Name: KeyRight}, true
	case KeyDelete:
		return Key{Name: KeyDelete}, true
	case KeyInsert:
		return Key{Name: KeyInsert}, true
	case KeyClear:
		return Key{Name: KeyClear}, true
	case KeyHome:
		return Key{Name: KeyHome}, true
	case KeyEnd:
		return Key{Name: KeyEnd}, true
	case KeyPageUp:
		return Key{Name: KeyPageUp}, true
	case KeyPageDown:
		return Key{Name: KeyPageDown}, true
	case KeyReturn:
		return Key{Name: KeyReturn}, true
	case KeyBackspace:
		return Key{Name: KeyBackspace}, true
	case KeyTab:
		if parsed.Shift {
			return Key{Name: KeyShiftTab}, true
		}
		return Key{Name: KeyTab}, true
	case "space":
		return Key{Char: ' '}, true
	}

	if strings.HasPrefix(parsed.Sequence, "\x1b") {
		text := strings.TrimPrefix(parsed.Sequence, "\x1b")
		if text != "" && text[0] != '\x1b' && text[0] != '[' && text[0] != 'O' {
			r, size := utf8.DecodeRuneInString(text)
			if r != utf8.RuneError && size == len(text) {
				return Key{Char: r}, true
			}
		}
	}

	return Key{}, false
}

// ParseEscapeSequence parses ANSI escape sequences
func ParseEscapeSequence(input string) Key {
	if len(input) == 0 || input[0] != '\x1b' {
		return Key{}
	}

	parsed := parseHookKeypress(input)
	if key, ok := keyFromParsedHookKeypress(parsed); ok {
		return key
	}

	// Return the escape key as fallback for unknown sequences
	return Key{Name: KeyEscape}
}

// readKeyWithReader reads a single key using an existing buffered reader
func readEscapedKeyWithPrefix(reader *bufio.Reader, prefix []byte, sequenceStart int) Key {
	buf := append([]byte(nil), prefix...)

	// Read until we hit the end character.
	for i := 0; i < 8; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		buf = append(buf, b)
		// CSI sequences can contain a second '[' (e.g. ESC [[5~) before the
		// real terminator. Otherwise ANSI sequences end with 0x40-0x7E (@-~).
		if b >= 64 && b <= 126 && !(buf[sequenceStart] == '[' && len(buf) == sequenceStart+2 && b == '[') {
			break
		}
	}

	return ParseEscapeSequence(string(buf))
}

func readKeyWithReader(reader *bufio.Reader) (Key, error) {
	if preview, _ := reader.Peek(utf8.UTFMax); len(preview) > 0 && preview[0] > 127 {
		if utf8.FullRune(preview) {
			r, size := utf8.DecodeRune(preview)
			if r != utf8.RuneError && size > 1 {
				_, _, err := reader.ReadRune()
				if err != nil {
					return Key{}, err
				}

				return Key{Char: r}, nil
			}
		}

		metaByte, readErr := reader.ReadByte()
		if readErr != nil {
			return Key{}, readErr
		}

		return ParseEscapeSequence(normalizeHighBitMeta(string([]byte{metaByte}))), nil
	}

	// Read first rune
	r, _, err := reader.ReadRune()
	if err != nil {
		return Key{}, err
	}

	// Handle escape sequences
	if r == '\x1b' {
		// Try to read more to determine if it's a sequence or just escape key.
		nextBytes, err := reader.Peek(2)
		if err == nil || len(nextBytes) > 0 {
			if len(nextBytes) > 0 {
				switch nextBytes[0] {
				case '[', 'O':
					prefixByte, _ := reader.ReadByte()
					return readEscapedKeyWithPrefix(reader, []byte{'\x1b', prefixByte}, 1), nil
				case '\x1b':
					if len(nextBytes) > 1 && (nextBytes[1] == '[' || nextBytes[1] == 'O') {
						reader.ReadByte()
						prefixByte, _ := reader.ReadByte()
						return readEscapedKeyWithPrefix(reader, []byte{'\x1b', '\x1b', prefixByte}, 2), nil
					}

					// Treat a second escape byte as part of the same keypress so
					// persistent readers don't split meta+escape into two events.
					reader.ReadByte()
					return ParseEscapeSequence("\x1b\x1b"), nil
				default:
					metaRune, _, err := reader.ReadRune()
					if err != nil {
						return Key{Name: KeyEscape}, nil
					}

					return ParseEscapeSequence("\x1b" + string(metaRune)), nil
				}
			}
		}

		// Just the escape key
		return Key{Name: KeyEscape}, nil
	}

	// Handle special keys BEFORE control keys
	switch r {
	case '\r', '\n':
		return Key{Name: KeyReturn}, nil
	case '\x08':
		return Key{Name: KeyBackspace}, nil
	case '\x7f':
		return Key{Name: KeyDelete}, nil
	case '\t':
		return Key{Name: KeyTab}, nil
	}

	// Handle control keys (0x01-0x1a), excluding special chars already handled
	if (r >= 0x01 && r <= 0x08) || (r == 0x0b) || (r >= 0x0c && r <= 0x1a) {
		return Key{Name: ctrlKeyName(byte(r))}, nil
	}

	// Regular character key
	return Key{Char: r}, nil
}

// ReadKey reads a single key from the reader
func ReadKey(r io.Reader) (Key, error) {
	reader := bufio.NewReader(r)
	return readKeyWithReader(reader)
}

// InputHandler manages terminal input with a persistent buffer
type InputHandler struct {
	bufferedReader *bufio.Reader
}

// NewInputHandler creates a new input handler
func NewInputHandler(r io.Reader) *InputHandler {
	return &InputHandler{
		bufferedReader: bufio.NewReader(r),
	}
}

// ReadKey reads a single key from the input handler
func (h *InputHandler) ReadKey() (Key, error) {
	return readKeyWithReader(h.bufferedReader)
}

// KeyHandler is a function that handles key press events
// Returns true if the app should exit
type KeyHandler func(Key, []string) bool

// StdinHandler handles stdin input with a callback
type StdinHandler struct {
	handler KeyHandler
	keys    []string // For tracking pressed keys
}

// NewStdinHandler creates a new stdin handler
func NewStdinHandler(handler KeyHandler) *StdinHandler {
	return &StdinHandler{
		handler: handler,
		keys:    make([]string, 0),
	}
}

// HandleInput processes input from stdin
func (h *StdinHandler) HandleInput(inputStr string) error {
	_, _, keys, err := NormalizeHookInput(inputStr)
	if err != nil {
		return err
	}

	reader := strings.NewReader(inputStr)
	key, err := ReadKey(reader)
	if err != nil {
		return err
	}

	var deliveredKeys []string
	if len(keys) > 0 {
		h.keys = h.keys[:0]
		h.keys = append(h.keys, keys...)
		deliveredKeys = append([]string(nil), h.keys...)
	} else {
		h.keys = h.keys[:0]
		deliveredKeys = nil
	}

	shouldExit := h.handler(key, deliveredKeys)
	if shouldExit {
		return fmt.Errorf("exit requested")
	}

	return nil
}

// DecodeUTF8Rune decodes a UTF-8 rune from bytes
func DecodeUTF8Rune(data []byte) (rune, int) {
	return utf8.DecodeRune(data)
}

func isHookShiftKeyCode(code string) bool {
	_, ok := hookShiftKeyCodes[code]
	return ok
}

func isHookCtrlKeyCode(code string) bool {
	_, ok := hookCtrlKeyCodes[code]
	return ok
}

func parseHookKeypress(raw string) parsedHookKeypress {
	raw = normalizeHighBitMeta(raw)

	parsed := parsedHookKeypress{
		Sequence: raw,
	}

	switch {
	case raw == "\r" || raw == "\x1b\r":
		parsed.Name = KeyReturn
		parsed.Option = len(raw) == 2
		return parsed
	case raw == "\n":
		parsed.Name = "enter"
		return parsed
	case raw == "\t":
		parsed.Name = KeyTab
		return parsed
	case raw == "\b" || raw == "\x1b\b":
		parsed.Name = KeyBackspace
		parsed.Meta = strings.HasPrefix(raw, "\x1b")
		return parsed
	case raw == "\x7f" || raw == "\x1b\x7f":
		parsed.Name = KeyDelete
		parsed.Meta = strings.HasPrefix(raw, "\x1b")
		return parsed
	case raw == "\x1b" || raw == "\x1b\x1b":
		parsed.Name = KeyEscape
		parsed.Meta = len(raw) == 2
		return parsed
	case raw == " " || raw == "\x1b ":
		parsed.Name = "space"
		parsed.Meta = len(raw) == 2
		return parsed
	case len(raw) == 1 && raw <= "\x1a":
		parsed.Name = string(rune(raw[0]) + 'a' - 1)
		parsed.Ctrl = true
		return parsed
	case len(raw) == 1 && raw >= "0" && raw <= "9":
		parsed.Name = "number"
		return parsed
	case len(raw) == 1 && raw >= "a" && raw <= "z":
		parsed.Name = raw
		return parsed
	case len(raw) == 1 && raw >= "A" && raw <= "Z":
		parsed.Name = strings.ToLower(raw)
		parsed.Shift = true
		return parsed
	}

	if matches := hookMetaKeyCodeRe.FindStringSubmatch(raw); matches != nil {
		parsed.Meta = true
		parsed.Shift = matches[1] >= "A" && matches[1] <= "Z"
		return parsed
	}

	if matches := hookFnKeyRe.FindStringSubmatch(raw); matches != nil {
		runes := []rune(raw)
		if len(runes) >= 2 && runes[0] == '\x1b' && runes[1] == '\x1b' {
			parsed.Option = true
		}

		codeParts := make([]string, 0, 4)
		for _, part := range []string{matches[1], matches[2], matches[4], matches[6]} {
			if part != "" {
				codeParts = append(codeParts, part)
			}
		}
		code := strings.Join(codeParts, "")

		modifierText := "1"
		if matches[3] != "" {
			modifierText = matches[3]
		} else if matches[5] != "" {
			modifierText = matches[5]
		}

		modifier, err := strconv.Atoi(modifierText)
		if err == nil {
			modifier--
			parsed.Ctrl = modifier&4 != 0
			parsed.Meta = modifier&10 != 0
			parsed.Shift = modifier&1 != 0
		}

		parsed.Name = hookKeyNames[code]
		if isHookShiftKeyCode(code) {
			parsed.Shift = true
		}
		if isHookCtrlKeyCode(code) {
			parsed.Ctrl = true
		}

		return parsed
	}

	return parsed
}

func buildHookKey(parsed parsedHookKeypress, text string) HookKey {
	key := HookKey{
		UpArrow:    parsed.Name == KeyUp,
		DownArrow:  parsed.Name == KeyDown,
		LeftArrow:  parsed.Name == KeyLeft,
		RightArrow: parsed.Name == KeyRight,
		PageDown:   parsed.Name == KeyPageDown,
		PageUp:     parsed.Name == KeyPageUp,
		Home:       parsed.Name == KeyHome,
		End:        parsed.Name == KeyEnd,
		Return:     parsed.Name == KeyReturn,
		Escape:     parsed.Name == KeyEscape,
		Ctrl:       parsed.Ctrl,
		Shift:      parsed.Shift,
		Tab:        parsed.Name == KeyTab,
		Backspace:  parsed.Name == KeyBackspace,
		Delete:     parsed.Name == KeyDelete,
		Meta:       parsed.Meta || parsed.Option || parsed.Name == KeyEscape,
	}

	if len(text) == 1 && text >= "A" && text <= "Z" {
		key.Shift = true
	}

	return key
}

// NormalizeHookInput converts raw terminal data into a hook-friendly input string, key object, and key names.
func NormalizeHookInput(raw string) (string, HookKey, []string, error) {
	if raw == "" {
		return "", HookKey{}, nil, io.EOF
	}

	if strings.HasPrefix(raw, bracketedPasteStart) && strings.HasSuffix(raw, bracketedPasteEnd) {
		return raw[len(bracketedPasteStart) : len(raw)-len(bracketedPasteEnd)], HookKey{}, nil, nil
	}

	parsed := parseHookKeypress(raw)
	text := parsed.Sequence
	if parsed.Ctrl {
		text = parsed.Name
	}
	if _, ok := hookNonAlphanumericKeyNames[parsed.Name]; ok {
		text = ""
	}
	if strings.HasPrefix(text, "\x1b") {
		text = text[1:]
	}

	key := buildHookKey(parsed, text)
	var keys []string

	switch {
	case parsed.Name == KeyTab && key.Shift:
		keys = append(keys, KeyTab, "shift", KeyShiftTab)
	case parsed.Name == KeyEscape && key.Meta:
		if parsed.Meta || parsed.Option {
			keys = append(keys, "meta", parsed.Name)
		} else {
			keys = append(keys, parsed.Name)
		}
	case key.Ctrl && len(parsed.Name) == 1:
		keys = append(keys, "ctrl", "ctrl-"+parsed.Name)
	case key.Ctrl && parsed.Name != "":
		keys = append(keys, "ctrl", parsed.Name)
	case key.Meta && parsed.Name != "":
		keys = append(keys, "meta", parsed.Name)
	case parsed.Name == KeyReturn:
		keys = append(keys, parsed.Name)
	case parsed.Name != "":
		if _, ok := hookNonAlphanumericKeyNames[parsed.Name]; ok {
			keys = append(keys, parsed.Name)
		}
	case key.Meta && text != "":
		keys = append(keys, "meta", text)
	}

	return text, key, keys, nil
}
