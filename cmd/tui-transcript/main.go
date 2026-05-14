package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/dh-kam/goink.go/internal/tuitest"
	testrenderer "github.com/dh-kam/goink.go/pkg/renderer"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	scenarioPath := flag.String("scenario", "", "scenario YAML/JSON path")
	manifestPath := flag.String("manifest", "tests/tui/runtimes.yaml", "runtime manifest YAML/JSON path")
	runtimeName := flag.String("runtime", "", "runtime name from manifest")
	quiet := flag.Duration("quiet", 150*time.Millisecond, "frame settle quiet duration")
	stepTimeout := flag.Duration("step-timeout", 10*time.Second, "maximum wait for a step to settle")
	commandTimeout := flag.Duration("command-timeout", 10*time.Second, "maximum wait for a static command to exit")
	pretty := flag.Bool("pretty", false, "pretty-print JSON output")
	flag.Parse()

	if *scenarioPath == "" {
		return errors.New("-scenario is required")
	}
	if *runtimeName == "" {
		return errors.New("-runtime is required")
	}

	scenario, err := tuitest.LoadSpecFile(*scenarioPath)
	if err != nil {
		return err
	}
	manifest, err := tuitest.LoadRuntimeManifestFile(*manifestPath)
	if err != nil {
		return err
	}

	binding, ok := manifest.Binding(*runtimeName)
	if !ok {
		return fmt.Errorf("runtime %q is not defined in %s", *runtimeName, *manifestPath)
	}
	command, ok := binding.Apps[scenario.App]
	if !ok {
		return fmt.Errorf("runtime %q has no app binding for %q", *runtimeName, scenario.App)
	}

	transcript, err := runScenarioCommand(*scenarioPath, *manifestPath, *runtimeName, scenario, binding, command, *quiet, *stepTimeout, *commandTimeout)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	if *pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(transcript)
}

func runScenarioCommand(scenarioPath string, manifestPath string, runtimeName string, scenario *tuitest.Spec, runtime tuitest.RuntimeBinding, command tuitest.CommandSpec, quiet time.Duration, stepTimeout time.Duration, commandTimeout time.Duration) (*tuitest.Transcript, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, command.Command, command.Args...)
	cmd.Dir = resolveCwd(filepath.Dir(manifestPath), command.Cwd)
	cmd.Env = mergedEnvironment(runtime.Environment, command.Environment, scenario.Environment)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(scenario.Viewport.Width),
		Rows: uint16(scenario.Viewport.Height),
	})
	if err != nil {
		return nil, fmt.Errorf("start PTY command %q: %w", command.Command, err)
	}
	defer ptmx.Close()

	chunks := make(chan string, 128)
	readDone := make(chan error, 1)
	go readPTY(ptmx, chunks, readDone)

	exitDone := make(chan error, 1)
	go func() {
		exitDone <- cmd.Wait()
	}()

	transcript := &tuitest.Transcript{
		SchemaVersion: tuitest.TranscriptSchemaVersion,
		Scenario:      scenario.Name,
		Runtime:       runtimeName,
		App:           scenario.App,
		Viewport:      scenario.Viewport,
		Environment:   effectiveEnvironment(runtime.Environment, command.Environment, scenario.Environment),
	}
	screen := tuitest.NewTerminalScreen(scenario.Viewport.Width, scenario.Viewport.Height)

	for index, step := range scenario.Steps {
		input := ""
		if step.Resize != nil {
			if err := pty.Setsize(ptmx, &pty.Winsize{
				Cols: uint16(step.Resize.Width),
				Rows: uint16(step.Resize.Height),
			}); err != nil {
				return nil, fmt.Errorf("step %q resize PTY: %w", step.Name, err)
			}
			screen.Resize(step.Resize.Width, step.Resize.Height)
		}
		if step.Input != nil {
			normalized, err := tuitest.NormalizeInput(*step.Input)
			if err != nil {
				return nil, fmt.Errorf("step %q input: %w", step.Name, err)
			}
			input = normalized
			if _, err := io.WriteString(ptmx, normalized); err != nil {
				return nil, fmt.Errorf("step %q write input: %w", step.Name, err)
			}
		}

		if step.Wait != "" {
			waitDuration, err := time.ParseDuration(strings.TrimSpace(step.Wait))
			if err != nil {
				return nil, fmt.Errorf("step %q wait: %w", step.Name, err)
			}
			time.Sleep(waitDuration)
		}

		raw, exited, exitErr, screenApplied := collectStepOutputForStep(chunks, exitDone, scenario, step, quiet, stepTimeout, commandTimeout, screen)
		if !screenApplied {
			screen.Apply(raw)
		}
		frame := tuitest.TranscriptFrame{
			Index:       index,
			Step:        step.Name,
			Input:       escapedInput(input),
			Raw:         raw,
			RawEscaped:  escapeForJSON(raw),
			Plain:       normalizePlain(raw),
			ScreenPlain: screen.PlainString(),
		}
		transcript.Frames = append(transcript.Frames, frame)

		if exitErr != nil && step.Exit == nil {
			return transcript, fmt.Errorf("step %q failed: %w", step.Name, exitErr)
		}

		if step.Exit != nil {
			start := time.Now()
			if !exited {
				exitErr = waitExitWithTimeout(ptmx, exitDone, tuitest.ExitTimeout(*step.Exit))
			}
			exitErr = normalizeExitError(exitErr)
			transcript.Exit = &tuitest.TranscriptExit{
				Step:       step.Name,
				OK:         exitErr == nil,
				DurationMS: time.Since(start).Milliseconds(),
			}
			if exitErr != nil {
				transcript.Exit.Error = exitErr.Error()
				return transcript, fmt.Errorf("step %q exit failed: %w", step.Name, exitErr)
			}
			break
		}

		if exited {
			if exitErr != nil && !isSignalExit(exitErr) {
				return transcript, fmt.Errorf("command exited during step %q: %w", step.Name, exitErr)
			}
			break
		}
	}

	_ = terminateProcess(ptmx, cmd)
	_ = ptmx.Close()
	select {
	case <-readDone:
	case <-time.After(200 * time.Millisecond):
	}
	return transcript, nil
}

