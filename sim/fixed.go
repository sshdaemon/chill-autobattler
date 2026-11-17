package sim

import "fmt"

// Fx is a Q48.16 fixed-point number: the raw int64 holds value*65536.
// All game maths uses Fx (never float) so the browser (WASM) and the server
// (native) produce bit-identical results. Products of two Fx must stay below
// ~2^47 in value, which is far beyond any map size we use.
type Fx int64

const (
	fracBits    = 16
	One      Fx = 1 << fracBits
	Half     Fx = One / 2
)

func I(n int) Fx            { return Fx(n) << fracBits }
func Ratio(num, den int) Fx { return Fx(int64(num) << fracBits / int64(den)) }
func (a Fx) Mul(b Fx) Fx    { return Fx((int64(a) * int64(b)) >> fracBits) }
func (a Fx) Div(b Fx) Fx    { return Fx((int64(a) << fracBits) / int64(b)) }
func (a Fx) Int() int       { return int(a >> fracBits) }
func (a Fx) Abs() Fx {
	if a < 0 {
		return -a
	}
	return a
}
func (a Fx) String() string { return fmt.Sprintf("%.4f", float64(a)/float64(One)) } // display only

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

// Length of the vector (x, z), exact integer square root.
func Length(x, z Fx) Fx {
	return Fx(isqrt(uint64(int64(x)*int64(x) + int64(z)*int64(z))))
}

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
