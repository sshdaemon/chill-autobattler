package sim

import "fmt"

// Guard: the unit takes Pct% of blows aimed at allies with Tag within Radius
// (Prince's guard).
func init() {
	registerEffect(Guard, &effectKind{
		name: "guard",
		check: func(e Effect, _ *UnitDef) error {
			if e.Radius <= 0 || e.Tag == 0 {
				return fmt.Errorf("guard needs a radius and the tags it guards")
			}
			return pctIn(e, 1, 90)
		},
		intercept: func(b *Battle, g, v, att *Unit, amount Fx, a *active) Fx {
			if v.Def.Tags&a.e.Tag == 0 || !Within(g.Pos, v.Pos, a.e.Radius) {
				return 0
			}
			return amount.MulDiv(a.e.Pct, 100)
		},
	})
}