func collectStepOutputForStep(chunks <-chan string, exitDone <-chan error, scenario *tuitest.Spec, step tuitest.StepSpec, quiet time.Duration, stepTimeout time.Duration, commandTimeout time.Duration, screen *tuitest.TerminalScreen) (string, bool, error, bool) {
	if len(scenario.Steps) == 1 && step.Input == nil && step.Exit == nil {
		raw, exited, err := collectUntilExit(chunks, exitDone, quiet, commandTimeout)
		return raw, exited, err, false
	}
	if step.WaitFor != nil {
		timeout := stepTimeout
		if step.WaitFor.Within != "" {
			parsed, err := time.ParseDuration(strings.TrimSpace(step.WaitFor.Within))
			if err != nil {
				return "", false, fmt.Errorf("parse waitFor.within: %w", err), false
			}
			timeout = parsed
		}
		raw, exited, err := collectUntilScreenContains(chunks, exitDone, screen, step.WaitFor.Text, quiet, timeout)
		return raw, exited, err, true
	}
	if step.Resize != nil {
		raw, exited, err := collectStepOutputAllowNoOutput(chunks, exitDone, quiet, stepTimeout)
		return raw, exited, err, false
	}
	if step.Expect != nil && step.Expect.SameAsPrevious {
		raw, exited, err := collectStepOutputAllowNoOutput(chunks, exitDone, quiet, stepTimeout)
		return raw, exited, err, false
	}
	raw, exited, err := collectStepOutput(chunks, exitDone, quiet, stepTimeout)
	return raw, exited, err, false
}

func readPTY(reader io.Reader, chunks chan<- string, done chan<- error) {
	defer close(chunks)
	buffer := make([]byte, 4096)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			chunks <- string(buffer[:n])
		}
		if err != nil {
			done <- err
			return
		}
	}
}

func collectStepOutput(chunks <-chan string, exitDone <-chan error, quiet time.Duration, timeout time.Duration) (string, bool, error) {
	var builder strings.Builder
	timer := time.NewTimer(quiet)
	if !timer.Stop() {
		<-timer.C
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	exitCh := exitDone
	exited := false
	var exitErr error

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				return builder.String(), exited, exitErr
			}
			builder.WriteString(chunk)
			resetTimer(timer, quiet)
		case err := <-exitCh:
			exited = true
			exitErr = err
			exitCh = nil
			resetTimer(timer, quiet)
		case <-timer.C:
			return builder.String(), exited, exitErr
		case <-deadline.C:
			if exited {
				return builder.String(), true, exitErr
			}
			return builder.String(), false, fmt.Errorf("timed out waiting for PTY output to settle after %s", timeout)
		}
	}
}

func collectStepOutputAllowNoOutput(chunks <-chan string, exitDone <-chan error, quiet time.Duration, timeout time.Duration) (string, bool, error) {
	var builder strings.Builder
	timer := time.NewTimer(quiet)
	defer timer.Stop()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	exitCh := exitDone
	exited := false
	var exitErr error

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				return builder.String(), exited, exitErr
			}
			builder.WriteString(chunk)
			resetTimer(timer, quiet)
		case err := <-exitCh:
			exited = true
			exitErr = err
			exitCh = nil
			resetTimer(timer, quiet)
		case <-timer.C:
			return builder.String(), exited, exitErr
		case <-deadline.C:
			if exited {
				return builder.String(), true, exitErr
			}
			return builder.String(), false, fmt.Errorf("timed out waiting for PTY output to settle after %s", timeout)
		}
	}
}

