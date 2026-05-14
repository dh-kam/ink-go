package tuitest

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dh-kam/ink-go/pkg/ink"
	testrenderer "github.com/dh-kam/ink-go/pkg/renderer"
)

const (
	CapturePlain      = "plain"
	CaptureScreen     = "screen"
	CaptureANSI       = "ansi"
	CaptureANSIEscape = "ansi-escaped"

	DefaultCtrlCExitTimeout  = 5 * time.Second
	CtrlCExitTimeoutGuidance = "Loop-driven Ctrl+C exit checks are environment-sensitive; adjust this timeout if this host is consistently slow, but an exit that never completes is a real shutdown bug."
)

type Components map[string]ink.ComponentFunc
type Renderers map[string]func() string

type Apps struct {
	Components Components
	Renderers  Renderers
}

func RunScenarioFile(t testing.TB, path string, components Components) {
	t.Helper()

	RunScenarioFileWithApps(t, path, Apps{Components: components})
}

func RunStaticScenarioFile(t testing.TB, path string, renderers Renderers) {
	t.Helper()

	RunScenarioFileWithApps(t, path, Apps{Renderers: renderers})
}

func RunScenarioFileWithApps(t testing.TB, path string, apps Apps) {
	t.Helper()

	spec, err := LoadSpecFile(path)
	if err != nil {
		t.Fatalf("load TUI scenario: %v", err)
	}

	RunScenarioWithApps(t, spec, filepath.Dir(path), apps)
}

func RunScenario(t testing.TB, spec *Spec, baseDir string, components Components) {
	t.Helper()

	RunScenarioWithApps(t, spec, baseDir, Apps{Components: components})
}

func RunScenarioWithApps(t testing.TB, spec *Spec, baseDir string, apps Apps) {
	t.Helper()

	restoreEnvironment := applyScenarioEnvironment(t, spec.Environment)
	defer restoreEnvironment()

	if renderer, ok := apps.Renderers[spec.App]; ok {
		runStaticScenario(t, spec, baseDir, renderer)
		return
	}

	component, ok := apps.Components[spec.App]
	if !ok {
		t.Fatalf("TUI scenario %q references unregistered app %q", spec.Name, spec.App)
	}

	stdout := &Recorder{}
	preview := newTmuxPreview(t, spec.Name)
	defer preview.Close(t)

	instance, err := ink.MountWithOptions(component, ink.RenderOptions{
		AppOptions: ink.AppOptions{
			Stdout: stdout,
			Stderr: stdout,
			Width:  spec.Viewport.Width,
			Height: spec.Viewport.Height,
		},
	})
	if err != nil {
		t.Fatalf("mount TUI scenario %q: %v", spec.Name, err)
	}
	defer func() {
		if err := instance.Unmount(); err != nil {
			t.Fatalf("unmount TUI scenario %q: %v", spec.Name, err)
		}
	}()

	mode := captureMode(spec.Capture.Mode)
	trimTrailingNewline := captureTrimTrailingNewline(spec.Capture)
	var screen *TerminalScreen
	screenWriteIndex := 0
	if mode == CaptureScreen {
		screen = NewTerminalScreen(spec.Viewport.Width, spec.Viewport.Height)
	}
	var previous string
	for index, step := range spec.Steps {
		writesBeforeStep := stdout.WriteCount()
		if step.Resize != nil {
			t.Fatalf("TUI scenario %q step %q defines resize; use the PTY transcript runner for resize scenarios", spec.Name, step.Name)
		}
		if step.Input != nil {
			input, err := NormalizeInput(*step.Input)
			if err != nil {
				t.Fatalf("TUI scenario %q step %q input: %v", spec.Name, step.Name, err)
			}
			if err := instance.HandleInput(input); err != nil {
				t.Fatalf("TUI scenario %q step %q handle input: %v", spec.Name, step.Name, err)
			}
		}

		if step.Wait != "" {
			waitDuration, err := parseDuration(step.Wait)
			if err != nil {
				t.Fatalf("TUI scenario %q step %q wait: %v", spec.Name, step.Name, err)
			}
			time.Sleep(waitDuration)
		}

		if step.WaitFor != nil {
			waitForRecorder(t, spec.Name, step.Name, stdout, screen, &screenWriteIndex, mode, trimTrailingNewline, *step.WaitFor)
		}

		if step.Exit != nil {
			assertExitWithin(t, spec.Name, step.Name, instance, ExitTimeout(*step.Exit))
			if exitNoExtraWrites(*step.Exit) && stdout.WriteCount() != writesBeforeStep {
				t.Fatalf("TUI scenario %q step %q expected exit to leave the final frame intact, got %d extra writes",
					spec.Name, step.Name, stdout.WriteCount()-writesBeforeStep)
			}
		}

		actual, previewOutput := captureRuntimeStep(stdout, screen, &screenWriteIndex, mode, trimTrailingNewline)
		preview.Frame(t, index, step.Name, previewOutput)
		assertStepExpectation(t, spec, step, baseDir, mode, trimTrailingNewline, previous, actual)
		previous = actual
	}
}

