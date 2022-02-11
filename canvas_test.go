package gosvg

import (
	"fmt"
	"image"
	"strings"
	"testing"

	"github.com/benoitkugler/webrender/backend"
	"github.com/benoitkugler/webrender/css/parser"
)

// test basic drawing commands

const template = `
	<?xml version="1.0" standalone="no"?>
	<svg width="600" height="600" viewBox="0 0 600 600"
		xmlns="http://www.w3.org/2000/svg">
		%s
	 </svg>
	 `

func TestDraw(t *testing.T) {
	s := fmt.Sprintf(template, `
		<circle style="fill:#8DD9FF;" cx="256" cy="256" r="256"/>
		<path style="fill:#FFFFF4;" d="M153.423,21.442C141.133,42.833,134,67.561,134,94c0,80.633,65.366,146,146,146s146-65.367,146-146
			c0-11.231-1.388-22.118-3.789-32.621C377.481,23.141,319.459,0,256,0C219.512,0,184.835,7.685,153.423,21.442z"/>
	`)
	// 	<g>
	// </g>
	// <g style="opacity:0.4;">
	// 	<circle style="fill:#FFFFF4;" cx="280" cy="94" r="68.639"/>
	// </g>
	img, err := Render(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	err = saveToPngFile("tmp.png", img)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRect(t *testing.T) {
	var width, height Fl = 600, 600
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	output := NewCanvas(0, 0, width, height, img)
	output.SetColorRgba(parser.RGBA{R: 0, G: 0.5, B: 0.5, A: 0.5}, false)
	output.SetColorRgba(parser.RGBA{R: 0.5, G: 0.1, B: 0.5, A: 1}, true)
	output.SetLineWidth(3)
	output.Rectangle(5, 5, 200, 200)
	output.Paint(backend.FillNonZero | backend.Stroke)
	saveToPngFile("tmp.png", output.image)
}
