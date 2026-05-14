package main

import (
	"testing"

	"github.com/dh-kam/goink.go/internal/tuitest"
	"github.com/dh-kam/goink.go/pkg/ink"
)

func TestStaticCompleteScenario(t *testing.T) {
	tuitest.RunStaticScenarioFile(t, "testdata/static-complete.scenario.yaml", tuitest.Renderers{
		"static-demo": func() string {
			resetTests()
			for number := 1; number <= 10; number++ {
				appendTest(number)
			}
			app := ink.NewApp(StaticDemo)
			return app.RenderOnce()
		},
	})
}
