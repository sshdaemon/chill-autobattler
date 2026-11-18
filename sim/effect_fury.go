package sim

import "fmt"

// Fury: Mod +Pct% while below Threshold% of its health (Battle fury).
func init() {
	registerEffect(Fury, &effectKind{
		name:   "fury",
		stacks: true,
		check: func(e Effect, _ *UnitDef) error {
			if e.Mod < ModDamage || e.Mod > ModSpeed || e.Threshold <= 0 || e.Threshold >= 100 {
				return fmt.Errorf("fury needs a modifier and a health threshold 1..99")
			}
			return pctIn(e, -90, 200)
		},
		self: func(b *Battle, u *Unit, a *active) {
			if u.HP.MulDiv(100, a.e.Threshold) < u.MaxHP {
				u.mods[a.e.Mod] += a.e.Pct
			}
		},
	})
}
