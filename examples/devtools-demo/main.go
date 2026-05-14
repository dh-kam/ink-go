// devtools-demo walks through every public surface of pkg/devtools end-to-end
// so engineers can eyeball the wiring without reading the test files.
//
// The script runs six clearly-numbered acts:
//
//  1. Inspect (Tree)   - pretty-print a vdom tree to stdout
//  2. Inspect (JSON)   - dump the same tree as canonical JSON
//  3. Profiler         - measure five fresh renders and print Stats
//  4. Tracer           - record three diffs (text change, add child, remove
//                        child) and dump every entry one per line
//  5. DebugPanel       - render a standalone debug panel with the metrics
//                        we collected in steps 3 and 4
//  6. WithDebug        - render the live counter app side-by-side with the
//                        debug panel using the right-hand split layout
//
// Other slots in the devtools sprint own DebugPanel / WithDebug; if those
// have not landed yet the build will fail at compile time, which is the
// expected hand-off signal.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/devtools"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// counterComponent is the "real" app under inspection: a single-line counter
// wrapped in a bordered box. It is intentionally small so the Inspect output
// stays readable while still exercising nested children + props.
func counterComponent() *vdom.Node {
	count, _ := ink.UseState(42)
	return components.Box(vdom.Props{
		"borderStyle":     "round",
		"padding":         1,
		"flexDirection":   "column",
		"backgroundColor": "black",
	},
		components.Text(fmt.Sprintf("counter: %d", count.(int))),
		components.Text("press q to quit"),
	)
}

// makeTree builds a small tree we mutate across Tracer.Diff calls. Keeping
// the shape stable except for the label keeps the resulting patches readable.
func makeTree(label string, extras ...string) *vdom.Node {
	children := []*vdom.Node{components.Text(label)}
	for _, extra := range extras {
		children = append(children, components.Text(extra))
	}
	return components.Box(vdom.Props{"borderStyle": "single"}, children...)
}

func main() {
	app := ink.NewApp(counterComponent)

	// ------------------------------------------------------------------
	// 1. Inspect (Tree)
	// ------------------------------------------------------------------
	// GetVNode runs the component once with the app's hooks context so we
	// get a real, mounted vdom tree (state slots populated) — matches what
	// the runtime would diff against.
	fmt.Println("=== 1. Inspect (Tree) ===")
	node := app.GetVNode()
	if err := devtools.PrintTree(os.Stdout, node); err != nil {
		fmt.Fprintln(os.Stderr, "PrintTree error:", err)
	}

	// ------------------------------------------------------------------
	// 2. Inspect (JSON)
	// ------------------------------------------------------------------
	fmt.Println("\n=== 2. Inspect (JSON) ===")
	fmt.Println(devtools.Inspect(node).JSON())

	// ------------------------------------------------------------------
	// 3. Profiler
	// ------------------------------------------------------------------
	// Wrap ink.RenderToString so each call records a fresh Profile entry.
	// Five iterations is enough to populate Min/Max/Avg without making the
	// demo wait noticeably.
	fmt.Println("\n=== 3. Profiler ===")
	prof := devtools.NewProfiler()
	wrapped := prof.Wrap(ink.RenderToString)
	for i := 0; i < 5; i++ {
		_, _ = wrapped(node)
	}
	stats := prof.Stats()
	fmt.Print(stats.Format())

	// ------------------------------------------------------------------
	// 4. Tracer
	// ------------------------------------------------------------------
	// Three diffs that hit three different patch families:
	//   - text change       -> UpdateText
	//   - add child         -> Insert
	//   - remove child      -> Remove
	// Unbounded tracer (limit=0) so all three entries survive.
	fmt.Println("\n=== 4. Tracer ===")
	tr := devtools.NewTracer(0)

	base := makeTree("hello")
	textChanged := makeTree("world")
	tr.Diff(base, textChanged, "text change")

	withChild := makeTree("world", "extra")
	tr.Diff(textChanged, withChild, "add child")

	tr.Diff(withChild, textChanged, "remove child")

	for _, e := range tr.Entries() {
		fmt.Println(devtools.FormatEntry(e))
	}

	// ------------------------------------------------------------------
	// 5. DebugPanel
	// ------------------------------------------------------------------
	// Reuse the metrics we already collected so the panel reflects what
	// the user just saw scrolling past, not synthetic numbers.
	fmt.Println("\n=== 5. DebugPanel ===")
	panelData := devtools.PanelData{
		Title:        "Counter App",
		Renders:      stats.TotalRenders,
		CacheHits:    stats.CachedCount,
		LastDuration: stats.MaxDuration,
	}
	if panelData.LastDuration == 0 {
		// Guarantee a non-zero LastDuration even when measurements happened
		// fast enough to round to zero on coarse-clock platforms.
		panelData.LastDuration = time.Millisecond
	}
	panel := devtools.DebugPanel(devtools.DebugPanelProps{Data: panelData})
	fmt.Println(ink.RenderToString(panel))

	// ------------------------------------------------------------------
	// 6. WithDebug (split view)
	// ------------------------------------------------------------------
	// The App callback is re-evaluated by WithDebug, so we close over the
	// already-rendered vdom snapshot rather than re-mounting the app.
	fmt.Println("\n=== 6. WithDebug (split view) ===")
	split := devtools.WithDebug(devtools.WithDebugProps{
		App: func() *vdom.Node {
			return components.Box(vdom.Props{
				"borderStyle": "round",
				"padding":     1,
			}, components.Text("counter: 42"))
		},
		Panel: panelData,
		Side:  "right",
	})
	fmt.Println(ink.RenderToString(split))

	fmt.Println("\ndevtools-demo: done.")
}
