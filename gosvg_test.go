package gosvg

import (
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

func TestMask(t *testing.T) {
	input := `
	<svg viewBox="-10 -10 150 150">
	<mask id="myMask">
		<!-- Everything under a white pixel will be visible -->
		<rect x="0" y="0" width="100" height="100" fill="white" />

		<!-- Everything under a black pixel will be invisible -->
		<path d="M10,35 A20,20,0,0,1,50,35 A20,20,0,0,1,90,35 Q90,65,50,95 Q10,65,10,35 Z" fill="black" />
	</mask>

	<polygon points="-10,110 110,110 110,-10" fill="orange" />

	<!-- with this mask applied, we "punch" a heart shape hole into the circle -->
	<circle cx="50" cy="50" r="50" mask="url(#myMask)" />
	</svg>
	`
	img, err := Render(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	err = saveToPngFile(filepath.Join(os.TempDir(), "svg_mask.png"), img)
	if err != nil {
		t.Fatal(err)
	}
}
