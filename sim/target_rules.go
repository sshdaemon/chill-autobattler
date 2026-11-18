package sim

// The other target rules (TargetNearest is in target_nearest.go). Each falls
// back to the nearest enemy when its own choice finds none.
func init() {
	registerTarget(TargetWounded, &targetRule{name: "wounded", pick: func(b *Battle, u *Unit) int {
		best := -1
		for _, i := range b.spotted[u.Team] { // lowest health share within WoundedReach
			e := b.Units[i]
			if Within(u.Pos, e.Pos, WoundedReach) && (best < 0 || e.HP.MulDiv(100, 1).Div(e.MaxHP) < b.Units[best].HP.MulDiv(100, 1).Div(b.Units[best].MaxHP)) {
				best = i
			}
		}
		if best < 0 {
			return b.nearestSpotted(u, nil)
		}
		return best
	}})
	registerTarget(TargetRangedFirst, &targetRule{name: "rangedFirst", pick: func(b *Battle, u *Unit) int {
		if i := b.nearestSpotted(u, func(e *Unit) bool { return e.Def.Tags&(TagRanged|TagCaster) != 0 }); i >= 0 {
			return i
		}
		return b.nearestSpotted(u, nil)
	}})
	registerTarget(TargetStrongest, &targetRule{name: "strongest", pick: func(b *Battle, u *Unit) int {
		best := -1
		for _, i := range b.spotted[u.Team] { // highest top damage; ties: the nearer
			e := b.Units[i]
			if best < 0 || e.Stats.DmgHi > b.Units[best].Stats.DmgHi ||
				e.Stats.DmgHi == b.Units[best].Stats.DmgHi && Dist(u.Pos, e.Pos) < Dist(u.Pos, b.Units[best].Pos) {
				best = i
			}
		}
		return best
	}})
	registerTarget(TargetFocus, &targetRule{name: "focus", pick: func(b *Battle, u *Unit) int {
		best, most := -1, 0
		for _, i := range b.spotted[u.Team] { // the enemy most allies target; ties: the nearer
			n := 0
			for _, a := range b.Units {
				if a != u && !a.Dead && a.Team == u.Team && a.Target == i {
					n++
				}
			}
			if n > most || n == most && n > 0 && Dist(u.Pos, b.Units[i].Pos) < Dist(u.Pos, b.Units[best].Pos) {
				best, most = i, n
			}
		}
		if best < 0 {
			return b.nearestSpotted(u, nil)
		}
		return best
	}})
	registerTarget(TargetProtect, &targetRule{name: "protect", pick: func(b *Battle, u *Unit) int {
		best, bd := -1, Fx(0)
		for _, a := range b.Units { // the enemy closest to one of its ranged or caster allies
			if a == u || a.Dead || a.Team != u.Team || a.Def.Tags&(TagRanged|TagCaster) == 0 {
				continue
			}
			for _, i := range b.spotted[u.Team] {
				if d := Dist(a.Pos, b.Units[i].Pos); d <= ProtectReach && (best < 0 || d < bd) {
					best, bd = i, d
				}
			}
		}
		if best < 0 {
			return b.nearestSpotted(u, nil)
		}
		return best
	}})
}
