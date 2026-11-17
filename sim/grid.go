package sim

// Map is the battlefield: obstacles plus the 0.5 m grids derived from them.
type Map struct {
	Obstacles []Obstacle
	blocked   [GridN * GridN]bool // can't stand here (rock/tree + unit radius, or the border)
	wet       [GridN * GridN]bool // inside a pond
	opaque    [GridN * GridN]bool // rock or tree itself: blocks sight
}

var (
	walkPad  = UnitRadius + Ratio(1, 10) // cells this close to a rock/tree are blocked
	wetInset = Ratio(1, 10)              // "in water" means this far inside a blob
	sample   = Cell / 2                  // segment checks sample every 0.25 m
)

func NewMap(obstacles []Obstacle) *Map {
	m := &Map{Obstacles: obstacles}
	for j := 0; j < GridN; j++ {
		for i := 0; i < GridN; i++ {
			c := V(CellCenter(i), CellCenter(j))
			edge := i == 0 || j == 0 || i == GridN-1 || j == GridN-1
			m.blocked[j*GridN+i] = edge || m.insideSolid(c, walkPad)
			m.opaque[j*GridN+i] = m.insideSolid(c, 0)
			m.wet[j*GridN+i] = m.InWater(c)
		}
	}
	return m
}

// ToCell maps a coordinate to its grid index, clamped to the grid.
func ToCell(v Fx) int {
	i := (v + ArenaHalf).Mul(I(2)).Floor() // / Cell
	if i < 0 {
		return 0
	}
	if i > GridN-1 {
		return GridN - 1
	}
	return i
}

func CellCenter(i int) Fx { return -ArenaHalf + Cell.MulDiv(2*i+1, 2) }

func cellOf(p Vec) (int, int) { return ToCell(p.X), ToCell(p.Z) }

func (m *Map) insideSolid(p Vec, pad Fx) bool {
	for _, o := range m.Obstacles {
		if o.Kind != Water && Within(p, o.Pos, o.R+pad) {
			return true
		}
	}
	return false
}

func (m *Map) InWater(p Vec) bool {
	for _, o := range m.Obstacles {
		if o.Kind == Water && Within(p, o.Pos, o.R-wetInset) {
			return true
		}
	}
	return false
}

// Free reports whether a unit may stand in cell (i, j).
func (m *Map) Free(i, j int) bool {
	return i >= 0 && j >= 0 && i < GridN && j < GridN && !m.blocked[j*GridN+i]
}

func (m *Map) FreeAt(p Vec) bool { return m.Free(cellOf(p)) }

func (m *Map) Wet(i, j int) bool { return m.wet[j*GridN+i] }

// segment returns the number of 0.25 m steps for a→b and a function giving
// the point at step s (0..steps).
func segment(a, b Vec) (int, func(s int) Vec) {
	steps := Dist(a, b).Div(sample).Ceil()
	return steps, func(s int) Vec { return Lerp(a, b, s, steps) }
}

// LineOfSight reports whether a unit can walk straight from a to b.
func (m *Map) LineOfSight(a, b Vec) bool {
	steps, at := segment(a, b)
	for s := 1; s <= steps; s++ {
		if !m.FreeAt(at(s)) {
			return false
		}
	}
	return true
}

// ClearView reports whether rocks and trees leave the view from a to b open.
// The end cell itself may be opaque: you can see a rock.
func (m *Map) ClearView(a, b Vec) bool {
	steps, at := segment(a, b)
	for s := 1; s < steps; s++ {
		i, j := cellOf(at(s))
		if m.opaque[j*GridN+i] {
			return false
		}
	}
	return true
}

// LineOfFire reports whether a shot from a to b clears rocks and trees.
// Same test as sight: the obstacles themselves block, not the walking
// padding round them (a unit may stand in a padded cell, and the prototype's
// walkability test then left two archers unable to shoot each other forever).
func (m *Map) LineOfFire(a, b Vec) bool { return m.ClearView(a, b) }

// WetLength estimates the metres of water along a straight segment.
func (m *Map) WetLength(a, b Vec) Fx {
	steps := Dist(a, b).Div(sample).Ceil()
	if steps < 1 {
		steps = 1
	}
	n := 0
	for s := 0; s <= steps; s++ {
		i, j := cellOf(Lerp(a, b, s, steps))
		if m.wet[j*GridN+i] {
			n++
		}
	}
	return Dist(a, b).MulDiv(n, steps+1)
}

// NearestFree returns (i, j) if free, else the first free cell found in
// growing square rings (up to 9 cells away).
func (m *Map) NearestFree(i, j int) (int, int) {
	if m.Free(i, j) {
		return i, j
	}
	for r := 1; r < 10; r++ {
		for dj := -r; dj <= r; dj++ {
			for di := -r; di <= r; di++ {
				if m.Free(i+di, j+dj) {
					return i + di, j + dj
				}
			}
		}
	}
	return i, j
}

// FreeSpot snaps a point to the centre of the nearest free cell.
func (m *Map) FreeSpot(p Vec) Vec {
	i, j := m.NearestFree(cellOf(p))
	return V(CellCenter(i), CellCenter(j))
}

// PushOut moves a unit out of rocks and trees and back inside the arena.
func (m *Map) PushOut(p Vec) Vec {
	for _, o := range m.Obstacles {
		if o.Kind == Water {
			continue
		}
		min := o.R + UnitRadius
		if d := p.Sub(o.Pos); !d.IsZero() && Within(p, o.Pos, min) {
			p = o.Pos.Add(d.Norm().Scale(min))
		}
	}
	return V(Clamp(p.X, -ClampLim, ClampLim), Clamp(p.Z, -ClampLim, ClampLim))
}
