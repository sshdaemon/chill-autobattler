package sim

import "fmt"

// Phalanx: Mod +Pct% for each ally of the unit's own type within Radius, up
// to Threshold of them (Phalanx: damage taken -8% each).
func init() {
	registerEffect(Phalanx, &effectKind{
		name:   "phalanx",
		stacks: true,
		check: func(e Effect, _ *UnitDef) error {
			if e.Mod < ModDamage || e.Mod > ModSpeed || e.Radius <= 0 || e.Threshold <= 0 {
				return fmt.Errorf("phalanx needs a modifier, a radius and a maximum count")
			}
			return pctIn(e, -50, 50)
		},
		self: func(b *Battle, u *Unit, a *active) {
			n := 0
			b.eachNear(u.Pos, a.e.Radius, u.Team, false, u, func(o *Unit) {
				if o.Class == u.Class {
					n++
				}
			})
			u.mods[a.e.Mod] += a.e.Pct * min(n, a.e.Threshold)
		},
	})
}
