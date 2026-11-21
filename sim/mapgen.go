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

// GenerateObstacles builds the map for a seed. The map is mirrored across
// the centre line (z → −z), so Blue and Red face the same field: each feature
// is generated in Red's half and reflected into Blue's. A rock or tree close
// to the centre line sits on it, once; a pond that overlaps its reflection
// becomes one pond. 1–2 ponds of 2–4 blobs (and their reflections), 2–3 rocks,
// 2–4 trees (each a pair, or one on the centre line).
func GenerateObstacles(seed uint64) []Obstacle {
	rnd := Sub(seed, StreamMap)
	var list []Obstacle
	// clear: o keeps its distance from everything placed and from extra,
	// except blobs of its own pond, which overlap on purpose
	clear := func(o Obstacle, extra []Obstacle) bool {
		lim := ArenaHalf - borderGap
		if o.Pos.X.Abs()+o.R > lim || o.Pos.Z.Abs()+o.R > lim {
			return false
		}
		if o.Pos.Z.Abs() > zoneClearZ && o.Pos.X.Abs() < zoneClearX {
			return false
		}
		for _, others := range [][]Obstacle{list, extra} {
			for _, p := range others {
				if o.Kind == Water && p.Kind == Water && o.Pond == p.Pond {
					continue
				}
				if !Within(o.Pos, p.Pos, o.R+p.R+obstacleGap+1) { // +1 raw: "more than", not "at least"
					continue
				}
				if Dist(o.Pos, p.Pos) <= o.R+p.R+obstacleGap {
					return false
				}
			}
		}
		return true
	}
	mirror := func(o Obstacle) Obstacle { o.Pos.Z = -o.Pos.Z; return o }
	// place puts count features: each made in Red's half, then a pair with its
	// reflection, or one on the centre line when it is too close to it
	place := func(make func() Obstacle, count int) {
		for n, tries := 0, 0; n < count && tries < 200; tries++ {
			o := make()
			set := []Obstacle{o, mirror(o)}
			if -o.Pos.Z < o.R+obstacleGap/2 {
				o.Pos.Z = 0
				set = []Obstacle{o}
			}
			ok := true
			for k, x := range set {
				ok = ok && clear(x, set[:k])
			}
			if ok {
				list = append(list, set...)
				n++
			}
		}
	}

	want := 1 + rnd.IntN(2)
	pond := 0
	for p, tries := 0, 0; p < want && tries < 150; tries++ {
		c := V(rnd.Between(I(-9), I(9)), rnd.Between(Ratio(-45, 10), 0))
		r := rnd.Between(Ratio(16, 10), Ratio(26, 10))
		blobs := []Obstacle{{Kind: Water, Pond: pond, Pos: c, R: r}}
		for b := 1 + rnd.IntN(3); b > 0; b-- {
			dir := Dir64(rnd.IntN(64))
			d := r.Mul(rnd.Between(Ratio(6, 10), Ratio(11, 10)))
			off := V(dir.X.Mul(d).Mul(Ratio(13, 10)), dir.Z.Mul(d).Mul(Ratio(6, 10))) // ponds stretch sideways
			blobs = append(blobs, Obstacle{Kind: Water, Pond: pond, Pos: c.Add(off), R: r.Mul(rnd.Between(Ratio(6, 10), Ratio(9, 10)))})
		}
		joined := false // the pond reaches its reflection: one pond
		for _, a := range blobs {
			for _, b := range blobs {
				joined = joined || Within(a.Pos, mirror(b).Pos, a.R+b.R)
			}
		}
		next := pond + 2
		var set []Obstacle
		for _, o := range blobs {
			m := mirror(o)
			if joined {
				next = pond + 1
			} else {
				m.Pond = pond + 1
			}
			set = append(set, o, m)
		}
		ok := true
		for k, o := range set {
			if o.Pos.Z.Abs()+o.R >= pondBand || !clear(o, set[:k]) {
				ok = false
				break
			}
		}
		if ok {
			list = append(list, set...)
			pond = next
			p++
		}
	}
	place(func() Obstacle {
		return Obstacle{Kind: Rock, Pos: V(rnd.Between(I(-11), I(11)), rnd.Between(I(-9), 0)), R: rnd.Between(Ratio(9, 10), Ratio(19, 10))}
	}, 2+rnd.IntN(2))
	place(func() Obstacle {
		return Obstacle{Kind: Tree, Pos: V(rnd.Between(I(-11), I(11)), rnd.Between(I(-10), 0)), R: Ratio(8, 10)}
	}, 2+rnd.IntN(3))
	return list
}
