package components

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// StackFrame describes a single parsed entry from a Go runtime stack trace.
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// ErrorOverviewProps configures an ErrorOverview block.
//
//   - Err is the underlying error to display. nil is permitted and renders a
//     placeholder body.
//   - Stack is an optional captured runtime.Stack output. When empty, no stack
//     section is rendered.
//   - SourceContext is the number of lines of surrounding source to display
//     above and below the failing line. A non-positive value disables the
//     source excerpt entirely.
type ErrorOverviewProps struct {
	Err           error
	Stack         string
	SourceContext int
}

// ErrorOverviewGroupProps configures an ErrorOverviewGroup block.
//
//   - Title is the header text (defaults to "ERRORS"). The trailing count is
//     appended automatically when there is more than one entry.
//   - Validation lists user-facing input/validation errors, typically the
//     field errors emitted by a Form. Each entry is rendered as a row in
//     a "Validation" sub-section.
//   - Runtime lists internal/runtime errors (panics, I/O, downstream call
//     failures). Each entry is rendered as a row in a "Runtime" sub-section.
//   - Stack is an optional joint stack trace shared by the runtime errors.
//     When empty, no stack section is rendered.
//   - SourceContext mirrors ErrorOverviewProps.SourceContext: how many
//     lines of surrounding source to display above and below the failing
//     line in the originating frame. A non-positive value disables the
//     excerpt entirely.
type ErrorOverviewGroupProps struct {
	Title         string
	Validation    []error
	Runtime       []error
	Stack         string
	SourceContext int
}

// ErrorOverviewGroup renders a summary block for multiple validation and/or
// runtime errors, with an optional joint stack trace. It is the natural
// complement to Form's field-level error reporting and panic recovery flows
// where several errors need to surface at once. nil entries are skipped.
//
// When both Validation and Runtime are empty, a "<no errors>" placeholder is
// rendered so the component remains visually present in conditional layouts
// without callers needing to nil-check first.
func ErrorOverviewGroup(props ErrorOverviewGroupProps) *vdom.Node {
	validation := filterNonNilErrors(props.Validation)
	runtime := filterNonNilErrors(props.Runtime)
	totalCount := len(validation) + len(runtime)

	sectionChildren := []*vdom.Node{
		errorOverviewGroupHeader(props.Title, totalCount),
	}

	if totalCount == 0 {
		sectionChildren = append(sectionChildren, Box(vdom.Props{"marginTop": 1},
			Text(styles.Dim("<no errors>")),
		))
		return Box(vdom.Props{"flexDirection": "column", "padding": 1}, sectionChildren...)
	}

	if len(validation) > 0 {
		sectionChildren = append(sectionChildren, errorOverviewSubsection("Validation", validation))
	}
	if len(runtime) > 0 {
		sectionChildren = append(sectionChildren, errorOverviewSubsection("Runtime", runtime))
	}

	frames := ParseGoStack(props.Stack)
	if origin, ok := firstUserFrame(frames); ok {
		sectionChildren = append(sectionChildren, errorOverviewLocation(origin))
		if props.SourceContext > 0 {
			if excerpt := errorOverviewExcerpt(origin, props.SourceContext); excerpt != nil {
				sectionChildren = append(sectionChildren, excerpt)
			}
		}
	}

	if stackBlock := errorOverviewStack(frames); stackBlock != nil {
		sectionChildren = append(sectionChildren, stackBlock)
	}

	return Box(vdom.Props{"flexDirection": "column", "padding": 1}, sectionChildren...)
}

func errorOverviewGroupHeader(title string, count int) *vdom.Node {
	if title == "" {
		title = "ERRORS"
	}

	headerLabel := " " + title + " "
	header := styles.WrapWithANSI(headerLabel, styles.Red.ToANSI(styles.Background), styles.White.ToANSI(styles.Foreground), styles.BoldCode())

	suffix := ""
	if count > 1 {
		suffix = fmt.Sprintf(" (%d)", count)
	}

	return Box(vdom.Props{"flexDirection": "row"},
		Text(header),
		Text(suffix),
	)
}

func errorOverviewSubsection(label string, errs []error) *vdom.Node {
	rows := make([]*vdom.Node, 0, len(errs)+1)
	rows = append(rows, Text(styles.Dim(label+":")))
	for _, err := range errs {
		message := err.Error()
		if message == "" {
			message = "<empty error>"
		}
		rows = append(rows, Text("  - "+message))
	}
	return Box(vdom.Props{"marginTop": 1, "flexDirection": "column"}, rows...)
}

