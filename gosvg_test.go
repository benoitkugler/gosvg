package gosvg

import (
	"bytes"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func renderIcon(t testing.TB, iconPath string) {
	t.Helper()

	f, err := os.Open(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	img, err := Render(f)
	if err != nil {
		t.Fatal(err)
	}

	name := strings.TrimSuffix(filepath.Base(iconPath), filepath.Ext(iconPath)) + ".png"
	pngPath := filepath.Join("testdata", "out", name)
	err = saveToPngFile(pngPath, img)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLandscapeIcons(t *testing.T) {
	for _, p := range []string{
		"beach",
		"cape", "iceberg", "island",
		"mountains", "sea", "trees", "village",
	} {
		renderIcon(t, "testdata/landscapeIcons/"+p+".svg")
	}
}

func BenchmarkRaster(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, p := range []string{
			"beach",
			"cape", "iceberg", "island",
			"mountains", "sea", "trees", "village",
		} {
			renderIcon(b, "testdata/landscapeIcons/"+p+".svg")
		}
	}
}

func TestTestIcons(t *testing.T) {
	for _, p := range []string{
		"astronaut", "jupiter", "lander", "school-bus", "telescope", "content-cut-light",
		"defs",
		"24px",
	} {
		renderIcon(t, "testdata/testIcons/"+p+".svg")
	}
}

func TestStrokeIcons(t *testing.T) {
	for _, p := range []string{
		"OpacityStrokeDashTest.svg",
		"OpacityStrokeDashTest2.svg",
		"OpacityStrokeDashTest3.svg",
		"TestShapes.svg",
		"TestShapes2.svg",
		"TestShapes3.svg",
		"TestShapes4.svg",
		"TestShapes5.svg",
		"TestShapes6.svg",
	} {
		renderIcon(t, "testdata/"+p)
	}
}

// TODO: support text
// func TestPercentagesAndText(t *testing.T) {
// renderIcon(t, "testdata/TestPercentages.svg")
// }

func toPngBytes(m image.Image) ([]byte, error) {
	var b bytes.Buffer

	// Write the image into the buffer
	err := png.Encode(&b, m)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func saveToPngFile(filePath string, m image.Image) error {
	b, err := toPngBytes(m)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, b, os.ModePerm)
	return err
}
