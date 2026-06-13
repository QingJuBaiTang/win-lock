package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
)

// makeTrayIcon generates a simple 32x32 lock icon for the system tray.
func makeTrayIcon() fyne.Resource {
	const sz = 32
	img := image.NewNRGBA(image.Rect(0, 0, sz, sz))

	bg := color.NRGBA{R: 30, G: 90, B: 200, A: 255}
	fg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	// Fill background
	for y := 0; y < sz; y++ {
		for x := 0; x < sz; x++ {
			img.Set(x, y, bg)
		}
	}

	fill := func(x0, y0, x1, y1 int, c color.NRGBA) {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				img.Set(x, y, c)
			}
		}
	}

	// Lock body
	fill(7, 17, 24, 28, fg)

	// Shackle outer (U-shape)
	fill(7, 8, 10, 18, fg)
	fill(21, 8, 24, 18, fg)
	fill(7, 8, 24, 11, fg)

	// Shackle inner cutout
	fill(11, 12, 20, 18, bg)

	// Keyhole
	fill(14, 20, 17, 25, bg)
	fill(13, 21, 18, 23, bg)

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return fyne.NewStaticResource("icon.png", buf.Bytes())
}
