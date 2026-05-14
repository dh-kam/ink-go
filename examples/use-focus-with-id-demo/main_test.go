package main

import (
	"testing"

	"github.com/dh-kam/goink.go/internal/tuitest"
)

func TestUseFocusWithIDNavigationScenario(t *testing.T) {
	tuitest.RunScenarioFile(t, "testdata/use-focus-with-id-navigation.scenario.yaml", tuitest.Components{
		"use-focus-with-id-demo": UseFocusWithIDDemo,
	})
}
