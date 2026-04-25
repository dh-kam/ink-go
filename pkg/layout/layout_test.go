package layout_test

import (
	"testing"

	"github.com/dh-kam/goink.go/pkg/layout"
)

// TestNodeCreation tests creating a layout node
func TestNodeCreation(t *testing.T) {
	node := layout.NewNode()

	if node == nil {
		t.Fatal("NewNode should return a valid node")
	}
}

// TestNodeDimensions tests setting and getting dimensions
func TestNodeDimensions(t *testing.T) {
	node := layout.NewNode()

	node.SetWidth(100)
	node.SetHeight(50)

	if node.GetComputedWidth() != 100 {
		t.Errorf("Expected width 100, got %f", node.GetComputedWidth())
	}

	if node.GetComputedHeight() != 50 {
		t.Errorf("Expected height 50, got %f", node.GetComputedHeight())
	}
}

// TestNodeFlexDirection tests flex direction
func TestNodeFlexDirection(t *testing.T) {
	node := layout.NewNode()

	node.SetFlexDirection(layout.FlexDirectionRow)
	if node.GetFlexDirection() != layout.FlexDirectionRow {
		t.Error("Expected FlexDirectionRow")
	}

	node.SetFlexDirection(layout.FlexDirectionColumn)
	if node.GetFlexDirection() != layout.FlexDirectionColumn {
		t.Error("Expected FlexDirectionColumn")
	}
}

func TestNodeFlexGrow(t *testing.T) {
	node := layout.NewNode()
	node.SetFlexGrow(2)

	if node.GetFlexGrow() != 2 {
		t.Errorf("Expected flex grow 2, got %f", node.GetFlexGrow())
	}
}

// TestNodeChildren tests adding children
func TestNodeChildren(t *testing.T) {
	parent := layout.NewNode()
	child1 := layout.NewNode()
	child2 := layout.NewNode()

	parent.AddChild(child1)
	parent.AddChild(child2)

	if parent.GetChildCount() != 2 {
		t.Errorf("Expected 2 children, got %d", parent.GetChildCount())
	}

	if parent.GetChild(0) != child1 {
		t.Error("First child mismatch")
	}

	if parent.GetChild(1) != child2 {
		t.Error("Second child mismatch")
	}
}

// TestSimpleRowLayout tests simple row layout
func TestSimpleRowLayout(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(300)
	parent.SetHeight(100)

	child1 := layout.NewNode()
	child1.SetWidth(100)
	child1.SetHeight(50)

	child2 := layout.NewNode()
	child2.SetWidth(150)
	child2.SetHeight(50)

	parent.AddChild(child1)
	parent.AddChild(child2)

	// Calculate layout
	parent.CalculateLayout()

	// In row layout, children should be positioned horizontally
	if child1.GetComputedLeft() != 0 {
		t.Errorf("Expected child1 left=0, got %f", child1.GetComputedLeft())
	}

	if child2.GetComputedLeft() != 100 {
		t.Errorf("Expected child2 left=100, got %f", child2.GetComputedLeft())
	}
}

// TestSimpleColumnLayout tests simple column layout
func TestSimpleColumnLayout(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(100)
	parent.SetHeight(300)

	child1 := layout.NewNode()
	child1.SetWidth(100)
	child1.SetHeight(50)

	child2 := layout.NewNode()
	child2.SetWidth(100)
	child2.SetHeight(80)

	parent.AddChild(child1)
	parent.AddChild(child2)

	parent.CalculateLayout()

	// In column layout, children should be positioned vertically
	if child1.GetComputedTop() != 0 {
		t.Errorf("Expected child1 top=0, got %f", child1.GetComputedTop())
	}

	if child2.GetComputedTop() != 50 {
		t.Errorf("Expected child2 top=50, got %f", child2.GetComputedTop())
	}
}

