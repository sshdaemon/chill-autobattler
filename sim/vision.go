package sim

// SeenCells marks every grid cell a living unit of team can see: within its
// sight range and not blocked by rocks or trees (the same rules as canSee).
// Presentation only (fog of war): it reads the state and changes nothing.
func (b *Battle) SeenCells(team int) []bool {
	seen := make([]bool, GridN*GridN)
	for _, u := range b.Units {
		if u.Dead || u.Team != team {
			continue
		}
		r := u.Stats.Sight
		rc := r.Div(Cell).Ceil()
		ci, cj := cellOf(u.Pos)
		for j := max(0, cj-rc); j <= min(GridN-1, cj+rc); j++ {
			for i := max(0, ci-rc); i <= min(GridN-1, ci+rc); i++ {
				k := j*GridN + i
				if seen[k] {
					continue
				}
				c := V(CellCenter(i), CellCenter(j))
				if Dist(u.Pos, c) <= r && b.Map.ClearView(u.Pos, c) {
					seen[k] = true
				}
			}
		}
	}
	return seen
}

// SeenBy reports whether team currently sees unit (its own units always).
func (b *Battle) SeenBy(team int, u *Unit) bool {
	if u.Team == team {
		return true
	}
	for _, i := range b.spotted[team] {
		if i == u.Index {
			return true
		}
	}
	return false
}
