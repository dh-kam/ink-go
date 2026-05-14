package tuitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const RuntimeManifestSchemaVersion = "goink.tuitest.runtime/v1alpha1"

type RuntimeManifest struct {
	SchemaVersion string                    `json:"schemaVersion" yaml:"schemaVersion"`
	Runtimes      map[string]RuntimeBinding `json:"runtimes" yaml:"runtimes"`
}

type RuntimeBinding struct {
	Environment Environment            `json:"environment,omitempty" yaml:"environment,omitempty"`
	Apps        map[string]CommandSpec `json:"apps" yaml:"apps"`
}

type CommandSpec struct {
	Command     string      `json:"command" yaml:"command"`
	Args        []string    `json:"args,omitempty" yaml:"args,omitempty"`
	Cwd         string      `json:"cwd,omitempty" yaml:"cwd,omitempty"`
	Environment Environment `json:"environment,omitempty" yaml:"environment,omitempty"`
}

func LoadRuntimeManifestFile(path string) (*RuntimeManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read runtime manifest %q: %w", path, err)
	}

	manifest := &RuntimeManifest{}
	switch ext := filepath.Ext(path); ext {
	case ".json":
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(manifest); err != nil {
			return nil, fmt.Errorf("decode JSON runtime manifest %q: %w", path, err)
		}
	case ".yaml", ".yml":
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		if err := decoder.Decode(manifest); err != nil {
			return nil, fmt.Errorf("decode YAML runtime manifest %q: %w", path, err)
		}
	default:
		return nil, fmt.Errorf("unsupported runtime manifest extension %q", ext)
	}

	if err := manifest.validate(); err != nil {
		return nil, fmt.Errorf("validate runtime manifest %q: %w", path, err)
	}

	return manifest, nil
}

func (manifest *RuntimeManifest) Binding(runtime string) (RuntimeBinding, bool) {
	if manifest == nil {
		return RuntimeBinding{}, false
	}
	binding, ok := manifest.Runtimes[runtime]
	return binding, ok
}

func (manifest *RuntimeManifest) validate() error {
	if manifest.SchemaVersion == "" {
		return fmt.Errorf("schemaVersion is required")
	}
	if manifest.SchemaVersion != RuntimeManifestSchemaVersion {
		return fmt.Errorf("unsupported schemaVersion %q", manifest.SchemaVersion)
	}
	if len(manifest.Runtimes) == 0 {
		return fmt.Errorf("runtimes must not be empty")
	}
	for runtimeName, runtime := range manifest.Runtimes {
		if runtimeName == "" {
			return fmt.Errorf("runtime name must not be empty")
		}
		if len(runtime.Apps) == 0 {
			return fmt.Errorf("runtime %q apps must not be empty", runtimeName)
		}
		for appName, app := range runtime.Apps {
			if appName == "" {
				return fmt.Errorf("runtime %q app name must not be empty", runtimeName)
			}
			if app.Command == "" {
				return fmt.Errorf("runtime %q app %q command is required", runtimeName, appName)
			}
			if filepath.IsAbs(app.Cwd) {
				return fmt.Errorf("runtime %q app %q cwd must be relative: %q", runtimeName, appName, app.Cwd)
			}
		}
	}
	return nil
}