func TestRowFlexGrowLayout(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(10)
	parent.SetHeight(1)

	left := layout.NewNode()
	left.SetWidth(2)
	left.SetHeight(1)

	spacer := layout.NewNode()
	spacer.SetHeight(1)
	spacer.SetFlexGrow(1)

	right := layout.NewNode()
	right.SetWidth(2)
	right.SetHeight(1)

	parent.AddChild(left)
	parent.AddChild(spacer)
	parent.AddChild(right)

	parent.CalculateLayout()

	if spacer.GetComputedWidth() != 6 {
		t.Errorf("Expected spacer width 6, got %f", spacer.GetComputedWidth())
	}

	if right.GetComputedLeft() != 8 {
		t.Errorf("Expected right child at x=8, got %f", right.GetComputedLeft())
	}
}

func TestColumnFlexGrowLayout(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(1)
	parent.SetHeight(10)

	top := layout.NewNode()
	top.SetWidth(1)
	top.SetHeight(2)

	spacer := layout.NewNode()
	spacer.SetWidth(1)
	spacer.SetFlexGrow(1)

	bottom := layout.NewNode()
	bottom.SetWidth(1)
	bottom.SetHeight(2)

	parent.AddChild(top)
	parent.AddChild(spacer)
	parent.AddChild(bottom)

	parent.CalculateLayout()

	if spacer.GetComputedHeight() != 6 {
		t.Errorf("Expected spacer height 6, got %f", spacer.GetComputedHeight())
	}

	if bottom.GetComputedTop() != 8 {
		t.Errorf("Expected bottom child at y=8, got %f", bottom.GetComputedTop())
	}
}

func TestColumnGapOverridesMainAxisGapInRowLayout(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(10)
	parent.SetHeight(1)
	parent.SetGap(1)
	parent.SetColumnGap(2)

	left := layout.NewNode()
	left.SetWidth(1)
	left.SetHeight(1)

	right := layout.NewNode()
	right.SetWidth(1)
	right.SetHeight(1)

	parent.AddChild(left)
	parent.AddChild(right)
	parent.CalculateLayout()

	if right.GetComputedLeft() != 3 {
		t.Fatalf("expected second child at x=3 with columnGap override, got %f", right.GetComputedLeft())
	}
}

func TestRowGapOverridesMainAxisGapInColumnLayout(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(1)
	parent.SetHeight(10)
	parent.SetGap(1)
	parent.SetRowGap(2)

	top := layout.NewNode()
	top.SetWidth(1)
	top.SetHeight(1)

	bottom := layout.NewNode()
	bottom.SetWidth(1)
	bottom.SetHeight(1)

	parent.AddChild(top)
	parent.AddChild(bottom)
	parent.CalculateLayout()

	if bottom.GetComputedTop() != 3 {
		t.Fatalf("expected second child at y=3 with rowGap override, got %f", bottom.GetComputedTop())
	}
}

func TestRowGapAppliesBetweenWrappedRowLines(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(3)
	parent.SetWrapMode(layout.WrapWrap)
	parent.SetGap(1)
	parent.SetRowGap(2)

	for index := 0; index < 3; index++ {
		child := layout.NewNode()
		child.SetWidth(1)
		child.SetHeight(1)
		parent.AddChild(child)
	}

	parent.CalculateLayout()

	if got := parent.GetChild(2).GetComputedTop(); got != 3 {
		t.Fatalf("expected wrapped second row to start at y=3, got %f", got)
	}
}

func TestColumnGapAppliesBetweenWrappedColumnLines(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetHeight(2)
	parent.SetWrapMode(layout.WrapWrap)
	parent.SetColumnGap(2)

	for index := 0; index < 3; index++ {
		child := layout.NewNode()
		child.SetWidth(1)
		child.SetHeight(1)
		parent.AddChild(child)
	}

	parent.CalculateLayout()

	if got := parent.GetChild(2).GetComputedLeft(); got != 3 {
		t.Fatalf("expected wrapped second column to start at x=3, got %f", got)
	}
}

// TestPadding tests padding
func TestPadding(t *testing.T) {
	node := layout.NewNode()
	node.SetWidth(100)
	node.SetHeight(100)
	node.SetPadding(layout.EdgeAll, 10)

	child := layout.NewNode()
	child.SetWidth(50)
	child.SetHeight(50)

	node.AddChild(child)
	node.CalculateLayout()

	// Child should be positioned with padding offset
	if child.GetComputedLeft() != 10 {
		t.Errorf("Expected child left=10 (padding), got %f", child.GetComputedLeft())
	}

	if child.GetComputedTop() != 10 {
		t.Errorf("Expected child top=10 (padding), got %f", child.GetComputedTop())
	}
}

