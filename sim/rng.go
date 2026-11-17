package sim

// RNG is SplitMix64, written out here on purpose: math/rand's helpers (IntN etc.)
// are allowed to change between Go versions, which would break old replays.
type RNG struct{ s uint64 }

func NewRNG(seed uint64) *RNG { return &RNG{s: seed} }

func (r *RNG) Uint64() uint64 {
	r.s += 0x9e3779b97f4a7c15
	z := r.s
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// IntN returns 0..n-1 (tiny modulo bias is fine for a game).
func (r *RNG) IntN(n int) int { return int(r.Uint64() % uint64(n)) }
