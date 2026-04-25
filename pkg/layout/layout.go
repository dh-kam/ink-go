package layout

import (
	"math"
)

// FlexDirection defines the direction of flex items
type FlexDirection int

const (
	FlexDirectionRow FlexDirection = iota
	FlexDirectionColumn
)

// JustifyContent defines alignment along main axis
type JustifyContent int

const (
	JustifyStart JustifyContent = iota
	JustifyCenter
	JustifyEnd
	JustifySpaceBetween
	JustifySpaceAround
	JustifySpaceEvenly
)

// AlignItems defines alignment along cross axis
type AlignItems int

const (
	AlignStretch AlignItems = iota
	AlignStart
	AlignCenter
	AlignEnd
)

// WrapMode defines wrapping behavior along the cross axis.
type WrapMode int

const (
	WrapNoWrap WrapMode = iota
	WrapWrap
	WrapWrapReverse
)

// Edge represents edges for padding/margin
type Edge int

const (
	EdgeLeft Edge = iota
	EdgeTop
	EdgeRight
	EdgeBottom
	EdgeAll
)

// Node represents a layout node in the flexbox tree
type Node struct {
	// Style properties
	width          float64
	height         float64
	widthSet       bool
	heightSet      bool
	widthPct       bool
	heightPct      bool
	minWidth       float64
	minHeight      float64
	minWidthSet    bool
	minHeightSet   bool
	minWidthPct    bool
	minHeightPct   bool
	flexGrow       float64
	flexShrink     float64
	flexBasis      float64
	flexBasisSet   bool
	flexBasisPct   bool
	flexDirection  FlexDirection
	wrapMode       WrapMode
	justifyContent JustifyContent
	alignItems     AlignItems
	alignSelf      AlignItems
	alignSelfSet   bool
	gap            float64
	rowGap         float64
	rowGapSet      bool
	columnGap      float64
	columnGapSet   bool
	padding        [4]float64 // left, top, right, bottom
	margin         [4]float64 // left, top, right, bottom

	// Computed layout
	computedLeft   float64
	computedTop    float64
	computedWidth  float64
	computedHeight float64
	widthHint      float64
	widthHintSet   bool
	measureHeight  func(width float64) float64
	sizeAdjusted   bool
	textLike       bool

	// Tree structure
	children []*Node
	parent   *Node
}

// NewNode creates a new layout node
func NewNode() *Node {
	return &Node{
		flexDirection:  FlexDirectionRow,
		justifyContent: JustifyStart,
		alignItems:     AlignStretch,
	}
}

// SetWidth sets the width of the node
func (n *Node) SetWidth(width float64) {
	n.width = width
	n.widthSet = true
	n.widthPct = false
	n.computedWidth = width
}

// SetHeight sets the height of the node
func (n *Node) SetHeight(height float64) {
	n.height = height
	n.heightSet = true
	n.heightPct = false
	n.computedHeight = height
}

// SetWidthHint records the inherited available width for cases where upstream
// percentage minimum sizing resolves against parent availability instead of the
// parent's final computed width.
func (n *Node) SetWidthHint(width float64) {
	n.widthHint = width
	n.widthHintSet = true
}

// SetWidthPercent sets the width as a percentage of the parent content width.
func (n *Node) SetWidthPercent(percent float64) {
	n.width = percent
	n.widthSet = true
	n.widthPct = true
}

// SetHeightPercent sets the height as a percentage of the parent content height.
func (n *Node) SetHeightPercent(percent float64) {
	n.height = percent
	n.heightSet = true
	n.heightPct = true
}

// SetMinWidth sets the minimum width of the node.
func (n *Node) SetMinWidth(width float64) {
	n.minWidth = width
	n.minWidthSet = true
	n.minWidthPct = false
}

// SetMinHeight sets the minimum height of the node.
func (n *Node) SetMinHeight(height float64) {
	n.minHeight = height
	n.minHeightSet = true
	n.minHeightPct = false
}

// SetMinWidthPercent sets the minimum width as a percentage of the parent content width.
func (n *Node) SetMinWidthPercent(percent float64) {
	n.minWidth = percent
	n.minWidthSet = true
	n.minWidthPct = true
}

