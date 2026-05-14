package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

var ansiSequencePattern = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\a]*(?:\a|\x1b\\)|[@-Z\\-_])`)

func SubprocessOutputDemo() *vdom.Node {
	app := ink.UseApp()
	outputRaw, setOutput := ink.UseState("")
	output := outputRaw.(string)

	ink.UseEffect(func() func() {
		cmd := newJestSubprocessCommand()
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			setOutput(err.Error())
			return nil
		}

		if err := cmd.Start(); err != nil {
			setOutput(err.Error())
			return nil
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			readSubprocessOutput(app, stdout, setOutput)
			_ = cmd.Wait()
			app.Schedule(func() {
				app.Exit()
			})
		}()

		return func() {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
		}
	}, []interface{}{})

	return ink.Box(vdom.Props{"flexDirection": "column", "padding": 1},
		ink.Text("Сommand output:"),
		ink.Box(vdom.Props{"marginTop": 1}, ink.Text(output)),
	)
}

func newJestSubprocessCommand() *exec.Cmd {
	if override := os.Getenv("GOINK_SUBPROCESS_JEST_CMD"); override != "" {
		cmd := exec.Command("bash", "-lc", override)
		cmd.Env = subprocessEnvironment()
		return cmd
	}

	if _, err := os.Stat(".tmp/tui/jest-demo"); err == nil {
		cmd := exec.Command(".tmp/tui/jest-demo")
		cmd.Env = subprocessEnvironment()
		return cmd
	}

	cmd := exec.Command("go", "run", "./examples/jest-demo")
	cmd.Env = subprocessEnvironment()
	return cmd
}

func subprocessEnvironment() []string {
	env := os.Environ()
	env = append(env, "GOINK_JEST_STREAM=1", "TERM=tmux-256color", "FORCE_COLOR=1")
	return env
}

func readSubprocessOutput(app ink.AppContext, reader io.Reader, setOutput ink.SetStateFunc) {
	buffer := make([]byte, 4096)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			output := lastFiveLines(stripANSI(string(buffer[:n])))
			app.Schedule(func() {
				setOutput(output)
			})
		}

		if err != nil {
			return
		}
	}
}

func stripANSI(value string) string {
	return ansiSequencePattern.ReplaceAllString(value, "")
}

func lastFiveLines(value string) string {
	lines := strings.Split(value, "\n")
	if len(lines) <= 5 {
		return strings.Join(lines, "\n")
	}

	return strings.Join(lines[len(lines)-5:], "\n")
}

func main() {
	if !terminal.StdoutIsTerminal() {
		app := ink.NewApp(SubprocessOutputDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(SubprocessOutputDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := instance.WaitUntilExit(); err != nil {
		fmt.Println(err)
	}
}
