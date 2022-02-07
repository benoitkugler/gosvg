package gosvg

import (
	"os"
	"testing"
)

func renderIcon(t *testing.T, iconPath string) {
	f, err := os.Open(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	err = Render(f)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLandscapeIcons(t *testing.T) {
	for _, p := range []string{
		"beach", "cape", "iceberg", "island",
		"mountains", "sea", "trees", "village",
	} {
		renderIcon(t, "testdata/landscapeIcons/"+p+".svg")
	}
}

func TestTestIcons(t *testing.T) {
	for _, p := range []string{
		"astronaut", "jupiter", "lander", "school-bus", "telescope", "content-cut-light", "defs",
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
