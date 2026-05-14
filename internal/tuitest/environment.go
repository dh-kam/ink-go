package tuitest

import (
	"os"
	"sync"
	"testing"
)

var scenarioEnvironmentMu sync.Mutex

func applyScenarioEnvironment(t testing.TB, environment Environment) func() {
	t.Helper()
	if len(environment) == 0 {
		return func() {}
	}

	scenarioEnvironmentMu.Lock()
	previous := map[string]environmentValue{}
	for key, value := range environment {
		oldValue, existed := os.LookupEnv(key)
		previous[key] = environmentValue{value: oldValue, existed: existed}

		if value == nil {
			_ = os.Unsetenv(key)
			continue
		}
		_ = os.Setenv(key, *value)
	}

	return func() {
		defer scenarioEnvironmentMu.Unlock()
		for key, value := range previous {
			if value.existed {
				_ = os.Setenv(key, value.value)
				continue
			}
			_ = os.Unsetenv(key)
		}
	}
}

type environmentValue struct {
	value   string
	existed bool
}
