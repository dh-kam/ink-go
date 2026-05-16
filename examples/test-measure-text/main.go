package main

import (
	"fmt"

	"github.com/dh-kam/ink-go/pkg/ink"
)

func main() {
	dimensions := ink.MeasureText("constructor")
	fmt.Printf("Width: %d\n", dimensions.Width)
}
