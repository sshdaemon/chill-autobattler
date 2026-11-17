package sim

// RNG is SplitMix64, written out here on purpose: math/rand's helpers (IntN etc.)
// are allowed to change between Go versions, which would break old replays.
type RNG struct{ s uint64 }

func NewRNG(seed uint64) *RNG { return &RNG{s: seed} }

// Stream tags: each part of a battle draws from its own stream, so adding a
// draw in one place (say, a new map feature) doesn't shift the others.
const (
	StreamMap uint64 = iota + 1
	StreamRedArmy
	StreamRedSkills
	StreamRedFormation
	StreamCombat
)

// Sub derives an independent stream for a tag from the battle seed.
func Sub(seed, tag uint64) *RNG {
	r := NewRNG(seed ^ (tag * 0xd1b54a32d192ed03))
	r.Uint64() // decorrelate neighbouring tags
	return NewRNG(r.Uint64())
}

func (r *RNG) Uint64() uint64 {
	r.s += 0x9e3779b97f4a7c15
	z := r.s
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// IntN returns 0..n-1 (tiny modulo bias is fine for a game).
func (r *RNG) IntN(n int) int { return int(r.Uint64() % uint64(n)) }

// Frac returns a uniform Fx in [0, 1).
func (r *RNG) Frac() Fx { return Fx(r.Uint64() >> (64 - fracBits)) }

// Between returns a uniform Fx in [a, b).
func (r *RNG) Between(a, b Fx) Fx { return a + (b - a).Mul(r.Frac()) }

// Chance reports true with probability p (an Fx fraction).
func (r *RNG) Chance(p Fx) bool { return r.Frac() < p }