// TestJustifyContent tests justify content
func TestJustifyContent(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(300)
	parent.SetHeight(100)
	parent.SetJustifyContent(layout.JustifyCenter)

	child := layout.NewNode()
	child.SetWidth(100)
	child.SetHeight(50)

	parent.AddChild(child)
	parent.CalculateLayout()

	// Child should be centered: (300 - 100) / 2 = 100
	if child.GetComputedLeft() != 100 {
		t.Errorf("Expected child centered at left=100, got %f", child.GetComputedLeft())
	}
}

func TestJustifyContentSpaceEvenly(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(10)
	parent.SetHeight(1)
	parent.SetJustifyContent(layout.JustifySpaceEvenly)

	left := layout.NewNode()
	left.SetWidth(1)
	left.SetHeight(1)

	right := layout.NewNode()
	right.SetWidth(1)
	right.SetHeight(1)

	parent.AddChild(left)
	parent.AddChild(right)
	parent.CalculateLayout()

	if int(left.GetComputedLeft()) != 2 {
		t.Errorf("Expected left child at x=2, got %f", left.GetComputedLeft())
	}

	if int(right.GetComputedLeft()) != 6 {
		t.Errorf("Expected right child at x=6, got %f", right.GetComputedLeft())
	}
}

func TestMarginAffectsChildPlacement(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(20)
	parent.SetHeight(5)

	child := layout.NewNode()
	child.SetWidth(3)
	child.SetHeight(1)
	child.SetMargin(layout.EdgeLeft, 2)
	child.SetMargin(layout.EdgeTop, 1)
	child.SetMargin(layout.EdgeRight, 1)

	parent.AddChild(child)
	parent.CalculateLayout()

	if child.GetComputedLeft() != 2 {
		t.Errorf("Expected child left=2, got %f", child.GetComputedLeft())
	}

	if child.GetComputedTop() != 1 {
		t.Errorf("Expected child top=1, got %f", child.GetComputedTop())
	}
}

func TestRootMarginAffectsOuterSize(t *testing.T) {
	node := layout.NewNode()
	node.SetMargin(layout.EdgeTop, 2)
	node.SetMargin(layout.EdgeBottom, 3)
	node.SetMargin(layout.EdgeLeft, 1)
	node.SetMargin(layout.EdgeRight, 4)

	child := layout.NewNode()
	child.SetWidth(2)
	child.SetHeight(1)
	node.AddChild(child)

	node.CalculateLayout()

	if node.GetOuterComputedWidth() != 7 {
		t.Errorf("Expected outer width 7, got %f", node.GetOuterComputedWidth())
	}

	if node.GetOuterComputedHeight() != 6 {
		t.Errorf("Expected outer height 6, got %f", node.GetOuterComputedHeight())
	}

	if child.GetComputedLeft() != 1 {
		t.Errorf("Expected child left=1 due to root margin, got %f", child.GetComputedLeft())
	}

	if child.GetComputedTop() != 2 {
		t.Errorf("Expected child top=2 due to root margin, got %f", child.GetComputedTop())
	}
}

func TestAlignItemsCenterInRow(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(10)
	parent.SetHeight(5)
	parent.SetAlignItems(layout.AlignCenter)

	child := layout.NewNode()
	child.SetWidth(2)
	child.SetHeight(1)

	parent.AddChild(child)
	parent.CalculateLayout()

	if child.GetComputedTop() != 2 {
		t.Errorf("Expected child top=2, got %f", child.GetComputedTop())
	}
}

func TestAlignItemsEndInColumn(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(10)
	parent.SetHeight(5)
	parent.SetAlignItems(layout.AlignEnd)

	child := layout.NewNode()
	child.SetWidth(3)
	child.SetHeight(1)

	parent.AddChild(child)
	parent.CalculateLayout()

	if child.GetComputedLeft() != 7 {
		t.Errorf("Expected child left=7, got %f", child.GetComputedLeft())
	}
}

