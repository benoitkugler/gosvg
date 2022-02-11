package gosvg

import (
	"image"
	"image/draw"

	"github.com/benoitkugler/webrender/backend"
)

func applyOpacity(img *image.RGBA, opacity backend.Fl) {
	// since img.Pix stores alpha-premultiplied values
	// applying an opacity factor amounts to multiply
	// every components
	for i, p := range img.Pix {
		img.Pix[i] = uint8(Fl(p) * opacity)
	}
}

func drawTo(dst, src *image.RGBA) {
	sr := src.Bounds()
	dp := dst.Bounds().Min
	r := image.Rectangle{dp, dp.Add(sr.Size())}
	draw.Draw(dst, r, src, sr.Min, draw.Over)
}
