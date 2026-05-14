package main

import (
	"testing"

	"github.com/dh-kam/goink.go/internal/tuitest"
)

func TestBordersDemoScenario(t *testing.T) {
	tuitest.RunStaticScenarioFile(t, "testdata/borders.scenario.yaml", tuitest.Renderers{
		"border-demo": func() string {
			return renderBordersDemo() + "\n"
		},
	})
}