// SetMinHeightPercent sets the minimum height as a percentage of the parent content height.
func (n *Node) SetMinHeightPercent(percent float64) {
	n.minHeight = percent
	n.minHeightSet = true
	n.minHeightPct = true
}

// SetFlexGrow sets how much this node should grow along the main axis.
func (n *Node) SetFlexGrow(flexGrow float64) {
	n.flexGrow = flexGrow
}

// SetFlexShrink sets how much this node should shrink along the main axis.
func (n *Node) SetFlexShrink(flexShrink float64) {
	n.flexShrink = flexShrink
}

// SetFlexBasis sets the preferred main-axis size for this node.
func (n *Node) SetFlexBasis(basis float64) {
	n.flexBasis = basis
	n.flexBasisSet = true
	n.flexBasisPct = false
}

// SetFlexBasisPercent sets the preferred main-axis size as a percentage of the parent content size.
func (n *Node) SetFlexBasisPercent(percent float64) {
	n.flexBasis = percent
	n.flexBasisSet = true
	n.flexBasisPct = true
}

// SetMeasureHeightFunc sets a callback for nodes whose height depends on the resolved width.
func (n *Node) SetMeasureHeightFunc(measure func(width float64) float64) {
	n.measureHeight = measure
}

// SetTextLike marks nodes whose measured height depends on text wrapping semantics.
func (n *Node) SetTextLike(textLike bool) {
	n.textLike = textLike
}

// GetFlexGrow returns the configured flex grow factor.
func (n *Node) GetFlexGrow() float64 {
	return n.flexGrow
}

// GetComputedWidth returns the computed width
func (n *Node) GetComputedWidth() float64 {
	return n.computedWidth
}

// GetComputedHeight returns the computed height
func (n *Node) GetComputedHeight() float64 {
	return n.computedHeight
}

// HasAdjustedSize reports whether flex layout changed this node's main-axis size.
func (n *Node) HasAdjustedSize() bool {
	return n.sizeAdjusted
}

// IsTextLike reports whether this node behaves like a text leaf for layout rounding.
func (n *Node) IsTextLike() bool {
	return n.textLike
}

// GetOuterComputedWidth returns the computed width including horizontal margins.
func (n *Node) GetOuterComputedWidth() float64 {
	return n.outerWidth()
}

// GetOuterComputedHeight returns the computed height including vertical margins.
func (n *Node) GetOuterComputedHeight() float64 {
	return n.outerHeight()
}

// GetComputedLeft returns the computed left position
func (n *Node) GetComputedLeft() float64 {
	return n.computedLeft
}

// GetComputedTop returns the computed top position
func (n *Node) GetComputedTop() float64 {
	return n.computedTop
}

// HasExplicitWidth reports whether width was explicitly set.
func (n *Node) HasExplicitWidth() bool {
	return n.widthSet
}

// HasExplicitHeight reports whether height was explicitly set.
func (n *Node) HasExplicitHeight() bool {
	return n.heightSet
}

// SetFlexDirection sets the flex direction
func (n *Node) SetFlexDirection(direction FlexDirection) {
	n.flexDirection = direction
}

// GetFlexDirection returns the flex direction
func (n *Node) GetFlexDirection() FlexDirection {
	return n.flexDirection
}

// SetJustifyContent sets justify content
func (n *Node) SetJustifyContent(justify JustifyContent) {
	n.justifyContent = justify
}

// SetAlignItems sets align items
func (n *Node) SetAlignItems(align AlignItems) {
	n.alignItems = align
}

// SetAlignSelf overrides cross-axis alignment for this node.
func (n *Node) SetAlignSelf(align AlignItems) {
	n.alignSelf = align
	n.alignSelfSet = true
}

// SetGap sets the gap between children along both axes.
func (n *Node) SetGap(gap float64) {
	n.gap = gap
}

// SetRowGap sets the vertical gap between rows.
func (n *Node) SetRowGap(gap float64) {
	n.rowGap = gap
	n.rowGapSet = true
}

// SetColumnGap sets the horizontal gap between columns.
func (n *Node) SetColumnGap(gap float64) {
	n.columnGap = gap
	n.columnGapSet = true
}

// SetWrapMode sets the flex wrap mode.
func (n *Node) SetWrapMode(mode WrapMode) {
	n.wrapMode = mode
}

