package svg

import (
	"encoding/xml"
	"errors"
	"strings"

	"golang.org/x/image/math/fixed"
)

func init() {
	// avoids cyclical static declaration
	// called on package initialization
	drawFuncs["use"] = useF
}

type svgFunc func(c *svgCursor, attrs []xml.Attr) error

var drawFuncs = map[string]svgFunc{
	"svg":            svgF,
	"g":              gF,
	"line":           lineF,
	"stop":           stopF,
	"rect":           rectF,
	"circle":         circleF,
	"ellipse":        circleF, // circleF handles ellipse also
	"polyline":       polylineF,
	"polygon":        polygonF,
	"path":           pathF,
	"desc":           descF,
	"defs":           defsF,
	"title":          titleF,
	"linearGradient": linearGradientF,
	"radialGradient": radialGradientF,
}

func svgF(c *svgCursor, attrs []xml.Attr) error {
	c.svg.ViewBox.X = 0
	c.svg.ViewBox.Y = 0
	c.svg.ViewBox.W = 0
	c.svg.ViewBox.H = 0
	var width, height float64
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "viewBox":
			err = c.getPoints(attr.Value)
			if len(c.points) != 4 {
				return errParamMismatch
			}
			c.svg.ViewBox.X = c.points[0]
			c.svg.ViewBox.Y = c.points[1]
			c.svg.ViewBox.W = c.points[2]
			c.svg.ViewBox.H = c.points[3]
		case "width":
			c.svg.Width = attr.Value
			width, err = parseBasicFloat(attr.Value)
		case "height":
			c.svg.Height = attr.Value
			height, err = parseBasicFloat(attr.Value)
		}
		if err != nil {
			return err
		}
	}
	if c.svg.ViewBox.W == 0 {
		c.svg.ViewBox.W = width
	}
	if c.svg.ViewBox.H == 0 {
		c.svg.ViewBox.H = height
	}
	return nil
}
func gF(*svgCursor, []xml.Attr) error { return nil } // g does nothing but push the style
func rectF(c *svgCursor, attrs []xml.Attr) error {
	var x, y, w, h, rx, ry float64
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "x":
			x, err = c.parseUnit(attr.Value, widthPercentage)
		case "y":
			y, err = c.parseUnit(attr.Value, heightPercentage)
		case "width":
			w, err = c.parseUnit(attr.Value, widthPercentage)
		case "height":
			h, err = c.parseUnit(attr.Value, heightPercentage)
		case "rx":
			rx, err = c.parseUnit(attr.Value, widthPercentage)
		case "ry":
			ry, err = c.parseUnit(attr.Value, heightPercentage)
		}
		if err != nil {
			return err
		}
	}
	if w == 0 || h == 0 {
		return nil
	}
	c.path.addRoundRect(x+c.curX, y+c.curY, w+x+c.curX, h+y+c.curY, rx, ry, 0)
	return nil
}

func circleF(c *svgCursor, attrs []xml.Attr) error {
	var cx, cy, rx, ry float64
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "cx":
			cx, err = c.parseUnit(attr.Value, widthPercentage)
		case "cy":
			cy, err = c.parseUnit(attr.Value, heightPercentage)
		case "r":
			rx, err = c.parseUnit(attr.Value, diagPercentage)
			ry = rx
		case "rx":
			rx, err = c.parseUnit(attr.Value, widthPercentage)
		case "ry":
			ry, err = c.parseUnit(attr.Value, heightPercentage)
		}
		if err != nil {
			return err
		}
	}
	if rx == 0 || ry == 0 { // not drawn, but not an error
		return nil
	}
	c.ellipseAt(cx+c.curX, cy+c.curY, rx, ry)
	return nil
}

