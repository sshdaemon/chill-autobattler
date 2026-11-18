package sim

import "fmt"

// SplashOnHit: the victim's allies within Radius take Pct% of a landed
// projectile's hit (Blast).
func init() {
	registerEffect(SplashOnHit, &effectKind{
		name: "splashOnHit",
		check: func(e Effect, d *UnitDef) error {
			if err := needsRanged(e, d); err != nil {
				return err
			}
			if e.Radius <= 0 {
				return fmt.Errorf("splash needs a radius")
			}
			return pctIn(e, 1, 100)
		},
		onShotLands: func(b *Battle, u, v *Unit, s *Shot, a *active) {
			sp := a.e
			b.emit(Event{Kind: EvBlast, Unit: v.Index, Other: u.Index})
			for _, e := range b.Units {
				if e != v && !e.Dead && e.Team == v.Team && Within(e.Pos, v.Pos, sp.Radius) {
					b.damage(u, e, s.Start, Ratio(sp.Pct, 100))
				}
			}
		},
	})
}