// SetPadding sets padding for an edge
func (n *Node) SetPadding(edge Edge, value float64) {
	switch edge {
	case EdgeLeft:
		n.padding[0] = value
	case EdgeTop:
		n.padding[1] = value
	case EdgeRight:
		n.padding[2] = value
	case EdgeBottom:
		n.padding[3] = value
	case EdgeAll:
		n.padding[0] = value
		n.padding[1] = value
		n.padding[2] = value
		n.padding[3] = value
	}
}

// SetMargin sets margin for an edge
func (n *Node) SetMargin(edge Edge, value float64) {
	switch edge {
	case EdgeLeft:
		n.margin[0] = value
	case EdgeTop:
		n.margin[1] = value
	case EdgeRight:
		n.margin[2] = value
	case EdgeBottom:
		n.margin[3] = value
	case EdgeAll:
		n.margin[0] = value
		n.margin[1] = value
		n.margin[2] = value
		n.margin[3] = value
	}
}

// AddChild adds a child node
func (n *Node) AddChild(child *Node) {
	child.parent = n
	n.children = append(n.children, child)
}

// GetChildCount returns the number of children
func (n *Node) GetChildCount() int {
	return len(n.children)
}

// GetChild returns a child by index
func (n *Node) GetChild(index int) *Node {
	if index < 0 || index >= len(n.children) {
		return nil
	}
	return n.children[index]
}

// CalculateLayout calculates the layout for this node and its children
func (n *Node) CalculateLayout() {
	n.measure()
	n.calculateLayoutInternal(0, 0)
}

func (n *Node) measure() {
	for _, child := range n.children {
		child.measure()
	}

	if len(n.children) == 0 {
		paddingLeft := n.padding[0]
		paddingTop := n.padding[1]
		paddingRight := n.padding[2]
		paddingBottom := n.padding[3]

		if n.widthSet && !n.widthPct {
			n.computedWidth = n.width
		} else if n.computedWidth < paddingLeft+paddingRight {
			n.computedWidth = paddingLeft + paddingRight
		}

		if n.heightSet && !n.heightPct {
			n.computedHeight = n.height
		} else if n.computedHeight < paddingTop+paddingBottom {
			n.computedHeight = paddingTop + paddingBottom
		}

		n.applyResolvedMinimums(0, 0)

		return
	}

	if n.widthSet && !n.widthPct {
		n.computedWidth = n.width
	}

	if n.heightSet && !n.heightPct {
		n.computedHeight = n.height
	}

	if n.widthSet && !n.widthPct && n.heightSet && !n.heightPct {
		n.applyResolvedMinimums(0, 0)
		return
	}

	paddingLeft := n.padding[0]
	paddingTop := n.padding[1]
	paddingRight := n.padding[2]
	paddingBottom := n.padding[3]

	var contentWidth float64
	var contentHeight float64

	if n.flexDirection == FlexDirectionRow {
		for _, child := range n.children {
			contentWidth += child.outerWidth()
			if child.outerHeight() > contentHeight {
				contentHeight = child.outerHeight()
			}
		}
		if len(n.children) > 1 {
			contentWidth += n.mainAxisGap() * float64(len(n.children)-1)
		}
	} else {
		for _, child := range n.children {
			contentHeight += child.outerHeight()
			if child.outerWidth() > contentWidth {
				contentWidth = child.outerWidth()
			}
		}
		if len(n.children) > 1 {
			contentHeight += n.mainAxisGap() * float64(len(n.children)-1)
		}
	}

	if !n.widthSet {
		n.computedWidth = contentWidth + paddingLeft + paddingRight
	}

	if !n.heightSet {
		n.computedHeight = contentHeight + paddingTop + paddingBottom
	}

	n.applyResolvedMinimums(0, 0)
}

