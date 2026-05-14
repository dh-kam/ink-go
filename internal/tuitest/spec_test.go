package tuitest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSpecFileSupportsYAMLAndEmbeddedExpectation(t *testing.T) {
	path := writeTempFile(t, "scenario.yaml", `schemaVersion: goink.tuitest/v1alpha1
name: sample/yaml
app: sample
viewport:
  width: 80
  height: 24
capture:
  mode: plain
environment:
  TERM: tmux-256color
  COLORTERM: ""
  FORCE_COLOR: null
steps:
  - name: initial
    expect:
      text: |-
        hello
`)

	spec, err := LoadSpecFile(path)
	if err != nil {
		t.Fatalf("load YAML spec: %v", err)
	}
	if spec.Name != "sample/yaml" {
		t.Fatalf("expected scenario name, got %q", spec.Name)
	}
	if spec.Steps[0].Expect == nil || spec.Steps[0].Expect.Text == nil || *spec.Steps[0].Expect.Text != "hello" {
		t.Fatalf("expected embedded golden text, got %#v", spec.Steps[0].Expect)
	}
	if spec.Environment["TERM"] == nil || *spec.Environment["TERM"] != "tmux-256color" {
		t.Fatalf("expected TERM env, got %#v", spec.Environment["TERM"])
	}
	if spec.Environment["COLORTERM"] == nil || *spec.Environment["COLORTERM"] != "" {
		t.Fatalf("expected empty COLORTERM env, got %#v", spec.Environment["COLORTERM"])
	}
	if spec.Environment["FORCE_COLOR"] != nil {
		t.Fatalf("expected FORCE_COLOR unset marker, got %#v", spec.Environment["FORCE_COLOR"])
	}
}

func TestLoadSpecFileSupportsJSONAndExternalExpectation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "frame.golden"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write golden: %v", err)
	}

	path := filepath.Join(dir, "scenario.json")
	if err := os.WriteFile(path, []byte(`{
  "schemaVersion": "goink.tuitest/v1alpha1",
  "name": "sample/json",
  "app": "sample",
  "viewport": {"width": 80, "height": 24},
  "capture": {"mode": "plain"},
  "steps": [
    {"name": "initial", "expect": {"file": "frame.golden"}}
  ]
}`), 0o644); err != nil {
		t.Fatalf("write scenario: %v", err)
	}

	spec, err := LoadSpecFile(path)
	if err != nil {
		t.Fatalf("load JSON spec: %v", err)
	}
	expected, ok := expectedText(t, spec.Steps[0], filepath.Dir(path), true, "")
	if !ok {
		t.Fatal("expected external golden to resolve")
	}
	if expected != "hello" {
		t.Fatalf("expected external golden content, got %q", expected)
	}
}

func TestApplyScenarioEnvironmentRestoresPreviousValues(t *testing.T) {
	t.Setenv("GOINK_TUITEST_EXISTING", "old")
	_ = os.Unsetenv("GOINK_TUITEST_EMPTY")
	_ = os.Unsetenv("GOINK_TUITEST_UNSET")

	empty := ""
	unset := (*string)(nil)
	restore := applyScenarioEnvironment(t, Environment{
		"GOINK_TUITEST_EXISTING": strPtr("new"),
		"GOINK_TUITEST_EMPTY":    &empty,
		"GOINK_TUITEST_UNSET":    unset,
	})

	if value := os.Getenv("GOINK_TUITEST_EXISTING"); value != "new" {
		t.Fatalf("expected overridden existing env, got %q", value)
	}
	if value, ok := os.LookupEnv("GOINK_TUITEST_EMPTY"); !ok || value != "" {
		t.Fatalf("expected empty env to be set, got %q exists=%v", value, ok)
	}
	if _, ok := os.LookupEnv("GOINK_TUITEST_UNSET"); ok {
		t.Fatal("expected unset env")
	}

	restore()

	if value := os.Getenv("GOINK_TUITEST_EXISTING"); value != "old" {
		t.Fatalf("expected existing env to restore, got %q", value)
	}
	if _, ok := os.LookupEnv("GOINK_TUITEST_EMPTY"); ok {
		t.Fatal("expected empty env key to be removed")
	}
	if _, ok := os.LookupEnv("GOINK_TUITEST_UNSET"); ok {
		t.Fatal("expected unset env key to remain absent")
	}
}

func writeTempFile(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", name, err)
	}
	return path
}

func strPtr(value string) *string {
	return &value
}
