package main

import (
	"testing"

	"github.com/dh-kam/goink.go/internal/tuitest"
)

func TestBoxBackgroundsDemoScenario(t *testing.T) {
	tuitest.RunStaticScenarioFile(t, "testdata/box-backgrounds.scenario.yaml", tuitest.Renderers{
		"box-backgrounds-demo": renderBoxBackgroundsDemo,
	})
}