// TestGetChildInvalidIndex tests getting child with invalid index
func TestGetChildInvalidIndex(t *testing.T) {
	node := layout.NewNode()

	if node.GetChild(-1) != nil {
		t.Error("Expected nil for negative index")
	}
	if node.GetChild(0) != nil {
		t.Error("Expected nil for index 0 (no children)")
	}
	if node.GetChild(100) != nil {
		t.Error("Expected nil for out of bounds index")
	}
}

// TestSetMargin tests setting margins
func TestSetMargin(t *testing.T) {
	node := layout.NewNode()
	node.SetMargin(layout.EdgeLeft, 10)
	node.SetMargin(layout.EdgeTop, 20)
	node.SetMargin(layout.EdgeRight, 30)
	node.SetMargin(layout.EdgeBottom, 40)

	// Can't directly access margin, but we can verify it doesn't crash
	node.SetWidth(100)
	node.SetHeight(100)
	node.CalculateLayout()
}

// TestSetMarginAll tests setting margin for all edges
func TestSetMarginAll(t *testing.T) {
	node := layout.NewNode()
	node.SetMargin(layout.EdgeAll, 15)

	// Verify it doesn't crash
	node.SetWidth(100)
	node.SetHeight(100)
	node.CalculateLayout()
}

// TestSetAlignItems tests setting align items
func TestSetAlignItems(t *testing.T) {
	node := layout.NewNode()

	node.SetAlignItems(layout.AlignStart)
	node.SetAlignItems(layout.AlignCenter)
	node.SetAlignItems(layout.AlignEnd)
	node.SetAlignItems(layout.AlignStretch)

	// Verify it doesn't crash
	node.CalculateLayout()
}

// TestJustifyEnd tests justify end alignment
func TestJustifyEnd(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(300)
	parent.SetHeight(100)
	parent.SetJustifyContent(layout.JustifyEnd)

	child := layout.NewNode()
	child.SetWidth(100)
	child.SetHeight(50)

	parent.AddChild(child)
	parent.CalculateLayout()

	// Child should be at the end: 300 - 100 = 200
	if child.GetComputedLeft() != 200 {
		t.Errorf("Expected child at left=200, got %f", child.GetComputedLeft())
	}
}

// TestJustifySpaceBetween tests space between alignment
func TestJustifySpaceBetween(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(400)
	parent.SetHeight(100)
	parent.SetJustifyContent(layout.JustifySpaceBetween)

	child1 := layout.NewNode()
	child1.SetWidth(100)
	child1.SetHeight(50)

	child2 := layout.NewNode()
	child2.SetWidth(100)
	child2.SetHeight(50)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.CalculateLayout()

	// First child at 0, second at 300 (400 - 100)
	// Space between = 200
	if child1.GetComputedLeft() != 0 {
		t.Errorf("Expected child1 at left=0, got %f", child1.GetComputedLeft())
	}
	if child2.GetComputedLeft() != 300 {
		t.Errorf("Expected child2 at left=300, got %f", child2.GetComputedLeft())
	}
}

// TestJustifySpaceAround tests space around alignment
func TestJustifySpaceAround(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(400)
	parent.SetHeight(100)
	parent.SetJustifyContent(layout.JustifySpaceAround)

	child1 := layout.NewNode()
	child1.SetWidth(100)
	child1.SetHeight(50)

	child2 := layout.NewNode()
	child2.SetWidth(100)
	child2.SetHeight(50)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.CalculateLayout()

	// Total space = 400 - 200 = 200
	// Space per child = 200 / 2 = 100
	// First child starts at 50 (half spacing before)
	// Second child starts at 150 (50 + 100 + 0 for space between)
	if child1.GetComputedLeft() != 50 {
		t.Errorf("Expected child1 at left=50, got %f", child1.GetComputedLeft())
	}
	if child2.GetComputedLeft() != 250 {
		t.Errorf("Expected child2 at left=250, got %f", child2.GetComputedLeft())
	}
}

