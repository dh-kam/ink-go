package tui_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dh-kam/goink.go/internal/tuitest"
)

func TestStaticScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscript := runTranscript(t, repoRoot, "examples/box-backgrounds-demo/testdata/box-backgrounds.scenario.yaml", "upstream-ink-node")
	goTranscript := runTranscript(t, repoRoot, "examples/box-backgrounds-demo/testdata/box-backgrounds.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscript, goTranscript, "screen")
}

func TestInteractiveCtrlCExitSmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "go")

	transcriptPath := runTranscript(t, repoRoot, "examples/aria-demo/testdata/aria-toggle.scenario.yaml", "goink-go")
	transcript := readTranscript(t, transcriptPath)
	if transcript.Exit == nil || !transcript.Exit.OK {
		t.Fatalf("expected successful Ctrl+C exit, got %+v", transcript.Exit)
	}
	if got, want := len(transcript.Frames), 5; got != want {
		t.Fatalf("expected %d frames, got %d", want, got)
	}
	if transcript.Frames[0].ScreenPlain == "" {
		t.Fatalf("expected screen projection in first frame")
	}
}

func TestInteractiveScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/aria-demo/testdata/aria-toggle.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/aria-demo/testdata/aria-toggle.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")
}

func TestChatMessagesScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/chat-demo/testdata/chat-messages.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/chat-demo/testdata/chat-messages.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "initial", "examples/chat-demo/testdata/chat-initial.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "type-hello-world", "examples/chat-demo/testdata/chat-type-hello-world.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "submit-hello-world", "examples/chat-demo/testdata/chat-submit-hello-world.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "type-thanks", "examples/chat-demo/testdata/chat-type-thanks.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "submit-thanks", "examples/chat-demo/testdata/chat-submit-thanks.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "type-exit", "examples/chat-demo/testdata/chat-type-exit.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "submit-exit", "examples/chat-demo/testdata/chat-submit-exit.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "ctrl-c-exit", "examples/chat-demo/testdata/chat-submit-exit.screen.golden")
	}
}

func TestCursorIMEKoreanEditingScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/cursor-ime-demo/testdata/cursor-ime-korean-editing.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/cursor-ime-demo/testdata/cursor-ime-korean-editing.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "initial", "examples/cursor-ime-demo/testdata/cursor-ime-initial.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "type-ascii-prefix", "examples/cursor-ime-demo/testdata/cursor-ime-ascii-prefix.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "split-ga-byte-1", "examples/cursor-ime-demo/testdata/cursor-ime-ascii-prefix.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "split-ga-byte-rest", "examples/cursor-ime-demo/testdata/cursor-ime-with-ga.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "type-rest-korean", "examples/cursor-ime-demo/testdata/cursor-ime-full-korean.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "backspace-01", "examples/cursor-ime-demo/testdata/cursor-ime-backspace-01.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "backspace-02", "examples/cursor-ime-demo/testdata/cursor-ime-backspace-02.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "backspace-03", "examples/cursor-ime-demo/testdata/cursor-ime-backspace-03.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "backspace-04", "examples/cursor-ime-demo/testdata/cursor-ime-backspace-04.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "backspace-05", "examples/cursor-ime-demo/testdata/cursor-ime-backspace-05.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "type-abc", "examples/cursor-ime-demo/testdata/cursor-ime-after-abc.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "split-da-byte-1", "examples/cursor-ime-demo/testdata/cursor-ime-after-abc.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "split-da-byte-rest", "examples/cursor-ime-demo/testdata/cursor-ime-after-split-da.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "type-final-korean", "examples/cursor-ime-demo/testdata/cursor-ime-final.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "ctrl-c-exit", "examples/cursor-ime-demo/testdata/cursor-ime-final.screen.golden")
	}
}

func TestUseInputMaxThenQuitPlainParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/use-input-demo/testdata/use-input-max-then-quit.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/use-input-demo/testdata/use-input-max-then-quit.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "plain")
}

func TestSelectInputScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/select-input-demo/testdata/select-input.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/select-input-demo/testdata/select-input.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")
}

func TestSelectInputWrapDownScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/select-input-demo/testdata/select-input-wrap.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/select-input-demo/testdata/select-input-wrap.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	transcript := readTranscript(t, nodeTranscriptPath)
	assertScreenGolden(t, repoRoot, transcript, "down-08", "examples/select-input-demo/testdata/select-input-down-08.screen.golden")
	assertScreenGolden(t, repoRoot, transcript, "down-09", "examples/select-input-demo/testdata/select-input-down-09.screen.golden")
}

func TestUseFocusNavigationScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/use-focus-demo/testdata/use-focus-navigation.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/use-focus-demo/testdata/use-focus-navigation.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "initial", "examples/use-focus-demo/testdata/use-focus-none.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "tab-01", "examples/use-focus-demo/testdata/use-focus-first.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "tab-02", "examples/use-focus-demo/testdata/use-focus-second.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "shift-tab-02", "examples/use-focus-demo/testdata/use-focus-third.screen.golden")
	}
}

func TestStaticCompleteScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/static-demo/testdata/static-complete.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/static-demo/testdata/static-complete.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "complete-10", "examples/static-demo/testdata/static-complete.screen.golden")
	}
}

func TestUseFocusWithIDNavigationScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/use-focus-with-id-demo/testdata/use-focus-with-id-navigation.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/use-focus-with-id-demo/testdata/use-focus-with-id-navigation.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "initial", "examples/use-focus-with-id-demo/testdata/use-focus-with-id-none.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "key-2", "examples/use-focus-with-id-demo/testdata/use-focus-with-id-second.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "tab-01", "examples/use-focus-with-id-demo/testdata/use-focus-with-id-third.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "key-1", "examples/use-focus-with-id-demo/testdata/use-focus-with-id-first.screen.golden")
	}
}

func TestUseStdoutTwoWritesScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/use-stdout-demo/testdata/use-stdout-two-writes.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/use-stdout-demo/testdata/use-stdout-two-writes.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "two-writes", "examples/use-stdout-demo/testdata/use-stdout-two-writes.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "ctrl-c-exit", "examples/use-stdout-demo/testdata/use-stdout-ctrl-c.screen.golden")
	}
}

func TestUseStderrTwoWritesScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/use-stderr-demo/testdata/use-stderr-two-writes.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/use-stderr-demo/testdata/use-stderr-two-writes.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "two-writes", "examples/use-stderr-demo/testdata/use-stderr-two-writes.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "ctrl-c-exit", "examples/use-stderr-demo/testdata/use-stderr-ctrl-c.screen.golden")
	}
}

func TestTableScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/table-demo/testdata/table.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/table-demo/testdata/table.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "render", "examples/table-demo/testdata/table.screen.golden")
	}
}

func TestTerminalResizeScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/terminal-resize-demo/testdata/terminal-resize.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/terminal-resize-demo/testdata/terminal-resize.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		assertScreenGolden(t, repoRoot, transcript, "initial-82", "examples/terminal-resize-demo/testdata/terminal-resize-82.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "resize-70", "examples/terminal-resize-demo/testdata/terminal-resize-70.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "type-goink", "examples/terminal-resize-demo/testdata/terminal-resize-goink.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "enter-clear", "examples/terminal-resize-demo/testdata/terminal-resize-70.screen.golden")
		assertScreenGolden(t, repoRoot, transcript, "ctrl-c-exit", "examples/terminal-resize-demo/testdata/terminal-resize-70.screen.golden")
	}
}

func TestTerminalResizeSweepScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/terminal-resize-demo/testdata/terminal-resize-sweep.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/terminal-resize-demo/testdata/terminal-resize-sweep.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		for width := 100; width >= 20; width -= 5 {
			assertScreenGolden(t, repoRoot, transcript,
				fmt.Sprintf("width-%d", width),
				fmt.Sprintf("examples/terminal-resize-demo/testdata/terminal-resize-sweep-width-%d.screen.golden", width),
			)
		}
		assertScreenGolden(t, repoRoot, transcript, "ctrl-c-exit", "examples/terminal-resize-demo/testdata/terminal-resize-sweep-width-20.screen.golden")
	}
}

