package sim

import "fmt"

// SlowOnHit: Pct% chance per landed projectile that the victim moves ×Factor
// for Ticks (Frost).
func init() {
	registerEffect(SlowOnHit, &effectKind{
		name: "slowOnHit",
		check: func(e Effect, d *UnitDef) error {
			if err := needsRanged(e, d); err != nil {
				return err
			}
			if e.Ticks <= 0 || e.Factor <= 0 || e.Factor >= One {
				return fmt.Errorf("slow needs ticks > 0 and a factor in (0, 1)")
			}
			return pctIn(e, 1, 100)
		},
		onShotLands: func(b *Battle, u, v *Unit, _ *Shot, a *active) {
			if sl := a.e; !v.Dead && b.rng.IntN(100) < sl.Pct {
				v.slow, v.slowFactor = sl.Ticks, sl.Factor
				b.emit(Event{Kind: EvFrost, Unit: v.Index, Other: u.Index})
			}
		},
	})
}
