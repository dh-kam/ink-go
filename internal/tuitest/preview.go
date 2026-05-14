package tuitest

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

var tmuxPreviewMu sync.Mutex

type tmuxPreview struct {
	scenarioName string
	target       string
	tty          *os.File
	delay        time.Duration
	locked       bool
}

func newTmuxPreview(t testing.TB, scenarioName string) *tmuxPreview {
	t.Helper()

	target := strings.TrimSpace(os.Getenv("GOINK_TEST_TMUX_TARGET"))
	if target == "" {
		target = strings.TrimSpace(os.Getenv("GOINK_TMUX_TARGET"))
	}
	if target == "" {
		return nil
	}

	tmuxPreviewMu.Lock()
	preview := &tmuxPreview{scenarioName: scenarioName, target: target, delay: tmuxPreviewDelay(), locked: true}
	if err := preview.open(t); err != nil {
		t.Logf("tmux preview disabled for target %q: %v", target, err)
		preview.unlock()
		return nil
	}

	return preview
}

func (preview *tmuxPreview) open(t testing.TB) error {
	t.Helper()

	output, err := exec.Command("tmux", "display-message", "-p", "-t", preview.target, "#{pane_tty}").Output()
	if err != nil {
		return err
	}

	ttyPath := strings.TrimSpace(string(output))
	tty, err := os.OpenFile(ttyPath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprint(tty, "\x1b[?25l\x1b[2J\x1b[H"); err != nil {
		_ = tty.Close()
		return err
	}

	preview.tty = tty
	return nil
}

func (preview *tmuxPreview) Frame(t testing.TB, index int, label string, screen string) {
	t.Helper()
	if preview == nil || preview.tty == nil {
		return
	}

	header := fmt.Sprintf("%s | frame %d | %s", preview.scenarioName, index, label)
	if _, err := fmt.Fprintf(preview.tty, "\x1b[2J\x1b[H%s\n\n%s", tmuxPreviewBanner(header), screen); err != nil {
		t.Logf("tmux preview write failed: %v", err)
		return
	}

	time.Sleep(preview.delay)
}

func (preview *tmuxPreview) Close(t testing.TB) {
	t.Helper()
	if preview == nil {
		return
	}
	defer preview.unlock()
	if preview.tty == nil {
		return
	}

	doneBanner := tmuxPreviewBanner(fmt.Sprintf("goink test preview complete | %s | prompt restored below", preview.scenarioName))
	if _, err := fmt.Fprintf(preview.tty, "\x1b[?25h\r\n\r\n%s\r\n", doneBanner); err != nil {
		t.Logf("tmux preview restore failed: %v", err)
	}
	if err := preview.tty.Close(); err != nil {
		t.Logf("tmux preview close failed: %v", err)
	}
	preview.tty = nil

	if err := exec.Command("tmux", "send-keys", "-t", preview.target, "C-m").Run(); err != nil {
		t.Logf("tmux preview prompt restore failed for target %q: %v", preview.target, err)
	}
}

func (preview *tmuxPreview) unlock() {
	if preview.locked {
		preview.locked = false
		tmuxPreviewMu.Unlock()
	}
}

func tmuxPreviewBanner(text string) string {
	return "\x1b[48;5;24m\x1b[38;5;231m\x1b[1m " + text + " \x1b[K\x1b[0m"
}

func tmuxPreviewDelay() time.Duration {
	raw := strings.TrimSpace(os.Getenv("GOINK_TEST_TMUX_DELAY"))
	if raw == "" {
		return 350 * time.Millisecond
	}

	delay, err := time.ParseDuration(raw)
	if err != nil {
		return 350 * time.Millisecond
	}
	return delay
}
