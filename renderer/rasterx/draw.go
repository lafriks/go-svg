package rasterx

import (
	"github.com/lafriks/go-svg"
	"github.com/lafriks/go-svg/renderer"

	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

// Draw the parsed SVG into the graphic context with the specified options.
func Draw(gc *rasterx.Dasher, s *svg.Svg, opts ...renderer.RenderOption) {
	opt := renderer.Options(s, opts...)
	for _, svgp := range s.SvgPaths {
		drawTransformed(gc, svgp, opt)
	}
}

func drawToStroker(gc *rasterx.Stroker, op svg.Operation, m svg.Matrix2D) {
	switch op := op.(type) {
	case svg.OpMoveTo:
		gc.Stop(false)
		gc.Start(m.MoveTo(op))
	case svg.OpLineTo:
		gc.Line(m.LineTo(op))
	case svg.OpQuadTo:
		gc.QuadBezier(m.QuadTo(op))
	case svg.OpCubicTo:
		gc.CubeBezier(m.CubicTo(op))
	case svg.OpClose:
		gc.Stop(true)
	}
}

func drawToFiller(gc *rasterx.Filler, op svg.Operation, m svg.Matrix2D) {
	switch op := op.(type) {
	case svg.OpMoveTo:
		gc.Stop(false)
		gc.Start(m.MoveTo(op))
	case svg.OpLineTo:
		gc.Line(m.LineTo(op))
	case svg.OpQuadTo:
		gc.QuadBezier(m.QuadTo(op))
	case svg.OpCubicTo:
		gc.CubeBezier(m.CubicTo(op))
	case svg.OpClose:
		gc.Stop(true)
	}
}

func toLineGap(gap svg.GapMode) rasterx.GapFunc {
	switch gap {
	case svg.FlatGap:
		return rasterx.FlatGap
	case svg.RoundGap:
		return rasterx.RoundGap
	case svg.CubicGap:
		return rasterx.CubicGap
	case svg.QuadraticGap:
		return rasterx.QuadraticGap
	}
	return rasterx.FlatGap
}

func toLineCap(cap svg.CapMode, def rasterx.CapFunc) rasterx.CapFunc {
	if cap == svg.NilCap {
		return def
	}
	switch cap {
	case svg.ButtCap:
		return rasterx.ButtCap
	case svg.CubicCap:
		return rasterx.CubicCap
	case svg.QuadraticCap:
		return rasterx.QuadraticCap
	case svg.RoundCap:
		return rasterx.RoundCap
	case svg.SquareCap:
		return rasterx.SquareCap
	}
	return rasterx.ButtCap
}

func toLineJoin(join svg.JoinMode) rasterx.JoinMode {
	switch join {
	case svg.Arc:
		return rasterx.Arc
	case svg.ArcClip:
		return rasterx.ArcClip
	case svg.Miter:
		return rasterx.Miter
	case svg.MiterClip:
		return rasterx.MiterClip
	case svg.Round:
		return rasterx.Round
	case svg.Bevel:
		return rasterx.Bevel
	}
	return rasterx.Miter
}

func toRasterxGradient(grad svg.Gradient) rasterx.Gradient {
	var (
		points   [5]float64
		isRadial bool
	)
	switch dir := grad.Direction.(type) {
	case svg.Linear:
		points[0], points[1], points[2], points[3] = dir[0], dir[1], dir[2], dir[3]
		isRadial = false
	case svg.Radial:
		points[0], points[1], points[2], points[3], points[4], _ = dir[0], dir[1], dir[2], dir[3], dir[4], dir[5] // in rasterx fr is ignored
		isRadial = true
	}
	stops := make([]rasterx.GradStop, len(grad.Stops))
	for i := range grad.Stops {
		stops[i] = rasterx.GradStop(grad.Stops[i])
	}
	return rasterx.Gradient{
		Points:   points,
		Stops:    stops,
		Bounds:   grad.Bounds,
		Matrix:   rasterx.Matrix2D(grad.Matrix),
		Spread:   rasterx.SpreadMethod(grad.Spread),
		Units:    rasterx.GradientUnits(grad.Units),
		IsRadial: isRadial,
	}
}

// drawTransformed draws the compiled SvgPath into the driver while applying transform t.
func drawTransformed(gc *rasterx.Dasher, svgp svg.SvgPath, opt *renderer.RenderOptions) {
	m := svgp.Style.Transform.Mult(opt.Target)

	if svgp.Style.FillerColor != nil {
		filler := &gc.Filler
		filler.Clear()
		filler.SetWinding(svgp.Style.UseNonZeroWinding)

		for _, op := range svgp.Path {
			drawToFiller(filler, op, m)
		}
		filler.Stop(false)

		switch color := svgp.Style.FillerColor.(type) {
		case svg.PlainColor:
			filler.SetColor(rasterx.ApplyOpacity(color, svgp.Style.FillOpacity*opt.Opacity))
		case svg.Gradient:
			_ = color.ApplyPathExtent(filler.GetPathExtent())
			g := toRasterxGradient(color)
			filler.SetColor(g.GetColorFunction(svgp.Style.FillOpacity * opt.Opacity))
		}
		filler.Draw()
	}
	if svgp.Style.LinerColor != nil {
		scale := (m.A + m.D) / 2
		stroker := &gc.Stroker
		stroker.Clear()
		stroker.SetStroke(
			fixed.Int26_6(svgp.Style.LineWidth*64*scale),
			svgp.Style.Join.MiterLimit,
			toLineCap(svgp.Style.Join.LeadLineCap, toLineCap(svgp.Style.Join.TrailLineCap, rasterx.ButtCap)),
			toLineCap(svgp.Style.Join.TrailLineCap, rasterx.ButtCap),
			toLineGap(svgp.Style.Join.LineGap),
			toLineJoin(svgp.Style.Join.LineJoin))

		for _, op := range svgp.Path {
			drawToStroker(stroker, op, m)
		}
		stroker.Stop(false)

		switch color := svgp.Style.LinerColor.(type) {
		case svg.PlainColor:
			stroker.SetColor(rasterx.ApplyOpacity(color, svgp.Style.LineOpacity*opt.Opacity))
		case svg.Gradient:
			_ = color.ApplyPathExtent(stroker.GetPathExtent())
			g := toRasterxGradient(color)
			stroker.SetColor(g.GetColorFunction(svgp.Style.LineOpacity * opt.Opacity))
		}
		stroker.Draw()
	}
}
