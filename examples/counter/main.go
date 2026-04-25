package main

import (
	"fmt"
	"time"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// Counter component with state
func Counter() *vdom.Node {
	count, setCount := ink.UseState(0)

	// Auto-increment every render (for demo)
	// In a real app, this would be triggered by user input
	currentCount := count.(int)

	if currentCount < 5 {
		// Schedule next increment
		go func() {
			time.Sleep(500 * time.Millisecond)
			setCount(currentCount + 1)
		}()
	}

	return vdom.CreateElement("box", nil,
		components.Text(fmt.Sprintf("Counter: %d", currentCount)),
	)
}

func main() {
	app := ink.NewApp(Counter)

	// Render multiple times to show state updates
	for i := 0; i < 6; i++ {
		output := app.RenderOnce()
		fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
		fmt.Println(output)
		fmt.Println("\n(Auto-incrementing counter demo)")
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\n✅ Counter demo complete!")
}
