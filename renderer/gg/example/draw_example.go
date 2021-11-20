package main

import (
	"github.com/lafriks/go-svg"
	"github.com/lafriks/go-svg/renderer"
	rendr_gg "github.com/lafriks/go-svg/renderer/gg"

	"github.com/fogleman/gg"
)

func main() {
	gc := gg.NewContext(256, 256)

	s, err := svg.ParseFile("../../../testdata/TestShapes.svg", svg.WarnErrorMode)
	if err != nil {
		panic(err)
	}

	rendr_gg.Draw(gc, s, renderer.Target(0, 0, 256, 256))

	err = gc.SavePNG("example.png")
	if err != nil {
		panic(err)
	}
}
