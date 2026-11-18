package sim

import "fmt"

// BonusAgainst: +Pct% damage against units with Tag (Bodkin points, Hunter).
func init() {
	registerEffect(BonusAgainst, &effectKind{
		name:   "bonusAgainst",
		stacks: true,
		check: func(e Effect, _ *UnitDef) error {
			if e.Tag == 0 {
				return fmt.Errorf("bonus against needs tags")
			}
			return pctIn(e, 1, 200)
		},
		onBlow: func(b *Battle, att, v *Unit, scale, amount Fx, a *active) Fx {
			if v.Def.Tags&a.e.Tag == 0 {
				return amount
			}
			return pctOf(amount, a.e.Pct)
		},
	})
}
