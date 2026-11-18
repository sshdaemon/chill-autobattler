package sim

import "fmt"

// Fear: every Ticks, enemies within Radius flee from the unit for Duration
// (Dread). Fleeing units do nothing else.
func init() {
	registerEffect(Fear, &effectKind{
		name: "fear",
		check: func(e Effect, _ *UnitDef) error {
			if e.Radius <= 0 || e.Ticks <= 0 || e.Duration <= 0 {
				return fmt.Errorf("fear needs a radius, a cooldown and a duration")
			}
			return nil
		},
		beforeAct: func(b *Battle, u *Unit, a *active) bool {
			if a.cd > 0 {
				return false
			}
			hit := false
			b.eachNear(u.Pos, a.e.Radius, u.Team, true, nil, func(o *Unit) {
				o.fear, o.fearFrom = a.e.Duration, u.Pos
				o.swing, o.dash = nil, nil
				b.emit(Event{Kind: EvFear, Unit: o.Index, Other: u.Index})
				hit = true
			})
			if hit {
				a.cd = a.e.Ticks
			}
			return false
		},
	})
}