func lineF(c *svgCursor, attrs []xml.Attr) error {
	var x1, x2, y1, y2 float64
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "x1":
			x1, err = c.parseUnit(attr.Value, widthPercentage)
		case "x2":
			x2, err = c.parseUnit(attr.Value, widthPercentage)
		case "y1":
			y1, err = c.parseUnit(attr.Value, heightPercentage)
		case "y2":
			y2, err = c.parseUnit(attr.Value, heightPercentage)
		}
		if err != nil {
			return err
		}
	}
	c.path.Start(fixed.Point26_6{
		X: fixed.Int26_6((x1 + c.curX) * 64),
		Y: fixed.Int26_6((y1 + c.curY) * 64),
	})
	c.path.Line(fixed.Point26_6{
		X: fixed.Int26_6((x2 + c.curX) * 64),
		Y: fixed.Int26_6((y2 + c.curY) * 64),
	})
	return nil
}

func polylineF(c *svgCursor, attrs []xml.Attr) error {
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "points":
			err = c.getPoints(attr.Value)
			if len(c.points)%2 != 0 {
				return errors.New("polygon has odd number of points")
			}
		}
		if err != nil {
			return err
		}
	}
	if len(c.points) > 4 {
		c.path.Start(fixed.Point26_6{
			X: fixed.Int26_6((c.points[0] + c.curX) * 64),
			Y: fixed.Int26_6((c.points[1] + c.curY) * 64),
		})
		for i := 2; i < len(c.points)-1; i += 2 {
			c.path.Line(fixed.Point26_6{
				X: fixed.Int26_6((c.points[i] + c.curX) * 64),
				Y: fixed.Int26_6((c.points[i+1] + c.curY) * 64),
			})
		}
	}
	return nil
}

func polygonF(c *svgCursor, attrs []xml.Attr) error {
	err := polylineF(c, attrs)
	if len(c.points) > 4 {
		c.path.Stop(true)
	}
	return err
}

func pathF(c *svgCursor, attrs []xml.Attr) error {
	var err error
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "d":
			err = c.compilePath(attr.Value)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func descF(c *svgCursor, attrs []xml.Attr) error {
	c.inDescText = true
	c.svg.Descriptions = append(c.svg.Descriptions, "")
	return nil
}

func titleF(c *svgCursor, attrs []xml.Attr) error {
	c.inTitleText = true
	c.svg.Titles = append(c.svg.Titles, "")
	return nil
}

func defsF(c *svgCursor, attrs []xml.Attr) error {
	c.inDefs = true
	return nil
}

func linearGradientF(c *svgCursor, attrs []xml.Attr) error {
	var err error
	c.inGrad = true
	// interpretation of percentage in direction depends
	// on gradientUnits: we first store the string values
	// and resolve them in a second pass
	directionStrings := [4]string{"0%", "0%", "100%", "0"} // default value
	c.grad = &Gradient{Bounds: c.svg.ViewBox, Matrix: Identity}
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "id":
			id := attr.Value
			if len(id) >= 0 {
				c.svg.grads[id] = c.grad
			} else {
				return errZeroLengthID
			}
		case "x1":
			directionStrings[0] = attr.Value
		case "y1":
			directionStrings[1] = attr.Value
		case "x2":
			directionStrings[2] = attr.Value
		case "y2":
			directionStrings[3] = attr.Value
		default:
			err = c.readGradAttr(attr)
		}
		if err != nil {
			return err
		}
	}
	// now we can resolve percentages
	bbox := Bounds{W: 1, H: 1} // default is ObjectBoundingBox
	if c.grad.Units == UserSpaceOnUse {
		bbox = c.grad.Bounds
	}
	var direction Linear
	direction[0], err = bbox.resolveUnit(directionStrings[0], widthPercentage)
	if err != nil {
		return err
	}
	direction[1], err = bbox.resolveUnit(directionStrings[1], heightPercentage)
	if err != nil {
		return err
	}
	direction[2], err = bbox.resolveUnit(directionStrings[2], widthPercentage)
	if err != nil {
		return err
	}
	direction[3], err = bbox.resolveUnit(directionStrings[3], heightPercentage)
	if err != nil {
		return err
	}
	c.grad.Direction = direction
	return nil
}

