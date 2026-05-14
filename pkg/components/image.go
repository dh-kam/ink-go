package components

import (
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// ImageData is a row-major RGB raster suitable for terminal half-block
// rendering. Pixels MUST have length Width * Height; rows are laid out
// top-to-bottom, left-to-right.
type ImageData struct {
	Width  int
	Height int
	Pixels [][3]uint8
}

// ImageProps configures Image.
type ImageProps struct {
	Image *ImageData
}

// halfBlockUpper is the unicode "Upper Half Block" glyph used to pack two
// vertical pixels (top = foreground, bottom = background) into a single
// terminal cell — the standard trick used by ink-image / chafa / viu.
const halfBlockUpper = "▀"

// Image renders an RGB raster as a block of ANSI 24-bit half-block glyphs.
//
// Encoding (per cell): "\x1b[38;2;<r>;<g>;<b>m\x1b[48;2;<r>;<g>;<b>m▀\x1b[0m"
// — top half uses the foreground color, bottom half uses the background.
//
// Heights that are not divisible by two are padded virtually with black.
// A nil image (or zero-sized image) renders an empty, but valid, node.
func Image(props ImageProps) *vdom.Node {
	output := encodeImage(props.Image)
	return vdom.CreateElement("image", markPublicComponentProps(nil), vdom.CreateTextNode(output))
}

func encodeImage(img *ImageData) string {
	if img == nil || img.Width <= 0 || img.Height <= 0 {
		return ""
	}

	expected := img.Width * img.Height
	if len(img.Pixels) < expected {
		// Defensive: refuse to read past the slice — render whatever we can,
		// but cap rows to what the slice supports.
		usableRows := len(img.Pixels) / img.Width
		if usableRows <= 0 {
			return ""
		}
		img = &ImageData{Width: img.Width, Height: usableRows, Pixels: img.Pixels}
	}

	var out strings.Builder
	// Rough preallocation: every pixel pair adds ~36 bytes.
	rows := (img.Height + 1) / 2
	out.Grow(rows * img.Width * 36)

	for y := 0; y < img.Height; y += 2 {
		for x := 0; x < img.Width; x++ {
			top := img.Pixels[y*img.Width+x]
			var bottom [3]uint8
			if y+1 < img.Height {
				bottom = img.Pixels[(y+1)*img.Width+x]
			}
			// else: leave bottom = {0,0,0} (black background pad).

			fmt.Fprintf(&out,
				"\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%s\x1b[0m",
				top[0], top[1], top[2],
				bottom[0], bottom[1], bottom[2],
				halfBlockUpper,
			)
		}

		if y+2 < img.Height {
			out.WriteByte('\n')
		}
	}

	return out.String()
}
