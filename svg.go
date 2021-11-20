// Provides parsing and rendering of SVG images.
// SVG files are parsed into an abstract representation,
// which can then be consumed by painting drivers.
package svg

import (
	"encoding/xml"
	"errors"
	"io"
	"os"

	"golang.org/x/net/html/charset"
)

// PathStyle holds the state of the SVG style
type PathStyle struct {
	FillOpacity, LineOpacity float64
	LineWidth                float64
	UseNonZeroWinding        bool

	Join                    JoinOptions
	Dash                    DashOptions
	FillerColor, LinerColor Pattern // either PlainColor or Gradient

	Masks []string

	Transform Matrix2D // current transform
}

// SvgPath binds a style to a path
type SvgPath struct {
	Path  Path
	Style PathStyle
}

// Bounds defines a bounding box, such as a viewport
// or a path extent.
type Bounds struct{ X, Y, W, H float64 }

// Svg holds data from parsed SVGs.
// See the `Draw` methods to use it.
type Svg struct {
	ViewBox      Bounds
	Titles       []string // Title elements collect here
	Descriptions []string // Description elements collect here
	SvgPaths     []SvgPath
	Transform    Matrix2D
	SvgMasks     map[string]*SvgMask

	Width, Height string // top level width and height attributes

	grads map[string]*Gradient
	defs  map[string][]definition
}

// Parse reads the Icon from the given io.Reader
// This only supports a sub-set of SVG, but
// is enough to draw many svgs. errMode determines if the svg ignores, errors out, or logs a warning
// if it does not handle an element found in the svg file.
func Parse(stream io.Reader, errMode ErrorMode) (*Svg, error) {
	svg := &Svg{
		defs:      make(map[string][]definition),
		grads:     make(map[string]*Gradient),
		SvgMasks:  make(map[string]*SvgMask),
		Transform: Identity,
	}
	svgCursor := &svgCursor{styleStack: []PathStyle{DefaultStyle}, svg: svg}
	svgCursor.errorMode = errMode
	decoder := xml.NewDecoder(stream)
	decoder.CharsetReader = charset.NewReaderLabel
	seenTag := false
	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				if !seenTag {
					return nil, errors.New("invalid svg xml svg")
				}
				break
			}
			return svg, err
		}
		// Inspect the type of the XML token
		switch se := t.(type) {
		case xml.StartElement:
			seenTag = true
			// Reads all recognized style attributes from the start element
			// and places it on top of the styleStack
			err = svgCursor.pushStyle(se.Attr)
			if err != nil {
				return svg, err
			}
			err = svgCursor.readStartElement(se)
			if err != nil {
				return svg, err
			}
		case xml.EndElement:
			// pop style
			svgCursor.styleStack = svgCursor.styleStack[:len(svgCursor.styleStack)-1]
			switch se.Name.Local {
			case "g":
				if svgCursor.inDefs {
					svgCursor.currentDef = append(svgCursor.currentDef, definition{
						Tag: "endg",
					})
				}
			case "mask":
				if svgCursor.mask != nil {
					svgCursor.svg.SvgMasks[svgCursor.mask.ID] = svgCursor.mask
					svgCursor.mask = nil
				}
				svgCursor.inMask = false
			case "title":
				svgCursor.inTitleText = false
			case "desc":
				svgCursor.inDescText = false
			case "defs":
				if len(svgCursor.currentDef) > 0 {
					svgCursor.svg.defs[svgCursor.currentDef[0].ID] = svgCursor.currentDef
					svgCursor.currentDef = make([]definition, 0)
				}
				svgCursor.inDefs = false
			case "radialGradient", "linearGradient":
				svgCursor.inGrad = false
			}
		case xml.CharData:
			if svgCursor.inTitleText {
				svg.Titles[len(svg.Titles)-1] += string(se)
			}
			if svgCursor.inDescText {
				svg.Descriptions[len(svg.Descriptions)-1] += string(se)
			}
		}
	}
	return svg, nil
}

// ParseFile reads the SVG from the named file
// This only supports a sub-set of SVG, but
// is enough to draw many svgs. errMode determines if the svg ignores, errors out, or logs a warning
// if it does not handle an element found in the svg file.
func ParseFile(name string, errMode ErrorMode) (*Svg, error) {
	fin, errf := os.Open(name)
	if errf != nil {
		return nil, errf
	}
	defer fin.Close()
	return Parse(fin, errMode)
}
