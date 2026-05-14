package renderer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ansiRE matches the ANSI escape sequences emitted by Goink's renderers.
// Three alternations cover (in priority order):
//   - private CSI (\x1b[?...letter), listed first so the leading ? doesn't
//     get consumed by the generic CSI branch
//   - standard 7-bit CSI (\x1b[parameters intermediate-bytes final-byte)
//   - two-byte ESC introducer (\x1b<letter>)
//
// The 8-bit CSI form (single byte 0x9b) is intentionally omitted: Go's
// regexp engine treats patterns and inputs as UTF-8, so a lone high byte
// would never round-trip correctly. Modern terminals universally emit the
// 7-bit ESC[ form anyway.
var ansiRE = regexp.MustCompile(
	`\x1b\[\?[0-9;]*[ -/]*[@-~]` +
		`|\x1b\[[0-9;]*[ -/]*[@-~]` +
		`|\x1b[@-_]`,
)

// StripANSI returns s with all ANSI escape sequences removed. Idempotent.
func StripANSI(s string) string {
	if s == "" {
		return s
	}
	return ansiRE.ReplaceAllString(s, "")
}

// snapshotDir is the directory snapshots live under, relative to the package
// running the test.
const snapshotDir = "testdata/__snapshots__"

// MatchSnapshot compares actual against the stored snapshot for name.
// First call: writes actual as the new snapshot and passes.
// Later calls: fails the test if actual differs from the stored content.
//
// Set the UPDATE_SNAPSHOTS env var to "1", "true", "yes", "on", "y", or
// "t" (case-insensitive) to force-rewrite snapshots — useful when an
// intentional change ripples through many goldens.
//
// Snapshot files live at testdata/__snapshots__/<sanitized-name>.txt
// relative to the test package. The name is sanitized to strip path
// separators — callers cannot escape the snapshot directory.
func MatchSnapshot(t testing.TB, name string, actual string) {
	t.Helper()
	if name == "" {
		t.Fatal("MatchSnapshot: name must not be empty")
	}
	safeName := sanitizeName(name)
	if safeName == "" {
		t.Fatalf("MatchSnapshot: name %q sanitized to empty string", name)
	}

	path := filepath.Join(snapshotDir, safeName+".txt")
	update := shouldUpdate()

	existing, err := os.ReadFile(path)
	switch {
	case os.IsNotExist(err):
		if err := writeSnapshot(path, actual); err != nil {
			t.Fatalf("MatchSnapshot: write %q: %v", path, err)
		}
		return
	case err != nil:
		t.Fatalf("MatchSnapshot: read %q: %v", path, err)
	}

	if update {
		if err := writeSnapshot(path, actual); err != nil {
			t.Fatalf("MatchSnapshot: update %q: %v", path, err)
		}
		return
	}

	if string(existing) != actual {
		t.Fatalf("MatchSnapshot: %q mismatch\nwant:\n%s\n\ngot:\n%s",
			path, string(existing), actual)
	}
}

func writeSnapshot(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// sanitizeName strips any path separator so callers cannot escape the
// snapshot directory via "../foo" or "a/b" tricks.
func sanitizeName(name string) string {
	r := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		string(os.PathSeparator), "_",
		"..", "_",
	)
	return strings.TrimSpace(r.Replace(name))
}

// shouldUpdate reports whether UPDATE_SNAPSHOTS is set to a truthy value.
func shouldUpdate() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("UPDATE_SNAPSHOTS")))
	switch v {
	case "1", "true", "yes", "y", "on", "t":
		return true
	}
	return false
}
