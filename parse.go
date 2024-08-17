package svg

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"golang.org/x/image/math/fixed"
)

type (
	// svgCursor is used while parsing SVG files
	svgCursor struct {
		pathCursor
		svg                                             *Svg
		styleStack                                      []PathStyle
		grad                                            *Gradient
		inTitleText, inDescText, inGrad, inDefs, inMask bool
		currentDef                                      []definition
		mask                                            *SvgMask
	}

	// definition is used to store what's given in a def tag
	definition struct {
		ID, Tag string
		Attrs   []xml.Attr
	}
)

func fToFixed(f float64) fixed.Int26_6 {
	return fixed.Int26_6(f * 64)
}

// treat the error according to the errorMode
func (c *svgCursor) handleError(originFmt string, args ...interface{}) error {
	formatted := fmt.Sprintf(originFmt, args...)
	if c.errorMode == StrictErrorMode {
		return errors.New(formatted)
	} else if c.errorMode == WarnErrorMode {
		log.Println(formatted) // then return nil
	}
	return nil
}

func (c *svgCursor) readTransformAttr(m1 Matrix2D, k string) (Matrix2D, error) {
	ln := len(c.points)
	switch k {
	case "rotate":
		if ln == 1 {
			m1 = m1.Rotate(c.points[0] * math.Pi / 180)
		} else if ln == 3 {
			m1 = m1.Translate(c.points[1], c.points[2]).
				Rotate(c.points[0]*math.Pi/180).
				Translate(-c.points[1], -c.points[2])
		} else {
			return m1, errParamMismatch
		}
	case "translate":
		if ln == 1 {
			m1 = m1.Translate(c.points[0], c.points[0])
		} else if ln == 2 {
			m1 = m1.Translate(c.points[0], c.points[1])
		} else {
			return m1, errParamMismatch
		}
	case "skewx":
		if ln == 1 {
			m1 = m1.SkewX(c.points[0] * math.Pi / 180)
		} else {
			return m1, errParamMismatch
		}
	case "skewy":
		if ln == 1 {
			m1 = m1.SkewY(c.points[0] * math.Pi / 180)
		} else {
			return m1, errParamMismatch
		}
	case "scale":
		if ln == 1 {
			m1 = m1.Scale(c.points[0], c.points[0])
		} else if ln == 2 {
			m1 = m1.Scale(c.points[0], c.points[1])
		} else {
			return m1, errParamMismatch
		}
	case "matrix":
		if ln == 6 {
			m1 = m1.Mult(Matrix2D{
				A: c.points[0],
				B: c.points[1],
				C: c.points[2],
				D: c.points[3],
				E: c.points[4],
				F: c.points[5],
			})
		} else {
			return m1, errParamMismatch
		}
	default:
		return m1, errParamMismatch
	}
	return m1, nil
}

func (c *svgCursor) parseTransform(v string) (Matrix2D, error) {
	ts := strings.Split(v, ")")
	m1 := c.styleStack[len(c.styleStack)-1].Transform
	// From the docs at https://devdoc.net/web/developer.mozilla.org/en-US/docs/Web/SVG/Attribute/transform.html:
	// The items in the transform list are separated by whitespace and/or commas, and are applied from right to left.
	for i := len(ts) - 1; i >= 0; i-- {
		t := strings.TrimSpace(ts[i])
		if len(t) == 0 {
			continue
		}
		d := strings.Split(t, "(")
		if len(d) != 2 || len(d[1]) < 1 {
			return m1, errParamMismatch // badly formed transformation
		}
		err := c.getPoints(d[1])
		if err != nil {
			return m1, err
		}
		m1, err = c.readTransformAttr(m1, strings.ToLower(strings.TrimSpace(d[0])))
		if err != nil {
			return m1, err
		}
	}
	return m1, nil
}

