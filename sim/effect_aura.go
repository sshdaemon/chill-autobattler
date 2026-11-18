package sim

import "fmt"

// Aura: allies within Radius (or, with Foes, enemies) get Mod +Pct% while the
// unit lives (songs, Foretelling, Mocking song). Not the unit itself.
func init() {
	registerEffect(Aura, &effectKind{
		name:   "aura",
		stacks: true,
		check: func(e Effect, _ *UnitDef) error {
			if e.Mod < ModDamage || e.Mod > ModSpeed || e.Radius <= 0 {
				return fmt.Errorf("aura needs a modifier and a radius")
			}
			return pctIn(e, -90, 100)
		},
		aura: func(b *Battle, src *Unit, a *active) {
			b.eachNear(src.Pos, auraRadius(src, a.e.Radius), src.Team, a.e.Foes, src, func(o *Unit) { o.mods[a.e.Mod] += a.e.Pct })
		},
	})
}
