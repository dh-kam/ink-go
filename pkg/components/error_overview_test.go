package components_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func collectText(node *vdom.Node) string {
	if node == nil {
		return ""
	}
	return node.TextContent()
}

func TestErrorOverviewSimpleErrorNoStack(t *testing.T) {
	node := components.ErrorOverview(components.ErrorOverviewProps{
		Err: errors.New("boom"),
	})
	if node == nil {
		t.Fatal("ErrorOverview returned nil")
	}

	out := collectText(node)
	if !strings.Contains(out, "ERROR") {
		t.Fatalf("missing ERROR header: %q", out)
	}
	if !strings.Contains(out, "boom") {
		t.Fatalf("missing error message: %q", out)
	}

	// No stack means no per-frame "- " lines.
	if strings.Contains(out, "- ") {
		t.Fatalf("unexpected stack rows in output without stack: %q", out)
	}
}

func TestErrorOverviewNilErrorIsSafe(t *testing.T) {
	node := components.ErrorOverview(components.ErrorOverviewProps{Err: nil})
	if node == nil {
		t.Fatal("ErrorOverview returned nil for nil err")
	}
	out := collectText(node)
	if !strings.Contains(out, "ERROR") {
		t.Fatalf("missing header for nil err: %q", out)
	}
	if !strings.Contains(out, "<nil error>") {
		t.Fatalf("missing nil placeholder: %q", out)
	}
}

func TestParseGoStackRealRuntimeOutput(t *testing.T) {
	buf := make([]byte, 8192)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	frames := components.ParseGoStack(stack)
	if len(frames) == 0 {
		t.Fatalf("expected to parse at least one frame from runtime.Stack output:\n%s", stack)
	}

	for i, frame := range frames {
		if frame.Function == "" {
			t.Fatalf("frame %d has empty function: %+v\nfull stack:\n%s", i, frame, stack)
		}
		if frame.File == "" {
			t.Fatalf("frame %d has empty file: %+v", i, frame)
		}
		if frame.Line <= 0 {
			t.Fatalf("frame %d has invalid line %d", i, frame.Line)
		}
	}

	// Function names should not include argument lists.
	for _, frame := range frames {
		if strings.Contains(frame.Function, "(") {
			t.Fatalf("function name should be stripped of args: %q", frame.Function)
		}
	}
}

func TestParseGoStackHandcraftedSample(t *testing.T) {
	sample := "goroutine 1 [running]:\n" +
		"main.inner()\n" +
		"\t/tmp/stacktest.go:10 +0x40\n" +
		"main.outer(...)\n" +
		"\t/tmp/stacktest.go:14\n" +
		"main.main()\n" +
		"\t/tmp/stacktest.go:16 +0x1c\n"

	frames := components.ParseGoStack(sample)
	if len(frames) != 3 {
		t.Fatalf("expected 3 frames, got %d: %+v", len(frames), frames)
	}

	expected := []components.StackFrame{
		{Function: "main.inner", File: "/tmp/stacktest.go", Line: 10},
		{Function: "main.outer", File: "/tmp/stacktest.go", Line: 14},
		{Function: "main.main", File: "/tmp/stacktest.go", Line: 16},
	}
	for i, want := range expected {
		got := frames[i]
		if got != want {
			t.Errorf("frame %d: got %+v want %+v", i, got, want)
		}
	}
}

func TestParseGoStackEmpty(t *testing.T) {
	if got := components.ParseGoStack(""); got != nil {
		t.Fatalf("expected nil for empty input, got %+v", got)
	}
}

func TestParseGoStackCreatedBy(t *testing.T) {
	sample := "goroutine 7 [running]:\n" +
		"created by main.spawn\n" +
		"\t/tmp/spawn.go:42 +0x80\n"
	frames := components.ParseGoStack(sample)
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d: %+v", len(frames), frames)
	}
	if frames[0].Function != "main.spawn" {
		t.Fatalf("created-by function not captured: %+v", frames[0])
	}
	if frames[0].File != "/tmp/spawn.go" || frames[0].Line != 42 {
		t.Fatalf("created-by location not captured: %+v", frames[0])
	}
}

