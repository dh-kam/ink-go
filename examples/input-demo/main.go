package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dh-kam/goink.go/pkg/input"
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
	defer oldState.Restore()

	// Clear screen and hide cursor
	terminal.ClearScreen()
	terminal.HideCursor()
	defer terminal.ShowCursor()

	// Print instructions
	terminal.MoveCursor(1, 1)
	fmt.Println("=== Input Demo ===")
	fmt.Println("Press keys to see them displayed")
	fmt.Println("Press 'q' or ESC to quit")
	fmt.Println()
	terminal.MoveCursor(6, 1)

	// Create input handler
	handler := input.NewInputHandler(os.Stdin)

	// Setup signal handling
	sigChan := terminal.SetupSignalHandler()
	defer close(sigChan)

	// Track key press count
	keyCount := 0
	lastKeys := make([]string, 0, 5)

	// Main loop
	running := true
	for running {
		select {
		case sig := <-sigChan:
			fmt.Printf("\n\nReceived signal: %v\n", sig)
			running = false

		default:
			// Non-blocking read with timeout
			key, err := handler.ReadKey()
			if err != nil {
				// EOF or no input
				time.Sleep(10 * time.Millisecond)
				continue
			}

			keyCount++
			keyName := key.String()

			// Keep last 5 keys
			lastKeys = append(lastKeys, keyName)
			if len(lastKeys) > 5 {
				lastKeys = lastKeys[1:]
			}

			// Clear and redraw status
			terminal.MoveCursor(6, 1)
			fmt.Printf("\x1b[K") // Clear line
			fmt.Printf("Key #%d: %s\n", keyCount, keyName)

			terminal.MoveCursor(7, 1)
			fmt.Printf("\x1b[K")
			fmt.Printf("Char: %c (0x%02x)\n", key.Char, key.Char)

			terminal.MoveCursor(8, 1)
			fmt.Printf("\x1b[K")
			fmt.Printf("Name: %s\n", key.Name)

			terminal.MoveCursor(9, 1)
			fmt.Printf("\x1b[K")
			fmt.Print("Recent keys: ")
			for i, k := range lastKeys {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(k)
			}

			// Check for quit
			if key.Char == 'q' || key.Name == input.KeyEscape {
				running = false
			}
		}
	}

	// Cleanup
	terminal.MoveCursor(12, 1)
	terminal.ShowCursor()
	fmt.Println("\nGoodbye!")
}
