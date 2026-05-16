package components_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// renderText concatenates the text content of a vdom subtree for assertion.
func renderText(node *vdom.Node) string {
	if node == nil {
		return ""
	}
	return node.TextContent()
}

func TestErrorBoundary_NoPanicReturnsChildren(t *testing.T) {
	want := vdom.CreateTextNode("hello-children")
	got := components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node { return want },
	})
	if got != want {
		t.Fatalf("expected children passthrough, got %v", got)
	}
}

func TestErrorBoundary_PanicTriggersDefaultFallback(t *testing.T) {
	got := components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node {
			panic(errors.New("boom"))
		},
	})
	if got == nil {
		t.Fatal("expected fallback node, got nil")
	}

	rendered := renderText(got)
	if !strings.Contains(rendered, "Error") {
		t.Fatalf("default fallback must contain 'Error', got %q", rendered)
	}
	if !strings.Contains(rendered, "boom") {
		t.Fatalf("default fallback must contain panic message, got %q", rendered)
	}
}

func TestErrorBoundary_OnErrorCallbackInvoked(t *testing.T) {
	var captured components.ErrorInfo
	called := 0

	components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node {
			panic(errors.New("kaboom"))
		},
		OnError: func(info components.ErrorInfo) {
			captured = info
			called++
		},
	})

	if called != 1 {
		t.Fatalf("OnError should run exactly once, got %d", called)
	}
	if captured.Err == nil || captured.Err.Error() != "kaboom" {
		t.Fatalf("captured Err mismatch: %v", captured.Err)
	}
	if captured.Stack == "" {
		t.Fatal("captured Stack must not be empty")
	}
	if !strings.Contains(captured.Stack, "goroutine") {
		t.Fatalf("Stack should look like a runtime stack, got %q", captured.Stack)
	}
}

func TestErrorBoundary_DefaultFallbackContainsErrorAndMessage(t *testing.T) {
	got := components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node {
			panic(errors.New("the-message"))
		},
	})

	rendered := renderText(got)
	if !strings.Contains(rendered, "Error") {
		t.Fatalf("default fallback missing 'Error': %q", rendered)
	}
	if !strings.Contains(rendered, "the-message") {
		t.Fatalf("default fallback missing message: %q", rendered)
	}
}

func TestErrorBoundary_DefaultFallbackRendersThroughApp(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return components.ErrorBoundary(components.ErrorBoundaryProps{
			Render: func() *vdom.Node {
				panic(errors.New("render failure"))
			},
		})
	})

	output := app.RenderOnce()
	if !strings.Contains(output, "ERROR") {
		t.Fatalf("rendered fallback missing ERROR header: %q", output)
	}
	if !strings.Contains(output, "render failure") {
		t.Fatalf("rendered fallback missing panic message: %q", output)
	}
}

func TestErrorBoundary_CustomFallbackUsed(t *testing.T) {
	got := components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node {
			panic("oh-no")
		},
		Fallback: func(info components.ErrorInfo) *vdom.Node {
			return vdom.CreateTextNode("custom:" + info.Err.Error())
		},
	})

	rendered := renderText(got)
	if rendered != "custom:oh-no" {
		t.Fatalf("custom fallback not used, got %q", rendered)
	}
}

func TestErrorBoundary_NilRenderIsSafe(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil Render should not panic, recovered %v", r)
		}
	}()

	got := components.ErrorBoundary(components.ErrorBoundaryProps{Render: nil})
	if got == nil {
		t.Fatal("nil Render should still produce a node")
	}
}

func TestErrorBoundary_PanicWithStringValue(t *testing.T) {
	var captured components.ErrorInfo
	components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node {
			panic("string-panic")
		},
		OnError: func(info components.ErrorInfo) {
			captured = info
		},
	})

	if captured.Err == nil {
		t.Fatal("string panic must be normalized to error")
	}
	if captured.Err.Error() != "string-panic" {
		t.Fatalf("string panic error mismatch: %q", captured.Err.Error())
	}
}

func TestErrorBoundary_PanicWithErrorValue(t *testing.T) {
	original := errors.New("real-error")
	var captured components.ErrorInfo

	components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node { panic(original) },
		OnError: func(info components.ErrorInfo) {
			captured = info
		},
	})

	if captured.Err == nil {
		t.Fatal("error panic must be captured")
	}
	if !errors.Is(captured.Err, original) {
		t.Fatalf("error panic should preserve identity, got %v", captured.Err)
	}
}

func TestErrorBoundary_PanicWithNonStringNonErrorValue(t *testing.T) {
	var captured components.ErrorInfo
	components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node { panic(42) },
		OnError: func(info components.ErrorInfo) {
			captured = info
		},
	})

	if captured.Err == nil {
		t.Fatal("int panic must be normalized to error")
	}
	if !strings.Contains(captured.Err.Error(), "42") {
		t.Fatalf("int panic should contain value, got %q", captured.Err.Error())
	}
}

func TestErrorBoundary_OnErrorAndCustomFallbackBothFire(t *testing.T) {
	called := false
	got := components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node { panic("multi") },
		OnError: func(info components.ErrorInfo) {
			called = true
		},
		Fallback: func(info components.ErrorInfo) *vdom.Node {
			return vdom.CreateTextNode("fb:" + info.Err.Error())
		},
	})
	if !called {
		t.Fatal("OnError must be invoked even when custom Fallback is provided")
	}
	if renderText(got) != "fb:multi" {
		t.Fatalf("custom fallback content unexpected: %q", renderText(got))
	}
}