func radialGradientF(c *svgCursor, attrs []xml.Attr) error {
	c.inGrad = true
	c.grad = &Gradient{Bounds: c.svg.ViewBox, Matrix: Identity}
	var setFx, setFy bool
	var err error
	directionStrings := [6]string{"50%", "50%", "50%", "50%", "50%", "50%"} // default values
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "id":
			id := attr.Value
			if len(id) >= 0 {
				c.svg.grads[id] = c.grad
			} else {
				return errZeroLengthID
			}
		case "cx":
			directionStrings[0] = attr.Value
		case "cy":
			directionStrings[1] = attr.Value
		case "fx":
			setFx = true
			directionStrings[2] = attr.Value
		case "fy":
			setFy = true
			directionStrings[3] = attr.Value
		case "r":
			directionStrings[4] = attr.Value
		case "fr":
			directionStrings[5] = attr.Value
		default:
			err = c.readGradAttr(attr)
		}
		if err != nil {
			return err
		}
	}
	if !setFx { // set fx to cx by default
		directionStrings[2] = directionStrings[0]
	}
	if !setFy { // set fy to cy by default
		directionStrings[3] = directionStrings[1]
	}

	// now we can resolve percentages
	bbox := Bounds{W: 1, H: 1} // default is ObjectBoundingBox
	if c.grad.Units == UserSpaceOnUse {
		bbox = c.grad.Bounds
	}
	var direction Radial
	direction[0], err = bbox.resolveUnit(directionStrings[0], widthPercentage)
	if err != nil {
		return err
	}
	direction[1], err = bbox.resolveUnit(directionStrings[1], heightPercentage)
	if err != nil {
		return err
	}
	direction[2], err = bbox.resolveUnit(directionStrings[2], widthPercentage)
	if err != nil {
		return err
	}
	direction[3], err = bbox.resolveUnit(directionStrings[3], heightPercentage)
	if err != nil {
		return err
	}
	direction[4], err = bbox.resolveUnit(directionStrings[4], diagPercentage)
	if err != nil {
		return err
	}
	direction[5], err = bbox.resolveUnit(directionStrings[5], diagPercentage)
	if err != nil {
		return err
	}

	c.grad.Direction = direction
	return nil
}

func stopF(c *svgCursor, attrs []xml.Attr) error {
	var err error
	if c.inGrad {
		stop := GradStop{Opacity: 1.0}
		for _, attr := range attrs {
			switch attr.Name.Local {
			case "offset":
				stop.Offset, err = readFraction(attr.Value)
			case "stop-color":
				// todo: add current color inherit
				var optColor optionnalColor
				optColor, err = parseSVGColor(attr.Value)
				stop.StopColor = optColor.asColor()
			case "stop-opacity":
				stop.Opacity, err = parseBasicFloat(attr.Value)
			}
			if err != nil {
				return err
			}
		}
		c.grad.Stops = append(c.grad.Stops, stop)
	}
	return nil
}

func useF(c *svgCursor, attrs []xml.Attr) error {
	var (
		href string
		x, y float64
		err  error
	)
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "href":
			href = attr.Value
		case "x":
			x, err = c.parseUnit(attr.Value, widthPercentage)
		case "y":
			y, err = c.parseUnit(attr.Value, heightPercentage)
		}
		if err != nil {
			return err
		}
	}
	c.curX, c.curY = x, y
	defer func() {
		c.curX, c.curY = 0, 0
	}()
	if href == "" {
		return errors.New("only use tags with href is supported")
	}
	if !strings.HasPrefix(href, "#") {
		return errors.New("only the ID CSS selector is supported")
	}
	defs, ok := c.svg.defs[href[1:]]
	if !ok {
		return errors.New("href ID in use statement was not found in saved defs")
	}
	for _, def := range defs {
		if def.Tag == "endg" {
			// pop style
			c.styleStack = c.styleStack[:len(c.styleStack)-1]
			continue
		}
		if err = c.pushStyle(def.Attrs); err != nil {
			return err
		}
		df, ok := drawFuncs[def.Tag]
		if !ok {
			errStr := "Cannot process svg element " + def.Tag
			return c.handleError(errStr)
		}
		if err := df(c, def.Attrs); err != nil {
			return err
		}
		if def.Tag != "g" {
			// pop style
			c.styleStack = c.styleStack[:len(c.styleStack)-1]
		}
	}
	return nil
}