func (n *Node) calculateLayoutInternal(offsetX, offsetY float64) {
	// Apply padding
	paddingLeft := n.padding[0]
	paddingTop := n.padding[1]
	paddingRight := n.padding[2]
	paddingBottom := n.padding[3]

	baseX := offsetX
	baseY := offsetY
	if n.parent == nil {
		baseX += n.margin[0]
		baseY += n.margin[1]
	}

	// Content area
	contentWidth := n.computedWidth - paddingLeft - paddingRight
	contentHeight := n.computedHeight - paddingTop - paddingBottom

	if len(n.children) == 0 {
		return
	}

	n.resolvePercentageAndMinimumConstraints(contentWidth, contentHeight)

	if n.wrapMode != WrapNoWrap {
		n.calculateWrappedLayout(baseX, baseY, contentWidth, contentHeight)
		return
	}

	n.resolveFlexBasis(contentWidth, contentHeight)

	// Calculate total size of children
	var totalChildSize float64
	var totalFlexGrow float64
	var totalFlexShrink float64
	for _, child := range n.children {
		if n.flexDirection == FlexDirectionRow {
			totalChildSize += child.outerWidth()
			totalFlexShrink += child.flexShrink * child.computedWidth
		} else {
			totalChildSize += child.outerHeight()
			totalFlexShrink += child.flexShrink * child.computedHeight
		}
		totalFlexGrow += child.flexGrow
	}
	if len(n.children) > 1 {
		totalChildSize += n.mainAxisGap() * float64(len(n.children)-1)
	}

	availableSpace := contentWidth - totalChildSize
	if n.flexDirection == FlexDirectionColumn {
		availableSpace = contentHeight - totalChildSize
	}

	if availableSpace > 0 && totalFlexGrow > 0 {
		for _, child := range n.children {
			if child.flexGrow <= 0 {
				continue
			}

			growth := availableSpace * (child.flexGrow / totalFlexGrow)
			if n.flexDirection == FlexDirectionRow {
				child.computedWidth += growth
			} else {
				child.computedHeight += growth
			}
			child.sizeAdjusted = true
		}

		totalChildSize += availableSpace
	}

	if availableSpace < 0 && totalFlexShrink > 0 {
		deficit := -availableSpace
		distributedShrink := 0.0

		for _, child := range n.children {
			mainSize := child.computedWidth
			if n.flexDirection == FlexDirectionColumn {
				mainSize = child.computedHeight
			}

			weight := child.flexShrink * mainSize
			if weight <= 0 || mainSize <= 0 {
				continue
			}

			shrink := deficit * (weight / totalFlexShrink)
			if shrink > mainSize {
				shrink = mainSize
			}

			if n.flexDirection == FlexDirectionRow {
				child.computedWidth -= shrink
			} else {
				child.computedHeight -= shrink
			}
			if shrink > 1e-9 {
				child.sizeAdjusted = true
			}

			distributedShrink += shrink
		}

		totalChildSize -= distributedShrink
	}

	n.reapplyChildMinimums(contentWidth, contentHeight)
	totalChildSize = n.totalChildMainSize()

	if n.flexDirection == FlexDirectionRow {
		n.remeasureChildrenForWidth()
		if n.shouldRecomputeRowCrossSize() {
			n.recomputeCrossSizeFromChildren()
			contentHeight = n.computedHeight - paddingTop - paddingBottom
		}
	}

	// Calculate starting position based on justify content
	var currentPos float64
	var spacing float64
	if n.flexDirection == FlexDirectionRow {
		switch n.justifyContent {
		case JustifyStart:
			currentPos = 0
		case JustifyCenter:
			currentPos = (contentWidth - totalChildSize) / 2
		case JustifyEnd:
			currentPos = contentWidth - totalChildSize
		case JustifySpaceBetween:
			currentPos = 0
			if len(n.children) > 1 {
				spacing = (contentWidth - totalChildSize) / float64(len(n.children)-1)
			}
		case JustifySpaceAround:
			if len(n.children) > 0 {
				spacing = (contentWidth - totalChildSize) / float64(len(n.children))
				currentPos = spacing / 2
			}
		case JustifySpaceEvenly:
			if len(n.children) > 0 {
				spacing = (contentWidth - totalChildSize) / float64(len(n.children)+1)
				currentPos = spacing
			}
		}
	} else {
		switch n.justifyContent {
		case JustifyStart:
			currentPos = 0
		case JustifyCenter:
			currentPos = (contentHeight - totalChildSize) / 2
		case JustifyEnd:
			currentPos = contentHeight - totalChildSize
		case JustifySpaceBetween:
			currentPos = 0
			if len(n.children) > 1 {
				spacing = (contentHeight - totalChildSize) / float64(len(n.children)-1)
			}
		case JustifySpaceAround:
			if len(n.children) > 0 {
				spacing = (contentHeight - totalChildSize) / float64(len(n.children))
				currentPos = spacing / 2
			}
		case JustifySpaceEvenly:
			if len(n.children) > 0 {
				spacing = (contentHeight - totalChildSize) / float64(len(n.children)+1)
				currentPos = spacing
			}
		}
	}

	// Position children
	adjustedChildren := n.hasAdjustedChildren()
	for index, child := range n.children {
		if n.flexDirection == FlexDirectionRow {
			if !adjustedChildren {
				currentPos += child.margin[0]
				child.computedLeft = baseX + paddingLeft + currentPos
				child.computedTop = baseY + paddingTop + n.crossAxisOffset(contentHeight, child, true)

				child.calculateLayoutInternal(child.computedLeft, child.computedTop)

				currentPos += child.computedWidth + child.margin[2]
				if index < len(n.children)-1 {
					currentPos += n.mainAxisGap()
				}

				if n.justifyContent == JustifySpaceBetween || n.justifyContent == JustifySpaceAround || n.justifyContent == JustifySpaceEvenly {
					currentPos += spacing
				}
				continue
			}

			currentPos += child.margin[0]
			exactStart := currentPos
			exactWidth := child.computedWidth
			exactEnd := exactStart + exactWidth

			roundedStart := math.Round(exactStart)
			finalWidth := child.computedWidth
			if child.textLike && len(n.children) > 1 {
				roundedStart = math.Floor(exactStart + 1e-9)
				if child.sizeAdjusted {
					finalWidth = math.Ceil(exactWidth - 1e-9)
				}
			} else {
				if child.sizeAdjusted {
					roundedEnd := math.Round(exactEnd)
					finalWidth = roundedEnd - roundedStart
					if finalWidth < 0 {
						finalWidth = 0
					}
				}
			}

			child.computedLeft = baseX + paddingLeft + roundedStart
			child.computedTop = baseY + paddingTop + n.crossAxisOffset(contentHeight, child, true)

			// Recursively calculate child layout using the exact allocated width. After that,
			// restore the rounded width used for rendering and sibling overlap parity.
			child.computedWidth = exactWidth
			child.calculateLayoutInternal(child.computedLeft, child.computedTop)
			child.computedWidth = finalWidth

			currentPos = exactEnd + child.margin[2]
			if index < len(n.children)-1 {
				currentPos += n.mainAxisGap()
			}

			if n.justifyContent == JustifySpaceBetween || n.justifyContent == JustifySpaceAround || n.justifyContent == JustifySpaceEvenly {
				currentPos += spacing
			}
		} else {
			if !adjustedChildren {
				currentPos += child.margin[1]
				child.computedLeft = baseX + paddingLeft + n.crossAxisOffset(contentWidth, child, false)
				child.computedTop = baseY + paddingTop + currentPos

				child.calculateLayoutInternal(child.computedLeft, child.computedTop)

				currentPos += child.computedHeight + child.margin[3]
				if index < len(n.children)-1 {
					currentPos += n.mainAxisGap()
				}

				if n.justifyContent == JustifySpaceBetween || n.justifyContent == JustifySpaceAround || n.justifyContent == JustifySpaceEvenly {
					currentPos += spacing
				}
				continue
			}

			currentPos += child.margin[1]
			exactStart := currentPos
			exactHeight := child.computedHeight
			exactEnd := exactStart + exactHeight
			roundedStart := math.Round(exactStart)
			finalHeight := child.computedHeight
			if child.textLike && len(n.children) > 1 {
				roundedStart = math.Floor(exactStart + 1e-9)
				if child.sizeAdjusted {
					finalHeight = math.Ceil(exactHeight - 1e-9)
				}
			} else if child.sizeAdjusted {
				roundedEnd := math.Round(exactEnd)
				finalHeight = roundedEnd - roundedStart
				if finalHeight < 0 {
					finalHeight = 0
				}
			}

			child.computedLeft = baseX + paddingLeft + n.crossAxisOffset(contentWidth, child, false)
			child.computedTop = baseY + paddingTop + roundedStart

			child.computedHeight = exactHeight
			child.calculateLayoutInternal(child.computedLeft, child.computedTop)
			child.computedHeight = finalHeight

			currentPos = exactEnd + child.margin[3]
			if index < len(n.children)-1 {
				currentPos += n.mainAxisGap()
			}

			if n.justifyContent == JustifySpaceBetween || n.justifyContent == JustifySpaceAround || n.justifyContent == JustifySpaceEvenly {
				currentPos += spacing
			}
		}
	}

	if n.flexDirection == FlexDirectionRow && n.shouldRecomputeRowCrossSize() {
		n.recomputeCrossSizeFromChildren()
	}

	parentContentWidth, parentContentHeight := n.parentContentDimensions()
	n.applyResolvedMinimums(parentContentWidth, parentContentHeight)
}

