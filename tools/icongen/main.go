// Command icongen draws the TrashDay icon (blue bin + sparkles) and writes
// build/windows/trashday.ico and docs/images/icon.png. Run from the repo root:
//
//	go run ./tools/icongen
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type layer struct {
	inside func(x, y float64) bool
	color  func(x, y float64) color.RGBA
}

func flat(c color.RGBA) func(x, y float64) color.RGBA {
	return func(_, _ float64) color.RGBA { return c }
}

func rect(x0, y0, x1, y1, g float64) func(x, y float64) bool {
	return func(x, y float64) bool { return x >= x0-g && x <= x1+g && y >= y0-g && y <= y1+g }
}

// body is a trapezoid, wider at the top. g grows it (used for the outline).
func body(g float64) func(x, y float64) bool {
	return func(x, y float64) bool {
		if y < .31-g || y > .88+g {
			return false
		}
		t := (y - .31) / .57
		return x >= .22+.05*t-g && x <= .68-.05*t+g
	}
}

// star is a 4-point sparkle (astroid).
func star(cx, cy, r float64) func(x, y float64) bool {
	return func(x, y float64) bool {
		return math.Sqrt(math.Abs(x-cx)/r)+math.Sqrt(math.Abs(y-cy)/r) <= 1
	}
}

func or(fs ...func(x, y float64) bool) func(x, y float64) bool {
	return func(x, y float64) bool {
		for _, f := range fs {
			if f(x, y) {
				return true
			}
		}
		return false
	}
}

func handle(g float64) func(x, y float64) bool {
	outer, inner := rect(.36, .12, .54, .22, g), rect(.40, .16-g, .50, .22, -g)
	return func(x, y float64) bool { return outer(x, y) && !inner(x, y) }
}

func lerp(a, b uint8, t float64) uint8 { return uint8(float64(a) + (float64(b)-float64(a))*t) }

var (
	outline = color.RGBA{0x16, 0x3A, 0x8F, 255}
	top     = color.RGBA{0x4D, 0x9A, 0xFF, 255}
	bottom  = color.RGBA{0x1E, 0x4F, 0xC2, 255}
	stripe  = color.RGBA{0xA8, 0xD0, 0xFF, 255}
	lid     = color.RGBA{0x2F, 0x6B, 0xE0, 255}
	gold    = color.RGBA{0xFF, 0xC8, 0x2E, 255}
)

var layers = []layer{
	{or(body(.03), rect(.17, .19, .73, .29, .03), handle(.03)), flat(outline)},
	{body(0), func(_, y float64) color.RGBA {
		t := math.Max(0, math.Min(1, (y-.31)/.57))
		return color.RGBA{lerp(top.R, bottom.R, t), lerp(top.G, bottom.G, t), lerp(top.B, bottom.B, t), 255}
	}},
	{or(rect(.325, .39, .355, .81, 0), rect(.435, .39, .465, .81, 0), rect(.545, .39, .575, .81, 0)), flat(stripe)},
	{or(rect(.17, .19, .73, .29, 0), handle(0)), flat(lid)},
	{or(star(.82, .30, .17), star(.84, .62, .09)), flat(gold)},
}

func render(size int) []byte {
	const ss = 4 // supersampling per axis
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			var r, g, b, a float64
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					x := (float64(px) + (float64(sx)+.5)/ss) / float64(size)
					y := (float64(py) + (float64(sy)+.5)/ss) / float64(size)
					for i := len(layers) - 1; i >= 0; i-- {
						if layers[i].inside(x, y) {
							c := layers[i].color(x, y)
							r, g, b, a = r+float64(c.R), g+float64(c.G), b+float64(c.B), a+1
							break
						}
					}
				}
			}
			if a > 0 {
				img.SetNRGBA(px, py, color.NRGBA{uint8(r / a), uint8(g / a), uint8(b / a), uint8(a / ss / ss * 255)})
			}
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// main writes an ICO whose entries are PNGs (supported since Windows Vista).
func main() {
	sizes := []int{16, 24, 32, 48, 64, 128, 256}
	var head, data bytes.Buffer
	binary.Write(&head, binary.LittleEndian, [3]uint16{0, 1, uint16(len(sizes))})
	offset := 6 + 16*len(sizes)
	for _, s := range sizes {
		p := render(s)
		dim := uint8(s % 256) // 0 means 256
		binary.Write(&head, binary.LittleEndian, struct {
			W, H, Colors, Reserved uint8
			Planes, Bpp            uint16
			Size, Offset           uint32
		}{dim, dim, 0, 0, 1, 32, uint32(len(p)), uint32(offset)})
		data.Write(p)
		offset += len(p)
	}
	if err := os.WriteFile("build/windows/trashday.ico", append(head.Bytes(), data.Bytes()...), 0o644); err != nil {
		panic(err)
	}
	// Large PNG for the README.
	if err := os.WriteFile("docs/images/icon.png", render(256), 0o644); err != nil {
		panic(err)
	}
}
