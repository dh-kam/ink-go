package main

import (
	"testing"

	"github.com/dh-kam/goink.go/internal/tuitest"
)

func TestUseFocusNavigationScenario(t *testing.T) {
	tuitest.RunScenarioFile(t, "testdata/use-focus-navigation.scenario.yaml", tuitest.Components{
		"use-focus-demo": UseFocusDemo,
	})
}
