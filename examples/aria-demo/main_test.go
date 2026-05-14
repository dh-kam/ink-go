package main

import (
	"testing"

	"github.com/dh-kam/goink.go/internal/tuitest"
)

func TestAriaDemoToggleScenario(t *testing.T) {
	tuitest.RunScenarioFile(t, "testdata/aria-toggle.scenario.yaml", tuitest.Components{
		"aria-demo": AriaExample,
	})
}
