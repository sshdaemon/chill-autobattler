package sim

import "fmt"

// Burst: every Ticks, when its target is within Value and in sight, the unit
// throws a ×Factor attack (Speed m/s) at it, with Pct% splash within Radius
// if Radius > 0 (Throwing spears, Fire pots). Throwing takes its turn.
func init() {
	registerEffect(Burst, &effectKind{
		name: "burst",
		check: func(e Effect, _ *UnitDef) error {
			if e.Value <= 0 || e.Ticks <= 0 || e.Factor <= 0 || e.Speed <= 0 || e.Radius < 0 {
				return fmt.Errorf("burst needs reach, cooldown, a damage factor and a speed")
			}
			if e.Radius > 0 {
				return pctIn(e, 1, 100)
			}
			return nil
		},
		approach: func(b *Battle, u, t *Unit, d Fx, a *active) bool {
			if a.cd > 0 || d > a.e.Value || !b.Map.LineOfFire(u.Pos, t.Pos) {
				return false
			}
			u.Face = t.Pos.Sub(u.Pos).Norm()
			b.throw(u, t, a.e)
			a.cd = a.e.Ticks
			b.emit(Event{Kind: EvBurst, Unit: u.Index, Other: t.Index})
			return true
		},
	})
}
