package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dh-kam/goink.go/pkg/hooks"
	"github.com/dh-kam/goink.go/pkg/renderloop"
	"github.com/dh-kam/goink.go/pkg/terminal"
)

func main() {
	// Check if we're in a terminal
	if !terminal.StdinIsTerminal() {
		fmt.Println("This example requires a terminal")
		os.Exit(1)
	}

	// Setup terminal
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("Failed to set raw mode: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		oldState.Restore()
		terminal.ShowCursor()
		terminal.ClearScreen()
	}()

	// Clear screen and hide cursor
	terminal.ClearScreen()
	terminal.HideCursor()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create hooks context for state management
	hooksCtx := hooks.NewContext()

	// Create counter state
	count, setCount := hooks.UseState(hooksCtx, 0)
	currentCount := count.(int)

	// Create render function
	renderFn := func() error {
		// Clear screen
		terminal.ClearScreen()
		terminal.MoveCursor(1, 1)

		// Render UI
		fmt.Println("=== Live Counter Demo ===")
		fmt.Println()
		fmt.Printf("Count: %d\n", currentCount)
		fmt.Println()
		fmt.Println("Counter is incrementing automatically...")
		fmt.Println("Press Ctrl+C to exit")

		// Increment counter
		currentCount++
		setCount(currentCount)

		return nil
	}

	// Create render loop app (10 FPS)
	app := renderloop.NewApp(10, renderFn)

	// Start app in background
	appDone := make(chan error, 1)
	go func() {
		appDone <- app.Start()
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Println("\n\nReceived interrupt signal")
		app.Stop()
		<-appDone
	case err := <-appDone:
		if err != nil && err != context.Canceled {
			fmt.Printf("\n\nRender loop error: %v\n", err)
		}
	}

	// Cleanup
	terminal.MoveCursor(10, 1)
	terminal.ShowCursor()
	fmt.Println("\nGoodbye!")
	time.Sleep(100 * time.Millisecond)
}
