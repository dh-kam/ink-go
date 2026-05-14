package main

import (
	"testing"

	"github.com/dh-kam/ink-go/internal/tuitest"
)

func TestChatMessagesScenario(t *testing.T) {
	nextMessageID = 0
	tuitest.RunScenarioFile(t, "testdata/chat-messages.scenario.yaml", tuitest.Components{
		"chat-demo": ChatDemo,
	})
}