func collectUntilScreenContains(chunks <-chan string, exitDone <-chan error, screen *tuitest.TerminalScreen, expected string, quiet time.Duration, timeout time.Duration) (string, bool, error) {
	var builder strings.Builder
	timer := time.NewTimer(quiet)
	if !timer.Stop() {
		<-timer.C
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	exitCh := exitDone
	exited := false
	var exitErr error
	found := false

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				return builder.String(), exited, exitErr
			}
			builder.WriteString(chunk)
			screen.Apply(chunk)
			if strings.Contains(screen.PlainString(), expected) {
				found = true
				resetTimer(timer, quiet)
			}
		case err := <-exitCh:
			exited = true
			exitErr = err
			exitCh = nil
			if found {
				resetTimer(timer, quiet)
			}
		case <-timer.C:
			return builder.String(), exited, exitErr
		case <-deadline.C:
			if found {
				return builder.String(), exited, exitErr
			}
			return builder.String(), exited, fmt.Errorf("timed out waiting for output containing %q after %s", expected, timeout)
		}
	}
}

func collectUntilExit(chunks <-chan string, exitDone <-chan error, quiet time.Duration, timeout time.Duration) (string, bool, error) {
	var builder strings.Builder
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	exited := false
	var exitErr error
	exitCh := exitDone
	timer := time.NewTimer(quiet)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				return builder.String(), exited, exitErr
			}
			builder.WriteString(chunk)
			if exited {
				resetTimer(timer, quiet)
			}
		case err := <-exitCh:
			exited = true
			exitErr = err
			exitCh = nil
			resetTimer(timer, quiet)
		case <-timer.C:
			return builder.String(), true, exitErr
		case <-deadline.C:
			return builder.String(), false, fmt.Errorf("timed out waiting for command exit after %s", timeout)
		}
	}
}

func waitExitWithTimeout(ptmx io.Writer, exitDone <-chan error, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-exitDone:
		if isSignalExit(err) {
			return nil
		}
		return err
	default:
	}

	if _, err := io.WriteString(ptmx, "\x03"); err != nil {
		return err
	}

	select {
	case err := <-exitDone:
		if isSignalExit(err) {
			return nil
		}
		return err
	case <-timer.C:
		return fmt.Errorf("timed out waiting for Ctrl+C exit after %s. %s", timeout, tuitest.CtrlCExitTimeoutGuidance)
	}
}

func terminateProcess(ptmx io.Writer, cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil || cmd.ProcessState != nil {
		return nil
	}
	_, _ = io.WriteString(ptmx, "\x03")
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		return cmd.Process.Kill()
	}
}

func resetTimer(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(duration)
}

func resolveCwd(baseDir string, cwd string) string {
	if cwd == "" {
		return baseDir
	}
	return filepath.Clean(filepath.Join(baseDir, cwd))
}

func mergedEnvironment(environments ...tuitest.Environment) []string {
	values := map[string]string{}
	unset := map[string]bool{}
	for _, item := range os.Environ() {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			values[key] = value
		}
	}
	for _, environment := range environments {
		for key, value := range environment {
			if value == nil {
				delete(values, key)
				unset[key] = true
				continue
			}
			values[key] = *value
			delete(unset, key)
		}
	}

	result := make([]string, 0, len(values))
	for key, value := range values {
		if unset[key] {
			continue
		}
		result = append(result, key+"="+value)
	}
	return result
}

func effectiveEnvironment(environments ...tuitest.Environment) map[string]string {
	result := map[string]string{}
	for _, environment := range environments {
		for key, value := range environment {
			if value == nil {
				result[key] = "<unset>"
				continue
			}
			result[key] = *value
		}
	}
	return result
}

func normalizePlain(raw string) string {
	plain := testrenderer.StripANSI(raw)
	plain = strings.ReplaceAll(plain, "\r\n", "\n")
	plain = strings.ReplaceAll(plain, "\r", "")
	return strings.TrimSuffix(plain, "\n")
}

func escapeForJSON(value string) string {
	quoted := strconv.QuoteToASCII(value)
	return quoted[1 : len(quoted)-1]
}

func escapedInput(input string) string {
	if input == "" {
		return ""
	}
	return escapeForJSON(input)
}

func isSignalExit(err error) bool {
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	return ok && status.Signaled()
}

func normalizeExitError(err error) error {
	if isSignalExit(err) {
		return nil
	}
	return err
}