func (n *Node) remeasureChildrenForWidth() {
	for _, child := range n.children {
		if child.measureHeight == nil || !child.sizeAdjusted {
			continue
		}

		child.computedHeight = child.measureHeight(child.computedWidth)
	}
}

func (n *Node) recomputeCrossSizeFromChildren() {
	paddingLeft := n.padding[0]
	paddingTop := n.padding[1]
	paddingRight := n.padding[2]
	paddingBottom := n.padding[3]

	if n.flexDirection == FlexDirectionRow {
		contentHeight := 0.0
		for _, child := range n.children {
			if child.outerHeight() > contentHeight {
				contentHeight = child.outerHeight()
			}
		}

		n.computedHeight = contentHeight + paddingTop + paddingBottom
		return
	}

	contentWidth := 0.0
	for _, child := range n.children {
		if child.outerWidth() > contentWidth {
			contentWidth = child.outerWidth()
		}
	}

	n.computedWidth = contentWidth + paddingLeft + paddingRight
}

func (n *Node) shouldRecomputeRowCrossSize() bool {
	if n.heightSet {
		return false
	}

	return n.parent == nil || n.parent.flexDirection == FlexDirectionRow
}

func (n *Node) shouldUseTextOverlapRounding() bool {
	if n.flexDirection != FlexDirectionRow || len(n.children) < 3 {
		return false
	}

	adjustedTextLike := 0
	for _, child := range n.children {
		if child.sizeAdjusted && child.textLike {
			adjustedTextLike++
		}
	}

	return adjustedTextLike >= 2
}

