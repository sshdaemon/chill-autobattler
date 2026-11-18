package sim

// TargetNearest: the closest enemy the unit's team sees; ties go to the lower index.
func init() {
	registerTarget(TargetNearest, &targetRule{name: "nearest", pick: func(b *Battle, u *Unit) int {
		best, bd := -1, Fx(0)
		for _, i := range b.spotted[u.Team] {
			if d := Dist(u.Pos, b.Units[i].Pos); best < 0 || d < bd {
				best, bd = i, d
			}
		}
		return best
	}})
}
