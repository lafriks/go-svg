package svg

// SvgMask is an SVG element that defines a mask for the referenced elements.
type SvgMask struct {
	ID string

	X, Y float64
	W, H float64

	SvgPaths  []SvgPath
	Transform Matrix2D
}
