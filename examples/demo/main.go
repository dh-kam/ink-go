package main

import (
	"fmt"
	"os"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/input"
	"github.com/dh-kam/ink-go/pkg/renderloop"
	sty "github.com/dh-kam/ink-go/pkg/styles"
	"github.com/dh-kam/ink-go/pkg/terminal"
)

func main() {
	// Check if we're in a terminal
	if !terminal.StdinIsTerminal() {
		fmt.Println("This demo requires a terminal")
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
	defer terminal.ShowCursor()

	// Create demo
	demo := NewDemo()

	// Create render loop
	renderFn := func() error {
		return demo.Render()
	}

	app := renderloop.NewApp(15, renderFn)

	// Start input handler in background
	inputDone := make(chan error, 1)
	go func() {
		inputDone <- demo.HandleInput()
	}()

	// Start app
	appDone := make(chan error, 1)
	go func() {
		appDone <- app.Start()
	}()

	// Wait for completion
	select {
	case <-appDone:
		// Normal exit
	case <-inputDone:
		app.Stop()
		<-appDone
	}
}

// Demo represents the demo application
type Demo struct {
	count       int
	selectedTab int
	running     bool
	handler     *input.InputHandler
}

// NewDemo creates a new demo app
func NewDemo() *Demo {
	return &Demo{
		count:   0,
		running: true,
	}
}

// HandleInput processes user input
func (d *Demo) HandleInput() error {
	d.handler = input.NewInputHandler(os.Stdin)

	for d.running {
		key, err := d.handler.ReadKey()
		if err != nil {
			return err
		}

		// Handle key press
		switch {
		case key.Char == 'q' || key.Name == input.KeyEscape:
			d.running = false
			return renderloop.ErrExitRequested

		case key.Char == '+', key.Char == '=':
			d.count++

		case key.Char == '-', key.Char == '_':
			if d.count > 0 {
				d.count--
			}

		case key.Name == input.KeyRight:
			if d.selectedTab < 4 {
				d.selectedTab++
			}

		case key.Name == input.KeyLeft:
			if d.selectedTab > 0 {
				d.selectedTab--
			}

		case key.Char == 'r':
			d.count = 0
		}
	}

	return nil
}

// Render renders the demo UI
func (d *Demo) Render() error {
	terminal.ClearScreen()
	terminal.MoveCursor(1, 1)

	d.renderHeader()
	d.renderTabs()
	d.renderTabContent()
	d.renderFooter()

	return nil
}

func (d *Demo) renderHeader() {
	title := sty.Bold(sty.Colorize("╔════════════════════════════════════════════════════════════╗", sty.Cyan, sty.Foreground))
	fmt.Println(title)

	subtitle := sty.Bold("║           Goink: React-like CLI Framework in Pure Go           ║")
	fmt.Println(subtitle)

	divider := sty.Bold("║" + sty.Colorize("──────────────────────────────────────────────────────────", sty.Cyan, sty.Foreground) + "║")
	fmt.Println(divider)

	bottom := sty.Bold("╚════════════════════════════════════════════════════════════╝")
	fmt.Println(bottom)
	fmt.Println()
}

func (d *Demo) renderTabs() {
	tabs := []string{"Widgets", "Styles", "Borders", "Colors", "Input"}

	for i, tab := range tabs {
		if i == d.selectedTab {
			fmt.Print(sty.Bold(sty.Colorize("["+tab+"] ", sty.Green, sty.Foreground)))
		} else {
			fmt.Print(" " + tab + "  ")
		}
	}
	fmt.Println()
	fmt.Println(sty.Colorize("─────────────────────────────────────────────────────", sty.Cyan, sty.Foreground))
	fmt.Println()
}

func (d *Demo) renderTabContent() {
	switch d.selectedTab {
	case 0:
		d.renderWidgetsTab()
	case 1:
		d.renderStylesTab()
	case 2:
		d.renderBordersTab()
	case 3:
		d.renderColorsTab()
	case 4:
		d.renderInputTab()
	}
}

func (d *Demo) renderWidgetsTab() {
	fmt.Println(sty.Bold("Box Components & Interactive Elements:"))
	fmt.Println()

	fmt.Println("┌────────────────────────────────────┐")
	fmt.Println("│ " + sty.Bold("Counter Demo:") + "                       │")
	fmt.Println("│                                    │")
	fmt.Printf("│     %s%d%s                            │\n", sty.Bold(sty.Colorize("Count: ", sty.Yellow, sty.Foreground)), d.count, sty.Reset())
	fmt.Println("│                                    │")
	fmt.Println("│     [+/-] Increment/Decrement      │")
	fmt.Println("│     [r]   Reset                     │")
	fmt.Println("│     [←/→] Switch tabs               │")
	fmt.Println("│     [q]   Quit                      │")
	fmt.Println("└────────────────────────────────────┘")

	fmt.Println()
	fmt.Println("Flex Layout Demo:")
	fmt.Println()
	fmt.Println("Row Layout:")
	fmt.Println("[Box 1] [Box 2] [Box 3]")
	fmt.Println()
	fmt.Println("Column Layout:")
	fmt.Println("┌───────┐")
	fmt.Println("│ Box 1 │")
	fmt.Println("├───────┤")
	fmt.Println("│ Box 2 │")
	fmt.Println("├───────┤")
	fmt.Println("│ Box 3 │")
	fmt.Println("└───────┘")
}

func (d *Demo) renderStylesTab() {
	fmt.Println(sty.Bold("Text Styles & Formatting:"))
	fmt.Println()

	textStyles := []struct {
		name string
		fn   func(string) string
	}{
		{"Bold", sty.Bold},
		{"Dim", sty.Dim},
		{"Italic", sty.Italic},
		{"Underline", sty.Underline},
		{"Strikethrough", sty.Strikethrough},
	}

	for _, s := range textStyles {
		fmt.Printf("  %s: %sHello World%s\n", s.name, s.fn(""), sty.Reset())
	}
	fmt.Println()

	fmt.Println("Combined Styles:")
	fmt.Println("  " + sty.Bold(sty.Colorize("Bold + Red", sty.Red, sty.Foreground)))
	fmt.Println("  " + sty.Italic(sty.Underline("Italic + Underline")))
	fmt.Println("  " + sty.Bold(sty.Dim(sty.Colorize("Bold + Dim + Blue", sty.Blue, sty.Foreground))))
}

func (d *Demo) renderBordersTab() {
	borderTypes := []struct {
		name  string
		style components.BorderStyle
	}{
		{"Single Border", components.BorderSingle},
		{"Double Border", components.BorderDouble},
		{"Rounded Border", components.BorderRounded},
		{"Bold Border", components.BorderBold},
	}

	for _, bt := range borderTypes {
		fmt.Println(bt.name + ":")
		printSimpleBorder(bt.style, bt.name)
		fmt.Println()
	}
}

func (d *Demo) renderColorsTab() {
	colors := []struct {
		name  string
		color sty.Color
	}{
		{"Red", sty.Red},
		{"Green", sty.Green},
		{"Yellow", sty.Yellow},
		{"Blue", sty.Blue},
		{"Magenta", sty.Magenta},
		{"Cyan", sty.Cyan},
		{"White", sty.White},
	}

	fmt.Println("Foreground Colors:")
	for _, c := range colors {
		fmt.Print("  " + sty.Colorize(c.name, c.color, sty.Foreground) + "  ")
	}
	fmt.Println()
	fmt.Println()

	fmt.Println("Background Colors:")
	for _, c := range colors {
		fmt.Print("  " + sty.Colorize(" "+c.name+" ", c.color, sty.Background) + " ")
	}
	fmt.Println()
	fmt.Println()

	fmt.Println("RGB Colors:")
	fmt.Println("  " + sty.Colorize("Custom Red", sty.RGB(255, 0, 0), sty.Foreground) + sty.Reset())
	fmt.Println("  " + sty.Colorize("Custom Green", sty.RGB(0, 255, 0), sty.Foreground) + sty.Reset())
	fmt.Println("  " + sty.Colorize("Custom Blue", sty.RGB(0, 0, 255), sty.Foreground) + sty.Reset())
}

func (d *Demo) renderInputTab() {
	fmt.Println(sty.Bold("Input Handling:"))
	fmt.Println()

	keys := []string{
		"↑ (Up Arrow)", "↓ (Down Arrow)", "← (Left Arrow)", "→ (Right Arrow)",
		"Enter/Return", "Tab", "Backspace", "Escape", "Ctrl+C",
	}

	fmt.Println("Supported Keys:")
	for _, key := range keys {
		fmt.Printf("  %s\n", sty.Colorize("• "+key, sty.Green, sty.Foreground))
	}

	fmt.Println()
	fmt.Println("Current State:")
	fmt.Printf("  Count: %d\n", d.count)
	fmt.Printf("  Tab: %d/5\n", d.selectedTab+1)
}

func (d *Demo) renderFooter() {
	fmt.Println()
	fmt.Println(sty.Colorize("─────────────────────────────────────────────────────", sty.Cyan, sty.Foreground))
	fmt.Printf("Count: %d | Tab: %d/5 | Running: %v | Press 'q' to quit\n",
		d.count, d.selectedTab+1, d.running)
}

// printSimpleBorder prints a simple border
func printSimpleBorder(style components.BorderStyle, text string) {
	var h, v, tl, tr, bl, br string

	switch style {
	case components.BorderSingle:
		h, v, tl, tr, bl, br = "─", "│", "┌", "┐", "└", "┘"
	case components.BorderDouble:
		h, v, tl, tr, bl, br = "═", "║", "╔", "╗", "╚", "╝"
	case components.BorderRounded:
		h, v, tl, tr, bl, br = "─", "│", "╭", "╮", "╰", "╯"
	case components.BorderBold:
		h, v, tl, tr, bl, br = "━", "┃", "┏", "┓", "┗", "┛"
	default:
		h, v, tl, tr, bl, br = "-", "|", "+", "+", "+", "+"
	}

	width := len(text) + 4
	fmt.Print(tl)
	for i := 0; i < width; i++ {
		fmt.Print(h)
	}
	fmt.Println(tr)

	fmt.Print(v)
	fmt.Print("  ")
	fmt.Print(text)
	fmt.Print("  ")
	fmt.Println(v)

	fmt.Print(bl)
	for i := 0; i < width; i++ {
		fmt.Print(h)
	}
	fmt.Println(br)
}