func filterNonNilErrors(in []error) []error {
	if len(in) == 0 {
		return nil
	}
	out := make([]error, 0, len(in))
	for _, err := range in {
		if err == nil {
			continue
		}
		out = append(out, err)
	}
	return out
}

// ErrorOverview renders an error message together with the originating stack
// frame, an optional source excerpt, and the full stack trace.  The output is
// arranged as a column-oriented Box and is safe to embed inside any other
// layout. The function never panics for nil/zero values — an empty placeholder
// is rendered instead.
func ErrorOverview(props ErrorOverviewProps) *vdom.Node {
	frames := ParseGoStack(props.Stack)
	sectionChildren := []*vdom.Node{
		errorOverviewHeader(props.Err),
	}

	if origin, ok := firstUserFrame(frames); ok {
		sectionChildren = append(sectionChildren, errorOverviewLocation(origin))
		if props.SourceContext > 0 {
			if excerpt := errorOverviewExcerpt(origin, props.SourceContext); excerpt != nil {
				sectionChildren = append(sectionChildren, excerpt)
			}
		}
	}

	if stackBlock := errorOverviewStack(frames); stackBlock != nil {
		sectionChildren = append(sectionChildren, stackBlock)
	}

	return Box(vdom.Props{
		"flexDirection": "column",
		"padding":       1,
	}, sectionChildren...)
}

func errorOverviewHeader(err error) *vdom.Node {
	header := styles.WrapWithANSI(" ERROR ", styles.Red.ToANSI(styles.Background), styles.White.ToANSI(styles.Foreground), styles.BoldCode())

	message := ""
	if err != nil {
		message = err.Error()
	}
	if message == "" {
		message = "<nil error>"
	}

	return Box(vdom.Props{"flexDirection": "row"},
		Text(header),
		Text(" "+message),
	)
}

func errorOverviewLocation(origin StackFrame) *vdom.Node {
	location := fmt.Sprintf("%s:%d", origin.File, origin.Line)
	return Box(vdom.Props{"marginTop": 1},
		Text(styles.Dim(location)),
	)
}

func errorOverviewExcerpt(origin StackFrame, contextLines int) *vdom.Node {
	if origin.File == "" || origin.Line <= 0 {
		return nil
	}

	lines, ok := readSourceWindow(origin.File, origin.Line, contextLines)
	if !ok || len(lines) == 0 {
		return nil
	}

	maxLine := lines[len(lines)-1].number
	width := len(strconv.Itoa(maxLine))

	rows := make([]*vdom.Node, 0, len(lines))
	for _, entry := range lines {
		gutter := fmt.Sprintf("%*d:", width, entry.number)
		body := " " + entry.text

		if entry.number == origin.Line {
			highlight := styles.WrapWithANSI(
				gutter+body,
				styles.Red.ToANSI(styles.Background),
				styles.White.ToANSI(styles.Foreground),
			)
			rows = append(rows, Text(highlight))
			continue
		}

		rows = append(rows, Text(styles.Dim(gutter+body)))
	}

	return Box(vdom.Props{"marginTop": 1, "flexDirection": "column"}, rows...)
}

func errorOverviewStack(frames []StackFrame) *vdom.Node {
	if len(frames) == 0 {
		return nil
	}

	rows := make([]*vdom.Node, 0, len(frames))
	for _, frame := range frames {
		display := frame.Function
		if frame.File != "" {
			if display == "" {
				display = fmt.Sprintf("(%s:%d)", frame.File, frame.Line)
			} else {
				display = fmt.Sprintf("%s (%s:%d)", display, frame.File, frame.Line)
			}
		}
		if display == "" {
			continue
		}

		rows = append(rows, Text(styles.Dim("- "+display)))
	}

	if len(rows) == 0 {
		return nil
	}

	return Box(vdom.Props{"marginTop": 1, "flexDirection": "column"}, rows...)
}

