package ink

import (
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func TestHandleInputProvidesClearModifierMatrix(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		expectedKeys string
		assertKey    func(t *testing.T, key InputKey)
	}{
		{
			name:         "clear",
			raw:          "\x1b[E",
			expectedKeys: "clear",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if key.Ctrl || key.Shift || key.Meta {
					t.Fatalf("expected plain clear with no modifiers, got %+v", key)
				}
			},
		},
		{
			name:         "shift-clear",
			raw:          "\x1b[e",
			expectedKeys: "clear",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Shift || key.Ctrl || key.Meta {
					t.Fatalf("expected shift+clear flags, got %+v", key)
				}
			},
		},
		{
			name:         "ctrl-clear",
			raw:          "\x1bOe",
			expectedKeys: "ctrl,clear",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Ctrl || key.Shift || key.Meta {
					t.Fatalf("expected ctrl+clear flags, got %+v", key)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			var modernInput string
			var modernKey InputKey
			var legacyInput interface{}
			var legacyKeys []string

			instance, err := MountWithOptions(func() *vdom.Node {
				UseInput(func(input string, key InputKey) {
					modernInput = input
					modernKey = key
				})
				UseInput(func(input interface{}, keys []string) bool {
					legacyInput = input
					legacyKeys = append([]string(nil), keys...)
					return false
				})

				return components.Text("Input")
			}, RenderOptions{
				AppOptions: AppOptions{Stdout: stdout},
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.HandleInput(testCase.raw); err != nil {
				t.Fatalf("handle input failed: %v", err)
			}

			if modernInput != "" {
				t.Fatalf("expected special key text input to be empty, got %q", modernInput)
			}
			if legacyInput != "" {
				t.Fatalf("expected legacy special key text input to be empty, got %#v", legacyInput)
			}
			if strings.Join(legacyKeys, ",") != testCase.expectedKeys {
				t.Fatalf("expected legacy keys %q, got %#v", testCase.expectedKeys, legacyKeys)
			}
			testCase.assertKey(t, modernKey)
		})
	}
}

func TestHandleInputPassesCtrlCToUseInputHooks(t *testing.T) {
	stdout := &recordingWriter{}
	var modernInput string
	var modernKey InputKey
	var legacyInput interface{}
	var legacyKeys []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			modernInput = input
			modernKey = key
		})
		UseInput(func(input interface{}, keys []string) bool {
			legacyInput = input
			legacyKeys = append([]string(nil), keys...)
			return false
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x03"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if modernInput != "c" {
		t.Fatalf("expected Ctrl+C input payload %q, got %q", "c", modernInput)
	}
	if !modernKey.Ctrl {
		t.Fatalf("expected Ctrl+C to set ctrl=true, got %+v", modernKey)
	}
	if legacyInput != "c" {
		t.Fatalf("expected legacy Ctrl+C input payload %q, got %#v", "c", legacyInput)
	}
	if strings.Join(legacyKeys, ",") != "ctrl,ctrl-c" {
		t.Fatalf("expected legacy Ctrl+C keys, got %#v", legacyKeys)
	}
}