func runStaticScenario(t testing.TB, spec *Spec, baseDir string, renderer func() string) {
	t.Helper()
	if len(spec.Steps) != 1 {
		t.Fatalf("static TUI scenario %q must have exactly one step, got %d", spec.Name, len(spec.Steps))
	}

	step := spec.Steps[0]
	if step.Input != nil || step.Exit != nil || step.Resize != nil {
		t.Fatalf("static TUI scenario %q step %q must not define input, exit, or resize", spec.Name, step.Name)
	}

	mode := captureMode(spec.Capture.Mode)
	trimTrailingNewline := captureTrimTrailingNewline(spec.Capture)
	raw := renderer()
	actual := captureStaticOutputWithViewport(raw, mode, trimTrailingNewline, spec.Viewport)
	preview := newTmuxPreview(t, spec.Name)
	defer preview.Close(t)
	preview.Frame(t, 0, step.Name, previewStaticOutputWithViewport(raw, mode, trimTrailingNewline, spec.Viewport))
	assertStepExpectation(t, spec, step, baseDir, mode, trimTrailingNewline, "", actual)
}

func captureMode(mode string) string {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case "", CapturePlain:
		return CapturePlain
	case CaptureScreen:
		return CaptureScreen
	case CaptureANSI:
		return CaptureANSI
	case CaptureANSIEscape:
		return CaptureANSIEscape
	default:
		return strings.TrimSpace(strings.ToLower(mode))
	}
}

func captureTrimTrailingNewline(capture CaptureSpec) bool {
	if capture.TrimTrailingNewline == nil {
		return true
	}
	return *capture.TrimTrailingNewline
}

func captureRuntimeStep(stdout *Recorder, screen *TerminalScreen, screenWriteIndex *int, mode string, trimTrailingNewline bool) (string, string) {
	if mode != CaptureScreen {
		actual := stdout.Capture(mode, trimTrailingNewline)
		return actual, stdout.Preview(mode, trimTrailingNewline)
	}

	for _, write := range stdout.WritesFrom(*screenWriteIndex) {
		screen.Apply(terminalScreenInput(write))
	}
	*screenWriteIndex = stdout.WriteCount()
	actual := screen.PlainString()
	return actual, actual
}

