package tuitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = "goink.tuitest/v1alpha1"

type Spec struct {
	SchemaVersion string      `json:"schemaVersion" yaml:"schemaVersion"`
	Name          string      `json:"name" yaml:"name"`
	App           string      `json:"app" yaml:"app"`
	Viewport      Viewport    `json:"viewport" yaml:"viewport"`
	Capture       CaptureSpec `json:"capture" yaml:"capture"`
	Environment   Environment `json:"environment,omitempty" yaml:"environment,omitempty"`
	Steps         []StepSpec  `json:"steps" yaml:"steps"`
}

type Environment map[string]*string

type Viewport struct {
	Width  int `json:"width" yaml:"width"`
	Height int `json:"height" yaml:"height"`
}

type CaptureSpec struct {
	Mode                string `json:"mode" yaml:"mode"`
	TrimTrailingNewline *bool  `json:"trimTrailingNewline,omitempty" yaml:"trimTrailingNewline,omitempty"`
}

type StepSpec struct {
	Name    string       `json:"name" yaml:"name"`
	Resize  *Viewport    `json:"resize,omitempty" yaml:"resize,omitempty"`
	Input   *InputSpec   `json:"input,omitempty" yaml:"input,omitempty"`
	Wait    string       `json:"wait,omitempty" yaml:"wait,omitempty"`
	WaitFor *WaitForSpec `json:"waitFor,omitempty" yaml:"waitFor,omitempty"`
	Exit    *ExitSpec    `json:"exit,omitempty" yaml:"exit,omitempty"`
	Expect  *ExpectSpec  `json:"expect,omitempty" yaml:"expect,omitempty"`
}

type InputSpec struct {
	Text string `json:"text,omitempty" yaml:"text,omitempty"`
	Key  string `json:"key,omitempty" yaml:"key,omitempty"`
	Hex  string `json:"hex,omitempty" yaml:"hex,omitempty"`
}

type WaitForSpec struct {
	Text   string `json:"text,omitempty" yaml:"text,omitempty"`
	Within string `json:"within,omitempty" yaml:"within,omitempty"`
}

type ExitSpec struct {
	Within        string `json:"within,omitempty" yaml:"within,omitempty"`
	NoExtraWrites *bool  `json:"noExtraWrites,omitempty" yaml:"noExtraWrites,omitempty"`
}

type ExpectSpec struct {
	Text           *string `json:"text,omitempty" yaml:"text,omitempty"`
	File           string  `json:"file,omitempty" yaml:"file,omitempty"`
	SameAsPrevious bool    `json:"sameAsPrevious,omitempty" yaml:"sameAsPrevious,omitempty"`
}

func LoadSpecFile(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario %q: %w", path, err)
	}

	spec := &Spec{}
	switch ext := filepath.Ext(path); ext {
	case ".json":
		if err := decodeJSON(data, spec); err != nil {
			return nil, fmt.Errorf("decode JSON scenario %q: %w", path, err)
		}
	case ".yaml", ".yml":
		if err := decodeYAML(data, spec); err != nil {
			return nil, fmt.Errorf("decode YAML scenario %q: %w", path, err)
		}
	default:
		return nil, fmt.Errorf("unsupported scenario extension %q", ext)
	}

	if err := spec.validate(); err != nil {
		return nil, fmt.Errorf("validate scenario %q: %w", path, err)
	}

	return spec, nil
}

func decodeJSON(data []byte, spec *Spec) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(spec)
}

func decodeYAML(data []byte, spec *Spec) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	return decoder.Decode(spec)
}

func (spec *Spec) validate() error {
	if spec.SchemaVersion == "" {
		return fmt.Errorf("schemaVersion is required")
	}
	if spec.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schemaVersion %q", spec.SchemaVersion)
	}
	if spec.Name == "" {
		return fmt.Errorf("name is required")
	}
	if spec.App == "" {
		return fmt.Errorf("app is required")
	}
	if spec.Viewport.Width <= 0 {
		return fmt.Errorf("viewport.width must be positive")
	}
	if spec.Viewport.Height <= 0 {
		return fmt.Errorf("viewport.height must be positive")
	}
	if len(spec.Steps) == 0 {
		return fmt.Errorf("steps must not be empty")
	}
	switch strings.TrimSpace(strings.ToLower(spec.Capture.Mode)) {
	case "", CapturePlain, CaptureScreen, CaptureANSI, CaptureANSIEscape:
	default:
		return fmt.Errorf("unsupported capture.mode %q", spec.Capture.Mode)
	}
	for key := range spec.Environment {
		if key == "" {
			return fmt.Errorf("environment key must not be empty")
		}
		if strings.Contains(key, "=") {
			return fmt.Errorf("environment key must not contain '=': %q", key)
		}
	}
	for index, step := range spec.Steps {
		if step.Name == "" {
			return fmt.Errorf("steps[%d].name is required", index)
		}
		if step.Resize != nil {
			if step.Resize.Width <= 0 {
				return fmt.Errorf("steps[%d].resize.width must be positive", index)
			}
			if step.Resize.Height <= 0 {
				return fmt.Errorf("steps[%d].resize.height must be positive", index)
			}
		}
		if step.Wait != "" {
			if _, err := parseDuration(step.Wait); err != nil {
				return fmt.Errorf("steps[%d].wait must be a duration: %w", index, err)
			}
		}
		if step.WaitFor != nil {
			if step.WaitFor.Text == "" {
				return fmt.Errorf("steps[%d].waitFor.text is required", index)
			}
			if step.WaitFor.Within != "" {
				if _, err := parseDuration(step.WaitFor.Within); err != nil {
					return fmt.Errorf("steps[%d].waitFor.within must be a duration: %w", index, err)
				}
			}
		}
		if step.Expect != nil && step.Expect.File != "" {
			if filepath.IsAbs(step.Expect.File) {
				return fmt.Errorf("steps[%d].expect.file must be relative: %q", index, step.Expect.File)
			}
			clean := filepath.Clean(step.Expect.File)
			if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
				return fmt.Errorf("steps[%d].expect.file must not escape the scenario directory: %q", index, step.Expect.File)
			}
		}
	}
	return nil
}

func parseDuration(value string) (time.Duration, error) {
	return time.ParseDuration(strings.TrimSpace(value))
}
