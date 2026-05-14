package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripANSIRemovesCSI(t *testing.T) {
	in := "\x1b[31mred\x1b[0m text"
	if got := StripANSI(in); got != "red text" {
		t.Fatalf("StripANSI = %q, want %q", got, "red text")
	}
}

func TestStripANSIRemovesPrivateCSI(t *testing.T) {
	in := "\x1b[?25hcursor\x1b[?25l"
	if got := StripANSI(in); got != "cursor" {
		t.Fatalf("StripANSI = %q, want %q", got, "cursor")
	}
}

func TestStripANSIRemovesESCLetter(t *testing.T) {
	in := "before\x1bAafter"
	if got := StripANSI(in); got != "beforeafter" {
		t.Fatalf("StripANSI = %q, want %q", got, "beforeafter")
	}
}

func TestStripANSIIdempotent(t *testing.T) {
	in := "plain text"
	once := StripANSI(in)
	twice := StripANSI(once)
	if once != twice || once != in {
		t.Fatalf("idempotency broken: in=%q once=%q twice=%q", in, once, twice)
	}
}

func TestStripANSIEmptyString(t *testing.T) {
	if got := StripANSI(""); got != "" {
		t.Fatalf("StripANSI(\"\") = %q, want \"\"", got)
	}
}

func TestMatchSnapshotCreatesThenCompares(t *testing.T) {
	dir := t.TempDir()
	withCwd(t, dir, func() {
		MatchSnapshot(t, "create-then-compare", "hello world")

		path := filepath.Join(dir, snapshotDir, "create-then-compare.txt")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("snapshot not created at %s: %v", path, err)
		}
		// Second call with same content should pass quietly.
		MatchSnapshot(t, "create-then-compare", "hello world")
	})
}

func TestMatchSnapshotMismatchFails(t *testing.T) {
	dir := t.TempDir()
	withCwd(t, dir, func() {
		MatchSnapshot(t, "mismatch", "first")
		fake := &fakeTB{TB: t}
		MatchSnapshot(fake, "mismatch", "second")
		if !fake.failed {
			t.Fatal("expected mismatch to call Fatalf on the testing.TB")
		}
	})
}

func TestMatchSnapshotUpdatesWithEnv(t *testing.T) {
	dir := t.TempDir()
	withCwd(t, dir, func() {
		MatchSnapshot(t, "updated", "v1")

		t.Setenv("UPDATE_SNAPSHOTS", "1")
		MatchSnapshot(t, "updated", "v2")

		path := filepath.Join(dir, snapshotDir, "updated.txt")
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(got) != "v2" {
			t.Fatalf("snapshot after update = %q, want v2", string(got))
		}
	})
}

func TestSanitizeNameStripsSeparators(t *testing.T) {
	out := sanitizeName("../escape/me")
	if strings.Contains(out, "/") || strings.Contains(out, "..") {
		t.Fatalf("sanitizeName left dangerous chars: %q", out)
	}
}

func TestShouldUpdateValues(t *testing.T) {
	cases := map[string]bool{
		"":      false,
		"1":     true,
		"TRUE":  true,
		"true":  true,
		"yes":   true,
		"on":    true,
		"y":     true,
		"t":     true,
		"false": false,
		"no":    false,
	}
	for v, want := range cases {
		t.Setenv("UPDATE_SNAPSHOTS", v)
		if got := shouldUpdate(); got != want {
			t.Errorf("shouldUpdate(%q) = %v, want %v", v, got, want)
		}
	}
}

// withCwd chdir'd into dir for the duration of fn. Restores cwd on exit.
func withCwd(t *testing.T, dir string, fn func()) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore Chdir: %v", err)
		}
	}()
	fn()
}

// fakeTB intercepts Fatalf so we can assert that MatchSnapshot triggered
// failure. Falls through to the embedded *testing.T for everything else.
type fakeTB struct {
	testing.TB
	failed bool
}

func (f *fakeTB) Fatal(args ...any)                 { f.failed = true }
func (f *fakeTB) Fatalf(format string, args ...any) { f.failed = true }
func (f *fakeTB) Errorf(format string, args ...any) { f.failed = true }