func TestErrorOverviewMissingFileSilentSkip(t *testing.T) {
	stack := "goroutine 1 [running]:\n" +
		"main.boom()\n" +
		"\t/this/path/does/not/exist_xyz.go:7 +0x12\n"

	node := components.ErrorOverview(components.ErrorOverviewProps{
		Err:           errors.New("kaboom"),
		Stack:         stack,
		SourceContext: 2,
	})
	if node == nil {
		t.Fatal("ErrorOverview returned nil")
	}
	out := collectText(node)
	if !strings.Contains(out, "kaboom") {
		t.Fatalf("missing message: %q", out)
	}
	if !strings.Contains(out, "/this/path/does/not/exist_xyz.go:7") {
		t.Fatalf("expected location even when source missing: %q", out)
	}
	// Should not contain a gutter prefix like " 7:" — excerpt would have rendered that.
	if strings.Contains(out, " 7: ") {
		t.Fatalf("did not expect a source excerpt for missing file: %q", out)
	}
}

func TestErrorOverviewSourceExcerptFromTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")

	const targetLine = 4
	content := "line 1\nline 2\nline 3\nFAILING LINE 4\nline 5\nline 6\nline 7\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write tmp source: %v", err)
	}

	stack := fmt.Sprintf("goroutine 1 [running]:\nmain.boom()\n\t%s:%d +0x10\n", path, targetLine)

	node := components.ErrorOverview(components.ErrorOverviewProps{
		Err:           errors.New("explode"),
		Stack:         stack,
		SourceContext: 2,
	})
	out := collectText(node)

	for _, want := range []string{"line 2", "line 3", "FAILING LINE 4", "line 5", "line 6"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected excerpt to contain %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "line 1\n") {
		t.Errorf("did not expect line 1 (outside context window):\n%s", out)
	}
	if strings.Contains(out, "line 7") {
		t.Errorf("did not expect line 7 (outside context window):\n%s", out)
	}
	if !strings.Contains(out, fmt.Sprintf("%s:%d", path, targetLine)) {
		t.Errorf("missing location header in:\n%s", out)
	}
}

func TestErrorOverviewSourceContextZeroSkipsExcerpt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "src.go")
	if err := os.WriteFile(path, []byte("a\nb\nUNIQ_TARGET\nd\n"), 0o644); err != nil {
		t.Fatalf("write tmp source: %v", err)
	}
	stack := fmt.Sprintf("goroutine 1 [running]:\nmain.boom()\n\t%s:3 +0x10\n", path)

	node := components.ErrorOverview(components.ErrorOverviewProps{
		Err:           errors.New("e"),
		Stack:         stack,
		SourceContext: 0,
	})
	out := collectText(node)
	if strings.Contains(out, "UNIQ_TARGET") {
		t.Fatalf("SourceContext=0 should suppress excerpt, but found UNIQ_TARGET in: %q", out)
	}
}

func TestErrorOverviewStackBlockRendered(t *testing.T) {
	stack := "goroutine 1 [running]:\n" +
		"runtime.gopanic(...)\n" +
		"\t/usr/local/go/src/runtime/panic.go:838 +0x100\n" +
		"main.boom()\n" +
		"\t/home/me/proj/main.go:24 +0x40\n"

	node := components.ErrorOverview(components.ErrorOverviewProps{
		Err:   errors.New("kab"),
		Stack: stack,
	})
	out := collectText(node)

	if !strings.Contains(out, "main.boom") {
		t.Fatalf("expected stack to include user frame: %q", out)
	}
	if !strings.Contains(out, "runtime.gopanic") {
		t.Fatalf("expected stack to include runtime frame: %q", out)
	}
	// The location header should anchor on the user frame.
	if !strings.Contains(out, "/home/me/proj/main.go:24") {
		t.Fatalf("expected user-frame location header: %q", out)
	}
}

func TestErrorOverviewDisplaysCurrentWorkspacePathsRelatively(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}

	path := filepath.Join(cwd, "fixture.go")
	stack := fmt.Sprintf("goroutine 1 [running]:\nmain.boom()\n\t%s:7 +0x10\n", path)
	node := components.ErrorOverview(components.ErrorOverviewProps{
		Err:   errors.New("kaboom"),
		Stack: stack,
	})
	out := collectText(node)

	if !strings.Contains(out, "fixture.go:7") {
		t.Fatalf("expected relative path in output: %q", out)
	}
	if strings.Contains(out, cwd) {
		t.Fatalf("current workspace path should be stripped: %q", out)
	}
}

