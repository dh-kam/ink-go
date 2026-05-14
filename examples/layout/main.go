package main

import (
"fmt"

"github.com/dh-kam/ink-go/pkg/layout"
)

func main() {
// Create a simple flexbox layout
container := layout.NewNode()
container.SetFlexDirection(layout.FlexDirectionRow)
container.SetWidth(300)
container.SetHeight(100)
container.SetJustifyContent(layout.JustifySpaceBetween)
container.SetPadding(layout.EdgeAll, 10)

// Add three boxes
box1 := layout.NewNode()
box1.SetWidth(80)
box1.SetHeight(60)

box2 := layout.NewNode()
box2.SetWidth(80)
box2.SetHeight(60)

box3 := layout.NewNode()
box3.SetWidth(80)
box3.SetHeight(60)

container.AddChild(box1)
container.AddChild(box2)
container.AddChild(box3)

// Calculate layout
container.CalculateLayout()

// Print results
fmt.Println("Flexbox Layout Demo (Row with SpaceBetween)")
fmt.Println("===========================================")
fmt.Printf("Container: %.0fx%.0f\n", container.GetComputedWidth(), container.GetComputedHeight())
fmt.Println()

for i := 0; i < container.GetChildCount(); i++ {
child := container.GetChild(i)
fmt.Printf("Box %d: position=(%.0f, %.0f) size=(%.0fx%.0f)\n",
i+1,
child.GetComputedLeft(),
child.GetComputedTop(),
child.GetComputedWidth(),
child.GetComputedHeight(),
)
}

fmt.Println("\n✅ Pure Go Flexbox layout working!")
}
