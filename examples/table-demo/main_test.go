package main

import (
	"testing"

	"github.com/dh-kam/ink-go/internal/tuitest"
	"github.com/dh-kam/ink-go/pkg/ink"
)

func TestTableScenario(t *testing.T) {
	tuitest.RunStaticScenarioFile(t, "testdata/table.scenario.yaml", tuitest.Renderers{
		"table-demo": func() string {
			app := ink.NewAppWithOptions(TableDemo, ink.AppOptions{Width: 100, Height: 30})
			return app.RenderOnce()
		},
	})
}
