package main

import (
	"bytes"
	"image"
	"image/png"
	"os"

	"github.com/lafriks/go-svg"
	"github.com/lafriks/go-svg/renderer"
	rendr_rasterx "github.com/lafriks/go-svg/renderer/rasterx"

	"github.com/srwiley/rasterx"
)

func main() {
	s, err := svg.ParseFile("../../../testdata/TestShapes6.svg", svg.WarnErrorMode)
	if err != nil {
		panic(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	scanner := rasterx.NewScannerGV(256, 256, img, img.Bounds())

	gc := rasterx.NewDasher(256, 256, scanner)

	rendr_rasterx.Draw(gc, s, renderer.Target(0, 0, 256, 256))

	var b bytes.Buffer
	// Write the image into the buffer
	err = png.Encode(&b, img)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("example.png", b.Bytes(), 0o644)
	if err != nil {
		panic(err)
	}
}