func waitForRecorder(t testing.TB, scenarioName string, stepName string, stdout *Recorder, screen *TerminalScreen, screenWriteIndex *int, mode string, trimTrailingNewline bool, waitFor WaitForSpec) {
	t.Helper()

	timeout := DefaultCtrlCExitTimeout
	if waitFor.Within != "" {
		parsed, err := parseDuration(waitFor.Within)
		if err != nil {
			t.Fatalf("TUI scenario %q step %q waitFor.within: %v", scenarioName, stepName, err)
		}
		timeout = parsed
	}

	deadline := time.Now().Add(timeout)
	for {
		actual, _ := captureRuntimeStep(stdout, screen, screenWriteIndex, mode, trimTrailingNewline)
		if strings.Contains(actual, waitFor.Text) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("TUI scenario %q step %q timed out after %s waiting for %q", scenarioName, stepName, timeout, waitFor.Text)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func captureStaticOutput(output string, mode string, trimTrailingNewline bool) string {
	return captureStaticOutputWithViewport(output, mode, trimTrailingNewline, Viewport{})
}

func captureStaticOutputWithViewport(output string, mode string, trimTrailingNewline bool, viewport Viewport) string {
	if mode == CaptureScreen {
		screen := NewTerminalScreen(viewport.Width, viewport.Height)
		screen.Apply(terminalScreenInput(output))
		return screen.PlainString()
	}

	var captured string
	switch mode {
	case CaptureANSI:
		captured = output
	case CaptureANSIEscape:
		captured = escapeForFixture(output)
	default:
		captured = stripANSI(output)
	}

	if trimTrailingNewline {
		captured = strings.TrimSuffix(captured, "\n")
	}
	return captured
}

func previewStaticOutput(output string, mode string, trimTrailingNewline bool) string {
	return previewStaticOutputWithViewport(output, mode, trimTrailingNewline, Viewport{})
}

func previewStaticOutputWithViewport(output string, mode string, trimTrailingNewline bool, viewport Viewport) string {
	if mode == CaptureScreen {
		return captureStaticOutputWithViewport(output, mode, trimTrailingNewline, viewport)
	}

	switch mode {
	case CaptureANSI, CaptureANSIEscape:
		return output
	default:
		return captureStaticOutput(output, CapturePlain, trimTrailingNewline)
	}
}

func terminalScreenInput(output string) string {
	if !strings.Contains(output, "\n") {
		return output
	}

	var builder strings.Builder
	builder.Grow(len(output) + strings.Count(output, "\n"))
	for index := 0; index < len(output); index++ {
		if output[index] == '\n' && (index == 0 || output[index-1] != '\r') {
			builder.WriteByte('\r')
		}
		builder.WriteByte(output[index])
	}
	return builder.String()
}

func exitNoExtraWrites(exit ExitSpec) bool {
	if exit.NoExtraWrites == nil {
		return true
	}
	return *exit.NoExtraWrites
}

func assertExitWithin(t testing.TB, scenarioName string, stepName string, instance *ink.Instance, timeout time.Duration) {
	t.Helper()

	done := make(chan error, 1)
	go func() {
		done <- instance.WaitUntilExit()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("TUI scenario %q step %q exit failed: %v", scenarioName, stepName, err)
		}
	case <-time.After(timeout):
		t.Fatalf("TUI scenario %q step %q timed out waiting for Ctrl+C exit after %s. %s",
			scenarioName, stepName, timeout, CtrlCExitTimeoutGuidance)
	}
}

func assertStepExpectation(t testing.TB, spec *Spec, step StepSpec, baseDir string, mode string, trimTrailingNewline bool, previous string, actual string) {
	t.Helper()
	if step.Expect == nil {
		return
	}

	expected, ok := expectedText(t, step, baseDir, trimTrailingNewline, previous)
	if !ok {
		return
	}
	if actual == expected {
		return
	}

	actualPath := writeActualFrame(t, spec.Name, step.Name, mode, actual)
	firstLine := firstDifferingLine(expected, actual)
	t.Fatalf("TUI scenario %q step %q mismatch in %s capture at line %d; actual frame written to %s\nexpected:\n%s\n\nactual:\n%s",
		spec.Name, step.Name, mode, firstLine, actualPath, expected, actual)
}

func expectedText(t testing.TB, step StepSpec, baseDir string, trimTrailingNewline bool, previous string) (string, bool) {
	t.Helper()

	expect := step.Expect
	if expect.SameAsPrevious {
		return previous, true
	}
	if expect.Text != nil {
		if trimTrailingNewline {
			return strings.TrimSuffix(*expect.Text, "\n"), true
		}
		return *expect.Text, true
	}
	if expect.File == "" {
		return "", false
	}

	path := filepath.Clean(filepath.Join(baseDir, expect.File))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read expected frame %q for step %q: %v", expect.File, step.Name, err)
	}
	expected := string(content)
	if trimTrailingNewline {
		expected = strings.TrimSuffix(expected, "\n")
	}
	return expected, true
}

func writeActualFrame(t testing.TB, scenarioName string, stepName string, mode string, actual string) string {
	t.Helper()

	filename := sanitizeFileName(scenarioName + "-" + stepName + "." + mode + ".actual")
	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, []byte(actual), 0o644); err != nil {
		t.Fatalf("write actual TUI frame %q: %v", path, err)
	}
	return path
}

func firstDifferingLine(expected string, actual string) int {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")
	lineCount := len(expectedLines)
	if len(actualLines) > lineCount {
		lineCount = len(actualLines)
	}
	for index := 0; index < lineCount; index++ {
		var expectedLine string
		if index < len(expectedLines) {
			expectedLine = expectedLines[index]
		}
		var actualLine string
		if index < len(actualLines) {
			actualLine = actualLines[index]
		}
		if expectedLine != actualLine {
			return index + 1
		}
	}
	return 0
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_", "..", "_")
	return strings.Trim(replacer.Replace(name), "_")
}

func escapeForFixture(value string) string {
	quoted := strconv.QuoteToASCII(value)
	return quoted[1 : len(quoted)-1]
}

func stripANSI(value string) string {
	return testrenderer.StripANSI(value)
}
