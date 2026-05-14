package components

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// ErrorInfo carries the recovered error and stack trace produced when an
// ErrorBoundary catches a panic from its protected render function.
//
// Err is always non-nil when surfaced through OnError or the Fallback render —
// non-error panic values (strings, ints, structs, etc.) are converted with
// errors.New / fmt.Errorf so callers can assume an error is available.
type ErrorInfo struct {
	// Err is the recovered error. If the original panic value was not an
	// error it is wrapped via fmt.Errorf("%v", v) so this field is always
	// populated when an error has been caught.
	Err error
	// Stack contains the goroutine stack captured at recovery time via
	// runtime.Stack. It is non-empty when a panic was caught.
	Stack string
}

// ErrorBoundaryProps configures an ErrorBoundary.
type ErrorBoundaryProps struct {
	// Render is invoked inside a defer/recover guard. Any panic raised while
	// constructing the protected subtree is converted to an ErrorInfo and
	// surfaced to OnError / Fallback instead of propagating.
	Render func() *vdom.Node
	// Fallback, when non-nil, produces the replacement subtree shown after a
	// panic. If nil, ErrorBoundary renders a default red bordered block with
	// the error message.
	Fallback func(info ErrorInfo) *vdom.Node
	// OnError, when non-nil, is invoked once per caught panic before the
	// fallback subtree is produced. Useful for logging / telemetry.
	OnError func(info ErrorInfo)
}

// ErrorBoundary is the upstream React error-boundary equivalent for ink-go.
//
// It wraps props.Render in a defer/recover guard. If the render function
// panics, ErrorBoundary:
//  1. Captures the recovered value and a runtime stack trace into ErrorInfo.
//  2. Invokes props.OnError with the captured info, when provided.
//  3. Returns props.Fallback(info) when provided, otherwise a default red
//     bordered block containing "Error" and the message.
//
// When props.Render is nil, an empty Box is returned without invoking any
// callbacks — this mirrors the "no children" case rather than treating the
// missing render as a runtime fault.
func ErrorBoundary(props ErrorBoundaryProps) *vdom.Node {
	if props.Render == nil {
		return Box(nil)
	}

	node, info, recovered := safeRender(props.Render)
	if !recovered {
		return node
	}

	if props.OnError != nil {
		props.OnError(info)
	}

	if props.Fallback != nil {
		return props.Fallback(info)
	}

	return defaultErrorFallback(info)
}

// safeRender invokes render inside a defer/recover guard and reports whether
// a panic was caught along with the resulting ErrorInfo.
func safeRender(render func() *vdom.Node) (node *vdom.Node, info ErrorInfo, recovered bool) {
	defer func() {
		if r := recover(); r != nil {
			recovered = true
			info = ErrorInfo{
				Err:   normalizePanic(r),
				Stack: captureStack(),
			}
			node = nil
		}
	}()

	node = render()
	return node, ErrorInfo{}, false
}

// normalizePanic converts an arbitrary panic value into an error.
func normalizePanic(v interface{}) error {
	if v == nil {
		return errors.New("panic: <nil>")
	}

	switch typed := v.(type) {
	case error:
		return typed
	case string:
		return errors.New(typed)
	default:
		return fmt.Errorf("%v", typed)
	}
}

// captureStack returns the current goroutine's stack trace.
func captureStack() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// defaultErrorFallback renders the built-in fallback UI: a single-bordered
// red box with a title line and the error message.
func defaultErrorFallback(info ErrorInfo) *vdom.Node {
	message := "<unknown>"
	if info.Err != nil {
		message = info.Err.Error()
	}

	title := styles.Colorize("Error", styles.Red, styles.Foreground)
	body := styles.Colorize("Error: "+message, styles.Red, styles.Foreground)

	column := vdom.CreateElement(
		"error-boundary-body",
		vdom.Props{"flexDirection": "column"},
		vdom.CreateTextNode(title),
		vdom.CreateTextNode(body),
	)

	return Border(
		BorderProps{Style: BorderSingle},
		vdom.Props{"borderColor": "red", "padding": 1},
		column,
	)
}
