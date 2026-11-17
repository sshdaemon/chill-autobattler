package sim

type ObstacleKind uint8

const (
	Water ObstacleKind = iota
	Rock
	Tree
)

// Obstacle is a circle on the ground. Water blobs of one pond share Pond.
type Obstacle struct {
	Kind ObstacleKind
	Pos  Vec
	R    Fx
	Pond int
}

// Generation rules (prototype's randomObstacles).
var (
	borderGap   = I(2)          // obstacles stay this far from the arena edge
	obstacleGap = Ratio(16, 10) // and this far from each other
	zoneClearX  = I(6)          // the middle of each deployment zone stays clear...
	zoneClearZ  = Ratio(75, 10) // ...beyond this |z|
	pondBand    = Ratio(65, 10) // ponds keep |z| + r below this
)

// GenerateObstacles builds the map for a seed: 2–3 ponds of 2–4 blobs, then
// 3–6 rocks, then 3–7 trees, each placed only where it fits.
func GenerateObstacles(seed uint64) []Obstacle {
	rnd := Sub(seed, StreamMap)
	var list []Obstacle
	fits := func(o Obstacle) bool {
		lim := ArenaHalf - borderGap
		if o.Pos.X.Abs()+o.R > lim || o.Pos.Z.Abs()+o.R > lim {
			return false
		}
		if o.Pos.Z.Abs() > zoneClearZ && o.Pos.X.Abs() < zoneClearX {
			return false
		}
		for _, p := range list {
			if !Within(o.Pos, p.Pos, o.R+p.R+obstacleGap+1) { // +1 raw: "more than", not "at least"
				continue
			}
			if Dist(o.Pos, p.Pos) <= o.R+p.R+obstacleGap {
				return false
			}
		}
		return true
	}
	place := func(make func() Obstacle, count int) {
		for n, tries := 0, 0; n < count && tries < 200; tries++ {
			if o := make(); fits(o) {
				list = append(list, o)
				n++
			}
		}
	}

	want := 2 + rnd.IntN(2)
	for p, tries := 0, 0; p < want && tries < 150; tries++ {
		c := V(rnd.Between(I(-9), I(9)), rnd.Between(Ratio(-45, 10), Ratio(45, 10)))
		r := rnd.Between(Ratio(16, 10), Ratio(26, 10))
		blobs := []Obstacle{{Kind: Water, Pond: p, Pos: c, R: r}}
		for b := 1 + rnd.IntN(3); b > 0; b-- {
			dir := Dir64(rnd.IntN(64))
			d := r.Mul(rnd.Between(Ratio(6, 10), Ratio(11, 10)))
			off := V(dir.X.Mul(d).Mul(Ratio(13, 10)), dir.Z.Mul(d).Mul(Ratio(6, 10))) // ponds stretch sideways
			blobs = append(blobs, Obstacle{Kind: Water, Pond: p, Pos: c.Add(off), R: r.Mul(rnd.Between(Ratio(6, 10), Ratio(9, 10)))})
		}
		ok := true
		for _, o := range blobs {
			if o.Pos.Z.Abs()+o.R >= pondBand || !fits(o) {
				ok = false
				break
			}
		}
		if ok {
			list = append(list, blobs...)
			p++
		}
	}
	place(func() Obstacle {
		return Obstacle{Kind: Rock, Pos: V(rnd.Between(I(-11), I(11)), rnd.Between(I(-9), I(9))), R: rnd.Between(Ratio(9, 10), Ratio(19, 10))}
	}, 3+rnd.IntN(4))
	place(func() Obstacle {
		return Obstacle{Kind: Tree, Pos: V(rnd.Between(I(-11), I(11)), rnd.Between(I(-10), I(10))), R: Ratio(8, 10)}
	}, 3+rnd.IntN(5))
	return list
}
