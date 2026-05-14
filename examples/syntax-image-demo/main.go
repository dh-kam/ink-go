package main

import (
	"fmt"

	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

// goSnippet is a small Go source listing used to exercise the Syntax component.
const goSnippet = `package main

// double returns n * 2.
func double(n int) int {
    return n * 2
}

var greeting = "hello, world"
`

// build4x4Raster produces a tiny RGB raster: a checker of red / cyan in the
// top half, green / magenta in the bottom — enough to show the half-block
// glyph encoding works for both the FG (top) and BG (bottom) halves.
func build4x4Raster() *components.ImageData {
	pixels := make([][3]uint8, 4*4)

	red := [3]uint8{220, 60, 60}
	cyan := [3]uint8{60, 200, 220}
	green := [3]uint8{60, 200, 100}
	magenta := [3]uint8{200, 60, 200}

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			var c [3]uint8
			topHalf := y < 2
			leftCell := (x % 2) == 0
			switch {
			case topHalf && leftCell:
				c = red
			case topHalf && !leftCell:
				c = cyan
			case !topHalf && leftCell:
				c = green
			default:
				c = magenta
			}
			pixels[y*4+x] = c
		}
	}

	return &components.ImageData{Width: 4, Height: 4, Pixels: pixels}
}

// App renders a syntax-highlighted Go block followed by a 4x4 RGB image.
func App() *vdom.Node {
	return components.Box(vdom.Props{"flexDirection": "column"},
		components.Text("Syntax (go):"),
		components.Newline(),
		components.Syntax(components.SyntaxProps{
			Code:     goSnippet,
			Language: components.SyntaxGo,
		}),
		components.Newline(),
		components.Text("Image (4x4 RGB):"),
		components.Newline(),
		components.Image(components.ImageProps{Image: build4x4Raster()}),
	)
}

func main() {
	fmt.Println(ink.RenderToString(App()))
}
