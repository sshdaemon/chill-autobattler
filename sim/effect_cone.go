package sim

import "fmt"

// Cone: the unit's blow hits every enemy within its range in a cone toward
// its target (Factor: the cosine of the cone's half-angle), each a full hit
// with the unit's on-hit effects (the siphonarios' fire jet). A trait: put it
// in the roster entry's Traits.
func init() {
	registerEffect(Cone, &effectKind{
		name: "cone",
		check: func(e Effect, d *UnitDef) error {
			if d.Stats.Ranged || e.Factor <= 0 || e.Factor >= One {
				return fmt.Errorf("a cone needs a non-ranged unit and a cosine in (0, 1)")
			}
			return nil
		},
		strike: func(b *Battle, u, v *Unit, a *active) bool {
			dir := v.Pos.Sub(u.Pos).Norm()
			b.emit(Event{Kind: EvJet, Unit: u.Index, Other: v.Index})
			for _, o := range b.Units {
				if o.Dead || o.Team == u.Team || !Within(u.Pos, o.Pos, u.Stats.Range+MeleeSlack) {
					continue
				}
				if o != v && dir.Dot(o.Pos.Sub(u.Pos).Norm()) < a.e.Factor {
					continue
				}
				b.damage(u, o, u.Pos, One)
				b.landed(u, o, u.Pos)
			}
			return true
		},
	})
}
