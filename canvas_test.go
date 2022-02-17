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
	output := newCanvas(0, 0, width, height, img, nil)
	output.State().SetColorRgba(parser.RGBA{R: 0, G: 0.5, B: 0.5, A: 0.5}, false)
	output.State().SetColorRgba(parser.RGBA{R: 0.5, G: 0.1, B: 0.5, A: 1}, true)
	output.State().SetLineWidth(3)
	output.Rectangle(5, 5, 200, 200)
	output.Paint(backend.FillNonZero | backend.Stroke)
	saveToPngFile("tmp.png", output.state.image)
}

func TestStack(t *testing.T) {
	var width, height Fl = 600, 600
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	output := newCanvas(0, 0, width, height, img, nil)

	output.OnNewStack(func() {
		output.State().SetColorRgba(parser.RGBA{R: 0, G: 0.5, B: 0.5, A: 0.5}, false)
		output.State().SetColorRgba(parser.RGBA{R: 0.5, G: 0.1, B: 0.5, A: 1}, true)
		output.State().SetLineWidth(3)
		output.Rectangle(5, 5, 200, 200)
		output.Paint(backend.FillNonZero | backend.Stroke)
	})

	output.OnNewStack(func() {
		output.State().SetColorRgba(parser.RGBA{R: 0, G: 0.5, B: 1, A: 0.5}, false)
		output.Rectangle(100, 100, 50, 50)

		output.OnNewStack(func() {
			output.State().SetColorRgba(parser.RGBA{R: 0, G: 0.5, B: 0, A: 0.5}, false)
			output.Rectangle(150, 150, 50, 50)
			output.Paint(backend.FillNonZero | backend.Stroke)
		})

		output.Rectangle(200, 200, 50, 50)
		output.Paint(backend.FillNonZero | backend.Stroke)
	})

	saveToPngFile("tmp.png", output.state.image)
}

func TestGradient(t *testing.T) {
	input := `
	<?xml version="1.0"?>
	<svg width="500" height="500"
		xmlns="http://www.w3.org/2000/svg"
		xmlns:xlink="http://www.w3.org/1999/xlink">

	<radialGradient id="rg" cx="50%" cy="50%" fx="30%" fy ="30%"  r="30%" gradientUnits="objectBoundingBox"  
		gradientTransform=" rotate(-35) " spreadMethod="reflect" >
	<stop offset="1%" stop-color="powderblue" stop-opacity="1.00"/>
	<stop offset="10%" stop-color="silver" stop-opacity="1.00"/>
	<stop offset="30%" stop-color="darkblue" stop-opacity="1.00"/>
	<stop offset="40%" stop-color="black" stop-opacity="1.00"/>
	<stop offset="100%" stop-color="lime" stop-opacity="0.50"/>
	</radialGradient>

	<radialGradient id="rg2" cx="50%" cy="50%"   r="40%" gradientUnits="objectBoundingBox"  
	spreadMethod="repeat" >
	<stop offset="10%" stop-color="goldenrod" />
	<stop offset="30%" stop-color="seagreen" />
	<stop offset="50%" stop-color="cyan" />
	<stop offset="70%" stop-color="black" />
	<stop offset="100%" stop-color="orange" />
	</radialGradient>

	<linearGradient id="lg" x1="0%" y1="0%" x2="0%" y2="100%" spreadMethod="pad" gradientTransform="rotate(25) ">
	<stop offset="0%" stop-color="green" stop-opacity="1.00"/>
	<stop offset="50%" stop-color="red" stop-opacity="1.00"/>
	<stop offset="100%" stop-color="blue" stop-opacity="1.00"/>
	</linearGradient>


	<ellipse cx="300" cy="150" rx="120" ry="100"  style="fill:url(#rg2)" /> 

	<rect x="100" y="340" width="200" height="140" style="fill:url(#lg)" /> 
	<ellipse cx="350" cy="350" rx="80" ry="100" style="fill:url(#rg)" /> 
	
	</svg>
`
	img, err := Render(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	err = saveToPngFile("tmp.png", img)
	if err != nil {
		t.Fatal(err)
	}
}