func TestUseInputManyHooksKeepsASingleRuntimeInputSubscription(t *testing.T) {
	stdin := &inputRecordingStdin{}
	stdout := &recordingWriter{}

	instance, err := MountWithOptions(func() *vdom.Node {
		for index := 0; index < 12; index++ {
			UseInput(func(input string, key InputKey) {})
		}

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  stdin,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	if stdin.InputListenerCount() != 1 {
		t.Fatalf("expected a single stdin subscription for the mounted instance, got %d", stdin.InputListenerCount())
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount failed: %v", err)
	}
	if stdin.InputListenerCount() != 0 {
		t.Fatalf("expected stdin subscription cleanup on unmount, got %d listeners", stdin.InputListenerCount())
	}
}

func TestUseInputUnmountCleanupReleasesRawModeOnlyAfterLastHookLikeUpstreamFixture(t *testing.T) {
	stdout := &recordingWriter{}
	renderFirstInput := true
	renderSecondInput := true

	render := func() *vdom.Node {
		children := make([]*vdom.Node, 0, 3)

		if renderFirstInput {
			UseInput(func(input string, key InputKey) {})
			children = append(children, components.Text("First"))
		}

		if renderSecondInput {
			UseInput(func(input string, key InputKey) {})
			children = append(children, components.Text("Second"))
		}

		if len(children) == 0 {
			return nil
		}

		return components.Box(vdom.Props{"flexDirection": "column"}, children...)
	}

	instance, err := MountWithOptions(render, RenderOptions{
		AppOptions: AppOptions{
			Stdout: stdout,
			Stdin:  rawModeTestStdin{},
		},
		Debug: true,
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if instance.app.rawModeUsers != 2 {
		t.Fatalf("expected two raw mode users with two mounted input hooks, got %d", instance.app.rawModeUsers)
	}
	if instance.app.rawState == nil {
		t.Fatal("expected raw mode state to be enabled while input hooks are mounted")
	}

	renderSecondInput = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender removing second input failed: %v", err)
	}

	if instance.app.rawModeUsers != 1 {
		t.Fatalf("expected one raw mode user after removing one input hook, got %d", instance.app.rawModeUsers)
	}
	if instance.app.rawState == nil {
		t.Fatal("expected raw mode state to remain enabled while one input hook is still mounted")
	}

	renderFirstInput = false
	if err := instance.Rerender(render); err != nil {
		t.Fatalf("rerender removing all input hooks failed: %v", err)
	}

	if instance.app.rawModeUsers != 0 {
		t.Fatalf("expected raw mode users to reach zero after removing all input hooks, got %d", instance.app.rawModeUsers)
	}
	if instance.app.rawState != nil {
		t.Fatal("expected raw mode state to be restored after all input hooks unmount")
	}
}

func TestUseInputInactiveSiblingDoesNotDuplicateRuntimeHandling(t *testing.T) {
	stdout := &recordingWriter{}
	callCount := 0

	instance, err := MountWithOptions(func() *vdom.Node {
		handler := func(input string, key InputKey) {
			callCount++
		}

		UseInput(handler)
		UseInput(handler, InputOptions{IsActive: false})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("x"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected only the active hook to receive input, got %d calls", callCount)
	}
}

func TestHandleInputPreservesPastedPayloads(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantKey func(t *testing.T, key InputKey)
	}{
		{
			name: "pasted-carriage-return",
			raw:  "\rtest",
			want: "\rtest",
			wantKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if key.Return || key.Tab || key.Delete {
					t.Fatalf("expected pasted payload not to be coerced into a special key, got %+v", key)
				}
			},
		},
		{
			name: "pasted-tab",
			raw:  "\ttest",
			want: "\ttest",
			wantKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if key.Return || key.Tab || key.Delete {
					t.Fatalf("expected pasted payload not to be coerced into a special key, got %+v", key)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			var receivedInput string
			var receivedKey InputKey

			instance, err := MountWithOptions(func() *vdom.Node {
				UseInput(func(input string, key InputKey) {
					receivedInput = input
					receivedKey = key
				})

				return components.Text("Input")
			}, RenderOptions{
				AppOptions: AppOptions{Stdout: stdout},
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.HandleInput(testCase.raw); err != nil {
				t.Fatalf("handle input failed: %v", err)
			}

			if receivedInput != testCase.want {
				t.Fatalf("expected pasted payload %q, got %q", testCase.want, receivedInput)
			}
			testCase.wantKey(t, receivedKey)
		})
	}
}

func TestHandleInputPreservesPastedPayloadsForLegacyUseInput(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "pasted-carriage-return", raw: "\rtest", want: "\rtest"},
		{name: "pasted-tab", raw: "\ttest", want: "\ttest"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			var legacyInput interface{}
			var legacyKeys []string

			instance, err := MountWithOptions(func() *vdom.Node {
				UseInput(func(input interface{}, keys []string) bool {
					legacyInput = input
					legacyKeys = keys
					return false
				})

				return components.Text("Input")
			}, RenderOptions{
				AppOptions: AppOptions{Stdout: stdout},
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.HandleInput(testCase.raw); err != nil {
				t.Fatalf("handle input failed: %v", err)
			}

			if legacyInput != testCase.want {
				t.Fatalf("expected pasted payload %q, got %#v", testCase.want, legacyInput)
			}
			if legacyKeys != nil {
				t.Fatalf("expected pasted payload not to synthesize legacy keys, got %#v", legacyKeys)
			}
		})
	}
}

func TestHandleInputMatchesRemainingUseInputFixtureMatrix(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantInput string
		assertKey func(t *testing.T, key InputKey)
	}{
		{
			name:      "uppercase-character",
			raw:       "Q",
			wantInput: "Q",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Shift || key.Return || key.Delete {
					t.Fatalf("expected uppercase input to set only shift, got %+v", key)
				}
			},
		},
		{
			name:      "backspace",
			raw:       "\b",
			wantInput: "",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Backspace || key.Delete || key.Return || key.Meta {
					t.Fatalf("expected backspace to set backspace-only flags, got %+v", key)
				}
			},
		},
		{
			name:      "carriage-return-is-not-uppercase",
			raw:       "\r",
			wantInput: "\r",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Return || key.Shift {
					t.Fatalf("expected carriage return to set return without shift, got %+v", key)
				}
			},
		},
		{
			name:      "escape",
			raw:       "\x1b",
			wantInput: "",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Escape || !key.Meta || key.Tab || key.Return {
					t.Fatalf("expected escape to set escape and meta flags, got %+v", key)
				}
			},
		},
		{
			name:      "delete",
			raw:       "\x7f",
			wantInput: "",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if key.Backspace || !key.Delete || key.Return || key.Meta {
					t.Fatalf("expected raw DEL byte to map to delete-only flags, got %+v", key)
				}
			},
		},
		{
			name:      "remove-delete-sequence",
			raw:       "\x1b[3~",
			wantInput: "",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Delete || key.Meta || key.Return {
					t.Fatalf("expected remove sequence to map to delete, got %+v", key)
				}
			},
		},
		{
			name:      "option-return",
			raw:       "\x1b\r",
			wantInput: "\r",
			assertKey: func(t *testing.T, key InputKey) {
				t.Helper()
				if !key.Meta || !key.Return || key.Shift {
					t.Fatalf("expected option+return flags, got %+v", key)
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			stdout := &recordingWriter{}
			var receivedInput string
			var receivedKey InputKey

			instance, err := MountWithOptions(func() *vdom.Node {
				UseInput(func(input string, key InputKey) {
					receivedInput = input
					receivedKey = key
				})

				return components.Text("Input")
			}, RenderOptions{
				AppOptions: AppOptions{Stdout: stdout},
			})
			if err != nil {
				t.Fatalf("mount failed: %v", err)
			}
			defer instance.Unmount()

			if err := instance.HandleInput(testCase.raw); err != nil {
				t.Fatalf("handle input failed: %v", err)
			}

			if receivedInput != testCase.wantInput {
				t.Fatalf("expected input %q, got %q", testCase.wantInput, receivedInput)
			}
			testCase.assertKey(t, receivedKey)
		})
	}
}