// TestColumnJustifyCenter tests column justify center
func TestColumnJustifyCenter(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(100)
	parent.SetHeight(300)
	parent.SetJustifyContent(layout.JustifyCenter)

	child := layout.NewNode()
	child.SetWidth(50)
	child.SetHeight(100)

	parent.AddChild(child)
	parent.CalculateLayout()

	// Child should be centered vertically: (300 - 100) / 2 = 100
	if child.GetComputedTop() != 100 {
		t.Errorf("Expected child centered at top=100, got %f", child.GetComputedTop())
	}
}

// TestColumnJustifyEnd tests column justify end
func TestColumnJustifyEnd(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(100)
	parent.SetHeight(300)
	parent.SetJustifyContent(layout.JustifyEnd)

	child := layout.NewNode()
	child.SetWidth(50)
	child.SetHeight(100)

	parent.AddChild(child)
	parent.CalculateLayout()

	// Child should be at the bottom: 300 - 100 = 200
	if child.GetComputedTop() != 200 {
		t.Errorf("Expected child at top=200, got %f", child.GetComputedTop())
	}
}

// TestColumnSpaceBetween tests column space between
func TestColumnSpaceBetween(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(100)
	parent.SetHeight(400)
	parent.SetJustifyContent(layout.JustifySpaceBetween)

	child1 := layout.NewNode()
	child1.SetWidth(50)
	child1.SetHeight(100)

	child2 := layout.NewNode()
	child2.SetWidth(50)
	child2.SetHeight(100)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.CalculateLayout()

	// First child at 0, second at 300 (400 - 100)
	if child1.GetComputedTop() != 0 {
		t.Errorf("Expected child1 at top=0, got %f", child1.GetComputedTop())
	}
	if child2.GetComputedTop() != 300 {
		t.Errorf("Expected child2 at top=300, got %f", child2.GetComputedTop())
	}
}

// TestColumnSpaceAround tests column space around
func TestColumnSpaceAround(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(100)
	parent.SetHeight(400)
	parent.SetJustifyContent(layout.JustifySpaceAround)

	child1 := layout.NewNode()
	child1.SetWidth(50)
	child1.SetHeight(100)

	child2 := layout.NewNode()
	child2.SetWidth(50)
	child2.SetHeight(100)

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.CalculateLayout()

	// First child at 50, second at 250
	if child1.GetComputedTop() != 50 {
		t.Errorf("Expected child1 at top=50, got %f", child1.GetComputedTop())
	}
	if child2.GetComputedTop() != 250 {
		t.Errorf("Expected child2 at top=250, got %f", child2.GetComputedTop())
	}
}

// TestEdgePadding tests individual edge padding
func TestEdgePadding(t *testing.T) {
	tests := []struct {
		name  string
		edge  layout.Edge
		value float64
	}{
		{"left", layout.EdgeLeft, 5},
		{"top", layout.EdgeTop, 10},
		{"right", layout.EdgeRight, 15},
		{"bottom", layout.EdgeBottom, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := layout.NewNode()
			node.SetWidth(100)
			node.SetHeight(100)
			node.SetPadding(tt.edge, tt.value)

			child := layout.NewNode()
			child.SetWidth(50)
			child.SetHeight(50)

			node.AddChild(child)
			node.CalculateLayout()

			// Verify the child is offset by the padding
			// For row layout with left padding
			if tt.edge == layout.EdgeLeft {
				if child.GetComputedLeft() != tt.value {
					t.Errorf("Expected child left=%f, got %f", tt.value, child.GetComputedLeft())
				}
			}
		})
	}
}

