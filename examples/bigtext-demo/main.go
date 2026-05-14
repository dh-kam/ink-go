// bigtext-demo prints the word "GOINK" using both embedded BigText
// fonts, then again with a foreground color applied — a quick visual
// smoke test for the components.BigText component.
package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/styles"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

const banner = "GOINK"

func main() {
	fmt.Println("BigText Component Demo")
	fmt.Println("======================")
	fmt.Println()

	fmt.Println("Block font (5x5):")
	printRows(components.BigText(components.BigTextProps{Text: banner, Font: components.FontBlock}))

	fmt.Println()
	fmt.Println("Tiny font (3x3):")
	printRows(components.BigText(components.BigTextProps{Text: banner, Font: components.FontTiny}))

	fmt.Println()
	fmt.Println("Block font with cyan color:")
	printRows(components.BigText(components.BigTextProps{
		Text:  banner,
		Font:  components.FontBlock,
		Color: styles.Cyan,
	}))

	fmt.Println()
	fmt.Println("Tiny font with magenta color:")
	printRows(components.BigText(components.BigTextProps{
		Text:  banner,
		Font:  components.FontTiny,
		Color: styles.Magenta,
	}))
}

// printRows prints the text-node children of a BigText node, one per
// line. Kept tiny so the demo doesn't reach for the runtime — the
// component's output already is a row-per-child vdom node.
func printRows(node *vdom.Node) {
	for _, child := range node.Children {
		if child == nil || child.Type != vdom.TextNode {
			continue
		}
		fmt.Println(child.Text)
	}
}
