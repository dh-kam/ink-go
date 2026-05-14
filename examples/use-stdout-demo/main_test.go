package main

import (
	"testing"

	"github.com/dh-kam/ink-go/internal/tuitest"
)

func TestUseStdoutTwoWritesScenario(t *testing.T) {
	tuitest.RunScenarioFile(t, "testdata/use-stdout-two-writes.scenario.yaml", tuitest.Components{
		"use-stdout-demo": UseStdoutDemo,
	})
}
