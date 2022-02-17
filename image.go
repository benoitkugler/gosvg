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

// apply mask to `src`
func applyOpacityMask(src *image.RGBA, mask *image.Alpha) {
	dst := src // update src in place
	sr := src.Bounds()
	dp := dst.Bounds().Min
	r := image.Rectangle{dp, dp.Add(sr.Size())}
	draw.DrawMask(dst, r, src, sr.Min, mask, sr.Min, draw.Src)
}

func rgbaToAlpha(r, g, b uint8) uint8 {
	const cR, cG, cB uint32 = 2989, 5870, 1141 // x 10_000

	v := (cR*uint32(r) + cG*uint32(g) + cB*uint32(b)) / 10_000
	return uint8(v)
}

// interprets `img` as an alpha mask
func rgbToAlpha(img *image.RGBA) *image.Alpha {
	b := img.Bounds()
	dst := image.NewAlpha(b)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			i1 := img.PixOffset(x, y)
			src := img.Pix[i1 : i1+3 : i1+3]

			i2 := dst.PixOffset(x, y)
			dst.Pix[i2] = rgbaToAlpha(src[0], src[1], src[2])
		}
	}
	return dst
}

func applyPattern(dst *image.RGBA, pattern *image.RGBA) {
	// FIXME:
	drawTo(dst, pattern)
}