func (n *Node) hasAdjustedChildren() bool {
	for _, child := range n.children {
		if child.sizeAdjusted {
			return true
		}
	}

	return false
}

func (n *Node) outerWidth() float64 {
	return n.margin[0] + n.computedWidth + n.margin[2]
}

func (n *Node) outerHeight() float64 {
	return n.margin[1] + n.computedHeight + n.margin[3]
}

func (n *Node) crossAxisOffset(availableSpace float64, child *Node, isRow bool) float64 {
	marginStart := child.margin[1]
	marginEnd := child.margin[3]
	childSize := child.computedHeight

	if !isRow {
		marginStart = child.margin[0]
		marginEnd = child.margin[2]
		childSize = child.computedWidth
	}

	usableSpace := availableSpace - childSize - marginStart - marginEnd
	if usableSpace < 0 {
		usableSpace = 0
	}

	align := n.alignItems
	if child.alignSelfSet {
		align = child.alignSelf
	}

	switch align {
	case AlignStart:
		return marginStart
	case AlignCenter:
		return marginStart + usableSpace/2
	case AlignEnd:
		return marginStart + usableSpace
	case AlignStretch:
		if isRow {
			if !child.heightSet {
				child.computedHeight = availableSpace - child.margin[1] - child.margin[3]
				if child.computedHeight < 0 {
					child.computedHeight = 0
				}
			}
			return child.margin[1]
		}

		if !child.widthSet {
			child.computedWidth = availableSpace - child.margin[0] - child.margin[2]
			if child.computedWidth < 0 {
				child.computedWidth = 0
			}
		}
		return child.margin[0]
	default:
		return marginStart
	}
}

