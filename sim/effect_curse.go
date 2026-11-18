package sim

import "fmt"

// Curse: the unit's hits make the victim take Pct% more damage for Ticks
// (Curse, Shield-breaker).
func init() {
	registerEffect(Curse, &effectKind{
		name: "curse",
		check: func(e Effect, _ *UnitDef) error {
			if e.Ticks <= 0 {
				return fmt.Errorf("curse needs a duration")
			}
			return pctIn(e, 1, 100)
		},
		onHit: func(b *Battle, u, v *Unit, _ Vec, a *active) {
			if v.Dead {
				return
			}
			v.addBuff(ModTaken, a.e.Pct, a.e.Ticks)
			b.emit(Event{Kind: EvCurse, Unit: v.Index, Other: u.Index})
		},
	})
}
