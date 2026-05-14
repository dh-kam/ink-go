package main

import (
	"testing"

	"github.com/dh-kam/ink-go/internal/tuitest"
)

func TestBordersDemoScenario(t *testing.T) {
	tuitest.RunStaticScenarioFile(t, "testdata/borders.scenario.yaml", tuitest.Renderers{
		"border-demo": func() string {
			return renderBordersDemo() + "\n"
		},
	})
}
