// error-demo demonstrates the ErrorBoundary component by rendering two
// subtrees: a normal one that succeeds, and a crashing one whose render
// function panics. The boundary converts the panic into a fallback UI and
// surfaces an ErrorInfo to the OnError callback so that the host process
// keeps running.
package main

import (
	"fmt"
	"os"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func main() {
	fmt.Println("--- normal render ---")
	fmt.Println(ink.RenderToString(safeApp()))

	fmt.Println("--- render with simulated panic ---")
	fmt.Println(ink.RenderToString(crashingApp()))

	fmt.Println("--- render with custom fallback ---")
	fmt.Println(ink.RenderToString(crashingAppWithFallback()))

	fmt.Println("(process survived all three renders)")
}

// safeApp renders a healthy subtree inside an ErrorBoundary. No panic is
// raised, so the boundary returns the wrapped tree untouched.
func safeApp() *vdom.Node {
	return components.ErrorBoundary(components.ErrorBoundaryProps{
		Render: func() *vdom.Node {
			return components.Box(
				vdom.Props{"flexDirection": "column", "padding": 1},
				components.Text("All good!"),
				components.Text("ErrorBoundary did not have to intervene."),
			)
		},
	})
}

// crashingApp renders a subtree whose builder panics. The boundary catches
// it, logs via OnError, and falls back to its built-in red bordered UI.
func crashingApp() *vdom.Node {
	return components.ErrorBoundary(components.ErrorBoundaryProps{
		OnError: func(info components.ErrorInfo) {
			fmt.Fprintln(os.Stderr, "error-demo caught panic:", info.Err)
		},
		Render: func() *vdom.Node {
			// divisionByZero stands in for any deeply nested logic that
			// might explode while the vdom is being built.
			divisionByZero()
			return doomedRender()
		},
	})
}

// crashingAppWithFallback shows that callers can provide a custom Fallback
// function that receives the ErrorInfo and produces any vdom they like.
func crashingAppWithFallback() *vdom.Node {
	return components.ErrorBoundary(components.ErrorBoundaryProps{
		OnError: func(info components.ErrorInfo) {
			fmt.Fprintln(os.Stderr, "error-demo (custom fallback) caught:", info.Err)
		},
		Fallback: func(info components.ErrorInfo) *vdom.Node {
			return components.Box(
				vdom.Props{"flexDirection": "column", "padding": 1},
				components.Text(fmt.Sprintf("custom fallback fired: %v", info.Err)),
				components.Text("the application keeps running."),
			)
		},
		Render: func() *vdom.Node {
			return doomedRender()
		},
	})
}

// divisionByZero deliberately triggers a runtime panic to mimic an
// unexpected failure deep inside a render tree.
func divisionByZero() {
	defer func() {
		// Swallow the runtime panic from the integer division and re-raise
		// a clearer message — this is what real call sites that wrap
		// low-level failures often do.
		if r := recover(); r != nil {
			panic(fmt.Sprintf("simulated render failure: %v", r))
		}
	}()

	a, b := 1, 0
	_ = a / b
}

// doomedRender is the canonical "this should never run" branch — if the
// preceding logic somehow returns, this still surfaces a panic so the
// boundary has something to catch.
func doomedRender() *vdom.Node {
	panic("simulated render failure")
}