// TestNestedLayout tests nested layout calculations
func TestNestedLayout(t *testing.T) {
	root := layout.NewNode()
	root.SetWidth(400)
	root.SetHeight(300)
	root.SetFlexDirection(layout.FlexDirectionRow)

	// Left panel
	leftPanel := layout.NewNode()
	leftPanel.SetWidth(200)
	leftPanel.SetHeight(300)

	// Right panel
	rightPanel := layout.NewNode()
	rightPanel.SetWidth(200)
	rightPanel.SetHeight(300)
	rightPanel.SetFlexDirection(layout.FlexDirectionColumn)

	// Add child to right panel
	rightChild := layout.NewNode()
	rightChild.SetWidth(100)
	rightChild.SetHeight(100)

	rightPanel.AddChild(rightChild)
	root.AddChild(leftPanel)
	root.AddChild(rightPanel)
	root.CalculateLayout()

	// Verify positions
	if leftPanel.GetComputedLeft() != 0 {
		t.Errorf("Expected leftPanel at left=0, got %f", leftPanel.GetComputedLeft())
	}
	if rightPanel.GetComputedLeft() != 200 {
		t.Errorf("Expected rightPanel at left=200, got %f", rightPanel.GetComputedLeft())
	}
	if rightChild.GetComputedLeft() != 200 {
		t.Errorf("Expected rightChild at left=200, got %f", rightChild.GetComputedLeft())
	}
	if rightChild.GetComputedTop() != 0 {
		t.Errorf("Expected rightChild at top=0, got %f", rightChild.GetComputedTop())
	}
}

// TestEmptyNodeLayout tests layout calculation on node with no children
func TestEmptyNodeLayout(t *testing.T) {
	node := layout.NewNode()
	node.SetWidth(100)
	node.SetHeight(100)

	// Should not crash
	node.CalculateLayout()

	if node.GetComputedWidth() != 100 {
		t.Errorf("Expected width 100, got %f", node.GetComputedWidth())
	}
	if node.GetComputedHeight() != 100 {
		t.Errorf("Expected height 100, got %f", node.GetComputedHeight())
	}
}

// TestConstantValues tests constant values
func TestConstantValues(t *testing.T) {
	// FlexDirection
	if layout.FlexDirectionRow != 0 {
		t.Errorf("Expected FlexDirectionRow=0, got %d", layout.FlexDirectionRow)
	}
	if layout.FlexDirectionColumn != 1 {
		t.Errorf("Expected FlexDirectionColumn=1, got %d", layout.FlexDirectionColumn)
	}

	// JustifyContent
	if layout.JustifyStart != 0 {
		t.Errorf("Expected JustifyStart=0, got %d", layout.JustifyStart)
	}
	if layout.JustifyCenter != 1 {
		t.Errorf("Expected JustifyCenter=1, got %d", layout.JustifyCenter)
	}
	if layout.JustifyEnd != 2 {
		t.Errorf("Expected JustifyEnd=2, got %d", layout.JustifyEnd)
	}
	if layout.JustifySpaceBetween != 3 {
		t.Errorf("Expected JustifySpaceBetween=3, got %d", layout.JustifySpaceBetween)
	}
	if layout.JustifySpaceAround != 4 {
		t.Errorf("Expected JustifySpaceAround=4, got %d", layout.JustifySpaceAround)
	}

	// AlignItems
	if layout.AlignStretch != 0 {
		t.Errorf("Expected AlignStretch=0, got %d", layout.AlignStretch)
	}
	if layout.AlignStart != 1 {
		t.Errorf("Expected AlignStart=1, got %d", layout.AlignStart)
	}
	if layout.AlignCenter != 2 {
		t.Errorf("Expected AlignCenter=2, got %d", layout.AlignCenter)
	}
	if layout.AlignEnd != 3 {
		t.Errorf("Expected AlignEnd=3, got %d", layout.AlignEnd)
	}

	// Edge
	if layout.EdgeLeft != 0 {
		t.Errorf("Expected EdgeLeft=0, got %d", layout.EdgeLeft)
	}
	if layout.EdgeTop != 1 {
		t.Errorf("Expected EdgeTop=1, got %d", layout.EdgeTop)
	}
	if layout.EdgeRight != 2 {
		t.Errorf("Expected EdgeRight=2, got %d", layout.EdgeRight)
	}
	if layout.EdgeBottom != 3 {
		t.Errorf("Expected EdgeBottom=3, got %d", layout.EdgeBottom)
	}
	if layout.EdgeAll != 4 {
		t.Errorf("Expected EdgeAll=4, got %d", layout.EdgeAll)
	}
}