func (n *Node) resolveFlexBasis(contentWidth float64, contentHeight float64) {
	for _, child := range n.children {
		if !child.flexBasisSet {
			continue
		}

		resolved := child.flexBasis
		if child.flexBasisPct {
			if n.flexDirection == FlexDirectionRow {
				resolved = contentWidth * child.flexBasis / 100
			} else {
				resolved = contentHeight * child.flexBasis / 100
			}
		}

		if n.flexDirection == FlexDirectionRow {
			if !child.widthSet {
				child.computedWidth = resolved
			}
			continue
		}

		if !child.heightSet {
			child.computedHeight = resolved
		}
	}
}

func (n *Node) resolvedMinimumWidth(parentContentWidth float64) float64 {
	if !n.minWidthSet {
		return 0
	}

	if n.minWidthPct {
		if n.parent != nil && n.parent.widthHintSet {
			return n.parent.widthHint * n.minWidth / 100
		}

		return parentContentWidth * n.minWidth / 100
	}

	return n.minWidth
}

func (n *Node) resolvedMinimumHeight(parentContentHeight float64) float64 {
	if !n.minHeightSet {
		return 0
	}

	if n.minHeightPct {
		return parentContentHeight * n.minHeight / 100
	}

	return n.minHeight
}

func (n *Node) applyResolvedMinimums(parentContentWidth, parentContentHeight float64) {
	minWidth := n.resolvedMinimumWidth(parentContentWidth)
	if n.computedWidth < minWidth {
		n.computedWidth = minWidth
	}

	minHeight := n.resolvedMinimumHeight(parentContentHeight)
	if n.computedHeight < minHeight {
		n.computedHeight = minHeight
	}
}

func (n *Node) resolvePercentageAndMinimumConstraints(parentContentWidth, parentContentHeight float64) {
	for _, child := range n.children {
		if child.widthSet && child.widthPct {
			child.computedWidth = parentContentWidth * child.width / 100
		}
		if child.heightSet && child.heightPct {
			child.computedHeight = parentContentHeight * child.height / 100
		}

		child.applyResolvedMinimums(parentContentWidth, parentContentHeight)
	}
}

func (n *Node) reapplyChildMinimums(parentContentWidth, parentContentHeight float64) {
	for _, child := range n.children {
		child.applyResolvedMinimums(parentContentWidth, parentContentHeight)
	}
}

func (n *Node) totalChildMainSize() float64 {
	total := 0.0
	for _, child := range n.children {
		if n.flexDirection == FlexDirectionRow {
			total += child.outerWidth()
		} else {
			total += child.outerHeight()
		}
	}

	if len(n.children) > 1 {
		total += n.mainAxisGap() * float64(len(n.children)-1)
	}

	return total
}

func (n *Node) parentContentDimensions() (float64, float64) {
	if n == nil || n.parent == nil {
		return 0, 0
	}

	width := n.parent.computedWidth - n.parent.padding[0] - n.parent.padding[2]
	height := n.parent.computedHeight - n.parent.padding[1] - n.parent.padding[3]
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return width, height
}

type wrapLine struct {
	children  []*Node
	mainSize  float64
	crossSize float64
}

