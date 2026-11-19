package sim

import "fmt"

// BurningGround: the unit's hits leave fire of Radius under the victim for
// Duration; enemies standing in it take Pct% of the unit's blow every second
// (Burning ground). A new fire close to one of its own refreshes that one;
// both emit EvBurn (Unit: the victim it lit under), so the client can keep
// the flames as long as the fire lasts.
func init() {
	registerEffect(BurningGround, &effectKind{
		name: "burningGround",
		check: func(e Effect, _ *UnitDef) error {
			if e.Radius <= 0 || e.Duration <= 0 {
				return fmt.Errorf("burning ground needs a radius and a duration")
			}
			return pctIn(e, 1, 100)
		},
		onHit: func(b *Battle, u, v *Unit, _ Vec, a *active) {
			for i := range b.zones {
				if z := &b.zones[i]; z.src == u.Index && Within(z.pos, v.Pos, a.e.Radius) {
					z.ticks = a.e.Duration
					b.emit(Event{Kind: EvBurn, Unit: v.Index, Other: u.Index})
					return
				}
			}
			b.zones = append(b.zones, zone{src: u.Index, team: u.Team, pos: v.Pos, radius: a.e.Radius, pct: a.e.Pct, ticks: a.e.Duration})
			b.emit(Event{Kind: EvBurn, Unit: v.Index, Other: u.Index})
		},
	})
}