func TestRowAlignSelfOverridesParentAlignment(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(10)
	parent.SetHeight(3)

	child := layout.NewNode()
	child.SetWidth(4)
	child.SetHeight(1)
	child.SetAlignSelf(layout.AlignCenter)

	parent.AddChild(child)
	parent.CalculateLayout()

	if child.GetComputedTop() != 1 {
		t.Fatalf("expected child to be vertically centered at y=1, got %f", child.GetComputedTop())
	}
}

func TestColumnAlignSelfOverridesParentAlignment(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(10)
	parent.SetHeight(3)

	child := layout.NewNode()
	child.SetWidth(4)
	child.SetHeight(1)
	child.SetAlignSelf(layout.AlignEnd)

	parent.AddChild(child)
	parent.CalculateLayout()

	if child.GetComputedLeft() != 6 {
		t.Fatalf("expected child to be right-aligned at x=6, got %f", child.GetComputedLeft())
	}
}

func TestRowFlexBasisSetsMainAxisSize(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(6)
	parent.SetHeight(1)

	child := layout.NewNode()
	child.SetHeight(1)
	child.SetFlexBasis(3)

	parent.AddChild(child)
	parent.CalculateLayout()

	if child.GetComputedWidth() != 3 {
		t.Fatalf("expected child width from flex basis to be 3, got %f", child.GetComputedWidth())
	}
}

func TestColumnFlexBasisPercentSetsMainAxisSize(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionColumn)
	parent.SetWidth(1)
	parent.SetHeight(6)

	child := layout.NewNode()
	child.SetWidth(1)
	child.SetFlexBasisPercent(50)

	parent.AddChild(child)
	parent.CalculateLayout()

	if child.GetComputedHeight() != 3 {
		t.Fatalf("expected child height from percentage flex basis to be 3, got %f", child.GetComputedHeight())
	}
}

func TestRowFlexShrinkShrinksChildrenToFit(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(10)
	parent.SetHeight(1)

	left := layout.NewNode()
	left.SetWidth(6)
	left.SetHeight(1)
	left.SetFlexShrink(1)

	right := layout.NewNode()
	right.SetWidth(6)
	right.SetHeight(1)
	right.SetFlexShrink(1)

	tail := layout.NewNode()
	tail.SetWidth(1)
	tail.SetHeight(1)

	parent.AddChild(left)
	parent.AddChild(right)
	parent.AddChild(tail)
	parent.CalculateLayout()

	if left.GetComputedWidth() != 5 {
		t.Fatalf("expected first shrinking child width 5, got %f", left.GetComputedWidth())
	}
	if right.GetComputedWidth() != 4 {
		t.Fatalf("expected second shrinking child width 4, got %f", right.GetComputedWidth())
	}
	if tail.GetComputedLeft() != 9 {
		t.Fatalf("expected tail child at x=9 after shrink, got %f", tail.GetComputedLeft())
	}
}

func TestRowFlexShrinkZeroPreservesOverflowingWidths(t *testing.T) {
	parent := layout.NewNode()
	parent.SetFlexDirection(layout.FlexDirectionRow)
	parent.SetWidth(16)
	parent.SetHeight(1)

	left := layout.NewNode()
	left.SetWidth(6)
	left.SetHeight(1)
	left.SetFlexShrink(0)

	middle := layout.NewNode()
	middle.SetWidth(6)
	middle.SetHeight(1)
	middle.SetFlexShrink(0)

	right := layout.NewNode()
	right.SetWidth(6)
	right.SetHeight(1)

	parent.AddChild(left)
	parent.AddChild(middle)
	parent.AddChild(right)
	parent.CalculateLayout()

	if left.GetComputedWidth() != 6 || middle.GetComputedWidth() != 6 || right.GetComputedWidth() != 6 {
		t.Fatal("expected shrink-disabled widths to remain unchanged")
	}
	if right.GetComputedLeft() != 12 {
		t.Fatalf("expected overflowing third child to remain at x=12, got %f", right.GetComputedLeft())
	}
}
