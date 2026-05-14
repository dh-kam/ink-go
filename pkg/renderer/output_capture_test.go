package renderer_test

import (
	"errors"
	"io"
	"testing"

	"github.com/dh-kam/ink-go/pkg/renderer"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func TestStdoutNilWhenNotConfigured(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"))
	defer inst.Cleanup()
	if inst.Stdout() != nil {
		t.Fatal("Stdout() should be nil without WithStdoutCapture option")
	}
	if frames := inst.StdoutFrames(); frames != nil {
		t.Fatalf("StdoutFrames() should be nil unconfigured, got %v", frames)
	}
}

func TestStderrNilWhenNotConfigured(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"))
	defer inst.Cleanup()
	if inst.Stderr() != nil {
		t.Fatal("Stderr() should be nil without WithStderrCapture option")
	}
	if frames := inst.StderrFrames(); frames != nil {
		t.Fatalf("StderrFrames() should be nil unconfigured, got %v", frames)
	}
}

func TestStdoutCaptureAccumulatesFramesPerWrite(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdoutCapture())
	defer inst.Cleanup()

	w := inst.Stdout()
	if w == nil {
		t.Fatal("Stdout() should be non-nil with WithStdoutCapture")
	}
	if n, err := w.Write([]byte("hello")); err != nil || n != 5 {
		t.Fatalf("Write hello: n=%d err=%v", n, err)
	}
	if n, err := w.Write([]byte("world")); err != nil || n != 5 {
		t.Fatalf("Write world: n=%d err=%v", n, err)
	}
	if n, err := w.Write(nil); err != nil || n != 0 {
		t.Fatalf("Write nil: n=%d err=%v", n, err)
	}

	got := inst.StdoutFrames()
	want := []string{"hello", "world", ""}
	if len(got) != len(want) {
		t.Fatalf("len(frames) = %d, want %d (%v)", len(got), len(want), got)
	}
	for i, g := range got {
		if g != want[i] {
			t.Fatalf("frame[%d] = %q, want %q", i, g, want[i])
		}
	}
}

func TestStderrCaptureAccumulatesFramesPerWrite(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStderrCapture())
	defer inst.Cleanup()

	w := inst.Stderr()
	if w == nil {
		t.Fatal("Stderr() should be non-nil with WithStderrCapture")
	}
	if _, err := io.WriteString(w, "boom"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	got := inst.StderrFrames()
	if len(got) != 1 || got[0] != "boom" {
		t.Fatalf("StderrFrames = %v, want [boom]", got)
	}
}

func TestStdoutFramesDefensiveCopy(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdoutCapture())
	defer inst.Cleanup()

	w := inst.Stdout()
	if _, err := w.Write([]byte("first")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	snap := inst.StdoutFrames()
	// Mutate the returned slice — must not bleed into Instance state.
	snap[0] = "MUTATED"
	snap = append(snap, "extra")
	_ = snap

	again := inst.StdoutFrames()
	if len(again) != 1 || again[0] != "first" {
		t.Fatalf("internal state mutated by caller: got %v", again)
	}
}

func TestStdoutWriteDefensiveCopyAgainstCallerBuffer(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdoutCapture())
	defer inst.Cleanup()

	buf := []byte("hello")
	if _, err := inst.Stdout().Write(buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	buf[0] = 'X' // mutate caller's buffer — must not affect captured frame

	frames := inst.StdoutFrames()
	if len(frames) != 1 || frames[0] != "hello" {
		t.Fatalf("captured frame aliases caller buffer: %v", frames)
	}
}

func TestStdoutClosedAfterCleanupReturnsErrClosedPipe(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdoutCapture())
	w := inst.Stdout()
	inst.Cleanup()

	n, err := w.Write([]byte("after"))
	if err == nil {
		t.Fatal("expected error writing to closed stdout capture")
	}
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("err = %v, want io.ErrClosedPipe", err)
	}
	if n != 0 {
		t.Fatalf("n = %d on closed write, want 0", n)
	}
}

func TestStderrClosedAfterCleanupReturnsErrClosedPipe(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStderrCapture())
	w := inst.Stderr()
	inst.Cleanup()

	_, err := w.Write([]byte("after"))
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("err = %v, want io.ErrClosedPipe", err)
	}
}

func TestStdoutAndStderrIndependent(t *testing.T) {
	inst := renderer.Render(
		vdom.CreateTextNode("x"),
		renderer.WithStdoutCapture(),
		renderer.WithStderrCapture(),
	)
	defer inst.Cleanup()

	if _, err := inst.Stdout().Write([]byte("OUT")); err != nil {
		t.Fatalf("stdout Write: %v", err)
	}
	if _, err := inst.Stderr().Write([]byte("ERR")); err != nil {
		t.Fatalf("stderr Write: %v", err)
	}

	out := inst.StdoutFrames()
	errs := inst.StderrFrames()
	if len(out) != 1 || out[0] != "OUT" {
		t.Fatalf("stdout frames = %v, want [OUT]", out)
	}
	if len(errs) != 1 || errs[0] != "ERR" {
		t.Fatalf("stderr frames = %v, want [ERR]", errs)
	}
}

func TestStdoutFramesEmptyBeforeAnyWrite(t *testing.T) {
	inst := renderer.Render(vdom.CreateTextNode("x"), renderer.WithStdoutCapture())
	defer inst.Cleanup()
	frames := inst.StdoutFrames()
	if frames == nil {
		t.Fatal("StdoutFrames should be non-nil (empty slice) once configured")
	}
	if len(frames) != 0 {
		t.Fatalf("StdoutFrames = %v, want []", frames)
	}
}
