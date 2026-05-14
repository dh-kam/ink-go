package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dh-kam/goink.go/internal/ttyinput"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

var transitionWords = []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}

func UseTransitionDemo() *vdom.Node {
	queryRaw, setQuery := ink.UseState("")
	deferredQueryRaw, setDeferredQuery := ink.UseState("")
	isPending, startTransition := ink.UseTransition()

	query := queryRaw.(string)
	deferredQuery := deferredQueryRaw.(string)

	ink.UseInput(func(input string, key ink.InputKey) {
		switch {
		case key.Backspace || key.Delete:
			next := trimLastRune(query)
			setQuery(next)
			startTransition(func() {
				setDeferredQuery(next)
			})
		case input != "" && !key.Ctrl && !key.Meta:
			next := query + input
			setQuery(next)
			startTransition(func() {
				setDeferredQuery(next)
			})
		}
	})

	searchValue := query
	if searchValue == "" {
		searchValue = "(type something)"
	}

	resultQualifier := "(showing first 10)"
	if deferredQuery != "" {
		resultQualifier = fmt.Sprintf("for %q", deferredQuery)
	}

	filteredItemsRaw := ink.UseMemo(func() interface{} {
		return generateTransitionItems(deferredQuery)
	}, []interface{}{deferredQuery})
	filteredItems, _ := filteredItemsRaw.([]string)
	searchChildren := []*vdom.Node{
		ink.Text("Search: "),
		ink.Text(vdom.Props{"color": "cyan"}, searchValue),
	}
	if isPending {
		searchChildren = append(searchChildren, ink.Text(vdom.Props{"color": "yellow"}, " (updating...)"))
	}

	children := []*vdom.Node{
		ink.Text(vdom.Props{"bold": true, "underline": true}, "useTransition Demo"),
		ink.Text(vdom.Props{"dimColor": true}, "(Type to search - input stays responsive while list updates)"),
		ink.Box(vdom.Props{"marginTop": 1}, searchChildren...),
		ink.Box(vdom.Props{"marginTop": 1, "flexDirection": "column"},
			buildTransitionResultNodes(resultQualifier, filteredItems, isPending)...,
		),
		ink.Box(vdom.Props{"marginTop": 1},
			ink.Text(vdom.Props{"dimColor": true}, "Press Ctrl+C to exit"),
		),
	}

	return ink.Box(vdom.Props{"flexDirection": "column"}, children...)
}

func buildTransitionResultNodes(resultQualifier string, items []string, isPending bool) []*vdom.Node {
	nodes := []*vdom.Node{
		ink.Text(vdom.Props{"bold": true}, fmt.Sprintf("Results %s:", resultQualifier)),
	}

	if len(items) == 0 {
		return append(nodes, ink.Text(vdom.Props{"dimColor": true}, " No items found"))
	}

	for _, item := range items {
		nodes = append(nodes, ink.Text(vdom.Props{"dimColor": isPending}, item))
	}

	return nodes
}

func generateTransitionItems(filter string) []string {
	items := make([]string, 0, 200)
	for i := 0; i < 200; i++ {
		items = append(items, fmt.Sprintf("Item %d: %s", i+1, transitionWords[i%len(transitionWords)]))
	}

	if filter == "" {
		return firstTransitionItems(items, 10)
	}

	start := time.Now()
	for time.Since(start) < 100*time.Millisecond {
	}

	lowerFilter := strings.ToLower(filter)
	filtered := make([]string, 0, 10)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), lowerFilter) {
			filtered = append(filtered, item)
			if len(filtered) == 10 {
				break
			}
		}
	}

	return filtered
}

func firstTransitionItems(items []string, limit int) []string {
	if len(items) < limit {
		return items
	}

	return items[:limit]
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}

	return string(runes[:len(runes)-1])
}

func main() {
	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(UseTransitionDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(UseTransitionDemo, ink.RenderOptions{})
	if err != nil {
		panic(err)
	}

	if err := runInputLoop(instance); err != nil {
		fmt.Println(err)
	}
}

func runInputLoop(instance *ink.Instance) error {
	return ttyinput.Run(os.Stdin, instance.HandleInput, func(input string) bool {
		return strings.Contains(input, "\x03")
	})
}