func TestTerminalResizeHeightSweepScreenParitySmoke(t *testing.T) {
	repoRoot := repoRoot(t)
	requireCommand(t, "node")
	requireCommand(t, "go")
	requireUpstreamInk(t, repoRoot)

	nodeTranscriptPath := runTranscript(t, repoRoot, "examples/terminal-resize-demo/testdata/terminal-resize-height-sweep.scenario.yaml", "upstream-ink-node")
	goTranscriptPath := runTranscript(t, repoRoot, "examples/terminal-resize-demo/testdata/terminal-resize-height-sweep.scenario.yaml", "goink-go")
	runCompare(t, repoRoot, nodeTranscriptPath, goTranscriptPath, "screen")

	nodeTranscript := readTranscript(t, nodeTranscriptPath)
	goTranscript := readTranscript(t, goTranscriptPath)
	for _, transcript := range []*tuitest.Transcript{nodeTranscript, goTranscript} {
		for height := 20; height >= 4; height-- {
			assertScreenGolden(t, repoRoot, transcript,
				fmt.Sprintf("height-%d", height),
				fmt.Sprintf("examples/terminal-resize-demo/testdata/terminal-resize-sweep-height-%d.screen.golden", height),
			)
		}
		assertScreenGolden(t, repoRoot, transcript, "ctrl-c-exit", "examples/terminal-resize-demo/testdata/terminal-resize-sweep-height-4.screen.golden")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Clean("../..")
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	return root
}

func requireCommand(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s is required for TUI parity smoke tests: %v", name, err)
	}
}

func requireUpstreamInk(t *testing.T, repoRoot string) {
	t.Helper()
	upstream := filepath.Join(repoRoot, "..", "ink")
	if _, err := os.Stat(filepath.Join(upstream, "package.json")); err != nil {
		t.Skipf("upstream Ink checkout is required for Node parity smoke tests: %v", err)
	}
}

func runTranscript(t *testing.T, repoRoot string, scenario string, runtime string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), runtime+".transcript.json")
	stdout := runGoCommand(t, repoRoot, 45*time.Second,
		"run", "./cmd/tui-transcript",
		"-scenario", scenario,
		"-manifest", "tests/tui/runtimes.yaml",
		"-runtime", runtime,
		"-command-timeout", "30s",
	)
	if err := os.WriteFile(path, stdout, 0o644); err != nil {
		t.Fatalf("write transcript %q: %v", path, err)
	}
	return path
}

func runCompare(t *testing.T, repoRoot string, left string, right string, mode string) {
	t.Helper()
	_ = runGoCommand(t, repoRoot, 30*time.Second,
		"run", "./cmd/tui-compare",
		"-left", left,
		"-right", right,
		"-mode", mode,
	)
}

func runGoCommand(t *testing.T, repoRoot string, timeout time.Duration, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("go %v timed out after %s\n%s", args, timeout, output)
	}
	if err != nil {
		t.Fatalf("go %v failed: %v\n%s", args, err, output)
	}
	return output
}

func readTranscript(t *testing.T, path string) *tuitest.Transcript {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read transcript %q: %v", path, err)
	}

	var transcript tuitest.Transcript
	if err := json.Unmarshal(data, &transcript); err != nil {
		t.Fatalf("decode transcript %q: %v", path, err)
	}
	return &transcript
}

func assertScreenGolden(t *testing.T, repoRoot string, transcript *tuitest.Transcript, step string, goldenPath string) {
	t.Helper()
	expected, err := os.ReadFile(filepath.Join(repoRoot, goldenPath))
	if err != nil {
		t.Fatalf("read screen golden %q: %v", goldenPath, err)
	}
	expectedText := string(expected)
	expectedText = strings.TrimSuffix(expectedText, "\n")

	for _, frame := range transcript.Frames {
		if frame.Step != step {
			continue
		}
		if frame.ScreenPlain != expectedText {
			t.Fatalf("screen golden mismatch for step %q\nexpected:\n%s\n\nactual:\n%s", step, expectedText, frame.ScreenPlain)
		}
		return
	}
	t.Fatalf("step %q not found in transcript %q", step, transcript.Scenario)
}