func TestErrorOverviewSkipsBoundaryInternalsForOrigin(t *testing.T) {
	stack := "goroutine 1 [running]:\n" +
		"github.com/dh-kam/ink-go/pkg/components.captureStack()\n" +
		"\t/work/pkg/components/error_boundary.go:112 +0x20\n" +
		"github.com/dh-kam/ink-go/pkg/components.safeRender.func1()\n" +
		"\t/work/pkg/components/error_boundary.go:83 +0x20\n" +
		"panic({0x1, 0x2})\n" +
		"\t/usr/local/go/src/runtime/panic.go:783 +0x20\n" +
		"main.Test()\n" +
		"\t/app/main.go:23 +0x20\n"

	node := components.ErrorOverview(components.ErrorOverviewProps{
		Err:   errors.New("Oh no"),
		Stack: stack,
	})
	out := collectText(node)

	if !strings.Contains(out, "/app/main.go:23") {
		t.Fatalf("expected user panic origin, got %q", out)
	}
	if strings.Contains(out, "/work/pkg/components/error_boundary.go:112") {
		t.Fatalf("boundary internals should not be used as origin: %q", out)
	}
}

func TestErrorOverviewGroupRendersValidationAndRuntimeSections(t *testing.T) {
	node := components.ErrorOverviewGroup(components.ErrorOverviewGroupProps{
		Validation: []error{
			errors.New("email is required"),
			errors.New("password too short"),
		},
		Runtime: []error{
			errors.New("network timeout"),
		},
	})
	if node == nil {
		t.Fatal("ErrorOverviewGroup returned nil")
	}

	out := collectText(node)
	for _, want := range []string{"ERRORS", "Validation", "Runtime", "email is required", "password too short", "network timeout"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
	// Total error count should appear in the header when count > 1.
	if !strings.Contains(out, "(3)") {
		t.Errorf("expected total count suffix '(3)' in header, got %q", out)
	}
}

func TestErrorOverviewGroupSkipsNilEntries(t *testing.T) {
	node := components.ErrorOverviewGroup(components.ErrorOverviewGroupProps{
		Validation: []error{nil, errors.New("kept"), nil},
		Runtime:    []error{nil},
	})
	out := collectText(node)
	if !strings.Contains(out, "kept") {
		t.Fatalf("expected non-nil error to render, got %q", out)
	}
	// Single non-nil error → no count suffix.
	if strings.Contains(out, "(1)") {
		t.Fatalf("did not expect count suffix for single error, got %q", out)
	}
	// Runtime had only nils → its sub-section header should be omitted.
	if strings.Contains(out, "Runtime") {
		t.Fatalf("expected Runtime section to be hidden when all entries are nil, got %q", out)
	}
}

func TestErrorOverviewGroupEmptyShowsPlaceholder(t *testing.T) {
	node := components.ErrorOverviewGroup(components.ErrorOverviewGroupProps{})
	out := collectText(node)
	if !strings.Contains(out, "<no errors>") {
		t.Fatalf("expected empty placeholder, got %q", out)
	}
	if !strings.Contains(out, "ERRORS") {
		t.Fatalf("expected default ERRORS title, got %q", out)
	}
}

func TestErrorOverviewGroupCustomTitleAndStack(t *testing.T) {
	stack := "goroutine 1 [running]:\n" +
		"main.fail()\n" +
		"\t/home/me/proj/app.go:42 +0x40\n"

	node := components.ErrorOverviewGroup(components.ErrorOverviewGroupProps{
		Title:   "VALIDATION FAILED",
		Runtime: []error{errors.New("downstream call failed")},
		Stack:   stack,
	})
	out := collectText(node)
	if !strings.Contains(out, "VALIDATION FAILED") {
		t.Fatalf("expected custom title, got %q", out)
	}
	if !strings.Contains(out, "main.fail") {
		t.Fatalf("expected stack frame in output, got %q", out)
	}
	if !strings.Contains(out, "/home/me/proj/app.go:42") {
		t.Fatalf("expected location header, got %q", out)
	}
}

func TestErrorOverviewNegativeSourceContextSkipsExcerpt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.go")
	_ = os.WriteFile(path, []byte("only-line\n"), 0o644)

	stack := fmt.Sprintf("main.f()\n\t%s:1 +0x0\n", path)
	node := components.ErrorOverview(components.ErrorOverviewProps{
		Err:           errors.New("z"),
		Stack:         stack,
		SourceContext: -3,
	})
	out := collectText(node)
	if strings.Contains(out, "only-line") {
		t.Fatalf("negative SourceContext should suppress excerpt: %q", out)
	}
}
