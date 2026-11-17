// Package sim is the deterministic battle logic. It is shared by the browser
// (compiled to WASM, see ../wasm) and the server (replay verification).
// Rules: no floats, no maps iterated for logic, no time, no goroutines,
// randomness only from the battle's RNG.
//
// This is a placeholder toy battle (dots walk and hit each other) that proves
// the pipeline; the real rules come from the port-battle-sim OpenSpec change.
package sim

import "hash/fnv"

// Version must change whenever the rules change, so old replays are checked
// with the logic that produced them.
const Version = "0.1.0-toy"

type Unit struct {
	Team   int
	X, Z   Fx
	HP     Fx
	CoolDn int // ticks until the next attack
}

type Battle struct {
	Tick  int
	Units []Unit
	rng   *RNG
}

const (
	TicksPerSecond = 20
	speed          = One / 10 // metres per tick (2 m/s)
	reach          = One + Half
	attackTicks    = 16
)

func New(seed uint64) *Battle {
	b := &Battle{rng: NewRNG(seed)}
	for i := 0; i < 3; i++ {
		b.Units = append(b.Units,
			Unit{Team: 0, X: I(i*3 - 3), Z: I(-10), HP: I(100)},
			Unit{Team: 1, X: I(i*3 - 3), Z: I(10), HP: I(100)})
	}
	return b
}

// Winner is the surviving team, -1 while both teams stand (or if none do).
func (b *Battle) Winner() int {
	alive := [2]bool{}
	for _, u := range b.Units {
		if u.HP > 0 {
			alive[u.Team] = true
		}
	}
	switch {
	case alive[0] && !alive[1]:
		return 0
	case alive[1] && !alive[0]:
		return 1
	}
	return -1
}

func (b *Battle) Step() {
	b.Tick++
	for i := range b.Units {
		u := &b.Units[i]
		if u.HP <= 0 {
			continue
		}
		t, dist := -1, Fx(0)
		for j, e := range b.Units { // nearest living enemy, ties -> lowest index
			if e.Team == u.Team || e.HP <= 0 {
				continue
			}
			if d := Length(e.X-u.X, e.Z-u.Z); t < 0 || d < dist {
				t, dist = j, d
			}
		}
		if t < 0 {
			continue
		}
		if u.CoolDn > 0 {
			u.CoolDn--
		}
		if dist > reach {
			u.X += (b.Units[t].X - u.X).Mul(speed).Div(dist)
			u.Z += (b.Units[t].Z - u.Z).Mul(speed).Div(dist)
		} else if u.CoolDn == 0 {
			b.Units[t].HP -= I(8 + b.rng.IntN(8))
			u.CoolDn = attackTicks
		}
	}
}

// Checksum hashes the whole state; equal checksums == identical battles.
func (b *Battle) Checksum() uint64 {
	h := fnv.New64a()
	w := func(v int64) {
		var buf [8]byte
		for i := range buf {
			buf[i] = byte(v >> (8 * i))
		}
		h.Write(buf[:])
	}
	w(int64(b.Tick))
	for _, u := range b.Units {
		w(int64(u.Team))
		w(int64(u.X))
		w(int64(u.Z))
		w(int64(u.HP))
		w(int64(u.CoolDn))
	}
	return h.Sum64()
}

type Result struct {
	Winner   int
	Ticks    int
	Checksum uint64
}

// Run plays a whole battle with no player input.
func Run(seed uint64, maxTicks int) Result {
	b := New(seed)
	for b.Tick < maxTicks && b.Winner() < 0 {
		b.Step()
	}
	return Result{Winner: b.Winner(), Ticks: b.Tick, Checksum: b.Checksum()}
}
