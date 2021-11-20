package main

import (
	"image"

	"github.com/lafriks/go-svg"
	"github.com/lafriks/go-svg/renderer"
	randr_draw2d "github.com/lafriks/go-svg/renderer/draw2d"

	"github.com/llgcode/draw2d/draw2dimg"
)

func main() {
	dest := image.NewRGBA(image.Rect(0, 0, 256, 256))
	gc := draw2dimg.NewGraphicContext(dest)

	s, err := svg.ParseFile("../../../testdata/TestShapes.svg", svg.WarnErrorMode)
	if err != nil {
		panic(err)
	}

	randr_draw2d.Draw(gc, s, renderer.Target(0, 0, 256, 256))

	err = draw2dimg.SaveToPngFile("example.png", dest)
	if err != nil {
		panic(err)
	}
}
