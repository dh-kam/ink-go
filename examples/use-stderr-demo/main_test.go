package main

import (
	"testing"

	"github.com/dh-kam/goink.go/internal/tuitest"
)

func TestUseStderrTwoWritesScenario(t *testing.T) {
	tuitest.RunScenarioFile(t, "testdata/use-stderr-two-writes.scenario.yaml", tuitest.Components{
		"use-stderr-demo": UseStderrDemo,
	})
}
