package main

import (
	"testing"

	"github.com/dh-kam/ink-go/internal/tuitest"
)

func TestSelectInputScenario(t *testing.T) {
	tuitest.RunScenarioFile(t, "testdata/select-input.scenario.yaml", tuitest.Components{
		"select-input-demo": SelectInput,
	})
}

func TestSelectInputWrapScenario(t *testing.T) {
	tuitest.RunScenarioFile(t, "testdata/select-input-wrap.scenario.yaml", tuitest.Components{
		"select-input-demo": SelectInput,
	})
}