func TestHandleInputTreatsDoubleEscapeAsMetaEscape(t *testing.T) {
	stdout := &recordingWriter{}
	var modernInput string
	var modernKey InputKey
	var legacyInput interface{}
	var legacyKeys []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			modernInput = input
			modernKey = key
		})
		UseInput(func(input interface{}, keys []string) bool {
			legacyInput = input
			legacyKeys = append([]string(nil), keys...)
			return false
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b\x1b"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if modernInput != "\x1b" {
		t.Fatalf("expected modern input to preserve the escaped escape byte, got %q", modernInput)
	}
	if !modernKey.Escape || !modernKey.Meta || modernKey.Tab || modernKey.Return {
		t.Fatalf("expected meta+escape flags, got %+v", modernKey)
	}
	if legacyInput != "\x1b" {
		t.Fatalf("expected legacy input to preserve the escaped escape byte, got %#v", legacyInput)
	}
	if strings.Join(legacyKeys, ",") != "meta,escape" {
		t.Fatalf("expected legacy meta+escape keys, got %#v", legacyKeys)
	}
}

func TestHandleInputTreatsSingleHighBitByteAsMetaCharacter(t *testing.T) {
	stdout := &recordingWriter{}
	var modernInput string
	var modernKey InputKey
	var legacyInput interface{}
	var legacyKeys []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			modernInput = input
			modernKey = key
		})
		UseInput(func(input interface{}, keys []string) bool {
			legacyInput = input
			legacyKeys = append([]string(nil), keys...)
			return false
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput(string([]byte{0xe9})); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if modernInput != "i" {
		t.Fatalf("expected modern input to normalize 8-bit meta byte to %q, got %q", "i", modernInput)
	}
	if !modernKey.Meta || modernKey.Ctrl || modernKey.Return || modernKey.Delete {
		t.Fatalf("expected 8-bit meta byte to set meta-only flags, got %+v", modernKey)
	}
	if legacyInput != "i" {
		t.Fatalf("expected legacy input to normalize 8-bit meta byte to %q, got %#v", "i", legacyInput)
	}
	if strings.Join(legacyKeys, ",") != "meta,i" {
		t.Fatalf("expected legacy keys to match meta+i, got %#v", legacyKeys)
	}
}

func TestHandleInputTreatsEscapedDELByteAsMetaDelete(t *testing.T) {
	stdout := &recordingWriter{}
	var modernInput string
	var modernKey InputKey
	var legacyInput interface{}
	var legacyKeys []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			modernInput = input
			modernKey = key
		})
		UseInput(func(input interface{}, keys []string) bool {
			legacyInput = input
			legacyKeys = append([]string(nil), keys...)
			return false
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("\x1b\x7f"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if modernInput != "" {
		t.Fatalf("expected escaped DEL byte not to produce text input, got %q", modernInput)
	}
	if !modernKey.Meta || modernKey.Backspace || !modernKey.Delete || modernKey.Return {
		t.Fatalf("expected meta+delete flags, got %+v", modernKey)
	}
	if legacyInput != "" {
		t.Fatalf("expected legacy escaped DEL byte input to stay empty, got %#v", legacyInput)
	}
	if strings.Join(legacyKeys, ",") != "meta,delete" {
		t.Fatalf("expected legacy meta+delete keys, got %#v", legacyKeys)
	}
}

func TestHandleInputPreservesPlainUTF8Characters(t *testing.T) {
	stdout := &recordingWriter{}
	var modernInput string
	var modernKey InputKey
	var legacyInput interface{}
	var legacyKeys []string

	instance, err := MountWithOptions(func() *vdom.Node {
		UseInput(func(input string, key InputKey) {
			modernInput = input
			modernKey = key
		})
		UseInput(func(input interface{}, keys []string) bool {
			legacyInput = input
			legacyKeys = append([]string(nil), keys...)
			return false
		})

		return components.Text("Input")
	}, RenderOptions{
		AppOptions: AppOptions{Stdout: stdout},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	defer instance.Unmount()

	if err := instance.HandleInput("🙂"); err != nil {
		t.Fatalf("handle input failed: %v", err)
	}

	if modernInput != "🙂" {
		t.Fatalf("expected modern input to preserve UTF-8 rune, got %q", modernInput)
	}
	if modernKey.Meta || modernKey.Ctrl || modernKey.Return || modernKey.Delete || modernKey.Tab {
		t.Fatalf("expected plain UTF-8 input not to set synthetic modifiers, got %+v", modernKey)
	}
	if legacyInput != "🙂" {
		t.Fatalf("expected legacy input to preserve UTF-8 rune, got %#v", legacyInput)
	}
	if legacyKeys != nil {
		t.Fatalf("expected plain UTF-8 input not to synthesize legacy keys, got %#v", legacyKeys)
	}
}