func (c *svgCursor) parseSelector(v string) (string, error) {
	if len(v) == 0 || v == "none" {
		return "", nil
	}
	if strings.HasPrefix(v, "url(") {
		i := strings.Index(v, ")")
		if i < 0 {
			return "", errParamMismatch
		}
		v = v[4:i]
		if !strings.HasPrefix(v, "#") {
			return "", fmt.Errorf("unsupported url selector: %s", v)
		}
		return v[1:], nil
	}
	return "", fmt.Errorf("unsupported selector: %s", v)
}

func (c *svgCursor) readStyleAttr(curStyle *PathStyle, k, v string) error {
	switch k {
	case "fill":
		gradient, ok := c.readGradURL(v, curStyle.FillerColor)
		if ok {
			curStyle.FillerColor = gradient
			break
		}
		optCol, err := parseSVGColor(v)
		curStyle.FillerColor = optCol.asPattern()
		return err
	case "fill-rule":
		switch v {
		case "evenodd":
			curStyle.UseNonZeroWinding = false
		case "nonzero":
			curStyle.UseNonZeroWinding = true
		default:
			return c.handleError("unsupported value '%s' for <fill-rule>", v)
		}
	case "stroke":
		gradient, ok := c.readGradURL(v, curStyle.LinerColor)
		if ok {
			curStyle.LinerColor = gradient
			break
		}
		optCol, errc := parseSVGColor(v)
		if errc != nil {
			return errc
		}
		curStyle.LinerColor = optCol.asPattern()
	case "stroke-linegap":
		switch v {
		case "flat":
			curStyle.Join.LineGap = FlatGap
		case "round":
			curStyle.Join.LineGap = RoundGap
		case "cubic":
			curStyle.Join.LineGap = CubicGap
		case "quadratic":
			curStyle.Join.LineGap = QuadraticGap
		default:
			return c.handleError("unsupported value '%s' for <stroke-linegap>", v)
		}
	case "stroke-leadlinecap":
		switch v {
		case "butt":
			curStyle.Join.LeadLineCap = ButtCap
		case "round":
			curStyle.Join.LeadLineCap = RoundCap
		case "square":
			curStyle.Join.LeadLineCap = SquareCap
		case "cubic":
			curStyle.Join.LeadLineCap = CubicCap
		case "quadratic":
			curStyle.Join.LeadLineCap = QuadraticCap
		default:
			return c.handleError("unsupported value '%s' for <stroke-leadlinecap>", v)
		}
	case "stroke-linecap":
		switch v {
		case "butt":
			curStyle.Join.TrailLineCap = ButtCap
		case "round":
			curStyle.Join.TrailLineCap = RoundCap
		case "square":
			curStyle.Join.TrailLineCap = SquareCap
		case "cubic":
			curStyle.Join.TrailLineCap = CubicCap
		case "quadratic":
			curStyle.Join.TrailLineCap = QuadraticCap
		default:
			return c.handleError("unsupported value '%s' for <stroke-linecap>", v)
		}
	case "stroke-linejoin":
		switch v {
		case "miter":
			curStyle.Join.LineJoin = Miter
		case "miter-clip":
			curStyle.Join.LineJoin = MiterClip
		case "arc-clip":
			curStyle.Join.LineJoin = ArcClip
		case "round":
			curStyle.Join.LineJoin = Round
		case "arc":
			curStyle.Join.LineJoin = Arc
		case "bevel":
			curStyle.Join.LineJoin = Bevel
		default:
			return c.handleError("unsupported value '%s' for <stroke-linejoin>", v)
		}
	case "stroke-miterlimit":
		mLimit, err := parseBasicFloat(v)
		if err != nil {
			return err
		}
		curStyle.Join.MiterLimit = fToFixed(mLimit)
	case "stroke-width":
		width, err := c.parseUnit(v, widthPercentage)
		if err != nil {
			return err
		}
		curStyle.LineWidth = width
	case "stroke-dashoffset":
		dashOffset, err := c.parseUnit(v, diagPercentage)
		if err != nil {
			return err
		}
		curStyle.Dash.DashOffset = dashOffset
	case "stroke-dasharray":
		if v != "none" {
			dashes := splitOnCommaOrSpace(v)
			dList := make([]float64, len(dashes))
			for i, dstr := range dashes {
				d, err := c.parseUnit(strings.TrimSpace(dstr), diagPercentage)
				if err != nil {
					return err
				}
				dList[i] = d
			}
			curStyle.Dash.Dash = dList
			break
		}
	case "opacity", "stroke-opacity", "fill-opacity":
		op, err := parseBasicFloat(v)
		if err != nil {
			return err
		}
		if k != "stroke-opacity" {
			curStyle.FillOpacity *= op
		}
		if k != "fill-opacity" {
			curStyle.LineOpacity *= op
		}
	case "transform":
		m, err := c.parseTransform(v)
		if err != nil {
			return err
		}
		curStyle.Transform = m
	case "mask":
		id, err := c.parseSelector(v)
		if err != nil {
			return err
		}
		curStyle.Masks = append(curStyle.Masks, id)
	}
	return nil
}

