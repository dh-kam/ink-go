package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func main() {
	// Create a bordered box with content
	borderProps := components.BorderProps{
		Style:  components.BorderDouble,
		Top:    true,
		Bottom: true,
		Left:   true,
		Right:  true,
		Label:  "Border Demo",
	}

	props := vdom.Props{
		"padding": 1.0,
	}

	content := components.Text("This is a bordered box!\nWith multiple lines\nof content.")

	_ = components.Border(borderProps, props, content)

	// Print the border
	fmt.Println("\nBorder Component Demo:")
	fmt.Println("=====================")
	printBorder(borderProps, "This is a bordered box!")
	fmt.Println()

	// Show different border styles
	styles := []components.BorderStyle{
		components.BorderSingle,
		components.BorderDouble,
		components.BorderRounded,
		components.BorderBold,
	}

	for _, style := range styles {
		fmt.Printf("%s Border:\n", style)
		printBorder(components.BorderProps{Style: style, Top: true, Bottom: true, Left: true, Right: true}, string(style)+" Style")
		fmt.Println()
	}

	// Show partial borders
	fmt.Println("Partial Borders:")
	fmt.Println("Top and bottom only:")
	printBorder(components.BorderProps{Style: components.BorderSingle, Top: true, Bottom: true}, "Top & Bottom")
	fmt.Println()
	fmt.Println("Left and right only:")
	printBorder(components.BorderProps{Style: components.BorderSingle, Left: true, Right: true}, "Left & Right")
}

func printBorder(props components.BorderProps, text string) {
	style := props.Style
	top := props.Top
	bottom := props.Bottom
	left := props.Left
	right := props.Right

	var hBar, vBar, tl, tr, bl, br string

	switch style {
	case components.BorderSingle:
		hBar = "─"
		vBar = "│"
		tl = "┌"
		tr = "┐"
		bl = "└"
		br = "┘"
	case components.BorderDouble:
		hBar = "═"
		vBar = "║"
		tl = "╔"
		tr = "╗"
		bl = "╚"
		br = "╝"
	case components.BorderRounded:
		hBar = "─"
		vBar = "│"
		tl = "╭"
		tr = "╮"
		bl = "╰"
		// br = "╯"
		br = "┘" // Fallback for simpler terminals
	case components.BorderBold:
		hBar = "━"
		vBar = "┃"
		tl = "┏"
		tr = "┓"
		bl = "┗"
		br = "┛"
	default:
		hBar = "-"
		vBar = "|"
		tl = "+"
		tr = "+"
		bl = "+"
		br = "+"
	}

	// Simple border rendering
	width := len(text) + 4

	// Top border
	if top {
		fmt.Print(tl)
		for i := 0; i < width-2; i++ {
			fmt.Print(hBar)
		}
		fmt.Println(tr)
	} else {
		if left {
			fmt.Print(vBar)
		}
		for i := 0; i < width; i++ {
			fmt.Print(" ")
		}
		if right {
			fmt.Println(vBar)
		} else {
			fmt.Println()
		}
	}

	// Content with side borders
	if left {
		fmt.Print(vBar)
	}
	fmt.Print("  ")
	fmt.Print(text)
	fmt.Print("  ")
	if right {
		fmt.Println(vBar)
	} else {
		fmt.Println()
	}

	// Bottom border
	if bottom {
		fmt.Print(bl)
		for i := 0; i < width-2; i++ {
			fmt.Print(hBar)
		}
		fmt.Println(br)
	} else {
		if left {
			fmt.Print(vBar)
		}
		for i := 0; i < width; i++ {
			fmt.Print(" ")
		}
		if right {
			fmt.Println(vBar)
		} else {
			fmt.Println()
		}
	}
}
