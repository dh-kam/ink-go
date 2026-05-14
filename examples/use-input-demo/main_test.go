package main

import (
	"testing"

	"github.com/dh-kam/ink-go/internal/tuitest"
)

func TestUseInputMaxThenQuitScenario(t *testing.T) {
	tuitest.RunScenarioFile(t, "testdata/use-input-max-then-quit.scenario.yaml", tuitest.Components{
		"use-input-demo": Robot,
	})
}
