package draw2d

import (
	"image/color"

	"github.com/lafriks/go-svg"
	"github.com/lafriks/go-svg/renderer"

	"github.com/llgcode/draw2d"
)

// Draw the parsed SVG into the graphic context with the specified options.
func Draw(gc draw2d.GraphicContext, s *svg.Svg, opts ...renderer.RenderOption) {
	opt := renderer.Options(s, opts...)
	for _, svgp := range s.SvgPaths {
		drawTransformed(gc, svgp, opt)
	}
}

func drawTo(gc draw2d.GraphicContext, op svg.Operation, m svg.Matrix2D) {
	switch op := op.(type) {
	case svg.OpMoveTo:
		gc.Close()
		gc.BeginPath()
		t := m.MoveTo(op)
		gc.MoveTo(float64(t.X)/64, float64(t.Y)/64)
	case svg.OpLineTo:
		t := m.LineTo(op)
		gc.LineTo(float64(t.X)/64, float64(t.Y)/64)
	case svg.OpQuadTo:
		t1, t2 := m.QuadTo(op)
		gc.QuadCurveTo(float64(t1.X)/64, float64(t1.Y)/64, float64(t2.X)/64, float64(t2.Y)/64)
	case svg.OpCubicTo:
		t1, t2, t3 := m.CubicTo(op)
		gc.CubicCurveTo(float64(t1.X)/64, float64(t1.Y)/64, float64(t2.X)/64, float64(t2.Y)/64, float64(t3.X)/64, float64(t3.Y)/64)
	case svg.OpClose:
		gc.Close()
	}
}

func toLineCap(cap svg.CapMode) draw2d.LineCap {
	switch cap {
	case svg.ButtCap, svg.CubicCap, svg.QuadraticCap:
		return draw2d.ButtCap
	case svg.RoundCap:
		return draw2d.RoundCap
	case svg.SquareCap:
		return draw2d.SquareCap
	}
	return draw2d.ButtCap
}

func toLineJoin(join svg.JoinMode) draw2d.LineJoin {
	switch join {
	case svg.Arc, svg.Miter, svg.MiterClip:
		return draw2d.MiterJoin
	case svg.Round:
		return draw2d.RoundJoin
	case svg.Bevel:
		return draw2d.BevelJoin
	}
	return draw2d.MiterJoin
}

func toColor(c color.Color, opacity float64) color.Color {
	r, g, b, a := c.RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(uint32(float64(a)*opacity) >> 8)}
}

func toGradient(g svg.Gradient, opacity float64) color.Color {
	return toColor(svg.GetColor(g), opacity)
}

// drawTransformed draws the compiled SvgPath into the driver while applying transform t.
func drawTransformed(gc draw2d.GraphicContext, svgp svg.SvgPath, opt *renderer.RenderOptions) {
	m := svgp.Style.Transform.Mult(opt.Target)

	if svgp.Style.FillerColor != nil {
		var fr draw2d.FillRule
		if svgp.Style.UseNonZeroWinding {
			fr = draw2d.FillRuleWinding
		}
		gc.SetFillRule(fr)
		switch c := svgp.Style.FillerColor.(type) {
		case svg.PlainColor:
			gc.SetFillColor(toColor(c, svgp.Style.FillOpacity*opt.Opacity))
		case svg.Gradient:
			gc.SetFillColor(toGradient(c, svgp.Style.FillOpacity*opt.Opacity))
		}
	}
	if svgp.Style.LinerColor != nil {
		gc.SetLineCap(toLineCap(svgp.Style.Join.TrailLineCap))
		gc.SetLineJoin(toLineJoin(svgp.Style.Join.LineJoin))
		switch c := svgp.Style.LinerColor.(type) {
		case svg.PlainColor:
			gc.SetStrokeColor(toColor(c, svgp.Style.LineOpacity*opt.Opacity))
		case svg.Gradient:
			gc.SetStrokeColor(toGradient(c, svgp.Style.LineOpacity*opt.Opacity))
		}
		gc.SetLineWidth(svgp.Style.LineWidth * m.LineWidthScale())
		gc.SetLineDash(svgp.Style.Dash.Dash, svgp.Style.Dash.DashOffset)
	}

	for _, op := range svgp.Path {
		drawTo(gc, op, m)
	}

	if svgp.Style.FillerColor != nil {
		if svgp.Style.LinerColor != nil {
			gc.FillStroke()
		} else {
			gc.Fill()
		}
	} else if svgp.Style.LinerColor != nil {
		gc.Stroke()
	}
}
