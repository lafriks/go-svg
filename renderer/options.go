package renderer

import (
	"github.com/lafriks/go-svg"
)

type RenderOptions struct {
	// Opacity is the opacity of the rendered image.
	Opacity float64
	// Target is the rectangle to render the image within.
	Target svg.Matrix2D
}

// RenderOption is a interface for renderer options.
type RenderOption interface {
	apply(s *svg.Svg, r *RenderOptions)
}

// Opacity is an option to specify the opacity of the rendered image.
type Opacity float64

func (o Opacity) apply(s *svg.Svg, r *RenderOptions) {
	r.Opacity = float64(o)
}

type targetOption struct {
	X, Y, W, H float64
}

func (o targetOption) apply(s *svg.Svg, r *RenderOptions) {
	scaleW := o.W / s.ViewBox.W
	scaleH := o.H / s.ViewBox.H
	r.Target = svg.Identity.Translate(o.X-s.ViewBox.X, o.Y-s.ViewBox.Y).Scale(scaleW, scaleH)
}

// Target specifies the rectangle to draw within.
func Target(x, y, w, h float64) RenderOption {
	return targetOption{X: x, Y: y, W: w, H: h}
}

// Options apply the options.
func Options(s *svg.Svg, opts ...RenderOption) *RenderOptions {
	r := &RenderOptions{
		Opacity: 1,
	}
	for _, o := range opts {
		o.apply(s, r)
	}
	return r
}
