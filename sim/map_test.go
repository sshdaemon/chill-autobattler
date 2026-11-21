package sim

import (
	"reflect"
	"testing"
)

// 2.1
func TestSameSeedSameMap(t *testing.T) {
	for _, seed := range []uint64{1, 2, 42, 0xdeadbeef} {
		a, b := GenerateObstacles(seed), GenerateObstacles(seed)
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("seed %d: maps differ", seed)
		}
	}
	if reflect.DeepEqual(GenerateObstacles(1), GenerateObstacles(2)) {
		t.Fatal("seeds 1 and 2 gave the same map")
	}
}

func TestMapRules(t *testing.T) {
	for seed := uint64(0); seed < 300; seed++ {
		obs := GenerateObstacles(seed)
		count := map[ObstacleKind]int{}
		ponds := map[int]int{}
		for i, o := range obs {
			count[o.Kind]++
			if o.Kind == Water {
				ponds[o.Pond]++
				if o.Pos.Z.Abs()+o.R >= pondBand {
					t.Errorf("seed %d: pond blob outside the middle band: %+v", seed, o)
				}
			}
			if o.Pos.X.Abs()+o.R > ArenaHalf-borderGap || o.Pos.Z.Abs()+o.R > ArenaHalf-borderGap {
				t.Errorf("seed %d: obstacle too close to the border: %+v", seed, o)
			}
			if o.Pos.Z.Abs() > zoneClearZ && o.Pos.X.Abs() < zoneClearX {
				t.Errorf("seed %d: obstacle in the middle of a deployment zone: %+v", seed, o)
			}
			for _, p := range obs[:i] {
				if o.Kind == Water && p.Kind == Water && o.Pond == p.Pond {
					continue // blobs of one pond overlap on purpose
				}
				if Dist(o.Pos, p.Pos) <= o.R+p.R+obstacleGap {
					t.Errorf("seed %d: obstacles too close: %+v %+v", seed, o, p)
				}
			}
		}
		if count[Rock] < 2 || count[Rock] > 6 || count[Tree] < 2 || count[Tree] > 8 || len(ponds) > 4 {
			t.Errorf("seed %d: obstacle counts out of range: %v, %d ponds", seed, count, len(ponds))
		}
		for p, n := range ponds {
			if n < 2 || n > 8 {
				t.Errorf("seed %d: pond %d has %d blobs", seed, p, n)
			}
		}
		// mirrored: every obstacle's reflection across the centre line is on the map
		for _, o := range obs {
			found := false
			for _, m := range obs {
				found = found || (m.Kind == o.Kind && m.R == o.R && m.Pos == V(o.Pos.X, -o.Pos.Z))
			}
			if !found {
				t.Errorf("seed %d: %+v has no mirror image", seed, o)
			}
		}
	}
}

// 2.2: grid queries on a hand-made map.
func TestGridQueries(t *testing.T) {
	m := NewMap([]Obstacle{
		{Kind: Rock, Pos: V(0, 0), R: I(1)},
		{Kind: Water, Pos: V(I(6), 0), R: I(2)},
	})
	if m.FreeAt(V(0, 0)) || m.FreeAt(V(Ratio(13, 10), 0)) {
		t.Error("rock (with unit padding) should be blocked")
	}
	if !m.FreeAt(V(I(2), 0)) || !m.FreeAt(V(I(6), 0)) {
		t.Error("open ground and water should be walkable")
	}
	if m.FreeAt(V(Ratio(139, 10), 0)) {
		t.Error("border cells should be blocked")
	}
	if !m.InWater(V(I(6), 0)) || m.InWater(V(Ratio(795, 100), 0)) || m.InWater(V(I(3), 0)) {
		t.Error("InWater")
	}
	left, right, up := V(I(-4), 0), V(I(4), 0), V(I(-4), I(4))
	if m.LineOfSight(left, right) || m.ClearView(left, right) {
		t.Error("rock should block walking and sight")
	}
	if !m.LineOfSight(left, up) || !m.ClearView(left, up) {
		t.Error("open line should be clear")
	}
	if !m.ClearView(V(I(3), 0), V(I(10), 0)) {
		t.Error("water must not block sight")
	}
	if wl := m.WetLength(V(I(3), 0), V(I(9), 0)); wl < I(3) || wl > I(4) {
		t.Errorf("wet length across a 4 m pond (inset 0.1) = %v", wl)
	}
	if m.WetLength(left, up) != 0 {
		t.Error("dry line has no wet length")
	}
	if p := m.FreeSpot(V(0, 0)); !m.FreeAt(p) || Dist(p, V(0, 0)) > I(2) {
		t.Errorf("FreeSpot from the rock centre = %v", p)
	}
	if p := m.PushOut(V(Ratio(5, 10), 0)); Dist(p, V(0, 0)) < I(1)+UnitRadius-2 {
		t.Errorf("PushOut left the unit inside the rock: %v", p)
	}
	if p := m.PushOut(V(I(20), I(-20))); p != V(ClampLim, -ClampLim) {
		t.Errorf("PushOut clamp = %v", p)
	}
	if ToCell(-ArenaHalf) != 0 || ToCell(ArenaHalf) != GridN-1 || ToCell(0) != GridN/2 || ToCell(-1) != GridN/2-1 {
		t.Error("ToCell")
	}
	if CellCenter(0) != -ArenaHalf+Cell/2 {
		t.Error("CellCenter")
	}
}