// pushStyle parses the style element, and push it on the style stack. Only color and opacity are supported
// for fill. Note that this parses both the contents of a style attribute plus
// direct fill and opacity attributes.
func (c *svgCursor) pushStyle(attrs []xml.Attr) error {
	var pairs []string
	for _, attr := range attrs {
		switch strings.ToLower(attr.Name.Local) {
		case "style":
			pairs = append(pairs, strings.Split(attr.Value, ";")...)
		default:
			pairs = append(pairs, attr.Name.Local+":"+attr.Value)
		}
	}
	// Make a copy of the top style
	curStyle := c.styleStack[len(c.styleStack)-1]
	for _, pair := range pairs {
		kv := strings.Split(pair, ":")
		if len(kv) >= 2 {
			k := strings.ToLower(kv[0])
			k = strings.TrimSpace(k)
			v := strings.TrimSpace(kv[1])
			err := c.readStyleAttr(&curStyle, k, v)
			if err != nil {
				return err
			}
		}
	}
	c.styleStack = append(c.styleStack, curStyle) // Push style onto stack
	return nil
}

// splitOnCommaOrSpace returns a list of strings after splitting the input on comma and space delimiters
func splitOnCommaOrSpace(s string) []string {
	return strings.FieldsFunc(s,
		func(r rune) bool {
			return r == ',' || r == ' '
		})
}

func (c *svgCursor) readStartElement(se xml.StartElement) (err error) {
	var skipDef bool
	if se.Name.Local == "radialGradient" || se.Name.Local == "linearGradient" || c.inGrad {
		skipDef = true
	}
	if c.inDefs && !skipDef {
		ID := ""
		for _, attr := range se.Attr {
			if attr.Name.Local == "id" {
				ID = attr.Value
			}
		}
		if ID != "" && len(c.currentDef) > 0 {
			c.svg.defs[c.currentDef[0].ID] = c.currentDef
			c.currentDef = make([]definition, 0)
		}
		c.currentDef = append(c.currentDef, definition{
			ID:    ID,
			Tag:   se.Name.Local,
			Attrs: se.Attr,
		})
		return nil
	}
	df, ok := drawFuncs[se.Name.Local]
	if !ok {
		errStr := "Cannot process svg element " + se.Name.Local
		if c.errorMode == StrictErrorMode {
			return errors.New(errStr)
		} else if c.errorMode == WarnErrorMode {
			log.Println(errStr)
		}
		return nil
	}
	err = df(c, se.Attr)

	if len(c.path) > 0 {
		// The svgCursor parsed a path from the xml element
		pathCopy := append(Path{}, c.path...)
		if c.inMask && c.mask != nil {
			c.mask.SvgPaths = append(c.mask.SvgPaths,
				SvgPath{Path: pathCopy, Style: c.styleStack[len(c.styleStack)-1]})
		} else if !c.inMask {
			c.svg.SvgPaths = append(c.svg.SvgPaths,
				SvgPath{Path: pathCopy, Style: c.styleStack[len(c.styleStack)-1]})
		}
		c.path = c.path[:0]
	}
	return
}
