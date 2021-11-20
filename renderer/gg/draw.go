package gg

import (
	"image/color"

	"github.com/lafriks/go-svg"
	"github.com/lafriks/go-svg/renderer"

	"github.com/fogleman/gg"
)

// Draw the parsed SVG into the graphic context with the specified options.
func Draw(gc *gg.Context, s *svg.Svg, opts ...renderer.RenderOption) {
	opt := renderer.Options(s, opts...)
	for _, svgp := range s.SVGPaths {
		drawTransformed(gc, svgp, opt)
	}
}

func drawTo(gc *gg.Context, op svg.Operation, m svg.Matrix2D) {
	switch op := op.(type) {
	case svg.OpMoveTo:
		gc.ClosePath()
		gc.NewSubPath()
		t := m.MoveTo(op)
		gc.MoveTo(float64(t.X)/64, float64(t.Y)/64)
	case svg.OpLineTo:
		t := m.LineTo(op)
		gc.LineTo(float64(t.X)/64, float64(t.Y)/64)
	case svg.OpQuadTo:
		t1, t2 := m.QuadTo(op)
		gc.QuadraticTo(float64(t1.X)/64, float64(t1.Y)/64, float64(t2.X)/64, float64(t2.Y)/64)
	case svg.OpCubicTo:
		t1, t2, t3 := m.CubicTo(op)
		gc.CubicTo(float64(t1.X)/64, float64(t1.Y)/64, float64(t2.X)/64, float64(t2.Y)/64, float64(t3.X)/64, float64(t3.Y)/64)
	case svg.OpClose:
		gc.ClosePath()
	}
}

func toLineCap(cap svg.CapMode) gg.LineCap {
	switch cap {
	case svg.ButtCap, svg.CubicCap, svg.QuadraticCap:
		return gg.LineCapButt
	case svg.RoundCap:
		return gg.LineCapRound
	case svg.SquareCap:
		return gg.LineCapSquare
	}
	return gg.LineCapButt
}

func toLineJoin(join svg.JoinMode) gg.LineJoin {
	switch join {
	case svg.Arc, svg.Miter, svg.MiterClip:
		return gg.LineJoinBevel
	case svg.Round:
		return gg.LineJoinRound
	case svg.Bevel:
		return gg.LineJoinBevel
	}
	return gg.LineJoinBevel
}

func toColor(c color.Color, opacity float64) color.Color {
	r, g, b, a := c.RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(uint32(float64(a)*opacity) >> 8)}
}

func toGradient(g svg.Gradient, opacity float64) gg.Gradient {
	var grad gg.Gradient
	switch dir := g.Direction.(type) {
	case svg.Linear:
		// x1, y1, x2, y2
		grad = gg.NewLinearGradient(dir[0], dir[1], dir[2], dir[3])
	case svg.Radial:
		// cx, cy, fx, fy, r, fr
		grad = gg.NewRadialGradient(dir[0], dir[1], dir[4], dir[2], dir[3], dir[5])
	}
	for _, stop := range g.Stops {
		grad.AddColorStop(stop.Offset, toColor(stop.StopColor, stop.Opacity*opacity))
	}
	return grad
}

// drawTransformed draws the compiled SvgPath into the driver while applying transform t.
func drawTransformed(gc *gg.Context, svgp svg.SvgPath, opt *renderer.RenderOptions) {
	m := svgp.Style.Transform.Mult(opt.Target)

	if svgp.Style.FillerColor != nil {
		if svgp.Style.UseNonZeroWinding {
			gc.SetFillRuleWinding()
		} else {
			gc.SetFillRuleEvenOdd()
		}
		switch c := svgp.Style.FillerColor.(type) {
		case svg.PlainColor:
			gc.SetFillStyle(gg.NewSolidPattern(toColor(c, svgp.Style.FillOpacity*opt.Opacity)))
		case svg.Gradient:
			gc.SetFillStyle(toGradient(c, svgp.Style.FillOpacity*opt.Opacity))
		}
	}
	if svgp.Style.LinerColor != nil {
		gc.SetLineCap(toLineCap(svgp.Style.Join.TrailLineCap))
		gc.SetLineJoin(toLineJoin(svgp.Style.Join.LineJoin))
		switch c := svgp.Style.LinerColor.(type) {
		case svg.PlainColor:
			gc.SetColor(toColor(c, svgp.Style.LineOpacity*opt.Opacity))
		case svg.Gradient:
			gc.SetColor(toColor(svg.GetColor(c), svgp.Style.LineOpacity*opt.Opacity))
		}
		gc.SetLineWidth(svgp.Style.LineWidth)
		gc.SetDash(svgp.Style.Dash.Dash...)
		gc.SetDashOffset(svgp.Style.Dash.DashOffset)
	}

	for _, op := range svgp.Path {
		drawTo(gc, op, m)
	}

	if svgp.Style.FillerColor != nil {
		if svgp.Style.LinerColor != nil {
			gc.FillPreserve()
		} else {
			gc.Fill()
		}
	}
	if svgp.Style.LinerColor != nil {
		gc.Stroke()
	}
}
