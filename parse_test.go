package svg

import (
	"strings"
	"testing"
)

func parseSvg(t *testing.T, svgPath string) {
	_, errSvg := ParseFile(svgPath, WarnErrorMode)
	if errSvg != nil {
		t.Error(errSvg)
	}
}

func TestLandscapeIcons(t *testing.T) {
	for _, p := range []string{
		"beach", "cape", "iceberg", "island",
		"mountains", "sea", "trees", "village",
	} {
		parseSvg(t, "testdata/landscapeIcons/"+p+".svg")
	}
}

func TestTestIcons(t *testing.T) {
	for _, p := range []string{
		"astronaut", "jupiter", "lander", "school-bus", "telescope", "content-cut-light", "defs",
		"24px",
	} {
		parseSvg(t, "testdata/testIcons/"+p+".svg")
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
		parseSvg(t, "testdata/"+p)
	}
}

func TestPercentages(t *testing.T) {
	parseSvg(t, "testdata/TestPercentages.svg")
}

func TestInvalidXML(t *testing.T) {
	_, err := Parse(strings.NewReader("dummy"), StrictErrorMode)
	if err == nil {
		t.Fatal("expected error on invalid input")
	}
}
