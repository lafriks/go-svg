package svg

// Implements SVG style matrix transformations.
// https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/transform

import (
	"math"

	"golang.org/x/image/math/fixed"
)

// Matrix2D represents an SVG style matrix
type Matrix2D struct {
	A, B, C, D, E, F float64
}

// matrix3 is a full 3x3 float64 matrix
// used for inverting
type matrix3 [9]float64

func otherPair(i int) (a, b int) {
	switch i {
	case 0:
		a, b = 1, 2
	case 1:
		a, b = 0, 2
	case 2:
		a, b = 0, 1
	}
	return
}

func (m *matrix3) coFact(i, j int) float64 {
	ai, bi := otherPair(i)
	aj, bj := otherPair(j)
	a, b, c, d := m[ai+aj*3], m[bi+bj*3], m[ai+bj*3], m[bi+aj*3]
	return a*b - c*d
}

func (m *matrix3) invert() *matrix3 {
	var cofact matrix3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			sign := float64(1 - (i+j%2)%2*2) // "checkerboard of minuses" grid
			cofact[i+j*3] = m.coFact(i, j) * sign
		}
	}
	deteriminate := m[0]*cofact[0] + m[1]*cofact[1] + m[2]*cofact[2]

	// transpose cofact
	for i := 0; i < 2; i++ {
		for j := i + 1; j < 3; j++ {
			cofact[i+j*3], cofact[j+i*3] = cofact[j+i*3], cofact[i+j*3]
		}
	}
	for i := 0; i < 9; i++ {
		cofact[i] /= deteriminate
	}
	return &cofact
}

// Invert returns the inverse matrix
func (a Matrix2D) Invert() Matrix2D {
	n := &matrix3{a.A, a.C, a.E, a.B, a.D, a.F, 0, 0, 1}
	n = n.invert()
	return Matrix2D{A: n[0], C: n[1], E: n[2], B: n[3], D: n[4], F: n[5]}
}

// Mult returns a*b
func (a Matrix2D) Mult(b Matrix2D) Matrix2D {
	return Matrix2D{
		A: a.A*b.A + a.C*b.B,
		B: a.B*b.A + a.D*b.B,
		C: a.A*b.C + a.C*b.D,
		D: a.B*b.C + a.D*b.D,
		E: a.A*b.E + a.C*b.F + a.E*b.A,
		F: a.B*b.E + a.D*b.F + a.F*b.D,
	}
}

// Identity is the identity matrix
var Identity = Matrix2D{1, 0, 0, 1, 0, 0}

// TFixed transforms a fixed.Point26_6 by the matrix
func (a Matrix2D) TFixed(x fixed.Point26_6) (y fixed.Point26_6) {
	y.X = fixed.Int26_6((float64(x.X)*a.A + float64(x.Y)*a.C) + a.E*64)
	y.Y = fixed.Int26_6((float64(x.X)*a.B + float64(x.Y)*a.D) + a.F*64)
	return
}

// Transform multiples the input vector by matrix m and outputs the results vector
// components.
func (a Matrix2D) Transform(x1, y1 float64) (x2, y2 float64) {
	x2 = x1*a.A + y1*a.C + a.E
	y2 = x1*a.B + y1*a.D + a.F
	return
}

// TransformVector is a modidifed version of Transform that ignores the
// translation components.
func (a Matrix2D) TransformVector(x1, y1 float64) (x2, y2 float64) {
	x2 = x1*a.A + y1*a.C
	y2 = x1*a.B + y1*a.D
	return
}

// Scale matrix in x and y dimensions
func (a Matrix2D) Scale(x, y float64) Matrix2D {
	return Matrix2D{
		A: a.A * x,
		B: a.B * y,
		C: a.C * x,
		D: a.D * y,
		E: a.E * x,
		F: a.F * y,
	}
}

// SkewY skews the matrix in the Y dimension
func (a Matrix2D) SkewY(theta float64) Matrix2D {
	return a.Mult(Matrix2D{
		A: 1,
		B: math.Tan(theta),
		C: 0,
		D: 1,
		E: 0,
		F: 0,
	})
}

// SkewX skews the matrix in the X dimension
func (a Matrix2D) SkewX(theta float64) Matrix2D {
	return a.Mult(Matrix2D{
		A: 1,
		B: 0,
		C: math.Tan(theta),
		D: 1,
		E: 0,
		F: 0,
	})
}

// Translate translates the matrix to the x , y point
func (a Matrix2D) Translate(x, y float64) Matrix2D {
	return Matrix2D{
		A: a.A,
		B: a.B,
		C: a.C,
		D: a.D,
		E: a.E + x,
		F: a.F + y,
	}
}

// Rotate rotate the matrix by theta
func (a Matrix2D) Rotate(theta float64) Matrix2D {
	return a.Mult(Matrix2D{
		A: math.Cos(theta),
		B: math.Sin(theta),
		C: -math.Sin(theta),
		D: math.Cos(theta),
		E: 0,
		F: 0,
	})
}

// matrixAdder add points to path after applying a  matrix M to all points
type matrixAdder struct {
	path *Path
	M    Matrix2D
}

// Reset sets the matrix M to identity
func (t *matrixAdder) Reset() {
	t.M = Identity
}

// Start starts a new path
func (t *matrixAdder) Start(a fixed.Point26_6) {
	t.path.Start(t.M.TFixed(a))
}

// Line adds a linear segment to the current curve.
func (t *matrixAdder) Line(b fixed.Point26_6) {
	t.path.Line(t.M.TFixed(b))
}

// QuadBezier adds a quadratic segment to the current curve.
func (t *matrixAdder) QuadBezier(b, c fixed.Point26_6) {
	t.path.QuadBezier(t.M.TFixed(b), t.M.TFixed(c))
}

// CubeBezier adds a cubic segment to the current curve.
func (t *matrixAdder) CubeBezier(b, c, d fixed.Point26_6) {
	t.path.CubeBezier(t.M.TFixed(b), t.M.TFixed(c), t.M.TFixed(d))
}

// MoveTo transform the operation `m` by applying `t`
func (t Matrix2D) MoveTo(m OpMoveTo) fixed.Point26_6 {
	return t.TFixed(fixed.Point26_6(m))
}

// LineTo transform the operation `m` by applying `t`
func (t Matrix2D) LineTo(m OpLineTo) fixed.Point26_6 {
	return t.TFixed(fixed.Point26_6(m))
}

// QuadTo transform the operation `m` by applying `t`
func (t Matrix2D) QuadTo(m OpQuadTo) (fixed.Point26_6, fixed.Point26_6) {
	return t.TFixed(fixed.Point26_6(m[0])), t.TFixed(fixed.Point26_6(m[1]))
}

// CubicTo transform the operation `m` by applying `t`
func (t Matrix2D) CubicTo(m OpCubicTo) (fixed.Point26_6, fixed.Point26_6, fixed.Point26_6) {
	return t.TFixed(fixed.Point26_6(m[0])), t.TFixed(fixed.Point26_6(m[1])), t.TFixed(fixed.Point26_6(m[2]))
}
