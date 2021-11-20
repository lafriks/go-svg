package gg

import (
	"github.com/lafriks/go-svg"
	"github.com/lafriks/go-svg/renderer"

	"github.com/fogleman/gg"
)

func ExampleDraw() {
	gc := gg.NewContext(256, 256)

	s, err := svg.ParseFile("../../testdata/landscapeIcons/beach.svg", svg.WarnErrorMode)
	if err != nil {
		panic(err)
	}

	Draw(gc, s, renderer.Target(0, 0, 256, 256))

	err = gc.SavePNG("TestShapes.png")
	if err != nil {
		panic(err)
	}
}
