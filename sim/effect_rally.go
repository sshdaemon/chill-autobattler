package sim

import "fmt"

// Rally: when an ally dies within Radius, the allies within Radius get Mod
// +Pct% for Ticks (Lament for the fallen).
func init() {
	registerEffect(Rally, &effectKind{
		name:   "rally",
		stacks: true,
		check: func(e Effect, _ *UnitDef) error {
			if e.Mod < ModDamage || e.Mod > ModSpeed || e.Radius <= 0 || e.Ticks <= 0 {
				return fmt.Errorf("rally needs a modifier, a radius and a duration")
			}
			return pctIn(e, -90, 200)
		},
		onDeath: func(b *Battle, holder, dead *Unit, a *active) {
			r := auraRadius(holder, a.e.Radius)
			if dead.Team != holder.Team || !Within(holder.Pos, dead.Pos, r) {
				return
			}
			b.eachNear(holder.Pos, r, holder.Team, false, nil, func(o *Unit) {
				o.addBuff(a.e.Mod, a.e.Pct, a.e.Ticks)
				b.emit(Event{Kind: EvRally, Unit: o.Index, Other: holder.Index})
			})
		},
	})
}
