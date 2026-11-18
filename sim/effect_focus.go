package sim

import "fmt"

// Focus: +Pct% damage when Threshold or more of its allies fight the same
// target (Volley).
func init() {
	registerEffect(Focus, &effectKind{
		name: "focus",
		check: func(e Effect, _ *UnitDef) error {
			if e.Threshold <= 0 {
				return fmt.Errorf("focus needs a number of allies")
			}
			return pctIn(e, 1, 200)
		},
		onBlow: func(b *Battle, att, v *Unit, scale, amount Fx, a *active) Fx {
			n := 0
			for _, o := range b.Units {
				if o != att && !o.Dead && o.Team == att.Team && o.Target == v.Index {
					n++
				}
			}
			if n < a.e.Threshold {
				return amount
			}
			return pctOf(amount, a.e.Pct)
		},
	})
}
