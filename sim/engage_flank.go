package sim

// EngageFlank: with no target, advance like a charge; with a target it can't
// hit yet, swing round its side first: head for a point FlankWide beside it
// (on the side the unit already is), then close in. Once round, or close,
// it chases straight like anyone else for the rest of the battle.
func init() {
	registerEngage(EngageFlank, &engageRule{name: "flank",
		idle: func(b *Battle, u *Unit) { b.advance(u) },
		chase: func(b *Battle, u, t *Unit) bool {
			if u.flanked || Within(u.Pos, t.Pos, FlankNear) {
				u.flanked = true
				return false
			}
			if u.flankSide == 0 {
				u.flankSide = One
				if u.Pos.X < t.Pos.X {
					u.flankSide = -One
				}
			}
			lim := ArenaHalf - I(2)
			goal := V(Clamp(t.Pos.X+u.flankSide.Mul(FlankWide), -lim, lim), t.Pos.Z)
			if Within(u.Pos, goal, AdvanceReach.Mul(I(2))) {
				u.flanked = true
				return false
			}
			b.moveToward(u, b.Map.FreeSpot(goal))
			return true
		},
	})
}
