package gosvg

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/benoitkugler/webrender/backend"
)

func sampleImage(L int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, L, L))
	b := img.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(L / 10 * x), G: uint8(y), B: uint8(x + y), A: 0xff})
		}
	}
	return img
}

func assertEqual(t *testing.T, img1, img2 image.Image) {
	t.Helper()

	b1, b2 := img1.Bounds(), img2.Bounds()
	if b1 != b2 {
		t.Fatal()
	}

	for x := b1.Min.X; x < b1.Max.X; x++ {
		for y := b1.Min.Y; y < b1.Max.Y; y++ {
			if img1.At(x, y) != img2.At(x, y) {
				t.Fatal()
			}
		}
	}
}

func applyOpacitySlow(img *image.RGBA, opacity backend.Fl) {
	b := img.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			c := img.RGBAAt(x, y)
			c.R = uint8(Fl(c.R) * opacity)
			c.G = uint8(Fl(c.G) * opacity)
			c.B = uint8(Fl(c.B) * opacity)
			c.A = uint8(Fl(c.A) * opacity)
			img.SetRGBA(x, y, c)
		}
	}
}

func TestApplyOpacity(t *testing.T) {
	tmp := os.TempDir()

	s1 := sampleImage(200)
	s2 := sampleImage(200)

	if err := saveToPngFile(filepath.Join(tmp, "tmp1.png"), s1); err != nil {
		t.Fatal(err)
	}

	applyOpacity(s1, 0.5)
	applyOpacitySlow(s2, 0.5)
	assertEqual(t, s1, s2)

	if err := saveToPngFile(filepath.Join(tmp, "tmp0.5.png"), s1); err != nil {
		t.Fatal(err)
	}

	applyOpacity(s1, 0.5)
	if err := saveToPngFile(filepath.Join(tmp, "tmp0.25.png"), s1); err != nil {
		t.Fatal(err)
	}
}

func TestDrawTo(t *testing.T) {
	tmp := os.TempDir()

	s := sampleImage(200)
	s2 := sampleImage(100)
	applyOpacity(s2, 0.5)
	saveToPngFile(filepath.Join(tmp, "copy_1.png"), s)
	drawTo(s, s2)

	saveToPngFile(filepath.Join(tmp, "copy_2.png"), s)
}

func Test_rgbaToAlpha(t *testing.T) {
	tests := []struct {
		r, g, b uint8
		want    uint8
	}{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{255, 255, 255, 255},
		{120, 120, 120, 120},
		{60, 60, 60, 60},
	}
	for _, tt := range tests {
		if got := rgbaToAlpha(tt.r, tt.g, tt.b); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("rgbaToAlpha() = %v, want %v", got, tt.want)
		}
	}
}

func TestAlphaMask(t *testing.T) {
	tmp := os.TempDir()

	s1 := sampleImage(200)
	if err := saveToPngFile(filepath.Join(tmp, "alpha_1.png"), s1); err != nil {
		t.Fatal(err)
	}

	alpha := rgbToAlpha(s1)
	asGray := (*image.Gray)(alpha)
	if err := saveToPngFile(filepath.Join(tmp, "alpha_2.png"), asGray); err != nil {
		t.Fatal(err)
	}
}

func TestApplyOpacityMask(t *testing.T) {
	tmp := os.TempDir()

	s1 := sampleImage(200)
	if err := saveToPngFile(filepath.Join(tmp, "src.png"), s1); err != nil {
		t.Fatal(err)
	}

	s2 := sampleImage(50)
	mask := rgbToAlpha(s2)
	asGray := (*image.Gray)(mask)
	if err := saveToPngFile(filepath.Join(tmp, "alpha.png"), asGray); err != nil {
		t.Fatal(err)
	}

	applyOpacityMask(s1, mask)
	if err := saveToPngFile(filepath.Join(tmp, "src_with_alpha.png"), s1); err != nil {
		t.Fatal(err)
	}
}