func (n *Node) calculateWrappedLayout(baseX, baseY, contentWidth, contentHeight float64) {
	n.resolveFlexBasis(contentWidth, contentHeight)

	lines := n.buildWrapLines(contentWidth, contentHeight)
	if len(lines) == 0 {
		return
	}

	totalCross := 0.0
	for _, line := range lines {
		totalCross += line.crossSize
	}
	if len(lines) > 1 {
		totalCross += n.crossAxisGap() * float64(len(lines)-1)
	}

	if n.flexDirection == FlexDirectionRow {
		if !n.heightSet {
			n.computedHeight = totalCross + n.padding[1] + n.padding[3]
			contentHeight = totalCross
		}
	} else if !n.widthSet {
		n.computedWidth = totalCross + n.padding[0] + n.padding[2]
		contentWidth = totalCross
	}

	parentContentWidth, parentContentHeight := n.parentContentDimensions()
	n.applyResolvedMinimums(parentContentWidth, parentContentHeight)

	crossCursor := 0.0
	crossLimit := contentHeight
	if n.flexDirection == FlexDirectionColumn {
		crossLimit = contentWidth
	}

	reverse := n.wrapMode == WrapWrapReverse
	if reverse {
		crossCursor = crossLimit
	}

	for lineIndex, line := range lines {
		lineCrossStart := crossCursor
		if reverse {
			lineCrossStart -= line.crossSize
		}

		mainCursor, spacing := n.mainAxisPlacement(line.mainSize, contentWidth, contentHeight, len(line.children))

		for childIndex, child := range line.children {
			if n.flexDirection == FlexDirectionRow {
				mainCursor += child.margin[0]
				child.computedLeft = baseX + n.padding[0] + mainCursor
				child.computedTop = baseY + n.padding[1] + lineCrossStart + n.crossAxisOffset(line.crossSize, child, true)
				child.calculateLayoutInternal(child.computedLeft, child.computedTop)
				mainCursor += child.computedWidth + child.margin[2]
			} else {
				mainCursor += child.margin[1]
				child.computedLeft = baseX + n.padding[0] + lineCrossStart + n.crossAxisOffset(line.crossSize, child, false)
				child.computedTop = baseY + n.padding[1] + mainCursor
				child.calculateLayoutInternal(child.computedLeft, child.computedTop)
				mainCursor += child.computedHeight + child.margin[3]
			}

			if childIndex < len(line.children)-1 {
				mainCursor += n.mainAxisGap() + spacing
			}
		}

		if reverse {
			crossCursor = lineCrossStart
			if lineIndex < len(lines)-1 {
				crossCursor -= n.crossAxisGap()
			}
			continue
		}

		crossCursor = lineCrossStart + line.crossSize
		if lineIndex < len(lines)-1 {
			crossCursor += n.crossAxisGap()
		}
	}
}

func (n *Node) buildWrapLines(contentWidth, contentHeight float64) []wrapLine {
	mainLimit := contentWidth
	if n.flexDirection == FlexDirectionColumn {
		mainLimit = contentHeight
	}

	lines := make([]wrapLine, 0, 1)
	current := wrapLine{}

	for _, child := range n.children {
		mainSize := child.outerWidth()
		crossSize := child.outerHeight()
		if n.flexDirection == FlexDirectionColumn {
			mainSize = child.outerHeight()
			crossSize = child.outerWidth()
		}

		required := mainSize
		if len(current.children) > 0 {
			required += n.mainAxisGap()
		}

		if len(current.children) > 0 && mainLimit > 0 && current.mainSize+required > mainLimit+1e-9 {
			lines = append(lines, current)
			current = wrapLine{}
		}

		if len(current.children) > 0 {
			current.mainSize += n.mainAxisGap()
		}
		current.children = append(current.children, child)
		current.mainSize += mainSize
		if crossSize > current.crossSize {
			current.crossSize = crossSize
		}
	}

	if len(current.children) > 0 {
		lines = append(lines, current)
	}

	return lines
}

func (n *Node) mainAxisGap() float64 {
	if n.flexDirection == FlexDirectionColumn {
		if n.rowGapSet {
			return n.rowGap
		}
		return n.gap
	}

	if n.columnGapSet {
		return n.columnGap
	}

	return n.gap
}

func (n *Node) crossAxisGap() float64 {
	if n.flexDirection == FlexDirectionColumn {
		if n.columnGapSet {
			return n.columnGap
		}
		return n.gap
	}

	if n.rowGapSet {
		return n.rowGap
	}

	return n.gap
}

func (n *Node) mainAxisPlacement(lineMainSize, contentWidth, contentHeight float64, childCount int) (float64, float64) {
	available := contentWidth - lineMainSize
	if n.flexDirection == FlexDirectionColumn {
		available = contentHeight - lineMainSize
	}

	currentPos := 0.0
	spacing := 0.0
	switch n.justifyContent {
	case JustifyStart:
		currentPos = 0
	case JustifyCenter:
		currentPos = available / 2
	case JustifyEnd:
		currentPos = available
	case JustifySpaceBetween:
		if childCount > 1 {
			spacing = available / float64(childCount-1)
		}
	case JustifySpaceAround:
		if childCount > 0 {
			spacing = available / float64(childCount)
			currentPos = spacing / 2
		}
	case JustifySpaceEvenly:
		if childCount > 0 {
			spacing = available / float64(childCount+1)
			currentPos = spacing
		}
	}

	return currentPos, spacing
}
