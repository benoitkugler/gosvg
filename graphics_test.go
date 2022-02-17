package gosvg

import (
	"image"
	"math"
	"testing"

	"github.com/benoitkugler/webrender/matrix"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

func flToFixed(x, y Fl) fixed.Point26_6 {
	return fixed.Point26_6{X: floatToFixed(x), Y: floatToFixed(y)}
}

func TestCoordinates(t *testing.T) {
	const width, height = 600, 600
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	f := rasterx.NewFiller(width, height, rasterx.NewScannerGV(width, height, img, img.Bounds()))
	f.Start(flToFixed(300, 300))
	f.Line(flToFixed(300, 400))
	f.Line(flToFixed(400, 400))
	f.Line(flToFixed(400, 300))
	f.Stop(true)

	f.Start(flToFixed(0, 0))
	f.Line(flToFixed(100, 0))
	f.Line(flToFixed(100, 100))
	f.Line(flToFixed(0, 100))
	f.Stop(true)

	// outside the image -> not displayed
	f.Start(flToFixed(700, 0))
	f.Line(flToFixed(700, 0))
	f.Line(flToFixed(700, 700))
	f.Stop(true)

	f.Draw()

	err := saveToPngFile("tmp.png", img)
	if err != nil {
		t.Fatal(err)
	}
}

func drawRectangle(dst *rasterx.Filler, mat matrix.Transform) {
	dst.Start(point{300, 300}.toFixed(mat))
	dst.Line(point{300, 400}.toFixed(mat))
	dst.Line(point{400, 400}.toFixed(mat))
	dst.Line(point{400, 300}.toFixed(mat))
	dst.Stop(true)
	dst.Draw()
	dst.Clear()
}

func TestTransforms(t *testing.T) {
	const width, height = 600, 600
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	f := rasterx.NewFiller(width, height, rasterx.NewScannerGV(width, height, img, img.Bounds()))

	drawRectangle(f, matrix.Identity())

	drawRectangle(f, matrix.Translation(150, 150))

	mt := matrix.Mul(matrix.Translation(350, 350), matrix.Mul(matrix.Rotation(math.Pi/4), matrix.Translation(-350, -350)))
	drawRectangle(f, mt)

	err := saveToPngFile("tmp.png", img)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGraphics(t *testing.T) {
	const width, height = 600, 600
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	scanner := rasterx.NewScannerGV(width, height, img, img.Bounds())
	d := rasterx.NewDasher(width, height, scanner)
	f := &d.Filler

	p := newRectangle(20, 20, 100, 100)
	s := shape{
		path:        p,
		fillColor:   plainColor{R: 1, A: 0.5},
		strokeColor: plainColor{B: 1, A: 0.8},
	}

	s.rasterize(f, d, matrix.Identity())

	s.rasterize(f, d, matrix.Translation(150, 150))

	mt := matrix.Mul(matrix.Translation(400, 400), matrix.Mul(matrix.Rotation(math.Pi/4), matrix.Translation(-60, -60)))
	s.rasterize(f, d, mt)

	err := saveToPngFile("tmp.png", img)
	if err != nil {
		t.Fatal(err)
	}
}
