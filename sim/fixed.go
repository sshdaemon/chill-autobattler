package sim

import "strconv"

// Fx is a Q48.16 fixed-point number: the raw int64 holds value*65536.
// All game maths uses Fx (never float) so the browser (WASM) and the server
// (native) produce bit-identical results. Products of two Fx must stay below
// ~2^47 in value, which is far beyond the 28 m arena.
type Fx int64

const (
	fracBits    = 16
	One      Fx = 1 << fracBits
	Half     Fx = One / 2
)

// I converts an integer to Fx.
func I(n int) Fx { return Fx(n) << fracBits }

// Ratio is num/den as Fx, e.g. Ratio(13, 10) = 1.3.
func Ratio(num, den int) Fx { return Fx(int64(num) << fracBits / int64(den)) }

// Mul rounds toward zero, like Div and MulDiv, so -a*b == -(a*b): mirror-image
// positions on the two halves of the field stay exact mirror images.
func (a Fx) Mul(b Fx) Fx { return Fx(int64(a) * int64(b) / int64(One)) }
func (a Fx) Div(b Fx) Fx { return Fx((int64(a) << fracBits) / int64(b)) }

// MulDiv returns a*num/den without losing precision in between.
func (a Fx) MulDiv(num, den int) Fx { return Fx(int64(a) * int64(num) / int64(den)) }

// Floor is the largest integer <= a.
func (a Fx) Floor() int { return int(a >> fracBits) }

// Ceil is the smallest integer >= a.
func (a Fx) Ceil() int { return int((a + One - 1) >> fracBits) }

func (a Fx) Abs() Fx {
	if a < 0 {
		return -a
	}
	return a
}

// String formats with 4 decimals using integer maths only (no floats in sim/).
func (a Fx) String() string {
	neg := a < 0
	if neg {
		a = -a
	}
	whole := int64(a) >> fracBits
	frac := (int64(a) & int64(One-1)) * 10000 >> fracBits
	s := strconv.FormatInt(whole, 10) + "." + leftPad(strconv.FormatInt(frac, 10), 4)
	if neg {
		return "-" + s
	}
	return s
}

func leftPad(s string, n int) string {
	for len(s) < n {
		s = "0" + s
	}
	return s
}

func Min(a, b Fx) Fx {
	if a < b {
		return a
	}
	return b
}

func Max(a, b Fx) Fx {
	if a > b {
		return a
	}
	return b
}

func Clamp(v, lo, hi Fx) Fx { return Max(lo, Min(hi, v)) }

func isqrt(n uint64) uint64 {
	if n < 2 {
		return n
	}
	x := n
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}
	return x
}

// Vec is a point or direction on the ground plane (x right, z toward Blue's edge).
type Vec struct{ X, Z Fx }

func V(x, z Fx) Vec               { return Vec{x, z} }
func (a Vec) Add(b Vec) Vec       { return Vec{a.X + b.X, a.Z + b.Z} }
func (a Vec) Sub(b Vec) Vec       { return Vec{a.X - b.X, a.Z - b.Z} }
func (a Vec) Scale(k Fx) Vec      { return Vec{a.X.Mul(k), a.Z.Mul(k)} }
func (a Vec) Dot(b Vec) Fx        { return a.X.Mul(b.X) + a.Z.Mul(b.Z) }
func (a Vec) IsZero() bool        { return a.X == 0 && a.Z == 0 }
func (a Vec) MulDiv(n, d int) Vec { return Vec{a.X.MulDiv(n, d), a.Z.MulDiv(n, d)} }

// Cross is the z component of a×b in raw units squared; only its sign is used.
func (a Vec) Cross(b Vec) int64 { return int64(a.X)*int64(b.Z) - int64(a.Z)*int64(b.X) }

// Len is the exact floor of the length.
func (a Vec) Len() Fx {
	return Fx(isqrt(uint64(int64(a.X)*int64(a.X) + int64(a.Z)*int64(a.Z))))
}

// Dist is the distance between two points.
func Dist(a, b Vec) Fx { return b.Sub(a).Len() }

// Within reports |a-b| < r, without a square root.
func Within(a, b Vec, r Fx) bool {
	d := b.Sub(a)
	return int64(d.X)*int64(d.X)+int64(d.Z)*int64(d.Z) < int64(r)*int64(r)
}

// Norm returns the unit vector, or zero for the zero vector.
func (a Vec) Norm() Vec {
	l := a.Len()
	if l == 0 {
		return Vec{}
	}
	return Vec{a.X.Div(l), a.Z.Div(l)}
}

// Lerp is a + (b-a)*num/den.
func Lerp(a, b Vec, num, den int) Vec { return a.Add(b.Sub(a).MulDiv(num, den)) }

// Rot rotates a direction by an angle given as its cos/sin (see the angle
// tables in data.go). Positive angles turn from +Z toward +X, like the
// prototype's yaw.
func (a Vec) Rot(r Turn) Vec {
	return Vec{a.X.Mul(r.Cos) + a.Z.Mul(r.Sin), a.Z.Mul(r.Cos) - a.X.Mul(r.Sin)}
}

// Turn is a rotation stored as cos and sin, precomputed so sim needs no trig.
type Turn struct{ Cos, Sin Fx }

// For is the rotation as a unit of team sees it: Red's field is Blue's
// reflected across the centre line, which turns a rotation the other way, so
// "try +θ first" means the same to both sides.
func (r Turn) For(team int) Turn {
	if team == 1 {
		return r.Neg()
	}
	return r
}

// Neg is the opposite rotation.
func (r Turn) Neg() Turn { return Turn{r.Cos, -r.Sin} }

// TurnToward turns a unit direction toward want by at most step and returns
// the new direction (normalised).
func TurnToward(face, want Vec, step Turn) Vec {
	if want.IsZero() {
		return face
	}
	if face.Dot(want) >= step.Cos {
		return want
	}
	c := face.Cross(want)
	if c == 0 { // straight behind: turn toward +X, which is the same for both sides of the centre line
		if a, b := face.Rot(step), face.Rot(step.Neg()); a.X >= b.X {
			return a.Norm()
		} else {
			return b.Norm()
		}
	}
	if c < 0 { // want is clockwise from face (toward +X when facing +Z)
		return face.Rot(step).Norm()
	}
	return face.Rot(step.Neg()).Norm()
}