// ParseGoStack parses output produced by runtime.Stack.  The input format is:
//
//	goroutine N [state]:
//	package.Function(args)
//	\tfile.go:line +0x...
//
// Inlined frames lack the trailing offset.  Goroutine headers and the
// optional "created by" suffix lines are skipped.  Anything that does not
// match the expected pattern is silently ignored to keep the renderer
// resilient to unfamiliar runtime output.
func ParseGoStack(s string) []StackFrame {
	if s == "" {
		return nil
	}

	frames := make([]StackFrame, 0, 8)
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var pending string
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty separators and goroutine state headers.
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			pending = ""
			continue
		}
		if strings.HasPrefix(trimmed, "goroutine ") && strings.HasSuffix(trimmed, ":") {
			pending = ""
			continue
		}

		// File/line rows are tab-indented in real runtime output but a leading
		// space is also tolerated to ease parsing of hand-formatted traces.
		if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "    ") {
			file, lineNo, ok := parseGoStackLocation(trimmed)
			if !ok {
				pending = ""
				continue
			}
			frames = append(frames, StackFrame{
				Function: pending,
				File:     file,
				Line:     lineNo,
			})
			pending = ""
			continue
		}

		// "created by ..." lines also belong to the trace; treat them as
		// function headers so the next location row attaches to them.
		// Newer Go runtimes append " in goroutine N" — strip that.
		if strings.HasPrefix(trimmed, "created by ") {
			rest := strings.TrimPrefix(trimmed, "created by ")
			if idx := strings.Index(rest, " in goroutine "); idx >= 0 {
				rest = rest[:idx]
			}
			pending = stripFunctionArgs(rest)
			continue
		}

		// Otherwise we're at a function header line.  We strip the trailing
		// argument list to keep the function name concise.
		pending = stripFunctionArgs(trimmed)
	}

	return frames
}

// firstUserFrame returns the first stack frame outside of the Go runtime,
// standard library panic helpers, and ParseGoStack itself.
func firstUserFrame(frames []StackFrame) (StackFrame, bool) {
	for _, frame := range frames {
		if isRuntimeFrame(frame) {
			continue
		}
		return frame, true
	}
	return StackFrame{}, false
}

func isRuntimeFrame(frame StackFrame) bool {
	fn := frame.Function
	if fn == "" {
		return true
	}
	switch {
	case strings.HasPrefix(fn, "runtime."),
		strings.HasPrefix(fn, "runtime/"),
		strings.HasPrefix(fn, "reflect."),
		strings.HasPrefix(fn, "testing."),
		strings.HasPrefix(fn, "panic"):
		return true
	}

	file := frame.File
	if file == "" {
		return false
	}
	// /usr/local/go/src/... or any path containing /src/runtime/
	if strings.Contains(file, "/src/runtime/") || strings.Contains(file, "/runtime/panic") {
		return true
	}
	return false
}

func parseGoStackLocation(line string) (string, int, bool) {
	// Expected forms:
	//   /path/to/file.go:42 +0x1c
	//   /path/to/file.go:42
	rest := line
	if idx := strings.Index(rest, " "); idx >= 0 {
		rest = rest[:idx]
	}

	colon := strings.LastIndex(rest, ":")
	if colon <= 0 || colon == len(rest)-1 {
		return "", 0, false
	}

	file := rest[:colon]
	lineNo, err := strconv.Atoi(rest[colon+1:])
	if err != nil || lineNo <= 0 {
		return "", 0, false
	}

	return file, lineNo, true
}

func stripFunctionArgs(line string) string {
	// The argument list is the LAST parenthesised group on the line.  Using
	// LastIndex preserves type qualifiers such as "pkg.(*T).Method" while still
	// dropping "pkg.Func(0x1, 0x2)" → "pkg.Func".
	if idx := strings.LastIndex(line, "("); idx >= 0 {
		return strings.TrimSpace(line[:idx])
	}
	return line
}

type sourceLine struct {
	number int
	text   string
}

func readSourceWindow(path string, target, contextLines int) ([]sourceLine, bool) {
	if path == "" || target <= 0 {
		return nil, false
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer file.Close()

	start := target - contextLines
	if start < 1 {
		start = 1
	}
	end := target + contextLines

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var window []sourceLine
	current := 0
	for scanner.Scan() {
		current++
		if current < start {
			continue
		}
		if current > end {
			break
		}
		window = append(window, sourceLine{number: current, text: scanner.Text()})
	}
	if err := scanner.Err(); err != nil {
		return nil, false
	}

	return window, true
}
