package sim

import "fmt"

// DamageTaken: the unit takes Pct% more damage (negative: less) when Cond
// holds: always, against ranged attacks, from the front, or while standing
// (Shield wall, Tale of the bohatyrs).
func init() {
	registerEffect(DamageTaken, &effectKind{
		name:   "damageTaken",
		stacks: true,
		check: func(e Effect, _ *UnitDef) error {
			if e.Cond > CondStanding {
				return fmt.Errorf("damage taken: unknown condition %d", e.Cond)
			}
			return pctIn(e, -90, 100)
		},
		onHurt: func(b *Battle, v, att *Unit, from Vec, amount Fx, a *active) Fx {
			switch a.e.Cond {
			case CondRanged:
				if !att.Stats.Ranged {
					return amount
				}
			case CondFront:
				if v.Face.Dot(from.Sub(v.Pos).Norm()) < CosFront {
					return amount
				}
			case CondStanding:
				if v.Moving() {
					return amount
				}
			}
			return pctOf(amount, a.e.Pct)
		},
	})
}
