package tuitest

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// NormalizeInput converts scenario input declarations into terminal bytes.
func NormalizeInput(input InputSpec) (string, error) {
	if strings.TrimSpace(input.Hex) != "" {
		if input.Key != "" || input.Text != "" {
			return "", fmt.Errorf("hex input cannot be combined with text or key")
		}
		decoded, err := hex.DecodeString(compactHex(input.Hex))
		if err != nil {
			return "", fmt.Errorf("decode hex input %q: %w", input.Hex, err)
		}
		return string(decoded), nil
	}

	if input.Key == "" {
		return input.Text, nil
	}

	switch strings.ToLower(strings.TrimSpace(input.Key)) {
	case "ctrl-c", "ctrl+c", "c-c":
		return "\x03", nil
	case "space":
		return " ", nil
	case "enter", "return":
		return "\r", nil
	case "escape", "esc":
		return "\x1b", nil
	case "tab":
		return "\t", nil
	case "shift-tab", "shift+tab", "backtab", "reverse-tab":
		return "\x1b[Z", nil
	case "backspace":
		return "\x7f", nil
	case "up", "arrow-up":
		return "\x1b[A", nil
	case "down", "arrow-down":
		return "\x1b[B", nil
	case "right", "arrow-right":
		return "\x1b[C", nil
	case "left", "arrow-left":
		return "\x1b[D", nil
	case "home":
		return "\x1b[H", nil
	case "end":
		return "\x1b[F", nil
	case "delete":
		return "\x1b[3~", nil
	default:
		return "", fmt.Errorf("unsupported key %q", input.Key)
	}
}

func compactHex(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		if r == ' ' || r == '\n' || r == '\t' || r == '_' || r == ':' {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

// ExitTimeout returns the parsed exit timeout, falling back to the shared default.
func ExitTimeout(exit ExitSpec) time.Duration {
	if strings.TrimSpace(exit.Within) == "" {
		return DefaultCtrlCExitTimeout
	}

	timeout, err := time.ParseDuration(exit.Within)
	if err != nil {
		return DefaultCtrlCExitTimeout
	}
	return timeout
}
