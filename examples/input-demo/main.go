package main

import (
	"fmt"
	"unicode/utf16"

	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

type currentKey struct {
	Char string
	Name string
}

func InputDemo() *vdom.Node {
	app := ink.UseApp()
	keyCountValue, setKeyCount := ink.UseState(0)
	lastKeysValue, setLastKeys := ink.UseState([]string{})
	currentKeyValue, setCurrentKey := ink.UseState(currentKey{})

	keyCount := keyCountValue.(int)
	lastKeys := lastKeysValue.([]string)
	pressedKey := currentKeyValue.(currentKey)

	ink.UseInput(func(input string, key ink.InputKey) {
		if input == "q" {
			app.Exit()
		}

		name := firstTrueKeyName(key)
		setKeyCount(keyCount + 1)
		setCurrentKey(currentKey{Char: input, Name: name})
		setLastKeys(appendRecentKey(lastKeys, firstNonEmpty(input, name)))
	})

	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Text("=== Input Demo ==="),
		ink.Text("Press keys to see them displayed"),
		ink.Text("Press 'q' or ESC to quit"),
		ink.Text("\n"),
		ink.Text(fmt.Sprintf("Key #%d: %s", keyCount, firstNonEmpty(pressedKey.Char, pressedKey.Name))),
		ink.Text(fmt.Sprintf("Char: %s (0x%s)", pressedKey.Char, charCodeHex(pressedKey.Char))),
		ink.Text(fmt.Sprintf("Name: %s", pressedKey.Name)),
		ink.Text(fmt.Sprintf("Recent keys: %s", joinKeys(lastKeys))),
	)
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(InputDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.Mount(InputDemo)
	if err != nil {
		panic(err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		panic(err)
	}
}

func firstTrueKeyName(key ink.InputKey) string {
	switch {
	case key.UpArrow:
		return "upArrow"
	case key.DownArrow:
		return "downArrow"
	case key.LeftArrow:
		return "leftArrow"
	case key.RightArrow:
		return "rightArrow"
	case key.PageDown:
		return "pageDown"
	case key.PageUp:
		return "pageUp"
	case key.Home:
		return "home"
	case key.End:
		return "end"
	case key.Return:
		return "return"
	case key.Escape:
		return "escape"
	case key.Ctrl:
		return "ctrl"
	case key.Shift:
		return "shift"
	case key.Tab:
		return "tab"
	case key.Backspace:
		return "backspace"
	case key.Delete:
		return "delete"
	case key.Meta:
		return "meta"
	default:
		return ""
	}
}

func appendRecentKey(keys []string, value string) []string {
	next := append(append([]string(nil), keys...), value)
	if len(next) > 5 {
		next = next[len(next)-5:]
	}

	return next
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func charCodeHex(value string) string {
	if value == "" {
		return "0"
	}

	codeUnits := utf16.Encode([]rune(value))
	if len(codeUnits) == 0 {
		return "0"
	}

	return fmt.Sprintf("%x", codeUnits[0])
}

func joinKeys(keys []string) string {
	if len(keys) == 0 {
		return ""
	}

	output := keys[0]
	for _, key := range keys[1:] {
		output += ", " + key
	}

	return output
}
