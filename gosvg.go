package gosvg

import (
	"image"
	"io"

	"github.com/benoitkugler/webrender/svg"
)

func Render(src io.Reader) (image.Image, error) {
	icon, err := svg.Parse(src, "", nil, nil)
	if err != nil {
		return nil, err
	}

	var width, height Fl = 600., 600.
	if vb := icon.ViewBox(); vb != nil {
		width, height = vb.Width, vb.Height
	}
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	output := newCanvas(0, 0, width, height, img, nil)
	icon.Draw(output, width, height, nil)

	return img, nil
}
