package main

import (
	"testing"

	"github.com/dh-kam/ink-go/internal/tuitest"
)

func TestJustifyContentDemoScenario(t *testing.T) {
	tuitest.RunStaticScenarioFile(t, "testdata/justify-content.scenario.yaml", tuitest.Renderers{
		"justify-content-demo": renderJustifyContentDemo,
	})
}
