package components

import "testing"

func TestFirstUserFrameSkipsRuntime(t *testing.T) {
	frames := []StackFrame{
		{Function: "runtime.gopanic", File: "/usr/local/go/src/runtime/panic.go", Line: 100},
		{Function: "runtime.goexit", File: "/usr/local/go/src/runtime/asm_amd64.s", Line: 1},
		{Function: "panic", File: "", Line: 0},
		{Function: "main.boom", File: "/home/me/proj/main.go", Line: 24},
		{Function: "main.main", File: "/home/me/proj/main.go", Line: 7},
	}

	got, ok := firstUserFrame(frames)
	if !ok {
		t.Fatal("expected a user frame")
	}
	if got.Function != "main.boom" {
		t.Fatalf("unexpected first user frame: %+v", got)
	}
}

func TestFirstUserFrameAllRuntimeReturnsFalse(t *testing.T) {
	frames := []StackFrame{
		{Function: "runtime.gopanic", File: "/usr/local/go/src/runtime/panic.go", Line: 1},
		{Function: "runtime.main", File: "/usr/local/go/src/runtime/proc.go", Line: 250},
	}
	if _, ok := firstUserFrame(frames); ok {
		t.Fatal("expected no user frame")
	}
}

func TestFirstUserFrameEmpty(t *testing.T) {
	if _, ok := firstUserFrame(nil); ok {
		t.Fatal("expected no user frame for empty list")
	}
}

func TestFirstUserFrameDetectsRuntimeByPath(t *testing.T) {
	frames := []StackFrame{
		// Function name doesn't start with runtime. but file is in /src/runtime/.
		{Function: "weird.name", File: "/usr/local/go/src/runtime/asm.s", Line: 1},
		{Function: "user.fn", File: "/me/code.go", Line: 12},
	}
	got, ok := firstUserFrame(frames)
	if !ok || got.Function != "user.fn" {
		t.Fatalf("expected user.fn, got %+v ok=%v", got, ok)
	}
}

func TestStripFunctionArgs(t *testing.T) {
	cases := map[string]string{
		"main.f(0x1, 0x2)":           "main.f",
		"main.f(...)":                "main.f",
		"main.f":                     "main.f",
		"":                           "",
		"testing.(*T).Run(0x1, 0x2)": "testing.(*T).Run",
		"pkg.(*Type).Method(0x1)":    "pkg.(*Type).Method",
	}
	for in, want := range cases {
		if got := stripFunctionArgs(in); got != want {
			t.Errorf("stripFunctionArgs(%q) = %q want %q", in, got, want)
		}
	}
}

func TestParseGoStackLocationInvalid(t *testing.T) {
	cases := []string{
		"",
		"no-colon-here",
		"file.go:not-a-number",
		"file.go:",
		":42",
		"file.go:0",
		"file.go:-3",
	}
	for _, in := range cases {
		if _, _, ok := parseGoStackLocation(in); ok {
			t.Errorf("parseGoStackLocation(%q) unexpectedly ok", in)
		}
	}
}

func TestParseGoStackLocationValid(t *testing.T) {
	file, line, ok := parseGoStackLocation("/tmp/x.go:17 +0x40")
	if !ok || file != "/tmp/x.go" || line != 17 {
		t.Fatalf("got %q %d ok=%v", file, line, ok)
	}

	file, line, ok = parseGoStackLocation("/tmp/x.go:42")
	if !ok || file != "/tmp/x.go" || line != 42 {
		t.Fatalf("got %q %d ok=%v", file, line, ok)
	}
}

func TestReadSourceWindowMissing(t *testing.T) {
	if _, ok := readSourceWindow("/no/such/file/here.go", 1, 2); ok {
		t.Fatal("expected ok=false for missing file")
	}
	if _, ok := readSourceWindow("", 1, 2); ok {
		t.Fatal("expected ok=false for empty path")
	}
	if _, ok := readSourceWindow("/anything", 0, 2); ok {
		t.Fatal("expected ok=false for non-positive target")
	}
}