// 2.3: there is always a walkable way from Blue's zone to Red's.
func TestAlwaysAWayThrough(t *testing.T) {
	for seed := uint64(0); seed < 1000; seed++ {
		m := NewMap(GenerateObstacles(seed))
		reach := m.flood(V(0, I(10)))
		for team := 0; team < 2; team++ {
			z := Zones[team]
			for j := ToCell(z.ZMin); j <= ToCell(z.ZMax); j++ {
				for i := ToCell(z.XMin); i <= ToCell(z.XMax); i++ {
					if m.Free(i, j) && !reach[j*GridN+i] {
						t.Fatalf("seed %d: free cell (%d,%d) in zone %d can't be reached from Blue's zone", seed, i, j, team)
					}
				}
			}
		}
	}
}

// flood marks every cell reachable on foot (8-connected, no corner cutting) from p.
func (m *Map) flood(p Vec) []bool {
	seen := make([]bool, GridN*GridN)
	si, sj := m.NearestFree(cellOf(p))
	stack := []int{sj*GridN + si}
	seen[stack[0]] = true
	for len(stack) > 0 {
		c := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, n := range m.neighbours(c) {
			if !seen[n.cell] {
				seen[n.cell] = true
				stack = append(stack, n.cell)
			}
		}
	}
	return seen
}

// On a generated (mirrored) map Red's paths and free spots are the mirror
// images of Blue's, so equal choices (which side of a rock) fall the same
// way for both sides.
func TestMirroredMapChoices(t *testing.T) {
	if NewMap([]Obstacle{{Kind: Rock, Pos: V(I(3), I(3)), R: I(1)}}).Mirrored {
		t.Error("a map with one rock off the centre line is not mirrored")
	}
	for seed := uint64(1); seed <= 40; seed++ {
		b := &Battle{Map: NewMap(GenerateObstacles(seed))}
		if !b.Map.Mirrored {
			t.Fatalf("seed %d: generated map not mirrored", seed)
		}
		from, to := V(I(-9)+Half/3, I(10)), V(I(8), I(-11)+Half/7)
		blue := b.findPath(0, from, to)
		red := b.findPath(1, flipZ(from), flipZ(to))
		if len(blue) != len(red) {
			t.Fatalf("seed %d: Blue's path has %d legs, Red's mirror %d", seed, len(blue), len(red))
		}
		for i := range blue {
			if red[i] != flipZ(blue[i]) {
				t.Fatalf("seed %d: leg %d: Blue %v, Red %v", seed, i, blue[i], red[i])
			}
		}
		for _, p := range []Vec{V(0, I(5)), V(I(4), I(-2)), V(I(-6), I(9))} {
			if b.freeSpot(1, flipZ(p)) != flipZ(b.freeSpot(0, p)) {
				t.Errorf("seed %d: free spots near %v aren't mirror images", seed, p)
			}
		}
	}
}
